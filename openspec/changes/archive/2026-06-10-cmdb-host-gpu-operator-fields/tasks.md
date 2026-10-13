## 1. CVM is_gpu 字段贯通

- [x] 1.1 新增迁移脚本 `scripts/sql/0053_20260609_cvm_is_gpu.sql`：`ALTER TABLE cvm ADD COLUMN is_gpu tinyint(1) NOT NULL DEFAULT 0`
- [x] 1.2 `pkg/dal/table/cloud/cvm/cvm.go`：`TableColumnDescriptor` 增加 `is_gpu` 列，`Table` 结构增加 `IsGPU bool` 字段
- [x] 1.3 `pkg/api/core/cloud/cvm/cvm.go`：`BaseCvm` 增加 `IsGPU bool` 字段
- [x] 1.4 `cmd/data-service/service/cloud/cvm/query.go` `convTableToBaseCvm`：补充 `IsGPU: one.IsGPU`

## 2. GPU 机型配置与识别

- [x] 2.1 `pkg/criteria/enumor/global_config.go`：新增 `GlobalConfigTypeGPUMachineType = "gpu_machine_type"`
- [x] 2.2 新增 `cmd/data-service/service/cloud/cvm/gpu.go`：实现 `isGPUMachine(kt, vendor, machineType)`，经 `dao.GlobalConfig().List` 读取机型清单做集合判断，加 TTL 缓存；Azure 直接返回 false
- [x] 2.3 `cmd/data-service/service/cloud/cvm/create.go` `batchCreateCvm`：入库前对每台计算并设置 `model.IsGPU`
- [x] 2.4 `cmd/data-service/service/cloud/cvm/update.go` `BatchUpdateCvm`：当 `update.MachineType != ""` 时重算并写入 `IsGPU`

## 3. CMDB 同步层填充三字段

- [x] 3.1 `cmd/data-service/service/cloud/logics/cmdb/types.go`：`AddCloudHostToBizReq[T]` 与 `AddBaseCloudHostToBizReq` 各增加 `Operators map[string]string`（key=CloudID）
- [x] 3.2 `cmd/data-service/service/cloud/logics/cmdb/cmdb.go` `AddCloudHostToBiz`/`AddBaseCloudHostToBiz`：`HostCreateParam` 填充 `IsGPU`、`OnShelfDate`（AWS 空时回退 `CloudLaunchedTime`）、`Operator`（取自 `req.Operators[host.CloudID]`）
- [x] 3.3 `cmd/data-service/service/cloud/cvm/cmdb.go` `upsertCmdbHosts`/`upsertBaseCmdbHosts`：构建 `Operators`——对 `bk_host_id <= 0` 的主机推导（`creator==BackendOperationUserKey` 则批量查 `dao.Account()` 取 `Managers` 逗号连接，否则取 `creator`），传入 logics 请求

## 4. CMDB 读结构对齐（可选）

- [x] 4.1 `pkg/thirdparty/api-gateway/cmdb/types.go`：`Host` 结构补充 `is_gpu`/`on_shelf_date`（`operator` 已存在）。`HostFields` 暂不改动，待 CMDB 侧确认字段可查询后再加入

## 5. 验证

- [x] 5.1 单测：`isGPUMachine` Azure/空机型分支、`gpuTypeCache` 命中/未命中/过期（命中机型清单依赖 DB，联调覆盖）
- [x] 5.2 单测：`buildCmdbOperators` 首次推送下发 operator（购买人）、存量(bk_host_id>0)/无 creator/资源池/Other 均不下发；账号负责人分支依赖 DB，联调覆盖
- [ ] 5.3 联调：is_gpu 与 on_shelf_date（含 AWS 回退）在创建/分配/更新路径正确同步到 CMDB
- [ ] 5.4 录入各厂商 `gpu_machine_type` 配置后端到端验证 GPU 识别与推送
