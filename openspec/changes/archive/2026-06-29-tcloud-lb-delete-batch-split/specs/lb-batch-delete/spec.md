## MODIFIED Requirements

### Requirement: 负载均衡批量删除 task 按 API 限制分批

`buildLBDeletionTasks` 函数在按 account+vendor+region 分组后，SHALL 对每组的 LB IDs 列表按 `constant.BatchListenerMaxLimit`（当前值为 20）进行分批拆分，每个批次生成一个独立的 `ts.CustomFlowTask`。单个 task 的 `IDs` 数量 SHALL 不超过 20。

#### Scenario: 单组 LB 数量 ≤ 20

- **WHEN** 某个 account+region 分组下有 ≤ 20 个 LB
- **THEN** 该组生成 1 个 task，task 的 IDs 包含该组全部 LB ID，行为与改造前一致

#### Scenario: 单组 LB 数量 > 20（如 25）

- **WHEN** 某个 account+region 分组下有 25 个 LB
- **THEN** 该组生成 2 个 task：第一个 task 的 IDs 为前 20 个 LB ID，第二个 task 的 IDs 为剩余 5 个 LB ID

#### Scenario: 单组 LB 数量恰好为 20 的整数倍（如 40）

- **WHEN** 某个 account+region 分组下有 40 个 LB
- **THEN** 该组生成 2 个 task，每个 task 的 IDs 各含 20 个 LB ID

#### Scenario: 多个 account+region 分组各自独立分批

- **WHEN** groupA（account1+tcloud+ap-guangzhou）有 25 个 LB，groupB（account2+tcloud+ap-shanghai）有 10 个 LB
- **THEN** groupA 生成 2 个 task，groupB 生成 1 个 task，共 3 个 task，各组分批互不影响

---

### Requirement: 每个 task 的 ActionID 保持唯一

分批后生成的多个 task SHALL 各自拥有唯一的 `ActionID`，通过 `counter.NewNumStringCounter(1, 10)` 递增生成，确保 flow 调度不冲突。

#### Scenario: 多批次 ActionID 递增

- **WHEN** 分批后共生成 3 个 task
- **THEN** 三个 task 的 ActionID 分别为 "1"、"2"、"3"（或等价的递增序列）

---

### Requirement: hc-service Validate 防线保持不变

`pkg/api/hc-service/load-balancer/tcloud.go` 的 `BatchDeleteLoadBalancerReq.Validate()` SHALL 保持 `len(r.IDs) > constant.BatchListenerMaxLimit` 校验不变，作为最后防线。cloud-server 层分批后，正常流程下不应触发此校验报错。

#### Scenario: 正常流程不触发 Validate 报错

- **WHEN** cloud-server 按方案 A 分批后，每个 task 的 IDs ≤ 20
- **THEN** task 执行到达 hc-service 时，`req.Validate()` 通过，不返回 `"batch delete limit is 20"` 错误

---

### Requirement: 每 batch task 拥有独立的 Params 副本

分批生成的每个 task SHALL 拥有独立的 `DeleteLoadBalancerOption` 对象，各 batch 之间不共享 IDs 切片引用，避免后续批次被覆盖。

#### Scenario: 批次间 IDs 隔离

- **WHEN** 某组 25 个 LB 分为 2 批（20 + 5）
- **THEN** 第一个 task 的 Params.IDs 长度为 20，第二个 task 的 Params.IDs 长度为 5，两者指向不同的底层数组
