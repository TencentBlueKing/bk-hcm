# 菜单与路由

> status: drafted · kind: module
> globs: `src/router/**`, `src/constants/menu-symbol.ts`, `src/common/menu-service.ts`, `src/components/layout/menu*.vue`, `src/components/layout/header.vue`

一级菜单与侧栏只在 `common/menu-service.ts` 声明；页面所属一级菜单按 `route.matched[0].name` 判定，顶栏高亮与侧栏都由它决定。

## 关键文件与职责

- **菜单符号** — `constants/menu-symbol.ts`，全部是 `Symbol('<描述>')`（描述串即一期之前的字符串值，与 PR #1897 一致）。路由 `name`、`meta.activeKey`、`meta.owner`、`menu.relative` 都直接用这些 Symbol。业务侧已全部是 Symbol（回收站 `MENU_BUSINESS_RECYCLEBIN`、申请页 `MENU_BUSINESS_APPLY_{CVM,VPC,DISK,SUBNET,CLB}`、负载均衡 `MENU_BUSINESS_LOAD_BALANCER_{LB,TG}_VIEW`）；未迁移的服务请求、资源选型仍有字符串 name（如 `'scheme-recommendation'`）。
- **一级菜单与侧栏** — `common/menu-service.ts` 的 `menus: IMenu[]`。顶层每项的 `id` 必须等于 `router/index.ts` 里对应顶层父路由的 `name`；数组顺序即顶栏从左到右的顺序；声明了 `menu` 的一级菜单才显示侧栏。
  - `getMenus()`：顶栏用，按 `visibility` 过滤（资源选型受 `ENABLE_CLOUD_SELECTION`；资源运营始终显示，`ENABLE_ACCOUNT_BILL` 只控其下「云账单管理」三项）。
  - `getCurrentTopMenu(route)`：侧栏与顶栏激活用。**一级菜单本身不按 `visibility` 过滤**（开关只管顶栏入口，直接访问地址时侧栏照常），子项照常过滤。
  - `getMenuRoute(views, id)`：在路由树里递归按 name 找记录，菜单项只按 name 引用路由。
  - `toMenuKey(id)`：Symbol 取 `description` 作 `bk-menu` 的 key；侧栏 `active-key` 用同一函数处理 `route.meta.activeKey`。
- **顶层父路由** — `router/index.ts`：每个一级菜单一个带 name、不挂组件、带 `redirect`（入口页）的父路由。状态路由 `404` / `error` 不在任何父路由下，因此不属于任何一级菜单（顶栏无激活、无侧栏）。未知地址（含旧书签 `/403/:id`）兜底进 `404`。各父路由的 `children` 来自 `views/index.ts`，导出时已经 `injectStatusComponents` 注入状态页 named view；新增视图集合也必须经它包一层。全局 `beforeEach`：业务拦截器（可能转址）→ `available` → `meta.auth.view`。
- **业务地址** — 业务资源根是 `/business/:bizId(\d+)?`，子路由全部相对路径，地址形如 `/business/<业务ID>/host`。`router/business-interceptor.ts` 只作用于业务路由，顺序与 #1897 相同：路径已有业务 ID 时以路径为准；路径没有时才把 `?bizs=` 转成路径参数；仍然没有时依次取 localStorage、第一个有权限业务、第一个业务。业务 ID 定下来后再判状态（见 auth 模块），通过后写 localStorage 与 `accountStore.bizs`。转址由拦截器自己 `next({ replace: true })`，不把缺业务 ID 的中间地址留在历史记录里。
- **业务内跳转** — 跳业务路由一律显式传 `params: { bizId }`（`route.params.bizId`、负载均衡页注入的 `currentGlobalBusinessId`、或 `getBizsId()`），**不要**再带 `?bizs=`；字符串地址写 `/#/business/${bizId}/…`。vue-router 4.1 按名字解析时只沿用**必填**路径参数，`bizId` 是可选参数，不传就会丢；丢了由拦截器按「localStorage → 第一个有权限业务 → 第一个业务」补上并 `replace`，只作兜底。读当前业务用 `route.params.bizId` 或 `useWhereAmI().getBizsId()`（路由 → `accountStore.bizs` → localStorage，组件外也可用）。`accountStore.bizs` 只是镜像，供取不到路由的旧代码读。
- **一级视图集合** — `views/index.ts` 导出 `businessViews` / `serviceViews` / `resourceViews` / `schemeViews`，路由注册与菜单服务共用同一份。`businessViews` 由各模块 `views/<模块>/route-config.ts` 拼接（主机、硬盘、镜像、VPC、子网、弹性 IP、网络接口、路由表、安全组、负载均衡、证书、单据、操作记录、任务、云账号管理、回收站），每条记录 `meta.owner = MENU_BUSINESS`。`resourceViews`（资源运营 `/resource` 根）= 资源纳管 + 账号管理 + 回收管理 + 操作记录 + 账单（`/resource/bill/*`）。资源侧各模块的旧地址数组（`*Legacy`）由 `router/index.ts` 在顶层注册；业务侧旧地址（`ip`、`host|drive/recyclebin/:type`）写在各自模块的 `route-config.ts` 里，随 `businessViews` 注册。

