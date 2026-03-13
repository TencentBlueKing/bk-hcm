# Spec: prediction-transit-pool-skip-inflight

## Purpose

在预测转移流程的 Phase 1（splitter）和 Phase 2（dispatcher 贪心匹配）中，通过 `IsInProcessing` 字段过滤正处于 CRP 审批流程中的候选预测，防止将锁定中的预测纳入转移计算。

## Requirements

### Requirement: Phase 1 跳过流程中的预测
在 Phase 1（splitter `matchTransferCRPDemands`）中，系统 SHALL 通过 `IsInProcessing` 字段检测候选预测是否处于 CRP 审批流程中，若为 1 则跳过该预测，不纳入 transferableCore 计算。

此过滤条件与现有的 `ReviewStatus == Pending`（跳过未评审预测）为两个独立的过滤条件，互相不替代。

#### Scenario: Phase 1 中过滤流程中预测
- **WHEN** 中转池返回的候选预测中某条 `IsInProcessing == 1`
- **THEN** `matchTransferCRPDemands` 跳过该预测（continue），不将其核数计入 transferableCore

#### Scenario: IsInProcessing 与 ReviewStatus 双重过滤
- **WHEN** 候选预测中一条 `IsInProcessing == 1`，另一条 `ReviewStatus == Pending`，其余正常
- **THEN** 两条均被跳过，只有状态正常的预测参与 transferableCore 计算

#### Scenario: 所有候选均流程中，Phase 1 判断不可转移
- **WHEN** 中转池所有候选预测均有 `IsInProcessing == 1`
- **THEN** 可用候选为空，transferableCore 为 0，Phase 1 判定不可转移，所有需求走常规追加流程

### Requirement: Phase 2 贪心算法中过滤流程中预测
在 Phase 2（dispatcher 贪心匹配函数 `greedyMatch`）中，系统 SHALL 在构建候选列表时过滤掉 `IsInProcessing == 1` 的预测，确保不会将正在审批流程中的预测纳入转移请求。

#### Scenario: Phase 2 贪心匹配时跳过流程中预测
- **WHEN** `queryTransferCRPDemands` 返回的候选中某条 `IsInProcessing == 1`
- **THEN** `greedyMatch` 构建候选列表时排除该预测，不消耗其核数

#### Scenario: Phase 2 跳过流程中预测后贪心不足
- **WHEN** 过滤 `IsInProcessing == 1` 后可用核数不足以覆盖需求
- **THEN** `greedyMatch` 按正常贪心算法找拆单目标或返回 (gap, nil)，由外层循环决定重入或 fail

#### Scenario: 并发场景下 Phase 2 动态感知预测锁定状态
- **WHEN** 可重入循环的第 N 次迭代中，某条预测在上次迭代后变为 `IsInProcessing == 1`（被其他并发请求锁定）
- **THEN** 重新查询后该预测被过滤，系统继续贪心匹配剩余候选
