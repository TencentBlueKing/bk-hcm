# Coding — feat-clb-exclusive-cluster-purchase

## 执行顺序

1. 购买支持独占型规格：先建立类型、接口和表单状态闭环，完成四层/七层与带宽包联动。
2. 单据展示独占集群参数：消费后端新增的申请单 `content` 结构，独立补充详情字段。
3. CLB 详情独占集群字段：消费资源详情 `extension`，复用现有 VIP 显示逻辑。

> 排序依据：购买单提供申请字段和核心交互；两类详情只读显示，可在购买模型稳定后分别落地。

## 共享改动 / 提交策略

- 跨单公共类型：独占集群条目及 `TGW` / `STGW` 判别类型放在现有负载均衡领域类型附近；申请单和资源详情复用同一字段形态，不新增全局抽象层。规格文本统一调用最新基线已有的 `getLoadBalancerInstanceSpecName`，其性能容量档位统一读取 `CLB_SPECS`。
- 组件策略：复用现有 `BandwidthPackageSelector`、`ConfigureList`、`display-value`、`DetailInfo` 和 `getInstVip`；不新增路由、菜单、权限和基础 UI 组件。
- 异步策略：标签、空闲 VIP、带宽包请求均以请求序号和发起时关键表单快照作双重校验，旧响应不得覆盖新条件。
- 提交策略：按三张 TAPD 单分别提交；跨单公共类型随第一张购买单提交，后两张只提交各自详情文件及必要类型补充。

---

## 单据 1: 购买支持独占型规格

