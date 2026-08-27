## Context

bk-hcm 已有 TCloud 安全组规则的批量修改接口（`batch/update`），底层使用腾讯云 `ReplaceSecurityGroupPolicies` API，每次只能更新单方向（出站或入站）的规则。业务场景需要一次性全量替换安全组的所有出入站规则，现有接口无法满足。

腾讯云提供了 `ModifySecurityGroupPolicies` API，支持在单次调用中原子性地替换安全组的全部规则，是本次实现的核心底层接口。

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 资源下和业务下各新增一个 TCloud 安全组规则全量覆盖接口
- 在 hc-service 新增对应 TCloud 接口，底层单次调用 `ModifySecurityGroupPolicies`，`Version` 不传（null），保证原子性
- 请求体同时支持出站（egress）和入站（ingress）规则，支持只传一个方向（另一方向会被清空）
- 覆盖成功后通过现有 `syncSGRule` 机制同步本地 DB

**Non-Goals:**
- cloud-server 路由设计预留多云扩展，但当前 hc-service 仅实现 TCloud；其他 vendor 暂返回 unsupported 错误，不在本次范围内
- 不修改现有 `batch/update` 接口行为
- 不引入分布式事务或补偿机制（单次 API 调用已保证原子性，无需额外处理）

## Decisions

### 1. 使用单次 ModifySecurityGroupPolicies 调用，Version 不传

**选择**：调用 `ModifySecurityGroupPolicies` 时 `SecurityGroupPolicySet.Version` 不赋值（null）。

**原因**：
- TCloud 在 Version 为 null 时跳过版本冲突校验，直接原子替换所有规则
- 无需提前拉取当前版本号（减少一次 DescribeSecurityGroupPolicies 调用）
- 避免分步调用（先 Version="0" 清空，再写入）带来的中间窗口期风险：若第二步失败，安全组规则永久丢失

**备选方案**：先以 `Version="0"` 清空再写入 → 已否决，存在不可接受的原子性风险。

### 2. 出站和入站规则均为必填，各自至少一条

**选择**：`egress_rule_set` 和 `ingress_rule_set` 均为必填字段，且各自至少包含一条规则，否则请求校验阶段直接返回参数错误。

**原因**：全量覆盖是高风险操作，漏传某个方向会导致该方向规则被静默清空。强制要求两个方向均非空，可以有效防止调用方因疏忽造成规则意外丢失。API 文档中须明确说明此约束。

### 3. 新增独立的 Option/Req 结构，不复用 TCloudUpdateOption

**选择**：新增 `TCloudOverwriteOption`（adaptor 层）和 `TCloudSGRuleOverwriteReq`（hc-service proto 层），不复用 `TCloudUpdateOption`。

**原因**：`TCloudUpdateOption` 包含 `Version` 字段（必填），覆盖接口语义上不需要 Version，强制复用会产生误导，且未来两者可能独立演进。

### 4. 复用现有 syncSGRule 全量同步本地 DB

**选择**：覆盖操作成功后，通过现有 `g.syncSGRule(kt, syncParam)` 对本地 DB 进行全量同步，不单独实现 DB 写入逻辑。

**原因**：`syncSGRule` 内部会对比云上最新规则与本地 DB 的差异，执行**删除已移除的旧规则 + 写入新规则**的全量替换，确保 DB 与云上完全一致。覆盖接口无需重复实现该逻辑。

## Risks / Trade-offs

- **Version=null 的云端行为依赖**：实现依赖 TCloud 在 Version 为 null 时跳过版本校验并原子替换的行为，需在对接测试中确认。若 TCloud 实际行为不符预期，可退回为先获取当前 version 再传入的方案。
- **并发写冲突**：Version=null 跳过版本校验，多个并发调用可能相互覆盖。调用方需自行保证串行化，平台层不做额外保护（与现有行为一致）。

## Migration Plan

纯新增接口，不影响现有接口，无迁移步骤，无需回滚策略。
