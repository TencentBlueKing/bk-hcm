# Coding — clb-random-zone-exclusive-cluster

## 目标

业务下申请公网 CLB 时，单可用区选择“随机可用区”仍查询当前地域的独占集群标签；查询成功且有可用数据后，沿用现有独占型配置和提交逻辑。需求依据：TAPD #1069995598138577666 的已澄清描述。

## 现状与原因

- `src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts` 在查询前要求 `formModel.zones` 非空，独占型可用性判断也要求它非空；随机可用区使用空值，因此不会发起标签查询，独占型选项也不会显示。
- 查询请求当前必传 `zones`，空值被包装成数组；`src/api/load_balancers/apply-clb/types.ts` 也将该字段声明为必填。
- 已选具体可用区的查询、查询失败与无数据时的表现、独占集群配置和购买提交已有现成流程。

## 实施方案

1. 在 `src/api/load_balancers/apply-clb/types.ts` 将独占集群标签查询的 `zones` 标为可选，并说明随机可用区时不传。
2. 在 `src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts` 移除查询与独占型可用性判断中的非空 `zones` 门槛；有具体可用区时保持原有数组请求值，随机可用区时从请求体省略 `zones`。保留现有请求快照、失败状态和重选清理逻辑。
3. 核对表单中的随机可用区取值及购买提交路径，确保上述改动只影响标签查询与选项可用性，不改变现有购买参数语义。若核对发现相关校验还阻挡随机可用区，再补最小必要调整并回填本文档。

## 验证

- 随机可用区：满足账号、地域、运营商条件时发起标签查询，请求体无 `zones`；有数据可选择独占型并继续配置。
- 指定可用区：请求体仍包含原有 `zones` 数组，独占型配置保持可用。
- 无数据或请求失败：独占型入口按现有逻辑不可用；切换账号、地域、运营商或可用区后不会使用旧请求结果。
- 对变更文件执行项目 lint，并检查相关类型与测试结果。

## 文件

- `src/api/load_balancers/apply-clb/types.ts`
- `src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts`

## 实施与校验记录

- 已将标签查询 `zones` 定义为可选；随机可用区时请求体省略该字段，具体可用区仍传数组。独占型可用性与查询前置条件不再要求可用区非空。
- 已核对 `src/components/zone-selector/index.vue` 的单选空值与 `src/views/service/service-apply/clb/children/bottom-bar.tsx` 的购买参数构造：本次仅调整标签查询；购买提交仍沿用原有空可用区表示。
- 本地 Prettier、ESLint 通过；`git diff --check` 通过。全量 `tsc --noEmit` 在未改动的 `@blueking/chat-x` 声明文件出现 TS1139 等语法错误，未能完成项目类型检查。
- 已将稳定行为补充到 `docs/modules/service.md` 并刷新该模块文档基线。

## 追加修补：购买后定位负载均衡单据 tab

业务单据页 `src/views/ticket/entry-biz.vue` 的负载均衡 tab 名为 `load-balancer`，当前购买成功跳转和 CLB 单据详情返回链接传的是 `load_balancer`，导致 URL 的 `type` 与 tab 名不匹配。

- 将 `src/views/service/service-apply/clb/apply-clb-internal.plugin.ts` 的业务购买成功跳转改为 `type=load-balancer`，保留现有 `bizs` 和路由名。
- 同步修正 `src/views/ticket/children/apply-detail/clb.vue` 的返回单据列表链接，避免从详情返回时再次丢失 tab 定位。
- 验证业务跳转组合为 `/business/ticket?bizs=<BIZ_ID>&type=load-balancer`，并核对已有路由名与 tab 消费逻辑；运行变更文件 lint 和 diff 检查。

**追加文件**：`src/views/service/service-apply/clb/apply-clb-internal.plugin.ts`、`src/views/ticket/children/apply-detail/clb.vue`。

**追加修补结果**：两个入口均已改用 `type=load-balancer`，与 `src/views/ticket/entry-biz.vue` 的 tab 名一致；`bizs` 和现有路由名保持原值。两个变更文件的本地 Prettier、ESLint 与 `git diff --check` 通过。模块稳定行为已补充至 `docs/modules/service.md` 和 `docs/modules/ticket.md` 并刷新基线。

## 追加优化：独占集群校验体验

用户提供当前页面与参考截图，要求沿用 BK UI Form/Checkbox/Select 能力：

1. 勾选四层集群时预选第一条四层标签，并保持集群、IP 为现有“随机分配”；勾选七层标签时预选第一条七层标签。无可选项时保留空值并显示对应校验提示。
2. 将独占集群校验规则放到“负载均衡规格类型”表单项，使错误位于该选择器下。未启用任一集群的规则与原逻辑一致；四层和七层未完整配置分别提示。
3. 勾选状态变化与独占集群配置变化后刷新该表单项校验，取消勾选时及时消除不再适用的提示。
4. 账号、地域、可用区、运营商等导致独占集群标签重新查询时，以及切离独占型时，同时清空配置值和两个复选框状态；编辑回填继续保留已存配置。

**预计文件**：`src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts`、`src/views/service/service-apply/clb/hooks/useRenderForm.tsx`。校验使用 BK UI FormItem rules、Form.validate 与 Checkbox/Select change 事件；不新增自绘错误样式。

**实现记录**：已将原有组合校验拆为“至少启用一项”、四层配置、七层标签、请求失败与 IP 有效性规则，分别给四层/七层错误文案，仍要求所有已启用项完整。规则挂在 `slaType` 的 FormItem，数组形式的表单项渲染已透传项级 rules；勾选时预选首个标签，四层集群和 VIP 沿用随机分配；配置字段与勾选状态变化后重新校验。标签查询条件变化和切离独占型时重置配置与两个勾选状态，编辑回填由既有 `noResetParams` 保护。BK UI Form/Checkbox/Select 事件与规则能力已核对，未增加手写错误样式。

**校验记录**：已运行本地 Prettier 写入及 ESLint 修复，最终 ESLint 检查与 `git diff --check` 通过。项目的 Prettier CLI 与 ESLint 内置 Prettier 规则对 TSX 换行有冲突，最终以项目 ESLint 规则修复后的结果为准。全量 TypeScript 检查仍受既有 `@blueking/chat-x` 声明文件语法错误阻挡。稳定行为已补充至 `docs/modules/service.md` 并刷新模块基线。
