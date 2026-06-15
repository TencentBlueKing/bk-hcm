## Why

现有单据明细查看接口 `GET /api/v1/cloud/applications/{application_id}` 的权限控制过于严格：要求**本人查看**或**全局单据管理权限**，导致业务管理者无法查看其业务下其他人提交的单据，且不符合按业务维度进行管理的真实场景。

因此需要新增一个业务级别的单据明细查看接口：鉴权维度从"本人/全局单据权限"降级为"**业务访问权限**"。对用户而言，只有拥有对应业务访问权限的用户才能查看"该业务视角下"的单据明细。

## What Changes

### 1. 新增接口

```
GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}
```

用于业务视角下查看单据明细。

### 2. 鉴权维度变更

新接口使用**业务访问权限**（`meta.Biz.Access`）鉴权，而非全局单据权限。

### 3. 归属校验

验证单据的 `bk_biz_ids` 与请求路径中的 `bk_biz_id` 是否匹配：
- 若 `bk_biz_id` **在** 单据的 `bk_biz_ids` 列表中：允许访问
- 若 `bk_biz_id` **不在** 单据的 `bk_biz_ids` 列表中：返回 `NotFound`（避免信息泄露）

### 4. 错误处理统一策略

| 场景 | 返回 | 说明 |
|-----|------|------|
| 用户无业务访问权限 | `NotFound` | 避免泄露单据存在性 |
| 单据不存在 | `NotFound` | 标准处理 |
| `bk_biz_id` 不在单据的 `bk_biz_ids` 中 | `NotFound` | 避免泄露单据归属其他业务 |
| 校验通过 | 正常返回单据明细 | - |

内部日志需区分鉴权失败/归属不匹配原因，便于排查。

### 5. 返回体

复用现有 `ApplicationGetResp` 结构。

### 6. 现有接口保持不变

`GET /api/v1/cloud/applications/{application_id}` 不做任何行为改变。

## Capabilities

### New Capabilities

- **`biz-application-detail`**: 业务视角下的单据明细查看能力
  - **接口路径**：`GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}`
  - **权限要求**：业务访问权限（`meta.Biz.Access`）
  - **归属校验**：校验单据的 `bk_biz_ids` 包含请求路径中的 `bk_biz_id`
  - **返回结构**：复用现有 `ApplicationGetResp`

### Modified Capabilities

- 无，现有接口保持不变

## Impact

### Service Layer (cloud-server)

| 文件 | 变更 |
|-----|------|
| `cmd/cloud-server/service/application/init.go` | 新增路由注册 |
| `cmd/cloud-server/service/application/get.go` | 新增 `GetBizApplication` 函数 |

**实现要点**：
1. 业务访问权限鉴权（`meta.Biz.Access`）
2. 归属校验（`bk_biz_id` ∈ `bk_biz_ids`）
3. 统一 `NotFound` 错误策略
4. 敏感字段脱敏（复用 `RemoveSenseField`）
5. ITSM 场景处理（复用现有逻辑获取 `TicketUrl`）

### Protocol Layer (pkg/api)

| 文件 | 变更 |
|-----|------|
| `pkg/api/cloud-server/application/get.go` | 无需修改，复用现有 `ApplicationGetResp` |

### Data Service Layer (data-service)

- 无需修改，复用现有 `GetApplication` 接口

### Documentation

| 文件 | 变更 |
|-----|------|
| `docs/api-docs/web-server/docs/biz/get_biz_application.md` | 新增接口文档 |

文档需明确：
- 权限要求为业务访问权限
- 单据必须归属于请求的业务才能查看
- 错误场景统一返回 `NotFound`

## Open Questions

### 1. ✅ 已确认：多业务场景处理口径

**结论**：业务ID传哪个，就只能访问哪个业务。
- 路径中的 `bk_biz_id` 必须在单据的 `bk_biz_ids` 列表中才允许查看

### 2. ✅ 已确认：错误处理策略

**结论**：统一返回 `NotFound`。
- 当用户无权限或 `bk_biz_id` 不匹配时，对外统一返回 `NotFound`
- 避免泄露单据存在性或其他业务关联信息
- 内部日志区分鉴权失败/归属不匹配原因便于排查

### 3. ✅ 已确认：敏感字段与脱敏

**结论**：保持与现有接口一致。
- 使用相同的脱敏逻辑（`RemoveSenseField`）
