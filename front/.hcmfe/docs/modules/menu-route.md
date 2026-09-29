# 菜单与路由

> status: drafted · kind: module
> globs: `src/router/**`, `src/constants/menu-symbol.ts`, `src/common/menu-service.ts`, `src/views/home/hooks/to-legacy-menus.ts`

一级菜单在 header-config；侧栏按路径段 switch 取数据，资源运营一支已由 menu-service 供数，其余仍靠路由 children + groupTitle 表达分组。

## 当前基线的真实形态（先读这条，别被 rule 误导）

`.cursor/rules/fe-menu-route-architecture.mdc` 与 `fe-deprecated.mdc` 描述的是完整的 `common/menu-service.ts` + `components/layout/menu.vue` 那套「菜单与路由解耦」形态。**那是一次整体重构的产物，未合入本基线**：本分支的 `common/menu-service.ts` 只是先行雏形（只覆盖资源运营），没有 `components/layout/menu.vue`，`views/home/index.tsx` 仍是活的布局入口（`App.vue` 直接 `import Home from '@/views/home'`）。按那两份 rule 的「已废弃」结论跳过 `views/home/index.tsx`，会找不到菜单到底在哪渲染。

## 关键文件与职责

- **一级菜单定义** — `src/router/header-config.ts` 的 `headRouteConfig`，`{ id, name, path }` 数组，数组顺序即从左到右的展示顺序；显示走 `t(name)`，点击按 path 跳转。
- **一级菜单显隐** — 在 `views/home/index.tsx` 的 header 渲染处按 `window.PROJECT_CONFIG` 的 env 开关就地过滤（`ENABLE_CLOUD_SELECTION` 控 `scheme`、`ENABLE_ACCOUNT_BILL` 控 `bill`）。开关只作用于顶栏入口，直接访问对应 URL 时侧栏照常显示。
- **侧栏数据来源** — `views/home/hooks/useChangeHeaderTab.ts` 按 `route.path` 的第一段做 switch，给 `menus.value` 赋值：
  - `bill` → `toLegacyMenus(getBillMenus(), billViews)`：菜单来自 `common/menu-service.ts`，路由来自 `views/index.ts` 的 `billViews`，两者各有来源。
  - `business` → `businessViews`（就是 `router/module/business.ts`）、`resource` → `router/module/resource.ts`、`service` → `service.ts`、`scheme` → `scheme.ts`：直接用路由数组当菜单。
- **侧栏渲染** — `views/home/index.tsx` 的 `menu` 插槽，遍历 `menus.value` 输出 `Menu.Group` / `Menu.Item`。它只认路由形状的数据。

## 侧栏渲染规则

判定只对 `menus.value` 的**顶层元素**生效：`Array.isArray(menuItem.children) && !menuItem.meta?.hasPageRoute` 为真就渲染成 `Menu.Group`（标题取 `meta.groupTitle`，key 取 `path`），否则渲染成单个 `Menu.Item`。分组下的子项无论自己有没有 children，都只渲染成一个菜单项。

菜单项读取的 meta：`title` 展示名、`icon` 图标类名、`activeKey` 高亮、`notMenu` 排除出菜单、`checkAuth` 按权限 action 过滤（整组子项都无权限时该分组自动隐藏）。`icon` 与 `notMenu` 全仓唯一消费方就是这里。

面包屑不归侧栏管，由 `hooks/use-breadcrumb.ts` 读路由：`meta.layout.breadcrumb.show ?? meta.isShowBreadcrumb` 决定显示，标题取 `meta.title`，返回箭头来自 `meta.menu.relative`（或 `_f` 历史栈）。字段名已对齐整体重构的 `layout.breadcrumb`（单数），新路由用它，不再写已废弃的 `isShowBreadcrumb`；但本分支 `Meta` **默认不显示**，与整体重构默认显示不同，需要面包屑的页面必须显式写 `show: true`。

`meta.hasPageRoute` 是给顶层元素用的反例标记（「我的 children 是页签子路由，别当分组」），目前全仓已无设置点，只剩判定式。

## 分组的两种写法

**写法一：路由包装节点（旧，business / service 等仍在用）**。菜单数组里放一个不可点击的节点：

```ts
{ path: '/business', children: [ ...真正的菜单路由 ], meta: { groupTitle: '资源' } }
```

它同时被注册进 vue-router，菜单结构因此混在路由树里，改分组就是改路由树。`router/module/business.ts` 末尾用这个写法排出了「资源 / 其他 / 账号 / 回收站」四个分组。**新代码不要再用。**

**写法二：menu-service（目标形态，资源运营已用）**。菜单在 `common/menu-service.ts` 里声明为 `IMenu`，分组是菜单项上的扁平字段 `group`，菜单只按 name 引用路由：

```ts
{ id: MENU_BILL_MANAGE, i18n: '云账单管理', icon: 'bkhcm-icon-bill-manage', group: '云账单管理',
  route: getMenuRoute(billViews, MENU_BILL_MANAGE) }
```

