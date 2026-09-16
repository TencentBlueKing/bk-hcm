## 1. SQL 迁移与表注册

- [x] 1.1 新增迁移脚本 `scripts/sql/9999_{date}_load_balancer_exclusive_cluster.sql`（文件头 `SQLVER=9999,HCMVER=v9.9.9`，发布时按规范重命名为下一个序号，当前最新为 `0053_20260609_cvm_add_gpu.sql`）。内容：`START TRANSACTION` → 建表 `load_balancer_exclusive_cluster`（字段与索引见 design.md/spec.md，唯一键 `(cloud_id, account_id, tenant_id)`）→ `INSERT INTO id_generator` 种子 → `CREATE OR REPLACE VIEW hcm_version` → `COMMIT`
- [x] 1.2 `pkg/dal/table/table.go` 新增 `LoadBalancerExclusiveClusterTable Name = "load_balancer_exclusive_cluster"`，并在 `TableMap` 注册 `{EnableTenant: true}`

## 2. 枚举与常量注册

- [x] 2.1 `pkg/criteria/enumor/load_balancer.go` 新增 `ClusterType`（`TGW`/`STGW`/`VPCGW`）枚举类型 + `Validate()`
- [x] 2.2 同文件新增 `ClusterNetwork`（`Public`/`Private`）枚举类型 + `Validate()`
- [x] 2.3 `pkg/criteria/enumor/cloud_resource_type.go` 新增 `LoadBalancerExclusiveClusterCloudResType`，并在 `typeMapping` 映射到 `table.LoadBalancerExclusiveClusterTable`
- [x] 2.4 `pkg/criteria/enumor/audit.go` 新增 `LoadBalancerExclusiveClusterAuditResType` 常量并注册进 `AuditResourceTypeEnums`

## 3. Table 层

- [x] 3.1 新增 `pkg/dal/table/cloud/load-balancer/exclusive_cluster.go`：定义 `Table` 结构体（对齐 DDL 全部字段，`Extension` 用 `types.JsonField`，`MaxConn` 用可空类型）、`Columns`/`ColumnDescriptor`
- [x] 3.2 实现 `TableName()` 返回 `table.LoadBalancerExclusiveClusterTable`
- [x] 3.3 实现 `InsertValidate()`：长度校验 + 必填字段（`cloud_id`/`name`/`vendor`/`account_id`/`region`/`cluster_type`）非空 + `id`/`created_at`/`updated_at` 不可预先设置 + `cluster_type`/`network`/`isp` 枚举 `Validate()`
- [x] 3.4 实现 `UpdateValidate()`：禁止更新 `created_at`、`creator`、`updated_at`

## 4. DAO 层

- [x] 4.1 新增 `pkg/dal/dao/types/load_balancer_exclusive_cluster.go`：定义 `ListLoadBalancerExclusiveClusterDetails{Count, Details}`
- [x] 4.2 新增 `pkg/dal/dao/cloud/load-balancer/exclusive_cluster.go`：定义 `ExclusiveCluster` interface（`List`/`BatchCreateWithTx`/`BatchUpdateWithTx`/`BatchDeleteWithTx`，共 4 个方法，不含专属的 `bk_biz_id` 更新方法，对齐 `LoadBalancerDao`/`AccountSecretDao` 现有写法）与 `ExclusiveClusterDao` 结构体（`Orm`/`IDGen`/`Audit`）
- [x] 4.3 实现 `BatchCreateWithTx`：`IDGen.Batch` 生成 ID → 逐条 `InsertValidate()` → `BulkInsert`（`ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID))`）→ 同事务写入 Create 审计（`ResType=LoadBalancerExclusiveClusterAuditResType`）→ 返回 IDs；批量上限校验 `constant.BatchOperationMaxLimit`
- [x] 4.4 实现唯一的更新方法 `BatchUpdateWithTx`：逐条 `UpdateValidate()` → `utils.RearrangeSQLDataWithOption` 生成动态 SET 表达式（忽略 `types.DefaultIgnoredFields`，只有非零值字段进 SET 子句）→ `UPDATE ... WHERE id = :id`；`bk_biz_id` 更新（分配业务）与云属性更新（同步）均复用本方法，由调用方（7.4/7.5）构造不同字段的 model 实现隔离，DAO 层不感知调用意图
- [x] 4.5 实现 `BatchDeleteWithTx`：`expr` 非空校验 → `SQLWhereExpr` → `DELETE FROM ... WHERE ...`
- [x] 4.6 实现 `List`：`opt` 非空校验 → `Columns.ColumnTypes()` 校验 filter 字段 → count 模式返回 `Count`，非 count 模式 `SELECT` + `PageSQLExpr` 分页，返回 `Details`
- [x] 4.7 `pkg/dal/dao/dao.go` 的 `Set` interface 新增 `LoadBalancerExclusiveCluster() daoexclusivecluster.ExclusiveCluster` 方法，并在 `set` 实现中注入 `Orm`/`IDGen`/`Audit`

## 5. Core API 模型

