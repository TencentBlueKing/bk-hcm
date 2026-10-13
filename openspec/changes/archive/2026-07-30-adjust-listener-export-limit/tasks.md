## 1. 常量定义

- [x] 1.1 在 `pkg/criteria/constant/clb.go` 的导出限制常量组中新增 `ExportSkipLimitLbCount = 5`，注释说明「勾选负载均衡数量不超过该值时，导出不做数量限制」，并注明参数层与业务层共用该阈值、置为 `0` 可关闭该特性

## 2. 参数层校验分流

- [x] 2.1 在 `pkg/api/cloud-server/load-balancer/load_balancer.go` 的 `ExportListenerReq.Validate` 中，按 `len(r.GetAllLbIDs()) <= constant.ExportSkipLimitLbCount` 分流：命中时仅执行 `listeners` 非空与逐元素 `lb_id` 非空校验
- [x] 2.2 未命中分流时保持原有三项数量校验（`listeners` 数量 ≤ 5000、去重后 `lbl_ids` 总数 ≤ 5000、单元素 `lbl_ids` ≤ 100）及原错误文案不变
- [x] 2.3 调整 `ExportListener.Validate` 的职责边界：数量校验的调用与否由 `ExportListenerReq.Validate` 决定，不在元素级方法签名上增加参数

## 3. 预检逻辑分流

- [x] 3.1 在 `cmd/cloud-server/logics/load-balancer/export_listener_excel.go` 中为 `listenerExporter` 新增判定方法 `skipCountLimit() bool`，返回 `len(l.params.GetAllLbIDs()) <= constant.ExportSkipLimitLbCount`，该方法不依赖 client 以便单测
  - 实现调整：判定方法改为定义在 `ExportListenerReq` 上（`SkipCountLimit()`），参数层与预检层复用同一实现，避免两处重复且保证口径一致
- [x] 3.2 在 `listenerExporter.PreCheck` 中，于 `checkClbListenerRel` 之后、`checkListenerCount` 之前插入 `skipCountLimit()` 判断，命中则 `return nil`
- [x] 3.3 分流命中时输出 Info 级别日志，包含勾选负载均衡数量、监听器 ID 数量与 `rid`
- [x] 3.4 在 count 类校验函数（`getListenerCountRule`、`checkRuleCount`、`getRsCountRule`）附近补充注释，说明其未分批的前提是「分流命中时整体跳过，未命中时参数层 5000 上限生效」，提醒后续调整阈值时同步评估

## 4. 查询分批改造

- [x] 4.1 改造 `getTCloudListenersByProtocol` 的 `lblIDs` 分支：以 `core.DefaultMaxPageLimit` 为粒度对 `lblIDs` 分批，每批内保留原有分页循环，结果并入同一个以监听器 ID 为键的 map
- [x] 4.2 改造 `getTCloudRulesByRuleType` 的 `lblIDs` 分支：同样按 500 分批，结果并入以规则 ID 为键的 map
- [x] 4.3 改造 `getTgLblRelClassifyProtocol`：将 `getTgLblRelRule` 的 `lb_id IN (...) OR lbl_id IN (...)` 单查询拆为 `lb_id` 维度与 `lbl_id` 维度两段独立分批查询，各自保留分页循环
- [x] 4.4 为 4.3 的合并结果引入按关联记录 ID 的去重，确保同一关联同时命中两个维度时只计入一次，再按 `listener_rule_type` 分流到四层/七层切片
- [x] 4.5 核对 `getLbs`、`checkClbListenerRel`、`getRsClassifyProtocol` 现有分批逻辑无需调整，并确认改造后无残留的未分批 `RuleIn` 使用无界切片

## 5. 单元测试

- [x] 5.1 在 `pkg/api/cloud-server/load-balancer/` 下补充 `ExportListenerReq.Validate` 的表驱动单测，覆盖：≤5 个 LB 且单元素 `lbl_ids` 为 3000 时通过、>5 个 LB 且单元素 `lbl_ids` 为 200 时报错、>5 个 LB 且 `lbl_ids` 总数超 5000 时报错、`listeners` 为空与 `lb_id` 为空在两种分支下均报错
- [x] 5.2 在 `pkg/api/cloud-server/load-balancer/` 下补充 `GetAllLbIDs` 的单测，覆盖同一 `lb_id` 拆分为多个 `listeners` 元素时的去重语义
- [x] 5.3 随 `SkipCountLimit` 的实现位置，在 `pkg/api/cloud-server/load-balancer/` 下补充其表驱动单测，覆盖边界值：勾选 1 个、5 个（命中跳过）、6 个（不跳过）、同一 `lb_id` 重复出现去重后不超过阈值
- [x] 5.4 确认新增单测遵循项目单测规范（表驱动 + `testify/assert`），且未引入新依赖

## 6. 接口文档

- [x] 6.1 更新 `docs/api-docs/web-server/docs/biz/load-balancer/export/export_listener_pre_check.md`：修正 `listeners` 与 `lbl_ids` 的长度描述（现为 100，实际为 5000），补充单个 `listeners` 元素的 `lbl_ids` 长度限制为 100 的说明
- [x] 6.2 在上述文档中补充分流规则说明：勾选负载均衡数量不超过 5 个时参数长度限制与资源数量限制均不生效，超过 5 个时按各项限制校验
- [x] 6.3 在 `export_listener.md` 中同步补充相同的分流规则说明，并补充单个元素 `lbl_ids` 限制 100 的说明
- [x] 6.4 确认两份文档的版本占位符按项目约定处理：两个接口均非新增（现有 `v1.8.5+`），本次仅为行为说明补充，无需新增 `v9.9.9` 占位符

## 7. 验证

- [x] 7.1 执行 `go build ./...` 与 `go vet ./cmd/cloud-server/... ./pkg/api/cloud-server/... ./pkg/criteria/constant/...`，确认无编译与静态检查问题
- [x] 7.2 执行 `go test ./cmd/cloud-server/logics/load-balancer/... ./pkg/api/cloud-server/load-balancer/...`，确认单测通过
- [ ] 7.3 手动验证：勾选 1 个负载均衡下超过 100 个监听器导出，`pre_check` 返回 `pass: true` 且能完成导出
- [ ] 7.4 手动验证：勾选 1 个监听器数量超过 5000 的负载均衡（仅传 `lb_id`），`pre_check` 返回 `pass: true` 且能完成导出
- [ ] 7.5 手动验证：构造监听器 ID 数量超过 10000 的请求，确认不出现 `at most have 10000 elements` 错误且导出文件完整
- [ ] 7.6 手动验证混合传参场景：同时传一个只带 `lb_id` 的元素与一个带 `lbl_ids` 的元素，比对导出文件中的 RS 行数与去重前后一致、无重复行
- [ ] 7.7 手动验证：勾选 6 个负载均衡且四层监听器合计超过 5000，`pre_check` 返回 `pass: false` 且 `reason` 与变更前一致
- [ ] 7.8 手动验证：勾选 1 个负载均衡但传入不属于该负载均衡的 `lbl_id`，`pre_check` 仍返回归属校验失败
- [ ] 7.9 手动验证：`tcloud` 与 `tcloud-ziyan` 两个厂商下分流与分批行为一致
- [ ] 7.10 回归验证：将 `ExportSkipLimitLbCount` 临时改为 `0`，确认行为与变更前完全一致
