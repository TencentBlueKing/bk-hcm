# load-balancer-exclusive-cluster-crud Specification

## Purpose
TBD - created by archiving change load-balancer-exclusive-cluster-dao. Update Purpose after archive.
## Requirements
### Requirement: 独占集群表 DDL

系统 SHALL 提供 `load_balancer_exclusive_cluster` 表的 SQL DDL 迁移脚本，包含以下字段：`id`, `cloud_id`, `name`, `vendor`, `account_id`, `bk_biz_id`, `region`, `zone`, `cluster_type`, `cluster_tag`, `network`, `isp`, `egress`, `ip_version`, `max_conn`（nullable）, `clb_resource_count`, `extension`, `memo`, `tenant_id`, `creator`, `reviser`, `created_at`, `updated_at`。唯一键 `(cloud_id, account_id, tenant_id)`，索引含 `bk_biz_id`、`cluster_tag`、`(region, cluster_type)`。DDL 文件 SHALL 在同一事务内完成建表、`id_generator` 插入种子记录、`hcm_version` 视图更新。`pkg/dal/table/table.go` SHALL 新增 `LoadBalancerExclusiveClusterTable` 表名枚举并在 `TableMap` 注册 `{EnableTenant: true}`。

#### Scenario: 迁移脚本执行成功
- **WHEN** 执行 DDL 迁移脚本
- **THEN** 创建 `load_balancer_exclusive_cluster` 表，主键为 `id`，唯一键 `(cloud_id, account_id, tenant_id)` 生效，`id_generator` 表已插入该表的种子记录（`max_id=0`）

#### Scenario: bk_biz_id 默认值
- **WHEN** 未显式指定 `bk_biz_id` 插入一行原始 SQL
- **THEN** 该列默认值为 `-1`（未分配）

#### Scenario: max_conn 允许为空
- **WHEN** 插入一行不提供 `max_conn` 的数据
- **THEN** 该列存储为 SQL `NULL`，不归一化为 `0`

### Requirement: 集群相关枚举与常量注册

`pkg/criteria/enumor/load_balancer.go` SHALL 新增 `ClusterType`（取值 `TGW`、`STGW`、`VPCGW`）与 `ClusterNetwork`（取值 `Public`、`Private`、`Hybrid`）两个枚举类型，均提供 `Validate()` 方法。`pkg/criteria/enumor/cloud_resource_type.go` SHALL 新增 `LoadBalancerExclusiveClusterCloudResType` 并在 `typeMapping` 中映射到 `table.LoadBalancerExclusiveClusterTable`。`pkg/criteria/enumor/audit.go` SHALL 新增 `LoadBalancerExclusiveClusterAuditResType` 常量并注册进 `AuditResourceTypeEnums`。

#### Scenario: ClusterType 校验合法值
- **WHEN** 调用 `ClusterType("TGW").Validate()`
- **THEN** 返回 nil

#### Scenario: ClusterType 校验非法值
- **WHEN** 调用 `ClusterType("XXX").Validate()`
- **THEN** 返回非 nil 错误

#### Scenario: ClusterNetwork 校验合法值
- **WHEN** 调用 `ClusterNetwork("Public").Validate()`
- **THEN** 返回 nil

#### Scenario: 云资源类型映射
- **WHEN** 通过 `LoadBalancerExclusiveClusterCloudResType` 查表名映射
- **THEN** 返回 `load_balancer_exclusive_cluster`

### Requirement: 独占集群 Table 结构体与校验

`pkg/dal/table/cloud/load-balancer/exclusive_cluster.go` SHALL 定义 `Table` 结构体（字段与 DDL 一一对应，`Extension` 使用 `types.JsonField`）、`Columns`/`ColumnDescriptor` 列描述符，并实现 `TableName()`、`InsertValidate()`、`UpdateValidate()`。`InsertValidate()` SHALL 校验必填字段（`cloud_id`、`name`、`vendor`、`account_id`、`region`、`cluster_type`）非空、`id`/`created_at`/`updated_at` 不可预先设置、`cluster_type`/`network`/`isp` 枚举字段通过 `Validate()`。`UpdateValidate()` SHALL 禁止更新 `created_at`、`creator`。

#### Scenario: InsertValidate 缺少必填字段
- **WHEN** 调用 `InsertValidate()`，`cloud_id` 为空字符串
- **THEN** 返回错误 "cloud_id is required"

