## Context

data-service 层的 `POST /load_balancer_exclusive_clusters/list`（B-04 已实现）返回 `dataproto.ExclusiveClusterListResult{Count, Details []corelb.ExclusiveClusterRaw}`，其中 `Extension` 是未解析的 `json.RawMessage`——因为该接口不区分 vendor，无法在 data-service 层确定具体反序列化成哪个 vendor 的强类型结构体。

已发布的 API 文档 `docs/api-docs/web-server/docs/resource/load-balancer/list_exclusive_cluster.md`（B-04 变更已同步字段）约定 cloud-server 层对外响应的 `extension` 是**结构化对象**（`extension[tcloud]` 12 个具名字段 + `clusters_zone` 嵌套对象），而不是原始 JSON 字符串或未加工的 `json.RawMessage`。

当前独占集群只支持腾讯云（`vendor=tcloud`，方案设计「一、背景与目标」1.3 范围边界已明确），因此 cloud-server 层只需处理单一 vendor 的反序列化，不需要引入通用的多 vendor 分发框架。

项目内资源视角列表查询的标准鉴权模式是 `handler.ListResourceAuthRes`（`cert.ListCert`、`load-balancer.ListLoadBalancer` 均采用），按 `meta.ResType` + `meta.Find` 对请求方做资源过滤后再转发查询，不是先查询再逐条鉴权。

## Goals / Non-Goals

**Goals:**
- 新增 cloud-server 资源视角列表接口，字段与过滤能力与已发布 API 文档一致
- `extension` 按 tcloud 强类型返回，与文档响应示例完全匹配
- 复用项目标准的 `ListResourceAuthRes` 鉴权模式，不引入新的鉴权路径

**Non-Goals:**
- 不实现业务视角查询接口（需求单未包含，且业务视角的语义是"当前业务已分配集群+标签聚合"，属于 N-04，另有单据）
- 不支持除 tcloud 外的其它 vendor（方案设计范围边界已排除，`ExclusiveClusterExtension` 接口约束目前只有一个实现类型）
- 不修改 data-service 层 `List` 接口或返回结构（`ExclusiveClusterRaw` 保持不变，供未来多 vendor 场景复用）

## Decisions

### 1. 鉴权模式采用 `handler.ListResourceAuthRes` + `meta.LoadBalancer`/`meta.Find`，不新增 IAM 资源类型

**选择**：与方案设计「十、权限与审计」10.1 一致，独占集群复用 `meta.LoadBalancer` 资源类型，不新增独立的 IAM 资源类型。Handler 直接调用 `handler.ListResourceAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find, Filter: req.Filter})`，返回的 `expr` 直接作为 data-service 查询的 filter，无权限时返回空列表（`noPermFlag=true` 分支），与 `cert.listCert` 完全一致的写法。

**原因**：`ListResourceAuthRes` 是项目内所有资源视角列表接口的统一模式，独占集群作为 CLB 的附属资源没有理由另立机制；不新增 IAM 资源类型可避免权限模型膨胀（方案设计已有此结论，DAO 层的分配业务接口设计也遵循同一决策）。

**备选方案**：为独占集群单独注册 IAM 资源类型 → 已否决，与方案设计"复用 `meta.LoadBalancer`，不新增 IAM 资源类型"结论冲突，且列表查询没有跨业务的复杂权限诉求，无需独立建模。

### 2. `extension` 在 cloud-server 层反序列化为 `TCloudExclusiveClusterExtension` 强类型，逐条转换，遇解析失败直接报错而非跳过

**选择**：Handler 拿到 data-service 返回的 `[]corelb.ExclusiveClusterRaw` 后，逐条 `json.Unmarshal(one.Extension, &TCloudExclusiveClusterExtension{})`，构造 `corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]` 列表返回给前端；任一条目 `Extension` 解析失败视为数据异常，整体返回 `errf.New(errf.DBOpFailed, ...)`，不做静默跳过（避免让前端拿到"看起来完整但缺条目"的错误数据）。

**原因**：当前唯一 vendor 是 tcloud，反序列化目标类型是确定的，不需要按 `vendor` 字段做运行时分支；写入侧（B-04 的 `BatchCreateExclusiveCluster`/`BatchUpdateExclusiveCluster`）已经保证 `extension` 是合法的 `TCloudExclusiveClusterExtension` JSON，正常路径不应出现解析失败，解析失败大概率意味着数据损坏或字段定义不一致，直接报错比静默跳过更利于尽早发现问题。

**备选方案**：保持 `json.RawMessage` 透传，把结构化解析全部交给前端 → 已否决，与已发布 API 文档"`extension` 是结构化对象"的响应契约不符，前端也无需关心 vendor 分发逻辑。

### 3. 复用 data-service 现有 `core.ListReq` 作为透传结构，不新建 cloud-server 专属请求体

**选择**：cloud-server 层的 `ListExclusiveCluster` 直接使用 `hcm/pkg/api/core` 的 `core.ListReq` 接收请求（`Filter`+`Page`），鉴权后把 `expr` 覆写进新的 `core.ListReq` 转发给 data-service，不定义 `pkg/api/cloud-server/load-balancer` 专属的 `ListExclusiveClusterReq`。

**原因**：本接口没有 cloud-server 特有的入参（不像 assign 接口需要 `bk_biz_id` 做目标业务传递），`core.ListReq` 字段与语义完全够用；`cert.ListCert` 虽然用了 `proto.ListReq`（cloud-server 包装类型），但那是历史遗留的字段扩展点，本接口没有类似扩展诉求，直接复用 `core.ListReq` 更简单，减少一层类型转换代码。

**备选方案**：仿照 `cert.ListCert` 定义 `cslb.ListExclusiveClusterReq` 包装类型 → 已否决，当前没有需要包装的额外字段，直接复用可减少约 20 行样板代码；若未来有 cloud-server 特有过滤需求，可在那时再引入包装类型，成本可控。

## Risks / Trade-offs

- **[Risk] `extension` 解析失败导致列表接口整体报错，运营在数据异常时完全看不到列表** → **Mitigation**：正常写入路径（B-04）已保证数据合法性，此风险仅在人工误改 DB 或未来误改 `TCloudExclusiveClusterExtension` 字段定义时触发；`errf.DBOpFailed` 错误信息需包含具体的集群 `id`，便于运营/开发定位问题记录而非盲猜
- **[Trade-off] 只支持单一 tcloud vendor 的强类型转换，未来若接入第二个云厂商需要在本接口补充按 `vendor` 字段的分支逻辑** → 与方案设计当前范围一致（方案设计 1.3 明确"本次仅支持腾讯云"），非本变更需要提前设计的问题，留给届时的新变更处理
- **[Trade-off] 复用 `meta.LoadBalancer` 而非独立 IAM 资源类型，意味着"能看 CLB 列表"和"能看独占集群列表"权限位相同** → 与方案设计既定决策一致，非本变更引入的新风险

## Migration Plan

纯新增只读接口，不涉及数据变更，无需迁移。部署顺序：随版本发布即可生效，无前置依赖（B-04 已发布）；回滚只需下线该路由，不影响任何存量数据或其它接口。
