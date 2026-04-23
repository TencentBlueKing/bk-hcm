## MODIFIED Requirements

### Requirement: TmplBaseInfo 结构体

系统 SHALL 在 `cmd/cloud-server/service/permission-policy-library/applier.go` 中新增 `TmplBaseInfo` 结构体：

```go
type TmplBaseInfo struct {
    Name string  `json:"name"`
    Memo *string `json:"memo"`
}
```

该结构体用于统一传递权限模板的名称和备注信息，供 `TCloudCreateCAMPolicy`、`TCloudCreateLocalTemplate` 及新增的 `ApplyCreateWithTmplInfo` 使用。

### Requirement: TCloud CAM Policy 创建函数签名

`PolicyLibraryApplier.TCloudCreateCAMPolicy` 方法签名 SHALL 将原有的独立 `name string, memo *string` 参数改为接受 `tmplInfo TmplBaseInfo`，方法体中 `PolicyName` 使用 `tmplInfo.Name`，`Description` 使用 `cvt.PtrToVal(tmplInfo.Memo)`。

现有唯一调用方 `tcloudApplyCreateForAccount`（同文件内）SHALL 改为传入 `TmplBaseInfo` 参数，行为与修改前完全一致。

#### Scenario: 原有调用方传入 library.Name/Memo
- **WHEN** `tcloudApplyCreateForAccount` 调用 `TCloudCreateCAMPolicy(kt, library, accountID, TmplBaseInfo{Name: library.Name, Memo: library.Memo})`
- **THEN** 创建的 CAM 策略名称与原行为相同，Description 与原行为相同

#### Scenario: 自定义名称调用
- **WHEN** create_permission_template 的 deliver 调用 `TCloudCreateCAMPolicy(kt, library, accountID, TmplBaseInfo{Name: "my-custom-name", Memo: nil})`
- **THEN** 创建的 CAM 策略名称为 "my-custom-name"，Description 为空字符串

### Requirement: TCloud 本地模板创建函数签名

`PolicyLibraryApplier.TCloudCreateLocalTemplate` 方法签名 SHALL 将原有的独立 `name string, memo *string` 参数改为接受 `tmplInfo TmplBaseInfo`。方法体中模板记录的 `Name` 字段使用 `tmplInfo.Name`，`Memo` 字段使用 `tmplInfo.Memo`（指针类型）。

现有唯一调用方 `tcloudApplyCreateForAccount` SHALL 改为传入 `TmplBaseInfo` 参数，行为与修改前完全一致。

#### Scenario: 原有调用方传入 library.Name/Memo
- **WHEN** `tcloudApplyCreateForAccount` 调用 `TCloudCreateLocalTemplate(kt, library, accountID, cloudPolicyID, TmplBaseInfo{Name: library.Name, Memo: library.Memo})`
- **THEN** 写入的本地模板记录 Name/Memo 与原行为相同

#### Scenario: 自定义名称和备注调用
- **WHEN** create_permission_template 的 deliver 调用 `TCloudCreateLocalTemplate(kt, library, accountID, cloudPolicyID, TmplBaseInfo{Name: "my-template", Memo: cvt.ValToPtr("my memo")})`
- **THEN** 写入的本地模板记录 Name 为 "my-template"，Memo 为 "my memo"

### Requirement: ApplyCreateWithTmplInfo 方法

系统 SHALL 在 `applier.go` 中新增 `ApplyCreateWithTmplInfo` 公开方法，签名为：

```go
func (a *PolicyLibraryApplier) ApplyCreateWithTmplInfo(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
    accountIDs []string, tmplInfo TmplBaseInfo) (*proto.ApplyPermissionPolicyLibraryResult, error)
```

该方法与原有 `ApplyCreate` 逻辑相同（GetPolicyLibraryDetail → CheckAccountsBizInScope → tcloudApplyCreate），区别在于使用调用方传入的 `tmplInfo` 而非 `TmplBaseInfo{Name: library.Name, Memo: library.Memo}`。

原有 `ApplyCreate` 方法 SHALL 保留，其行为不变，内部调用 `tcloudApplyCreate` 时传入 `TmplBaseInfo{Name: library.Name, Memo: library.Memo}`。

#### Scenario: create_permission_template deliver 调用
- **WHEN** `ApplyCreateWithTmplInfo(kt, vendor, libraryID, []string{accountID}, TmplBaseInfo{Name: "custom", Memo: nil})` 被调用
- **THEN** 创建的 CAM 策略名称和本地模板名称均为 "custom"，Memo 为 nil

#### Scenario: 原有 ApplyCreate 行为不变
- **WHEN** `ApplyCreate(kt, vendor, libraryID, accountIDs)` 被调用
- **THEN** 创建的 CAM 策略名称使用 library.Name，Memo 使用 library.Memo，行为与修改前完全一致
