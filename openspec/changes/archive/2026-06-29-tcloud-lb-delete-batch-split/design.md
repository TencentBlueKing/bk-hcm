## 背景

### 现状

cloud-server 的负载均衡批量删除流程（`cmd/cloud-server/service/load-balancer/delete.go`）：

1. `batchDeleteLoadBalancer` 接收用户传入的 LB ID 列表，通过 `ListResBasicInfo` 获取每个 LB 的 `AccountID`、`Vendor`、`Region` 信息
2. 调用 `buildLBDeletionTasks(infoMap)` 按 `genAccountRegionKey(info)` = `AccountID_Vendor_Region` 分组
3. 每组生成一个 `ts.CustomFlowTask`，其 `Params` 为 `actionlb.DeleteLoadBalancerOption`（内含 `hcproto.BatchDeleteLoadBalancerReq.IDs`）
4. 所有 task 封装为 `ts.AddCustomFlowReq` 提交给 task-server 创建 custom flow

**问题根因**（`delete.go:213-242`）：

```go
func buildLBDeletionTasks(infoMap map[string]types.CloudResourceBasicInfo) (tasks []ts.CustomFlowTask) {
    reqMap := make(map[string]*actionlb.DeleteLoadBalancerOption, len(infoMap))
    for id, info := range infoMap {
        key := genAccountRegionKey(info)
        if reqMap[key] == nil {
            reqMap[key] = &actionlb.DeleteLoadBalancerOption{...}
        }
        req := reqMap[key]
        req.IDs = append(req.IDs, id)  // <-- 无上限累加
    }
    getNextID := counter.NewNumStringCounter(1, 10)
    for _, req := range reqMap {
        tasks = append(tasks, ts.CustomFlowTask{...})  // <-- 每组仅 1 个 task
    }
    return tasks
}
```

当某个 account+region 下 LB 数量 > 20 时，task 执行链路 `DeleteLoadBalancerAction.Run` → `hc-service.BatchDeleteTCloudLoadBalancer` → `req.Validate()` 校验 `len(r.IDs) > constant.BatchListenerMaxLimit` 失败，返回 `"batch delete limit is 20"` 错误。

### 调用链确认

```
cloud-server batchDeleteLoadBalancer
  └→ buildLBDeletionTasks (分组，无分批) ← 改造点
    └→ TaskServer CreateCustomFlow
      └→ DeleteLoadBalancerAction.Run (透传，不改动)
        └→ hc-service BatchDeleteTCloudLoadBalancer
          └→ req.Validate() (len(IDs) > 20 报错) ← 防线保留
            └→ adaptor client.DeleteLoadBalancer
```

## 目标 / 非目标

**目标：**
- 在 cloud-server `buildLBDeletionTasks` 中对每组 IDs 按 `BatchListenerMaxLimit`(20) 分批，生成多个 task
- 保证每个 task 的 `IDs` 数量 ≤ 20，从源头消除 Validate 报错
- 保持 hc-service Validate 防线不被破坏
- 保持 task-server 透传逻辑不变

**非目标：**
- 不修改 hc-service 的 `BatchDeleteLoadBalancerReq.Validate()` 或 `BatchDeleteTCloudLoadBalancer` 逻辑
- 不修改 task-server 的 `DeleteLoadBalancerAction` 结构或 `Run` 方法
- 不引入通用的 slice chunk 工具函数（项目中无此先例，for 循环+切片即可）
- 不处理非腾讯云 vendor 的分批（当前仅 TCloud 有 20 限制，且 `DeleteLoadBalancerOption` 仅支持 TCloud）

## 设计决策

### 决策 1：在 cloud-server `buildLBDeletionTasks` 层分批（方案 A）

**方案 A（选定）**：在现有"按 account+region 分组"基础上，对每组 IDs 再按 `BatchListenerMaxLimit`(20) 拆分为多个 task。

