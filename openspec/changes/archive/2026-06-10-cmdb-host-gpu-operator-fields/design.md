## Context

HCM 在 CVM 创建、云资源同步、分配业务、CC 信息回填等场景下，通过 data-service 的 `upsertCmdbHosts` / `upsertBaseCmdbHosts` / `deleteCmdbHosts` 将主机同步到蓝鲸 CMDB（最终 `POST /createmany/cloud_hosts`，upsert 语义）。当前 `HostCreateParam` 仅含基础字段。

关键现状约束：

- 购买机与云上增量机都经 res-sync 进资源池（`bk_biz_id=-1`、`bk_host_id=UnBindBkHostID=-1`），**创建阶段因 `UnassignedBiz` 被跳过、不推 CMDB**；首次推 CMDB 都发生在"分配业务"时（`SyncCvmToCmdb` 或经 `BatchUpdateCvmCommonInfo` 的 `upsertBaseCmdbHosts`）。
- 后台同步任务用户为 `constant.BackendOperationUserKey`（`hcm-backend-admin`）；云上发现的增量机 `creator` 即此值，HCM 购买机 `creator` 为购买人。
- `machine_type` 列即云上实例规格 ID（如 `S5.LARGE8`/`t2.micro`/`g4dn.xlarge`），是 GPU 判断输入。
- CMDB 返回的 `bk_host_id` 会回写 CVM，故"首次推送"可由 `bk_host_id <= 0` 判定。

## Goals / Non-Goals

**Goals:**

- 在所有 CMDB 同步路径下发 `is_gpu`、`on_shelf_date`。
- 仅在主机首次入 CMDB 时下发 `operator`，并按 creator/账号负责人正确推导；后续同步以 CMDB 为准不覆盖。
- GPU 识别基于可运维配置（`global_config`）的厂商机型清单，落 CVM `is_gpu` 列。

**Non-Goals:**

- 不实现早期方案的 `gpu_count`/`gpu_name`/`gpu_memory`（基于 instance-type 表）字段。
- 不实现存量主机批量补刷（可后续单独需求）。
- 不在 CMDB 同步时回调云 API；不负责 CMDB 侧字段建模。

## Decisions

### D1: `is_gpu` 落 CVM 列，在 data-service 计算

在 data-service 创建/更新流程计算 `is_gpu`（拥有 `machine_type`+`vendor`，并可直接访问 `dao.GlobalConfig()`），写入新增的 `cvm.is_gpu` 列。CMDB 同步层只读取该列下发。

- 备选：在 hc-service res-sync 计算后随 create/update 请求带入。否决：会分散到 5 个厂商分包、且与 data-service 的更新路径（升降配重算）重复。

### D2: GPU 机型清单用 `global_config`，类型 `gpu_machine_type`

`config_type=gpu_machine_type`，`config_key=<vendor>`，`config_value=["机型",...]`。新增 `enumor.GlobalConfigTypeGPUMachineType`。识别函数读取后做集合判断，加轻量 TTL 缓存（`sync.Map`+过期）避免逐台查库。Azure 直接返回 false。

- 备选：复用早期方案查 instance-type 资源表的 GPU 字段。否决：当前需求明确改为按配置机型清单，简单可控、便于运维调整。

### D3: Operator 只在"新增到 CMDB"时下发，用 bk_host_id 区分新增/修改

CMDB 的 `createmany/cloud_hosts` 是 upsert 接口，新增与修改共用，因此 `operator` 必须只在"新增"时下发、"修改"时绝不下发（以 CMDB 侧为准）。

关键：这里的"新增 vs 修改"指的是**该主机是否已存在于 CMDB**，而**不是 HCM 侧的"创建调用 vs 更新调用"**。因为购买机与云上增量机都先经 res-sync 进资源池（创建调用阶段因 `UnassignedBiz` 被跳过、不推 CMDB），真正"第一次进 CMDB"都发生在**分配业务时**，而分配业务走的是更新类调用（`SyncCvmToCmdb` / 经 `BatchUpdateCvmCommonInfo` 的 `upsertBaseCmdbHosts`）。所以用"调用方是 create 还是 update"作判据是错的，会漏掉真正的首推时机。

正确判据：`bk_host_id <= 0` 表示该主机尚未绑定 CMDB 主机 ID（= 对 CMDB 而言是新增）；CMDB 创建成功后会把真实 `bk_host_id` 回写 CVM，之后即视为"修改"。