#### Scenario: InsertValidate cluster_type 非法
- **WHEN** 调用 `InsertValidate()`，`cluster_type` 为不在枚举范围内的值
- **THEN** 返回 `ClusterType.Validate()` 产生的错误

#### Scenario: InsertValidate 已设置 id
- **WHEN** 调用 `InsertValidate()`，`id` 字段非空
- **THEN** 返回错误 "id can not set"

#### Scenario: UpdateValidate 禁止修改 created_at
- **WHEN** 调用 `UpdateValidate()`，`created_at` 字段非空
- **THEN** 返回错误

### Requirement: 独占集群 DAO 接口

`pkg/dal/dao/cloud/load-balancer/exclusive_cluster.go` SHALL 定义 `ExclusiveCluster` interface 与 `ExclusiveClusterDao` 实现，方法包括：`List(kt, opt *types.ListOption) (*types.ListLoadBalancerExclusiveClusterDetails, error)`、`BatchCreateWithTx(kt, tx, models []Table) ([]string, error)`、`BatchUpdateWithTx(kt, tx, models []Table) error`、`BatchDeleteWithTx(kt, tx, expr *filter.Expression) error`。`BatchUpdateWithTx` SHALL 是唯一的更新方法，逐条对 `models` 中每个元素执行 `UpdateValidate()` 后通过 `utils.RearrangeSQLDataWithOption` 生成动态 SET 表达式（对齐 `pkg/dal/dao/cloud/load-balancer/load_balancer.go` 中 `LoadBalancerDao.Update` 与 `account-secret`/`sub-account-secret` 的 `BatchUpdate` 写法），只有非零值字段会被拼进 SQL 的 SET 子句；`bk_biz_id` 更新（分配业务）与云属性更新（同步）均调用本方法，通过调用方构造的 model 字段差异实现隔离，不额外定义 `BatchUpdateBizIDWithTx`。`pkg/dal/dao/dao.go` SHALL 注册该 DAO 到全局 `Set` 接口。

#### Scenario: BatchCreateWithTx 固定 bk_biz_id
- **WHEN** 调用 `BatchCreateWithTx` 创建 3 条集群记录
- **THEN** 返回 3 个新生成的 ID，且写入 DB 后每条记录的 `bk_biz_id` 均为 `constant.UnassignedBiz`（-1），并在同一事务内写入 Create 审计（`ResType=LoadBalancerExclusiveClusterAuditResType`）

#### Scenario: BatchCreateWithTx 超过批量上限
- **WHEN** 调用 `BatchCreateWithTx`，`models` 长度超过 `constant.BatchOperationMaxLimit`（100）
- **THEN** 返回 `errf.InvalidParameter` 错误，不写入任何记录

#### Scenario: BatchCreateWithTx 校验失败
- **WHEN** 调用 `BatchCreateWithTx`，某条 model 的 `InsertValidate()` 失败（如 `cluster_type` 非法）
- **THEN** 返回校验错误，不写入任何记录

#### Scenario: BatchUpdateWithTx 更新云属性不改变 bk_biz_id
- **WHEN** 已存在记录 `bk_biz_id=213`，调用 `BatchUpdateWithTx`，model 只赋值 `ID` 与云属性字段（如 `egress`、`clb_resource_count`），不赋值 `BkBizID`（保持零值）
- **THEN** 云属性更新成功，`bk_biz_id` 仍为 `213`（`BkBizID` 字段因零值被 `isBlank()` 跳过，未被拼进 SET 子句）

#### Scenario: BatchUpdateWithTx 分配业务
- **WHEN** 已存在记录 `bk_biz_id=-1`，调用 `BatchUpdateWithTx`，model 只赋值 `ID` 与 `BkBizID=213`（其余字段保持零值）
- **THEN** 该记录 `bk_biz_id` 更新为 `213`，其余列因零值被跳过而不受影响

#### Scenario: BatchDeleteWithTx 要求 filter
- **WHEN** 调用 `BatchDeleteWithTx`，`expr` 为 nil
- **THEN** 返回 `errf.InvalidParameter` 错误

