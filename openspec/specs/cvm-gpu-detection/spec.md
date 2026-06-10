# cvm-gpu-detection Specification

## Purpose
TBD - created by archiving change cmdb-host-gpu-operator-fields. Update Purpose after archive.
## Requirements
### Requirement: CVM 持久化 is_gpu 标识

系统 SHALL 在 CVM 资源记录上持久化 `is_gpu` 布尔标识，表示该主机是否为 GPU 机器。该标识 MUST 在主机创建入库时计算并写入，并在主机机型（`machine_type`）发生变更时重新计算。

#### Scenario: 创建 GPU 机型主机

- **WHEN** 创建/同步入库一台 vendor 为 TCloud/AWS/GCP/HuaWei、且 `machine_type` 命中该厂商 GPU 机型清单的主机
- **THEN** 该 CVM 记录的 `is_gpu` 写为 `true`

#### Scenario: 创建非 GPU 机型主机

- **WHEN** 创建/同步入库一台 `machine_type` 未命中 GPU 机型清单的主机
- **THEN** 该 CVM 记录的 `is_gpu` 写为 `false`

#### Scenario: 机型升降配后重算

- **WHEN** 更新主机时传入了新的 `machine_type`
- **THEN** 系统 SHALL 依据新机型重新计算并更新 `is_gpu`

### Requirement: 基于 global_config 的厂商 GPU 机型清单

系统 SHALL 通过 `global_config` 表读取各云厂商的 GPU 机型清单进行匹配。配置 MUST 使用统一类型 `config_type=gpu_machine_type`，以厂商标识为 `config_key`（tcloud/aws/gcp/huawei），以机型字符串数组为 `config_value`。匹配 SHALL 为精确集合判断（`machine_type` 是否在清单中）。

#### Scenario: 命中配置清单

- **WHEN** 主机 vendor=aws、`machine_type=g4dn.xlarge` 且 aws 的 GPU 机型清单包含 `g4dn.xlarge`
- **THEN** 识别结果为 GPU 机器（`is_gpu=true`）

#### Scenario: 厂商未配置或清单为空

- **WHEN** 某厂商在 `global_config` 中没有 `gpu_machine_type` 配置或清单为空
- **THEN** 该厂商所有主机 `is_gpu` 识别为 `false`，且不阻塞主机创建/同步流程

#### Scenario: Azure 不做 GPU 识别

- **WHEN** 主机 vendor=azure
- **THEN** 系统 SHALL 直接将 `is_gpu` 置为 `false`，不查询机型清单

