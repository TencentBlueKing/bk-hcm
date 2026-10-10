# 菜单与路由

> status: drafted · kind: module
> globs: `src/router/**`, `src/constants/menu-symbol.ts`, `src/common/menu-service.ts`, `src/components/layout/menu*.vue`, `src/components/layout/header.vue`

一级菜单与侧栏只在 `common/menu-service.ts` 声明；页面所属一级菜单按 `route.matched[0].name` 判定，顶栏高亮与侧栏都由它决定。

## 关键文件与职责

- **菜单符号** — `constants/menu-symbol.ts`，全部是 `Symbol('<描述>')`（描述串即一期之前的字符串值，与 PR #1897 一致）。路由 `name`、`meta.activeKey`、`meta.owner`、`menu.relative` 都直接用这些 Symbol。未迁移的老路由仍有字符串 name（如 `'resourceDetail'`、`'businessRecyclebin'`、`'scheme-recommendation'`），二者并存。
- **一级菜单与侧栏** — `common/menu-service.ts` 的 `menus: IMenu[]`。顶层每项的 `id` 必须等于 `router/index.ts` 里对应顶层父路由的 `name`；数组顺序即顶栏从左到右的顺序；声明了 `menu` 的一级菜单才显示侧栏。
  - `getMenus()`：顶栏用，按 `visibility` 过滤（资源选型受 `ENABLE_CLOUD_SELECTION`；资源运营始终显示，`ENABLE_ACCOUNT_BILL` 只控其下「云账号管理」三项）。
  - `getCurrentTopMenu(route)`：侧栏与顶栏激活用。**一级菜单本身不按 `visibility` 过滤**（开关只管顶栏入口，直接访问地址时侧栏照常），子项照常过滤。
  - `getMenuRoute(views, id)`：在路由树里递归按 name 找记录，菜单项只按 name 引用路由。
  - `toMenuKey(id)`：Symbol 取 `description` 作 `bk-menu` 的 key；侧栏 `active-key` 用同一函数处理 `route.meta.activeKey`。
- **顶层父路由** — `router/index.ts`：每个一级菜单一个带 name、不挂组件、带 `redirect`（入口页）的父路由，子路由可写绝对路径，地址不受包装影响。状态路由 `404` / `error` 与 `/403/:id` 不在任何父路由下，因此不属于任何一级菜单（顶栏无激活、无侧栏）。未知地址兜底进 `404`（不再跳回首页）。各父路由的 `children` 来自 `views/index.ts`，导出时已经 `injectStatusComponents` 注入状态页 named view；新增视图集合也必须经它包一层。全局 `beforeEach` 只做视图级权限（业务拦截器 → `available` → `meta.auth.view`），不改地址。
- **一级视图集合** — `views/index.ts` 导出 `businessViews` / `serviceViews` / `resourceViews` / `schemeViews`，路由注册与菜单服务共用同一份。`resourceViews`（资源运营 `/resource` 根）= 资源纳管 + 原资源接入页面 + 账单；账单记录写绝对路径 `/bill/*`，地址不变但归属资源运营。

## 侧栏与顶栏渲染

- `components/layout/menu.vue`：取 `getCurrentTopMenu(route).menu`，按 `group` 聚合（首次出现顺序），分组用 `bk-menu-group`（**不可折叠**，不做 #1897 的折叠交互与收藏）。菜单项权限过滤读 `constants/view-auth.ts` 中 `hideMenu` 项的视图权限缓存（菜单 id = `viewId`，见 auth 模块）；整组为空时分组隐藏。业务资源下顶部显示 `BusinessSelector`（侧栏展开时），菜单链接带 `?bizs=`。
- `components/layout/menu-item.vue`：一律按 `{ name }` 跳转（一期为相对路径路由维护的 `getRouteLinkParams` 白名单已随旧壳删除）。显示名取 `IMenu.i18n` 原文；需要翻译的项（单据管理、权限策略库）在 menu-service 里已用 `t()` 求值，渲染器不再套 `t()`。
- `components/layout/header.vue`：遍历 `getMenus()`（再过滤 `router.hasRoute`），`RouterLink` 跳 `{ name: 一级菜单 id }`，由父路由 `redirect` 落到入口页；业务资源额外带 `?bizs=`。激活态照搬 PR 用 `.router-link-active`：链接指向一级父路由，当前路由 matched 含它即点亮（与 `getCurrentTopMenu` 等价，`query` 不参与）。只做跳转的一级入口（资源接入 `MENU_RESOURCE_ACCESS`）不注册路由，菜单项声明 `route`（目标页面）：`header.vue` 对带 `route` 的一级菜单链到 `route.name`、`hasRoute` 按它过滤，并传空的 `active-class` / `exact-active-class`，否则会和目标所属一级菜单同时点亮。点击资源运营、资源接入之外的一级菜单时清空资源账号选择。

