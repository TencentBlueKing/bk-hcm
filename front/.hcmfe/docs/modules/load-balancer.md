# 负载均衡

> status: drafted · kind: module
> globs: `src/views/load-balancer/**`

CLB / 监听器 / 目标组 / 设备，以及资源接入下的独占集群列表。

## 职责

业务视角 CLB / 监听器 / 设备页。条件同步与单条同步走公共弹窗 `src/components/sync-account-resource`，**不在弹窗内写 CLB 文案**。

CLB 成功/失败反馈在模块内 `src/views/load-balancer/use-clb-sync-feedback.ts`，由列表/监听器等调用方注入 `successHandler` / `errorHandler`。

资源接入入口是 `src/views/load-balancer/entry-rsc.vue`：用 query `subtype` 在「负载均衡 / 独占集群」之间切换，radio 是 `children/resource-subtype-switch.vue`（模块内子组件放 `children/`），经 slot `toolbar-prefix` 插进当前子页 toolbar。`activeType` 是 `query.subtype` 的可写 computed（setter 里 `router.replace`），切换时先落 URL 再渲染子页；若改成 `ref` + 双向 watch，子页会在导航落地前挂载并按旧 query 先请求一次，首次切换即发双份列表请求。CLB 列表在 `src/views/resource/resource-manage/children/manage/load-balancer-manage.vue`，独占集群列表在 `src/views/load-balancer/exclusive-cluster/`，两页互不 import。默认「负载均衡」（不写 subtype）；`exclusive-cluster` 刷新可回到独占集群。购买只出现在 CLB 分支。

## 独占集群

列表在 `src/views/load-balancer/exclusive-cluster/`。资源接入筛选用公共 `src/components/resource-search-select`，字段来自模块内 `search-condition.ts`（经 `fields` 传入），不要在 `option-common.ts` 再写一份，也不要用局部网格 Search。列表只听 `route.query`（`useSearchQs` + `usePage()` + `transformSimpleCondition`）。固定条件来自账号树 `props.filter`（含资源 tab 右上角分配状态：未分配 `bk_biz_id eq -1`、已分配 `bk_biz_id neq -1`）与搜索条件 AND；点到具体账号时 `props.filter` 会丢掉 `vendor`，列表用 `vendorInResourcePage` / `selectedAccountId` 把 `vendor`、`account_id` 补全后再请求。业务筛选下拉不含「未分配」；表格里未分配仍是 `bk_biz_id === -1` 展示 `--`。分配复用资源接入 `BatchDistribution`：工具栏走勾选，行内「分配」调用 `open([row])`，同一弹窗、同一 `assign/bizs`；已分配行不可再勾选。弹窗的目标业务来自 `useAccountBusiness(行的 account_id)`，所以列表响应必须带 `account_id`，为空下拉就是空。空值展示 `--`。请求层在 `src/store/load-balancer/exclusive-cluster.ts`（pinia，与 `clb.ts` 同族），DTO 也放那里；页面只做筛选编排，不在 `views/` 下建 `api.ts`；仓库里不放 mock/占位数据，后端未就绪就是空列表，本地造数据用 devtools 改写响应。请求不要传 `globalError: false`——http 层 `handleResponse` 只在 `globalError` 为真时对 `code !== 0` reject，传 false 会把业务错误当成功（分配失败会弹成功提示）。接口挂在负载均衡路径下：`/api/v1/cloud/load_balancers/exclusive_clusters/{list,assign/bizs}`，分配 body 是 `cluster_ids` + `bk_biz_id`（>0），混入已分配集群后端整批拒绝。表格「集群内实例数」取 `clb_resource_count`；响应不含业务名，业务列靠 business 映射解析。切换「负载均衡 / 独占集群」时清掉 `filter`/`page`，避免搜索串台。列宽由 `column.ts` 的 `@Column` `width` 声明：固定形态字段（`tgw-` + 8 位串的集群ID、枚举类）给窄值，名称列不按"完整显示"取值（30～40 字符需 ~300px，会把后面的列挤出首屏），长尾交给 `show-overflow-tooltip`。toolbar 与资源页 CLB 同构，搜索框自适应依赖容器 `min-width: 0`，见 resource 模块 toolbar 那条。

## 实例规格

云上把规格拆成 `exclusive`（是否独占型）与 `sla_type`（性能容量型档位）两个正交字段，后端只回原始值，前端合成一列「实例规格」，位置在「网络类型」之后、不排序。判定优先级：`exclusive === 1` → 独占型；否则 `sla_type` 非空 → `CLB_SPECS` 档位名（与详情页「规格类型」同一套文案）；否则 → 共享型；字段全缺 → `--`。合成与筛选口径只有一份，放模块内 `src/views/load-balancer/utils.ts`（`getLoadBalancerInstanceSpecName` / `buildLoadBalancerInstanceSpecFilterRules`），两个 CLB 列表壳都只接线：业务下模型驱动列表在 `children/display/field-clb.ts` + `children/search/condition-clb.ts`，资源接入老列表在 `use-columns.tsx` + `load-balancer-manage.vue` 的 `clbsSearchData` / `conditionFormatterMapper`。筛选项是 `LOAD_BALANCER_INSTANCE_SPEC_SEARCH_NAME`（同在 `utils.ts`，与规则 builder 必须同步；`{ ...LOAD_BALANCER_INSTANCE_SPEC_NAME, ...CLB_SPECS }`，先独占型 / 共享型两个粗粒度分类，再列各档位）。展示读顶层 `exclusive`/`sla_type`，筛选走 `extension.*`：独占型是 `exclusive json_eq 1`；档位与共享型都必须再叠一条 `exclusive json_eq 0`——展示时 `exclusive === 1` 优先判独占型，档位规则只看 `sla_type` 会筛出列里显示「独占型」的行，共享型漏了 `sla_type=''` 则会把性能容量型算成共享型。档位统一一条 `sla_type json_in [...]`（单选也用 in，`JSONInOp` 接受单元素数组），不逐档 `json_eq`：`pkg/runtime/filter` 每层 rules 上限 10 条，逐档时 10 个选项全选加上 `useFilter` 的外层包装会贴死上限，合并后最多 3 条（独占 / 共享 / 档位各一条）。`instance_spec` 不是后端字段，靠 `filterRules` / `conditionFormatterMapper` 全量改写，不会带进 rules。

## 关键流程 / 注意事项

- 提交仍走 `sync_by_cond`。成功且 `task_management_id` 非空 → Toast「同步任务已创建，可在【任务管理-负载均衡】查看进度」，点击跳 `MENU_BUSINESS_TASK_MANAGEMENT_DETAILS`（`resourceType=clb`，`bizs` 取 `getBizsId()`：store → URL → `localStorage.bizs`）。
- 空 id →「没有可处理的负载均衡」。`2000002` →「该账号地域同步任务进行中」。其它错误用接口 `message`。
- 单条同步预填账号/地域并禁用，body 带该 CLB 的 `cloud_ids`。
- 模块专用 hook 放在 `views/load-balancer/`，不进 `src/hooks`。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
