## Context

### 现有流程

预测追加单审批通过后，异步 dispatcher 分两个阶段处理：

```
Phase 1: createSubTicket (splitter)
  ticket → SplitAddTicket
    ├─ queryTransferCRPDemands → 拉中转池可用预测
    ├─ matchTransferCRPDemands → 匹配可转移核数
    └─ splitDemandInAddScenarios
         ├─ Transfer 子单 (transferableCore)
         └─ Add 子单      (nonTransferableCore)

Phase 2: createCrpTicket (dispatcher/CrpTicketCreator)
  sub_ticket → CreateCRPTicket
    Transfer 子单 → constructAddTransferAdjustReqParams
      ├─ queryTransferCRPDemands（独立重查 CRP）
      ├─ prePrepareTransferableData（精确匹配 SliceId）
      └─ adjustOrder → CRP 提单
    Add 子单 → createAddCrpTicket → AddCvmCbsPlan
```

### 现有问题

- Phase 2 `prePrepareTransferableData` 贪心匹配 CRP SliceId 时，会消耗整个大预测（如 32 核），即使业务只需 16 核。CRP 对整个 SliceId 加锁，其他业务被阻塞。
- Phase 1 和 Phase 2 均未检查 `IsInProcessing` 字段，导致已处于审批流程中的预测也会被选中，产生冲突。
- Phase 2 中转移不够时直接 `return error` → 子单 Failed，无重试。

### 涉及文件

| 文件 | 角色 |
|------|------|
| `cmd/woa-server/logics/plan/splitter/add.go` | Phase 1 拆子单，`matchTransferCRPDemands` |
| `cmd/woa-server/logics/plan/dispatcher/crp_adjust.go` | Phase 2 CRP 提单，`constructAddTransferAdjustReqParams` / `prePrepareTransferableData` |
| `cmd/woa-server/logics/plan/dispatcher/crp_query.go` | CRP 查询，`queryTransferCRPDemands` |
| `cmd/woa-server/logics/plan/dispatcher/crp_transfer.go` | 转移提单，`createTransferOutCrpTicket` / `constructTransReq` |
| `cmd/woa-server/logics/plan/dispatcher/crp_create.go` | 子单处理入口，`createCrpTicket` |
| `pkg/thirdparty/cvmapi/constvar.go` | 中转池产品常量 `TransferPlanProductName` 等 |

## Goals / Non-Goals

**Goals:**

- 缩小中转池转移时 CRP 锁的冲突域：被转移的 CRP 单核数之和 == 业务需求核数（±1~2 核误差）
- 跳过处于流程中的预测（`IsInProcessing == 1`），避免卡单
- 拆出的子单被抢时可重入重试，提升转移成功率

**Non-Goals:**

- 不引入"转移失败降级为追加"的新实体——不够就 fail，保持现有行为
- 不修改对外 API，仅改动异步任务内部逻辑
- 不改动 Phase 1（splitter）的拆分算法，仅增加 `IsInProcessing` 过滤和部门前置检查

## Decisions

### D1: Phase 1 仅加过滤，复杂逻辑下沉到 Phase 2

**选择**：Phase 1（splitter `matchTransferCRPDemands`）只增加 `IsInProcessing == 1` 过滤，不做贪心组合、不做拆单、不做重入循环。

**理由**：
- Phase 2 的 `constructAddTransferAdjustReqParams` 独立重查 CRP（L242-243），不复用 Phase 1 缓存
- Phase 1 本质上只决定 transferableCore / nonTransferableCore 的拆分比例，精确匹配放在 Phase 2 不影响正确性
- 改动集中在一处，降低回归风险

**改动点**：

1. `splitter/add.go` `prepareAddSubTickets` 入口处增加部门前置检查（详见 D6）
2. `splitter/add.go` `matchTransferCRPDemands` 循环内增加：

```go
if transAbleD.IsInProcessing == 1 {
    continue
}
```

### D2: Phase 2 引入贪心组合 + 拆单 + 可重入循环

**选择**：改造 `constructAddTransferAdjustReqParams`，在查询中转池后对所有 demands 统一执行贪心匹配，汇总缺口后一次 `adjustOrder` 拆分所有需拆的 SliceId，被抢时重入。

