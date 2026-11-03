# Coding — feat-menu-phase2-resource-detail

需求文档：仓库根 `docs/reqs/二期资源详情迁移.md`（TAPD [#1069995598138702819](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702819)，验收以 AC-001～AC-009、AC-P01、AC-S01 为准）。设计参照 PR #1897（base `8fea9924df`，head `df8865b539`；详情相关取 `5c513eb559` 与 `c573e933b5` 中的详情部分）。基线 `01bca385c`。

**目标**：资源详情挂到资源纳管下 `/resource/manage/<resourceType>/details/<id>`（`MENU_RESOURCE_DETAIL`），在 ③ 的 `resource-entry.vue` 内渲染（云账号树 + 右栏面包屑：标题 + 返回 + 操作按钮）；详情内容与操作不变，资源侧按钮改用 `hcm-auth`（relation `[account_id]`）；旧地址 `/resource/detail/:type` 转址；改写全站进入资源详情的链接。**本单即上线**（用户 2026-10-10 决定：每单上线即用新地址）：详情子路由随 ③ 的模块生效，撤掉旧详情路由 `resourceDetail`，链接直接改到新详情；详情内进入账号详情的链接已由 ⑤ 改写，不在本单。

## 跨单约定

| 项 | 约定 |
|----|------|
| 提交位置 | 首批 7 个 commit：① 布局 → ② 视图权限 → ③ 资源纳管 + 资源运营翻转 → ④ 回收管理 / 操作记录 → ⑤ 账号管理 → **⑥ 本单** → ⑦ 账单迁址与清理 |
| 菜单符号 | ① 起 `src/constants/menu-symbol.ts` 为 PR 的 `Symbol`。本单新增 `MENU_RESOURCE_DETAIL = Symbol('menu_resource_detail')`（PR 同值）；name 比较一律直接用常量，不 `String()` / `toString()`。① 移植的 `history-storage.ts` 按 menu-symbol 导出建 `description → symbol` 注册表，新常量自动生效 |
| ③ 提供 | `views/resource-manage/route-config.ts` 的 `MENU_RESOURCE_MANAGE`（`manage`，`resource-entry.vue`）与子路由 `:resourceType`（`MENU_RESOURCE_RESOURCE_LIST`）；入口对非列表子路由渲染面包屑（`!isListPage`）；云账号树在详情页切账号时回 `{ name: LIST, params: { resourceType: route.params.resourceType }, query: { accountId } }`；③ 已注册模块并把 `/resource` 翻成资源运营根，旧详情 `resourceDetail` 暂用 ③ 的旧资源接入过渡布局（`legacyBreadcrumbRoutes`、`legacyBackConfig`） |
| ⑤ 提供 | `MENU_RESOURCE_ACCOUNT_DETAIL`（`account/details/:accountId` → `/basic`）；详情页里的账号链接、`useRouteLinkBtn` 账号分支已由 ⑤ 改到新账号详情 |
| 链接改法 | 新详情路由本单即生效，链接直接改到最终目标。各调用方按 PR 内联路由对象（§2）；原「⑥ 返回旧目标、⑦ 翻转函数体」的两步方案取消 |
| 资源 / 业务侧判定 | 详情组件两侧都用 `hcm-auth`（2026-10-12 用户决定业务侧一并按 PR 迁移）；签名按 `route.name === MENU_RESOURCE_DETAIL` 选资源侧（`AUTH_*` + 账号）或业务侧（`AUTH_BIZ_*` + 业务）。不照 PR 用「有无 bizId」判定：`getBizsId()` 会回退到 localStorage，资源侧也可能取到业务 ID |
| 撤旧与过渡布局 | 同一 commit：`router/module/resource.ts` 删除旧详情记录（`/resource/detail/:type`，`resourceDetail`）；`router/index.ts` 在 catch-all 前注册 `resourceDetailLegacy`；`resource-detail.vue` 删 `useVerify`、`permission-dialog`、`bus.$on('auth')`、`provide('authVerifyData')`（旧路由撤掉后无消费方，对齐 PR）。④ 已把旧操作记录详情移出，本单再去掉 `resourceDetail` 后 `main-content.vue` 的右栏过渡面包屑无使用方：删除 `legacyBreadcrumbRoutes` / `showLegacyBreadcrumb` / `legacyBackConfig`、`.legacy-breadcrumb-view` 模板分支与样式（旧云账号树容器仍在，⑦ 删） |
| ② 关系 | `useVerify` 在 HEAD 可用（全局守卫预取 `pageAuthData`），② 保留这条预取供按钮级旧消费方使用；本单后资源详情组件与 `business-detail.vue` 都不再使用（方案、账单等其它页面仍在用）。`.cursor/rules/fe-auth-migration.mdc` 说它「已失效」描述的是 PR 状态，与 HEAD 不符。详情路由不声明 `auth.view`（② §8：现网详情无页面级拦截） |
| 终扫 | 本单提交前 `rg "'resourceDetail'" src` 为 0：旧路由撤掉后，任何按该 name 的跳转（含 ⑦ 才删的旧资源接入页内部）都会报 No match，必须本单改 |

## 执行顺序

1. `menu-symbol.ts` 加 `MENU_RESOURCE_DETAIL`（§1）
2. 详情地址类型与链接写法（§2，各调用方内联）
3. 详情路由组合式函数 `use-resource-detail-route.ts`（§3）；§4 已由 ① 修订完成
4. 共享详情组件：取 ID、按钮权限（§5，依赖 3）
5. `resource-detail.vue`（§6）
6. `route-config.ts` 追加详情子路由 + `resourceDetailLegacy`（§7，依赖 1、2）；③ 列表补 gcp/template 回退（§8）
7. 站内链接改走链接工具（§9，依赖 2）
8. 撤旧详情路由、注册转址、删 ③ 的右栏过渡面包屑（跨单约定「撤旧与过渡布局」，依赖 7：链接改完旧 name 才无引用）
9. `bkdevbuddy_lint` + `npm run build` + `tsc` 对比；浏览器跑 AC；业务详情回归

> 2 先于 4、7：组件与链接都依赖映射与工具；6 依赖 1 的符号与 2 的映射。

## 改动点

### 1. 菜单符号
**文件**：`src/constants/menu-symbol.ts`

资源纳管段追加 `export const MENU_RESOURCE_DETAIL = Symbol('menu_resource_detail');`。

### 2. 详情地址的类型与链接写法（照 PR，2026-10-12 第 11 轮修订）

不建链接工具与类型映射文件，各调用方内联路由对象（PR 写法）：

- 资源侧：`{ name: MENU_RESOURCE_DETAIL, params: { resourceType: <调用方类型>, id }, query: { accountId: route.query.accountId, vendor } }`；业务侧：`{ name: MENU_BUSINESS_*_DETAILS, params: { id }, query: { [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY], vendor } }`（多类型的调用方在本文件内联 `BUSINESS_DETAIL_ROUTE_MAP`）；统一 `routerAction.redirect(routeInfo, { history: true })`，分支按 `whereAmI.value === Senarios.business`。href 类直接拼字符串（`/#/resource/manage/<类型>/details/<id>?vendor=`、`/#/business/<类型>/detail/<id>?vendor=&bizs=`）。
- 列表页签与详情地址用同一套类型名（第 12 轮，用户决定）：弹性 IP 统一为 `eip`、路由表统一为 `routing`（`RESOURCE_TYPES`、列表页签、详情 `componentMap`、跳转类型都改）；GCP 防火墙、安全组模板沿用 PR 的 `gcp`、`template`。旧类型名只做兼容：`resource-detail.vue` 的 `resourceTypeToComponentKey`（`ip` / `eips → eip`、`route → routing`）、列表的 `DETAIL_TYPE_TO_TAB`、旧详情转址里的类型映射。
- 云厂商 query 键统一 `vendor`（照 PR）；详情组件读 `route.query.vendor || detail.vendor`（主机、安全组），镜像读 `route.query.vendor`，`security-rule.vue` 改用 `props.vendor`，子网详情的 GCP 判断改为读详情返回的 `vendor`（照 PR）。
- 与 PR 的差异只有一处：业务侧跳转带 `?bizs=`。PR 的业务 ID 在路径里（`/business/:bizId`，属收尾单）；现网在 query，且业务选择器只在挂载时补参数，按 name 跳转不带会丢失当前业务。

### 3. 详情路由组合式函数
**文件**：`src/views/resource/resource-manage/hooks/use-resource-detail-route.ts`（新，与 `use-detail.ts` 同目录，只服务这批详情组件）

- `useResourceDetailRoute()`：返回 `isResourceDetail`（`route.name === MENU_RESOURCE_DETAIL`）、`detailId`（`route.params.id`；两侧旧 `?id=` 都在路由层转成路径参数，见「业务侧详情」）。
- `useResourceDetailAuth(accountId)`：返回 `updateSign` / `deleteSign`（资源侧 `AUTH_UPDATE/DELETE_IAAS_RESOURCE` + `[accountId]`，业务侧 `AUTH_BIZ_UPDATE/DELETE_IAAS_RESOURCE` + `[getBizsId()]`）与 `authPending`（关联 ID 未就绪，用作 `hcm-auth` 的 `ignore` 并禁用按钮）。

### 4. 标题与右侧操作（已由 ① 修订完成，本单不处理）

① 修订（2026-10-09）已按 PR 在 11 个详情组件里删掉 `<detail-header>`：标题 `useBreadcrumb().setTitle(...)`（文案与原标题栏逐字一致，业务 / 资源两侧都生效）、右侧操作 `Teleport defer to="#breadcrumbExtra"`；业务详情路由已改为显示全局面包屑并补 `menu.relative`。本单新详情路由只需补 `title` 与 `menu.relative`，并保证 ③ 入口右栏渲染面包屑（含 `#breadcrumbExtra`）。原计划的 `DetailHeader` 自适应、`useDetailTitle` 作废。

### 5. 共享详情组件
**文件**：`src/views/resource/resource-manage/children/detail/{host,vpc,subnet,security,gcp,drive,ip,routing,image,network-interface}-detail.vue`、`src/views/resource/resource-manage/children/detail/template-detail/index.tsx`、`src/views/resource/resource-manage/children/components/vpc/{vpc-route,vpc-subnet}.vue`、`src/views/resource/resource-manage/children/components/host/{host-drive,host-ip,host-security}.vue`、`src/views/resource/resource-manage/children/components/security/{security-rule,security-info}.vue`、`src/views/resource/resource-manage/children/components/security/security-relate/platform.vue`、`src/views/resource/resource-manage/children/components/security/security-relate/data-list/collapse-data-list.vue`

- **取 ID**：全部 `route.query.id` / `route.query?.id`（上列 detail 文件、`template-detail`、`vpc-route`、`vpc-subnet`、`subnet-detail` 的 `countSubnetIps`、`ip-detail` 的 `eip_id`）改为 `detailId`（只读 `params.id`，照 PR）。
- **按钮权限**：按每个按钮现用的 action id 映射——`(biz_)iaas_resource_operate → updateSign`、`(biz_)iaas_resource_delete → deleteSign`（子组件取 `props.data / props.detail` 的 `account_id`），PR 与 HEAD 不一致处以 HEAD 为准（主机回收、子网删除 HEAD 用 operate）。统一写法：`<hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">`，按钮 `:disabled="noPerm || authPending || 原有业务条件"`；删除 `authVerifyData`、`actionName`、`showAuthDialog` / `bus.$emit('auth')`、`hcm-no-permision-*` 类与「当前用户无权限操作该按钮」提示分支（无权限由 `hcm-auth` 置灰加锁、点击弹申请）。
- `business-detail.vue` 照 PR 删 `useVerify`、`permission-dialog`、`bus.$on('auth')`、`provide('authVerifyData')`。业务侧行为保持（AC-009）：无权限时由原「可点、点后弹申请」变为「置灰加锁、点后弹申请」，与资源侧和 PR 一致；子网详情业务侧「有权限时置灰」的原判定照 PR 保留为 `:disabled="!noPerm"`。`host-security.vue` 已有的 `hcm-auth`（按 `isResourcePage` 选资源 / 业务签名）不动，只处理解绑按钮的旧判定。
- 实施落点：表格 render 内的按钮（`host-drive` 卸载、`host-ip` / `host-security` 解绑、`security-rule` 编辑 / 删除）用 `resolveComponent('hcm-auth')` 包裹（PR 写法）；`security-rule` 新增 `accountId` prop，由 `security-detail` 传 `detail.account_id`（PR 同）；`security-info` 的编辑图标（资产归属四个、业务使用范围两个）与「使用范围 · 编辑」按钮包 `hcm-auth`，去掉 `useVerify`；`platform.vue`、`collapse-data-list.vue`（业务侧「当前业务」下）的绑定 / 解绑按钮包 `hcm-auth`，去掉 `useVerify`。
- `drive-detail.vue`、`gcp-detail.vue` 现网鉴权已注释、PR 也未加，保持无鉴权。`info-list.tsx` 的 `authAction` 只有账单页注入，不动。
- `bk-loading` 位置、`.detail-content-wrap` 替换 `.i-detail-tap-wrap` 已由 ① 修订（2026-10-09）随标题栏一起照 PR 取（业务、资源两侧共用）；`detail-tab.scss` 去 `height: 100%` 仍不取，新布局下如有间距问题在 §6 用容器修正。

### 6. 资源详情容器
**文件**：`src/views/resource/resource-manage/resource-detail.vue`

- 取 PR：`resourceType = route.params.resourceType`（旧路由本单撤掉，不再兼容 `params.type`），经 `resourceTypeToComponentKey`（旧名 `ip` / `eips → eip`、`route → routing`）查 `componentMap`。
- 查不到组件（类型非法）→ `bk-exception`（404，「资源类型不存在」）+「返回资源纳管」按钮（`routerAction.redirect({ name: MENU_RESOURCE_MANAGE })`），不白屏（F-001 异常）。
- `<component :is :key="`${resourceType}-${id}`">`：详情间跳转（同一路由记录只换 params）时重建组件，避免组件 setup 时读到的旧 ID。
- 外层不加内边距：① 修订已把详情组件的 `.i-detail-tap-wrap` 换成 `.detail-content-wrap`（自带 24px，固定头部的 52px 偏移随 `DetailHeader` 一起去掉），③ 入口内容区也无内边距；组件里资源侧的行内 `padding: 0` 保留。1366px 下截图核对（**待核实**）。
- 删除 `useVerify` / `permission-dialog` / `bus.$on('auth')` / `provide('authVerifyData')`（见跨单约定「撤旧与过渡布局」）。

### 7. 路由与旧地址转址
**文件**：`src/views/resource-manage/route-config.ts`（③ 产物）

在 `MENU_RESOURCE_MANAGE.children` 中 `:resourceType` 之前追加：

```ts
{ name: MENU_RESOURCE_DETAIL, path: ':resourceType/details/:id',
  component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
  meta: { ...new Meta({ owner: MENU_RESOURCE, activeKey: MENU_RESOURCE_MANAGE, title: '资源详情',
    menu: { i18n: '资源详情', relative: MENU_RESOURCE_RESOURCE_LIST }, layout: { breadcrumb: { show: false } } }) } },
```

- `layout.breadcrumb.show: false`：整页顶部不画；③ 入口按「非列表页」在右栏画。`title` 必填（`Meta` 构造函数用 `title` 覆盖 `menu.i18n`，面包屑标题读它）。
- `menu.relative` 用 `MENU_RESOURCE_RESOURCE_LIST` 而非 PR / ③ 写的 `MENU_RESOURCE_MANAGE`：无 `_f` 时 `useBack` 回退 `{ name: relative }`，vue-router 4.1.3 按 name 解析会沿用当前路由同名的必填参数（已核对 `node_modules/vue-router/dist/vue-router.mjs` 的 `resolve`），于是回到同类型列表；指向 `MANAGE` 会一律落到主机。从列表进入时链接带 `{ history: true }`，返回走 `_f` 历史栈，页签与 `accountId` 都还原（AC-002）。
- `resourceDetailLegacy`（绝对路径、无 name，本单注册）：

```ts
export const resourceDetailLegacy: RouteRecordRaw[] = [{
  path: '/resource/detail/:resourceType',
  redirect: (to) => {
    const { id, type, _f, ...query } = to.query;
    const { resourceType } = to.params;
    if (!id) return { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType }, query: { accountId: query.accountId } }; // 修 PR 的 /details/undefined
    return { name: MENU_RESOURCE_DETAIL, params: { resourceType, id }, query: { ...query, vendor: type } };
  },
}];
```

类型原样带过去（PR 同）；旧地址的 `type` 是云厂商，转成 `vendor`；缺 id 回列表时由 §8 把类型转成页签；类型无法识别时由 `resource-detail.vue` 显示「资源类型不存在」。

### 8. 列表对详情类型的回退
**文件**：`src/views/resource-manage/resource-list.vue`（③ 产物）

③ 的「类型不在 `tabs` 内 → replace 到 `tabs[0]`」之前，先查内联的 `DETAIL_TYPE_TO_TAB`（旧名 `ip` / `eips → eip`、`route → routing`；`gcp` / `template → security` + `scene`），命中则 replace 到对应页签。覆盖旧列表地址（如 `/resource/resource?type=ip`）、从 GCP 防火墙 / 模板详情返回（无 `_f` 时按 `menu.relative` 沿用当前 `resourceType`）、在这两类详情页切账号三条路径。PR 未处理（会落到空页签），用户 2026-10-12 确认补上。

### 9. 站内链接改写
**文件**：见下节清单的每一行

按 §2 内联写法改（不再展开整个 `route.query`，只带账号 / 业务与 `vendor`）；href 类直接拼新地址，`target` 等属性保留（F-005 新窗口行为不变）。

## 业务侧详情（第 10 轮）

对照 #1897 的 `business-detail.vue`、各业务模块 `route-config.ts` 与站内链接，意图是「业务详情与资源详情同一套路由约定」：

| #1897 改动 | 意图 | 本单 |
|---|---|---|
| 详情路由 name 改为 `MENU_BUSINESS_{HOST,DRIVE,VPC,SUBNET,EIP,NIF,ROUTE_TABLE,SECURITY_GROUP,GCP,TEMPLATE,IMAGE}_DETAILS`，路径 `<类型>/detail/:id?`，`beforeEnter: normalizeDetailId` | ID 进路径；旧 `detail?id=` 书签可用 | 取。`router/module/business.ts` 原地改（按模块拆 `route-config.ts` 属收尾单 AC-011）；新增 `router/utils/normalize-detail-id.ts`（照 PR） |
| 站内业务详情跳转改为 `routerAction.redirect({ name: MENU_BUSINESS_*_DETAILS, params: { id } }, { history: true })`，散在各文件的 `BUSINESS_DETAIL_ROUTE_MAP` / `RECYCLEBIN_DETAIL_MAP` | 不再按字符串拼 name / 路径；返回走历史栈 | 取，按 §2 内联（PR 写法）；保留 `?bizs=`，云厂商键统一 `vendor`；不再展开整个 `route.query` |
| `business-detail.vue` 删 `useVerify` 与权限弹窗 | 按钮统一 `hcm-auth` | 已取（第 9 轮） |
| `business-detail.vue` 仍按 `route.path.includes(<类型>)` 取组件 | — | 改按路由 name 取：ID 进路径后，路径子串会被 ID 干扰（如 ID 含 `ip`）；加 `:key="route.params.id"`，同类详情互跳时重建 |
| 详情组件读 `route.params.id` | 同上 | `useResourceDetailRoute().detailId` 改为只读 `params.id`（两侧旧 `?id=` 都在路由层转换） |
| 详情页切换业务回列表（#1897 由新业务选择器承担） | — | 现网 `business-selector.tsx` 按 name 含 `BusinessDetail` 判断、按路径截 `/detail` 回列表，改名后失效且 GCP / 模板截出的路径不存在；改为按文件内联的业务详情 name 列表判断、回 `meta.menu.relative`，只带 `?bizs=` |

改到的入口：`use-columns.tsx`（列表 ID 列）、`security-manage.vue`（安全组 / GCP 防火墙 / 模板 / 生效确认）、`drive-detail.vue`（挂载主机）、`ip-info.vue`、`route-info.vue`、`host-drive.vue`、`useRouteLinkBtn.tsx`、`recyclebin-manager.vue`（业务回收站查看详情）、`load-balancer/.../rs-ip-group.vue`（RS 主机新窗口）、`network-interface-info-gcp.vue`（业务侧 href）。

`business-manage.vue`（业务列表页）的 #1897 改动不属详情，且依赖业务路由重建，本单不跟，整体归「二期收尾体验」（用户 2026-10-12 确认，已写入该单 F-008）：

| 改动 | 依赖 / 归属 |
|---|---|
| 组件按列表路由 name 选取、新建按 `MENU_BUSINESS_APPLY_*` 跳转 | 申请页符号随业务路由按模块拆分定义 → 二期收尾体验（AC-011） |
| 去掉等待 `accountStore.bizs` 的 loading、鉴权改 `getBizsId()` | `/business/:bizId` 下业务 ID 总在地址里 → 二期收尾体验 |
| 去掉 `filter` / `isResourcePage` / `bkBizId` 传参（列表改 `useFilterFromRoute`） | 与业务列表组件改动成对 → 二期收尾体验 |
| 移除内嵌回收站（业务回收站并入共享 `views/recyclebin`，`MENU_BUSINESS_RECYCLEBIN`） | 业务回收站迁移 → 二期收尾体验 |

## 旧地址与站内链接清单

行号为基线 `01bca385c`（rg 结果），实施时以当前文件为准。

| 位置 | 现状 | 改为 |
|------|------|------|
| `src/views/resource/resource-manage/hooks/use-columns.tsx:114` | `'resourceDetail'` + 展开 `route.query` | `/resource/manage/<页签>/details/<id>?type=&accountId=` |
| `src/views/resource/resource-manage/children/manage/security-manage.vue:399` | 安全组，`vendor` | `security/details/<id>?vendor=` |
| `…/children/manage/security-manage.vue:713` | GCP 防火墙 | `gcp/details/<id>` |
| `…/children/manage/security-manage.vue:923` | 安全组模板（id = `cloud_id`） | `template/details/<cloud_id>` |
| `…/children/manage/security-manage.vue:1138` | 生效确认后进规则页，`activeTab: 'rule'` | `security/details/<id>?activeTab=rule&vendor=` |
| `…/children/detail/drive-detail.vue:125` | 挂载主机 | `host/details/<id>` |
| `…/children/components/ip/ip-info.vue:35` | `txtBtn` 主机 / 网络接口 | `host`、`network-interface` |
| `…/children/components/route/route-info.vue:114` | 所属 VPC | `vpc/details/<id>` |
| `…/children/components/host/host-drive.vue:110` | 硬盘 | `drive/details/<id>` |
| `src/hooks/useRouteLinkBtn.tsx:53` | 资源侧 VPC / 子网 / 镜像 | 新详情；第 12 行注释同步改（账号分支已由 ⑤ 改） |
| `src/views/resource/recyclebin-manager/recyclebin-manager.vue:620` | 资源侧「查看详情」（④ 后只剩业务回收站在用，此分支仍需改以满足终扫） | `host` / `drive` 新详情 |
| `src/views/recyclebin/index.vue` `handleClick`（④ 新建，行号**待核实**） | `'resourceDetail'`，`{ history: true }` | 新详情 |
| `src/views/business/load-balancer/group-view/specific-target-group-manager/target-group-detail/index.tsx:62` | `href="/#/resource/detail/vpc?type=tcloud&id=…"`，`target='_blank'` | `/#/resource/manage/vpc/details/<vpc_id>?vendor=<detail.vendor>`（AC-007，PR 同）。① 也改此文件，以 ① 版本为底 |
| `…/children/components/network-interface/network-interface-info-gcp.vue:121` | `/#/resource/detail/host?id=…&type=gcp` | `host/details/<id>?type=gcp` |
| `src/router/module/resource.ts` 旧详情记录 | `/resource/detail/:type`（`resourceDetail`） | 删除；由 `resourceDetailLegacy` 转址 |

实施补充：`host-drive.vue`（硬盘 ID → `drive` 详情）、`recyclebin-manager.vue` 资源侧分支（`router.push`，不入历史栈，与原行为一致）同样改走链接工具。`security-manage.vue` 生效确认后带 `activeTab=rule`，但 `security-detail.vue` 读的是 `query.active`，现网与 PR 都不生效，本单保持原样。

**不属本单**：账号详情链接（8 个文件 + `useRouteLinkBtn` 账号分支，⑤ 已改）；旧资源接入页内部的 `RESOURCE_DETAIL_TABS` 等（随旧页在 ⑦ 删除）。

## 从 PR #1897 取 / 不取

| PR 内容 | 取舍 |
|---|---|
| 子路由 `:resourceType/details/:id`（`MENU_RESOURCE_DETAIL`）挂在 `MENU_RESOURCE_MANAGE` 下 | 取；补 `title`，`relative` 改 `MENU_RESOURCE_RESOURCE_LIST` |
| `resource-detail.vue` 的 `resourceTypeToComponentKey`、去 `useVerify` | 取；补非法类型提示与 `:key` |
| `legacyRedirects` 旧详情转址 | 取（类型原样带过去）并修正：缺 id 回列表、`type` 转 `vendor`、保留其余 query；沿用本分支各单的模块导出写法 `resourceDetailLegacy` |
| 链接改 `MENU_RESOURCE_DETAIL` + `routerAction` + 只带 `accountId`/厂商 | 取，经链接工具集中生成，本单生效 |
| 链接 query 厂商键统一改 `vendor`、详情读 `route.query.vendor \|\| detail.vendor` | 取（第 11 轮） |
| 链接内联、不建工具文件，地址类型沿用调用方类型 | 取（第 11 轮）；列表补详情类型回退（§8） |
| 详情组件只读 `route.params.id` | 取（业务详情路由补 `:id?` + `normalizeDetailId`） |
| 详情组件删 `<detail-header>`、各自 Teleport 到 `#breadcrumbExtra` + `setTitle` | 已由 ① 修订（2026-10-09）照 PR 取，本单不处理 |
| `useVerify → hcm-auth`（资源 / 业务按 `bizId` 选签名） | 取（两侧）；选签名改按路由判定，action 映射以 HEAD 为准 |
| 详情样式调整（`bk-loading` 位置、`detail-tab.scss`、`.detail-content-wrap`） | 不取（共享组件，影响业务）；容器级修正 |
| 业务详情路由 Symbol 名 + `detail/:id?`、站内业务详情跳转 | 取（见「业务侧详情」） |
| `host-detail` 回收后跳 `MENU_BUSINESS_RECYCLEBIN`、业务路由 `/business/:bizId`、`business-manage.vue` | 不取，归二期收尾体验（F-008、F-010） |
| PR 的 `account/vendor-group` 详情页切账号回列表 | ③ 已实现，本单只补 gcp/template 回退（§8） |

## HEAD 新增能力必须保留

自 PR base 以来，本单要改的文件里 HEAD 只有以下差异（`git diff 8fea9924df HEAD` 已核对），**一律在 HEAD 文件上逐处改，不拿 PR 整文件覆盖**：

- `host-detail.vue`：「更多」下拉的 `:disabled` 与 `:popover-options="{ clickContentAutoHide: true }"`（`0a7a62faff`）。
- `host-ip.vue`：绑定弹窗网络接口 `bk-select` 的 `filterable`（`c1938521b9`）。
- `security-manage.vue`：`getMenuList` 异步搜索项（`9c6ac9204f`）；同文件引用的克隆弹窗「选择目标地域」（`c5b61eff78`，`children/dialog/clone-security` 等）不改。
- 主机批量分配（`4348ae1301`，`assign-host/`）、`batch-distribution/` 分配弹窗与 `resource_assign` 确定时预鉴权：本单不碰这些文件；`host-detail` 自带的分配弹窗与请求体保持原样（内容不变，已知请求体过期问题不在本单修）。
- `host-security.vue` 已有的 `hcm-auth`（基线即有）保留。

## 风险与验证

| # | 风险 | 怎么验 |
|---|------|--------|
| 1 | 入口全覆盖 | 资源纳管各列表点资源、详情内互跳、业务目标组 VPC 链接、回收管理查看详情、业务回收站资源侧分支：都进 `/resource/manage/<类型>/details/<id>`，顶栏高亮资源运营、侧栏高亮资源纳管 |
| 2 | 旧 name 与旧代码残留 | `rg "'resourceDetail'" src` 为 0；`resource-detail.vue` 无 `useVerify`；`rg "legacyBreadcrumb\|legacy-breadcrumb" src` 为 0；点任一链接无 No match |
| 3 | 业务详情回归（AC-009） | `/business/{host,vpc,subnet,security,drive,eip,routing,image,network-interface}/detail?id=`、`gcp/detail`、`template/detail`：标题、按钮、内部跳转与改前一致；有 / 无业务 `biz_iaas_resource_*` 权限时按钮可用 / 置灰加锁可申请；面包屑无变化 |
| 4 | 新路由下按钮权限（AC-004、AC-S01） | 有 / 无某账号 `iaas_resource_operate`、`iaas_resource_delete` 的用户逐类详情点按钮，无权限置灰可申请；Network 中鉴权请求合并、`account_id` 返回前无请求 |
| 5 | Teleport 目标缺失导致按钮消失 | 新详情右栏面包屑出现标题、返回、按钮；控制台无 Teleport 警告；`isOtherVendor` 时主机按钮仍隐藏 |
| 6 | 返回与上下文（AC-001、AC-002） | 选中账号后从主机 / 弹性 IP / 路由表 / 安全组 / GCP 防火墙 / 模板列表进详情再返回：页签、`scene`、账号一致；新窗口直接打开详情返回到该类型列表；GCP / 模板返回到安全组页签对应子场景 |
| 7 | 旧地址转址（AC-005、AC-006） | `/resource/detail/vpc?type=tcloud&id=123` → `/resource/manage/vpc/details/123?vendor=tcloud`；`eips → eip`、`route → routing`；缺 id → 对应列表（`gcp` → 安全组子场景）；未知类型 → 「资源类型不存在」 |
| 8 | 间距错位（1366px） | 11 类详情在侧栏展开 / 折叠、1366px 下截图：无 52px 空白、内容不被树遮挡（**待核实** ① 全局样式） |
| 9 | Symbol name 进历史栈 | 依赖 ① 的 `history-storage.ts`；列表→详情→详情→返回两次，逐级回到原页 |
| 10 | 资源不存在 | 非法 id：接口错误由 `http` 全局提示，页面框架与按钮区不白屏（与现网同）；非法类型显示 404 提示（§6） |
| 11 | 性能 AC-P01 | 新详情首屏请求：详情接口 1 次 + 鉴权合并请求；与改前同环境对比打开耗时差异 ≤10% |
| 12 | 构建与类型 | `bkdevbuddy_lint`、`npm run build`、`tsc` 与 HEAD 对比无新增错误（2026-10-12：eslint / stylelint 通过，build 通过，tsc 与 ⑤ 提交后基线一致） |

pre-commit 整文件检查顺带修的原有问题：`security-rule.vue` 两处 `v-model:isShow` 改 `v-model:is-show`；`recyclebin-manager.vue` 样式按 stylelint `--fix` 补空行、`#ffffff` 缩写、去掉与 `@keyframes move` 重复的 `@-webkit-keyframes move`。

## 已确认（2026-10-08）

1. GCP 防火墙规则、安全组模板的详情地址按 PR：`/resource/manage/gcp/details/<id>`、`/resource/manage/template/details/<id>`，返回时回到安全组页签。
2. 详情页面包屑标题沿用现有文案（如「主机详情 - ID x」），不改为资源实例名称。

## 待确认（已在上方确认，以下为原始记录）

1. **GCP 防火墙、安全组模板的地址类型**：沿用 PR 的 `gcp` / `template`（`/resource/manage/gcp/details/<id>`），它们不是资源纳管页签，返回和切账号时映射回「安全组」页签对应子场景（§8）。需求 F-001.2 字面要求「地址类型与页签一致」，如需严格一致，可改为 `security/details/<id>?scene=gcp|template`，代价是组件选择依赖 query。默认按 PR。
2. **面包屑标题文案**：沿用各详情现有标题（如「主机详情 - ID x」「云硬盘：ID（x）」，与 PR 一致），不改为资源实例名称。需求 F-002.2 写「资源名称」，如指实例名需另定文案。