推导所需的 `creator`、`account_id`、`bk_host_id` 在 CVM 记录上均可得，无需新增 `operator` 列。在 `upsertCmdbHosts`/`upsertBaseCmdbHosts` 中构建 `Operators map[cloudID]string`：

1. 仅对 `bk_host_id <= 0`（新增）的主机推导 operator。
2. `creator == hcm-backend-admin`（云上增量机）→ 取所属 `Account.Managers`（批量查 `dao.Account()`，逗号连接）；否则（HCM 购买机）取 `creator`。
3. 其余主机（`bk_host_id > 0`，即修改）不入 map → 经 `Operator,omitempty` 不下发，CMDB 侧 operator 保持不变。

将该 map 经 logics 请求传入，在 `AddCloudHostToBiz`/`AddBaseCloudHostToBiz` 映射处写 `HostCreateParam.Operator`。

- 备选 A：新增 `operator` 列在创建时落库。否决：首次推送可实时推导，避免冗余列与回填购买流程改动。
- 备选 B：上层调用显式传 `isCreate` 标记（create.go=true，其余=false）。否决：与 CMDB 的新增/修改语义不一致——首推发生在分配业务（更新类调用）时，该标记会把首推误判为"修改"而漏填 operator。

### D4: `on_shelf_date` 复用 `cloud_created_time`，AWS 回退 `cloud_launched_time`

不新增列。映射在 logics 层完成：`OnShelfDate = host.CloudCreatedTime`，AWS 且为空时用 `host.CloudLaunchedTime`，均空则不下发（`omitempty`）。

### D5: 字段填充集中在 logics + data-service cmdb.go

`HostCreateParam` 三字段映射放在 [logics/cmdb/cmdb.go](cmd/data-service/service/cloud/logics/cmdb/cmdb.go) 的 `AddCloudHostToBiz`/`AddBaseCloudHostToBiz`；Operators 推导（含账号负责人查询、首推判定）放在 [cvm/cmdb.go](cmd/data-service/service/cloud/cvm/cmdb.go) 的两个 upsert 函数。`is_gpu` 经 `BaseCvm` 字段贯通（`convTableToBaseCvm` 补充）。

## Risks / Trade-offs

- [新增/修改判据 `bk_host_id <= 0` 误判] → 依赖资源池机器 `bk_host_id=-1`、入 CMDB 后回写真实 id 的既有机制；存量已在 CMDB 的机器（id>0）天然不再下发 operator，符合"修改时以 CMDB 为准"。极端情况：若历史数据中某主机已在 CMDB 但 `bk_host_id` 未回写（同步失败/数据不一致），会被判为新增而下发一次 operator 覆盖 CMDB 值。属低概率边界，可接受；必要时后续以"是否已成功同步过 CMDB"的标记位增强。
- [`creator == hcm-backend-admin` 区分云上增量机] → 手动触发的全量同步拉入的存量机器，`creator` 会是触发人而非后台用户，可能被当作"购买人"。属已知边界，按确认判据执行。
- [升降配机型未回写] → res-sync 更新路径默认不传 `machine_type`（zero 值跳过），`is_gpu` 重算以收到的 `MachineType` 为准；若机型未随更新下传则 `is_gpu` 不变。
- [GPU 清单未录入] → `is_gpu` 全为 false；需运维先通过 global-config API 录入各厂商机型。
- [CMDB 侧字段未就绪] → 下发字段被忽略或报错；沿用 `CmdbSyncFailed` 容错不阻塞主流程。

## Migration Plan

1. 执行 DB 迁移新增 `cvm.is_gpu` 列（默认 0）。
2. 部署新版 data-service（含 GPU 识别与 CMDB 字段下发）。
3. 通过 global-config API 录入各厂商 `gpu_machine_type` 机型清单。
4. 后续主机创建/同步/分配自动带新字段；存量数据随下次同步逐步生效。
5. 回滚：下线新版本即可，新增列默认 0、CMDB 下发字段 `omitempty` 不影响旧逻辑。

## Open Questions

- CMDB 云主机模型 `is_gpu`/`operator`/`on_shelf_date` 字段名与类型最终以 CMDB 团队确认为准（当前 `HostCreateParam` 已按 `is_gpu`/`operator`/`on_shelf_date` 命名）。
- 是否需要存量 GPU 主机/Operator 的批量补刷能力（本次 Non-Goal，待评估）。