#### Scenario: List 支持 filter 与分页
- **WHEN** 调用 `List`，`opt.Filter` 指定 `cluster_type=TGW AND region=ap-guangzhou`，`opt.Page.Count=false`
- **THEN** 返回 `ListLoadBalancerExclusiveClusterDetails{Details: [...]}`，仅包含满足条件的记录，`extension` 字段完整返回

#### Scenario: List Count 模式
- **WHEN** 调用 `List`，`opt.Page.Count=true`
- **THEN** 返回 `ListLoadBalancerExclusiveClusterDetails{Count: N}`，`Details` 为空

### Requirement: DAO List 结果类型

`pkg/dal/dao/types/load_balancer_exclusive_cluster.go` SHALL 定义 `ListLoadBalancerExclusiveClusterDetails` 结构体，含 `Count uint64` 与 `Details []tableexclusivecluster.Table`。

#### Scenario: 结果类型序列化
- **WHEN** 将 `ListLoadBalancerExclusiveClusterDetails` 序列化为 JSON
- **THEN** 输出含 `count` 和 `details` 两个字段（`omitempty`）

### Requirement: 独占集群核心模型与 Extension 强类型

`pkg/api/core/cloud/load-balancer/exclusive_cluster.go` SHALL 定义 `BaseExclusiveCluster` 核心模型（对应表结构体除 `extension` 外的所有字段）与 `TCloudExclusiveClusterExtension` 结构体，后者包含 `max_in_flow`、`max_out_flow`、`max_in_pkg`、`max_out_pkg`、`max_new_conn`、`http_max_new_conn`、`https_max_new_conn`、`http_qps`、`https_qps`、`load_balance_director_count`（均 `int64`）、`clusters_version`、`disaster_recovery_type`（均 `string`）、`clusters_zone`（嵌套对象，含 `master_zone`/`slave_zone` 两个 `string` 数组）。

#### Scenario: Extension 序列化
- **WHEN** data-service handler 将 `TCloudExclusiveClusterExtension` 序列化为 JSON 字符串
- **THEN** 生成的 JSON 含全部 12 个标量字段及 `clusters_zone` 嵌套对象，可直接写入 `types.JsonField`

#### Scenario: Extension 反序列化
- **WHEN** data-service handler 从 DB 读取的 `types.JsonField` 反序列化为 `TCloudExclusiveClusterExtension`
- **THEN** 所有字段正确还原，未出现在 JSON 中的字段保持零值

### Requirement: data-service 列表查询接口

系统 SHALL 提供 `POST /load_balancer_exclusive_clusters/list` 接口，支持 `filter` + `page` 分页查询，支持 `count` 模式，返回记录含完整 `extension`。单页 `limit` 上限 500。

#### Scenario: 分页查询
- **WHEN** 发送 POST 请求，`filter` 指定 `region=ap-guangzhou AND cluster_type=TGW`，`page.count=false`
- **THEN** 返回匹配记录列表，每条记录的 `extension` 完整返回 12 个子字段

#### Scenario: fields 未包含 extension
- **WHEN** 发送 POST 请求，`fields` 未包含 `extension`
- **THEN** 正常返回记录列表，每条记录的 `extension` 为 `{}`

#### Scenario: Count 模式
- **WHEN** 发送 POST 请求，`page.count=true`
- **THEN** 返回 `{count: N}`，`details` 为空数组

#### Scenario: limit 超过上限
- **WHEN** 发送 POST 请求，`page.limit` 超过 500
- **THEN** 返回参数校验错误

### Requirement: data-service 按 vendor 批量创建接口

系统 SHALL 提供 `POST /vendors/{vendor}/load_balancer_exclusive_clusters/batch/create` 接口，请求体为待创建集群数组（不含 `bk_biz_id` 字段），单批次上限 `constant.BatchOperationMaxLimit`（100）。创建时 `bk_biz_id` 由服务端固定写入 `-1`。

#### Scenario: 批量创建成功
- **WHEN** 发送 POST 请求，`vendor=tcloud`，包含 3 条合法集群数据
- **THEN** 返回 3 个新 ID，DB 中对应记录 `bk_biz_id` 均为 `-1`

#### Scenario: 超过批量上限
- **WHEN** 发送 POST 请求，集群数组长度为 101
- **THEN** 返回 `InvalidParameter`，不写入任何记录

### Requirement: data-service 按 vendor 批量更新云属性接口

