# Coding — feat-menu-phase1-rename-bill-group

菜单机制的稳定结构见 `.hcmfe/docs/modules/menu-route.md`（本轮已写实），这里只记本迭代的改动。

## 执行顺序

1. 补符号 — `menu-symbol.ts` 新增 `MENU_BILL_*` 段
2. 建模块路由 — 新建 `views/bill/route-config.ts`，把 bill 的路由迁成去中心化新模式
3. 汇总与接线 — `views/index.ts` 导出 `billViews`，`router/index.ts`、`useChangeHeaderTab.ts`、`home/index.tsx` 三个消费点对齐
4. 改名 — `header-config.ts` 两个 `name` + `lang.ts` 补词条
5. 清理 — 删除已搬空的 `router/module/bill.ts`
6. 自测前跑 `bkdevbuddy_lint`
7. PRD v2 — 「云账号管理」拆成一级账号、二级账号两个菜单项（见 §7）

> 排序依据：2 依赖 1，3 依赖 2，5 必须等 3 全部改完（还有引用就不能删）。4 与其余互不依赖，放在后面单独验证更好定位问题。

## 本期为什么做「中间形态」而不是只包一层分组

按 `menu-route.md` 的三条重构信号对照，本期 **0 命中**，所以不做 menu-service 重构。但项目里已经有一条成熟的**局部迁移配方**（任务管理、操作记录、云账号管理、负载均衡、单据五个模块都已迁完），bill 这一支借这次改层级顺手迁过去，二期做「资源接入的迁移」时就已经有 `billViews` 这种一等集合可搬，不必先拆一遍再搬。

范围严格限定在 bill 一支，**不动** business / resource / service 三支——它们没有本期需求驱动。

## 迁移配方

完整步骤已沉淀到 `menu-route.md`「模块路由迁移配方」。本期对 bill 的关键点：

1. 路由按新模式重写，`views/index.ts` 只导出纯路由数组 `billViews`，**不放分组包装节点**。
2. 菜单在 `common/menu-service.ts` 声明（雏形，见 §3），经 `to-legacy-menus.ts` 喂给现有侧栏。
3. **把新符号加进 `views/home/index.tsx` 的 `getRouteLinkParams` 白名单**，否则相对 path 会让侧栏跳错。这一步 rule 里没写。
4. 补 `/bill` 父路由承载相对路径（原来是顶层 `...bill` 平铺）。

---

## 单据 1: 菜单整合一期：一级菜单更名与云账单分组

