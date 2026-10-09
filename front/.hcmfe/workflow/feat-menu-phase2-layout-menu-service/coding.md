# Coding — feat-menu-phase2-layout-menu-service

需求文档：[`docs/reqs/二期布局菜单服务化.md`](../../../../docs/reqs/二期布局菜单服务化.md)（TAPD [1069995598138702243](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702243)，验收以其 AC-001～AC-013、AC-P01、AC-S01 为准）。设计参照 PR #1897（base `8fea9924df`，head `df8865b539`），HEAD 基线 `01bca385c`。

## 目标

旧壳 `views/home/index.tsx` 停止渲染并删除；`app.vue` 按 PR #1897 直接组装 `bk-navigation` + `HcmHeader` / `HcmMenu` / `HcmMainContent` / `HcmFooter`（不建 `components/layout/index.vue`）。5 个一级菜单与各侧栏只在 `common/menu-service.ts` 声明一次；侧栏按 `route.matched[0].name`（所属一级菜单）取，顶栏高亮用 `RouterLink` 的 `.router-link-active`（链接指向同一顶层父路由，二者等价）。菜单符号全部改为 `Symbol`，面包屑配置全部改为 PR 的 `layout.breadcrumb`。菜单项、地址、各页是否显示面包屑与现网逐项一致；按 PR 有意变化的只有四处：面包屑默认显示、顶栏链接样式、最小宽度、未知地址进 404。

## 跨单约定（与本单相关部分）

首批 7 单同一 PR、按单拆 commit：① 本单 → ② 视图权限体系 → ③ 资源纳管主体 → ④ 回收管理与操作记录 → ⑤ 账号管理 → ⑥ 资源详情页迁移 → ⑦ 资源运营一级视图重组与资源接入跳转（组装）。

- **菜单符号从 ① 起全部是 `Symbol`**：`constants/menu-symbol.ts` 每个导出都是 `Symbol('<HEAD 现字符串值>')`，名称与 PR 一致；本单新增 `MENU_SCHEME = Symbol('menu_scheme')`。②～⑦ 新增符号同样写 `Symbol(...)`；非菜单常量的历史路由名（`'resourceDetail'`、`'accountDetail'`、`'businessRecyclebin'`、`'hostBusinessDetail'`、`'scheme-*'`、`'403'` 等）保持字符串。需要稳定字符串时用 `symbol.description`，不要用 `toString()`/`String()`。
- `MENU_RESOURCE` 在本单是**过渡的「资源接入」顶层父路由名**；⑦ 把它改为「资源运营」根。凡本单为「资源接入」写的过渡代码统一以 `route.matched[0]?.name === MENU_RESOURCE` 判定，下文标【⑦删】，⑦ 必须同时删除，否则资源运营会带上账号树。⑥ 迁移资源详情页时 `/resource/detail/:type` 仍在本单的 `MENU_RESOURCE` 父路由下，直到 ⑦ 重挂。
- 「资源运营」在本单仍是一期结构：顶层父路由 `MENU_BILL`（`/bill`）+ 一期三项侧栏。⑦ 改为 PR 最终结构、删除 `MENU_BILL` 一级菜单。
- 本单提供给 ②～⑦ 的基础设施：`IMenu`、`getMenus()`（顶栏，按 visibility 过滤）、`getCurrentTopMenu(route)`（侧栏与按一级菜单的副作用）、`getMenuRoute(views, id)`（递归按 name 查）、`toMenuKey(id)`；`views/index.ts` 导出扁平的 `businessViews` / `serviceViews` / `billViews`。
- **面包屑约定**：显示 = `layout.breadcrumb.show !== false`，不配即显示；隐藏写 `layout: { breadcrumb: { show: false } }`；`isShowBreadcrumb` 已删除。`Meta` **不为未传入的 `layout` 产出默认值**，所以 `layout` 沿 matched 链继承（父记录写 `show:false`，未写 `layout` 的子记录同样隐藏）。③～⑥ 从 PR 移植的路由若依赖 PR「Meta 默认 `show: true`」覆盖父级 `false`，须在子记录显式写 `show: true`；各单用 §3 审计脚本自检。
- 留给 ② 替换的过渡点（本单原样搬）：`menu.vue` 的 `isMenuAuthorized`（读路由 `meta.checkAuth` + `useVerify().authVerifyData`）、`app.vue` 的 `accountStore.bizs` 监听、`NoPermission`（② 删）、路由守卫、`/403/:id`；`meta.ts` 的 `auth.view` / `available` 函数类型、`permissionData` / `errorMessage` 字段由 ② 按 PR 补。
- 同文件先后改：`views/bill/route-config.ts` 本单只给 `billManage` 加 `show:false`，路径迁移归 ⑦；`views/index.ts` 本单定形，② 包 `injectStatusComponents`，⑦ 加 `resourceViews`、删 `billViews` / `resourceAccessViews`。

## 执行顺序

