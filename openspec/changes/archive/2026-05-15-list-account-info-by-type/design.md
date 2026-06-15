## Context

当前 `cloud-server` 中的 `accountSvc` 已有 `GetAccountInfo` 接口（单个账号查询），以及 `accountResourceAuthChecker` 接口和 `subAccountChecker` 实现。但该接口仅支持单账号查询，权限校验逻辑仅覆盖 `sub_account` 一种资源类型，已不满足业务需求，需要替换为支持批量和多资源类型的新接口。

现有代码模式：
- 权限校验通过 `authorizer.Authorize` 实现业务访问权限检查
- `accountResourceAuthChecker` 接口用于资源级别的权限校验（将被调整扩展）
- `Vendor.GetMainAccountIDField()` 用于获取不同云厂商的主账号 ID 字段名
- `ListWithExtension` 已有批量查询带扩展字段账号信息的实现

新接口需要支持三种资源类型（sub_account、sub_account_secret、permission_template），每种类型的权限校验逻辑不同，且未来可能扩展更多类型，需要一种可扩展的策略模式。现有的 `accountResourceAuthChecker` 将被调整为支持批量过滤的 `accountTypeAuthChecker`，`GetAccountInfo` 接口将被删除。

## Goals / Non-Goals

**Goals:**
- 实现 `ListAccountInfoByType` 接口，支持根据资源类型批量查询二级账号元数据信息
- 调整现有 `accountResourceAuthChecker` 接口为支持批量过滤的 `accountTypeAuthChecker`，替代原有单账号维度的校验模式
- 删除现有的 `GetAccountInfo` 接口，仅保留批量查询接口
- 设计可扩展的权限校验策略模式，便于新增资源类型时只需添加新的校验器
- 复用项目中已有的工具函数和权限校验模式
- 每个函数不超过 80 行，按职责合理拆分
- 支持不同云厂商的扩展字段（extension）

**Non-Goals:**
- 不实现新增资源类型的校验逻辑（仅实现当前三种已有类型）

## Decisions

### 1. 权限校验策略模式：调整 `accountResourceAuthChecker` 为 `accountTypeAuthChecker`

**决策**：将现有 `accountResourceAuthChecker` 接口调整为 `accountTypeAuthChecker`，方法签名从单账号校验改为批量过滤。

**理由**：
- 现有 `accountResourceAuthChecker.check(kt, accountID, bizID) (bool, error)` 是针对单个账号+业务维度的校验，已不满足批量查询需求
- 新接口需要的是：给定一组账号 ID 和业务 ID，返回其中有权限的账号 ID 列表（批量过滤）
- 直接调整现有接口，避免同时维护两套校验器，保持代码整洁
- 原有 `GetAccountInfo` 接口将被删除，`accountResourceAuthChecker` 无其他调用方

**接口定义**：
```go
type accountTypeAuthChecker interface {
    // filterAuthorizedIDs 过滤出有权限的账号ID列表
    filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64) ([]string, error)
}
```

### 2. 校验器注册：使用 map[string]accountTypeAuthChecker

**决策**：在 `accountSvc` 中维护一个 `typeCheckerMap map[string]accountTypeAuthChecker`，key 为资源类型字符串。

**理由**：
- 简单直观，新增类型只需在初始化时注册
- 与项目中 `vendorInfoMap` 等模式一致
- 避免大量 switch-case 分支

### 3. 各资源类型校验逻辑

#### sub_account（三级账号）

校验逻辑：查询 sub_account 表，查看是否存在满足以下条件的 sub_account 记录：
- `account_id IN ids`：二级账号 ID 在传入的 ID 列表中
- `bk_biz_ids JSON_CONTAINS bizID`：使用业务列表包含当前业务 ID
- `vendor = vendor`：对应云厂商

实现方式：通过 countPage 方式查询 data-service，使用 `filter` 构建上述条件，返回匹配的 account_id 列表。

#### sub_account_secret（三级账号密钥）

校验逻辑：在 sub_account 校验通过的基础上，额外增加一个查询条件：
- 先与 sub_account 相同的条件查询出满足条件的 sub_account 记录
- 再查询 sub_account_secret 表，确认这些 sub_account 记录下是否存在密钥记录
- 即：只有同时存在 sub_account 记录且该记录下存在密钥的账号才通过校验

实现方式：先查询 sub_account 表获取满足条件的 account_id 列表，再查询 sub_account_secret 表过滤出有密钥的 account_id。

#### permission_template（权限模版）

校验逻辑：校验当前业务是否属于 account_id 对应的 account 的使用业务（usage_biz_ids）。
- 查询 account 表，条件为 `id IN ids`
- 过滤出 `usage_biz_ids JSON_CONTAINS bizID` 的账号
- 即：当前业务 ID 在该账号的使用业务列表中，才表示有权限

实现方式：查询 account 表，构建 `id IN ids AND usage_biz_ids JSON_CONTAINS bizID` 条件，返回匹配的 account id 列表。

### 4. 扩展字段获取：复用 `Vendor.GetMainAccountIDField()`

**决策**：批量查询账号后，通过 vendor 获取对应的主账号 ID 字段名，从扩展字段中提取。

**理由**：复用已有的 `Vendor.GetMainAccountIDField()` 方法，与 `ListBizAccount` 等接口一致。

### 5. 请求/响应结构体

**请求** `AccountListByTypeReq`：
- `IDs []string`：二级账号 ID 列表，最大 100
- `Type string`：资源类型

**响应** `AccountListByTypeResp`：
- `Details []AccountInfoByTypeDetail`：账号信息列表
- 每个 detail 包含：ID、Name、BkBizID、UsageBizIDs、Managers、Extension(map 类型)

### 6. 文件组织

- 删除 `cmd/cloud-server/service/account/info.go` 中的 `GetAccountInfo` 方法、`accountResourceAuthChecker` 接口和 `subAccountChecker` 实现
- 新增 `info_by_type.go` 文件在 `cmd/cloud-server/service/account/` 目录下，包含：
  - `ListAccountInfoByType` 入口函数
  - `filterAuthorizedAccountIDs` 按 type 分派校验
  - `batchGetAccountBaseInfo` 批量获取账号基本信息
  - `buildAccountInfoByTypeDetails` 构建响应详情
  - `getAccountExtensions` 批量获取云厂商扩展字段
- 新增 `info_by_type_checker.go` 文件，包含：
  - `accountTypeAuthChecker` 接口定义
  - 各资源类型的校验器实现（subAccountTypeChecker、subAccountSecretTypeChecker、permissionTemplateTypeChecker）
  - `typeCheckerMap` 初始化逻辑

## Risks / Trade-offs

- [批量查询性能风险] ids 最大 100 个，每种校验器需查询关联表 → 使用 `slice.Split` 分批查询，限制单次查询量
- [扩展性权衡] 使用 map 注册校验器而非反射/插件 → 简单直接，新增类型需修改初始化代码，但可接受
- [sub_account_secret 校验链路较长] 需通过 sub_account 中转 → 只查询必要的字段，减少数据传输
