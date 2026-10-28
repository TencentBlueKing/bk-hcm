# 门户框架/错误页/通知

> status: drafted · kind: module
> globs: `src/app.vue`, `src/components/layout/**`, `src/views/status/**`, `src/views/error-pages/**`, `src/views/notice/**`, `src/style/app-shell.scss`

应用外壳：`app.vue` 直接组装 `bk-navigation` 与顶栏 / 侧栏 / 主区 / 页脚，承担全局副作用；404 / error 状态页；全局通知。

## 职责与关键文件

- **`app.vue`**（无额外 `layout/index.vue`，对齐 PR #1897）：
  - 布局：`bk-navigation`（`top-bottom`，展开宽度用默认 260px）的 `#header` / `#menu` / `#default` / `#footer` 分别放 `HcmHeader` / `HcmMenu` / `HcmMainContent` / `HcmFooter`；`need-menu` = 当前一级菜单声明了 `menu`（见 menu-route）；页脚只在主机列表页（`MENU_BUSINESS_HOST_MANAGEMENT`）渲染。
  - 全局副作用（原旧壳 `views/home/index.tsx` 的职责，旧壳已删除）：挂 `window.hcmPermissionDialog`、`useChromeNotice()`；预取地域（腾讯云 / 华为云）、业务映射、云区域。`accountStore.bizs` 不在这里维护：业务页由业务拦截器写入，资源运营与无归属页面由 `router/index.ts` 的 `afterEach` 清零（资源选型三期已并入工作台的 `MENU_SERVICE`，不再单独判断）。页面无权限不再整页替换，由主区的状态页承接（见 auth 模块「视图级」）。
  - 侧栏收起状态 `navigationState.collapse`（照 #1897）：`bk-navigation` 的 `toggle` 回传的是「当前是否展开」（含悬停临时展开），取反存入，只用于 `default-open`（侧栏不接收该状态，业务选择器收起时照常渲染，与 #1897 一致）。`need-menu` 看当前一级菜单是否声明 `menu`，状态页取不到一级菜单即不显示侧栏。
  - 跑马灯：`<Notice>` 排在导航之前，有跑马灯时导航加 `has-notice` 减去 40px。#1897 把 `<Notice>` 放在导航之后且没有减高度，会多出 40px 滚动条，不照搬。
  - 页脚：只在主机列表页渲染。`app-shell.scss` 把 `container-content` 设为纵向 flex、`main-content` 占满剩余高度，页脚自动排在下面，不需要 #1897 的 `has-footer` 减高度；页面也不要为页脚单独调整内边距。
  - 切换业务由侧栏整页刷新（见 menu-route），主区不再有 `isRouterAlive` / `reload()`。
