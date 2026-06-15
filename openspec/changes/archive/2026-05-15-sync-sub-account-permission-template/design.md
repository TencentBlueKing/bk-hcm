## Context

系统已有完整的子账号（`sub_account`）同步能力，通过 `cmd/hc-service/logics/res-sync/tcloud/sub_account.go` 实现。同步流程包括：从云上拉取子账号列表 → 与本地对比 → 新增/更新/删除。同时系统已有 `permission_template` 表存储权限模板，以及 `ListPolicies`、`GetPolicyDetail` 等 CAM 策略相关 adaptor 方法。

当前缺失的能力：
1. 无法获取子账号绑定的策略列表（缺少 `ListAttachedUserAllPolicies` adaptor 方法）
2. `sub_account` 表没有字段存储子账号绑定的权限模板 ID 列表
3. 同步流程中缺少同步子账号权限模板的步骤

腾讯云 CAM `ListAttachedUserAllPolicies` API 文档：https://cloud.tencent.com/document/product/598/67728

## Goals / Non-Goals

**Goals:**
- 在 `sub_account` 表新增 `permission_template_ids` 字段（JSON 数组类型）
- 新增 TCloud CAM `ListAttachedUserAllPolicies` adaptor 方法，支持获取子用户绑定的策略列表（含限流重试）
- 在 res-sync tcloud 层新增 `SubAccountPermissionTemplate` 同步方法
- 在 `SyncSubAccount` 服务层调用链中增加权限模板同步步骤
- 同步失败时返回错误，中断同步流程

**Non-Goals:**
- 不实现其他云厂商（AWS、Azure 等）的子账号权限模板同步
- 不实现权限模板的自动创建（如果云上策略在本地不存在，报错提示）
- 不修改子账号同步的核心逻辑，仅新增权限模板同步步骤
- 不实现权限模板绑定的创建/删除操作（仅同步现有绑定关系）

## Decisions

### Decision 1: `permission_template_ids` 使用 `types.StringArray` 类型

**选择**：使用 `types.StringArray`（JSON 数组）存储权限模板 ID 列表。

**理由**：
- 与现有 `managers` 字段（`types.StringArray`）保持一致
- 权限模板 ID 数量通常较少（10 个以内），JSON 数组性能可接受
- 便于后续扩展（如存储更多绑定信息）

**替代方案**：新建关联表 `sub_account_permission_template_rel`。否决：增加复杂度，子账号与权限模板是多对多关系，但当前场景只需单向查询（子账号→模板），JSON 数组足够。

### Decision 2: 同步失败返回错误

**选择**：`SubAccountPermissionTemplate` 同步失败时，记录错误日志，返回错误，中断同步流程。

**理由**：
- 权限模板同步是核心功能，失败应该让用户感知并处理
- 同步失败可能意味着数据不一致，需要人工介入排查
- 返回错误可以让上层调用方（如定时任务、手动触发）进行重试或告警

**替代方案**：同步失败时降级处理，记录日志但不返回错误。否决：隐藏错误会导致数据不一致问题难以发现。

### Decision 3: 云上策略不存在于本地时报错

**选择**：遍历子账号绑定的云上策略时，如果 `permission_template` 表中找不到对应的 `cloud_id` 记录，记录错误日志并跳过该策略，继续处理其他策略。

**理由**：
- 权限模板需要先通过 `PermissionTemplate` 同步接口同步到本地
- 如果缺少模板，说明前置同步未完成或失败，需要人工介入
- 跳过而非中断，确保已匹配的模板能正确同步

**替代方案**：自动创建缺失的权限模板。否决：需要调用 `GetPolicyDetail` 获取策略详情，增加复杂度和 API 调用次数；策略详情同步应由 `PermissionTemplate` 同步流程负责。

### Decision 4: `ListAttachedUserAllPolicies` 支持分页和限流重试

**选择**：adaptor 方法支持分页（`Page`、`Rp` 参数），并继承 adaptor 已有的 `SetRateLimitRetryWithRandomInterval` 限流重试机制。

**理由**：
- 子账号绑定的策略数量可能较多，分页避免单次请求过大
- CAM API 有频率限制，重试机制保证同步稳定性
- 与现有 `ListPolicies` 等方法保持一致

## Risks / Trade-offs

- **Risk**: 子账号数量较多时，逐个调用 `ListAttachedUserAllPolicies` 可能触发 API 限流
  - **Mitigation**: 使用 adaptor 已有的限流重试机制；考虑并发控制（限制并发数）

- **Risk**: 云上策略不存在于本地 `permission_template` 表时，该绑定关系无法同步
  - **Mitigation**: 记录错误日志，提示用户先同步权限模板；后续可考虑增加自动化修复工具

- **Risk**: 新增字段 `permission_template_ids` 需要 SQL 迁移
  - **Mitigation**: 使用标准 DDL 迁移脚本，字段设置为 `DEFAULT NULL` 兼容现有数据

## Migration Plan

1. **阶段一：数据库变更**
   - 新增 SQL DDL 文件，为 `sub_account` 表添加 `permission_template_ids` 字段（`json` 类型，`DEFAULT NULL`）
   - 更新 Table 定义、DAO、API Model 层

2. **阶段二：Adaptor 封装**
   - 新增 `ListAttachedUserAllPolicies` 类型定义和 adaptor 方法
   - 更新 `TCloud` interface

3. **阶段三：同步逻辑实现**
   - 在 res-sync tcloud 层新增 `SubAccountPermissionTemplate` 方法
   - 在 `SyncSubAccount` 服务层增加调用

4. **阶段四：测试验证**
   - 单元测试 adaptor 方法
   - 集成测试同步流程

## Open Questions

- 是否需要支持批量查询子账号的权限模板绑定（减少 API 调用次数）？当前设计是逐个查询。
  - **当前决策**: 逐个查询，后续如遇性能问题再优化
  
- `permission_template_ids` 是否需要包含策略类型（预设/自定义）？
  - **当前决策**: 仅存储模板 ID，策略类型可通过关联 `permission_template` 表获取
