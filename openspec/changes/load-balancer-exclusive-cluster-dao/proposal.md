## Why

腾讯云独占集群（TGW 四层 / STGW 七层）是 CLB 的独享转发资源，HCM 当前完全没有对应的表、DAO、data-service 接口：运营无法查看/管理独占集群清单，"分配业务"（把集群归属指定给业务）无落地存储，购买链路的归属校验无从做起。后续「独占集群同步」「CLB 同步补齐独占字段」「资源视角查询+分配业务」「购买链路支持独占型规格」「CLB 查询接口修改」等 5 个后端子任务均直接依赖本变更产出的表结构、DAO 接口与枚举定义，是 Wave 1 基础设施，需最先落地。

技术方案：[负载均衡-购买支持独占集群-后端方案设计](https://iwiki.woa.com/p/4040244998)「五、数据模型」「九、data-service 层改造」「十二、后端任务拆解 Wave1」；已澄清的需求单：[购买支持独占集群DAO层](https://tapd.woa.com/tapd_fe/69995598/story/detail/1069995598138302497)（F-001~F-004，AC-001~AC-007）。

## What Changes

- 新增 `load_balancer_exclusive_cluster` 表 + SQL 迁移脚本（`scripts/sql/9999_*`），注册表名枚举、`id_generator` 种子
- 新增 `ClusterType`（TGW/STGW）、`ClusterNetwork`（Public/Private）两个枚举，各带 `Validate()`；补齐云资源类型 `LoadBalancerExclusiveClusterCloudResType` 与审计资源类型 `LoadBalancerExclusiveClusterAuditResType` 常量注册
- 新增独占集群 DAO 层接口：`List`、`BatchCreateWithTx`（固定写 `bk_biz_id=-1`）、`BatchUpdateWithTx`（通用批量更新，对齐负载均衡自身 `LoadBalancerDao.Update` 的写法）、`BatchDeleteWithTx`
- 新增 data-service 层 5 个 HTTP 接口（列表查询、按 vendor 批量创建、按 vendor 批量更新云属性、批量更新 `bk_biz_id`、批量删除）+ 对应核心模型/请求响应体/客户端封装；云属性更新与 `bk_biz_id` 分配业务两个接口各自使用独立的窄请求体，但底层复用同一个 `BatchUpdateWithTx`，靠调用方构造的字段差异（零值自动跳过）实现协议隔离
- **不包含**（详见 Impact 与需求单"本期不包含"）：`TCloudClbExtension` 扩展字段、适配层接入 `DescribeExclusiveClusters`、res-sync 同步逻辑、cloud-server 层查询/分配业务的鉴权与前置校验、审计 build 方法实现、购买链路改造——均归属其余 5 个后端子任务



## Capabilities



### New Capabilities

- `load-balancer-exclusive-cluster-crud`：独占集群表结构、SQL 迁移、枚举/常量注册、DAO 层批量 CRUD（含协议隔离的 `bk_biz_id` 独立更新接口）、data-service 层 HTTP 路由与客户端封装



### Modified Capabilities

（无：本变更纯新增表与新增接口，不修改任何既有 spec 的行为）

## Impact

- **pkg/dal/table**：新增 `pkg/dal/table/cloud/load-balancer/exclusive_cluster.go`；`pkg/dal/table/table.go` 新增表名枚举
- **pkg/dal/dao**：新增 `pkg/dal/dao/cloud/load-balancer/exclusive_cluster.go`、`pkg/dal/dao/types/load_balancer_exclusive_cluster.go`；`pkg/dal/dao/dao.go` 注册新 DAO
- **pkg/criteria/enumor**：`load_balancer.go` 新增 `ClusterType`/`ClusterNetwork`；`cloud_resource_type.go`、`audit.go` 新增常量
- **pkg/api**：新增 `pkg/api/core/cloud/load-balancer/exclusive_cluster.go`（核心模型+Extension 强类型）、`pkg/api/data-service/cloud/load-balancer/exclusive_cluster.go`（请求/响应体）
- **pkg/client**：新增 `pkg/client/data-service/global/load_balancer_exclusive_cluster.go`
- **cmd/data-service**：新增 `cmd/data-service/service/cloud/load-balancer/exclusive_cluster*.go`；`service.go` 的 `InitService` 注册路由
- **scripts/sql**：新增迁移脚本 `9999_{date}_load_balancer_exclusive_cluster.sql`
- 纯新增，不影响任何现有表结构、接口或调用方；无需数据迁移

