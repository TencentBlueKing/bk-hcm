## 1. cloud-server 响应类型

- [x] 1.1 在 `pkg/api/cloud-server/load-balancer/exclusive_cluster.go`（新文件）定义 `ListExclusiveClusterResult = core.ListResultT[corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]]`，复用 `pkg/api/core/cloud/load-balancer` 已有的 `ExclusiveCluster[Ext]` 泛型与 `TCloudExclusiveClusterExtension`，不新建重复字段的结构体

## 2. cloud-server handler

- [x] 2.1 新增 `cmd/cloud-server/service/load-balancer/exclusive_cluster.go`：实现 `ListExclusiveCluster(cts *rest.Contexts) (any, error)`
- [x] 2.2 解析请求为 `core.ListReq` 并 `Validate()`
- [x] 2.3 调用 `handler.ListResourceAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find, Filter: req.Filter})` 做鉴权与过滤条件收窄（`noPermFlag=true` 时直接返回 `count:0, details:[]`），与 `cert.listCert` 写法一致
- [x] 2.4 用鉴权后的 `expr` 构造新的 `core.ListReq{Filter: expr, Page: req.Page}`，调用 `svc.client.DataService().Global.ListExclusiveCluster`（`pkg/client/data-service/global` 已有方法；实测该方法挂在 `Global` 顶层而非 `Global.LoadBalancer` 下，已按实际签名调用）
- [x] 2.5 `page.count=true` 时直接透传 `Count`，`details` 返回空数组
- [x] 2.6 `page.count=false` 时遍历 `result.Details`（`[]corelb.ExclusiveClusterRaw`），逐条 `json.Unmarshal(one.Extension, &corelb.TCloudExclusiveClusterExtension{})`；解析失败时返回错误并在错误信息中包含该记录 `id`，不做部分成功（`errf.DBOpFailed` 在当前 `errf` 错误码表中不存在，改用 `fmt.Errorf` 包装，与 `permission-template/list.go` 现有 extension 解析失败处理方式一致）
- [x] 2.7 解析成功的记录组装为 `corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]{BaseExclusiveCluster: one.BaseExclusiveCluster, Extension: &ext}`，汇总为 `ListExclusiveClusterResult` 返回

## 3. 路由注册

- [x] 3.1 `cmd/cloud-server/service/load-balancer/load_balancer.go` 的资源视角路由注册处新增 `h.Add("ListExclusiveCluster", http.MethodPost, "/load_balancers/exclusive_clusters/list", svc.ListExclusiveCluster)`

## 4. 文档核对

- [x] 4.1 核对 `docs/api-docs/web-server/docs/resource/load-balancer/list_exclusive_cluster.md` 的所需权限描述（当前为"资源查看"，需确认与 `meta.LoadBalancer`+`meta.Find` 的中文表述一致）、响应字段与本变更最终实现（字段名、`extension` 结构）逐项核对，如有偏差以实现为准更新文档 —— 核对结果：权限表述与其它同类只读接口一致，无需改动

## 5. 自测（用户执行，不在本次实现范围内）

- [ ] 5.1 本地起 cloud-server + data-service，验证 `ListExclusiveCluster` 鉴权、分页、`extension` 反序列化行为

