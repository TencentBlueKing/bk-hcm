## Why

当前 `ListBizSubAccountExt` 返回的三级账号数据缺少“当前业务是否可操作该三级账号”的明确标识，也没有直接返回所属二级账号名称，前端和调用方无法稳定判断可操作性与展示信息。需要在服务端统一补齐判定与转换逻辑，避免重复判断和歧义。

## What Changes

- 在 `ListBizSubAccountExt` 调用 `svc.listSubAccountExt(...)` 后新增可操作性判定流程：
  - 以三级账号 `account_id` 为维度，批量判断“当前 `bk_biz_id` 是否等于该二级账号在 account 表中的业务”。
  - 命中则 `operable=true`，否则 `operable=false`。
- 允许在 `ListBizSubAccountExt` / `listSubAccountExt` 上做重构，但必须保持原有行为语义等价：
  - 权限范围与鉴权结果一致；
  - filter 组合规则一致（包含业务与 vendor 约束）；
  - vendor 路由与返回语义一致。
- 新增服务层公共函数（放在 `cmd/cloud-server/logics/account`）：
  - 输入：`bk_biz_id` 与 `account_id` 数组
  - 输出：`map[account_id]bool`（或等价结构，至少可表达 account 与 boolean 的映射）
- 在判定后，通过 `account_id` 批量查询二级账号名称，并补齐到返回结构中的 `account_name`。
- 不修改 `pkg/api/core/cloud/sub-account/sub_account.go` 的 `SubAccount` 结构；新增组合结构体包装：
  - 包含原 `SubAccount`
  - 新增 `operable bool`
  - 新增 `account_name string`
- 补充转换封装函数，统一完成“原始 SubAccount -> 扩展响应结构”的组装。

## Capabilities

### New Capabilities
- `biz-sub-account-ext-list`: 在业务三级账号扩展列表中返回可操作标识与二级账号名称，并提供可复用的服务层判定函数。

### Modified Capabilities

## Impact

- `cmd/cloud-server/service/sub-account/list.go`：扩展列表接口编排逻辑
- `cmd/cloud-server/logics/account`：新增可复用的业务判定函数与账号名称查询辅助逻辑
- `pkg/api/cloud-server` 或对应返回结构定义处：新增列表响应组合结构（不改核心 `SubAccount`）
- 相关单元测试/集成测试：新增 operable 与 account_name 的断言
