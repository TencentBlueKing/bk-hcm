## 1. BaseSecret 与 AWS SDK 基础改造

- [x] 1.1 `pkg/adaptor/types/secret.go`：`BaseSecret` 新增 `CloudSessionToken string` 字段
- [x] 1.2 `pkg/adaptor/aws/client.go`：`newClientSet` 中 `""` 改为 `secret.CloudSessionToken`
- [x] 1.3 验证非 STS 场景（CloudSessionToken 零值）行为不变

## 2. AssumeRole 核心能力

- [x] 2.1 `pkg/adaptor/aws/assume_role.go`：实现 `AssumeRole` 函数（AK/SK + Role ARN → STS 临时凭证）
- [x] 2.2 实现 `buildRoleArn` 辅助函数，按 site 区分 `aws` / `aws-cn` partition
- [x] 2.3 STS Region 默认：国际站 `us-east-1`，中国站 `cn-north-1`

## 3. 临时凭证缓存

- [x] 3.1 `cmd/hc-service/logics/cloud-adaptor/`：新增缓存结构体（`map[string]*CachedCredential` + `sync.Mutex`）
- [x] 3.2 缓存 key：`cloudAccountID + ":" + roleArn`
- [x] 3.3 缓存命中：距过期 > 10 分钟直接返回
- [x] 3.4 提前刷新：距过期 ≤ 10 分钟时调用 STS
- [x] 3.5 降级策略：刷新失败但旧凭证未过期时返回旧凭证 + WARN 日志
- [x] 3.6 首次获取：调用 STS + INFO 日志

## 4. AwsWithAssumeRole 编排方法（支持 Role Chain）

- [ ] 4.1 `cmd/hc-service/logics/cloud-adaptor/cloud_client.go`：修改 `AwsWithAssumeRole(kt, rootAccountID, cloudID, roleChain, externalId)` 签名，改用 `AwsRoot()` 获取根账号凭证（参考 GCP 的 `GcpRoot`），不再内部反查 sub_account
- [ ] 4.2 移除 `cmd/hc-service/logics/cloud-adaptor/secret.go` 中的 `AwsSubAccountByCloudID` 方法（不再需要）
- [x] 4.3 编排流程：AwsRoot 获取 AK/SK → 按 roleChain 顺序链式 AssumeRole（中间角色用管理账号 ID，最终角色用 cloudID）→ 构建 Aws client
- [x] 4.4 凭证缓存 `GetOrRefresh` 改为接收调用方构建的 cacheKey，支持 Role Chain 各步独立缓存

## ~~5. AWS sub_account 同步链路补完~~（已移除：改用 root_account + main_account 体系，不需要 sub_account 同步）

- [ ] 5.1 移除 `pkg/client/hc-service/aws/account.go` 中新增的 `SyncSubAccount` client 方法
- [ ] 5.2 移除 `cmd/cloud-server/service/sync/aws/sub_account.go` 同步入口文件
- [ ] 5.3 移除 `cmd/cloud-server/service/sync/aws/sync_all_resource.go` 中 syncOrder 的 sub_account 注册

## 6. 实例类型 GPU 字段补全

- [x] 6.1 `pkg/adaptor/types/instance-type/aws.go`：`AwsInstanceType` 新增 `GPUMemory`、`GPUName`、`GPUManufacturer`
- [x] 6.2 `pkg/adaptor/aws/instance_type.go`：`toAwsInstanceType` 补充 `GpuInfo` 解析
- [x] 6.3 `pkg/api/hc-service/instance-type/aws.go`：API 层结构体同步新增 GPU 字段

## 7. GPU 数据透传接口（hc-service）

- [ ] 7.1 `pkg/api/hc-service/`：更新 GPU 实例类型查询请求结构体（入参：`root_account_id` + `main_account_id` + `role_chain` + `region`）
- [ ] 7.2 `pkg/api/hc-service/`：更新 GPU 实例列表查询请求结构体（入参：`root_account_id` + `main_account_id` + `role_chain` + `region`）
- [ ] 7.3 `cmd/hc-service/`：更新 GPU 实例类型查询 handler（用 root_account_id 调 AwsRoot + main_account_id 查 CloudID → AwsWithAssumeRole → adaptor.ListInstanceType）
- [ ] 7.4 `cmd/hc-service/`：更新 GPU 实例列表查询 handler（同上模式 → adaptor.ListCvm）
- [x] 7.5 注册 hc-service 路由

## 8. GPU 接口 cloud-server 入口、开放接口与文档

- [x] 8.1 `pkg/client/hc-service/aws/`：新增 `ListGpuInstanceType` 和 `ListGpuInstance` client 方法
- [ ] 8.2 `cmd/cloud-server/service/instance-type/svc.go`：更新资源视角 handler（移除 `lookupAccountIDByCloudID` 反查逻辑）
- [x] 8.3 `cmd/cloud-server/service/instance-type/init.go`：注册路由 `/vendors/aws/gpu/instance_types/list` 和 `/vendors/aws/gpu/instances/list`
- [x] 8.4 `docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml`：注册两个 GPU 开放接口
- [ ] 8.5 `docs/api-docs/web-server/docs/resource/list_aws_assume_role_instance_type.md`：更新接口文档
- [ ] 8.6 `docs/api-docs/web-server/docs/resource/list_aws_assume_role_instance.md`：更新接口文档