1. 在 HEAD 上跑 §3 审计脚本，存基线 JSON
2. 菜单符号 `Symbol` 化及消费方适配（§0）
3. 路由归属：`router/index.ts` 顶层父路由 + 状态路由 + 404 兜底；`views/index.ts` 展平分组包装节点（§1）
4. 菜单服务（§2）
5. 面包屑：`Meta` / `use-breadcrumb` 按 PR，全量改写 `isShowBreadcrumb`，补 `show:false`（§3）
6. 布局组件、状态页、样式迁移（§4）；接线 `app.vue`（§5）
7. 清理旧壳（§6）
8. 重跑审计脚本与基线对比为零差异；`npm run lint`、`npm run build`、`bkdevbuddy_lint`；按「风险与验证」逐项对照

> 1 必须在任何改动前完成，否则无法证明面包屑逐页一致；3 依赖 2（顶层父路由 name 即一级菜单 id，均为 Symbol）；7 等接线后旧壳无引用再删。

## 0. 菜单符号 Symbol 化

**文件**：`src/constants/menu-symbol.ts`、`src/router/utils/history-storage.ts`、`src/router/meta.ts`、`src/vue-router.d.ts`、`src/router/hooks/use-back.ts`

`menu-symbol.ts` 58 个导出全部改为 `Symbol('<原字符串>')`。已 rg 审计 HEAD 中依赖「值是字符串」的点（排除随旧壳删除的 `views/home/**`）：

| 类别 | 审计结果 | 处理 |
|---|---|---|
| 与字符串字面量比较（`'menu_…'`） | 0 处（`rg "['\"\`]menu_[a-z_]+['\"\`]"` 除定义外无命中） | — |
| `includes(route.name)` / `name?.includes` | `views/scheme/index.tsx` 比的是 `'scheme-*'` 字符串名，非菜单常量；`business-selector` 的 `route.name?.includes?.('BusinessDetail')` 对 Symbol 返回 `undefined`（已有 `@ts-ignore`），详情页名均为字符串，行为不变 | 不改 |
| `activeKey?.toString()` / 菜单 key | 仅旧壳 3 处（`views/home/index.tsx` 276/303/341），随旧壳删除 | 新 `menu.vue`/`menu-item.vue` 两侧统一用 `toMenuKey(id)` = `typeof id === 'symbol' ? id.description : id`（`bk-menu :active-key` 与 `bk-menu-item :key` 同一函数，兼容 `'businessRecyclebin'` 等字符串 id） |
| `getRouteLinkParams` 的 `includes(config.name)` 白名单 | 旧壳，随删 | 侧栏一律按 name 跳转 |
| 序列化（JSON / sessionStorage / query） | `router/utils/history-storage.ts`：`routerAction.redirect(…, { history: true })` 把 `currentRoute.name` `JSON.stringify` 进 sessionStorage，Symbol 会变 `undefined`，返回箭头与 `_f` 回退全部失效；`remove(name)` 按 name 比对同样失效 | 取 PR 版：`serialize` 把 symbol name 写成 `description`，`deserialize` 用 `menu-symbol` 导出建的 `description → symbol` 注册表还原；类型改用 `RouteRecordName`（已安装 `vue-router@4.1.3` 无 HEAD 引用的 `RouteRecordNameGeneric`） |
| 作对象 / Map key | 0 处（`rg "\[MENU_[A-Z_]+\]"`、`new Map(…MENU_` 无命中） | — |
| `String(name)` / 模板插值 / i18n `t(MENU_…)` | 0 处 | — |
| `name` 经 `string` 类型 props 传递 | 0 处；`bill/bill/header/index.tsx` 的 `links` 只把 name 交给 `RouterLink`，可接受 Symbol | 不改 |
| `router.push/redirect/open({ name: MENU_… })`（约 30 处） | vue-router 原生支持 Symbol name | 不改 |
| `menu.relative`、`activeKey`、`owner` 类型 | `meta.ts` 现为 `any` / `symbol \| string` / `symbol \| string`，默认 `owner = ''` | 见 §3 |
| `RouteMeta` 类型增强 | HEAD 无 `vue-router.d.ts` | 取 PR：`interface RouteMeta extends RouteMetaConfig {}` |

- `use-back.ts` 取 PR：`HistoryStorage.pop()` 结果无 `name` 时回退 `menu.relative`。
- 类型检查：`npm run build` 是否做 TS 类型检查待核实；若不做，在编辑器 TS 诊断中确认改动文件无新错。

## 1. 路由归属

**文件**：`src/router/index.ts`、`src/router/module/common.ts`、`src/views/index.ts`

