## ADDED Requirements

### Requirement: 同步 is_gpu 到 CMDB

系统在将主机同步到 CMDB 时 SHALL 下发 `is_gpu` 字段，取值为 CVM 记录的 `is_gpu`。该字段 MUST 在所有 CMDB 同步路径（创建、更新、分配业务、CC 信息回填）均下发。

#### Scenario: GPU 主机同步到 CMDB

- **WHEN** 一台 `is_gpu=true` 的主机触发任意 CMDB 同步
- **THEN** CMDB 主机记录的 `is_gpu` 为 `true`

#### Scenario: 非 GPU 主机同步到 CMDB

- **WHEN** 一台 `is_gpu=false` 的主机触发任意 CMDB 同步
- **THEN** CMDB 主机记录的 `is_gpu` 为 `false`

### Requirement: 同步 on_shelf_date 到 CMDB

系统在将主机同步到 CMDB 时 SHALL 下发 `on_shelf_date`（上架时间），取主机购买时间 `cloud_created_time`。当 vendor 为 AWS 且 `cloud_created_time` 为空时，系统 SHALL 回退使用 `cloud_launched_time`。该字段在所有 CMDB 同步路径均下发。

#### Scenario: 普通厂商取购买时间

- **WHEN** 一台 TCloud/HuaWei/GCP/Azure 主机同步到 CMDB 且 `cloud_created_time` 非空
- **THEN** CMDB 主机记录的 `on_shelf_date` 等于 `cloud_created_time`

#### Scenario: AWS 回退启动时间

- **WHEN** 一台 AWS 主机同步到 CMDB，其 `cloud_created_time` 为空、`cloud_launched_time` 非空
- **THEN** CMDB 主机记录的 `on_shelf_date` 等于 `cloud_launched_time`

#### Scenario: 时间均为空

- **WHEN** 主机 `cloud_created_time` 与（AWS）`cloud_launched_time` 均为空
- **THEN** 不下发 `on_shelf_date`（留空，不覆盖 CMDB 侧已有值）

### Requirement: 首次入 CMDB 时推导并下发 Operator

系统 SHALL 仅在主机**首次被创建到 CMDB**时下发 `operator`，判定条件为该 CVM 的 `bk_host_id <= 0`（尚未绑定 CMDB 主机 ID）。Operator 取值规则：当 CVM 的 `creator` 等于后台同步用户（`hcm-backend-admin`）时，取该主机所属二级账号（`Account`）的负责人 `Managers`（多个以逗号连接）；否则取 `creator`（HCM 购买人）。主机已存在于 CMDB（`bk_host_id > 0`）时的任何同步均 MUST NOT 下发 `operator`，以 CMDB 侧值为准。

#### Scenario: HCM 购买机首次入 CMDB

- **WHEN** 一台 `bk_host_id <= 0`、`creator` 为真实用户（非后台用户）的主机首次被分配业务并同步到 CMDB
- **THEN** CMDB 主机记录的 `operator` 等于该 CVM 的 `creator`

#### Scenario: 云上增量机首次入 CMDB

- **WHEN** 一台 `bk_host_id <= 0`、`creator` 等于 `hcm-backend-admin` 的主机首次被分配业务并同步到 CMDB
- **THEN** CMDB 主机记录的 `operator` 等于该主机所属二级账号的 `Managers`（逗号连接）

#### Scenario: 存量主机更新同步不覆盖 Operator

- **WHEN** 一台 `bk_host_id > 0`（已在 CMDB）的主机因云上同步、信息更新或再次分配而触发 CMDB 同步
- **THEN** 同步请求 MUST NOT 包含 `operator` 字段，CMDB 侧 `operator` 保持不变

### Requirement: 沿用未分配业务与容错规则

系统 SHALL 维持现有 CMDB 同步边界与容错：`bk_biz_id` 未分配（`-1`）或 `vendor=Other` 的主机不同步到 CMDB；CMDB 同步失败按现有 `CmdbSyncFailed` 机制记录日志，不阻塞 CVM 主流程。

#### Scenario: 未分配业务主机不同步

- **WHEN** 一台 `bk_biz_id=-1` 的主机被创建或更新
- **THEN** 不触发对该主机的 CMDB 字段下发（含 is_gpu/operator/on_shelf_date）

#### Scenario: CMDB 同步失败不阻塞

- **WHEN** 向 CMDB 下发新字段过程中 CMDB 接口返回失败
- **THEN** 记录 `CmdbSyncFailed` 错误日志，CVM 的写入/更新主流程不回滚
