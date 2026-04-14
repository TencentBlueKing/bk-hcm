## Why

GPU 资源分析平台需要通过 HCM 获取 AWS 所有成员账号下的实例类型规格（含 GPU 显存/型号/制造商）和实例列表数据。AWS 成员账号没有在 HCM 中录入独立 AK/SK，必须通过管理账号的 AK/SK + STS AssumeRole 获取临时凭证后访问成员账号资源。当前 HCM 尚未实现 AssumeRole 能力。

## What Changes

- **新增 AssumeRole 能力**：在 `pkg/adaptor/aws/` 新增 `AssumeRole` 函数，通过 STS API 获取临时凭证访问成员账号
- **新增 `AwsWithAssumeRole` 编排方法（支持 Role Chain）**：在 `CloudAdaptorClient` 中封装完整链路（用 root_account_id 取根账号 AK/SK → 用 main_account_id 查 main_account 获取目标成员账号 CloudID → 按 role_chain 顺序链式 AssumeRole → 构建 client），参考 GCP GPU monitoring 的 `GcpRoot` + `MainAccount` 模式
- **新增进程内凭证缓存**：缓存 AssumeRole 临时凭证，提前 10 分钟刷新，含降级策略和可观测性日志。Role Chain 场景下每步独立缓存
- **BaseSecret 新增 CloudSessionToken 字段**：支持 STS 临时凭证透传，对已有厂商零影响
- **newClientSet 支持 SessionToken**：将硬编码的 `""` 改为读取 `secret.CloudSessionToken`
- **实例类型 GPU 字段补全**：`AwsInstanceType` 新增 `GPUMemory`、`GPUName`、`GPUManufacturer` 三个字段，转换函数补充解析 `GpuInfo`
- **GPU 数据透传接口**：hc-service 实现 GPU 实例类型查询和实例列表查询 handler，入参为 `root_account_id` + `main_account_id` + `role_chain` + `region`
- **GPU 接口 cloud-server 入口与开放接口注册**：cloud-server 提供资源视角入口（含鉴权），在蓝鲸 API 网关 YAML 中注册为开放接口
- **GPU 接口文档**：按 HCM 接口文档规范编写接口文档

## Capabilities

### New Capabilities
- `aws-assume-role`: AssumeRole 跨账号访问能力，包括 STS 调用、Role ARN 自动拼接、临时凭证缓存、`AwsWithAssumeRole` 编排方法（基于 root_account + main_account 账单账号体系 + Role Chain 支持）
- `aws-gpu-instance-data`: 基于 AssumeRole 的 GPU 实例数据透传接口（实例类型 + 实例列表），含 GPU 字段补全、cloud-server 入口、API 网关注册、接口文档

### Modified Capabilities
（无现有 spec 需要修改）

## Impact

- **hc-service**（资源层）：新增 AssumeRole 函数、AwsWithAssumeRole 方法（支持 Role Chain）、凭证缓存、GPU 数据透传 handler
- **cloud-server**（服务层）：新增 GPU 接口资源视角入口 handler（含鉴权），转调 hc-service
- **pkg/adaptor/aws/**：新增 `AssumeRole` 导出函数、修改 `newClientSet` 支持 SessionToken
- **pkg/adaptor/types/**：`BaseSecret` 新增 `CloudSessionToken` 字段（零影响兼容）、`AwsInstanceType` 新增 GPU 字段
- **pkg/client/hc-service/aws/**：新增 `ListAssumeRoleInstanceType`、`ListAssumeRoleInstance` client 方法
- **蓝鲸 API 网关**：在 `bk_apigw_resources_bk-hcm.yaml` 中注册两个 AssumeRole 开放接口
- **接口文档**：新增 `list_aws_assume_role_instance_type.md` 和 `list_aws_assume_role_instance.md`
- **前置依赖**：AWS 运维需在所有成员账号中创建统一 IAM Role 并提供角色名；需确认管理账号 IAM User 有 `sts:AssumeRole` 权限
