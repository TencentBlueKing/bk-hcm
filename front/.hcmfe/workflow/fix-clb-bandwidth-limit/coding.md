# CLB 规格带宽联动修复

## 来源与范围

- 来源：MR !3454，源提交 `a1b17db0cc0fbee76fc480f3df11f34b67b80573`，经工蜂 MCP 读取并与本仓库源码对照。
- 参考地址：`https://<GIT_HOST>/bcc/hcm/-/merge_requests/3454`。
- 用户确认 lite，仅本地修复与文档，跳过 TAPD 澄清与新建单据。
- 本地基线：`27cedc81c2ab80c0a9ee0aba96927b4c6aa68735`。
- lite 不含 PRD / Design / API 阶段；无接口或页面结构变更。

## 根因与实现

**文件**：`src/views/service/service-apply/clb/hooks/useRenderForm.tsx` `src/views/business/load-balancer/components/lb-spec-type-dialog/index.tsx`

旧实现用单个 selectedSlaTypeBandWidthLimit 保存上次弹框确认的上限，切回共享型、重置规格或编辑其它配置条目时可能沿用旧值。

按 MR 相同方式修复：

1. 用按 sla_type 索引的响应式映射记录确认的上限。
2. 当前规格非空且非 shared 时，依次取该规格缓存、CLB_SPEC_TYPE_COLUMNS_MAP 对应规格上限、共享型兜底；shared / 空规格固定取共享型 10240。
3. 监听计算出的 max，下调时仅将超限的 internet_max_bandwidth_out 收敛至 max；范围内与上调时保留原值。
4. 沿用现有 Slider 的 minValue / maxValue / customContent 接线。

## 仓库差异

- 本仓库仅有云端规格弹框 index.tsx，confirm 的 bandwidthLimit 来源于 CLB_SPEC_TYPE_COLUMNS_MAP。
- 源变更中的其它部署变体不适用于当前仓库，仅移植当前入口的修复。
- 本地规格表最大值为 61440；验证使用本地实际规格值，不照抄来源文档中的 81920。
- 不复制来源工作流的审批记录、done 状态或测试结论；本地保持真实状态。

## 验证

执行记录见 test.md。编码完成后登记产物、核对 freshness 与 drift；阶段推进等待用户确认。

2026-10-20 复核：当前 watch 回调参数增加了 `max: number` 类型标注，带宽取值和收敛逻辑与已验证版本一致。保留该标注，并按用户“继续”确认对齐基线、推进 test。

## 2026-10-20 自测问题修复：未知规格

用户提供的响应包含 clb.c5.medium / clb.c5.small，而参数及名称映射缺失；直接解构 undefined 会中断整批列表赋值。

- 列表转换在映射缺失时使用空对象，缺失数值显示 --；未知带宽排在已知带宽之后，空输入列表可安全处理。
- 型号列从 row 读取，名称优先映射、其次显示接口 SpecType；已选规格预览采用相同兜底。
- 确认时安全读取可选 bandwidthLimit，表单接收可选上限；未知规格仍按现有逻辑兜底 10240，不将该值声明为云端实际上限。
- 不补写未经确认的 c5 性能数值，不更改接口或规格过滤规则。
- 保持 test 阶段，新增逻辑检查后等待用户继续手动验证。