| 顶层记录 | 本单内容 |
|---|---|
| `statusRouters` | 取 PR：`{ name: '404', path: '/404' }`、`{ name: 'error', path: '/error' }`，`meta.layout.breadcrumb.show = false` |
| `MENU_BUSINESS` `/business` | `redirect: { name: MENU_BUSINESS_HOST_MANAGEMENT }`，`children: businessViews`；**不**改为 `/business/:bizId` |
| `MENU_SERVICE` `/service` | `redirect: { name: MENU_SERVICE_TICKET_MANAGEMENT }`，`children: serviceViews`（`serviceInside` 的 `/service/service-apply/account-add` 并入，不再顶层平铺；绝对路径不变，侧栏才会显示） |
| `MENU_RESOURCE` `/resource`【⑦删】 | `redirect: { name: MENU_RESOURCE_RESOURCE_MANAGEMENT }`，`children: resourceAccessViews`（= `router/module/resource.ts` + `resource-inside.ts`） |
| `MENU_SCHEME` `/scheme` | `redirect: { name: 'scheme-recommendation' }`，`children: schemeViews`（子记录 `/scheme` 与父同址，vue-router 先匹配子记录，裸地址表现不变） |
| `MENU_BILL` `/bill`【⑦删】 | 保留 `children: billViews`，加 `redirect: { path: '/bill/account-manage' }`（落到一期权限分流） |
| `/403/:id` | 从 `common.ts` 保留，补 `layout.breadcrumb.show = false`；HEAD 的 `{ path: '/403', redirect: '/403' }` 原样保留（归 ②） |
| `/` | 取 PR：`redirect: { name: MENU_BUSINESS }`（query 保留） |
| catch-all | 取 PR：`{ path: '/:pathMatch(.*)*', redirect: () => ({ name: '404', query: {} }) }`；删除 `common.ts` 三条 `redirect: '/'` 兜底 |

- 父路由不带 component，`RouterView` 跳过无组件记录，**页面地址零变化**；`redirect` 对象形式默认保留 query（`?bizs=` 随之带过去）。路由守卫本单不改（归 ②）。
- `views/index.ts`：`businessViews = business.flatMap((group) => group.children)`、`serviceViews = [...service.flatMap(...), ...serviceInside]`（去掉 `groupTitle` 包装节点；其无组件，子路由 URL 不变；全仓无其它 `groupTitle` 读取方，已 Grep）；新增过渡导出 `resourceAccessViews`【⑦删】、`schemeViews`；`billViews` 不变。
- 已 Grep：HEAD `src/` 无 `.matched` 读取，包父路由不影响现有页面。

## 2. 菜单服务

**文件**：`src/common/menu-service.ts`

结构取 PR（`IMenu`、`menus` 数组、`getMenus` 递归 visibility 过滤），内容按 HEAD：

```ts
const menus: IMenu[] = [
  { id: MENU_BUSINESS, i18n: '业务资源', menu: businessMenus },
  { id: MENU_SERVICE, i18n: '服务请求', menu: serviceMenus },
  { id: MENU_RESOURCE, i18n: '资源接入' },                                   // 无 menu = 不显示侧栏【⑦改】
  { id: MENU_SCHEME, i18n: '资源选型', visibility: () => ENABLE_CLOUD_SELECTION === 'true' },
  { id: MENU_BILL, i18n: '资源运营', visibility: () => ENABLE_ACCOUNT_BILL === 'true', menu: billMenus }, // 【⑦改】
];
export const getMenus = () => filter(menus);                       // 顶栏
export const getCurrentTopMenu = (route) => {                      // 一级菜单不过滤 visibility，子项过滤
  const top = menus.find((item) => item.id === route.matched[0]?.name);
  return top && { ...top, menu: filter(top.menu ?? []) };
};
```

- **开关只作用于顶栏**（与 HEAD 一致）：PR 的 `menu.vue` 用过滤后的 `getMenus()` 查侧栏，开关关时侧栏会空，故改用 `getCurrentTopMenu`。
- `getMenuRoute` 改为递归（业务主机等是 `path: ''` 子记录），找不到返回 `{ name: '', path: '' }`；侧栏跳过 `route.name` 为空的项、顶栏跳过 `!router.hasRoute(id)` 的项（单项配置错误只影响该项）。
- 菜单项 `id` = 路由 `meta.activeKey`；`group` 渲染为不可折叠的 `bk-menu-group`；`icon` 保留 HEAD 逐项图标（PR 业务侧只有 groupIcon，不取）。

| 一级 | 分组 | 菜单项（i18n · icon · 路由 name） |
|---|---|---|
| 业务资源 | 资源 | 主机 `bkhcm-icon-host` `MENU_BUSINESS_HOST_MANAGEMENT`；硬盘 `bkhcm-icon-disk`；VPC `bkhcm-icon-vpc`；子网 `bkhcm-icon-subnet`；弹性IP `bkhcm-icon-eip`；网络接口 `bkhcm-icon-network-interface`；路由表 `bkhcm-icon-route-table`；安全组 `bkhcm-icon-security-group`；负载均衡 `bkhcm-icon-loadbalancer` `MENU_BUSINESS_LOAD_BALANCER`；证书托管 `bkhcm-icon-cert` `MENU_BUSINESS_CERT_MANAGEMENT`（镜像不列） |
| 业务资源 | 其他 | `t('单据管理')` `bkhcm-icon-my-apply` `MENU_BUSINESS_TICKET_MANAGEMENT`；任务管理 `bkhcm-icon-bushu`；操作记录 `bkhcm-icon-operation-record` |
| 业务资源 | 账号 | 云账号管理 `bkhcm-icon-account-manage` `MENU_BUSINESS_CLOUD_ACCOUNT` |
| 业务资源 | 回收站 | 回收站 `bkhcm-icon-recyclebin` `'businessRecyclebin'`（业务侧沿用现名，字符串） |
| 服务请求 | 资源 | `t('单据管理')` `bkhcm-icon-my-apply` `MENU_SERVICE_TICKET_MANAGEMENT`；`t('权限策略库')` `bkhcm-icon-my-apply` `MENU_SERVICE_PERMISSION_POLICY` |
| 资源运营 | 云账单管理 | 一期 `billMenus` 三项原样（一级账号、二级账号、云账单管理） |