**核心原则**：一个 sub_ticket 中所有 demands 共享消耗状态，最终合并为一次 CRP adjustOrder 提单。拆单也遵循同样原则——先让所有 demands 完成贪心匹配，再汇总缺口统一拆单，而非逐个 demand 拆单。

**匹配状态与提交状态分离**：`greedyMatch` 将匹配结果写入临时的 `matchResult`（local map），不直接操作 `adjCRPDemandsRst`。所有 demands 匹配完毕且 allMatched 后，一次性将 `matchResult` commit 到 `adjCRPDemandsRst`。好处：(1) 每次循环迭代中若某个 demand 匹配失败需要拆单重入，`adjCRPDemandsRst` 始终干净，无 partial 状态；(2) `greedyMatch` 可作为近纯函数，更易测试。

**与 `constructAdjustDemandDetails` 的关系**：改造后的循环直接调用 `greedyMatch`，不再经由 `constructAdjustDemandDetails` 壳。安全性分析：transfer-add 场景下 `constructAdjustDemandDetails` 的唯一实际副作用是调 `prePrepareTransferableData`（`demand.Original == nil` + `adjustType == Transfer` 分支），其余分支（`prePrepareAdjustAbleData` 因 `demand.Original == nil` 不触发；`constructAdjustAppendData` 因 transfer 类型在 L584-586 被跳过）在 transfer-add 路径上均不执行，因此绕过 `constructAdjustDemandDetails` 不影响正确性。

**改动点**：`crp_adjust.go` `constructAddTransferAdjustReqParams` 重构为循环结构：

```
func constructAddTransferAdjustReqParams(kt, subTicket):
    for retry := 0; retry < maxSplitRetry; retry++:
        // 每次迭代重新查询并重置状态
        reset(c.transferAbleDemands)
        queryTransferCRPDemands(kt, obsProjects, technicalClasses)

        // 所有 demands 统一贪心匹配，使用临时 matchResult
        matchResult = {}  // 临时匹配状态，不直接写 adjCRPDemandsRst
        splitTargets = []
        allMatched = true
        for each demand in subTicket.Demands:
            gap, target = greedyMatch(kt, demand, matchResult)
            if gap > 0 && target != nil:
                splitTargets = append(splitTargets, {target, gap})
                allMatched = false
            else if gap > 0 && target == nil:
                return error  // 确实不够，和现有行为一致

        if allMatched:
            commit(matchResult → c.adjCRPDemandsRst)  // 一次性提交
            break  // 成功，跳出重入循环

        // 有缺口，按 SliceId 合并 splitTargets，一次 adjustOrder 多路拆分
        err, orderSN = splitAdjustOrder(kt, mergeSplitTargets(splitTargets))
        if err is InProcessing:
            continue  // 被抢，触发重入（重入后 requery 会过滤掉 InProcessing 的 SliceId）
        if err != nil:
            return error
        // 轮询 CRP 单据状态，等待免审完成（替代固定 sleep）
        err = pollSplitOrderUntilApproved(kt, orderSN)
        if err != nil:
            return error
        continue  // 拆单已审批通过，重入以获取新 SliceId

    if retry >= maxSplitRetry:
        return error

    构建 srcData + updatedData (现有逻辑)
```

### D3: 贪心组合选单算法

**选择**：在 `prePrepareTransferableData` 基础上改造匹配逻辑。改造后的 `greedyMatch` 替换现有 `prePrepareTransferableData`，在重入循环内直接调用（不经由 `constructAdjustDemandDetails` 壳，安全性分析详见 D2）。`greedyMatch` 将匹配结果写入临时 `matchResult`，而非直接操作 `adjCRPDemandsRst`。

**现有逻辑**：按顺序遍历 `transferAbleDemands`，`canConsume = min(need, remain)`，贪心吃掉尽可能多的核数。

**新逻辑**：`greedyMatch` 负责贪心匹配、记录缺口，并在 Step 2 选定 splitTarget 时将 gap 作为 WillConsume 预扣到 matchResult（防止后续 demand 重复选择同一 SliceId 导致 gap 总和超出实际容量）。拆单由 D2 中的外层循环统一处理。