- [x] 5.1 新增 `pkg/api/core/cloud/load-balancer/exclusive_cluster.go`：定义 `BaseExclusiveCluster`（对应表结构体除 `extension` 外全部字段）
- [x] 5.2 同文件定义 `TCloudExclusiveClusterExtension`（`max_in_flow`/`max_out_flow`/`max_in_pkg`/`max_out_pkg`/`max_new_conn`/`http_max_new_conn`/`https_max_new_conn`/`http_qps`/`https_qps`/`load_balance_director_count` 均 `int64`；`clusters_version`/`disaster_recovery_type` 均 `string`；`clusters_zone` 嵌套对象含 `master_zone`/`slave_zone` 两个 `string` 数组）

## 6. data-service 请求/响应体

- [x] 6.1 新增 `pkg/api/data-service/cloud/load-balancer/exclusive_cluster.go`：定义 `ListReq`（`Filter`+`Page`）与 `ListResult`（`Count`+`Details []BaseExclusiveCluster`）
- [x] 6.2 同文件定义 `BatchCreateReq`（集群数组，元素含必填字段 + `Extension`，不含 `bk_biz_id`）与 `BatchCreateResp`（`IDs []string`）
- [x] 6.3 同文件定义 `BatchUpdateReq`（更新集合，不含 `bk_biz_id` 字段）
- [x] 6.4 同文件定义 `BatchUpdateBizIDReq`（`ClusterIDs []string` 必填 `min=1,max=100`，`BkBizID int64` 必填 `gt=0`）
- [x] 6.5 同文件定义 `BatchDeleteReq`（`Filter *filter.Expression`）

## 7. data-service 路由与 handler

- [x] 7.1 新增 `cmd/data-service/service/cloud/load-balancer/exclusive_cluster.go`：`InitExclusiveClusterService(cap *rest.Capability)` 注册 5 个路由
- [x] 7.2 实现 `List` handler：`POST /load_balancer_exclusive_clusters/list`，`page.limit` 上限 500
- [x] 7.3 实现 `BatchCreate` handler：`POST /vendors/{vendor}/load_balancer_exclusive_clusters/batch/create`，固定写入 `bk_biz_id=constant.UnassignedBiz`，`Extension` 序列化为 `types.JsonField`
- [x] 7.4 实现 `BatchUpdate` handler：`PATCH /vendors/{vendor}/load_balancer_exclusive_clusters`，请求体 `BatchUpdateReq` 不含 `bk_biz_id`；构造的 `Table` model 不赋值 `BkBizID`（保持零值），调用 DAO `BatchUpdateWithTx`
- [x] 7.5 实现 `BatchUpdateBizID` handler：`PATCH /load_balancer_exclusive_clusters/biz`；为每个 `cluster_id` 构造只赋值 `ID`/`BkBizID`/`Reviser` 的 `Table` model，调用与 7.4 **同一个** DAO `BatchUpdateWithTx`（不新增专属方法）；不做业务前置校验/鉴权/审计（注释标注仅供 cloud-server 内部调用）
- [x] 7.6 实现 `BatchDelete` handler：`DELETE /load_balancer_exclusive_clusters/batch`
- [x] 7.7 `cmd/data-service/service/service.go` 的 `InitService` 中注册 `InitExclusiveClusterService`

## 8. 客户端封装

- [x] 8.1 新增 `pkg/client/data-service/global/load_balancer_exclusive_cluster.go`：`List`、`BatchUpdateBizID`、`BatchDelete` 方法
- [x] 8.2 vendor-scoped 客户端方法 `BatchCreate`、`BatchUpdate`（按 `pkg/client/data-service/tcloud` 现有封装方式，拼接 `/vendors/tcloud` 前缀）

## 9. 文档与自测

- [x] 9.1 校对 `docs/api-docs/web-server/docs/resource/load-balancer/list_exclusive_cluster.md`，删除已废弃的 `resource_count`/`idle_resource_count`/`sync_time` 查询字段说明（design.md 决策 1：本变更只落地 `clb_resource_count` 单列），保持与最终表结构一致
- [ ] 9.2 ~~新增 4 份 data-service 层接口文档到 `docs/api-docs/web-server/docs/data-service/load-balancer/`~~ —— 跳过：排查全仓库后确认不存在 `docs/api-docs/.../data-service/` 这一文档目录惯例（现有 api-docs 只覆盖 api-server/web-server/task-server/auth-server 等对外服务；data-service 是内部服务，cert、load-balancer 等既有 DAO 接口均未单独出过文档），故未新造该目录结构，仅保留/更新了 web-server 层的 `list_exclusive_cluster.md`（9.1）
- [ ] 9.3 本地起 data-service，执行迁移脚本，手工验证 AC-001~AC-007：建表与唯一键、批量创建默认 `bk_biz_id=-1`、批量更新云属性不动 `bk_biz_id`、`BatchUpdateBizID` 生效、List filter+extension 完整返回、超批量上限报错、非法枚举值报错 —— 用户自测，不在本次自动化实现范围内
