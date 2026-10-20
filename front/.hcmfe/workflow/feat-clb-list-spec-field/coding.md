# Coding — feat-clb-list-spec-field

## 执行顺序

1. 模块常量 + 可复用方法（展示值合成 / 查询条件构建）
2. 模型驱动列表（业务下 CLB）：列 + 搜索条件
3. 老写法列表（资源接入入口）：列 + 搜索条件

> 排序依据：两个列表共用同一份展示/筛选口径，先把口径落成模块内通用方法，两个列表壳只做接线，避免各写一份判定。

## 落码入口

- lite 链，无 prd/design/api 产物。需求真相源：TAPD 单据正文（F-004 + AC-001~004）+ PR #2078 的 `list_load_balancer_with_delete_protection.md` / `list_biz_load_balancer_with_delete_protection.md`。
- 无新页面、无新路由、无新接口。两个列表都已在用 `load_balancers/with/delete_protection/list`，本轮只加列与筛选项。
- 模型驱动列表走 `comp-field-model`：列进 `children/display/field-clb.ts`，筛选进 `children/search/condition-clb.ts`。老列表沿用 `use-columns.tsx` + 页面内 `clbsSearchData`，**不**顺手迁模型。
- 复用形态选「模块内通用方法」而非组件：两个列表壳一个是 `display-value` + `column.render`、一个是 JSX `render`，都只需要一个字符串；组件会强行绑定渲染层。方法放 `views/load-balancer/utils.ts`（模块内新建，与 `constants.ts` 同级），`use-columns.tsx` 已有 `@/views/load-balancer/constants` 的先例，跨目录引用不新增耦合方向。

## 口径决策（与 TAPD 正文的差异已确认）

- 位置：按 TAPD 正文放「网络类型」之后（口述里的「运营商之后」已当面改回正文口径）。
- 展示：按 API 文档三段式，覆盖 TAPD 正文「只展示独占/共享、不展示购买用规格类型」——`exclusive === 1` → 独占型；否则 `sla_type` 非空 → 档位名；否则 → 共享型；再否则（都没有）→ `--`（AC-004）。
- 档位文案复用 `CLB_SPECS`（标准型规格 / 高阶型1规格 / 超强型1规格…），与详情页「规格类型」一致，不另起一套去「规格」后缀的文案。
- 搜索项：初版只给「独占型 / 共享型」两项；**验收阶段已推翻该决策**，现为「独占型 + 8 个档位 + 共享型」共 10 项，见下方[验收调整]。
- 不排序（AC-003）：两个列表的该列都不标 `sort`。

## 共享改动 / 提交策略

- 跨单公共改动：`views/load-balancer/utils.ts`、`views/load-balancer/constants.ts` 会被「CLB详情独占集群字段」等兄弟单复用（展示口径同一份）
- 提交：本分支上一张子单（独占集群管理页面）改动尚未提交，本单文件与其无重叠，提交时按单拆开，带 `--story=1069995598138114064`

---

## 单据 1: CLB列表实例规格字段