## 面包屑

`hooks/use-breadcrumb.ts` 读路由 `meta.layout.breadcrumb`：`show !== false` 即显示（默认显示，对齐 #1897），`back` 控制返回箭头，标题取 `menu.i18n || title`。`isShowBreadcrumb` 已全量改写为 `layout.breadcrumb` 并删除。

**浅合并陷阱**：路由 `meta` 沿 matched 链浅合并，子路由若带默认 `layout` / `activeKey`，会盖掉父记录上的配置。因此 `router/meta.ts` 的 `Meta` **不为未传入的字段产出默认值**，默认语义交给消费方；需要隐藏面包屑的页面在自身或父记录上显式写 `layout: { breadcrumb: { show: false } }`。从 PR 移植路由时，若依赖「默认显示」去覆盖父级的隐藏，必须显式写 `show: true`。`Meta` 仍保留 HEAD 的 `menu.i18n = title` 写法，只写 `menu.i18n` 不写 `title` 的路由，标题会被覆盖为空。

## 返回与历史栈

`router/utils/history-storage.ts` 把路由 name 写进 sessionStorage：Symbol 序列化为 description，读回时查表还原为 Symbol（`JSON.stringify` 会直接丢掉 Symbol）。`router/hooks/use-back.ts` 在历史记录缺 name 时回退到 `meta.menu.relative`。

## 模块路由迁移配方（去中心化模式）

把一个模块从 `router/module/*.ts` 迁到 `views/<模块>/route-config.ts` 时：

1. 模块自己定义路由：`name` 用 `menu-symbol.ts` 的 Symbol，`path` 用相对路径，`meta` 用 `...new Meta({...})`。
2. 在 `views/index.ts` 把路由并入所属一级视图集合（纯路由数组，不放分组包装节点）。
3. 在 menu-service 对应一级菜单的 `menu` 里加 `IMenu`（`group` 是分组标题）。
4. 一级视图要有承载相对路径的顶层父路由；地址变了就补旧地址兼容。固定目标用 `redirect: { name }`（默认保留 query）；目标取决于权限时用 `beforeEnter` 返回 `{ name, query: to.query, replace: true }`，不能用 `redirect` 函数，因为它在全局守卫加载权限之前求值。实例见 `bill` 模块「旧地址兼容」。

`children` 不是禁区：负载均衡（`entry-biz.vue`）、云账单管理（`bill/index.tsx`）有真实布局层（页签栏、`provide` 上下文、内含 `<RouterView>`），保留嵌套。禁止的是**拿 children 当菜单分组**。

## 变体差异

菜单项集合随构建变体不同。**判断「有没有某个菜单」一律以当前分支实际注册的数据为准**：上游需求文档提到、而本分支没有注册的菜单项，不代表需求写错，也不要照文字凭空新增；反之本分支有而文档没提的，也不要顺手删。核对时要连 menu-service 的 `visibility` 开关一起看。

## 过渡形态（菜单整合二期进行中）

二期首批每单上线即用新地址。资源纳管单起 `MENU_RESOURCE` 已是资源运营根；侧栏按最终菜单树结构只放已有最终页面的项（除资源纳管外都是无云账号树的独立页面，各由自己的单加入，不用原页面临时顶替；回收管理、操作记录已上线，账号管理待加入）。尚未迁出的原资源接入页面（账号列表 / 页签、资源详情）只承接旧地址，仍在 `router/module/resource.ts`，作为根的子记录挂在外壳记录下：外壳 `path: '/resource'` 与根同路径（vue-router 优先匹配子记录，所以外壳带 `redirect` 到资源纳管），`meta.legacyResourceAccess` 让 `main-content.vue` 为这些页面渲染原云账号树两栏布局。原资源列表 `/resource/resource` 由外壳下的空路径子路由转到资源纳管（按 name 跳 `MENU_RESOURCE_RESOURCE_MANAGEMENT` 不会触发转址，站内已全部改走新 name）。过渡代码标 `TODO(菜单整合二期·组装)`，随各模块迁出逐步删除；账单迁到 `/resource/bill/*` 由资源运营重组单完成。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