- **显示名逐字一致**：HEAD 侧栏显示路由 `meta.title`，其中三处单据管理/权限策略库是模块加载时 `t()` 的结果，其余是原始中文；menu-service 照此写，`menu-item.vue` 与 PR 一样直接输出 `menu.i18n`。顶栏 `t(nav.i18n)` 与 HEAD `t(name)` 同 key。
- 删除一期的 `getBillMenus`（唯一消费方随旧壳删除）。

## 3. 面包屑：按 PR 全量改写

**文件**：`src/router/meta.ts`、`src/hooks/use-breadcrumb.ts`、`src/typings/common.ts`、`src/router/module/business.ts`、`src/router/module/scheme.ts`、`src/router/module/service-inside.ts`、`src/router/module/common.ts`、`src/views/load-balancer/route-config.ts`、`src/views/task/route-config.ts`、`src/views/ticket/route-config.ts`、`src/views/operation-log/route-config.ts`、`src/views/cloud-account-manage/route-config.ts`、`src/views/cloud-account-manage/permission-policy/route-config.ts`、`src/views/bill/route-config.ts`

- **`use-breadcrumb.ts` 取 PR**：`immediate` + `!oldMeta` 首次初始化；`display = meta.layout?.breadcrumb?.show !== false`；`back = meta.layout?.breadcrumb?.back ?? true`（`IBreadcrumb.back`）；返回值改名 `data`（外部唯一调用方 `task/details/clb.vue` 只用 `setTitle`）。标题在 PR 的 `meta.menu?.i18n` 后加回退 `|| meta.title`：HEAD 裸 meta 路由（如主机列表 `title: '主机'`）没有 `menu`，不回退会丢标题。
- **`meta.ts` 取 PR 形态**：`RouteMetaConfig` 删除 `isShowBreadcrumb`；`activeKey?: symbol | string`（保留 string：`'businessRecyclebin'`、`'resourceAccount'`、`'serviceApply'` 等历史 key）、`owner?: symbol`、`menu.relative?: string | symbol`、`layout.breadcrumb { show?, back? }`；`title` 标 `@deprecated`。`checkAuth`（② 前菜单在用）、`notMenu`/`icon`（已无消费方，随 `router/module` 在 ⑦ 清理）类型暂留。
- **默认值设计（保证改写前后逐页一致）**：选「`Meta` 不为未传入的字段产出默认值」，具体是 `layout`、`activeKey`、`owner` 三个字段**只在传入时出现**在 meta 上（去掉 HEAD 的 `layout = {}`、`owner = ''` 初始值，不取 PR 的 `Symbol.for('')` / `{ show: true, back: true }` 初始值），默认语义全部由消费方给出（`show !== false`、`back ?? true`、`activeKey` 缺省即不高亮），与 PR 运行结果相同。`title`、`menu`（含 `i18n = title`）、`auth`、`extra`、`available`、`authKey`、`view` 维持 HEAD 的产出方式，构造器保留按传入键通配复制（不然会丢 `checkAuth`、`isFilterAccount`）。理由：vue-router 的 `route.meta` 是 matched 链上各记录 meta 的浅合并。改写前，子记录 `new Meta` 产出的 `layout: {}` 只遮住父级 `layout`，父级的 `isShowBreadcrumb` 照样继承下来；改写把 `isShowBreadcrumb` 搬进 `layout`，`layout` 就必须像原来的 `isShowBreadcrumb` 一样能继承，即子记录不传就不能产出 `layout`。若采用 PR 的默认值，所有「父写可见性、子用 `new Meta` 不写」的记录（如主机/硬盘等列表的 `path: ''` 子记录、负载均衡页签）都要逐条补显式值，改动更大，后续新增子路由也容易漏。`menu` 不改，是因为它决定返回箭头（`menu.relative`），HEAD 依赖子级 `menu` 遮住父级。
- **改写规则**：`isShowBreadcrumb: false` → `layout: { breadcrumb: { show: false } }`；`isShowBreadcrumb: true` → 删除（默认即显示）。若审计显示父链上有 `false` 要被覆盖，则写 `show: true`；现网有 27 处，分布为 `business.ts` 19、`operation-log` 3、`ticket` 2、`task` 1、`cloud-account-manage` 1、`permission-policy` 1。现网既无 `isShowBreadcrumb` 也无 `layout.breadcrumb` 的叶子路由原本不显示，新默认会显示，需补 `show:false`；静态阅读得到的清单如下，以脚本为准：`business.ts` 镜像列表与详情、`gcp/detail`、`template/detail`、`applyCvm`/`applyVPC`/`applyDisk`/`applySubnet`；`load-balancer` 入口记录（页签子记录 Meta 不再产出 `layout`，可继承，脚本确认）；`task` 列表；`ticket` 两个详情；`bill` 的 `billManage`；`scheme.ts` 父记录 `resource-scheme`；`service-inside.ts` 的 `applyAccount`；`common.ts` 的 `/403/:id`。
- 资源接入（`router/module/resource*.ts`）不补：现网由壳按地址隐藏，本单由 `main-content.vue` 的同一个过渡分支隐藏【⑦删】，这些路由在 ⑦ 删除或由 ③/⑥ 重建。
- 完成门槛：`rg isShowBreadcrumb src` 为 0；审计脚本前后零差异。
- **审计脚本**（已登录页面控制台；改前在 HEAD 跑存基线，改后重跑对比）：