```
func greedyMatch(kt, demand, matchResult) -> (gap int64, splitTarget *CvmCbsPlanQueryItem):
    候选 = transferAbleDemands[expectYear] 中过滤：
        - IsInProcessing == 1 → 跳过
        - ReviewStatus == Pending → 跳过
        - ObsProject / TechnicalClass 不匹配 → 跳过
        - 已被其他 demand 占用（matchResult 中已消耗）→ 扣减 remain

    needCores = demand.Updated.Cvm.CpuCore

    // Step 1: 优先选 ≤ needCores 的候选，按核数降序
    sortDesc(候选, by remain)
    for each 候选:
        if 候选.remain <= 0: continue
        if 候选.remain <= needCores:
            选中该候选（整单，记录到 matchResult），needCores -= 候选.remain
        if needCores <= 0: break

    // Step 2: needCores > 0，贪心不够，找最小的 > needCores 的候选
    if needCores > 0:
        overCandidates = 候选中 remain > needCores 的，按核数升序
        if len(overCandidates) == 0:
            return needCores, nil  // 确实不够，无可拆候选
        // 误差范围内视为精确匹配（详见下方核数误差处理）
        if overCandidates[0].remain - needCores <= TransferCoreToleranceThreshold:
            选中该候选（消耗 needCores，记录到 matchResult）
            return 0, nil
        // 预扣：将 gap 作为 WillConsume 记录到 matchResult，确保后续 demand 看到的是扣减后的 remain
        记录 overCandidates[0] 到 matchResult，WillConsume += needCores
        return needCores, overCandidates[0]  // 记录缺口和拆单目标

    // Step 3: 全部匹配成功
    return 0, nil
```

**Step 2 预扣机制**：当 `greedyMatch` 选定某个 SliceId 作为 splitTarget 时，将本次 demand 的 gap 作为 WillConsume 预扣到 matchResult。这保证了后续 demand 看到该 SliceId 的 remain 是真实剩余值（`RealCoreAmount - 已预扣总量`），从而：(1) 避免多个 demand 对同一 SliceId 的 gap 总和超出其实际容量；(2) 后续 demand 可能转向 Step 1 消耗该 SliceId 的剩余部分（当剩余 ≤ needCores 时），减少不必要的拆单。

**Step 1 降序排列的理由**：优先选大的候选，可以用更少的 SliceId 凑够需求核数。使用的 SliceId 越少，CRP adjustOrder 中的 srcData 条目越少，锁冲突面越小。同时减少触发 Step 2 拆单的概率。

**同一 sub_ticket 内 demand 无需排序**：多个 demand 共享 `adjCRPDemandsRst` 消耗状态。若中转池总量不够，无论哪个 demand 先匹配都会最终 fail；若总量够，顺序只影响拆单次数（差异极小），不影响正确性。真正的争抢发生在不同 sub_ticket 的并发之间，由重入循环解决。

**核数误差处理**：当候选核数与需求之差 ≤ `TransferCoreToleranceThreshold`（常量，值为 2）核时，视为精确匹配，不触发预拆单（`splitAdjustOrder`），而是在最终的 transfer `adjustOrder` 中直接从该候选调减 needCores 给业务、剩余部分留在中转池。误差容忍省去的是额外的拆单 round-trip，转移精度不变——业务始终精确获得 needCores 核数。

### D4: adjustOrder 拆单使用中转池产品身份

**选择**：拆单请求的 `PlanProductName` / `ProductName` 使用 `cvmapi.TransferPlanProductName` / `cvmapi.TransferOpProductName`。

**理由**：中转池的预测归属于中转产品，拆单操作也在中转产品上下文内进行。CRP 识别到是中转池内部拆单会自动免审。

**参数构造**：复用现有 `CvmCbsPlanAdjustReq` 结构，`srcData` 为拆分前的完整 SliceId。`updatedData` 支持多路拆分——当同一 SliceId 被多个 demand 选为 splitTarget 时，按 SliceId 合并各 demand 的 gap，拆成 N+1 段（每个 demand 的 gap 各一段 + 剩余段）。例如 SliceId_X(36核) 被 demand A(gap=20) 和 demand B(gap=10) 同时选中，则 `updatedData` 包含三条：[20核, 10核, 6核]。每个 updatedData 条目使用唯一 UUID 作为 SliceId。CRP 要求拆单前后核数严格一致（`srcData[i].CoreAmount == sum(对应 updatedData.CoreAmount)`），不允许拆分过程中追加或减少核数。实现中须加断言校验，防止构造参数 bug 导致 CRP 拒绝请求。