- **`components/layout/`**：`header.vue`（一级菜单、语言、版本日志、用户）、`menu.vue` / `menu-item.vue` / `menu-favorites.vue`（侧栏、收藏区）、`main-content.vue`（面包屑 + 主视图）、`breadcrumb.vue`、`footer.vue`。业务选择器在 `components/business-selector/global.vue`。
- **`main-content.vue`**：`.main-scroll` 与 `.main-layout` 通过纵向 flex 分配全局面包屑和主视图高度。非 `fit` 页面保留内部 `min-width: 980px`（与 `html/body` 兜底同值；主区在侧栏 260px 之后，整页下限约 1240px，业务视角工具条尚未适配，见「注意事项」的遗留专项）；状态页及显式配置 `meta.layout.contentWidth: 'fit'` 的页面免最小宽度。`fit` 页面关闭主区横滚，内部表格保留列宽、固定列与横滚。当前只在 `/resource/account` 列表及 `/resource/manage/:resourceType` 列表启用；详情、登记和购买路由保持默认策略，不能按路由前缀泛化。保留 `#breadcrumbLeft` 与 `#breadcrumbExtra` 挂载点。
- **宽度与横滚边界**：`style/common.scss` 的 `html/body` 保留 `min-width: 980px`，`body overflow: auto hidden`——页面级最小宽度是所有页面的兜底，防止窄视口下顶栏一级菜单被压缩重叠（此前去掉后出现过顶栏文字挤压）；980px 已足够容纳 logo、一级菜单与右侧工具区，不要再调回 1366px。`html/body` 与 `.main-layout` 都取 980px；因为主区在侧栏（260px）之后，非 `fit` 页面（业务管理下的主机、资源等各资源页签）的整页下限实际约 1240px，而 `fit` 页面才是真正的 980px。业务视角工具条的窄宽适配属于下一条遗留专项。`app.vue` 覆盖导航容器 `max-width` 为 `none`，让导航跟随文档宽度而非视口，侧栏不收缩。`fit` 列表的工具栏允许换行：不换行时搜索区仍贴右（保留 `margin-left: auto`，与拆分前观感一致），一旦换行则由公共 hook `hooks/use-wrap-detect.ts`（`useWrapDetect`）按子项 `offsetTop` 判断并给容器加 `is-wrapped`，整行统一左对齐；`resource-list.vue` 里传入 `{ selector: '.toolbar, .top-bar', scope: '.resource-main' }`。该 hook 与工具栏无关：任何靠 `flex-wrap` 换行的区域（操作条、标签组等）传入对应选择器即可，需要自定义换行后的行为时用 `onChange(wrapped, el)` 回调，`wrapClass: ''` 可关闭默认类名；工具栏内搜索区外层结构不统一（`.search` / `.flex-row` / `.search-selector-container` / 直接是 `bk-search-select`），统一由工具栏控制 180–500px（`flex: 1 1 180px`，flex-basis 必须与 `min-width` 同为 180px——换行按 basis 判定，basis 写大会在剩余不足时就提前换行），内层 `bk-search-select` / `resource-search-select` 铺满外层（覆盖 `.w500`、`.resource-search-select` 的固定 500px）——否则云硬盘这类多包一层 `.flex-row` 的列表搜索框会按内容宽度收缩，宽度和贴右位置与其它列表不一致；按钮组不压缩，空间不足时搜索框单独成行；表格以 flex 占据实际剩余高度，不再用固定 48px/52px 扣高。资源类型仍由 `bk-tab` 承担窄宽自适应滚动，不要再加类型下拉或隐藏页签条。页面级搜索（镜像 / 路由表 / 网络接口）是 loading 容器的直接子项，不能给 `.bk-search-select` 设横向 `flex-basis`，否则在纵向 flex 下会变成高度把搜索区撑满。局部适配样式不改变共享组件在业务视角下的布局。
- **`style/app-shell.scss`**：原旧壳 `index.scss` 中仍被页面依赖的全局样式（导航内容区容器、原资源接入页面的云账号树两栏容器 `.resource-manage-container`：资源运营有侧栏，树用 flex 定宽 240px，不能再 `position: fixed` 贴视口左边）。侧栏外观对齐 PR #1897，用 bkui 默认暗色（侧栏 `#182132`，展开分组 `#151d2c`），不再覆盖侧栏、菜单、选择器底色与列表上内边距；业务选择器外边距写在 `app.vue`（`0 8px 10px`）。侧栏展开宽度与菜单宽度都用 bkui 默认 260px（`bk-navigation` 不传 `hover-width`、`bk-menu` 不设宽度，两者须一致，否则菜单右侧被裁）。业务选择器依赖 `@blueking/app-select@0.0.1-beta.19`：`beta.6` 不支持 `clearable`，会出现清除按钮。不要把旧壳的失效选择器或旧页脚 52px 高度规则带回来——组件库总会渲染页脚容器，那条规则会让所有页面底部多出 52px 空白。
- **视图集合**：`views/index.ts` 的 `serviceViews` 直接在文件内联四组路由 —— `ticketRoutes`、`permissionPolicyRoutes`、`scheme`、`accountManageRsc`（工作台只是一个视图级聚合，没有自己的 `route-config.ts`；`router/module/` 三期已清空）。账号管理是三期从 `resourceViews` 迁入的，`resourceViews` 仍含资源纳管、回收管理、操作记录与账单。
- **状态页**：`views/status/404.vue`（未知地址兜底，返回首页用 `routerAction`）、`views/status/error.vue`、`permission.vue`、`business.vue`。404 / error 既是独立路由，也和两个无权限页一起作为 named view 注入各页面路由；`main-content.vue` 的出口按 `route.meta.view` 渲染（`default` 之外即状态页，居中、免最小宽度，原资源接入页面此时不渲染云账号树）。旧 `/403/:id` 页（`views/error-pages/`）已删除，旧书签落 404。