```js
const r = document.querySelector('#app').__vue_app__.config.globalProperties.$router;
const key = (n) => (typeof n === 'symbol' ? n.description : n);
copy(JSON.stringify(r.getRoutes().filter((x) => x.components?.default && !x.redirect).map((x) => {
  const path = x.path.replace(/:\w+(\([^)]*\))?[?*+]?/g, '1');
  const { meta: m, matched } = r.resolve(path);
  const legacy = /^\/resource\/.+$/.test(path); // 改后换成 key(matched[0]?.name) === 'menu_resource'
  return { path, name: key(matched.at(-1)?.name), top: key(matched[0]?.name), activeKey: key(m.activeKey),
    show: !legacy && (BEFORE ? !!(m.layout?.breadcrumb?.show ?? m.isShowBreadcrumb) : m.layout?.breadcrumb?.show !== false),
    title: BEFORE ? m.title : m.menu?.i18n || m.title, back: m.menu?.relative && key(m.menu.relative) };
})));
```

  改前 `BEFORE = true`、改后 `false`；按 `path` 对比 `show`/`title`/`activeKey`/`back` 必须逐条相等（`top` 只看改后）。`__vue_app__` 在生产包是否可取待核实，取不到就临时在 `main.ts` 挂 `window.__router`，不提交。
- **实施记录（2026-10-08）**：编码时无浏览器，同一套判定逻辑改在 node 里跑：仓库外脚本用 `@swc/core` 转译 `src/router/index.ts` 及其依赖（i18n / store 打桩、`createWebHashHistory` 换 `createMemoryHistory`），对 `getRoutes()` 逐条 `resolve` 后套用上面的 `show`/`title`/`activeKey`/`back` 公式；HEAD 快照（`git archive HEAD`）与改后各跑一次。改后 84 条页面路由零差异（新增的只有 `/404`、`/error`）；所属一级菜单的唯一变化是 `/403/:id` 由「资源接入」变为无（见「风险与验证」）。补 `show:false` 的最终清单与上文静态清单一致：`load-balancer` 入口与购买页、`task` 列表、`ticket` 两个详情、`billManage`、`resource-scheme`、`applyAccount`、`business.ts` 镜像父记录、`gcp/detail`、`template/detail`、四个 `apply*`、`/403/:id`。上线前仍需在浏览器跑一次本脚本复核。

## 4. 布局组件

**文件**：`src/components/layout/header.vue`、`src/components/layout/menu.vue`、`src/components/layout/menu-item.vue`、`src/components/layout/main-content.vue`、`src/components/layout/breadcrumb.vue`、`src/components/layout/footer.vue`、`src/components/layout/business-selector.tsx`、`src/components/release-note/index.vue`、`src/views/status/404.vue`、`src/views/status/error.vue`、`src/style/app-shell.scss`、`src/style/index.scss`、`src/style/common.scss`、`src/language/lang.ts`