系统 SHALL 提供 `PATCH /vendors/{vendor}/load_balancer_exclusive_clusters` 接口，请求体结构体 `BatchUpdateReq` SHALL NOT 包含 `bk_biz_id` 字段，用于同步逻辑更新集群云属性（如 `egress`、`max_conn`、`clb_resource_count`、`extension`）。handler 内部构造的 `Table` model SHALL NOT 赋值 `BkBizID` 字段，调用 DAO 层统一的 `BatchUpdateWithTx`。

#### Scenario: 更新云属性不影响 bk_biz_id
- **WHEN** 已存在记录 `bk_biz_id=213`，发送 PATCH 请求更新该记录的 `clb_resource_count`
- **THEN** `clb_resource_count` 更新成功，`bk_biz_id` 仍为 `213`（因 handler 未对 model 的 `BkBizID` 赋值，DAO 层零值跳过）

### Requirement: data-service 批量更新 bk_biz_id 接口（分配业务）

系统 SHALL 提供 `PATCH /load_balancer_exclusive_clusters/biz` 接口（无 vendor 路径参数），请求体 `BatchUpdateBizIDReq` 仅含 `cluster_ids`（本地 ID 数组，最多 100 个）与 `bk_biz_id`（必须大于 0），用于把集群分配给业务。handler 内部为每个 `cluster_id` 构造一个只赋值 `ID`/`BkBizID`/`Reviser` 的 `Table` model（其余字段保持零值），调用与云属性更新**同一个** DAO 层 `BatchUpdateWithTx`，不额外定义专属的分配业务 DAO 方法。本接口 SHALL NOT 做"是否已分配"业务前置校验、鉴权与审计（由上层 cloud-server 负责），仅保证按传入参数正确写库。

#### Scenario: 分配业务写库成功
- **WHEN** 发送 PATCH 请求，`cluster_ids=["00000001"]`，`bk_biz_id=213`
- **THEN** 对应记录的 `bk_biz_id` 更新为 `213`，其余列不变

#### Scenario: cluster_ids 超过上限
- **WHEN** 发送 PATCH 请求，`cluster_ids` 长度超过 100
- **THEN** 返回参数校验错误

### Requirement: data-service 批量删除接口

系统 SHALL 提供 `DELETE /load_balancer_exclusive_clusters/batch` 接口，通过 `filter.Expression` 指定删除条件，用于同步逻辑删除云上已不存在的集群记录。

#### Scenario: 按 filter 删除
- **WHEN** 发送 DELETE 请求，`filter` 指定 `id in [...]`
- **THEN** 系统物理删除匹配记录

#### Scenario: filter 为空
- **WHEN** 发送 DELETE 请求，`filter` 为 nil
- **THEN** 返回 `InvalidParameter` 错误

### Requirement: 创建审计记录

系统 SHALL 为批量创建操作记录审计，在 DAO `BatchCreateWithTx` 内的同一事务中写入，`ResType=enumor.LoadBalancerExclusiveClusterAuditResType`，`Action=Create`。批量更新云属性、更新 `bk_biz_id`、删除操作的审计 build 方法与触发点不在本能力范围内。

#### Scenario: 创建审计写入
- **WHEN** `BatchCreateWithTx` 成功创建记录
- **THEN** 同一事务内写入审计记录，包含 `ResID`、`ResType=LoadBalancerExclusiveClusterAuditResType`、`Action=Create`、`Vendor`、`Operator`

### Requirement: Client 封装

`pkg/client/data-service/global/load_balancer_exclusive_cluster.go` SHALL 提供 Global client 方法：`List`、`BatchDelete`、`BatchUpdateBizID`；vendor-scoped 方法（`BatchCreate`、`BatchUpdate`）SHALL 按 vendor 路径封装。

#### Scenario: Global client 调用 List
- **WHEN** 调用 `global.LoadBalancerExclusiveCluster.List(kt, req)`
- **THEN** 发送 POST 到 `/load_balancer_exclusive_clusters/list`，返回 `ListResult`

#### Scenario: Global client 调用 BatchUpdateBizID
- **WHEN** 调用 `global.LoadBalancerExclusiveCluster.BatchUpdateBizID(kt, req)`
- **THEN** 发送 PATCH 到 `/load_balancer_exclusive_clusters/biz`

