# Spec: aws-gpu-instance-data

## Purpose

提供 AWS GPU 实例数据透传能力，包括 GPU 实例类型查询和 GPU 实例列表查询接口，供下游 GPU 资源分析平台通过蓝鲸 API 网关调用。

## Requirements

### Requirement: 实例类型 GPU 字段补全
`AwsInstanceType` 数据模型 SHALL 包含以下 GPU 相关字段：`GPUMemory`（单卡显存 MiB）、`GPUName`（GPU 型号，如 T4、A100）、`GPUManufacturer`（制造商，如 NVIDIA）。转换函数 `toAwsInstanceType` SHALL 从 `DescribeInstanceTypes` 响应的 `GpuInfo` 字段解析这些值。仅涉及 AWS 云厂商。

#### Scenario: GPU 实例类型解析
- **GIVEN** AWS DescribeInstanceTypes 返回的实例类型包含 GpuInfo（如 p3.2xlarge）
- **WHEN** 系统执行 toAwsInstanceType 转换
- **THEN** GPUCount > 0，GPUMemory/GPUName/GPUManufacturer 均有值

#### Scenario: 非 GPU 实例类型解析
- **GIVEN** AWS DescribeInstanceTypes 返回的实例类型不包含 GpuInfo（如 t3.micro）
- **WHEN** 系统执行 toAwsInstanceType 转换
- **THEN** GPUCount = 0，GPUMemory/GPUName/GPUManufacturer 为零值

### Requirement: GPU 实例类型透传接口
hc-service SHALL 提供 GPU 实例类型查询接口，接收 `root_account_id`（根账号 HCM 内部 ID）、`main_account_id`（二级账号 HCM 内部 ID）、`role_chain`（角色名数组，支持 Role Chaining）、`region` 参数。系统通过 `root_account_id` 调 `AwsRoot()` 获取根账号 AK/SK，通过 `main_account_id` 查 `main_account` 表获取目标成员账号 CloudID，再按 role_chain 顺序执行链式 AssumeRole 获取成员账号权限后，调用 `DescribeInstanceTypes` 并返回含 GPU 字段的实例类型列表。参考 GCP GPU monitoring 的入参模式。该接口为 AWS 数据透传，不持久化到本地数据库。仅涉及 AWS 云厂商。

#### Scenario: 查询 GPU 实例类型成功
- **GIVEN** root_account_id、main_account_id、role_chain、region 均有效
- **WHEN** 调用 GPU 实例类型查询接口
- **THEN** 返回该 region 下的 AWS 实例类型列表，每条记录包含 InstanceType、CPU、Memory、GPU（Count/Memory/Name/Manufacturer）

#### Scenario: region 无效
- **GIVEN** region 参数无效或成员账号未激活该 region
- **WHEN** 调用 GPU 实例类型查询接口
- **THEN** 返回 AWS API 错误信息

#### Scenario: main_account_id 无效
- **GIVEN** main_account_id 在 main_account 表中不存在
- **WHEN** 调用 GPU 实例类型查询接口
- **THEN** 返回错误，提示该 main_account_id 无效

### Requirement: GPU 实例列表透传接口
hc-service SHALL 提供 GPU 实例列表查询接口，接收 `root_account_id`（根账号 HCM 内部 ID）、`main_account_id`（二级账号 HCM 内部 ID）、`role_chain`（角色名数组，支持 Role Chaining）、`region` 参数。系统通过 `root_account_id` 调 `AwsRoot()` 获取根账号 AK/SK，通过 `main_account_id` 查 `main_account` 表获取目标成员账号 CloudID，再按 role_chain 顺序执行链式 AssumeRole 获取成员账号权限后，调用 `DescribeInstances` 并返回该 region 下的 EC2 实例列表。参考 GCP GPU monitoring 的入参模式。该接口为 AWS 数据透传，不持久化到本地数据库。仅涉及 AWS 云厂商。

#### Scenario: 查询实例列表成功
- **GIVEN** root_account_id、main_account_id、role_chain、region 均有效
- **WHEN** 调用 GPU 实例列表查询接口
- **THEN** 返回该 region 下的 EC2 实例列表，每条记录包含 InstanceId、InstanceType、State 等核心字段