## 侧栏与顶栏渲染

- `components/layout/menu.vue`：分固定区与滚动区。固定区 = 业务选择器（`components/business-selector/global.vue`，仅业务资源；侧栏收起时随侧栏变窄，与 #1897 一致）+「我的收藏」（独立 `bk-menu`，`menu-favorites.vue`）；滚动区 = 常规菜单，按 `group` 聚合（首次出现顺序），分组用可折叠的 `bk-submenu`，图标取组内第一项的 `groupIcon`。图标照设计稿只用在分组标题与不在分组里的独立菜单（`IMenu.icon`），分组内的菜单项不配 `icon`。菜单项权限过滤读 `constants/view-auth.ts` 中 `hideMenu` 项的视图权限缓存（菜单 id = `viewId`，见 auth 模块；**不**照 #1897 按 `hasViewPermission` 过滤全部菜单）；整组为空时分组隐藏。
  - 收藏 / 折叠：`hooks/use-menu-favorite.ts`、`hooks/use-menu-collapse.ts` 是模块级单例，按「用户名 × 一级菜单 description」存 localStorage。失效收藏按「菜单是否存在」清理，无权限的只是不显示；无收藏时收藏区强制收起且不可展开。从收藏区点击时常规菜单的激活项弱化，反之亦然（`lastClickSource`）。
  - 切换业务：`router.resolve({ name: meta.menu.relative ?? route.name, params: { bizId } })` 后整页刷新，详情页回到所属列表。选择器的业务收藏（`hooks/use-business-favorite.ts`）只在当前业务有权限时查询，否则收藏接口会触发全局申请弹窗。
- `components/layout/menu-item.vue`：按 `{ name }` 跳转，业务菜单另带 `params.bizId`（`menu.vue` 经 `provide('isBusinessNav' / 'bizId')` 下发，收藏区同理），悬停显示收藏星标。显示名取 `IMenu.i18n` 原文；需要翻译的项（单据管理、权限策略库）在 menu-service 里已用 `t()` 求值，渲染器不再套 `t()`。
- `components/layout/header.vue`：遍历 `getMenus()`（再过滤 `router.hasRoute`），`RouterLink` 跳 `{ name: 一级菜单 id }`，由父路由 `redirect` 落到入口页（业务资源由拦截器补业务 ID）。激活态照搬 PR 用 `.router-link-active`：链接指向一级父路由，当前路由 matched 含它即点亮（与 `getCurrentTopMenu` 等价，`query` 不参与）。只做跳转的一级入口（资源接入 `MENU_RESOURCE_ACCESS`）不注册路由，菜单项声明 `route`（目标页面）：`header.vue` 对带 `route` 的一级菜单链到 `route.name`、`hasRoute` 按它过滤，并传空的 `active-class` / `exact-active-class`，否则会和目标所属一级菜单同时点亮。顶栏不再清空资源账号选择：离开资源纳管时由 `resource-entry.vue` 卸载清空。

## 面包屑