**多路拆分的必要性**：若将多个 demand 的 gap 合并为单块（如 20+10=30），拆出的 30 核块比任何单个 demand 的需求都大，下一轮迭代的 Step 1 无法消耗，Step 2 又会触发新的拆单请求，形成级联拆单（cascading splits），可能耗尽 `maxSplitRetry` 导致本不应失败的场景失败。多路拆分（拆成 [20, 10, 6]）使重入后每个 demand 在 Step 1 即可精确消耗对应的块，通常只需 2 次迭代即可完成。

### D5: 拆单 adjustOrder 的 InProcessing 处理

**选择**：中转池内拆单 `adjustOrder` 返回 `AdjustDemandIsInProcessingException` 时，触发重入循环（continue），不直接 fail。

**理由**：拆单 `adjustOrder` 与"查询到 `IsInProcessing=0`"之间存在时间窗口，其他并发请求可能在窗口内对目标 SliceId 发起操作。此时重入循环重新查询即可获取最新状态。重入后 `queryTransferCRPDemands` 返回的数据中，被锁定的 SliceId 会带有 `IsInProcessing=1`，`greedyMatch` 会将其过滤掉，因此不会重复命中同一个 InProcessing 的 SliceId——系统会转向匹配其他可用候选。

**与外层 `createCrpTicket` 的关系**：最终转移 adjustOrder（非拆单）碰到 InProcessing 时，保持现有行为——`createCrpTicket` catch 该错误后 return nil，子单留在队列等待重试。两层处理互不干扰。

### D6: 仅技术运营部（1041）可走转移逻辑

**选择**：三个场景（add/adjust/delete）均需在入口处增加部门前置检查。当 `ticket.VirtualDeptID != cvmapi.CvmCbsPlanDeptId`（1041）时，跳过整个中转逻辑，所有需求走常规流程。

**理由**：中转池属于技术运营部（1041）的专属资源，非该部门的业务不得进入。

- **Add/Adjust 方向（中转池 → 本业务）**：若允许非 1041 部门消耗中转池，会侵占本属于 1041 的资源，产生资源污染。
- **Delete 方向（本业务 → 中转池）**：若允许非 1041 部门的预测流入中转池，中转池内会混入其他部门的预测，导致"部门污染"——1041 之后从中转池取资源时可能拿到非本部门的预测，引发错误转移。

因此 delete 场景同样必须限制在 1041 部门内，与 add/adjust 对称。

**改动点**：

1. `splitter/add.go` `prepareAddSubTickets`（覆盖 add / adjust 复用场景）入口处增加：

```go
if virtualDeptID != cvmapi.CvmCbsPlanDeptId {
    return false, cvmDemands, nil
}
```

2. `splitter/delete.go` `prepareDeleteSubTickets` 入口处增加：

```go
if virtualDeptID != cvmapi.CvmCbsPlanDeptId {
    // 非 1041 部门：直接走常规删除流程，不进入中转池
    return s.regularDeleteFlow(kt, ticketID, ticketType, demands, planProductName, opProductName, deviceTypeMap)
}
```

对应的，`SplitDeleteTicket` 签名需增加 `virtualDeptID int64` 参数，调用方 `dispatcher/sub_ticket.go` 和 `plan/sub_ticket.go` 传入 `ticket.VirtualDeptID`。

### D7: 失败行为不变

**选择**：贪心+重入循环耗尽后若仍不够，`prePrepareTransferableData` 仍返回 `"crp demand remained is not enough to deduction"` error，子单标记 Failed。

**理由**：如无必要，勿增实体。Phase 1 已经排除了 `IsInProcessing` 后判断"够"，Phase 2 执行时仅在极端竞态下才会"不够"，此时 fail 是合理的。

### D8: matchResult 复用 `map[string]*AdjustAbleRemainObj`

**选择**：`greedyMatch` 使用的临时 `matchResult` 直接复用 `adjCRPDemandsRst` 的类型 `map[string]*AdjustAbleRemainObj`（key 为 SliceId），不引入新的数据结构。

**理由**：
- `AdjustAbleRemainObj` 已有的 `WillConsume`、`TransferTarget`、`OriginDemand` 三个字段精确满足 `greedyMatch` 的需求：跨 demand 共享消耗追踪、记录转移目标、保存原始预测快照
- commit 操作退化为一行赋值 `c.adjCRPDemandsRst = matchResult`，无需额外转换函数，杜绝字段映射遗漏风险
- `AdjustType`/`ExpectTime` 字段在 transfer-add 路径上为 zero-value 且不被读取，不影响正确性
- 丢弃操作（重入时）仅需忽略该 map，下次迭代新建即可

