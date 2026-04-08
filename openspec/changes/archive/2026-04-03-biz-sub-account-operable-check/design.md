## Context

`ListBizSubAccountExt` 当前依赖 `svc.listSubAccountExt(cts, listBizSubAccountAuthRes)` 返回三级账号扩展列表，但结果中没有“当前业务是否可操作”与“所属二级账号名称”两个关键字段。业务判断逻辑如果分散在接口层会导致重复实现和口径不一致，因此需要在 cloud-server 服务层新增公共能力并在列表接口中做后处理组装。

当前需求明确约束：
- 允许重构 `ListBizSubAccountExt` 与 `listSubAccountExt` 的调用关系和内部实现。
- 重构后必须保持行为语义等价（鉴权范围、过滤组合、vendor 路由、返回语义）。

## Goals / Non-Goals

**Goals:**
- 在 `cmd/cloud-server/logics/account` 新增可复用的批量判定函数，支持输入 `bk_biz_id + account_id[]`，输出 `account_id -> operable` 映射。
- 在 `ListBizSubAccountExt` 的后处理阶段补充 `operable` 与 `account_name` 字段。
- 不直接修改核心 `SubAccount` 结构，新增组合响应结构统一承载扩展信息。
- 封装转换函数，避免在 handler 中散落 map 访问与字段拼装逻辑。

**Non-Goals:**
- 不调整 `listSubAccountExt` 的查询逻辑与返回结构。
- 不修改 `pkg/api/core/cloud/sub-account/sub_account.go` 中已有 `SubAccount` 定义。
- 不引入新的数据库表或外部依赖。

## Decisions

- **决策 1：允许重构，但保持行为等价**
  - 允许根据实现可维护性重构 `ListBizSubAccountExt` 与 `listSubAccountExt` 的流程编排。
  - operable/account_name 的处理阶段可放在独立分支或后处理步骤，只要外部行为保持等价。
  - 原因：减少重复逻辑，提升复用性，同时避免行为回归。

- **决策 2：公共函数放到 account logic 层**
  - 在 `cmd/cloud-server/logics/account` 增加公共函数，形如：
    - 批量查询账号简要信息（`account_id -> account(bk_biz_id,name)`）
    - 批量生成可操作映射（`account_id -> bool`）
  - 原因：`account` 归属判断是账号域逻辑，避免在 sub-account service 里重复拼 filter。

- **决策 3：新增组合结构，不改 SubAccount**
  - 新增响应结构（名称待代码实现确认），包含：
    - `SubAccount`（inline 或嵌套）
    - `Operable bool`
    - `AccountName string`
  - 原因：保持 core 结构稳定，满足扩展字段按场景注入。

- **决策 4：转换函数统一封装**
  - 新增转换函数，输入：
    - 原 `[]SubAccount` 列表
    - operable map
    - account name map
  - 输出：扩展结构数组。
  - 原因：避免 handler 里手工拼字段，便于单测覆盖。

## Risks / Trade-offs

- [Risk] account_id 去重不足导致重复查询/重复 map 覆盖 → Mitigation：先去重再批量查询。
- [Risk] 某些三级账号的 account_id 在 account 表不存在 → Mitigation：`operable=false`，`account_name` 置空字符串，并记录调试日志。
- [Risk] 扩展结构与现有响应契约不一致 → Mitigation：保持原字段兼容，仅新增字段，补充接口层测试断言。
- [Risk] 重构引入行为漂移（鉴权或 filter 语义变化）→ Mitigation：增加行为等价测试（无权限、带过滤、不同 vendor、count 查询场景）。