## 注意事项

- **遗留：所有资源页的搜索组件统一与排版优化（下一期专项）**。`fit` 列表（`/resource/manage/:resourceType`、`/resource/account`）的工具栏已统一：换行检测（`is-wrapped`）、搜索区 180–500px、内层搜索输入铺满外层。但业务视角下仍有多处资源页的工具栏区域没有做窄宽适配，窗口变窄时按钮组与搜索框会被撑破，`.main-layout` 现取 980px（整页约 1240px），`html/body` 的 980px 兜底在这些页面触发不到。该专项要把各资源页（含业务视角）的工具栏结构收敛成同一套后再放宽主区宽度下限，顺带清理各页自带的搜索区 hack 样式（如 `.w500`、`.resource-search-select` 的固定宽度、外层 `.flex-row` 包一层等）。
- **页面标题统一走全局面包屑**，页面不要自带固定定位的标题栏（旧 `DetailHeader` 已随账号管理迁移删除）：静态标题写路由 `menu.i18n`（`Meta` 已无 `title` 字段），动态标题 `useBreadcrumb().setTitle()`；右侧操作按钮 `<Teleport defer to="#breadcrumbExtra">`；返回目标靠路由 `menu.relative`（业务页返回时缺的业务 ID 由拦截器按 localStorage 补上，见 menu-route）。内容区容器用 `page-container`（列表 / 普通页）或 `detail-content-wrap`（详情页），新建 / 申请页用 `components/layout/sticky-bottom-container` 让底部按钮吸底。
- `main-content.vue` 只负责全局面包屑与按 `meta.view` 渲染 named view（状态页），不再有原资源接入的云账号树两栏；云账号树与右栏面包屑只存在于资源纳管入口 `views/resource-manage/resource-entry.vue`。

- 语言切换照 #1897：`js-cookie` 先删后写 `blueking_language`（`expires: 366`，`domain: BK_DOMAIN`），有 `BK_COMPONENT_API_URL` 时调用户语言接口，最后刷新；当前语言默认 `zh-cn`。#1897 写的 `BK_HCM_DOMAIN` 已在 #1647 改名为 `BK_DOMAIN`（`index.html` 只有后者），照抄会写成当前域名的 cookie，盖不住蓝鲸统一域名下的语言设置。
- 退出登录用 `utils/login-helper.ts` 的 `gotoLoginPage(undefined, true)`（带 `is_from_logout=1`，由登录服务清理登录态）。不要手拼 `c_url=${location.href}`：hash 路由下未转义的 `#` / `&` 会截断回跳地址。
- `app.vue` 无外层容器，`bk-navigation.hcm-app` 作根（`height: 100%`）。通知栏（`ENABLE_NOTICE`）在文档流中占 40px，必须排在导航之前，有跑马灯时导航加 `has-notice` 高度减 40px；固定定位元素的 top 偏移用 `hooks/use-calc-top-with-notice.ts`（原 `views/home/hooks/useCalcTopWithNotice.ts`）。

- **栅格（`components/layout/grid-container`）的窄宽约定**：列宽一律 `minmax(0, 1fr)`（`fixed` 模式除外），`.grid-item` 与 `.item-content` 均 `min-width: 0`，内容区直接子元素 `min-width: 0` + `max-width: 100%`，`.form-element` 内的 `bk-input / bk-select / bk-date-picker / bk-tag-input` 同样补 `min-width: 0`。原因：grid/flex 子项默认 `min-width: auto` 等于内容 min-content，长 ID、无空格长串、固定宽控件都会把轨道顶宽，最终整块撑破容器（页面横滚、右侧列被裁、表头错位）。注意组件只保证「不撑破」，**长串换行要页面侧兜底**：展示长文本的 `.item-content` 需加 `overflow-wrap: anywhere`（`anywhere` 还会参与 min-content 计算，与 `minmax(0,1fr)` 配对才彻底生效，只写 `break-word` 不够）；`content-min-width` 是内容列的硬下限，窄容器场景传 `0`（工单详情、云账号搜索已统一），不要传 200/300 这类值；`vertical`（搜索块）与 `bordered` 走同一套规则，无需额外处理。新增栅格页面直接沿用，不要再写裸 `1fr`。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
