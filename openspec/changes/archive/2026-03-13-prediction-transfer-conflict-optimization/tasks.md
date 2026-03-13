## 1. 常量与基础定义

- [x] 1.1 在 `pkg/criteria/constant/` 中新增 `maxSplitRetry = 5` 常量，用于拆单竞态重试计数
- [x] 1.2 确认 `cvmapi.CvmCbsPlanDeptId`（1041）常量已存在于 `pkg/thirdparty/cvmapi/constvar.go`，如不存在则添加
- [x] 1.3 在 `pkg/criteria/constant/` 中新增 `TransferCoreToleranceThreshold int64 = 2` 常量，用于贪心匹配时核数误差容忍阈值（差值 ≤ 该阈值时跳过预拆单）

## 2. Phase 1 改造（splitter）

- [x] 2.1 在 `cmd/woa-server/logics/plan/splitter/add.go` 的 `prepareAddSubTickets` 入口处增加部门前置检查：当 `ticket.VirtualDeptID != cvmapi.CvmCbsPlanDeptId` 时直接返回 `false, cvmDemands, nil`
- [x] 2.2 在 `matchTransferCRPDemands` 循环内增加 `IsInProcessing == 1` 过滤（`if transAbleD.IsInProcessing == 1 { continue }`）

## 3. Phase 2 贪心匹配函数（dispatcher）

- [x] 3.1 在 `cmd/woa-server/logics/plan/dispatcher/crp_adjust.go` 中新增 package-level 函数 `greedyMatch`（非 `CrpTicketCreator` 方法），签名为 `func greedyMatch(kt, demand, candidates []*CvmCbsPlanQueryItem, matchResult map[string]*AdjustAbleRemainObj) (gap int64, splitTarget *CvmCbsPlanQueryItem, err error)`。`matchResult` 直接复用 `adjCRPDemandsRst` 的类型 `map[string]*AdjustAbleRemainObj`，匹配结果写入 `matchResult`（不直接操作 `adjCRPDemandsRst`）。作为 package-level 函数可物理隔绝对 receiver `c` 中 `adjCRPDemandsRst` 的直接访问，防止误传
- [x] 3.2 `greedyMatch` 实现候选过滤：跳过 `IsInProcessing == 1`、`ReviewStatus == Pending`、ObsProject/TechnicalClass 不匹配、已被占用（扣减 remain）的候选
- [x] 3.3 `greedyMatch` 实现 Step 1：对 `remain <= needCores` 的候选按 remain 降序排列，贪心累积
- [x] 3.4 `greedyMatch` 实现 Step 2：若 needCores > 0，在 `remain > needCores` 的候选中选核数最小的作为 splitTarget 返回，同时将 gap 作为 WillConsume 预扣到 matchResult（确保后续 demand 看到真实剩余，防止多 demand 对同一 SliceId 的 gap 总和超出实际容量）
- [x] 3.5 `greedyMatch` 实现误差处理：候选 remain 与 needCores 之差 ≤ `TransferCoreToleranceThreshold` 时视为精确匹配，跳过预拆单，在最终 transfer adjustOrder 中直接消耗 needCores（业务精确获得 needCores，剩余留池）

## 4. Phase 2 拆单函数（dispatcher）

- [x] 4.1 新增 `mergeSplitTargets` 函数：按 SliceId 合并多个 demand 的 splitTarget，同一 SliceId 的多个 gap 聚合为列表
- [x] 4.2 新增 `splitAdjustOrder` 函数：接收合并后的 splitTargets，构造 `CvmCbsPlanAdjustReq`。支持多路拆分——同一 SliceId 被多个 demand 选中时拆成 N+1 段（每个 demand 的 gap 各一段 + 剩余段），每段使用唯一 UUID 作为 SliceId。须加断言校验 `srcData[i].CoreAmount == sum(对应 updatedData.CoreAmount)`，确保拆单前后核数严格一致。返回 (orderSN, error)
- [x] 4.3 `splitAdjustOrder` 使用 `cvmapi.TransferPlanProductName` / `cvmapi.TransferOpProductName` 作为产品身份
- [x] 4.4 错误处理：`AdjustDemandIsInProcessingException` 返回特定 sentinel error（用于外层判断是否重入），其他错误直接透传
- [x] 4.5 新增 `pollSplitOrderUntilApproved` 函数：接收 orderSN，通过 `QueryPlanOrder` 接口轮询 CRP 单据状态（轮询间隔 2s，超时 30s）。`PlanOrderStatusApproved` 时返回 nil；`PlanOrderStatusRejected` 时返回 error；超时返回 error。复用现有 `checkCrpTicket` 中的 `QueryPlanOrder` 调用模式