| 文件 | 取 PR | 按 HEAD 保留 / 改 |
|---|---|---|
| `header.vue` | `RouterLink` 链接样式与 `.router-link-active` 激活（链接指向一级父路由，当前路由 matched 含该父路由即激活，与 `getCurrentTopMenu` 等价；404/403 不在任何父路由下，自然无激活；`query` 不参与激活判断）、版本日志 / 语言 / 用户布局；退出登录不再用 `usePagePermissionStore`（同 PR），但不照搬 PR 的字符串拼接：改用 `utils/login-helper.ts` 的 `gotoLoginPage(undefined, true)`，`c_url` 经 `URLSearchParams` 转义——hash 路由下未转义的 `#` / `&` 会截断回跳地址 | 业务资源链接带 `query: { bizs: accountStore.bizs }`；点击非 `MENU_RESOURCE` 项时 `useResourceAccountStore().clear()`【⑦改】；语言切换沿用 HEAD：同值不处理，`document.cookie = blueking_language=…; domain=BK_DOMAIN`、`jsonp` 后 `reload`（HEAD `index.html` 无 `BK_HCM_DOMAIN`，已 Grep） |
| `menu.vue` | 按 `group` 聚合（首次出现顺序）、未分组项直出、`activeKey` 高亮 | 去收藏、去 `bk-submenu` 折叠，分组用 `bk-menu-group :name`；`:active-key="toMenuKey(route.meta.activeKey)"`；`isMenuAuthorized(menu)` = `router.resolve({ name }).meta.checkAuth` 不存在或 `authVerifyData.permissionAction[checkAuth]` 为真，整组空时隐藏（②替换）；业务资源且侧栏展开时顶部渲染 `BusinessSelector :reload`；根节点保留 `home-menu` class；宽度 240 |
| `menu-item.vue` | `RouterLink` + `bk-menu-item`、`menu.icon` | 去收藏星标；`:key="toMenuKey(menu.id)"`；`to = { name: menu.route.name, query: 业务资源 ? { bizs } : undefined }`；保留 `text-ov` 省略结构 |
| `main-content.vue` | 面包屑 + 主区、`min-width: 1089px`、状态页 `status-view`（免最小宽度）、`:deep(.page-container)`/`:deep(.detail-content-wrap)`（供 ③～⑥ 移植页） | 主区同时带 `view-warp` class（HEAD 全局样式依赖，保留 `flex:1; overflow-y:auto`）；`/service/ticket` 加 `no-padding`；`isRouterAlive` prop 控制 `RouterView`；`isLegacyResourceAccess` 分支：不渲染面包屑，渲染 `resource-manage-container` + `AccountVendorGroup` + `RouterView.router-view-content`【⑦删】；named view 归 ② |
| `breadcrumb.vue` | 结构、`data.back !== false && from` 才显示返回 | 保留 `#breadcrumbLeft`（`cloud-account-manage/index.vue` Teleport 目标，3904a5ef3b）与 `#breadcrumbExtra`；不加 PR 的 `#breadcrumbHead`（无消费方） |
| `footer.vue` | 原样；`app.vue` 仅 `route.name === MENU_BUSINESS_HOST_MANAGEMENT` 时渲染（等价 HEAD `path === '/business/host'`） | — |
| `views/status/404.vue`、`error.vue` | 原样；`lang.ts` 补「当前页面不存在或功能暂未开放」「返回首页」「页面加载出错，请稍后重试」「刷新页面」英文（HEAD 与 PR 均无） | — |
| `components/release-note/index.vue` | 由 `views/home/release-note/index.vue` 移入（PR 版只多 `@ts-expect-error` 与颜色） | — |
| `business-selector.tsx` | — | 由 `views/home/business-selector/index.tsx` 原样移入（PR 的 `global.vue` 属 `/business/:bizId` 方案，本单不取；做 `/business/:bizId` 时按 PR 的方式改用 `components/business-selector/global.vue`，见子单「菜单整合二期：收尾体验（菜单收藏、业务路由、分组折叠）」） |
| `style/app-shell.scss` | — | `views/home/index.scss` 全量迁入、在 `style/index.scss` 引入，选择器不变（`.view-warp` 系列、`.navigation-*`、`.bk-hcm-app-selector*`、`.hcm-create-business-dialog-*`、`.home-menu`、侧栏底色、`.resource-manage-container`）；PR 让该文件随旧壳失效，会丢页面依赖的全局样式，不跟 |
| `style/common.scss` | `html, body { min-width: 1366px }` | — |

实施偏差（2026-10-08）：

- `style/app-shell.scss` 未逐字全量迁入：删掉只服务旧壳 DOM、迁入后无任何匹配的选择器（`.left-header`、`.bk-hcm-header`、`.biz-select-warp`、`.navigation-breadcrumb`、`.navigation-content`、`.all-content`），以及 `.container-footer { height: 52px }` + `.home-page .no-footer` 这一对（bk-navigation 总是渲染 `.container-footer`，新壳没有 `.home-page`，留着会让所有页面底部多出 52px 空白；现在页脚只在主机列表通过 `#footer` 插槽渲染，高度由 `HcmFooter` 自身 52px 决定）。其余选择器原样保留，含 `.menu-warp .menu-icon`（bk-menu 图标容器类名就是 `menu-icon`，`menu.vue` 的 `bk-menu` 保留 `menu-warp` class）；`font-family: MicrosoftYaHei` 加 stylelint 行内豁免以保持原值。
- `main-content.vue` 不用 PR 的 `.main-layout` 包裹层与 `height: calc(100% - 52px)`：`.main-content` 用 `flex: 1; min-height: 0` 占满 `.container-content`（HEAD 已是纵向 flex），主区 `.main-view.view-warp` 用 `flex: 1` 撑开，因此 `app.vue` 也不需要 PR 的 `has-footer` 高度修正。
- `views/status/404.vue` 的「返回首页」用 `routerAction.redirect({ name: MENU_BUSINESS })` 代替 PR 的 `router.push('/')`（`fe-router-action` 规则）；`header.vue` / `menu-item.vue` 去掉 PR 的 `pr5`/`pl5`/`pr16` 间距工具类，改为组件内语义 class。
- `app.vue` 用 `import logo` 渲染 `#side-icon`，scoped 样式把 `.title-desc` 调成 18px、左距 10px，对齐 HEAD 旧标题。