#### Scenario: 无实例
- **GIVEN** 成员账号在该 region 下无 EC2 实例
- **WHEN** 调用 GPU 实例列表查询接口
- **THEN** 返回空列表，不报错

#### Scenario: AssumeRole 失败
- **GIVEN** AssumeRole 过程中出错（权限不足、角色不存在等）
- **WHEN** 调用任一 GPU 数据接口
- **THEN** 返回 AssumeRole 相关错误信息，不透出 AWS 原始凭证内容

### Requirement: GPU 接口 cloud-server 资源视角入口
cloud-server SHALL 提供 GPU 实例类型查询和 GPU 实例列表查询的资源视角入口 handler。入口 handler SHALL 执行鉴权后，通过微服务调用 hc-service 对应接口。路由注册于 cloud-server 的资源视角路由组。仅涉及 AWS 云厂商。

#### Scenario: cloud-server 路由注册
- **GIVEN** cloud-server 启动
- **WHEN** 初始化资源视角路由
- **THEN** `/vendors/aws/gpu/instance_types/list` 和 `/vendors/aws/gpu/instances/list` 两个路由 SHALL 被注册

#### Scenario: 鉴权通过后转调 hc-service
- **GIVEN** 下游平台通过 API 网关调用 cloud-server GPU 接口
- **WHEN** 鉴权通过
- **THEN** cloud-server handler 调用 hc-service 的对应接口并返回结果

### Requirement: GPU 接口蓝鲸 API 网关注册
GPU 实例类型查询和 GPU 实例列表查询接口 SHALL 在蓝鲸 API 网关 YAML（`bk_apigw_resources_bk-hcm.yaml`）中注册为开放接口，供下游 GPU 资源分析平台通过 API 网关调用。仅涉及 AWS 云厂商。

#### Scenario: API 网关注册
- **GIVEN** 部署 API 网关配置
- **WHEN** 加载 `bk_apigw_resources_bk-hcm.yaml`
- **THEN** GPU 实例类型查询和 GPU 实例列表查询两个接口 SHALL 出现在开放接口列表中

### Requirement: GPU 接口文档
SHALL 按 HCM 接口文档规范（`docs/api-docs/web-server/docs/resource/`）编写 GPU 实例类型查询和 GPU 实例列表查询的接口文档，包含请求参数、响应结构、错误码说明。仅涉及 AWS 云厂商。

#### Scenario: 接口文档完整
- **GIVEN** 开发完成 GPU 透传接口
- **WHEN** 检查接口文档
- **THEN** `list_aws_gpu_instance_type.md` 和 `list_aws_gpu_instance.md` 两个文档 SHALL 存在且包含请求参数、响应结构、示例

### Requirement: 下游平台调用流程
下游 GPU 资源分析平台 SHALL 按以下步骤调用 HCM：
1. 以 `root_account_id`（根账号 HCM 内部 ID）+ `main_account_id`（二级账号 HCM 内部 ID）+ `role_chain`（角色名数组）+ `region` 直接调用 GPU 数据接口
2. HCM 内部自动用 `root_account_id` 取根账号 AK/SK，用 `main_account_id` 查 `main_account` 表获取目标成员账号 CloudID，按 role_chain 顺序完成链式 AssumeRole

与 GCP GPU monitoring 的入参模式一致。仅涉及 AWS 云厂商。

#### Scenario: 单步调用
- **GIVEN** 下游平台持有 root_account_id、main_account_id，role_chain 为 `["gpu-readonly"]`
- **WHEN** 以 root_account_id + main_account_id + role_chain + region 调用 GPU 实例类型接口和 GPU 实例列表接口
- **THEN** 成功获取目标成员账号在该 region 的实例类型规格和实例列表

#### Scenario: 多步 Role Chain 调用
- **GIVEN** 下游平台持有 root_account_id、main_account_id，role_chain 为 `["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"]`
- **WHEN** 以 root_account_id + main_account_id + role_chain + region 调用 GPU 数据接口
- **THEN** HCM 先在管理账号中 Assume GPUInventoryCallerRole，再在成员账号中 Assume GPUInventoryReadOnlyRole，成功获取数据

#### Scenario: main_account_id 无效
- **GIVEN** 下游平台传入的 main_account_id 在 HCM 的 main_account 表中不存在
- **WHEN** 调用 GPU 数据接口
- **THEN** 返回错误提示该 main_account_id 无效
