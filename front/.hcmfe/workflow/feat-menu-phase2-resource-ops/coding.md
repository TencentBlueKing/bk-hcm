# Coding — feat-menu-phase2-resource-ops

需求文档：仓库根 `docs/reqs/二期资源运营重组.md`（TAPD [#1069995598138702368](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702368)），验收以其中归属本单的 AC 为准（见需求文档「交付归属」；标「已移至资源纳管主体」的不在本单验收）。菜单机制见 `.hcmfe/docs/modules/menu-route.md`，账单见 `.hcmfe/docs/modules/bill.md`。

**目标**：首批 PR 的最后一个 commit（⑦）。资源运营翻转、菜单树、资源接入入口、迁移提示条、资源管理相关站内链接**已由 ③ 交付**（用户 2026-10-10 决定：每单上线即用新地址），④⑤⑥ 各自注册新页面并转址自己的旧地址。本单只剩：账单从 `/bill/*` 迁到 `/resource/bill/*` 并兼容旧地址、`useWhereAmI`、删除 ③～⑥ 留下的过渡布局与旧路由模块、删除无引用的旧文件、全量终扫。

## 跨单约定

提交顺序：① 布局 → ② 视图权限 → ③ 资源纳管 + 资源运营翻转 → ④ 回收管理 / 操作记录 → ⑤ 账号管理 → ⑥ 资源详情迁移 → ⑦ 本单。

| 项 | 约定 |
|----|------|
| 菜单符号 | ① 起为 PR #1897 的 `Symbol`；③ 已新增 `MENU_RESOURCE_ACCESS`、删除 `MENU_BILL`，`MENU_RESOURCE` 自 ③ 起指资源运营根。本单删除 `MENU_RESOURCE_RESOURCE_MANAGEMENT`（旧 `/resource/resource` 的 name，PR 也已去掉），删完 `rg "MENU_BILL\b\|MENU_RESOURCE_RESOURCE_MANAGEMENT" src` 为 0 |
| ③ 提供 | `/resource` 根（redirect `MENU_RESOURCE_MANAGE`）与 `resourceViews` 汇总；顶栏资源接入入口（菜单项直接链到资源纳管，无路由记录）；资源运营最终菜单树；账单记录已改为绝对路径 `/bill/*` 挂在 `/resource` 根下，裸 `/bill` 暂留 redirect；`resourceManageLegacy`（5 个旧申请页）与 `toResourceList`（挂在旧 `/resource/resource` 记录的空路径子路由上）；迁移提示条；资源管理相关站内链接已改；旧资源接入页面过渡布局（`route.meta.legacyResourceAccess`、两栏云账号树容器、右栏面包屑） |
| ④⑤⑥ 提供 | 各自注册新页面与旧地址转址、改站内链接、把自己的菜单项加入资源运营侧栏，并从旧 `resource.ts` / `resource-inside.ts` / `service-inside.ts` 撤掉已迁走的记录；⑥ 撤掉旧详情 `resourceDetail` 后 `main-content.vue` 的 `legacyBreadcrumbRoutes` 只剩旧操作记录详情（④ 撤掉后为空） |
| 本单独占 | 账单 `/resource/bill/*` 与 `/bill/*` 旧地址、`useWhereAmI`、删除过渡布局与旧路由模块、`toResourceList` 改为独立记录、删除旧文件、全量终扫 |

## 执行顺序

1. 账单路由：`views/bill/route-config.ts` 路径改为相对 `/resource` 的 `bill/*`，旧地址改为绝对路径的 `billLegacy`；`router/index.ts` 注册 `billLegacy`，删除 ③ 的裸 `/bill` redirect
2. `useWhereAmI`
3. 旧路由模块收尾：`toResourceList` 改为独立的 `{ path: '/resource/resource', redirect: toResourceList }`；删除 `router/module/resource.ts`、`resource-inside.ts`、`service-inside.ts`；`MENU_SERVICE` 的 children 去掉 `serviceInside`
4. 布局去过渡：删除 `main-content.vue` 中标 `TODO(菜单整合二期·组装)` 的代码与 `app-shell.scss` 的旧云账号树容器样式；`header.vue` 清账号选择的条件随筛选写入地址单处理，本单不动
5. 删除 `MENU_RESOURCE_RESOURCE_MANAGEMENT` 与旧文件（§4）
6. `bkdevbuddy_lint`、`npm run build`、`rg` 终扫

> 排序依据：5 必须在 3、4 把引用清完之后。

## 改动点

### 1. 账单迁址（`src/views/bill/route-config.ts`、`src/router/index.ts`）

- 路径由 ③ 的绝对 `/bill/*` 改为相对 `/resource` 的 `bill/root-account`、`bill/root-account/create`、`bill/main-account`、`bill/main-account/create`、`bill/manage`（子路由 `summary`→`manage`/`operation-record`、`detail`、`adjust` 原样）。组件、name、`activeKey`、`menu.relative`、四个账号页的 `layout.breadcrumb.show: true`、嵌套结构保留 HEAD 写法；账单管理改名「账单管理」已由 ③ 完成。
- 每条记录补 `owner: MENU_RESOURCE`（与 PR 一致）。
- `billAccountManageLegacy` 改名 `billLegacy`，全部写成绝对路径（见对照表账单段），由 `router/index.ts` 顶层注册（不再并入 `resourceViews`）；除分流记录外都是 `redirect: (to) => ({ name, query: to.query })`。只有 `/bill/account-manage` 保留 `beforeEnter` 分流：有一级账号查看权限去 `MENU_BILL_ROOT_ACCOUNT`，否则 `MENU_BILL_MAIN_ACCOUNT`，`{ name, query: to.query, replace: true }`；权限读取方式以 ② 落地后的写法为准。裸 `/bill` 不再注册，落 404（删除 ③ 的 redirect）。
- 账单内部跳转已全部按 name，路径变了不受影响；`plugin-handler/bill-manage.tsx` 只读 query。逐项回归，不改代码。

### 2. `useWhereAmI`（`src/hooks/useWhereAmI.ts`）

只取 PR 的一行：在 `/^\/resource\/.+$/` 之前加 `/^\/resource\/bill\/.+$/ → Senarios.bill`。`useTable` 按 `Senarios.bill` 切账单接口、`use-columns` 等按 `isResourcePage` 分支，靠它保证账单页不被当成资源纳管。`/bill/` 行保留（兜底）；`getBizsId`、`/403/` 行不按 PR 改。

### 3. 旧路由模块与过渡布局退场

- `views/index.ts` 的 `resourceViews` 去掉 `resource`，删除 `router/module/resource.ts`（`resource-inside.ts` / `service-inside.ts` 已由 ⑤ 删除）。
- 旧 `/resource/resource` 记录删除后，③ 的 `toResourceList` 改为独立记录 `{ path: '/resource/resource', redirect: toResourceList }`（并入 `resourceManageLegacy`，`toResourceList` 不再导出）；`/resource/resource/{record,recycle,account*}` 由 ④⑤ 的旧地址数组承接。
- `main-content.vue`：删除 ① / ③ 为旧资源接入页面写的过渡（`legacyResourceAccess` 判断、云账号树容器及对旧 `account/vendor-group` 的引用；右栏过渡面包屑 `legacyBreadcrumbRoutes` 等已由 ⑥ 删除，若有残留一并删），删完 `rg "菜单整合二期·组装" src` 为空。`app-shell.scss` 删除旧云账号树容器样式。之后云账号树与右栏面包屑只存在于 ③ 的 `resource-entry.vue`。`layout.sideMenu` 字段保留（资源选型仍用）。

### 4. 删除旧文件

本分支不留 `@deprecated` 副本，无引用即删（见 ③ coding.md「对照 PR 的旧文件与过渡逻辑退场」）：旧资源接入页 `views/resource/resource-manage/resource-manage.vue`、旧云账号树目录 `views/resource/resource-manage/account/`（`vendor-group/index.vue`、`accountList/**`；`vendor.plugin.ts` / `vendor-property.ts` 已由 ③ 移入新目录）、`views/resource/resource-manage/resource-entry.vue`（旧外壳）、随之成为死代码的 `RESOURCE_TABS` / `RESOURCE_DETAIL_TABS`。删前逐个 `rg` 确认只剩被删文件自身引用。

**不删**：旧回收站 `views/resource/recyclebin-manager/` 仍是业务回收站页面（④ 只替代了资源侧），随业务回收站并入共享页面归「二期收尾体验」F-010。被删页面引用的 manage 组件、表单、`account-import` 等仍被 `resource-list.vue` 等使用，无新增死文件。服务请求里无引用的 `service-apply/components/basic-resource`（含 `'applyAccount'` 跳转）是早于本期的死代码，属三期工作台范围，不在本单处理。

### 5. 终扫

| 检查 | 期望 |
|------|------|
| `rg "MENU_BILL\b\|MENU_RESOURCE_RESOURCE_MANAGEMENT\|'resourceDetail'\|'accountDetail'\|'accountAdd'\|'applyAccount'" src` | 0（2026-10-12 实测：仅剩无引用的 `basic-resource` 里的 `'applyAccount'`，见 §4） |
| `rg "/resource/resource\|/resource/detail/\|/#/resource/account/detail\|'/bill/" src` | 只剩旧地址数组 |
| `router.getRoutes()` 的 name | 无重复（2026-10-12 按路由记录静态核对无重复） |
| `rg "菜单整合二期·组装" src` | 0（实测为 0；`header.vue` 另有筛选写入地址单的 TODO，不属本单） |

**文件**：`src/views/bill/route-config.ts` `src/views/index.ts` `src/router/index.ts` `src/router/module/resource.ts` `src/router/module/resource-inside.ts` `src/router/module/service-inside.ts` `src/views/resource-manage/route-config.ts` `src/hooks/useWhereAmI.ts` `src/components/layout/main-content.vue` `src/style/app-shell.scss` `src/constants/menu-symbol.ts`，以及 §4 所列删除文件

## 全量旧地址对照表

「注册单」指该转址记录随哪一单生效；本单只新增账单段与 `/resource/resource` 独立记录。

| 旧地址 | 新地址 / 处置 | 注册单 |
|--------|--------------|------|
| `/resource`、`/resource/`（旧：账号列表 `alias: ''`） | 资源运营根，redirect 资源纳管 → `/resource/manage/host` | ③ |
| `/resource/resource`（含 `?type=&accountId=`） | `/resource/manage/<type>`，`accountId` 保留；无 type → host | ③（空路径子路由）→ ⑦（独立记录） |
| `/resource/service-apply/{cvm,vpc,disk,subnet,clb}` | `/resource/manage/service-apply/*`，参数保留 | ③ |
| `/resource/resource/record` | `/resource/record` | ④ |
| `/resource/resource/recycle` | `/resource/recyclebin`，`accountId` 作初始条件 | ④ |
| `/resource/recyclebin`、`/resource/record/details` | 同址，新页面 | ④ |
| `/resource/resource/account` | `/resource/account` | ⑤ |
| `/resource/resource/account/detail?accountId=`、`…/account/manage?accountId=` | `/resource/account/details/<id>` | ⑤ |
| `/resource/resource/account/resource?accountId=` | 资源纳管 `?accountId=<id>&showAccountStatus=1` | ⑤ |
| `/resource/account` | 同址，账号管理新页 | ⑤ |
| `/resource/account/add`、`/service/service-apply/account-add` | `/resource/account/check-in`，参数保留 | ⑤ |
| `/resource/account/detail?accountId=\|id=` | `/resource/account/details/<id>` | ⑤ |
| `/resource/detail/:type?id=&type=` | `/resource/manage/<type>/details/<id>`；缺 id → 对应列表 | ⑥ |
| `/bill/account-manage` | 有一级账号权限 → `/resource/bill/root-account`，否则 `/resource/bill/main-account`，query 保留 | ⑦ |
| `/bill/root-account`、`/bill/main-account` | `/resource/bill/root-account`、`/resource/bill/main-account` | ⑦ |
| `/bill/root-account/create`、`/bill/account-manage/first-account` | `/resource/bill/root-account/create` | ⑦ |
| `/bill/main-account/create`、`/bill/account-manage/second-account` | `/resource/bill/main-account/create` | ⑦ |
| `/bill/bill-manage`、`…/summary`、`…/summary/manage` | `/resource/bill/manage/summary/manage` | ⑦ |
| `/bill/bill-manage/summary/operation-record` / `detail` / `adjust` | `/resource/bill/manage/summary/operation-record` / `detail` / `adjust` | ⑦ |
| 裸 `/bill`、其它 `/bill/*`、`/resource/<未知>` | 404（① 的 catch-all）；裸 `/bill` 在 ③～⑥ 期间仍转 `/bill/account-manage` | ⑦ |

## 从 PR #1897 取 / 不取

| 项 | 取 / 不取 | 说明 |
|----|----------|------|
| 账单 `bill/*` 路径、`owner: MENU_RESOURCE` | 取 | 组件、breadcrumb、`menu.relative` 保留 HEAD；改名已由 ③ 做 |
| 裸 `/bill` 落 404 | 取 | 站内无链接，现网本就是空白页 |
| `useWhereAmI` 优先匹配 `/resource/bill/` | 取 | 只取这一行 |
| 旧路由模块退场 | 取（且直接删除文件） | PR 只标 `@deprecated` |
| PR 删掉 `/bill/*` 兼容 | 不取 | 本单补 `billLegacy` |
| 资源运营根、菜单树、入口、提示条 | 已由 ③ 取 | — |

## HEAD 新增能力必须保留

- 一期账单：一级 / 二级账号是**两个不同组件文件**（`root-account-list.vue` / `main-account-list.vue`）；四个账号页显式 `layout.breadcrumb.show: true`；新建页 `menu.relative`；账单管理的 `summary` 空壳层保留（页签祖先高亮）；`/bill/account-manage` 的 `beforeEnter` + `children: []` 分流写法。PR 的账单路由没有这些，不能整文件替换。
- ① 的递归 `getMenuRoute`、`filterVisible`、`layout.sideMenu`（资源选型仍用）。

## 风险与验证

| 风险 | 验证 |
|------|------|
| 账单被当成资源页（出现云账号树、走资源接口） | 账单各页、页签、新建、调整、汇总跳操作记录逐项走一遍 |
| 分流在首次加载拿不到权限 | 有 / 无一级账号权限各自刷新打开 `/bill/account-manage` |
| 账单旧地址 query 丢失 | 对照表账单段每行带 query 打开，新地址 query 一致；裸 `/bill`、`/bill/unknown` 进 404 |
| 拦截随路径搬家失效 | 无一级 / 二级账号权限时分别访问新旧地址 |
| 删除旧模块后旧资源地址断链 | 对照表 ③～⑥ 段全部重走一遍（这些地址由各单注册，本单删旧记录后不能退化）；裸 `#/resource` 仍进资源纳管 |
| 删除过渡布局后残留 | 终扫表全部满足；资源运营各页无第二条面包屑、无旧云账号树 |
| 完整菜单树（需求 AC-001） | ③④⑤ 各自加入菜单项后，侧栏自上而下为云资源（资源纳管、账号管理、回收管理）→ 云账号管理（一级账号、二级账号、账单管理）→ 操作记录；除资源纳管外各页左侧无云账号树；`menu-service.ts` 里 ③ 留的 TODO 已删 |

## 已确认

1. 2026-10-08：资源详情页迁移作为 ⑥ 并入首批；裸 `/bill` 与 PR 一致进 404；菜单常量从 ① 起为 PR 的 `Symbol`。
2. 2026-10-09：顶栏激活态按 PR 用 `.router-link-active`。2026-10-10 改：资源接入入口不再注册 `/resource-access` 转址记录，菜单项直接链到资源纳管、自身不参与高亮。
3. 2026-10-10：每单上线即用新地址；资源运营翻转、菜单树、入口、提示条、资源管理相关站内链接移入 ③，本单收缩为账单迁址、清理与终扫。

## 待确认

无（提示条下线日期与文案随 ③ 发版前确认）。
