## Why

下游 GPU 资源分析平台需要从 CMDB 获取主机是否为 GPU 机器、上架时间与运维负责人（Operator），但当前 HCM 同步到 CMDB 的 `HostCreateParam` 仅含基础字段，缺少 `is_gpu`、`operator`、`on_shelf_date`。需要在主机创建与同步到 CMDB 时补齐这三个字段，并按业务规则正确推导 Operator（以 CMDB 侧为准、避免覆盖）。

## What Changes

- CVM 资源新增 `is_gpu` 持久化列：对 GCP/TCloud/AWS/HuaWei 通过机型（`machine_type`）判断是否 GPU 机器，机型清单存于 `global_config`（按厂商 key、机型数组 value）；Azure 不做识别，恒为 `false`。
- 创建主机与云上同步入库时计算并落库 `is_gpu`；机型变更（升降配）时重算。
- 主机同步到 CMDB 时新增下发三字段：
  - `is_gpu`：所有同步路径都下发。
  - `on_shelf_date`：取购买时间 `cloud_created_time`，AWS 回退用 `cloud_launched_time`；所有同步路径都下发。
  - `operator`：仅在主机**首次入 CMDB**（`bk_host_id <= 0`）时下发；`creator` 为后台用户（`hcm-backend-admin`，即云上发现的增量机）时取所属二级账号负责人（`Account.Managers`），否则取 `creator`（HCM 购买人）。后续任何同步均不下发 `operator`，以 CMDB 侧为准。
- 复用现有 CMDB 同步容错机制（`CmdbSyncFailed`），同步失败不阻塞 CVM 主流程。

## Capabilities

### New Capabilities

- `cvm-gpu-detection`: 基于 `global_config` 配置的各厂商 GPU 机型清单，对 CVM 进行 GPU 识别并将结果持久化到 CVM 的 `is_gpu` 列，覆盖创建与机型变更场景。
- `cmdb-host-field-sync`: 在主机同步到 CMDB 时填充并下发 `is_gpu`、`on_shelf_date`、`operator`，其中 Operator 遵循"首次入 CMDB 才下发、按 creator/账号负责人推导、后续不覆盖"的规则。

### Modified Capabilities

<!-- 无现有 spec 涉及主机 CMDB 同步，全部为新增能力 -->

## Impact

- 数据库：`cvm` 表新增 `is_gpu` 列（新增迁移脚本 `scripts/sql/`）。
- 配置：`global_config` 新增 `config_type=gpu_machine_type` 类型，按厂商录入 GPU 机型数组。
- 代码：
  - `pkg/dal/table/cloud/cvm/cvm.go`、`pkg/api/core/cloud/cvm/cvm.go`、`cmd/data-service/service/cloud/cvm/query.go`（新增 `is_gpu` 字段贯通）。
  - `pkg/criteria/enumor/global_config.go`（新增配置类型枚举）。
  - `cmd/data-service/service/cloud/cvm/`（新增 GPU 识别逻辑，create/update 集成）。
  - `cmd/data-service/service/cloud/cvm/cmdb.go`、`cmd/data-service/service/cloud/logics/cmdb/`（CMDB 同步层填充三字段、推导 Operator）。
  - `pkg/thirdparty/api-gateway/cmdb/types.go`（`HostCreateParam` 三字段已就绪，`Host` 读结构按需对齐）。
- 外部依赖：蓝鲸 CMDB 云主机模型需具备 `is_gpu`/`operator`/`on_shelf_date` 对应字段。
