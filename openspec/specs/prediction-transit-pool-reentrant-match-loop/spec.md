# Spec: prediction-transit-pool-reentrant-match-loop

## Purpose

在 Phase 2（dispatcher `constructAddTransferAdjustReqParams`）中实现可重入匹配循环，通过循环执行"查询 → 贪心选单 → 拆单 → 轮询免审 → 重入"流程，应对拆出子单被并发请求抢占的竞态场景，并通过独立常量 `maxSplitRetry` 控制最大重试次数。

## Requirements

### Requirement: Phase 2 可重入匹配循环
在 Phase 2（dispatcher `constructAddTransferAdjustReqParams`）中，系统 SHALL 以循环方式执行"查询中转池 → 贪心选单 → 拆单 → 转移"流程，处理拆出子单被并发请求抢占的竞态场景。

循环行为：
- 每次迭代开始时，重置 `transferAbleDemands` 并重新调用 `queryTransferCRPDemands`，不复用上次缓存
- 所有 demands 统一执行贪心匹配，匹配结果写入临时 `matchResult`（不直接操作 `adjCRPDemandsRst`），demands 间通过 `matchResult` 共享消耗状态
- 若所有 demands 均贪心匹配成功（allMatched == true），将 `matchResult` 一次性 commit 到 `adjCRPDemandsRst`，跳出循环继续构建 CRP 请求
- 若有 demands 存在缺口，按 SliceId 合并各 demand 的 gap 后一次调用 `adjustOrder` 完成多路拆分（同一 SliceId 拆成 N+1 段），然后通过 `pollSplitOrderUntilApproved` 轮询 CRP 单据状态确认免审完成（轮询间隔 2s，超时 30s）后继续重入
- 拆单 `adjustOrder` 返回 `AdjustDemandIsInProcessingException` 时触发重入（continue），不直接 fail
- 最大重试次数为 `maxSplitRetry = 5`，超过后返回 error，子单标记 Failed

#### Scenario: 首次迭代贪心匹配成功，无需拆单
- **WHEN** 中转池候选核数之和 ≥ 业务总需求，且 ≤ 需求的候选可精确覆盖所有 demands
- **THEN** 系统在第 1 次迭代完成贪心匹配，跳出循环，不调用 adjustOrder

#### Scenario: 首次拆单后重入获取新 SliceId
- **WHEN** 首次迭代贪心不足，调用 adjustOrder 多路拆分成功，`pollSplitOrderUntilApproved` 确认免审通过后重入
- **THEN** 第 2 次迭代重新查询中转池，获取拆出的新 SliceId（每个 demand 的 gap 对应一个精确大小的块），贪心匹配成功，跳出循环

#### Scenario: 拆单后子单被并发请求抢占，触发重入
- **WHEN** adjustOrder 拆单成功，但 poll 等待期间新 SliceId 被其他并发 sub_ticket 抢先消费
- **THEN** 第 2 次迭代查询后该 SliceId 已被消耗，系统继续执行贪心 + 拆单逻辑，触发第 3 次迭代

#### Scenario: 拆单免审通过后通过 poll 及时感知
- **WHEN** 拆单 adjustOrder 成功，CRP 免审在 1~2s 内完成
- **THEN** `pollSplitOrderUntilApproved` 在首次 2s 轮询后即检测到 `PlanOrderStatusApproved`，立即继续重入，无需等待固定的 5~10s

#### Scenario: poll 超时
- **WHEN** 拆单 adjustOrder 成功，但 CRP 免审在 30s 内未完成
- **THEN** `pollSplitOrderUntilApproved` 返回超时 error，外层循环返回 error，子单标记 Failed

#### Scenario: adjustOrder 返回 InProcessing，触发重入
- **WHEN** 拆单 adjustOrder 返回 `AdjustDemandIsInProcessingException`
- **THEN** 系统执行 continue，开始下一次迭代（不经过 poll 等待）。重入后 `queryTransferCRPDemands` 返回最新状态，被锁定的 SliceId 带有 `IsInProcessing=1`，`greedyMatch` 将其过滤，不会重复命中同一个 InProcessing 的 SliceId

#### Scenario: 达到最大重试次数后 fail
- **WHEN** 连续 5 次迭代均未能完成贪心匹配（中转池持续被竞争耗尽）
- **THEN** 系统返回 error，子单标记 Failed，与现有失败行为一致

#### Scenario: 中转池确实不够时立即 fail
- **WHEN** 贪心后有 gap > 0 且无可拆候选（overCandidates 为空）
- **THEN** 系统立即返回 error，不继续重入，子单标记 Failed

### Requirement: maxSplitRetry 独立常量定义
系统 SHALL 定义独立常量 `maxSplitRetry = 5` 用于拆单竞态重试计数，与现有的 `CreateCrpTicketDefaultRetryTimes`（CRP 限流重试）语义分离，两者互不复用。

#### Scenario: 常量值固定为 5
- **WHEN** 系统初始化
- **THEN** `maxSplitRetry` 值为 5，不可被运行时配置覆盖

#### Scenario: 两个重试常量独立计数
- **WHEN** Phase 2 同时触发拆单竞态重试和 CRP 限流重试
- **THEN** 两个计数器相互独立，不共享状态
