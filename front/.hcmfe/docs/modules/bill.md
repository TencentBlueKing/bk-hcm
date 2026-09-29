# 账单

> status: drafted · kind: module
> globs: `src/views/bill/**`

账单与账号账单视图。

## 职责

一级菜单「资源运营」下「云账单管理」分组里的三个入口页：

- **一级账号** — `account/account-manage/root-account-list.vue`（`/bill/root-account`）。
- **二级账号** — `account/account-manage/main-account-list.vue`（`/bill/main-account`）。
- **云账单管理** — `bill/index.tsx`，账单汇总 / 明细 / 调整。

一级 / 二级账号原本是「云账号管理」页里的两个页签，现在各自独立成页。两个列表页只是薄壳，内容都复用 `account/account-manage/panel`，按 `accountLevel` 区分搜索项、列、接口和详情侧滑。`Panel` 只在 `setup` 时读一次 `accountLevel`，所以两条路由必须挂**不同的组件文件**：共用一个组件、靠路由 props 传级别时，两路由间切换会复用实例，列表不会刷新。

「录入一级账号」「创建二级账号」是两个非菜单页（`account/create-account/*`，`/bill/root-account/create`、`/bill/main-account/create`），`activeKey` 分别指向一级 / 二级账号菜单。

四个账号页的页头都用布局面包屑（路由 `meta.layout.breadcrumb.show: true`，标题取 `meta.title`），页面里不再自己画标题或 `DetailHeader`。两个新建页用 `meta.menu.relative` 指回各自列表，面包屑上的返回箭头由 `useBack` 按它生成。页面根节点按「面包屑之下的 `view-warp` 无留白」来写：`height: 100%` 加自带 `padding: 24px`，不要再为固定标题栏留 `margin-top`。云账单管理整组不显示面包屑，页签栏就是它的页头。`Panel` 用 `routerAction.redirect` 按路由名（`MENU_BILL_ROOT_ACCOUNT_CREATE` / `MENU_BILL_MAIN_ACCOUNT_CREATE`）跳过去，并透传当前 query。本模块内跳转与重定向一律按路由名，不写路径串。

注意与「业务资源 → 账号 → 云账号管理」（`views/cloud-account-manage`）无关，那是资源账号的管理页。

## 路由

`views/bill/route-config.ts` 导出 `billAccountManage` / `billManage` / `billAccountManageLegacy`，由 `views/index.ts` 汇成纯路由数组 `billViews`；路由名用 `constants/menu-symbol.ts` 的 `MENU_BILL_*`（命名对齐未合入的整体重构）。侧栏菜单（显示名、图标、「云账单管理」分组）不在路由里，在 `common/menu-service.ts` 的 `billMenus` 声明，所以改菜单名或分组去那里，改页面标题、权限、高亮才动路由。通用约定见 `menu-route` 模块。

## 查看权限

| 页面 | 面向 | 侧栏显隐（路由 `meta.checkAuth`） | 访问拦截（`store/common.ts` 的 `pageAuthData.path`） |
| --- | --- | --- | --- |
| 一级账号 | 管理员 | `root_account_find`，无权限隐藏 | `/bill/root-account` → 无权限进 `/403/root_account_find`（联系管理员，无申请按钮） |
| 二级账号 | 普通用户 | **不配**，始终显示 | `/bill/main-account` → 无权限进 `/403/main_account_find` |
| 云账单管理 | 管理员 | `account_bill_find`，无权限隐藏 | 不拦截（历来如此：页面能打开，接口由后端 `AuthorizeWithPerm` 拒绝，只弹错误提示） |
| 两个新建页 | — | — | 不拦截 |

- `checkAuth` 只管侧栏显隐，`pageAuthData.path` 只管路由拦截，两者互不替代，按页面面向谁分别决定配不配。
- 二级账号是普通用户的自助入口：菜单必须始终可见，无权限的人点进去在 403 页自行申请。**不要**给它加 `checkAuth`，否则零权限用户连申请入口都找不到。
- 一级账号的权限需由管理员主动授权，`root_account_find` 在 403 页的 `NO_SELF_APPLY_KEYS` 里：只提示联系管理员授权，不显示「申请权限」按钮。
- 云账单的前端权限点是 `account_bill_find`，后端映射到 IAM 动作 `account_bill_manage`（「云账单-云账单管理」）。
- `main_account_find` 是不带实例的校验，后端走 `AuthorizeAny`，含义是「至少对一个二级账号有查看权限」。二级账号列表本身由后端按账号实例过滤（`cmd/account-server/.../main-account/list.go`），零权限时返回空数组而不报错，所以前端拦截只是把「空表格」换成「403 引导申请」，不改变数据范围。
- 一级账号列表后端是整体校验 `root_account` 查看权限，无权限直接报错。

## 旧地址兼容

`billAccountManageLegacy` 承接旧「云账号管理」时期的地址：

- `/bill/account-manage`（旧书签，**也是 `header-config.ts` 里顶栏「资源运营」的入口地址**）→ 有 `root_account_find` 去一级账号，否则去二级账号，保留 query。没有二级权限的人落到二级账号后，由上表的拦截进 403 申请。
- `/bill/account-manage/first-account`、`/second-account` → 按 name 重定向到两个新建页，vue-router 默认保留 query。

分流要读权限数据，所以写在 `beforeEnter` 而不是 `redirect`：`redirect` 在全局守卫加载权限之前求值，首次加载时拿不到。这条记录还要带 `children: []`，因为 vue-router 类型要求 `component` / `children` / `redirect` 三选一，而带 `redirect` 的记录不执行 `beforeEnter`。副作用是刷新旧地址时，权限接口会多请求一次：重定向产生的那次导航，`from` 仍是起始位置 `/`。

## 云账单管理是布局型页面

`bill/index.tsx` 不是普通列表页，它是下游三个页签的布局层：

- `provide` 三个值给所有子页面：`currentMonth`（默认上个月）、`bill_year`、`bill_month`。子页面一律 `inject` 取月份，不要各自再维护一份。
- 渲染 `bill/header`（月份选择、当月汇率、三个页签）+ `<RouterView>`。
- 页签是 `RouterLink` 按路由名跳转，并透传 `BILL_BIZS_KEY` / `BILL_MAIN_ACCOUNTS_KEY` 两个 query，切页签时筛选条件不丢。

所以它的路由必须保持嵌套 `children`，**不能按任务管理那样打平**——打平会让每个子页面丢掉页签栏和注入的月份上下文。

`summary` 这层（`bill/summary/index.tsx`）只裹了一个 `<RouterView>`，看着像空壳，但**不能删**：「账单汇总」页签指向 `MENU_BILL_MANAGE_SUMMARY`，靠祖先匹配才能在停留于「操作记录」子页时保持高亮。

## 注意事项

- 请求仍走 `src/api/bill`，那是废弃层，只允许老代码继续调用；新增请求按项目约定进 `src/store/<模块>/`。
- 汇总 / 调整等页面的部分操作栏、搜索组件与挂载逻辑经 `@pluginHandler/bill-manage` 注入，会随构建变体替换，改这些行为前先看 `plugin-handler` 模块。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
