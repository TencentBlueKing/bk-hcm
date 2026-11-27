# 闸门评审报告

**变更 ID**: `tcloud-lb-delete-batch-split`
**评审日期**: 2025-07-11
**评审人**: Architect

---

## 1. 需求完整性与歧义检查

**结论**: ✅ 通过

需求描述清晰，调用链分析完整，边界条件列举充分（20/21/40/0/多组混合/ActionID 唯一性）。问题根因定位精确到代码行号（`delete.go:213-242`），方案方向明确（方案 A vs 方案 B 的权衡分析合理）。

无歧义点。

## 2. 技术方案合理性

**结论**: ✅ 通过

- 方案 A（cloud-server 层分批）正确选择：符合现有"按账号+地域分组"设计模式的自然延伸，保留 hc-service Validate 防线，task 级独立重试粒度更细
- 分批常量 `constant.BatchListenerMaxLimit` 复用已有定义（`pkg/criteria/constant/clb.go:29`），未引入魔法数字
- for 循环 + 切片的实现方式与项目风格一致（项目中无通用 chunk 工具，不强行引入）
- 每 batch 独立 `DeleteLoadBalancerOption` 副本的设计正确避免了 IDs 引用共享问题

## 3. 落地可行性

**结论**: ✅ 通过

- 改造范围仅限 `cmd/cloud-server/service/load-balancer/delete.go` 单文件
- 改动量约 15 行（新增 import 1 行 + 修改第二阶段循环体 ~12 行）
- 不涉及接口签名变更、数据库 schema 变更、配置变更
- 不影响其他 vendor（当前 `DeleteLoadBalancerOption` 仅支持 TCloud）

## 4. 安全性

**结论**: ✅ 通过

- hc-service 的 `BatchDeleteLoadBalancerReq.Validate()` 保持不变，仍作为最后防线拦截超限请求
- 不引入新的攻击面
- 鉴权逻辑（`validHandler`）在 `buildLBDeletionTasks` 之前执行，分批不影响鉴权

## 5. 稳定性与可靠性

**结论**: ✅ 通过

- 每 batch task 独立重试（`Retry: tableasync.NewRetryWithPolicy(3, 1000, 5000)`），失败隔离粒度从"整组"细化到"每 20 个"
- 多个 task 由 flow 并行调度，吞吐不降反升
- 分批后 task 数量增多对 flow 调度压力可忽略（100 个 LB 仅 5 个 task）

## 6. 可测试性

**结论**: ✅ 通过

- 边界条件明确，可编写表驱动测试覆盖（20/21/40/多组混合）
- `buildLBDeletionTasks` 是包级函数，可直接构造 `infoMap` 入参进行单测
- 验证点清晰：task 数量、每 task IDs 长度、ActionID 唯一性

## 7. 可上线与可运维性

**结论**: ✅ 通过

- 无配置变更，无数据库迁移，灰度发布无特殊要求
- 回滚策略：直接回退代码版本即可，无数据兼容性问题
- 监控：现有 flow/task 监控自动覆盖新增 task

## 8. 实现前疑问点识别

**结论**: ✅ 无阻塞疑问

- `BatchListenerMaxLimit` 常量名为 "Listener" 但实际用于 LB 删除限制，语义稍有偏差，但这是已有命名，不在本次改造范围内
- `DeleteLoadBalancerOption` 当前仅支持 TCloud vendor，若未来扩展其他 vendor 且其 API 限制不同，需考虑按 vendor 区分 batch size。但当前需求仅针对 TCloud，不做过度设计

---

## 评审结论

### 【可直接进入开发】

**阻塞项**: 无

**建议项**:
1. 实现时建议为 `buildLBDeletionTasks` 补充单元测试（表驱动），覆盖边界条件
2. 可在注释中标注分批原因和常量来源，便于后续维护者理解

**评估摘要**:

| 维度 | 结论 |
|---|---|
| 需求完整性 | ✅ |
| 技术方案合理性 | ✅ |
| 落地可行性 | ✅ |
| 安全性 | ✅ |
| 稳定性与可靠性 | ✅ |
| 可测试性 | ✅ |
| 可上线与可运维性 | ✅ |
| 疑问点 | ✅ 无阻塞 |
