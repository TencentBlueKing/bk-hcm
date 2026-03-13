## Why

当用户提预测追加单时，若中转池存在可用预测，系统（异步任务中）会优先将中转池中的整个大预测转移给业务。转移过程会在 CRP 上对原预测整体加锁，导致其他业务无法并发使用该预测，审批流被迫串行等待。通过在转移前精确拆单、跳过流程中的预测并引入可重入匹配循环，可缩小 CRP 锁的粒度，提升多业务并发过单效率。

## What Changes

- **转移前精确选单与拆单**：发起转移前，须保证每条被转移的 CRP 单核数恰好等于该次转移所需核数（允许 ±1～2 核浮点误差）。当单条中转池预测核数超出需求（超出误差范围）时，以中转池产品身份调用 `adjustOrder` 接口将其拆分为"精确匹配需求的子单 + 剩余子单"，仅对匹配的子单发起转移。CRP 识别到是中转池内部拆单时会自动免审。

- **多单贪心组合算法**：当单条中转池预测不足以覆盖业务需求时，需要组合多条预测。选单策略如下：
  1. 过滤掉 `IsInProcessing == 1`（流程中）和 `ReviewStatus == Pending`（未评审）的预测；
  2. 在剩余候选中，优先选取核数 ≤ 剩余需求的预测，按核数降序贪心累积；
  3. 若贪心后仍有剩余需求，则在所有 > 剩余需求的候选中选核数最小的一条进行拆单，拆出恰好等于剩余需求核数的子单；
  4. 最终所有被选中单据的核数之和 == 业务总需求核数。

- **转移时跳过流程中的预测**：通过 CRP 返回的 `IsInProcessing` 字段（`1` 表示该预测正处于审批流程中）识别并跳过锁定中的预测，避免等待已加锁预测造成卡单。这与现有的 `ReviewStatus == Pending`（跳过未评审）是两个独立的跳过条件。

- **可重入匹配循环**：上述拆单与转移过程在 Phase 2（dispatcher 创建 CRP 单据阶段）以循环方式运行。拆单完成后，目标子单可能被其他并发请求抢先转移，此时循环重入，重新查询中转池并重新执行选单算法。循环退出条件：
  - **成功**：完成转移，累计转移核数 == 需求核数；
  - **失败**：中转池中已无可用预测可继续匹配，或达到最大重试次数上限，子单报错失败（保持现有行为，不引入降级实体）。

- **容忍性设计**：业务 A 拆单期间若业务 B 发起转移，B 因 `IsInProcessing == 1` 直接跳过走追加；若 A 的提单最终被驳回，B 已走追加，存在"错失一次转移机会"的情况。该概率极低（免审过单极快），可容忍。

- **部门限制**：仅技术运营部（部门 ID 1041）的业务可走中转池转移逻辑。非该部门的 ticket 在 Phase 1 拆子单时直接跳过中转判断，所有需求走常规的追加、调整和删除流程。

## Capabilities

### New Capabilities

- `prediction-transit-pool-split-before-transfer`: 中转池转移前精确选单与拆单能力——以中转池产品身份调用 `adjustOrder` 将超出需求的大预测拆分为精确匹配的小预测，结合多单贪心组合算法（优先选 ≤ 需求中最大的，贪心累积，缺口则拆最小超出单），确保每次转移的 CRP 单核数之和恰好等于业务需求，最小化 CRP 锁的冲突域。

- `prediction-transit-pool-reentrant-match-loop`: 中转池匹配可重入循环能力——在 Phase 2（dispatcher）中以循环方式执行"查询中转池 → 过滤流程中预测 → 选单/拆单 → 转移"流程，处理拆出子单被并发请求抢占的竞态场景，直到成功转移或不够时报错失败（含最大重试次数保护，保持现有失败行为）。

- `prediction-transit-pool-skip-inflight`: 中转池转移时跳过流程中预测能力——在 Phase 1（splitter）和 Phase 2（dispatcher）中均通过 `IsInProcessing` 字段检测预测是否处于审批流程中，若是则跳过，避免串行等待造成卡单。

### Modified Capabilities

<!-- 暂无已存在的 spec 需要修改 -->

## Impact

- **woa-server / splitter（小改）**：`cmd/woa-server/logics/plan/splitter/add.go` 的 `prepareAddSubTickets` 增加部门前置检查（非技术运营部 1041 直接跳过转移），`matchTransferCRPDemands` 增加 `IsInProcessing == 1` 过滤条件，其余拆分逻辑不变
- **woa-server / dispatcher（主改动）**：`crp_adjust.go` 中 `constructAddTransferAdjustReqParams` 及 `prePrepareTransferableData` 为核心改造点，需增加 `IsInProcessing` 过滤、贪心组合选单、adjustOrder 拆单和可重入循环逻辑
- **CRP 接口调用**：新增以中转池产品身份调用 `adjustOrder` 接口（用于中转池内拆单）；拆单参数中传入目标核数与剩余核数两段
- **并发与竞态**：最大重试次数上限固定为 5 次（`maxSplitRetry = 5`），防止极端情况下无限循环；每次重入需重新查询中转池状态，不可复用上次缓存；拆单 adjustOrder 碰到 InProcessing 时触发重入而非直接 fail
- **无外部 API 变更**：本次优化为异步任务内部流程调整，不涉及对外接口变化
