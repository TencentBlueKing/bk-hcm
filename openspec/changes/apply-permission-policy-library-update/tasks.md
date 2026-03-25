## 1. TCloud CAM UpdatePolicy Adaptor

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud_policy.go` 新增 `TCloudUpdatePolicyOption`（`PolicyID uint64`、`PolicyDocument string`）类型定义及 `Validate` 方法
- [x] 1.2 在 `pkg/adaptor/tcloud/cam_policy.go` 实现 `UpdatePolicy` 方法，封装 `cam.NewUpdatePolicyRequest` 调用（传 `PolicyId` + `PolicyDocument`）
- [x] 1.3 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `UpdatePolicy` 方法声明

## 2. hc-service CAM 策略更新接口

- [x] 2.1 在 `pkg/api/hc-service/permission-template/permission_template.go` 新增 `UpdateCAMPolicyReq`（`AccountID string`、`PolicyID uint64`、`PolicyDocument string`）及 `Validate` 方法
- [x] 2.2 在 `cmd/hc-service/service/permission-template/cam_policy.go` 新增 `TCloudUpdateCAMPolicy` handler，获取 TCloud adaptor 并调用 `UpdatePolicy`，成功返回 nil data
- [x] 2.3 在 `cmd/hc-service/service/permission-template/service.go` 路由注册中新增 `PATCH /vendors/tcloud/permission_templates/cam/update_policy`

## 3. hc-service Client 封装

- [x] 3.1 在 `pkg/client/hc-service/tcloud/permission_template.go` 新增 `UpdateCAMPolicy(kt, req) error` client 方法，调用 hc-service 新接口

## 4. cloud-server applier 扩展

- [x] 4.1 在 `cmd/cloud-server/service/permission-policy-library/applier.go` 新增 `GetAccountTemplate(kt, libraryID, accountID)` 方法，返回 `*PermissionTemplateExt[TCloud]`（nil 表示未应用）
- [x] 4.2 新增 `TCloudUpdateCAMPolicy(kt, library, accountID, cloudPolicyID uint64) error` 方法，调用 hc-service client `UpdateCAMPolicy`
- [x] 4.3 新增 `TCloudUpdateLocalTemplate(kt, library, templateID string) error` 方法，调用 data-service `BatchUpdate`（更新 `PolicyDocument`、`PolicyLibraryVersion`、`PolicyLibrarySyncTime`）
- [x] 4.4 新增 `tcloudApplyUpdateForAccount` 方法：`GetAccountTemplate`（nil → failed）→ 解析 CloudID `strconv.ParseUint` → `TCloudUpdateCAMPolicy` → `TCloudUpdateLocalTemplate` → `RecordApplyAudit`
- [x] 4.5 新增 `tcloudApplyUpdate` 方法：循环调用 `tcloudApplyUpdateForAccount`，收集结果
- [x] 4.6 新增 `ApplyUpdate(kt, vendor, libraryID, accountIDs)` 入口方法：`GetPolicyLibraryDetail` → `CheckAccountsBizInScope` → 按 vendor 分派 `tcloudApplyUpdate`

## 5. cloud-server Apply Update Handler

- [x] 5.1 在 `cmd/cloud-server/service/permission-policy-library/` 的 `apply.go`，实现 `ApplyPermissionPolicyLibraryUpdate` handler（vendor 校验 → id 校验 → 解码 → 鉴权 → 调 `applier.ApplyUpdate`）
- [x] 5.2 在 `cmd/cloud-server/service/permission-policy-library/service.go` 路由注册中新增 `PUT /vendors/{vendor}/permission_policy_libraries/{id}/apply` 路由，绑定 `ApplyPermissionPolicyLibraryUpdate`

## 6. 验证与编译

- [x] 6.1 执行 `go build ./...` 确认编译通过
- [x] 6.2 检查 linter 无新增错误
