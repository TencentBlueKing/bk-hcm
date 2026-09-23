# 账单

> status: drafted · kind: module
> globs: `src/views/bill/**`

账单与账号账单视图。

## 职责

一级菜单「资源运营」下「云账单管理」分组里的两个入口页：

- **云账号管理** — `account/account-manage`，账单视角的一级 / 二级账号管理。「录入一级账号」「创建二级账号」是两个非菜单页（`account/create-account/*`），从 `account/account-manage/panel` 用 `routerAction.redirect` 按路由名（`MENU_BILL_ROOT_ACCOUNT_CREATE` / `MENU_BILL_MAIN_ACCOUNT_CREATE`）跳过去，并透传当前 query。本模块内跳转与重定向一律按路由名，不写路径串。
- **云账单管理** — `bill/index.tsx`，账单汇总 / 明细 / 调整。

注意与「业务资源 → 账号 → 云账号管理」（`views/cloud-account-manage`）是**两个不同页面**，只是同名。

## 路由

`views/bill/route-config.ts` 导出 `billAccountManage` / `billManage`，由 `views/index.ts` 汇成纯路由数组 `billViews`；路由名用 `constants/menu-symbol.ts` 的 `MENU_BILL_*`。侧栏菜单（显示名、图标、「云账单管理」分组）不在路由里，在 `common/menu-service.ts` 的 `billMenus` 声明，所以改菜单名或分组去那里，改页面标题、权限、高亮才动路由。通用约定见 `menu-route` 模块。

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