显示名（`i18n`）、图标、分组、显隐（`visibility`）归菜单；页面标题、权限（`checkAuth`）、高亮（`activeKey`）留在路由上。菜单列哪些项由 menu-service 决定，路由上不再需要 `notMenu` / `icon`。`IMenu`、`getMenuRoute` 与 visibility 过滤和那次未合入的整体重构保持一致，将来可原样并入完整版。

**两种写法之间的桥是 `views/home/hooks/to-legacy-menus.ts`**：把 `IMenu[]` 按 `group` 聚合（顺序按首次出现），转成旧渲染器认得的写法一形状，结果只活在内存里、不注册进路由。转换时按 name 取完整路由记录、复制一份并用菜单的 `i18n` / `icon` 覆盖 `meta.title` / `meta.icon`，同时**去掉 `children`**，保证带页签子路由的菜单项永远不会被误判成分组。侧栏改为直接按 `IMenu` 渲染后删除这个文件。

## 模块路由迁移配方（去中心化模式）

把一个模块从 `router/module/*.ts` 迁到 `views/<模块>/route-config.ts` 时，以下几步缺一不可。已按此迁完的有：任务管理、操作记录、云账号管理、负载均衡、单据、账单。

1. **模块自己定义路由**：`name` 用 `constants/menu-symbol.ts` 的常量，`path` 用**相对路径**，`meta` 用 `...new Meta({...})` 展开。
2. **汇入一级视图集合**：在 `views/index.ts` 导出纯路由数组 `xxxViews`，不要在里面放分组包装节点。
3. **菜单在 menu-service 声明**：按写法二加 `IMenu` 项，`useChangeHeaderTab` 里用 `toLegacyMenus(getXxxMenus(), xxxViews)` 喂给侧栏。
4. **把新符号加进 `views/home/index.tsx` 的 `getRouteLinkParams` 白名单**。该函数白名单内走 `{ name }` 跳转、其余走 `{ path: config.path }`；**路由改成相对 path 后不进白名单，侧栏 RouterLink 会拿相对路径去跳，直接跳错**。这一步最容易漏。
5. **一级视图要有承载相对路径的父路由**：`{ name, path: '/xxx', children: xxxViews }`，URL 不变，旧地址继续可用。
6. **地址变了就补旧地址兼容**，并核对 `header-config.ts` 里的顶栏入口地址是否也是旧地址。固定目标用 `redirect: { name }`（默认保留 query）；目标取决于权限时用 `beforeEnter` 返回 `{ name, query: to.query, replace: true }`，不能用 `redirect` 函数，因为它在全局守卫加载权限之前求值。实例见 `bill` 模块「旧地址兼容」。整体重构分支搬地址时漏了这一步，从那边移植时要补上。

**`children` 不是禁区**。任务管理、操作记录是平的，因为列表与详情是彼此独立的整页；负载均衡（`entry-biz.vue`）和云账单管理（`bill/index.tsx`）有真实布局层——提供页签栏与 `provide` 上下文、内含 `<RouterView>`——所以保留嵌套。禁止的是**拿 children 当菜单分组**。

## 变体差异

菜单项集合随构建变体不同。**判断「有没有某个菜单」一律以当前分支实际注册的数据为准**：上游需求文档里提到、而本分支没有注册的菜单项，不代表需求写错，也不要照文字凭空新增；反之本分支有而文档没提的，也不要顺手删。核对时要连 header 的 env 开关一起看，否则会把「开关关掉」误判成「没有这个菜单」。

## 架构演进：menu-service 化

目标形态是菜单与路由彻底解耦：所有一级视图的菜单（含一级菜单本身）都在 menu-service 声明，侧栏与顶栏直接按 `IMenu` 渲染，`visibility` 是一等字段，菜单搬家不动路由。

**已完成的先行部分**：资源运营一支的侧栏已由 menu-service 供数（写法二 + 适配桥）。**剩余工作**：其余一级视图的菜单迁入 menu-service；一级菜单名从 `header-config.ts` 迁入；侧栏 / 顶栏改为按 `IMenu` 渲染；删除 `to-legacy-menus.ts` 与路由上的 `groupTitle` 包装节点、`hasPageRoute` 判定。

当前采用分期浅层改动，**命中下面任一条信号就该在那一期做剩余重构，而不是继续叠浅层改动**：

1. 同一轮改动需要在 ≥3 处各写一份菜单事实（`header-config.ts`、`useChangeHeaderTab` 的 switch、路由上的 `groupTitle` / `children`、menu-service）。
2. 需要让「菜单显隐」在顶栏与侧栏间联动（按权限、变体或灰度控制某一项），而顶栏现状只能在 JSX 里加 env 判断。
3. 需要把已有菜单项从一个一级视图搬到另一个。即便该项已由 menu-service 供数，`useChangeHeaderTab` 仍按 URL 第一段决定侧栏归属与顶栏高亮，搬家就等于改 URL，与「旧地址必须继续可用」冲突。

真要做时，把那次未合入的整体重构当**供体**按期裁剪搬运（它已有完整 menu-service、一级视图集合、新侧栏），不要重写；也不要再一次性全量合入——上次就是因为改动过大没能合进来。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