`hooks/use-breadcrumb.ts` 读路由 `meta.layout.breadcrumb`：`show !== false` 即显示，`back` 控制返回箭头，标题只取 `menu.i18n`。`router/meta.ts` 的 `Meta` 与 #1897 一致：默认 `layout.breadcrumb = { show: true, back: true }`、`activeKey` / `owner` 默认 `Symbol.for('')`，不再有 `title` 字段（标题一律写 `menu.i18n`）；`isShowBreadcrumb`、`notMenu`、`icon` 也已删除（显示名、图标、是否进菜单归 menu-service）。

**浅合并陷阱**：路由 `meta` 沿 matched 链浅合并，用 `Meta` 构造的子路由一定带 `layout` / `activeKey` / `owner` 默认值，会盖掉父记录上的同名配置。因此有子路由时，每条子路由都要自己写全 `layout`、`activeKey`、`owner`（如负载均衡、账号详情），不要指望从父记录继承；不挂页面的中间层记录（如负载均衡的 `resource`）不写 `meta`。

## 返回与历史栈

`router/utils/history-storage.ts` 把路由 name 写进 sessionStorage：Symbol 序列化为 description，读回时查表还原为 Symbol（`JSON.stringify` 会直接丢掉 Symbol）。`router/hooks/use-back.ts` 在历史记录缺 name 时回退到 `meta.menu.relative`。历史记录存了完整 `params`（含 `bizId`）；回退到 `menu.relative` 时不带 `bizId`，由拦截器按 localStorage 里当前业务补上。面包屑不再补 `?bizs=`。

## 模块路由迁移配方（去中心化模式）

把一个模块从 `router/module/*.ts` 迁到 `views/<模块>/route-config.ts` 时：

1. 模块自己定义路由：`name` 用 `menu-symbol.ts` 的 Symbol，`path` 用相对路径，`meta` 用 `...new Meta({...})`。
2. 在 `views/index.ts` 把路由并入所属一级视图集合（纯路由数组，不放分组包装节点）。
3. 在 menu-service 对应一级菜单的 `menu` 里加 `IMenu`（`group` 是分组标题）。
4. 一级视图要有承载相对路径的顶层父路由；地址变了就补旧地址兼容。固定目标用 `redirect: { name }`（默认保留 query）；目标取决于权限时用 `beforeEnter` 返回 `{ name, query: to.query, replace: true }`，不能用 `redirect` 函数，因为它在全局守卫加载权限之前求值。实例见 `bill` 模块「旧地址兼容」。

`children` 不是禁区：负载均衡（`entry-biz.vue`）、云账单管理（`bill/index.tsx`）有真实布局层（页签栏、`provide` 上下文、内含 `<RouterView>`），保留嵌套。禁止的是**拿 children 当菜单分组**。

## 变体差异

菜单项集合随构建变体不同。**判断「有没有某个菜单」一律以当前分支实际注册的数据为准**：上游需求文档提到、而本分支没有注册的菜单项，不代表需求写错，也不要照文字凭空新增；反之本分支有而文档没提的，也不要顺手删。核对时要连 menu-service 的 `visibility` 开关一起看。

## 旧地址兼容（菜单整合二期）

二期首批每单上线即用新地址，资源运营各模块在自己的 `route-config.ts` 导出旧地址数组，由 `router/index.ts` 在 catch-all 之前注册：`resourceManageLegacy`（`/resource/resource?type=` 与旧新建页）、`resourceDetailLegacy`、`recyclebinRscLegacy`、`operationLogRscLegacy`、`accountManageLegacy`、`billLegacy`。原资源接入的路由文件 `router/module/resource.ts` / `resource-inside.ts` / `service-inside.ts` 与业务路由 `router/module/business.ts`、旧 403 路由 `router/module/common.ts` 已删除；侧栏按最终菜单树结构，除资源纳管外都是无云账号树的独立页面。目标形态（菜单与路由解耦、侧栏与顶栏按 `IMenu` 渲染、`visibility` 一等字段）已由上文的 menu-service 落地；`header-config.ts`、`to-legacy-menus.ts`、`useChangeHeaderTab` 与路由上的 `groupTitle` 包装节点已删除，不再按「剩余工作」继续叠浅层改动。首批删除 / 迁移的文件与合入 v1.9.x 时的冲突处置见 `../menu-phase2-merge-v1.9.x.md`。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