**TAPD**: [#1069995598138114064](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138114064)

**文件**: `src/views/load-balancer/constants.ts` `src/views/load-balancer/utils.ts` `src/views/load-balancer/children/display/field-clb.ts` `src/views/load-balancer/children/search/condition-clb.ts` `src/store/load-balancer/clb.ts` `src/views/resource/resource-manage/hooks/use-columns.tsx` `src/views/resource/resource-manage/children/manage/load-balancer-manage.vue`

**改动点**:

- `constants.ts`：`LoadBalancerInstanceSpec`（`exclusive` / `shared`）+ `LOAD_BALANCER_INSTANCE_SPEC_NAME`，沿用模块内 `LoadBalancerIsp` / `LOAD_BALANCER_ISP_NAME` 的命名与形态（纯 `Record<string, string>`，`enum` 搜索项与 `option` 都能直接吃）。
- `utils.ts`（新建）：
  - `getLoadBalancerInstanceSpecName(data)` — 展示值合成，判定顺序即 API 文档的优先级；`sla_type` 落不到 `CLB_SPECS` 时兜原始值，字段全缺时 `--`。
  - `buildLoadBalancerInstanceSpecFilterRules(value)` — 查询条件。独占型是单条 `extension.exclusive json_eq 1`；共享型必须两条 `and`（`extension.exclusive json_eq 0` + `extension.sla_type json_eq ''`），否则会把性能容量型也算进共享型。多选时外层 `or` 包一层，单选直接返回该条，避免多套一层空 `or`。展示读顶层 `exclusive`/`sla_type`，筛选走 `extension.*` + `json_eq`，两边字段路径不同是后端契约本身如此。
- `field-clb.ts`：`lb_type` 之后插 `instance_spec`（`string`，width 100，不标 `sort`）。值由两个字段合成，所以 `render: ({ row }) => …` 直接写在 `@Column` 里（同模块 `field-rs.ts` 的「端口」已是这个写法），`data-list.vue` 的 `bk-table-column :render` 优先于默认 slot 的 `display-value`，`row.instance_spec` 不存在也不影响。
- `condition-clb.ts`：`lb_type` 之后插 `instance_spec`（`enum` + `option` + `meta.search.filterRules`）。`instance_spec` 不是后端字段，靠 `filterRules` 全量改写，`transformSimpleCondition` 不会把 id 带进 rules。
- `use-columns.tsx`：`lb` 列表在「网络类型」之后插「实例规格」，`isDefaultShow: true`，`render` 取 `data`（该文件 render 入参约定用 `data` 而非 `row`）。
- `load-balancer-manage.vue`：`clbsSearchData` 在 `lb_type` 之后插 `instance_spec`（`children` 由 `LOAD_BALANCER_INSTANCE_SPEC_NAME` 展开），并在 `useFilter` 的 `conditionFormatterMapper` 里接同一个 builder。`useFilter` 会把 formatter 结果再包一层 `{ op: or, rules: [...] }`，共享型因此是 3 层嵌套 —— `pkg/runtime/filter` 的 `Expression` 本身实现了 `RuleFactory`、支持递归嵌套，只限每层 rules 条数，不限深度。
- `store/load-balancer/clb.ts`：`ILoadBalancerWithDeleteProtectionItem` 补 `exclusive?: number` / `sla_type?: string`（可选，后端未上环境时就是缺）。
- 顺带清理（`eslint --fix --report-unused-disable-directives` 的结果，非本单需求）：`load-balancer-manage.vue` 里对 `useWhereAmI()` 的 `vue/no-dupe-keys` disable（props 无同名 key，早已失效）、`use-columns.tsx` 里两条 `no-nested-ternary` disable（文件顶部已有文件级 disable）。
- 两个列表的表格列设置都是每次由 columns 现算（`use-table-settings.ts` / `generateColumnsSettings` 都不落 localStorage），新列默认可见，老用户不需要手动勾。
- 其它消费 `DisplayFieldClb` / `SearchConditionClb` 的地方（`batch-delete-dialog.vue`、`load-balancer-list.vue`、`details.vue`）都按显式 id 列表取字段，加列不会漏进去。

## 已知限制

- web-server 现在**不返回** `exclusive` / `sla_type`（PR #2078 只有 api-docs + DAO + 表结构，接口层在兄弟单「负载均衡-购买支持独占集群-后端」）。所以当前环境两个列表该列都显示 `--`，筛选「独占型」大概率空列表 —— 这正是 AC-004 的期望态，按确认不加 mock，也不往共用 CLB store 里塞占位数据。
- 后端上环境后需要复验的是：`extension.exclusive` / `extension.sla_type` 两个 `json_eq` 是否被列表接口的 RuleFields 放行（老列表那条 3 层嵌套 filter 一并验）。

## 验收调整（分支 `feat-clb-exclusive-cluster-mgmt-patch1`）

### 实例规格筛选项补齐性能容量型档位

**现象**：验收确认原决策（筛选只给独占型 / 共享型）不够用，非独占但有档位名的规格也要能筛。属需求变更，不是缺陷；原「不做」条目因此作废。

**修复**：筛选项从 2 项扩到 10 项，两个列表共用同一份映射，`utils.ts` 一处改动覆盖业务侧与资源接入侧。

- 新增 `LOAD_BALANCER_INSTANCE_SPEC_SEARCH_NAME`（放 `utils.ts`，与它必须同步的规则 builder 同文件，不放 `constants.ts`——那边是无 import 的纯常量文件，而这份映射要 spread `CLB_SPECS`）。顺序是先独占型 / 共享型两个粗粒度分类，再列 8 个档位，实现上就是 `{ ...LOAD_BALANCER_INSTANCE_SPEC_NAME, ...CLB_SPECS }`。
- 档位取 `CLB_SPECS` **全部 8 项**，直接 spread，不另维护一份子集。
- 档位与共享型的规则都叠加 `extension.exclusive json_eq 0`。展示时 `exclusive === 1` 优先判独占型，档位规则若只看 `sla_type`，会把列里显示「独占型」的行筛出来，筛选与展示口径就不一致了。
- 档位统一走**一条** `extension.sla_type json_in [...]`，选一个也不退回 `json_eq`（`JSONInOp.ValidateValue` 只要求数组非空、元素是基础类型，单元素合法，SQL 就是 `IN (?)`，没必要为此多一个分支）。逐档 `json_eq` 时 8 个档位加独占、共享同层就是 10 条，正好顶到 `pkg/runtime/filter` 的每层上限（`len(rules) > 10` 才报错，等于 10 是零余量）；合并后固定最多 3 条。
- 多值是真实可达的，两套搜索壳走的是不同路径：`ResourceSearchSelect` 硬编码 `unique-select`（同字段只能一个 tag），但 `mapSearchFieldsToItems` 对 `enum` 自动置 `multiple: true`，一个 tag 内可勾多项；CLB 两个列表的壳（业务侧 `children/search/search.vue`、资源接入侧 `useFilter`）每个 tag 只能选一个值，但都没开 `unique-select`，同字段可以加多个 tag。前者经 `getSimpleConditionBySearchSelect` 合成同一个数组交给 `filterRules`，后者由 `useFilter` 按字段归成 `{op: OR, rules: [...]}`、builder 每个 tag 各调一次。
- 语义与 `enum` 默认规则（`getDefaultRule` 给的 `IN` + 数组）一致：字段之间 AND（`transformSimpleCondition` 顶层 `op: and`；`useFilter` 与 `props.filter.rules` 合并），同一字段的多个值 OR。
- 共享型语义保持不变（`exclusive=0` + `sla_type=''`），否则它会囊括所有档位、与并列的档位选项重复。

**文件**: `src/views/load-balancer/utils.ts` `src/views/load-balancer/children/search/condition-clb.ts` `src/views/resource/resource-manage/children/manage/load-balancer-manage.vue`

**验收**：两个列表的「实例规格」下拉都出 10 项；单选档位、多选档位、档位叠加独占型/共享型各筛一次，确认请求 rules 与列展示一致。后端未返回 `exclusive` / `sla_type` 的环境上仍是空结果（AC-004 预期态）。

## 不做

- 后端 Go / `docs/api-docs`
- CLB 详情页「实例规格」、单据展示、购买页独占型规格（兄弟单，另有处理人）
- 老列表迁模型驱动