## 5. Phase 2 可重入循环（dispatcher）

- [x] 5.1 重构 `constructAddTransferAdjustReqParams`，将核心逻辑包裹在 `for retry := 0; retry < maxSplitRetry; retry++` 循环中
- [x] 5.2 每次迭代开始时清空重置 `transferAbleDemands`，重新调用 `queryTransferCRPDemands`
- [x] 5.3 循环内通过 `make(map[string]*AdjustAbleRemainObj)` 新建临时 `matchResult`，对 sub_ticket 所有 demands 依次调用 `greedyMatch(kt, demand, candidates, matchResult)`，收集所有 splitTargets 并判断 allMatched
- [x] 5.4 allMatched == true 时通过 `c.adjCRPDemandsRst = matchResult` 一次性 commit，break 并继续构建 CRP 请求（现有逻辑）
- [x] 5.5 有 splitTargets 时先调用 `mergeSplitTargets` 按 SliceId 合并，再调用 `splitAdjustOrder` 执行多路拆分，成功后调用 `pollSplitOrderUntilApproved` 等待免审完成并 continue
- [x] 5.6 `splitAdjustOrder` 返回 InProcessing sentinel error 时直接 continue（不经过 poll）
- [x] 5.7 `splitAdjustOrder` 返回其他错误时直接 return error
- [x] 5.8 循环耗尽（retry >= maxSplitRetry）时 return error，子单标记 Failed

## 6. 单元测试

- [x] 6.1 为 `greedyMatch` 编写单元测试：覆盖精确匹配、贪心降序、需要拆单、误差范围内匹配、中转池不足、多 demand 选同一 splitTarget 时预扣正确性、预扣后 gap 总和不超出容量等场景
- [x] 6.2 为 `mergeSplitTargets` 编写单元测试：覆盖不同 SliceId、同一 SliceId 合并、单个 demand 等场景
- [x] 6.3 为 `splitAdjustOrder` 编写单元测试：覆盖正常两段拆单、多路拆分（N+1 段）、多 SliceId 合并请求、InProcessing 返回特定 error、其他错误透传等场景
- [x] 6.4 为 `pollSplitOrderUntilApproved` 编写单元测试：覆盖首次 poll 即通过、多次 poll 后通过、驳回、超时等场景
- [x] 6.5 为重构后的 `constructAddTransferAdjustReqParams` 编写集成测试（mock CRP 接口）：覆盖首次成功、拆单后重入成功、多路拆分后重入、被抢重入、达到 maxSplitRetry 失败等场景
- [x] 6.6 为 Phase 1 `matchTransferCRPDemands` 的 `IsInProcessing` 过滤新增单元测试
- [x] 6.7 为 `prepareAddSubTickets` 的部门前置检查新增单元测试

## 7. 集成验证

- [x] 7.1 在本地/测试环境验证：非技术运营部 ticket 走常规追加流程（不进入转移分支）
- [x] 7.2 验证：中转池预测核数恰好等于需求时，不触发拆单
- [x] 7.3 验证：中转池预测核数超出需求时，拆单后转移正确核数
- [x] 7.4 验证：`IsInProcessing == 1` 的预测在 Phase 1 和 Phase 2 均被跳过
- [x] 7.5 验证：可重入循环在拆单成功后正确获取新 SliceId 完成转移
- [x] 7.6 验证：多个 demand 选中同一 SliceId 时，多路拆分后各 demand 在第 2 次迭代精确消耗对应块
- [x] 7.7 验证：拆单免审后 poll 及时感知，不出现不必要的长等待