**误传防护**：`matchResult` 与 `c.adjCRPDemandsRst` 类型相同，存在误传风险。通过以下措施防护：
1. `greedyMatch` 不作为 `CrpTicketCreator` 的方法，而是 package-level 函数，签名中不携带 receiver `c`，物理上隔绝对 `c.adjCRPDemandsRst` 的直接访问
2. `greedyMatch` 同时接收 `transferAbleDemands` 参数（只读），不通过 receiver 隐式访问
3. 外层循环中 `matchResult` 在每次迭代内通过 `make(map[string]*AdjustAbleRemainObj)` 新建，`adjCRPDemandsRst` 仅在 commit 时被赋值

**TransferTarget 写入格式**：`greedyMatch` 选中候选时，须以 `cvt.PtrToVal(demand.Updated.Clone())` 作为 `TransferTarget` 的 key，value 为该 SliceId 分摊给该 demand 的核数。这与现有 `prePrepareTransferableData`（L662）的写法完全一致。下游 `constructTransferAppendDataToBiz`（L878）遍历 `TransferTarget` 时会从 key 中读取 `ObsProject`、`RegionName`、`ZoneName`、`ExpectTime`、`Cvm.DeviceFamily`、`Cbs.DiskType` 等业务字段来构造转移到业务的 updatedData，因此 key 的完整性至关重要。

**下游衔接验证**：commit 后下游构建 `srcData + updatedData`（L258-305）从 `adjCRPDemandsRst` 读取的字段与 `greedyMatch` 写入的字段对齐情况：
- `OriginDemand`：下游用于 Clone() 构建 srcItem/updatedItem → `greedyMatch` 通过 `transAbleD.Clone()` 填充，一致
- `WillConsume`：下游用于计算 willChangeCvm 和 CoreAmount 调减 → `greedyMatch` 累加填充，一致
- `TransferTarget`：下游传给 `constructTransferAppendDataToBiz` → `greedyMatch` 按上述格式填充，一致
- `AdjustType`/`ExpectTime`：下游在 transfer-add 路径上不读取，zero-value 安全

**函数签名**：

```go
func greedyMatch(
    kt *kit.Kit,
    demand rpt.ResPlanDemand,
    candidates []*cvmapi.CvmCbsPlanQueryItem,
    matchResult map[string]*AdjustAbleRemainObj,
) (gap int64, splitTarget *cvmapi.CvmCbsPlanQueryItem, err error)
```

## Risks / Trade-offs

**[风险] adjustOrder 拆单后 CRP 返回的新 SliceId 需要通过重新查询才能获取**
→ 每次重入循环开头重新 `queryTransferCRPDemands`，不缓存旧结果。`CrpTicketCreator` 的 `transferAbleDemands` 和 `adjCRPDemandsRst` 在每次循环迭代开始时清空重置。

**[风险] 重入循环对 CRP 查询接口产生额外压力**
→ 最大重试次数限制为 5 次（`maxSplitRetry = 5`）；每次拆单后通过 `pollSplitOrderUntilApproved` 等待审批完成（轮询间隔 2s，超时 30s），正常情况下不会重入（免审过单极快），仅竞态时触发。

**[风险] Phase 1 粗判"够"但 Phase 2 发现"不够"（时间窗口竞态）**
→ Phase 1 已过滤 `IsInProcessing`，两阶段间时间窗口极短。极端情况下 Phase 2 fail 是可接受的（保持现有行为，不新增降级）。

**[风险] 拆单 adjustOrder 本身可能失败（CRP 异常）**
→ 区分错误类型：若 CRP 返回 `AdjustDemandIsInProcessingException`，触发重入循环（详见 D5）；其他错误直接 return error，子单 Failed，与现有异常处理行为一致。

**[权衡] 贪心算法不保证全局最优拆单方案**
→ 贪心降序+最小超出单拆分在绝大多数实际场景下已足够（中转池 SliceId 数量有限）。无需引入复杂的背包算法。

## Resolved Questions