## 5. 接线：`app.vue`

**文件**：`src/app.vue`

按 PR 结构：`bk-navigation`（`top-bottom`、`side-title` + `side-icon`、`need-menu`、`default-open`/`@toggle`、`has-footer`），四个插槽挂 HcmHeader / HcmMenu / HcmMainContent / HcmFooter；scoped 覆盖 `.navigation-container { max-width: none }`。在 PR 基础上：

- `needMenu = !!getCurrentTopMenu(route)?.menu?.length`：资源接入、资源选型不声明 `menu`，403/404 无所属一级菜单，均不显示侧栏；`hover-width="240"` 沿用 HEAD。
- 按 PR 去掉外层容器，`bk-navigation.hcm-app` 直接作根（`height: 100%`，依赖 `html/body/#app` 的 100%）；与 PR 的差异：`Notice` 放在 `bk-navigation` 之前（PR 放在之后，跑马灯会被挤出可视区），有跑马灯时加 `has-notice` 把高度减 40px（与 `useCalcTopWithNotice` 同口径）；`onMounted` 保留 `useChromeNotice()`（PR 漏掉）。
- 旧壳副作用落点：全局预取 `fetchRegions(TCLOUD/HUAWEI)` / `fetchBusinessMap` / `fetchAllCloudAreas`；`watch(accountStore.bizs, …, { immediate: true })` 改写 `pageAuthData.bk_biz_id` 并 `getAuthVerifyData`（②删）；`watch(route.path, …, { immediate: true })` 复刻 `changeMenus` 的业务 ID 副作用：一级菜单为 `MENU_RESOURCE`【⑦删】/`MENU_SCHEME` 时 `updateBizsId(0)`，无所属一级菜单时 `updateBizsId('')`；`isRouterAlive` + `reload()` 在此持有并下传；根部挂 `<GlobalPermissionDialog />`（② 删）；`hasPagePermission` 为假时整页 `NoPermission`（全仓无 `setHasPagePermission` 调用，实际不触发，原样保留，② 删除，见 ② coding.md §10）；`provideBreadcrumb` / `providePermissionDialog` / `window.hcmPermissionDialog` 不变。

## 6. 清理

**文件**：`src/views/home/index.tsx`、`src/views/home/index.scss`、`src/views/home/breadcrumb.tsx`、`src/views/home/breadcrumb.scss`、`src/views/home/hooks/useChangeHeaderTab.ts`、`src/views/home/hooks/to-legacy-menus.ts`、`src/views/home/hooks/useCalcTopWithNotice.ts`、`src/views/home/release-note/index.vue`、`src/views/home/business-selector/index.tsx`、`src/router/header-config.ts`、`src/hooks/use-calc-top-with-notice.ts`、`src/views/business/load-balancer/group-view/specific-target-group-manager/target-group-detail/index.tsx`、`src/views/resource/resource-manage/common/header/detail-header.tsx`、`src/views/resource/resource-manage/account/createAccount/index.tsx`

- 删除整个 `views/home/` 与 `router/header-config.ts`；`getRouteLinkParams` 白名单作废。
- `useCalcTopWithNotice` 按 PR 移到 `hooks/use-calc-top-with-notice.ts`，同步改 3 处 import。
- 实施补充：`router/module/common.ts` 删除三条兜底时一并删掉文件尾部注释掉的旧路由（其中引用 `@/views/home/RootPath`），使 `rg views/home src` 为 0。

## 从 PR #1897 取 / 不取

| 项 | 处理 |
|---|---|
| `app.vue` 结构、header/menu/menu-item/main-content/breadcrumb/footer 拆分 | 取（按上表裁剪） |
| 菜单符号 `Symbol`、`history-storage` 序列化、`vue-router.d.ts`、`use-back` 回退 | 取 |
| 菜单收藏、分组折叠（`bk-submenu`、`use-menu-collapse`） | 不取 |
| `/business/:bizId`、`business-interceptor.ts`、`business-selector/global.vue` | 本单不取；由二期收尾子单按 PR 方式实现 |
| `use-breadcrumb` 语义（`show !== false`、`back`、`immediate`）、全量改写 `isShowBreadcrumb` | 取（标题加 `|| meta.title` 回退） |
| `Meta` 字段形态与类型 | 取；字段初始默认值（`Symbol.for('')`、`layout` 默认对象）不取，去通配复制不取（见 §3） |
| 顶栏链接样式与 `.router-link-active` 激活 | 取 |
| `html/body` 1366、主区 1089、状态页免最小宽度；404/error 与 catch-all | 取 |
| 状态 named view、`injectStatusComponents`、视图权限守卫、`meta.auth.view` 类型 | 不取（②） |
| 菜单树内容（工作台、资源选型并入工作台、资源运营新结构） | 不取（资源运营归 ⑦） |
| `views/index.ts` 业务侧按模块拆 route-config | 不取（只展平包装节点） |
| `components/release-note`、`hooks/use-calc-top-with-notice` 位置 | 取 |
| 语言切换新写法（`BK_HCM_DOMAIN`）、`main.ts` 去掉 `bkui-vue/dist/style.css` | 不取 |