**方案 B（放弃）**：在 hc-service `BatchDeleteTCloudLoadBalancer` 层循环分批调用 adaptor。
- 缺点 1：需放开 Validate 的 20 限制，该 Validate 也保护直接调 hc-service API 的其他入口，放开有风险
- 缺点 2：单 task 内循环删除中途失败时已删部分无法回滚，错误上报不清晰
- 缺点 3：task 重试粒度变粗（整组重试 vs 每批独立重试）

**选择方案 A 的理由：**
1. 符合现有"按账号+地域分列表"设计模式，自然延伸为"按账号+地域+批次分"
2. 保留 hc-service 的 Validate 作为最后防线不被破坏
3. 每个 task 独立重试（已有 `Retry: tableasync.NewRetryWithPolicy(3, 1000, 5000)`），失败隔离粒度更细
4. 多个 task 可由 flow 并行调度，不降低吞吐

### 决策 2：分批实现方式——for 循环 + 切片

项目中 `pkg/tools` 下无现成的 Split/Chunk/Batch 工具函数。为保持简洁，直接在 `buildLBDeletionTasks` 函数内使用 for 循环 + 切片实现分批，不新建工具函数。

```go
for i := 0; i < len(req.IDs); i += constant.BatchListenerMaxLimit {
    end := i + constant.BatchListenerMaxLimit
    if end > len(req.IDs) {
        end = len(req.IDs)
    }
    batchIDs := req.IDs[i:end]
    // 为每批创建独立 task
}
```

### 决策 3：ActionID 唯一性

现有代码使用 `counter.NewNumStringCounter(1, 10)` 生成递增 ActionID。分批后每组可能产生多个 task，但 `getNextID()` 在遍历所有批次时被调用，天然保证全局递增唯一。无需额外处理。

### 决策 4：每批 task 的 Params 构造

每批需要独立的 `DeleteLoadBalancerOption` 副本，共享 `AccountID`、`Region`、`Vendor`，但 `IDs` 为该批的子集。关键：不能直接复用原 `req` 指针，必须为每批构造新的 option 对象，避免 IDs 引用共享导致后续批次的 IDs 被覆盖。

```go
batchOpt := actionlb.DeleteLoadBalancerOption{
    Vendor: req.Vendor,
    BatchDeleteLoadBalancerReq: hcproto.BatchDeleteLoadBalancerReq{
        AccountID: req.AccountID,
        Region:    req.Region,
        IDs:       batchIDs,
    },
}
```

## 改造后的代码结构

```
cmd/cloud-server/service/load-balancer/delete.go
├── import 新增: "hcm/pkg/criteria/constant"
├── buildLBDeletionTasks(infoMap)  ← 修改
│   ├── 第一阶段: 按 account+region 分组到 reqMap (不变)
│   └── 第二阶段: 遍历 reqMap，对每组 IDs 按 20 分批，每批生成一个 task (修改)
├── genAccountRegionKey(info)  ← 不变
```

## 风险 / 权衡

| 风险 | 缓解措施 |
|---|---|
| 分批后 task 数量增多，flow 调度压力增大 | 实际场景中单次删除 > 20 个 LB 的 case 不常见；即使 100 个 LB 也仅 5 个 task，flow 可并行处理 |
| 批次间部分成功部分失败 | 每 task 独立重试（3 次），失败 task 不影响成功 task；flow 层面已有失败处理机制 |
| 分批切片边界错误 | 边界条件已在设计中明确列举，代码实现时严格遵循 |

## 边界条件验证

| 场景 | 输入 | 预期输出 |
|---|---|---|
| 单组恰好 20 个 | IDs=[1..20] | 1 个 task，IDs=[1..20] |
| 单组 21 个 | IDs=[1..21] | 2 个 task：[1..20] + [21] |
| 单组 40 个 | IDs=[1..40] | 2 个 task：[1..20] + [21..40] |
| 单组 0 个 | 不会出现 | infoMap 非空才会进入函数 |
| 多组各自分批 | groupA=25, groupB=10 | groupA→2 task, groupB→1 task, 共 3 task |
| ActionID 唯一性 | 3 个 task | ActionID 分别为 "1"、"2"、"3" |

## 待解决问题

- 无（方案已明确，可直接实施）
