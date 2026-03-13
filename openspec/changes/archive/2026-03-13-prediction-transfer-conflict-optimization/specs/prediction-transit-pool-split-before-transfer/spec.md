## ADDED Requirements

### Requirement: 转移前精确选单——贪心降序组合算法
转移前，系统 SHALL 对候选中转池预测执行贪心降序组合算法，确保最终被转移的所有 CRP 单核数之和与业务需求核数精确匹配（允许 ±2 核浮点误差）。

算法步骤：
1. 过滤掉 `IsInProcessing == 1` 和 `ReviewStatus == Pending` 的候选预测，以及 ObsProject / TechnicalClass 不匹配的候选
2. 在剩余候选中，按核数降序排列，优先选取核数 ≤ 剩余需求的预测，贪心累积直到需求满足
3. 若贪心后仍有剩余需求（gap > 0），在所有核数 > 剩余需求的候选中选取核数最小的一条记录为拆单目标，将 gap 作为 WillConsume 预扣到 matchResult（确保后续 demand 看到的是真实剩余），返回 (gap, splitTarget)
4. 若贪心后无剩余需求，返回 (0, nil) 表示精确匹配成功
5. 若既无法贪心满足、又无可拆候选，返回 (gap, nil) 表示中转池不足

#### Scenario: 候选预测核数之和恰好等于需求
- **WHEN** 中转池有两条预测分别为 8 核和 8 核，业务需求 16 核
- **THEN** 贪心算法选中两条，总核数 == 16，gap 为 0，无需拆单

#### Scenario: 需优先选大的候选减少 SliceId 数量
- **WHEN** 中转池有三条预测分别为 16 核、8 核、4 核，业务需求 20 核
- **THEN** 系统先选 16 核，再选 4 核，共 2 条，总核数 == 20，不选 8+4+4 以减少 SliceId 数量

#### Scenario: 贪心不足需标记拆单目标并预扣
- **WHEN** 中转池仅有一条 32 核预测，业务需求 16 核
- **THEN** 贪心无法用 ≤16 核候选满足，系统找到最小超出候选（32 核）作为拆单目标，将 gap=16 作为 WillConsume 预扣到 matchResult，返回 gap=16, splitTarget=32核预测。后续 demand 看到该 SliceId 的 remain 为 32-16=16

#### Scenario: splitTarget 预扣防止 gap 总和超出容量
- **WHEN** 中转池仅有一条 32 核预测，demand A 需求 20 核，demand B 需求 20 核
- **THEN** demand A 选中 32核预测为 splitTarget，预扣 gap=20（remain 变为 12）。demand B 在 Step 1 消耗剩余 12 核后仍有 gap=8 且无可拆候选（remain 已为 0），返回 (8, nil)，外层判定中转池不足。32 核确实无法满足 40 核总需求，fail 是正确行为

#### Scenario: 核数差在误差范围内视为精确匹配，跳过预拆单
- **WHEN** 候选预测剩余核数与需求之差 ≤ `TransferCoreToleranceThreshold`（常量 = 2）核
- **THEN** 视为精确匹配，不触发预拆单（`splitAdjustOrder`），在最终 transfer `adjustOrder` 中直接从该候选消耗 needCores 核数给业务、剩余部分留在中转池。业务精确获得 needCores，不存在多转或少转

#### Scenario: 中转池不足时无拆单目标
- **WHEN** 中转池候选全部 ≤ 已满足核数，贪心后仍有剩余需求，且无 > 剩余需求的候选
- **THEN** 返回 (gap, nil)，外层判定中转池不足，按现有失败行为处理

### Requirement: 转移前拆单——以中转池产品身份调用 adjustOrder
当贪心算法识别到需要拆单时，系统 SHALL 以中转池产品身份（`TransferPlanProductName` / `TransferOpProductName`）调用 `adjustOrder` 接口拆分大预测。支持多路拆分：当同一 SliceId 被多个 demand 选为 splitTarget 时，按 SliceId 合并各 demand 的 gap，拆成 N+1 段（每个 demand 的 gap 各一段 + 剩余段），使重入后各 demand 可在 Step 1 精确消耗对应块。

#### Scenario: 单个 demand 拆单（两段）
- **WHEN** 需拆单目标为 32 核，gap 为 16 核
- **THEN** 系统以中转池产品身份发起 adjustOrder，srcData 包含 32 核 SliceId，updatedData 包含 16 核子单 + 16 核余量子单，CRP 自动免审。拆单前后核数严格一致（srcData.CoreAmount == sum(updatedData.CoreAmount)），实现中须加断言校验

#### Scenario: 多个 demand 选中不同 SliceId 时合并为一次请求
- **WHEN** 同一 sub_ticket 内两个 demand 各自存在缺口，涉及不同 SliceId
- **THEN** 系统将两个拆单目标合并为一次 adjustOrder 请求，srcData 包含两个 SliceId，updatedData 包含对应拆后子单

#### Scenario: 多个 demand 选中同一 SliceId 时多路拆分
- **WHEN** 同一 sub_ticket 内 demand A(gap=20) 和 demand B(gap=10) 均选中 SliceId_X(36核) 作为 splitTarget
- **THEN** 系统按 SliceId 合并 gap，将 SliceId_X 拆成三段：[20核, 10核, 6核]，每段使用唯一 UUID 作为 SliceId，拆单前后核数严格一致（36 == 20+10+6）。重入后 demand A 在 Step 1 消耗 20核段，demand B 消耗 10核段，通常 2 次迭代完成

#### Scenario: adjustOrder 返回 AdjustDemandIsInProcessingException
- **WHEN** 拆单 adjustOrder 返回 `AdjustDemandIsInProcessingException` 错误
- **THEN** 系统不直接 fail，而是触发可重入循环重试（continue）

#### Scenario: adjustOrder 返回其他错误
- **WHEN** 拆单 adjustOrder 返回非 InProcessing 的错误
- **THEN** 系统返回 error，子单标记 Failed，与现有异常处理行为一致

### Requirement: 仅技术运营部（1041）可走转移逻辑
系统 SHALL 在 Phase 1（splitter `prepareAddSubTickets`）入口处增加部门前置检查。当 ticket 的 `VirtualDeptID` 不等于 `cvmapi.CvmCbsPlanDeptId`（1041）时，直接设置 canTransfer = false，跳过所有转移判断，所有需求走常规的追加、调整和删除流程。

#### Scenario: 非技术运营部 ticket 跳过转移
- **WHEN** ticket.VirtualDeptID 为 1042（非技术运营部）
- **THEN** prepareAddSubTickets 提前返回 false，不进入 canTransferByQuota 及后续转移逻辑

#### Scenario: 技术运营部 ticket 正常走转移逻辑
- **WHEN** ticket.VirtualDeptID 为 1041（技术运营部）
- **THEN** prepareAddSubTickets 继续执行 canTransferByQuota 及后续转移判断
