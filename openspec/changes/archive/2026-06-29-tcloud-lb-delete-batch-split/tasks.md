## 1. 修改 `buildLBDeletionTasks` 函数实现分批逻辑

- [x] 1.1 在 `cmd/cloud-server/service/load-balancer/delete.go` 的 import 块中新增 `"hcm/pkg/criteria/constant"`
- [x] 1.2 修改 `buildLBDeletionTasks` 函数的第二阶段（`for _, req := range reqMap` 循环体）：将原来每组直接生成 1 个 task 的逻辑，改为对 `req.IDs` 按 `constant.BatchListenerMaxLimit`(20) 分批，每批构造独立的 `actionlb.DeleteLoadBalancerOption` 副本（共享 AccountID/Region/Vendor，IDs 为该批子集），生成一个 `ts.CustomFlowTask`
- [x] 1.3 确保分批切片边界正确：`for i := 0; i < len(req.IDs); i += constant.BatchListenerMaxLimit`，`end := min(i + constant.BatchListenerMaxLimit, len(req.IDs))`，`batchIDs := req.IDs[i:end]`
- [x] 1.4 确保每批 task 的 `ActionID` 通过 `getNextID()` 递增生成，保持全局唯一
- [x] 1.5 确保每批 task 的 `Retry` 策略保持 `tableasync.NewRetryWithPolicy(3, 1000, 5000)` 不变

## 2. 验证

- [x] 2.1 编译通过：`go build ./cmd/cloud-server/...`
- [x] 2.2 现有单测通过：`go test ./cmd/cloud-server/service/load-balancer/...`（如有）
- [x] 2.3 手动/code review 验证边界条件：单组 20 个（1 task）、21 个（2 task）、40 个（2 task）、多组混合各自分批