## HEAD 新增能力必须保留（`git log 8fea9924df..HEAD` 核对）

- 业务侧「账号 / 云账号管理」菜单（3904a5ef3b、8119f45a4d）与 `#breadcrumbLeft`（3904a5ef3b）。
- 业务侧单据管理 `ticketRoutesBiz` 及菜单（07430fcaed）；服务请求「权限策略库」路由、菜单与 `checkAuth: 'cloud_vendor_config'`（b8a95d0915、38e7334a72 给 Meta 加 `checkAuth` 类型）。
- 一期资源运营：顶栏更名、`billMenus`、`/bill/account-manage` 的 `beforeEnter` 分流与旧地址兼容、`layout.breadcrumb` 字段名（01bca385c）。
- `useChromeNotice`（7dbb2e3a1b，早于 PR base，PR 的 app.vue 丢了）。
- 其余 HEAD 新增（安全组克隆目标地域 c5b61eff78、searchselect async 9c6ac9204f 等）在页面内部，本单只可能因 Symbol 化触及其中的 `{ name: MENU_… }` 跳转，行为不变。

## 风险与验证

| 风险 | 验证 |
|---|---|
| 面包屑、标题、返回箭头、高亮逐页一致（AC-008/012） | §3 审计脚本前后零差异；`rg isShowBreadcrumb src` 为 0；抽查主机列表/详情、负载均衡页签、云账单管理、资源选型、单据详情 |
| Symbol 化后历史返回失效 | 任务列表 → 详情（`history: true`）→ 返回箭头回到原列表且带原 query；刷新详情页后返回仍可用（sessionStorage 还原 Symbol） |
| Symbol 化后菜单高亮 key 对不上 | 逐项点击侧栏 + 直开详情/新建页，高亮项同改前；`'businessRecyclebin'` 字符串 id 同样高亮 |
| 顶栏 5 项顺序、开关、入口、英文、激活（AC-001/002/010） | 两开关各关一次：顶栏不显示，直接访问 `/scheme/recommendation`、`/bill/root-account` 页面与侧栏照常；逐项点击核对落点与 `.router-link-active` 高亮（业务资源带 `?bizs=` 仍高亮，404/403 无高亮）；切英文对照 |
| 侧栏逐项一致与权限过滤（AC-003/004/005/S01） | 逐组对比改前截图；无 `root_account_find` / `account_bill_find` / `cloud_vendor_config` 时对应项隐藏、二级账号仍在 |
| 业务参数 | 切换业务后点侧栏各项、顶栏「业务资源」，地址带 `bizs`；选择器切换后刷新、详情页回列表 |
| 资源接入/资源选型布局（AC-006/007） | `/resource/resource`、账号详情、`/resource/service-apply/clb`、`/resource/account/add`：无侧栏、无面包屑、左侧账号树、切账号刷新；经顶栏离开后账号选择被清空 |
| 旧书签与 404（AC-009/011） | 遍历顶栏入口、`/`、`/service/my-apply`、`/bill/account-manage/*`；随机地址进 404、无侧栏面包屑、顶栏无激活 |
| 403 页顶栏不再高亮「资源接入」 | 有意变化（旧逻辑按地址兜底的副作用），`/403/<id>` 无侧栏 |
| 全局样式迁移遗漏 | 抽查底部固定操作栏申请页（`.action-bar`/`.button-group`）、资源详情（`.i-detail-tap-wrap`）、业务选择器下拉与新建业务弹窗 |
| 最小宽度（AC-013） | 窗口 1280px 出现横向滚动；资源接入固定定位账号树（`width: 15%`）不遮挡内容 |
| logo/标题改用 `side-title` | 目视对比左上角，必要时在 app.vue scoped 样式补颜色/字重 |
| 性能（AC-P01） | 同环境对比首屏与切换一级菜单耗时，差异 ≤ 10% |

## 已确认（2026-10-08）

1. **菜单符号改为 PR 的 `Symbol`**：本单把整份 `menu-symbol.ts` 转为 `Symbol('<原字符串>')`，并适配 §0 审计出的全部依赖点；非菜单常量的历史路由名保持字符串；首批其余单从 ① 起一律使用 Symbol。
2. **面包屑完全按 PR**：全量改写 `isShowBreadcrumb` 为 `layout.breadcrumb`、采用 `show !== false` 默认显示与 PR 的 `Meta` 字段形态，删除 `isShowBreadcrumb`；为避免浅合并覆盖，`Meta` 不为未传入的 `layout`/`activeKey`/`owner` 产出默认值（§3），以审计脚本零差异作为验收门槛。