**TAPD**: [#1069995598138114049](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114049)

**文件**: `src/api/load_balancers/apply-clb/types.ts` `src/api/load_balancers/apply-clb/index.ts` `src/views/service/service-apply/clb/index.tsx` `src/views/service/service-apply/clb/hooks/useRenderForm.tsx` `src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts` `src/views/service/service-apply/clb/children/bottom-bar.tsx` `src/views/service/service-apply/clb/children/calc-price.tsx` `src/views/service/service-apply/components/common/BandwidthPackageSelector/index.tsx` `src/views/service/service-apply/components/common/configuration-list/index.vue` `src/views/service/service-apply/clb/index.module.scss`

**改动点**:

- 扩展 `ApplyClbModel`，增加提交字段及仅前端使用的四层/七层启用、标签、具体集群、IP选择状态；类型化标签聚合、集群和空闲 VIP 响应。
- 在现有 API 模块封装业务视角标签聚合和空闲 VIP 请求，不把 URL 与请求拼装散落在渲染逻辑中。
- 新建局部 `useExclusiveCluster` hook，集中负责：
  - 标签查询及 TGW/STGW 分组。
  - 复选框启用状态、保留值和上游条件变化后的失效清理。
  - 四层随机/具体集群与空闲 VIP 级联。
  - 四层、七层有效 egress 集合及交集计算。
  - 请求序号与表单快照竞态保护。
- 规格类型增加独占型，仅业务视角、公网且标签数据可用时允许选择；独占型区域按稿面组合 Checkbox 与 Select，未勾选层禁用但保留值。
- 保存表单前校验至少启用一层、启用层字段完整、具体四层集群存在空闲 VIP，以及共享带宽包计费已选择有效候选。
- 提交参数严格由启用状态生成：未启用层不输出；独占型清空 `sla_type`，四层随机提交候选 ID 数组，随机 IP 提交空字符串，七层提交标签。
- 询价仍沿用现有共享型/性能容量型前置条件；新增的独占集群选择器缓存和启用状态不透传给询价接口，避免污染非独占请求。
- 改造 `BandwidthPackageSelector`：
  - 新增可选 egress 集合输入。
  - 使用 `@blueking/roll-request` 按 `offset` / `limit`、`total_count` / `packages` 全量读取。
  - 全量完成后再按 egress、状态和既有地域约束过滤。
  - 关键条件变化时清空选中值；旧请求结果不得回写。
  - 非独占调用方保持原行为，避免影响 CVM 等复用页面。
- 配置清单规格显示按独占型、性能容量型、共享型优先级计算；克隆/编辑配置时完整保留前端状态。

### 实施检查

- Coding 前使用项目 BKUI 文档工具确认 Checkbox、Select 的 `modelValue`、禁用、loading 与 change 事件用法；保持目标目录现有导入方式。
- 四层具体集群 `count=0` 时不允许保存该集群；接口失败保持表单并展示错误。
- 不把未勾选层的保留值带入 egress 或提交参数。

## 单据 2: 单据展示独占集群参数

**TAPD**: [#1069995598138114062](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114062)

**文件**: `src/views/ticket/children/apply-detail/clb.vue`

**改动点**:

- 扩展申请单 `content` 的局部类型，安全解析 `exclusive`、`sla_type`、`vip` 和 `clusters`。
- 删除页面私有的 `DISPLAY_CLB_SPECS_MAP`；解析后生成统一实例规格展示值，调用 `getLoadBalancerInstanceSpecName({ exclusive, sla_type })`，不再维护第二套规格文案。
- 从 `clusters` 中分别定位 TGW/STGW；独占型在规格字段后插入四层标签、四层名称、四层 IP、七层标签。
- 四层 IP只读 `content.vip`；随机或空字符串展示项目统一空值，不查询资源详情。
- STGW 元素缺失时允许使用顶层 `cluster_tag` 作为七层标签兼容值。
- 共享型和性能容量型不渲染独占字段；JSON解析失败或字段缺失时不阻断其它详情内容。
- 复核业务/资源视角现有请求层返回形态；展示组件只消费已经归一后的 `applicationDetail.content`。
- 已落地：直接解析 `content.clusters` 的 TGW/STGW 项和 `content.vip`，规格文案统一走全局方法；缺失项与随机值保留空值展示，不反查资源详情。

## 单据 3: CLB 详情独占集群字段

**TAPD**: [#1069995598138114066](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114066)

**文件**: `src/store/load-balancer/clb.ts` `src/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail/index.tsx`

**改动点**:

- 扩展 `ILoadBalancerDetails.extension`：`exclusive` 与独占集群 `clusters` 类型。
- 在详情组件中分别计算 TGW/STGW 信息；“规格类型”统一调用 `getLoadBalancerInstanceSpecName(props.detail.extension)`，移除详情组件对 `CLB_SPECS` 的直接映射与导入。
- 独占型在配置信息末尾展示四层标签、名称、IP和七层标签；非独占型只展示实例规格。
- 四层集群 IP 直接复用 `getInstVip(props.detail)`，可用 computed 缓存，同现有“负载均衡 VIP”保持完全一致的地址优先级与空值。
- 集群名称或标签缺失时使用现有 `--` 空值规范，不改变监听器、安全组和跨域配置。
- 已落地：新增字段位于配置信息末尾，四层 IP 与现有负载均衡 VIP 共用 `getInstVip`；旧数据缺少 `exclusive` 时按共享型兼容显示。

## 验证与产物保鲜

1. 每完成一组代码改动，更新对应单据的 `**文件**` 与 `**改动点**`，再读取 workflow status。
2. 若 coding freshness 为 stale/untracked，重新登记/对齐 `coding.md`；不得通过修改 PRD/Design 迁就代码。
3. 运行 bkdevbuddy lint；再执行前端 build，区分本轮错误与仓库基线问题。
4. Coding 完成后进入 Test，按 `wf-test-checklist` 生成 P0/P1/P2 手测清单；默认由用户自测。

## 本轮实现与验证记录

- 购买单已落地独占型规格、分层复选框、随机/指定四层集群及 IP、七层标签、有效出口交集与带宽包全量分页过滤。编辑配置时保留原勾选和值；重启用指定四层时重新确认空闲 IP。表单条件切换后的旧标签、VIP、带宽包响应不回写。
- 购买提交仅输出启用层字段；配置清单、申请单详情和 CLB 详情统一使用全局实例规格文案方法。
- 目标文件 ESLint 通过；生产构建通过（仅有项目既有 Sass/包体积警告）。全仓 TypeScript 检查因依赖声明及其它模块基线错误未通过，但本轮改动文件未出现新的 TypeScript 诊断。
- 工作区已有的 package.json、E2E、配置文件等改动为本迭代前的用户文件；不归入三张单据的文件清单，也不在本轮提交中处理。

## D-01 修复记录（Test 阶段回流）

**TAPD**: [#1069995598138114049](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114049)

**文件**: `src/views/service/service-apply/clb/hooks/useRenderForm.tsx`

**改动点**:

- 移除「独占集群」`FormItem` 的 `required: true`。该属性会按 `property` 对模型字段判空，而 `exclusive_config` 并不是 `ApplyClbModel` 的字段，必填规则恒判为空并输出「独占集群不能为空」，导致独占型永远无法保存，自定义校验文案也被覆盖。
- 把独占集群的自定义校验从 form 级 `rules.exclusive_config` 迁到 `FormItem` 的**项级 `rules`**（BKUI `FormItem` 支持 `Array<IFormItemRule>`，已核对组件文档），校验语义集中在「独占集群」一项，不再依赖不存在的模型字段。
- `ClbFormItemOption` 增加可选 `rules`，渲染单表单项时透传给 `FormItem`。数组行与其它表单项行为不变。
- 预期行为：两层未启用或启用层配置不完整时仍拦截提交并提示「请至少启用并完整配置一个独占集群」；配置完整时可正常保存与提交。
- 视觉副作用：该字段不再显示必填星号（BKUI `FormItem` 无独立控制星号的属性，`required` 同时承担判空与星号），校验行为由项级 validator 承担。

**验证**: `e2e.spec.ts` 共 22 条用例全部通过，含此前被 D-01 阻塞的 `P0-02`、`P0-03`、`P0-04`、`P0-05`、`P1-08`、`P1-09`。

## isp 入参收窄为三网直连（2026-09-20）

**TAPD**: [#1069995598138114049](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114049)

**背景**: 独占集群标签接口 `POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/tags/list` 的 `isp` 入参只支持三网直连，BGP（含自研云按 `TypeSet` 拆分出的 BGP 系取值）等其它运营商不支持独占集群。

**改动点**:

- `src/api/load_balancers/apply-clb/types.ts`: 新增 `ExclusiveClusterIsp = 'CMCC' | 'CUCC' | 'CTCC'`，`ExclusiveClusterTagsReqData.isp` 由 `string` 收窄为 `ExclusiveClusterIsp`。
- `src/views/service/service-apply/clb/hooks/useExclusiveCluster.ts`: 新增 `EXCLUSIVE_CLUSTER_ISP_TYPES` 与类型谓词 `isExclusiveClusterIsp`；`isExclusiveAvailable` 的运营商条件改为 `isExclusiveClusterIsp(formModel.vip_isp)`；`loadExclusiveClusterTags` 前置校验改用同一谓词——非三网运营商**不发请求**，并按既有逻辑清空标签与 `formModel.exclusive_cluster_tags`。
- `src/views/service/service-apply/clb/hooks/useRenderForm.tsx`: `exclusiveConfigRules` 增加运营商兜底校验（`slaType === '2'` 且运营商非三网时判定不通过），防止克隆 / 编辑配置等入口带入独占态。

**行为**:

- 「独占型」选项仅在运营商为 `CMCC` / `CUCC` / `CTCC` 时出现；切换到 BGP 等其它运营商后 `isExclusiveAvailable` 变为 false，既有 `watch` 自动回退为共享型（`slaType='0'`、`sla_type='shared'`、`exclusive=0`）并清空独占选择。
- 自研云仅在运营商选择移动 / 联通 / 电信（原样 `Isp` 值）时支持独占型；选择 BGP 系（`ziyan` / `ziyan_normal_bgp` / `ziyan_mianliu` 等 `TypeSet[0].Type` 取值）时不支持。免流 / 直通开关与 `tgw_group_name` 推导链不受影响。

**校验**: 改动文件 ESLint 0 error、IDE 诊断 0、生产构建通过；`api.md` §2 的 `isp` 参数表已同步为枚举口径。