**TAPD**: [#1069995598138543785](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138543785)

**文件**: `src/constants/menu-symbol.ts` `src/views/bill/route-config.ts` `src/views/index.ts` `src/common/menu-service.ts` `src/views/home/hooks/to-legacy-menus.ts` `src/router/index.ts` `src/views/home/hooks/useChangeHeaderTab.ts` `src/views/home/index.tsx` `src/router/header-config.ts` `src/language/lang.ts` `src/views/bill/bill/header/index.tsx` `src/views/bill/bill/summary/primary/index.tsx` `src/views/bill/account/account-manage/panel/index.tsx` `src/router/module/bill.ts` `src/store/common.ts` `src/views/bill/account/account-manage/root-account-list.vue` `src/views/bill/account/account-manage/main-account-list.vue` `src/views/bill/account/account-manage/account-list.scss` `src/views/bill/account/account-manage/index.tsx` `src/views/bill/account/account-manage/index.scss` `src/router/meta.ts` `src/hooks/use-breadcrumb.ts` `src/router/module/business.ts` `src/views/bill/account/create-account/create-first-account/index.tsx` `src/views/bill/account/create-account/create-first-account/index.scss` `src/views/bill/account/create-account/create-second-account/index.tsx` `src/views/bill/account/create-account/create-second-account/index.scss` `src/views/error-pages/403.tsx`

**改动点**:

### 1. 符号命名（`src/constants/menu-symbol.ts`）

命名对齐 PR #1897（`refactor-router-menu`）的 `menu-symbol.ts`，前缀 `MENU_BILL_*`，跟功能语义走、不跟展示名走。PR 里已有的同义节点**直接复用它的名字与值**，将来合流时是机械替换；只有当前基线独有的合并页需要新增。

**值仍写成字符串常量**（`export const X = 'menu_bill_manage'`），与本分支现有风格一致；PR #1897 那边是 `Symbol('menu_bill_manage')`，描述串相同，合流时只换构造方式。

| 当前 name | 新符号 | 来源 |
|-----------|--------|------|
| `'account-manage'` | ~~`MENU_BILL_ACCOUNT_MANAGE`~~ → `MENU_BILL_ROOT_ACCOUNT` / `MENU_BILL_MAIN_ACCOUNT` | v1 为合并页新增过 `MENU_BILL_ACCOUNT_MANAGE`；v2 拆页后删除，改为复用 PR #1897 的两个符号（见 §7） |
| `'录入一级账号'` | `MENU_BILL_ROOT_ACCOUNT_CREATE` | 复用 PR #1897 |
| `'创建二级账号'` | `MENU_BILL_MAIN_ACCOUNT_CREATE` | 复用 PR #1897 |
| `'bill-manage'` | `MENU_BILL_MANAGE` | 复用 PR #1897 |
| `'billSummary'` | `MENU_BILL_MANAGE_SUMMARY` | 复用 PR #1897 |
| `'billSummaryManage'` | `MENU_BILL_MANAGE_SUMMARY_MANAGE` | 复用 PR #1897 |
| `'billSummaryOperationRecord'` | `MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD` | 复用 PR #1897 |
| `'billDetail'` | `MENU_BILL_MANAGE_DETAIL` | 复用 PR #1897 |
| `'billAdjust'` | `MENU_BILL_MANAGE_ADJUST` | 复用 PR #1897 |

另加 `MENU_BILL = 'menu_bill'` 作为一级视图父路由的 name（PR #1897 也有）。

**命名雷区**：本分支的 `MENU_RESOURCE` 指「资源接入」，PR #1897 的 `MENU_RESOURCE` 指合并后的「资源运营」，同名不同义。本期**不要**碰 `MENU_RESOURCE`，也不要把账单符号命名成 `MENU_RESOURCE_OPERATION_*`，否则合流时会撞车。

`activeKey` 同步换成新符号常量。

### 2. 模块路由（新建 `src/views/bill/route-config.ts`）

把 `router/module/bill.ts` 的四条路由按新模式重写。v1 时 component 一个不换、URL 一个不变；v2 把账号那三条换成拆页后的五条（§7），下面只列云账单管理：

```ts
export const billManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_MANAGE,
    path: 'bill-manage',
    component: () => import('./bill/index'),
    redirect: { name: MENU_BILL_MANAGE_SUMMARY },  // 重定向一律按 name，summary 层同理指向 MENU_BILL_MANAGE_SUMMARY_MANAGE
    children: [ /* summary / detail / adjust，name 换符号，path 保持相对，层级保持 */ ],
    meta: { ...new Meta({ title: '云账单管理', activeKey: MENU_BILL_MANAGE, checkAuth: 'account_bill_find' }) },
  },
];
```

要注意：

- **`hasPageRoute` 删除**（详见下方「关于 hasPageRoute 与 children」）。
- **`children` 层级保持不变**，不打平（同上）。
- **`checkAuth: 'account_bill_find'` 必须保留**，它是云账单管理的菜单级权限过滤依据。
- **路由上不再写 `icon` 与 `notMenu`**：两者全仓唯一消费方是 `views/home/index.tsx` 的侧栏，bill 的侧栏改由 menu-service 供数后，图标归菜单项、菜单列哪些项由 menu-service 决定（对齐 `fe-deprecated` §2 / §3）。

### 3. menu-service 雏形：菜单与路由分离（只覆盖资源运营一支）

路由侧回归纯路由，菜单侧写成目标形态（PR #1897 的 `IMenu`），中间用一层过渡适配喂给现有侧栏渲染器，渲染器不改。

```ts
// src/views/index.ts —— 纯路由，不再有分组假节点（v2 追加旧地址兼容 billAccountManageLegacy）
export const billViews = [...billAccountManage, ...billManage, ...billAccountManageLegacy];

// src/router/index.ts —— { name: MENU_BILL, path: '/bill', children: billViews }   替代  ...bill

// src/common/menu-service.ts —— IMenu 接口、getMenuRoute、visibility 过滤与 PR #1897 一致；文件头注明本期只收录资源运营
const billMenus: IMenu[] = [
  { id: MENU_BILL_ROOT_ACCOUNT, i18n: '一级账号',   icon: 'bkhcm-icon-account-manage', group: '云账单管理', route: getMenuRoute(billViews, MENU_BILL_ROOT_ACCOUNT) },
  { id: MENU_BILL_MAIN_ACCOUNT, i18n: '二级账号',   icon: 'bkhcm-icon-account-manage', group: '云账单管理', route: getMenuRoute(billViews, MENU_BILL_MAIN_ACCOUNT) },
  { id: MENU_BILL_MANAGE,       i18n: '云账单管理', icon: 'bkhcm-icon-bill-manage',    group: '云账单管理', route: getMenuRoute(billViews, MENU_BILL_MANAGE) },
];
export const getBillMenus = () => filterVisible(billMenus);

// src/views/home/hooks/to-legacy-menus.ts —— 过渡适配，换成按 IMenu 渲染的新侧栏后删除
// 按 group 聚合（与 PR #1897 menu.vue 同一套 Map 逻辑），输出旧渲染器认得的 { path, children, meta: { groupTitle } }
// 该结构只用于渲染，不注册进路由

// useChangeHeaderTab.ts —— case 'bill': menus.value = toLegacyMenus(getBillMenus(), billViews)
// views/home/index.tsx   —— getRouteLinkParams 白名单 += MENU_BILL_ROOT_ACCOUNT, MENU_BILL_MAIN_ACCOUNT, MENU_BILL_MANAGE
```

适配规则：

- 按 `menu.route.name` 找到完整路由记录，复制一份再用菜单项的 `i18n` / `icon` 覆盖 `meta.title` / `meta.icon`；原路由记录不被修改，面包屑等仍读路由自己的 `title`。
- 复制时**去掉 `children`**：旧渲染器对顶层元素按「有 children 即分组」判定，带着页签子路由的菜单项一旦未分组就会被误判。去掉后该判定对 bill 永远不会误触，这也是 `hasPageRoute` 可以彻底删除的原因。
- 未设 `group` 的菜单项原样平铺为一级菜单项；设了 `group` 的聚合成分组，分组顺序按首次出现。
- `title`、`checkAuth`、`activeKey` 仍在路由上，渲染器照旧读取；权限过滤不迁。

结果：「资源运营」侧栏只有一个分组「云账单管理」，下挂一级账号、二级账号、云账单管理三项（F-003 / F-004 / F-005），分组与子项同名（R-002）。

边界：

- **一级菜单名仍在 `header-config.ts`**，不进雏形。顶栏只读 `header-config`（按 `ENABLE_CLOUD_SELECTION` / `ENABLE_ACCOUNT_BILL` 就地过滤、`t(name)` 显示、按 path 跳转），雏形不经过这条链路；在 menu-service 里再写一份一级菜单名会造成同名两处定义。
- **`visibility` 字段保留但本期不赋值**。当前开关只作用于顶栏入口，直接访问 `/bill/...` 时侧栏照常显示；加上会改变现有显隐，与 PRD 冲突。
- **其它一级视图不迁**，仍由路由结构（包装节点 + `groupTitle`）表达菜单，两套机制并存，已写入 `menu-route.md`。

### 4. 跳转一律按 name

- 路由重定向：`bill-manage` → `{ name: MENU_BILL_MANAGE_SUMMARY }`，`summary` → `{ name: MENU_BILL_MANAGE_SUMMARY_MANAGE }`。
- `views/bill/account/account-manage/panel/index.tsx` 的「录入一级账号 / 创建二级账号」按钮：`router.push({ path: '/bill/account-manage/first-account' | 'second-account', query })` 改为 `routerAction.redirect({ name: MENU_BILL_ROOT_ACCOUNT_CREATE | MENU_BILL_MAIN_ACCOUNT_CREATE, query })`，`useRouter` 随之去掉。`routerAction.redirect` 无 options 时等价于 push，仅额外清空项目自有的历史栈；两个录入页返回用的是原生 `router.back()` / `router.go(-1)`，不受影响。
- 不在本期处理：`create-second-account` 创建成功后按 name 跳单据详情（已是 name，目标不是 bill 路由）；`sub-account-selector` 只改当前页 query；各处 `router.back()` / `router.go(-1)`（换成 `routerAction.back` 会改变返回语义，另议）；顶栏 `header-config` 按 path 跳转（顶栏链路本期不动）。

### 4.1 页面内 tab 的符号化

- `views/bill/bill/header/index.tsx` 的 `links` 数组：三个字符串 name 换成 `MENU_BILL_MANAGE_SUMMARY` / `_DETAIL` / `_ADJUST`。
- `views/bill/bill/summary/primary/index.tsx:119`：`router.push({ name: 'billSummaryOperationRecord' })` 换成新符号。**顺带把 `router.push` 改成 `routerAction.redirect`**——`fe-router-action` 明令禁止直接用 `router.push`，既然这一行本来就要改，就一并改正，不留新的违规行。

### 5. 一级菜单更名（F-001 / F-002 / F-006）

`src/router/header-config.ts` **只改两个 `name` 字段**，`id` 与 `path` 一律不动（`id` 是 `useChangeHeaderTab` 的 switch 依据和 `handleHeaderMenuClick` 的入参）：

- `id: 'business'` 的 `name`：`资源管理` → `业务资源`
- `id: 'bill'` 的 `name`：`账号管理` → `资源运营`

数组顺序原样保留（F-006）。

`src/language/lang.ts` 补词条（用户已定译名）：新增 `业务资源: ['Business']`、`资源运营: ['Resource']`；原有 `资源管理` / `账号管理` 两条**保留**，其它页面文案仍在用。

现状是：一级菜单名走 `t(name)`（`home/index.tsx`），五项里三项配了词条（`资源接入` / `资源管理` / `账号管理`），`服务请求` / `资源选型` 没配、英文环境本就显示中文；侧栏 `meta.title` 完全不走 `t()`。我们要改的两项恰好都是配了词条的，所以不补词条等于英文环境从 "Business" / "Account Manage" 退回中文，是回退而非维持现状。

### 6. 清理

`router/module/bill.ts` 搬空后删除。确认 `router/index.ts` 与 `useChangeHeaderTab.ts` 两处 import 都已改掉再删。

### 7. PRD v2：一级 / 二级账号拆成菜单项（F-004 / F-007 / F-008）

形态、符号、地址对齐 PR #1897（`refactor-router-menu`），但 #1897 没做的旧地址兼容本期补上。

**路由**（`views/bill/route-config.ts`）：

| name | path | component | meta |
|------|------|-----------|------|
| `MENU_BILL_ROOT_ACCOUNT` | `root-account` | `account-manage/root-account-list` | title 一级账号，`checkAuth: 'root_account_find'` |
| `MENU_BILL_ROOT_ACCOUNT_CREATE` | `root-account/create` | `create-account/create-first-account`（不变） | `activeKey` → `MENU_BILL_ROOT_ACCOUNT` |
| `MENU_BILL_MAIN_ACCOUNT` | `main-account` | `account-manage/main-account-list` | title 二级账号，**不配 `checkAuth`**（v2.1：面向普通用户，菜单始终显示，无权限点进去由 403 引导申请） |
| `MENU_BILL_MAIN_ACCOUNT_CREATE` | `main-account/create` | `create-account/create-second-account`（不变） | `activeKey` → `MENU_BILL_MAIN_ACCOUNT` |

**旧地址兼容**（新导出 `billAccountManageLegacy`，汇入 `billViews`）：

- `account-manage`：`beforeEnter` 读 `useCommonStore().authVerifyData`，有 `root_account_find` 去 `MENU_BILL_ROOT_ACCOUNT`，否则去 `MENU_BILL_MAIN_ACCOUNT`，带 `query: to.query, replace: true`。不能用 `redirect` 函数：它在全局守卫拉权限之前求值，首次加载拿不到权限。记录带 `children: []`，只为满足类型（`component` / `children` / `redirect` 三选一，带 `redirect` 的记录不执行 `beforeEnter`）。
- `account-manage/first-account` → `redirect: { name: MENU_BILL_ROOT_ACCOUNT_CREATE }`，`second-account` 同理，默认保留 query。
- `header-config.ts` 的顶栏入口仍是 `/bill/account-manage`，**不改**，默认入口和旧书签走同一条分流。

**直接访问拦截**（`store/common.ts`）：`pageAuthData` 里原有的 `root_account_find`、`main_account_find` 两项补上 `path: '/bill/root-account'`、`'/bill/main-account'`，由 `router/index.ts` 现成的全局守卫拦到 `/403/:id`。`useVerify` 拼权限请求时不读 `path`，加上它只影响拦截。新建页不加拦截（PRD：新建页权限规则不变）。

**一级账号的无权限页（PRD v2.2 / Q-006）**：该权限需由管理员主动授权。`views/error-pages/403.tsx` 新增 `NO_SELF_APPLY_KEYS = ['root_account_find']`，命中时顶部提示改为「请联系管理员授权」、补一段管理员权限说明（权限名取 IAM 注册名「云账号-一级账号管理」），并隐藏「申请权限」按钮。其余权限点和嵌入 403 的资源选型页不受影响（`urlKey` 不在列表里）。本分支 `PROJECT_CONFIG` 没有管理员联系地址，只给文字，不做可点击的联系入口；其它构建变体若已有联系人组件，合并时在这一处替换即可，这块与它们的 403 页改动相邻，合并时注意冲突。

**云账单管理的权限不动**：历来只有侧栏 `checkAuth: 'account_bill_find'`（IAM 动作 `account_bill_manage`），`pageAuthData` 无 `path`，直接访问页面能打开、接口由后端拒绝并弹错误提示，本期保持原样。

**页面**：`root-account-list.vue` / `main-account-list.vue` 是两个薄壳（新功能页面用 `.vue`），外层 `account-list.scss` 给出原页签面板的留白，内容都是 `<Panel :account-level=... />`。不照搬 #1897 新写的两个 TSX 列表文件：本分支 `Panel` 的二级账号搜索会动态填充业务下拉候选，#1897 的版本没有，照搬会退化。必须是两个组件文件，不能一个组件靠路由 props 区分：`Panel` 只在 `setup` 时读一次 `accountLevel`，在两路由间复用实例时列表不会刷新。原页签页 `account-manage/index.tsx` 与 `index.scss` 删除（`.account-sideslider-header` 全仓无引用）；`Panel` 去掉已无人传入的 `authVerifyData` prop。

**面包屑（对齐 #1897 新模式）**：拆页后原页签页自画的「云账号管理」标题没了，改用布局面包屑。

- 四个账号路由都加 `layout: { breadcrumb: { show: true } }`；两个新建页补 `title`（录入一级账号 / 创建二级账号）和 `menu: { relative: 列表符号 }`，面包屑返回箭头据此指回列表。
- 两个新建页删掉顶部固定定位的 `DetailHeader`，scss 按 #1897 改成 `height: 100%` 加 `padding: 24px`，去掉为固定标题栏预留的 `margin-top` / `calc(100vh - …)` 与按钮、提示条上的 `ml24` / `mr-24`。与布局无关的 #1897 改动（另一套目录的 import 路径、去掉 `filterable`、颜色写法）不带。
- meta 字段名对齐 #1897：`layout.breadcrumbs` → `layout.breadcrumb`，同步改 `router/meta.ts`、`hooks/use-breadcrumb.ts` 和 `router/module/business.ts` 里仅有的一处用法。**默认值不跟**：#1897 的 `Meta` 默认显示面包屑，本分支仍默认不显示（沿用 `?? meta.isShowBreadcrumb`），否则全站没配的路由都会冒出面包屑。

**菜单与白名单**：`billMenus` 与 `getRouteLinkParams` 白名单都换成两个新符号（见 §3）；`MENU_BILL_ACCOUNT_MANAGE` 删除，全仓已无引用。两项图标都沿用 `bkhcm-icon-account-manage`。

**分流与权限实跑结果**（内存 history 跑真实 `billViews`，全局守卫按 `router/index.ts` 同逻辑复刻，`pageAuthData` 取自真实 `store/common.ts`）：

| 权限 | 旧 `/bill/account-manage` | 直接访问 | 侧栏 |
|------|---------------------------|----------|------|
| 全有 | → `/bill/root-account`（query 保留） | 两个列表都可进 | 一级账号、二级账号、云账单管理 |
| 只有二级 | → `/bill/main-account` | 一级账号 → `/403/root_account_find` | 二级账号 |
| 只有一级 | → `/bill/root-account` | 二级账号 → `/403/main_account_find` | 一级账号、二级账号 |
| 都没有 | → `/403/main_account_find` | — | 二级账号 |

v2.1 去掉二级账号的 `checkAuth` 后，侧栏列按渲染逻辑更新（二级账号不再参与过滤）；`checkAuth` 不参与路由，前三列实跑结果不变。

另验：两个旧新建页地址带 query 落到新建页、`activeKey` 正确；`Panel` 的 `{ name, query }` 跳转、侧栏 `{ name }` 跳转、云账单管理及页签旧地址均不变。临时脚本已删除。

## 关于 hasPageRoute 与 children

**`hasPageRoute` 删除。** 它只有一个消费点：`views/home/index.tsx:274` 的分组判定 `Array.isArray(menuItem.children) && !menuItem.meta?.hasPageRoute`，语义是「我的 children 是页面内子路由，不要把我当分组渲染」。该判定**只对 `menus.value` 的顶层元素生效**；迁移后 `bill-manage` 降成分组的子项，渲染子项时只读 `child.meta.title`，不会再看它的 `children` 或 `hasPageRoute`。全仓也只有 `router/module/bill.ts` 这一处设置过它，删掉不影响其它一级视图（它们从来没设过）。

**`children` 层级保持，不打平。** 任务管理、操作记录是平的，因为它们的列表与详情是两个独立整页、没有共享布局。云账单管理不同：`views/bill/bill/index.tsx` 是布局层，`provide` 了 `currentMonth` / `bill_year` / `bill_month` 给下游页面，并渲染 `<Header />` + `<RouterView>`；打平会让每个页面丢掉页签栏和注入的月份上下文，那是页面组件重构，不在本期范围。

已迁移模块里有同构先例：**负载均衡** `views/load-balancer/route-config.ts` 在新模式下照样用 `children`（`entry-biz.vue` 作布局、`redirect` 指向子路由、子路由 `activeKey` 回指父菜单项）。所以新模式禁止的是**拿 children 当菜单分组**，不是禁止 children 本身。

**`summary` 那层空壳也不能删**：`summary/index.tsx` 虽然只裹了一个 `<RouterView>`、`.bill-summary-module` 也没有任何样式，但「账单汇总」页签是 `RouterLink to={{ name: billSummary }}`，靠祖先匹配才能在停留于 `operation-record` 时保持高亮；打平后高亮会失效。

## 不改的东西

- `views/home/index.tsx` 的分组渲染逻辑（能力已具备，只加白名单成员）。
- 云账单管理及其页签的 URL、component、`checkAuth`（账号页的地址与权限按 v2 调整，见 §7）。
- `header-config.ts` 的顶栏入口地址（仍指向旧地址，由分流承接）。
- 「录入一级账号 / 创建二级账号」按钮的操作级权限（PRD 范围外）。
- 「业务资源」下的四个分组与菜单项（R-004）。`business.ts` 只做了面包屑键名的机械改名，行为不变。
- 不新增父单列出的第三个账单类菜单项（R-003，该菜单项不在本分支注册范围内）。
- `useWhereAmI` 的 `/^\/bill\/.+$/` 判定（URL 不变，无需调整）。

## 风险与验证点

| 风险 | 说明 | 怎么验 |
|------|------|--------|
| 相对路径 + 白名单漏配 | 漏加白名单则侧栏 RouterLink 拿相对 path 跳转，跳错或跳空 | 从侧栏点三个菜单项，确认落到 `/bill/root-account`、`/bill/main-account`、`/bill/bill-manage/summary/manage` |
| 旧地址分流 | 首次加载时权限数据未就绪会分错 | 分别用「有一级权限」「只有二级权限」的账号刷新 `/bill/account-manage`，以及点顶栏「资源运营」 |
| 访问拦截 | `checkAuth` 只管侧栏、`pageAuthData.path` 只管路由，按页面面向谁分别配 | 无一级权限：侧栏无「一级账号」、直接访问进 403 且只提示联系管理员、无申请按钮；无二级权限：侧栏仍有「二级账号」，点击或直接访问都进 403 且可申请 |
| 新建页布局 | 去掉固定 `DetailHeader` 后改由面包屑占位，scss 偏移若没清干净会留白或被遮挡 | 两个新建页：面包屑显示标题与返回箭头、表单顶部不留空、二级账号右侧指引栏高度铺满、页面可滚动到提交按钮 |
| 刷新旧地址多一次权限请求 | 重定向产生的导航 `from` 仍是 `/`，全局守卫会再拉一次 | 可接受；Network 面板里看到两次 `auth/verify` 属预期 |
| 父路由无 component | `{ name: MENU_BILL, path: '/bill', children }` 与 business 同构，但 bill 原来是顶层平铺 | 控制台无 vue-router 警告；直接输入旧 URL 能打开 |
| 高亮失效 | `activeKey` 换成新常量，层级也变了 | 逐项点击 + 旧 URL 直接进入，两种方式都要正确高亮 |
| 页面内三个页签 | name 全换，`RouterLink` 与 `routerAction` 两处引用 | 汇总/明细/调整互相切换；汇总页「操作记录」跳转正常 |
| 权限过滤 | 一级账号、云账单管理各自的 `checkAuth`；二级账号不过滤 | 无一级 / 云账单权限时对应项不出现；二级账号始终在，分组始终可见（AC-007 / AC-011 / AC-012） |
| 英文环境 | 新名称缺词条会回退中文 | 切英文确认两个一级菜单显示新译名 |
| 一级菜单顺序 | 只改 name 不动数组 | 对照改动前顺序（AC-010） |

## 待确认

无。
