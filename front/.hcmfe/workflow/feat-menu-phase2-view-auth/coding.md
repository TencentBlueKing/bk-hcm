# Coding — feat-menu-phase2-view-auth

需求文档：仓库根 `docs/reqs/二期视图权限体系.md`（TAPD [#1069995598138702302](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702302)）。机制背景见 `.hcmfe/docs/modules/auth.md`、`.hcmfe/docs/modules/menu-route.md`。

**目标**：页面查看权限改为「路由 `meta.auth.view` 声明 + 启动批量预鉴权 + 守卫判定 + 框架内 named view 无权限页」，替换 `store/common.ts` 的 `pageAuthData` 按地址拦截与 `/403/:id` 跳转；侧栏显隐改读视图权限；接口无权限自动弹申请弹窗。**拦截范围逐页与现网一致（R-004），不新增、不取消。**

## 跨单约定（只列影响本单的）

- 首批共 7 张子单、7 个 commit：① 页面布局与菜单服务化 → ② 视图权限体系（本单）→ ③ 资源纳管主体 → ④ 回收管理与操作记录 → ⑤ 账号管理 → ⑥ 资源详情页迁移 → ⑦ 资源运营重组（账单迁 `/resource/bill/*` 在 ⑦）。
- 菜单 / 路由 id 从 ① 起改为 PR #1897 的 `Symbol` 常量（`src/constants/menu-symbol.ts` 由 ① 转换并适配消费方）；`constants/auth-symbols.ts` 的 `AUTH_*` 本来就是 `Symbol`。本单所有以菜单 id 为键的结构（`viewAuthConfig.viewId`、`viewPermissions`、`hideMenu` 集合、`menu.vue` 过滤）直接按 Symbol 设计，与 PR 一致。旧模块路由（`router/module/scheme.ts`、`resource.ts`）的 name 仍是字符串 / i18n 文案，所以缓存键类型取 vue-router 的 `RouteRecordName`（`string | symbol`），两种都能存。
- ①（页面布局与菜单服务化，已提交 `943fd921e`）实际提供：
  - `router/index.ts`：顶层 `statusRouters`（`404` / `error`，组件 `views/status/404.vue`、`error.vue`）、`common`（`/403/:id`），以及五个无组件、带 `redirect` 的一级父路由 `MENU_BUSINESS` / `MENU_SERVICE` / `MENU_RESOURCE` / `MENU_SCHEME` / `MENU_BILL`，`children` 分别为 `views/index.ts` 导出的 `businessViews` / `serviceViews` / `resourceAccessViews` / `schemeViews` / `billViews`；末尾 catch-all 转 `404`。
  - `views/index.ts`：业务、服务请求的分组包装节点已展平（`flatMap(group => group.children)`）；`resourceAccessViews = [...resource, ...resourceInside]`、`schemeViews = scheme`、`billViews = [...billAccountManage, ...billManage, ...billAccountManageLegacy]`。
  - `main-content.vue`：两个 `RouterView`——资源接入（`matched[0].name === MENU_RESOURCE`）时包在云账号树容器里，其余直出；`isStatusView` 目前按 `route.name` 为 `404` / `error` 判定。
  - `menu.vue`：`isMenuAuthorized` 读 `router.resolve({ name }).meta.checkAuth` + `useVerify().authVerifyData`。
  - `app.vue`：保留 `accountStore.bizs` watcher（改写 `pageAuthData.bk_biz_id` 并 `getAuthVerifyData`）、`NoPermission` 分支、`GlobalPermissionDialog`。
  - 本单在其上改：`main-content.vue` 两个 `RouterView` 都加 `:name="view"`；`menu.vue` 的 `checkAuth` 过滤替换为视图权限；`router/index.ts` 守卫整体重写；`app.vue` 按 §10 清理。
- 原 `views/home/index.tsx` 里「`accountStore.bizs` 变化 → 给 `pageAuthData` 补 `bk_biz_id` → `getAuthVerifyData`」的 watcher 由 ① 原样迁走，本单**保留它**（旧按钮鉴权仍靠它，见 §6）。
- 业务侧沿用 `?bizs=` + 现有 BusinessSelector，不引入 `/business/:bizId`。
- ③–⑦ 新路由（含 ⑥ 资源详情页迁移的详情路由）按 §8 的接口声明查看权限，本单不替它们写路由。

## 执行顺序

1. 类型与工具：`auth-service.ts` 补 `IPermission` / `IVerifyResourceInstance` / `extractSignPermission`
2. 视图权限 store：`store/auth.ts` 加 `viewPermissions` 与 fetch / check 方法
3. 预鉴权配置：新建 `constants/view-auth.ts`，`store/index.ts` 的 `preload` 接入
4. meta 类型：`router/meta.ts` 的 `auth.view` 改为 `IAuthSign`；清掉任务管理上的占位 `auth.view`
5. 状态页与注入：`views/status/permission.vue`、`business.vue`，`views/index.ts` 的 `injectStatusComponents`，`main-content.vue` named view
6. 守卫：新建 `router/business-interceptor.ts`，重写 `router/index.ts` 的 `beforeEach`，旧按钮鉴权预取改挂在守卫首航
7. 迁移拦截：账单、权限策略库、资源选型、账号列表、回收站逐页声明 `auth.view`；`store/common.ts` 去掉 `path`
8. 侧栏：`menu.vue` 改按视图权限过滤，删除 `checkAuth`
9. 接口兜底：`http/index.ts` 无权限弹窗
10. `bkdevbuddy_lint` + 构建 + 逐页对照验证；收口时更新 `auth.md` 视图级章节与「现存缺口」

> 排序依据：5 必须在 6 之前（守卫一旦切 `view`，没注入状态组件的路由会白屏）；7 必须与 6 同一 commit（守卫切换后 `pageAuthData.path` 不再生效，不迁就等于放开拦截）。

## 1. 类型与工具（`src/common/auth-service.ts`、`src/store/auth.ts`）

**文件**：`src/common/auth-service.ts`、`src/store/auth.ts`

- 从 PR 取 `IVerifyResourceInstance`、`IPermission` 两个接口与 `extractSignPermission(sign, permission)`（按定义 `id` 过滤 actions、按 `resource_id` / `bk_biz_id` 过滤 instances）。
- `store/auth.ts` 现有的 `IVerifyResourceInstance` / `IVerifyResult` 改为从 auth-service 引入并 re-export（`hooks/use-permission-dialog.ts`、`components/permission/apply-dialog.vue` 从 store 引类型，不能断）。`getApplyPermUrl` 参数改为 `IPermission`。
- **只加不改**：`AUTH_DEFINITIONS` 既有条目一条不动。PR 对 `AUTH_UPDATE_ACCOUNT`（import→update + transform）与新增 `AUTH_DELETE_ACCOUNT` 属于账号按钮鉴权，不在本单。

## 2. 视图权限 store（`src/store/auth.ts`）

照搬 PR 的 `viewPermissions` / `fetchViewPermissions` / `hasViewPermission` / `getViewPermissionData` / `isViewPermissionCached` / `checkViewPermission` / `resetViewPermissions`，改动三处：

| 改动 | 原因 |
|------|------|
| `Map<symbol, …>` → `Map<RouteRecordName, ViewPermissionValue>`，`viewId: RouteRecordName` | 菜单 id 是 Symbol，但本批仍有字符串 name 的旧路由要走同一缓存；守卫直接用 `to.name` 作键，**不得 `String()`**（Symbol 转字符串后与预鉴权键对不上） |
| 接口失败**不写缓存**：`fetchViewPermissions` 失败只 `console.error`；`checkViewPermission` 动态请求失败直接抛出 | PR 失败时写 `[false, null]`，会显示「无权限」而不是 F-001 要求的「加载失败，可重试」；不缓存则守卫会再试一次、再失败进 error 视图 |
| 保留 PR 语义：`hasViewPermission` 未缓存时返回 `true`（守卫用）；侧栏与入口分流改用严格判定 `isViewPermissionCached(id) && hasViewPermission(id)` | F-004「权限数据未返回 / 接口失败时受控菜单不显示」 |

约束：`fetchViewPermissions` 按下标对齐 `results`，所以预鉴权的 sign 的 `transform` 不能返回数组（本单 5 条都不是）。

## 3. 预鉴权配置（新建 `src/constants/view-auth.ts`，改 `src/store/index.ts`）

**文件**：`src/constants/view-auth.ts`、`src/store/index.ts`

```ts
export interface IViewAuthConfig {
  viewId: RouteRecordName; // = 路由 name = 菜单 id（Symbol）
  authSign: IAuthSign;
  hideMenu?: boolean;  // 无权限时侧栏隐藏同 id 的菜单项
}
export const VIEW_AUTH_BIZ_ACCESS = Symbol('view_auth_biz_access');
export const viewAuthConfig: IViewAuthConfig[] = [
  { viewId: MENU_BILL_ROOT_ACCOUNT, authSign: { type: AUTH_FIND_ROOT_ACCOUNT }, hideMenu: true },
  { viewId: MENU_BILL_MAIN_ACCOUNT, authSign: { type: AUTH_FIND_MAIN_ACCOUNT } },
  { viewId: MENU_BILL_MANAGE, authSign: { type: AUTH_FIND_ACCOUNT_BILL }, hideMenu: true },
  { viewId: MENU_SERVICE_PERMISSION_POLICY, authSign: { type: AUTH_FIND_PERMISSION_POLICY_LIBRARY }, hideMenu: true },
  { viewId: VIEW_AUTH_BIZ_ACCESS, authSign: { type: AUTH_ACCESS_BIZ } }, // 任一业务访问，替代旧 biz_access
];
```

- **菜单显隐与页面拦截解耦**：是否拦截看路由有没有 `meta.auth.view`；是否隐藏菜单看这里的 `hideMenu`。所以云账单管理「侧栏按权限隐藏、直接访问可打开」= 有预鉴权 + `hideMenu`、路由不声明 `auth.view`；二级账号「菜单常显、点进去拦截」= 路由声明 `auth.view`、不配 `hideMenu`。PR 是「凡预鉴权即参与菜单过滤」，表达不了这两种情况，这是本单对 PR 的主要改动。
- `preload` 改为 `Promise.all([getFullBusiness(), getAuthorizedBusiness(), authStore.fetchViewPermissions(viewAuthConfig)])`，5 条合成一次 `/auth/verify`（AC-P01）。`main.ts` 已在 `preload().finally` 之后才装路由、挂载，首航守卫拿得到缓存。
- 不取 PR 的 `MENU_RESOURCE_MANAGE`（资源接入 / 资源纳管无页面拦截）、`MENU_RESOURCE_OPERATION_LOG`（现网资源侧操作记录无拦截）、`MENU_RESOURCE_RECYCLEBIN` / `MENU_SERVICE_ACCOUNT_MANAGEMENT` / `MENU_SCHEME_*`（本批这些页面还是旧路由名，见 §7）。

## 4. meta 类型（`src/router/meta.ts`、`src/views/task/route-config.ts`）

- `Auth.view` 改为 `IAuthSign | AuthViewFn`（PR 写法；`IAuthSign.type` 是 `symbol`，直接接收 `AUTH_*`；HEAD 现为 `{ type: string }`，必须改），`RouteMetaConfig` 补 `permissionData?`、`errorMessage?`、`available?: boolean | AvailableFn`；删除 `checkAuth?`。
- **Meta 构造函数不动**：PR 的构造函数把面包屑默认改成显示，会改变现网页面。
- **必须删** `views/task/route-config.ts` 的占位 `auth: { view: { type: 'biz_access' } }`：新守卫会对它做 `getAuthResources`，字符串 type 查不到定义直接抛错 → 任务管理进 error 视图。业务访问统一由业务拦截器处理。

## 5. 状态页与 named view

**文件**：`src/views/status/permission.vue`、`src/views/status/business.vue`、`src/views/index.ts`、`src/router/index.ts`、`src/components/layout/main-content.vue`

**`permission.vue`**：以 PR 版为底（`descriptions` 文案表、`getApplyPermUrl` 申请、申请后切「已申请，刷新页面」并 `reload`），补：

- 按 `route.meta.auth.view` 取 sign；`descriptions` 键为 auth symbol，补 `AUTH_FIND_ROOT_ACCOUNT`（文案取自 `views/error-pages/403.tsx`：「当前无"云账号-一级账号管理"权限」「该功能面向平台管理员……请联系管理员授权」）。`AUTH_FIND_PERMISSION_POLICY_LIBRARY` 现网 403 无专属文案，不补；`AUTH_FIND_ACCOUNT_BILL` 本单无拦截，不补。
- 一期规则移植：`const NO_SELF_APPLY_SIGNS = [AUTH_FIND_ROOT_ACCOUNT]`。命中时标题下提示「您还没有该功能的权限，请联系管理员授权」且**不渲染申请按钮**；其余显示「……可以点击下方的"去申请"获得权限」。标题统一「抱歉，您暂无该功能的权限」（PR 未设标题）。
- 无 `permissionData`（后端没返回）时不显示申请按钮，与 PR 一致。

**`business.vue`**：以 PR 版为底，三种状态 `noBiz` / `bizNotFound` / `bizUnauthed`；`noBiz` **补申请按钮**（现网 `/403/biz_access` 可申请，PR 去掉了），`permissionData` 来自 `VIEW_AUTH_BIZ_ACCESS` 的缓存。

**`injectStatusComponents(routes, permission = statusPermission)`**（`views/index.ts`，导出供 router 用；注入 `notFound` / `error` / `permission` 三个 named view），相对 PR 两处改动：

- **没有 `component` 的记录只递归、不注入**。给无组件记录挂 `components` 后 `RouterView` 不再跳过它，会整页白屏。① 已展平分组节点，现存的无组件记录是一级父路由（定义在 `router/index.ts`，不经过本函数）与账单旧地址的 `beforeEnter` + `children: []` 记录，规则仍要保留作兜底。
- 业务无权限页按「挂在哪个视图集合」选，而不是 PR 的 `meta.owner === MENU_BUSINESS`（HEAD 业务路由都没设 owner）：`businessViews` 用 `statusBusiness`；`serviceViews` / `resourceAccessViews` / `schemeViews` / `billViews` 用默认 `statusPermission`。五个集合都在 `views/index.ts` 导出处包一层。
- `router/index.ts` 顶层的 `common`（`/403/:id`）与 `statusRouters` 也包一层，守卫异常时 error 视图才有组件可渲染。
- 嵌套路由（如资源接入的 `resource-entry.vue` → 账号列表）由 `main-content.vue` 的出口渲染「第一个带组件的记录」，所以祖先记录同样带 `permission` / `error` 组件即可整块替换，内层 `RouterView` 不渲染，不必改各页面组件。

**`main-content.vue`**：两个 `RouterView` 都改为 `:name="view"`，`view = route.meta.view ?? 'default'`；`isStatusView` 改为 `view !== 'default'` 或 `route.name` 为 `404` / `error`（沿用 ① 的 `status-view` 居中样式）；资源接入在 `view !== 'default'` 时**不渲染云账号树容器**，直接出状态页（AC-S02：无权限页不展示页面业务数据）。

## 6. 守卫（`src/router/index.ts`、新建 `src/router/business-interceptor.ts`）

**文件**：`src/router/index.ts`、`src/router/business-interceptor.ts`、`src/store/common.ts`

`beforeEach` 按 PR 结构重写：

1. **旧按钮鉴权预取（兼容）**：首次导航 `await getAuthVerifyData(commonStore.pageAuthData)`，失败只记日志、不阻断（现网 `.then` 失败时导航会卡死，顺带修掉）。只为 `useVerify().authVerifyData` 的按钮级消费方（见下表）保留，**不再参与拦截**。不放进 `preload` 是因为 `hooks/useVerify.ts` 引 `@/store`，放进去成环。
2. **旧 403 页兼容**：`to.name === '403'` 且首次导航且 `permissionAction.biz_access` 为真时 `next('/')`，其余直接放行——与现网 `toCurrentPage` 完全一致（Q-001：旧地址可打开，不作为新拦截落点）。
3. **业务拦截器** `before(to)`（从 PR `business-interceptor.ts` 改写，不做 bizId 路由、不改地址、不写 localStorage / `accountStore`，这些仍由 BusinessSelector 负责）：
   - 不在 `MENU_BUSINESS` 下 → 跳过
   - `VIEW_AUTH_BIZ_ACCESS` 无权限 → `noBiz`（与现网 `biz_access` 同一判定）
   - `?bizs=` 缺省或为 0 → 放行（BusinessSelector 挂载后补参数，再进一次守卫）
   - `businessFullList` 非空且不含该业务 → `bizNotFound`
   - 在 `businessAuthorizedList` 里 → 放行；否则 `checkViewPermission(Symbol.for(\`biz_${id}\`), { type: AUTH_ACCESS_BIZ, relation: [id] })`，有权限放行，无权限 → `bizUnauthed` 并挂 `permissionData`
   - 命中状态即 `to.meta.view = 'permission'`、`to.meta.extra.businessStatus = status`
4. `to.meta.view = 'default'`；`available === false` → `notFound`（PR 原样，HEAD 无路由用到）
5. `meta.auth.view` → `checkViewPermission(to.name, sign)`（Symbol 原样作键），无权限 `view = 'permission'` + `permissionData`
6. 整体 `try/catch`：异常 → `view = 'error'` + `errorMessage`

删除：`toCurrentPage`、`pageAuthData.find(path)`、资源选型两页在站内导航时的放行特判、`isArray/isRegExp/isString` 引入。

**`store/common.ts`**：`pageAuthData` 保留（按钮级仍在读），**删掉全部 `path` 字段**，注释改为「旧按钮鉴权预取，不参与路由拦截」；字段名与 `updatePageAuthData` 不改名，避免牵动 bizs watcher 与 `403.tsx`。

仍读旧数据、本单保持可用的消费方（只读 `permissionAction` / `urlParams`，不受拦截迁移影响）：

| 消费方 | 读的 id | 本单处理 |
|--------|---------|----------|
| `views/scheme/scheme-recommendation/index.tsx`、`scheme-list/index.tsx` 页内 `return () => <ErrorPage />` | `cloud_selection_recommend` / `cloud_selection_find` | **删除这两处页内拦截与 `ErrorPage` 引入**（由守卫接管，否则两套判定可能不一致）；`scheme-list` 的按钮级判定保留 |
| `views/bill/route-config.ts` 的 `account-manage` `beforeEnter` | `root_account_find` | 改为 `authStore.isViewPermissionCached(MENU_BILL_ROOT_ACCOUNT) && authStore.hasViewPermission(...)`，不再依赖旧数据的加载时机 |
| `views/error-pages/403.tsx`、`views/ticket/.../account-apply-detail/index.tsx`、资源详情 / 主机 / 安全组 / 回收站 / 账号 / 服务申请等按钮 | 各按钮 id | 不动（按钮级改造不在本单）|

## 7. 迁移现有拦截（F-002）

| 页面（现网拦截来源） | 路由（文件） | 声明 | 预鉴权 / 菜单 |
|------|------|------|------|
| 方案推荐 `/scheme/recommendation` | `scheme-recommendation`（`src/router/module/scheme.ts`） | `auth.view: { type: AUTH_CREATE_CLOUD_SELECTION_SCHEME }` | 不预鉴权（守卫首访动态查一次后缓存）；无侧栏 |
| 部署方案列表 `/scheme/deployment/list` | `scheme-list`（同上） | `AUTH_FIND_CLOUD_SELECTION_SCHEME` | 同上 |
| 账号列表 `/resource/account` | `t('账户')`（`src/router/module/resource.ts`） | `AUTH_FIND_ACCOUNT` | 同上 |
| 回收站 `/resource/recyclebin` | `t('回收站')`（同上） | `AUTH_FIND_RECYCLE_BIN` | 同上 |
| 一级账号 `/bill/root-account` | `MENU_BILL_ROOT_ACCOUNT`（`src/views/bill/route-config.ts`） | `AUTH_FIND_ROOT_ACCOUNT` | 预鉴权 + `hideMenu` |
| 二级账号 `/bill/main-account` | `MENU_BILL_MAIN_ACCOUNT`（同上） | `AUTH_FIND_MAIN_ACCOUNT` | 预鉴权，菜单常显 |
| 云账单管理（无页面拦截） | `MENU_BILL_MANAGE`（同上） | **不声明** | 预鉴权 + `hideMenu` |
| 权限策略库 `/service/permission-policy` | `MENU_SERVICE_PERMISSION_POLICY`（`src/views/cloud-account-manage/permission-policy/route-config.ts`） | `AUTH_FIND_PERMISSION_POLICY_LIBRARY` | 预鉴权 + `hideMenu` |
| 业务资源 `/business/*`（`biz_access`） | 业务拦截器 | —— | `VIEW_AUTH_BIZ_ACCESS` |

- 只在列表路由声明，创建页（`root-account/create` 等）、方案详情不声明——现网按 `to.path` 精确匹配，同样不拦。
- 两个旧模块路由的 name 是字符串 / i18n 文案，不为它们新建常量：④、⑤ 会用新路由替代它们，届时新路由用常量名并进预鉴权。`router/module/*.ts` 在 #1897 中已废弃，但本分支仍由它们承载现网路由，本单只加 `meta.auth.view` 一项，不做其它整理；各文件的退场批次见 `docs/reqs/二期拆单总览.md`「旧路由文件 `router/module/` 退场计划」。
- 一并删除 `bill/route-config.ts`、`permission-policy/route-config.ts` 的 `checkAuth`，更新 `bill/route-config.ts` 顶部两行说明。

## 8. 侧栏（`src/components/layout/menu.vue`）与给 ③–⑦ 的接口

**文件**：`src/components/layout/menu.vue`

```ts
const hideMenuIds = new Set(viewAuthConfig.filter((c) => c.hideMenu).map((c) => c.viewId));
const filterMenusByPermission = (menus: IMenu[]) =>
  menus.filter((m) => !hideMenuIds.has(m.id) || (authStore.isViewPermissionCached(m.id) && authStore.hasViewPermission(m.id)));
```

替换 ① 中按 `meta.checkAuth` + 旧 `authVerifyData` 的过滤；分组按过滤后的结果聚合，全空自然不显示。

菜单 id 与 `viewId` 都是同一个 Symbol 常量（同一模块导出的同一引用），`Set.has` 按引用比较即可；不要用 `Symbol.for` 或 `description` 比较。

**③–⑦ 声明新页面查看权限的方式**：

1. 路由 name 用 `menu-symbol.ts` 的 Symbol 常量，`meta: { ...new Meta({ auth: { view: { type: AUTH_XXX } } }) }`；只用静态 sign（缓存按路由 name，带动态 relation 的 sign 会串）。
2. 有菜单的页面在 `viewAuthConfig` 加 `{ viewId: <路由 name>, authSign }`；无权限要隐藏菜单再加 `hideMenu: true`，菜单 id 必须等于 `viewId`。
3. 路由数组必须经 `views/index.ts` 的 `injectStatusComponents` 汇入，否则无权限时白屏。
4. 需要说明文案的在 `permission.vue` 的 `descriptions` 加一项；不可自助申请的加进 `NO_SELF_APPLY_SIGNS`。
5. 业务资源下的页面不需要声明，业务拦截器统一处理。
6. 详情类路由（⑥）默认不声明——现网详情页没有页面级拦截；若 ⑥ 需要，同样按 1–3 声明。
7. 搬地址（⑦ 账单迁 `/resource/bill/*`）时 `meta.auth.view` 随路由走、路由 name 不变，预鉴权与菜单关联不用改；改了 name 必须同步 `viewAuthConfig`。

## 9. 接口无权限弹窗（`src/http/index.ts`）

**文件**：`src/http/index.ts`、`src/components/layout/business-selector.tsx`

照搬 PR diff：`handleResponse` 的 `reject` 带上 `data`、`permission`；HTTP 403 与业务码 `2030403`（`pkg/criteria/errf/code.go` 的 `PermissionDenied`，后端在响应体 `permission` 字段返回缺失权限）有 `permission` 时调 `window.hcmPermissionDialog?.show(permission)`，没有时保留原失败提示；删除无人消费的 `bus.$emit('show-forbidden')` 与 `bus` 引入。补一处：弹窗实例未挂载（`preload` 阶段）时退回原失败提示。单实例弹窗，后到的覆盖先到的，始终只有一个（F-005）。`globalError: false` 的调用方自行处理，不受影响。

兜底弹窗上线后，壳层自身的业务级请求不能在业务无权限 / 不存在时发出，否则主区显示业务状态页的同时还会弹窗（业务不存在时弹窗的关联实例为空）。现有唯一一处是业务选择器的收藏查询 `GET /bizs/<业务ID>/collections/bizs`（后端按地址中的业务鉴权，返回 `2030403`）：照 PR `components/business-selector/global.vue`，当前业务不在 `businessGlobalStore.businessAuthorizedList` 中时跳过查询。`getBizListWithAuth` 不带业务 ID，不受影响。

## 10. 删除旧整页无权限与旧全局权限弹窗（`src/app.vue` 等）

**文件**：`src/app.vue`、`src/views/business/host/children/host-operations/use-single-operation.tsx`、`src/plugin-handler/business-host-manage.tsx`、`src/language/lang.ts`、`src/views/resource/NoPermission.tsx`、`src/store/usePagePermissionStore.ts`、`src/components/global-permission-dialog/index.tsx`、`src/store/useGlobalPermissionDialog.ts`

- `app.vue` 对齐 PR：删除 `<NoPermission>`、`<GlobalPermissionDialog />`、两者的引入与 `usePagePermissionStore` 解构；`<template v-else>` 拆掉，`bk-navigation` 直接平铺。全仓无 `setHasPagePermission` 调用，`hasPagePermission` 恒为 `true`，整页无权限分支现网从不渲染；页面无权限由 §5 的框架内无权限页承接。
- `use-single-operation.tsx` 照搬 PR diff：删除「预鉴权」分支（`operationMap[type]` 的 `authId` / `actionName` + `authVerifyData` 判断、`handleAuth` + `globalPermissionDialog.setShow(true)`）及 `useVerify`、`useGlobalPermissionDialog` 引入。返回值里的 `noPermission` 字段无人消费。行为变化：无业务资源操作权限时按钮不再前置拦截，点击后接口返回 403，由 §9 的统一权限申请弹窗处理。`operationMap` 的 `authId` / `actionName` 字段 PR 未动，本单也不动。
- `plugin-handler/business-host-manage.tsx`（`use-single-operation` 的唯一消费方，读返回值的 `noPermission` 给按钮置灰）照搬 PR 同一处改动：回收按钮包 `HcmAuth`（业务下 `AUTH_BIZ_DELETE_IAAS_RESOURCE`，资源下 `AUTH_DELETE_IAAS_RESOURCE`，经 `getAuthSignByBusinessId`），更多操作下拉包 `HcmAuth`（`*_UPDATE_IAAS_RESOURCE`），无权限时置灰、点击由 `HcmAuth` 弹申请；下拉项只按 `disabled` 置灰。HEAD 已有 `components/auth/auth.vue`、`getAuthSignByBusinessId`、`useWhereAmI().getBizsId`，无需新增依赖。这是本单唯一一处按钮级迁移，原因是删预鉴权后不迁会丢掉置灰提示。
- 改完后以下文件无引用，删除（PR 只标 `@deprecated`，用户确认直接删）：`views/resource/NoPermission.tsx`、`store/usePagePermissionStore.ts`（① 的 `header.vue` 已按 PR 不再用其 `logout`）、`components/global-permission-dialog/index.tsx`、`store/useGlobalPermissionDialog.ts`。删前 `rg "NoPermission'|usePagePermissionStore|useGlobalPermissionDialog|global-permission-dialog" src` 只剩这四个文件自身。

## 实施记录（2026-10-09）

- 守卫的「首次导航」改用模块级标志判定（原为 `from.path === '/'`），旧按钮预取只在首次导航执行一次；旧 403 页兼容逻辑放在守卫最前、先于业务拦截与视图权限。
- `http/index.ts` 抽 `showPermissionDialog(permission, message)`：有 `permission` 且 `window.hcmPermissionDialog` 已注册才弹窗，否则 `Message` 提示；HTTP 403 与 `2030403` 共用。
- `lang.ts` 为状态页新增的短文案补英文（标题、提示、按钮、业务状态三种标题与描述）；沿用旧 403 页的长段说明不补译，与现网一致。
- `views/status/permission.vue` 的提示分三种：不可自助申请 → 联系管理员；可申请且有 `permissionData` → 提示点「去申请」；无 `permissionData` → 只说明无权限（不提按钮）。
- 验证：`bkdevbuddy_lint` 通过；`npm run build` 通过；`tsc --noEmit` 与 HEAD 对比 `src` 下无新增错误（`.vue` 不在 tsc 范围，由构建覆盖）。

## 从 PR #1897 取 / 不取

| 内容 | 取 | 说明 |
|------|----|------|
| `store/auth.ts` 视图权限方法、`extractSignPermission`、`IPermission` | 取 | 键类型放宽为 `RouteRecordName`；失败不缓存 |
| `constants/view-auth.ts` 结构 | 取 | 加 `hideMenu`；条目按现网拦截重列 |
| 守卫结构（业务拦截 → available → auth.view → try/catch） | 取 | 加旧按钮预取与旧 403 兼容 |
| `business-interceptor.ts` 的 bizId 路由、默认业务回填、写 localStorage | 不取 | 保留 `?bizs=` 与 BusinessSelector |
| `injectStatusComponents` / named view | 取 | 跳过无组件节点；业务页按视图集合选 |
| `views/status/*` | 取 | permission 补一级账号规则与文案；business 的 noBiz 补申请 |
| `http/index.ts` 403 / 2030403 弹窗 | 取 | 补未挂载兜底 |
| `menu.vue` 按 symbol 过滤 | 改写 | 仍是 Symbol id，但只过滤 `hideMenu` 项，严格判定 |
| `Meta` 构造函数（面包屑默认显示等） | 不取 | 改变现网 |
| 资源接入 / 资源侧操作记录 / 云账单管理的页面拦截 | 不取 | 现网没有（R-004） |
| 各页面按钮 `useVerify` → `HcmAuth` 迁移（同一 commit 的大部分文件） | 不取（主机操作列例外，见 §10） | 按钮级不在本单 |
| `AUTH_UPDATE_ACCOUNT` 改动、`AUTH_DELETE_ACCOUNT` | 不取 | 属账号管理按钮 |

## HEAD 新增能力必须保留

以 `git log --oneline 8fea9924df..HEAD -- <file>` 核对过：

- `common/auth-service.ts`、`constants/auth-symbols.ts`：权限策略库 8 个、权限模板 3 个、云密钥 3 个、二级 / 三级子账号 5 个 symbol 与定义（#1932 #1934 #1943 #1954 #1967 #1983 #2002）；`AUTH_FIND_PERMISSION_POLICY_LIBRARY` 的 id `cloud_vendor_config`（#1934 修正）。本单只追加，不重排、不改值。
- `store/common.ts`：一期给 `/bill/root-account`、`/bill/main-account` 加的拦截（#2085）→ 迁为 §7 两行；权限策略库条目（#1932）→ 同上。条目本身保留供按钮预取。
- `views/error-pages/403.tsx`：一期 `NO_SELF_APPLY_KEYS` 与一级账号文案（#2085）→ 403 页原样保留，规则移植到 `permission.vue`。
- `views/bill/route-config.ts`：一期 `account-manage` 按权限分流（#2085）→ 保留行为，改读新 store。
- `store/index.ts`：二级 / 三级账号、云密钥 store 导出（HEAD 新增）→ 只改 `preload`，导出不动。
- 待核实：`auth.md` 描述的分配（`resource_assign`）双实例预鉴权在 HEAD `auth-service.ts` 未见对应实现（`AUTH_ASSIGN_IAAS_RESOURCE` 仍是旧定义），本单不碰这两条定义。

## 风险与验证

| 风险 | 怎么验 |
|------|--------|
| 漏迁 = 放开拦截（AC-S01） | 用无对应权限的账号逐个打开 §7 前 8 行地址（直接输地址 + 站内点击各一次），全部为框架内无权限页、地址不变；有全部权限的账号逐个打开均正常（AC-009） |
| 多拦（R-004） | 无云账单权限直接访问账单管理可打开；资源接入各页、资源侧操作记录、方案详情、一级 / 二级账号创建页不出现无权限页 |
| 任务管理占位 `auth.view` 抛错 | 业务资源 → 任务管理正常打开 |
| 无组件分组节点被注入导致白屏 | 业务资源每个菜单、服务请求两项逐一打开 |
| `/resource` 裸地址命中账号列表的 `alias: ''`，新机制按记录拦截会多拦一个地址（旧机制按 `to.path` 不拦） | 已核实：vue-router 同路径时子记录优先，`/#/resource` 落到账号列表（与 HEAD 一致，① 的父路由 redirect 不生效；站内顶栏按 name 跳，不经过此地址）。处理见「已确认」 |
| 业务无权限 | 无任何业务权限 → noBiz 页可申请；`?bizs=` 改成无权限业务 → bizUnauthed 页，业务选择器可切走（AC-006）；改成不存在的 id → bizNotFound |
| 一级账号不可自助申请（AC-003） | 无一级账号权限：侧栏无一级账号；直接访问显示「请联系管理员授权」+ 权限名，无申请按钮；二级账号菜单仍在（AC-005） |
| `extractSignPermission` 在 sign 不带 relation、后端又返回了实例时把 instances 裁成 `[[]]`（PR 的 `.filter(Boolean)` 不过滤空数组），申请链接可能异常 | 照搬 PR，自测核对：二级账号、权限策略库、方案推荐、noBiz 的「去申请」均能打开权限中心且权限项正确；异常再改为过滤空实例组 |
| 权限接口失败 | 断开 `/auth/verify`（DevTools 屏蔽）刷新：受控菜单不显示、其它正常；进受控页显示加载失败可刷新，不白屏 |
| 接口兜底（AC-007） | 在有页面权限、无操作权限的页面触发一个未预鉴权的操作：弹申请弹窗、列出缺失项；连发多个只一个弹窗；后端未带 `permission` 时仍是原提示 |
| 旧 `/403/:id` | 打开 `/#/403/account_find` 仍是旧页、可申请；刷新时有业务权限跳首页（与现网一致） |
| 旧按钮状态回归 | 主机详情、回收站、方案列表、二级账号详情的按钮禁用态与改前一致 |
| 预鉴权合并（AC-P01） | 刷新后 Network 中启动阶段视图权限只有一次 `/auth/verify`（另有旧按钮预取一次，现网即有） |

| Symbol 键串用 | 预鉴权项的菜单在侧栏的显隐与直接访问的拦截各自正确；全仓检索本单新增代码无 `String(to.name)`、`Symbol.for('menu_` 比较 |

## 已确认（2026-10-08）

1. **指定业务无权限 / 业务不存在**：按 F-006 与 PR #1897 处理——`?bizs=` 指向无权限业务时，在框架内显示业务无权限页（`bizUnauthed`，可申请）；指向不存在的业务时显示「业务不存在」（`bizNotFound`）；没有任何业务权限时显示 `noBiz`（可申请）。业务选择器保留，可直接切走。此项视为 F-006 明确要求，不按 R-004 的「新增拦截」处理。
2. **无权限页按钮**：沿用 PR——先显示「去申请」，点过后换成「已申请，刷新页面」；需求 AC-001 已按此更新。
3. **菜单 id 为 Symbol**：从 ① 起 `menu-symbol.ts` 改为 PR #1897 的 Symbol 常量，本单的 `viewAuthConfig`、`viewPermissions`、`hideMenu`、`menu.vue` 过滤、`Meta.auth.view` 均按 Symbol 设计（见「跨单约定」与 §2、§8）。
4. **首批 7 张子单**：「资源详情页迁移」为 ⑥，用本单 §8 的接口给自己的路由声明查看权限；「资源运营重组」顺延为 ⑦。

## 已确认（2026-10-09）

1. **裸地址 `/#/resource` 多拦**：接受。该地址经 `alias: ''` 落到账号列表，新机制对无 `account_find` 的用户显示无权限页（现网按 `to.path` 不拦）。只有手输 / 旧书签会进入，站内顶栏按 name 跳转；⑤ 账号管理迁移后旧路由退场。守卫不为别名加特例。
