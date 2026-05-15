## Context

- 业务子账号扩展列表 `ListBizSubAccountExt` 已在 cloud-server 对 data-service 结果做后处理：通过 `logicaccount.BatchBuildOperableAndNameMap` 得到 `operableMap`，再 `buildBizSubAccountExtDetails` 组装 `BizSubAccountItem`（含 `Operable`）。
- 业务子账号密钥列表 `ListSubAccountSecret`（`cmd/cloud-server/service/subaccount-secret/list.go`）当前在 TCloud 分支直接返回 `ListSubAccountSecretJoinExt` 的 data-service 结果，响应体无 `operable`。
- 账号 operable 判定规则已在 `cmd/cloud-server/logics/account/list.go` 的 `BuildOperableMapByAccountMap` 中定义：`accountInfo.BkBizID == bkBizID`。

## Goals / Non-Goals

**Goals:**

- 为 join 列表每条密钥明细增加 `operable`，语义与业务子账号列表一致。
- 在 `pkg/api/cloud-server/sub-account-secret` 定义组合类型 `BizSubAccountSecretJoinExtDetail`，内嵌或持有现有 join 明细类型并增加 `Operable bool`（JSON `operable`），避免修改 core 或 data-service 协议中的「纯资源」结构。
- 从列表结果中收集 `account_id`，批量调用已有 account logic，再映射到各行。

**Non-Goals:**

- 不改变 data-service join 查询 SQL、filter 或 IAM 列表鉴权规则（除响应组装外）。
- 不为其他 vendor 实现列表（除非当前代码路径已存在）；仅在与现有 `switch vendor` 一致的路径上返回新形态。
- 不在此变更中强制增加 `account_name`（用户仅要求 `operable`）；若后续产品与「子账号扩展列表」对齐再单独变更。

## Decisions

1. **复用 `logicaccount.BatchBuildOperableAndNameMap`**  
   **Rationale**: 与 `convertBizSubAccountExtList` 相同数据源与批处理，避免重复 Global Account 分页查询逻辑。仅需 `operableMap` 时可忽略 `accountMap` 的展示用途或按需保留以备扩展。

2. **响应类型放在 cloud-server API 包**  
   **Rationale**: `operable` 是「当前业务上下文下的派生字段」，不属于持久化实体；与 `BizSubAccountItem` 模式一致。

3. **封装转换函数**  
   **Rationale**: `ListSubAccountSecret` 内保持「鉴权 → 组 dsReq → 调 data-service → convert」清晰分层；转换函数接收 `bkBizID`、`operableMap` 或内部先调 BatchBuild，与 sub-account 的 `convertBizSubAccountExtList` / `buildBizSubAccountExtDetails` 命名风格对齐。

4. **TCloud 类型参数**  
   **Rationale**: 若 join 结果为泛型或 TCloud 专用 extension，转换函数可与现有 `ListSubAccountSecretJoinExt` 返回类型一致（需阅读 `pkg/api/data-service/cloud/sub_account_secret.go` 与 client 返回类型后实现）。

**Alternatives considered**

- 仅在 JSON 层用匿名 struct 拼接：类型不安全、不利于文档与 handler 复用 → 拒绝。
- 在 data-service 返回 `operable`：需把 `bk_biz_id` 语义注入数据层，且与「业务派生字段」分层不符 → 拒绝。

## Risks / Trade-offs

- **[Risk] 列表很大时额外一次 Global Account 批量查询**  
  **Mitigation**: 与 sub-account ext 相同模式；仅对去重后的 `account_id` 分批，已有 `BatchListBasicInfoByAccountIDs` 分页。

- **[Risk] 返回类型从「裸 join 结果」变为「包装结果」**  
  **Mitigation**: 对外 API 文档与前端约定新增字段；保持原有 join 字段不变，仅增加 `operable`。

## Migration Plan

- 部署 cloud-server 后新字段出现；旧客户端忽略未知字段即可，无数据迁移。

## Open Questions

- 若未来需要 `account_name`，可复用同一 `accountMap` 与 `BuildAccountNameMapByAccountMap`，本设计已预留与 sub-account 一致的路径。
