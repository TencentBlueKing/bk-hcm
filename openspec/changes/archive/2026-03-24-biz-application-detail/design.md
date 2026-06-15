## Context

现有单据明细接口 `GET /api/v1/cloud/applications/{application_id}` 的权限模型：申请人本人可直接查看，非申请人需要全局单据查找权限（`meta.Application.Find`）。这导致业务管理者无法查看其管理业务下其他人提交的单据。

本次变更新增业务视角的单据查看接口，使用业务访问权限鉴权，实现按业务维度的单据查看能力。Cloud-Server 层已有成熟的业务鉴权模式（如 `woa-server/logics/biz/biz.go` 中的 `ListAuthorizedBiz`），本次变更遵循现有模式。

## Goals / Non-Goals

**Goals:**
- 提供基于业务访问权限的单据明细查看接口
- 遵循项目现有的鉴权模式，保持一致性
- 统一返回 `NotFound` 避免信息泄露

**Non-Goals:**
- 不修改现有接口 `GET /api/v1/cloud/applications/{application_id}` 的行为
- 不修改 Data Service 层代码
- 不引入新的返回结构（复用 `ApplicationGetResp`）

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Request                           │
│   GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{app_id}     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Cloud-Server Layer                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              GetBizApplication Handler                   │   │
│  │  1. 业务访问权限鉴权 (meta.Biz.Access)                    │   │
│  │  2. 调用 Data Service 获取单据                           │   │
│  │  3. 归属校验 (bk_biz_id ∈ bk_biz_ids)                    │   │
│  │  4. 敏感字段脱敏 (RemoveSenseField)                       │   │
│  │  5. ITSM 场景处理 (获取 TicketUrl)                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Data-Service Layer                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         GetApplication (现有接口，无需修改)               │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### Cloud-Server Layer (Access Layer)
- **路由注册**：`GET /bizs/{bk_biz_id}/applications/{application_id}`
- **鉴权**：校验用户对 `bk_biz_id` 的业务访问权限
- **归属校验**：校验 `bk_biz_id` 是否在单据的 `bk_biz_ids` 中
- **返回体构建**：复用 `ApplicationGetResp`，调用 `RemoveSenseField` 脱敏
- **日志记录**：区分鉴权失败/归属不匹配/不存在的日志

### Data-Service Layer
- **无需修改**：复用现有 `GetApplication` 接口

## Decisions

### 1. 接口路径设计
**选择**: `GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}`
**理由**: 符合 RESTful 资源嵌套规范（业务下的单据），与现有业务维度接口风格一致。

### 2. 统一返回 NotFound
**选择**: 所有错误场景（无权限、不归属、不存在）统一返回 `NotFound`
**替代方案**: 区分返回 `PermissionDenied` 和 `NotFound`
**理由**: 避免泄露单据存在性和归属信息，简化前端错误处理。内部日志区分具体原因便于排查。

### 3. 复用现有实现
**选择**: 复用 Data Service 层 `GetApplication` 接口、`ApplicationGetResp` 返回结构、`RemoveSenseField` 脱敏逻辑
**理由**: 减少代码重复，保持一致性，降低维护成本。

### 4. 鉴权在 Cloud-Server 层实现
**选择**: 在 Cloud-Server 层进行业务访问权限鉴权
**理由**: 符合现有架构分层（Access Layer 负责鉴权），Data Service 层保持纯数据服务职责。

## Risks / Trade-offs

- **[日志需区分错误原因]** → 对外统一返回 NotFound，内部日志需明确区分鉴权失败/归属不匹配/不存在，便于问题排查。

## Implementation Notes

### 关键代码复用

1. **鉴权模式**：参考 `cmd/cloud-server/service/` 下其他业务接口的鉴权方式
2. **脱敏逻辑**：复用 `cmd/cloud-server/service/application/get.go` 中的 `RemoveSenseField`
3. **ITSM 处理**：复用现有 `GetApplication` 中获取 `TicketUrl` 的逻辑

### 测试覆盖

1. 有权限且归属匹配 → 返回单据明细
2. 无业务访问权限 → 返回 NotFound
3. 有权限但不归属 → 返回 NotFound
4. 单据不存在 → 返回 NotFound
5. ITSM 来源单据 → 返回包含 TicketUrl