- **adjustOrder 免审时延**：CRP 免审过程是异步的，虽然通常很快但不可假设同步完成。拆单 adjustOrder 成功后，通过 `pollSplitOrderUntilApproved` 轮询 `QueryPlanOrder` 接口确认单据状态（轮询间隔 2s，超时 30s）。相比固定 sleep(5~10s)，poll 策略在免审快时可更早进入下一步（通常 1~2s 即完成），在偶发慢处理时不会因 sleep 太短查到旧数据，且驳回场景可立即感知并 fail。现有代码 `checkCrpTicket`（`sub_ticket.go`）已有成熟的 `QueryPlanOrder` 调用模式可复用。
- **maxSplitRetry 配置**：新增独立的 `maxSplitRetry` 常量，固定值为 5，不复用 `CreateCrpTicketDefaultRetryTimes`。两者语义不同——前者是竞态拆单重试，后者是 CRP 限流重试。正常情况下 1~2 次迭代即可完成，5 次留有充足余量应对极端竞态。
- **拆单碎片**：adjustOrder 拆出的小 SliceId 若未被转移（如后续失败），会留在中转池中成为碎片。无需专门清理机制，碎片会随后续转移请求被自然消耗，随时间收敛。
- **部门限制**：仅技术运营部（`VirtualDeptID == 1041`）可走转移逻辑。该常量已存在于 `cvmapi.CvmCbsPlanDeptId`。
- **CRP adjustOrder 支持一次请求拆分多个 SliceId**：当多个 demand 的缺口涉及不同 SliceId 时，可合并为一次 adjustOrder 请求，`srcData` 包含多个待拆 SliceId，`updatedData` 包含对应的拆后结果。CRP 要求拆分前后核数严格一致（srcData 和 updatedData 核数之和相等），不允许拆分过程中追加或减少核数。
- **同一 SliceId 多路拆分**：当同一 SliceId 被多个 demand 选为 splitTarget 时，`splitAdjustOrder` 按 SliceId 合并各 demand 的 gap，拆成 N+1 段（每个 demand 的 gap 各一段 + 剩余段）。例如 SliceId_X(36核) 被 demand A(gap=20) 和 demand B(gap=10) 同时选中，拆成 [20, 10, 6] 三段。greedyMatch 的 Step 2 预扣机制保证合并后的 gap 总和不会超出 SliceId 的 RealCoreAmount。若仅拆成合并后的单块（30+6），该块比任何单个 demand 都大，下一轮 Step 1 无法消耗，会触发级联拆单（cascading splits），可能耗尽 maxSplitRetry。多路拆分使重入后各 demand 在 Step 1 即可精确消耗对应块，通常 2 次迭代完成。
- **`constructAdjustDemandDetails` 绕过安全性**：在 transfer-add 场景下，`constructAdjustDemandDetails` 的唯一实际副作用是调 `prePrepareTransferableData`（`demand.Original == nil` + `adjustType == Transfer` 分支）。`prePrepareAdjustAbleData` 分支因 `demand.Original == nil` 不触发；`constructAdjustAppendData` 分支因 transfer 类型在 L584-586 被跳过。因此重入循环中直接调用 `greedyMatch` 替代，不经由 `constructAdjustDemandDetails`，不影响正确性。
- **核数误差容忍语义**：`TransferCoreToleranceThreshold`（常量 = 2）的含义是：当候选 remain 与 needCores 之差在阈值内时，跳过预拆单（`splitAdjustOrder`）round-trip，在最终 transfer `adjustOrder` 中直接从该候选消耗 needCores 核数。业务始终精确获得 needCores，不存在多转或少转。
- **CBS-only demand 无需特殊处理**：Phase 1 `separateAndProcessDemands`（`splitter/add.go` L142-145）在进入转移匹配之前已将 CBS-only demand（`Cvm.IsEmpty()`）分流至 `adjSplitGroupDemands[Add]`，仅 CVM demand 参与 `matchTransferCRPDemands` 和后续拆分。因此 transfer sub-ticket 的 demands 中不可能出现 CBS-only demand。现有 `prePrepareTransferableData` L612 的 `if demand.Updated.Cvm.IsEmpty()` 检查是纯防御性代码，正常流程不会触发。`greedyMatch` 中无需额外的 CBS-only 跳过逻辑——即使防御场景下 CBS-only 意外进入，`needCores = demand.Updated.Cvm.CpuCore = 0`，Step 1 循环不执行，直接返回 `(0, nil)`，行为天然正确。
