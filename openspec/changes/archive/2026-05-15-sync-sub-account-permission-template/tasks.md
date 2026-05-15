## 1. 数据库变更

- [x] 1.1 新增 SQL DDL 文件，为 `sub_account` 表添加 `permission_template_ids` 字段（`json` 类型，`DEFAULT NULL`）
- [x] 1.2 更新 `pkg/dal/table/cloud/sub-account/sub_account.go` Table 定义，新增 `PermissionTemplateIDs` 字段（`types.StringArray` 类型）
- [x] 1.3 更新 `pkg/dal/table/cloud/sub-account/sub_account.go` ColumnDescriptor，新增 `permission_template_ids` 列定义
- [x] 1.4 更新 `pkg/api/core/cloud/sub-account/` API Model，新增 `permission_template_ids` 字段
- [x] 1.5 更新 DAO 层，支持 `permission_template_ids` 字段的读写操作

## 2. Adaptor 封装

- [x] 2.1 在 `pkg/adaptor/types/account/` 新增 `TCloudListAttachedUserAllPoliciesOption` 类型定义（包含 `TargetUin`、`Page`、`Rp` 字段）
- [x] 2.2 在 `pkg/adaptor/types/account/` 新增 `TCloudAttachedPolicy` 类型定义（包含 `PolicyId`、`PolicyName`、`AddTime`、`PolicyType` 等字段）
- [x] 2.3 在 `pkg/adaptor/types/account/` 新增 `TCloudListAttachedUserAllPoliciesResult` 类型定义（包含策略列表和分页信息）
- [x] 2.4 在 `pkg/adaptor/tcloud/` 新增 `ListAttachedUserAllPolicies` 方法实现，支持分页和限流重试
- [x] 2.5 更新 `pkg/adaptor/tcloud/interface.go`，在 `TCloud` interface 新增 `ListAttachedUserAllPolicies` 方法签名

## 3. 同步逻辑实现

- [x] 3.1 在 `cmd/hc-service/logics/res-sync/tcloud/` 新增 `sub_account_permission_template.go` 文件
- [x] 3.2 实现 `SubAccountPermissionTemplate` 方法：
  - 查询指定账号下所有子账号（通过 data-service）
  - 逐个调用 `ListAttachedUserAllPolicies` 获取绑定的策略列表
  - 通过 `cloud_id` 匹配本地 `permission_template` 记录
  - 收集匹配的本地模板 ID 列表
  - 批量更新 `sub_account` 表的 `permission_template_ids` 字段
  - 记录同步日志（成功/失败/跳过数量）
- [x] 3.3 更新 `cmd/hc-service/logics/res-sync/tcloud/client.go` Interface，新增 `SubAccountPermissionTemplate` 方法签名
- [x] 3.4 在 `cmd/hc-service/service/sync/tcloud/sub_account.go` 的 `SyncSubAccount` 方法中增加 `SubAccountPermissionTemplate` 调用（在 `SubAccount` 同步之后）

## 4. 单元测试

- [ ] 4.1 为 `ListAttachedUserAllPolicies` adaptor 方法编写单元测试
- [ ] 4.2 为 `SubAccountPermissionTemplate` 同步方法编写单元测试

## 5. 集成测试

- [ ] 5.1 手动测试同步流程：创建子账号 → 绑定策略 → 执行同步 → 验证 `permission_template_ids` 正确更新
- [ ] 5.2 测试异常场景：云上策略不存在于本地 → 验证日志记录和跳过逻辑
- [ ] 5.3 测试限流重试：模拟 API 限流 → 验证重试机制生效
