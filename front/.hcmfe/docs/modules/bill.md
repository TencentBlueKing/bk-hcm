# 账单

> status: drafted · kind: module
> globs: `src/views/bill/**`

账单与账号账单视图。

## 职责

一级菜单「资源运营」下「云账单管理」分组里的三个入口页：

- **一级账号** — `account/account-manage/root-account-list.vue`（`/resource/bill/root-account`）。
- **二级账号** — `account/account-manage/main-account-list.vue`（`/resource/bill/main-account`）。
- **账单管理** — `bill/index.tsx`，账单汇总 / 明细 / 调整（页内页签栏标题仍为「云账单管理」，与 PR 一致）。

一级 / 二级账号原本是「云账号管理」页里的两个页签，现在各自独立成页。两个列表页只是薄壳，内容都复用 `account/account-manage/panel`，按 `accountLevel` 区分搜索项、列、接口和详情侧滑。`Panel` 只在 `setup` 时读一次 `accountLevel`，所以两条路由必须挂**不同的组件文件**：共用一个组件、靠路由 props 传级别时，两路由间切换会复用实例，列表不会刷新。

「录入一级账号」「创建二级账号」是两个非菜单页（`account/create-account/*`，`/resource/bill/root-account/create`、`/resource/bill/main-account/create`），`activeKey` 分别指向一级 / 二级账号菜单。

四个账号页的页头都用布局面包屑（路由 `meta.layout.breadcrumb.show: true`，标题取 `meta.menu.i18n`），页面里不再自己画标题或 `DetailHeader`。两个新建页用 `meta.menu.relative` 指回各自列表，面包屑上的返回箭头由 `useBack` 按它生成。页面根节点按「面包屑之下的主视图 `.main-view` 无留白」来写：`height: 100%` 加自带 `padding: 24px`，不要再为固定标题栏留 `margin-top`。云账单管理整组不显示面包屑，页签栏就是它的页头。`Panel` 用 `routerAction.redirect` 按路由名（`MENU_BILL_ROOT_ACCOUNT_CREATE` / `MENU_BILL_MAIN_ACCOUNT_CREATE`）跳过去，并透传当前 query。本模块内跳转与重定向一律按路由名，不写路径串。

注意与「业务资源 → 云账号」（`views/cloud-account-manage`）无关，那是资源账号的管理页。

## 路由

`views/bill/route-config.ts` 导出 `billAccountManage` / `billManage`（并入 `views/index.ts` 的 `resourceViews`，相对资源运营 `/resource` 根的 `bill/*`，即 `/resource/bill/root-account`、`/resource/bill/manage/summary/manage` 等，各记录 `owner: MENU_RESOURCE`）与 `billLegacy`（旧 `/bill/*` 绝对路径转址，由 `router/index.ts` 顶层注册）。`useWhereAmI` 先匹配 `/resource/bill/` 判为 `Senarios.bill`（`useTable` 据此切账单接口），再匹配 `/resource/`。裸 `/bill` 不再注册，落 404。路由名用 `constants/menu-symbol.ts` 的 `MENU_BILL_*`（命名对齐未合入的整体重构；一级父路由 `MENU_BILL` 已删除）。侧栏菜单（显示名、图标、「云账单管理」分组、`ENABLE_ACCOUNT_BILL` 显隐）不在路由里，在 `common/menu-service.ts` 的资源运营菜单树声明，所以改菜单名或分组去那里，改页面标题、权限、高亮才动路由。通用约定见 `menu-route` 模块。

## 查看权限

| 页面 | 面向 | 侧栏显隐（`constants/view-auth.ts` 的 `hideMenu`） | 访问拦截（路由 `meta.auth.view`） |
| --- | --- | --- | --- |
| 一级账号 | 管理员 | `AUTH_FIND_ROOT_ACCOUNT`，无权限隐藏 | 声明，无权限显示框架内无权限页（联系管理员，无申请按钮） |
| 二级账号 | 普通用户 | 预鉴权但**不配** `hideMenu`，始终显示 | 声明，无权限显示无权限页（可申请） |
| 云账单管理 | 管理员 | `AUTH_FIND_ACCOUNT_BILL`，无权限隐藏 | 不声明（历来如此：页面能打开，接口由后端 `AuthorizeWithPerm` 拒绝，由接口兜底弹申请） |
| 两个新建页 | — | — | 不拦截 |

- `hideMenu` 只管侧栏显隐，`meta.auth.view` 只管页面拦截，两者互不替代，按页面面向谁分别决定配不配。
- 二级账号是普通用户的自助入口：菜单必须始终可见，无权限的人点进去在无权限页自行申请。**不要**给它配 `hideMenu`，否则零权限用户连申请入口都找不到。
- 一级账号的权限需由管理员主动授权，`AUTH_FIND_ROOT_ACCOUNT` 在 `views/status/permission.vue` 的 `NO_SELF_APPLY_SIGNS` 里：只提示联系管理员授权，不显示「去申请」按钮。
- 云账单的前端权限点是 `account_bill_find`，后端映射到 IAM 动作 `account_bill_manage`（「云账单-云账单管理」）。
- `main_account_find` 是不带实例的校验，后端走 `AuthorizeAny`，含义是「至少对一个二级账号有查看权限」。二级账号列表本身由后端按账号实例过滤（`cmd/account-server/.../main-account/list.go`），零权限时返回空数组而不报错，所以前端拦截只是把「空表格」换成「403 引导申请」，不改变数据范围。
- 一级账号列表后端是整体校验 `root_account` 查看权限，无权限直接报错。

## 旧地址兼容

`billLegacy` 承接旧 `/bill/*` 地址（query 原样保留）：

- `/bill/root-account`、`/bill/main-account` 及各自 `/create`、`/bill/bill-manage`（含 `summary`、`summary/manage`、`summary/operation-record`、`detail`、`adjust`）→ 按 name 转到 `/resource/bill/*` 对应页。
- `/bill/account-manage`（旧「云账号管理」书签）→ 一级账号视图权限已缓存且有权限去一级账号，否则去二级账号，保留 query。没有二级权限的人落到二级账号后，由上表的拦截显示无权限页申请。
- `/bill/account-manage/first-account`、`/second-account` → 按 name 重定向到两个新建页，vue-router 默认保留 query。

分流读 `preload` 预鉴权的视图权限缓存（`authStore.isViewPermissionCached` + `hasViewPermission`），写在 `beforeEnter`。这条记录要带 `children: []`，因为 vue-router 类型要求 `component` / `children` / `redirect` 三选一，而带 `redirect` 的记录不执行 `beforeEnter`。预鉴权接口失败时缓存为空，按无权限去二级账号。

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
