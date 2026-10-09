# 门户框架/错误页/通知

> status: drafted · kind: module
> globs: `src/app.vue`, `src/components/layout/**`, `src/views/status/**`, `src/views/error-pages/**`, `src/views/notice/**`, `src/style/app-shell.scss`

应用外壳：`app.vue` 直接组装 `bk-navigation` 与顶栏 / 侧栏 / 主区 / 页脚，承担全局副作用；404 / error 状态页；全局通知。

## 职责与关键文件

- **`app.vue`**（无额外 `layout/index.vue`，对齐 PR #1897）：
  - 布局：`bk-navigation`（`top-bottom`、`hover-width=240`）的 `#header` / `#menu` / `#default` / `#footer` 分别放 `HcmHeader` / `HcmMenu` / `HcmMainContent` / `HcmFooter`；`need-menu` = 当前一级菜单声明了 `menu`（见 menu-route）；页脚只在主机列表页（`MENU_BUSINESS_HOST_MANAGEMENT`）渲染。
  - 全局副作用（原旧壳 `views/home/index.tsx` 的职责，旧壳已删除）：挂 `window.hcmPermissionDialog`、`useChromeNotice()`；预取地域（腾讯云 / 华为云）、业务映射、云区域；`accountStore.bizs` 变化时重算 `pageAuthData` 的业务 ID 并 `getAuthVerifyData`；按所属一级菜单初始化全局业务 ID（资源接入、资源选型置 0，无归属页面置空）；`usePagePermissionStore` 无页面权限时整页显示 `NoPermission`、渲染旧的 `GlobalPermissionDialog`（两者在视图权限体系改造中删除）。
  - `isRouterAlive` + `reload()`：业务选择器切换业务时由侧栏调用，主区卸载再挂载 `RouterView`。
- **`components/layout/`**：`header.vue`（一级菜单、语言、版本日志、用户）、`menu.vue` / `menu-item.vue`（侧栏）、`main-content.vue`（面包屑 + 主视图）、`breadcrumb.vue`、`footer.vue`、`business-selector.tsx`（原旧壳业务选择器）。
- **`main-content.vue`**：主视图最小宽度 1089px（状态页除外），本身无内边距（由页面容器 `page-container` / `detail-content-wrap` 提供）；保留 `#breadcrumbLeft`（云账号管理页 Teleport 目标）与 `#breadcrumbExtra` 挂载点。
- **最小宽度**：`style/common.scss` 中页面最小宽度 1366px；`app.vue` 覆盖导航容器组件内联的 `max-width: calc(100vw - 60px)`，使其跟随文档宽度。
- **`style/app-shell.scss`**：原旧壳 `index.scss` 中仍被页面依赖的全局样式（如 `.menu-warp .menu-icon`、`.home-menu`、资源接入左侧账号树容器）。不要把旧壳的失效选择器或旧页脚 52px 高度规则带回来——组件库总会渲染页脚容器，那条规则会让所有页面底部多出 52px 空白。
- **状态页**：`views/status/404.vue`（未知地址兜底，返回首页用 `routerAction`）、`views/status/error.vue`；不显示侧栏与面包屑。旧的 `views/error-pages/403.tsx` 仍是 `/403/:id` 的页面（视图权限体系改造前）。

## 注意事项

- **页面标题统一走全局面包屑**，页面不要自带固定定位的标题栏（旧 `views/resource/resource-manage/common/header/detail-header.tsx` 已废弃，只剩旧账号详情在用，随账号管理迁移删除）：静态标题写路由 `title`（用 `new Meta` 时写 `title`，构造函数会用它覆盖 `menu.i18n`；普通 meta 对象写 `menu.i18n`），动态标题 `useBreadcrumb().setTitle()`；右侧操作按钮 `<Teleport defer to="#breadcrumbExtra">`；返回目标靠路由 `menu.relative`，面包屑返回时自动带上 `?bizs=`。内容区容器用 `page-container`（列表 / 普通页）或 `detail-content-wrap`（详情页），新建 / 申请页用 `components/layout/sticky-bottom-container` 让底部按钮吸底。
- 资源接入过渡布局下全局面包屑隐藏，详情 / 申请 / 任务详情页的面包屑由 `main-content.vue` 渲染在右侧内容区（`legacyBreadcrumbRoutes`），随资源运营重组删除。

- 语言切换沿用 HEAD 写法：写 `blueking_language` cookie（`domain=BK_DOMAIN`）并调用户语言接口后刷新。PR #1897 用的 `BK_HCM_DOMAIN` 在本分支的 `index.html` 中不存在。
- 退出登录用 `utils/login-helper.ts` 的 `gotoLoginPage(undefined, true)`（带 `is_from_logout=1`，由登录服务清理登录态）。不要手拼 `c_url=${location.href}`：hash 路由下未转义的 `#` / `&` 会截断回跳地址。
- `app.vue` 无外层容器，`bk-navigation.hcm-app` 作根（`height: 100%`）。通知栏（`ENABLE_NOTICE`）在文档流中占 40px，必须排在导航之前，有跑马灯时导航加 `has-notice` 高度减 40px；固定定位元素的 top 偏移用 `hooks/use-calc-top-with-notice.ts`（原 `views/home/hooks/useCalcTopWithNotice.ts`）。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
