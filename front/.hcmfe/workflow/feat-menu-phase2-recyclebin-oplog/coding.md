# Coding — feat-menu-phase2-recyclebin-oplog

需求文档：仓库根 `docs/reqs/二期回收与操作记录.md`（TAPD [#1069995598138702618](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702618)，验收以其 AC-001 ~ AC-010、AC-S01 为准）。设计参照 PR #1897（base `8fea9924df`，head `df8865b539`），基线 `01bca385c`。

**首批 commit 顺序**：① 布局 → ② 视图权限 → ③ 资源纳管 + 资源运营翻转 → ④ 本单 → ⑤ 账号管理 → ⑥ 资源详情页迁移 → ⑦ 账单迁址与清理。

**目标**：交付资源侧「回收管理」「操作记录」两个模块并**本单即上线**（用户 2026-10-10 决定：每单上线即用新地址）：路由、查看权限、旧地址转址，以及无账号树的回收管理页面（页内选账号、按账号设置回收保留时间）。新路由挂到 ③ 的资源运营根下，`/resource/recyclebin`、`/resource/record`、`/resource/record/details` 即为新页面；旧页签地址 `/resource/resource/{record,recycle}` 转址到新页面；「回收管理」「操作记录」加入资源运营侧栏（③ 只放了已有最终页面的项）；指向旧回收站 / 操作记录的站内链接本单改。业务侧回收站和操作记录保持原样。

## 跨单约定

| 项 | 约定 |
|----|------|
| 菜单符号 | 从 ① 起，`src/constants/menu-symbol.ts` 已整体改为 PR #1897 的 `Symbol` 常量。本单新增 `export const MENU_RESOURCE_RECYCLEBIN = Symbol('menu_resource_recyclebin')`（名称和描述与 PR 一致）。操作记录沿用 ① 转换后的 `MENU_RESOURCE_OPERATION_LOG(_DETAILS)`。业务侧回收站符号不动。本单所有路由的 `name`、`owner`、`activeKey`、`menu.relative` 都直接引用这些 Symbol，不写字符串 |
| 导出 · 回收 | `views/recyclebin/route-config.ts` 导出 `recyclebinRsc`（相对路径 `recyclebin`）和 `recyclebinRscLegacy`（绝对路径转址）。**本批不导出 `recyclebinBiz`**：业务回收站继续用 `router/module/business.ts`，组件仍是 `views/resource/recyclebin-manager/recyclebin-manager.vue` |
| 导出 · 操作记录 | `views/operation-log/route-config.ts` 中，`operationLogRsc` 改为最终形态（相对路径 `record`、`record/details`）；新增 `operationLogRscLegacy` 用于转址。`operationLogBiz` 不动 |
| 本单注册 | `views/index.ts` 的 `resourceViews` 加 `...recyclebinRsc, ...operationLogRsc`（③ 的资源运营根 children），得到 `/resource/recyclebin`、`/resource/record`、`/resource/record/details`；`router/index.ts` 在 catch-all 前注册 `...recyclebinRscLegacy, ...operationLogRscLegacy`。同一 commit 从 `router/module/resource.ts` 撤掉：旧回收站记录（`t('回收站')`，`/resource/recyclebin`，与新页同路径）、旧 `/resource/resource` 下的 `record` / `recycle` 子路由、对 `operationLogRsc` 的引用（含旧操作记录详情） |
| 加入菜单 | `common/menu-service.ts` 资源运营菜单树：「云资源」组末项加 `{ id: MENU_RESOURCE_RECYCLEBIN, i18n: '回收管理', icon: 'bkhcm-icon-recyclebin', group: '云资源' }`（⑤ 之后账号管理插在它前面）；末尾加不分组的 `{ id: MENU_RESOURCE_OPERATION_LOG, i18n: '操作记录', icon: 'bkhcm-icon-operation-record' }`；`route` 均 `getMenuRoute(resourceViews, id)`，并删掉 ③ 留的 TODO 中对应两项。菜单显隐读 ② 的预鉴权结果 |
| ③ 过渡布局 | 从 `main-content.vue` 的 `legacyBreadcrumbRoutes` 去掉 `MENU_RESOURCE_OPERATION_LOG_DETAILS`：新详情不在旧资源接入外壳下（无 `legacyResourceAccess`），面包屑由 `main-content` 正常渲染 |
| ② 查看权限 | 本单只声明：`meta.auth.view` 为 `{ type: AUTH_FIND_RECYCLE_BIN }` 和 `{ type: AUTH_BIZ_FIND_AUDIT }`，这两个符号 HEAD `constants/auth-symbols.ts` 已有；另在 ② 的 `constants/view-auth.ts` 预鉴权表里加两项 `{ viewId, authSign, hideMenu: true }`（无权限隐藏菜单，F-005）。无权限页由 ② 渲染。② 已把 `Auth.view` 改为 `IAuthSign`，接受 symbol（已核实） |
| ② 拦截 | ② 已把旧 `/resource/recyclebin` 的拦截迁到路由 `auth.view`，本单新回收站路由沿用（同一权限）。操作记录旧页签现网无页面级拦截，本单起按 PR 加 `AUTH_BIZ_FIND_AUDIT`（PR 目标状态；无审计权限的用户此前能进页签但列表无数据，现在显示无权限页） |
| 交给 ⑥ 的链接 | 回收管理新页的「查看详情」（点主机/硬盘 ID）在本 commit 仍跳 `name: 'resourceDetail'`（`/resource/detail/:type`，`query: { id, type: vendor }`），与现网相同；由 ⑥ 统一把站内所有进入资源详情的链接改到新详情路由，其中包括 `views/recyclebin/index.vue` 的 `handleClick` |
| 本单改的链接 | `children/detail/drive-detail.vue` 回收硬盘后跳 `/resource/resource/recycle` → `{ name: MENU_RESOURCE_RECYCLEBIN, query: { type: 'disk' } }`。旧资源接入页的 `RESOURCE_TABS`、`useResourceAccount.tsx` 的 `WHITE_LIST` 不改：旧页签点击经本单转址进入新页面，随旧页在 ⑦ 删除 |

## 执行顺序

1. `menu-symbol.ts` 加符号
2. `views/operation-log/route-config.ts`：`operationLogRsc` 改为最终形态，加转址数组
3. 新建 `views/recyclebin/index.vue`、`views/recyclebin/route-config.ts`
4. 在 ② 的 `constants/view-auth.ts` 加两项预鉴权
5. 注册与撤旧（见「本单注册」）、加入菜单、调整 ③ 过渡布局、改 `drive-detail.vue` 链接
6. `bkdevbuddy_lint` + `npm run build` + `tsc` 对比；浏览器冒烟

> 5 依赖 2、3：撤掉旧记录前新路由必须就位，否则同 Symbol name 的操作记录会短暂无路由。

## 改动点

### 1. 菜单符号
**文件**：`src/constants/menu-symbol.ts`
在 ① 转换后的「资源」段、`MENU_RESOURCE_OPERATION_LOG` 旁加 `export const MENU_RESOURCE_RECYCLEBIN = Symbol('menu_resource_recyclebin');`。如果 ① 已经按 PR 整体搬入并带上了这一行，本步跳过，只核对名称。

### 2. 回收管理路由
**文件**：`src/views/recyclebin/route-config.ts`（新）

- `recyclebinRsc`：`{ name: MENU_RESOURCE_RECYCLEBIN, path: 'recyclebin', component: () => import('@/views/recyclebin/index.vue'), meta: { ...new Meta({ owner: MENU_RESOURCE, title: '回收管理', activeKey: MENU_RESOURCE_RECYCLEBIN, auth: { view: { type: AUTH_FIND_RECYCLE_BIN } }, menu: { i18n: '回收管理' } }) } }`。保留 `title` 是因为 HEAD 的 `use-breadcrumb` 读 `meta.title`；面包屑沿用 `Meta` 默认值（不显示），与原页签一致。
- `recyclebinRscLegacy`：`[{ path: '/resource/resource/recycle', redirect: { name: MENU_RESOURCE_RECYCLEBIN } }]`。目标固定，用对象形式的 `redirect`，vue-router 默认保留 query，所以 `accountId`、`type`、`cvm` 会原样带过去，`accountId` 就是初始账号（AC-008）。`/resource/recyclebin` 本身就是新地址，不需要转址。

### 3. 操作记录路由
**文件**：`src/views/operation-log/route-config.ts`、`src/router/module/resource.ts`

- `resource.ts` 不再引用 `operationLogRsc`（旧页签的 `record` 子路由与旧详情一并撤掉）；新旧两组路由 name 是同一个 Symbol，撤旧与注册新在同一 commit，不会同时存在。原计划的 `operationLogRscTab`（为不给旧页签新增拦截而保留旧形态）随「本单即上线」取消。
- `operationLogRsc`（最终形态）：
  - `[0]` 列表：保留 `name`、`path`、`component`；meta 为 `owner: MENU_RESOURCE`、`title: '操作记录'`、`activeKey: MENU_RESOURCE_OPERATION_LOG`、`auth: { view: { type: AUTH_BIZ_FIND_AUDIT } }`、`menu: { i18n: '操作记录' }`。
  - `[1]` 详情：meta 为 `owner: MENU_RESOURCE`、`title: '操作记录详情'`、`activeKey: MENU_RESOURCE_OPERATION_LOG`、`menu: { i18n: '操作记录详情', relative: MENU_RESOURCE_OPERATION_LOG }`，**不配** `layout.breadcrumb.show: false`：`flow-task` 的 `DetailHeader` 已由 ① 修订（2026-10-09）去掉，标题与返回改由全局面包屑承担（`setTitle` 在页面内）。去掉 `notMenu`。按 PR，详情页不加 `auth.view`。
- `operationLogRscLegacy`：`[{ path: '/resource/resource/record', redirect: toRecordList }]`。`toRecordList(to)` 返回 `{ name: MENU_RESOURCE_OPERATION_LOG, query }`：去掉 `accountId`，其余 query 原样保留。如果 `accountId` 是非空字符串，就生成 `filter = qs.stringify({ account_id, created_at: getDateRange('last30d') }, { arrayFormat: 'comma', encode: false })`，序列化选项与 `hooks/use-search-qs.ts` 的 `set` 相同。这样做的原因：`entry-rsc.vue` 一旦读到地址上的 `filter`，就不再套「近 30 天」默认条件，所以要一并写入。账号条件只在转址这一层生成，`entry-rsc.vue` 不读 `accountId`，旧页签行为保持不变。`/resource/record/details` 地址不变，不需要转址。
- 取 PR 给 `entry-rsc.vue` 的 `.entry-rsc { height: 100%; padding: 24px; }`（与业务侧 `entry-biz.vue` 一致）：① 的 `view-warp` 无内边距，旧页签本单起转址到新页面，不再嵌在原资源接入页里。

### 4. 回收管理页面
**文件**：`src/views/recyclebin/index.vue`（新）

以 PR 的 `views/recyclebin/index.vue` 为底，**只服务资源侧**。按 `fe-conventions` 改写为 `<script setup lang="ts">`，顺序为 script → template → style。业务逻辑与 PR 一致：主机/硬盘页签、按 ID 搜索、所属主机跳转、恢复/销毁及其弹窗与提示、`recycle_records/list` 分页排序、过期时间相对时长。

- **账号来源（补 PR 的缺口）**：PR 只从 `route.query.accountId` 读账号。本页把 `accountId` 作为唯一事实来源：`computed` 只认非空字符串，选择器变化时调 `routerAction.redirect({ name: MENU_RESOURCE_RECYCLEBIN, query: { ...route.query, accountId: val || undefined } }, { replace: true })`。
- **账号选择器**：在页签 `#setting` 插槽放 `bk-select`（`filterable`、`clearable`，占位「全部账号」）。选项取 `useAccountSelectorStore().getAuthorizedResourceAccountList({ filter: type = resource })`，与原左侧账号树 `account/vendor-group/index.vue` 同源，用户看到的账号范围不变。不复用 `AccountSelector`（`index-new.vue`）：它把 `clearable` 写死为 `false`，选了账号就回不到「全部账号」；它的默认数据源也和账号树不同。
- **列表条件**：监听 `accountId`（`immediate`），维护 `account_id` 规则。删除规则用 `splice(idx, 1)`，修掉 HEAD 和 PR 中 `splice(idx)` 连带删除后续搜索规则的问题。没有 `accountId` 时不带账号条件，即「全部可见账号」，由后端按权限过滤（F-002.3）。
- **回收保留时间（F-003）**：有 `accountId` 时调 `accountStore.getAccountDetail(id)`（默认不走缓存，保证重新进入后显示新值，AC-003），写入 `savedReserveTime`。请求返回时 id 已变化就丢弃结果。详情请求失败视为参数无效：清空 `accountId`，回到默认列表（F-006）。`#setting` 插槽显示「回收保留时间：N天」和原齿轮按钮，文案复用 `RESERVE_TIME_SET` 的 label。打开弹窗时下拉框初始化为 `savedReserveTime`。确定时 `http.patch('/api/v1/cloud/accounts/${id}', { account_id, recycle_reserve_time })`，不拼 `BK_HCM_AJAX_URL_PREFIX`，因为 `http` 的 `baseURL` 已经是它。成功：更新 `savedReserveTime`、提示「配置成功」、关闭弹窗。失败：错误提示由 http 拦截器弹出，`savedReserveTime` 保持修改前的值，弹窗不关闭。没有账号时整块不渲染（AC-004）。`recycle_reserve_time` 为 0 时也是合法值，不能用真值判断（PR 的 `if (account?.recycle_reserve_time)` 会把 0 跳过）。
- **操作权限（AC-S01）**：批量和行内的恢复/销毁都用 `hcm-auth :sign="{ type: AUTH_MANAGE_RECYCLE_BIN }"`（PR 做法，HEAD 已有组件和符号），替换 `useVerify` + 旧的 `permission-dialog`。
- **去掉业务分支**：不用 `useWhereAmI`、`accountStore.bizs`。行禁用条件为 `recycle_type === 'related' || bk_biz_id !== -1`，与原页签在资源侧的判定相同；操作列始终显示。
- **跳转**：「查看详情」保持现网目标：`routerAction.redirect({ name: 'resourceDetail', params: { type: 'host' | 'drive' }, query: { id, type: vendor } }, { history: true })`，改写由 ⑥ 负责（见「交给 ⑥ 的链接」）。切页签用 `routerAction.redirect({ name: MENU_RESOURCE_RECYCLEBIN, query }, { replace: true })`。替换原来的 `router.push` 和写死的 `/resource/resource/recycle`。
- 删除模板没有引用的死代码：`handleShowDialog`、资源详情弹窗及其 `HostInfo`、`HostDrive` 引用，以及 `handleJump`；PR 未接到界面的 `isAccurate` 一并去掉。
- **页签与所属主机跳转**：`selectedType` 跟随 `route.query.type`；手动切页签时 replace 地址（带 `type`、去掉 `cvm`）并清空搜索。「所属主机」改为 `routerAction.redirect` 写入 `type=cvm&cvm=<id>`，由 `route.query.cvm` 的监听回填搜索（PR 直接改 `route.query.cvm` 对象，不可靠）。
- **布局**：根节点用 `page-container`（`main-content` 统一给 24px、`height: 100%`），页签区撑满高度，替代 PR 的 `calc(100vh - 200px)`。`bk-tab` 传 `:label-height="42"`：默认 `labelHeight` 50 以内联 `line-height` 加在页签头上，`#setting` 里的文字与下拉继承后把页签头撑到 50px，与 42px 的页签按钮之间空出 8px（PR 与原回收站页签都有此问题；PR 资源列表用 `line-height: normal !important` 处理同类问题）。
- **首次请求**：`useQueryCommonList` 只深度监听 `filter`、不立即执行，须在它之后注册各 `immediate` 监听（`res_type` 规则每次替换对象），保证首屏触发一次请求。

### 5. 预鉴权
**文件**：`src/constants/view-auth.ts`（② 新建）
在 `viewAuthConfig` 中加 `{ viewId: MENU_RESOURCE_RECYCLEBIN, authSign: { type: AUTH_FIND_RECYCLE_BIN }, hideMenu: true }` 和 `{ viewId: MENU_RESOURCE_OPERATION_LOG, authSign: { type: AUTH_BIZ_FIND_AUDIT }, hideMenu: true }`（`hideMenu` 让侧栏在无权限时隐藏菜单项，页面拦截仍由路由 `auth.view` 负责）。`AUTH_BIZ_FIND_AUDIT` 的 transform 不带业务时仍返回单个对象，满足预鉴权「只用静态 sign」的约束。

## 从 PR #1897 取 / 不取

| 内容 | 取舍 |
|------|------|
| `MENU_RESOURCE_RECYCLEBIN` 的 Symbol 定义 | 取（① 已切到 Symbol） |
| `recyclebinRsc` 路由形态（`recyclebin`、owner、activeKey、`auth.view`、`menu.i18n`） | 取；加 `title` 兼容 HEAD 面包屑 |
| `recyclebinBiz` 和业务侧改用 `views/recyclebin/index.vue` | 不取（业务侧不动） |
| `recyclebin/index.vue` 的列表、页签、恢复/销毁、`hcm-auth`、`routerAction` | 取 |
| 回收站从 `route.query.accountId` 读账号、没有选择入口 | 不取；改为页内选择器写入 query |
| `MENU_RESOURCE_DETAIL`、`resourceType` 参数的详情跳转 | 本单不取，保持 `resourceDetail`；由 ⑥ 统一改 |
| `BK_HCM_AJAX_URL_PREFIX` 拼接、`splice(idx)`、`if (recycle_reserve_time)` | 不取（修正） |
| `.recycle-manager-page` 的 `padding: 24px`、`entry-rsc` 的 padding | 取：回收管理用 `page-container`（`main-content` 统一给 24px），`entry-rsc` 照 PR 自带 24px |
| `operationLogRsc` meta 去 `title`、`notMenu`、`isShowBreadcrumb`，加 `auth.view`、`menu.i18n` | 取，但保留 `title`；`isShowBreadcrumb` 换成 `layout.breadcrumb` |
| `operationLogBiz` 的 meta 改造 | 不取（业务侧不动） |
| `entry-rsc` / `entry-biz` 详情链接去掉 `name`、`bizs` 参数，`search.vue` 和 `flow-task` 的样式改动 | 不取（不在本单范围，且会覆盖 HEAD 的改动） |
| `recyclebin-manager.vue` 加 `@deprecated` | 不取：本分支不留废弃副本；业务侧仍在用，资源侧引用本单撤掉，文件保留到业务侧迁移 |

## HEAD 新增能力必须保留

- `views/operation-log/**` 自 PR base 以来的改动，本单不碰组件，只改 `route-config.ts`：操作来源条件（#1899）、云账号管理操作记录适配和 `OPERATION_LOG_RES_TYPES`（#1962）、`condition*.ts` 和 `constants.ts` 的调整。
- `recyclebin-manager.vue` 不改，业务回收站 `/business/recyclebin` 和 `business-manage.vue` 的引用不受影响（AC-010）。
- 新页沿用 HEAD 的 `SearchSelect` 用法、`timeFormatter`、`getRegionName`、moment 的中文相对时间配置。

## 风险与验证

| 风险 | 怎么验 |
|------|--------|
| 业务侧不受影响 | `/business/recyclebin`、业务操作记录与改前逐项一致（AC-010） |
| 侧栏与高亮 | 资源运营侧栏点「回收管理」「操作记录」进新页面，侧栏高亮对应项、顶栏高亮资源运营；操作记录详情面包屑在整页顶部、返回回列表 |
| Symbol 路由名 | `router.getRoutes()` 中 `MENU_RESOURCE_OPERATION_LOG(_DETAILS)` 各只有一条；`entry-rsc.vue` 按 `MENU_RESOURCE_OPERATION_LOG_DETAILS` 跳转详情可用 |
| 转址 | `/resource/resource/recycle?accountId=x&type=disk` 落到 `/resource/recyclebin`，选择器显示 x、页签为硬盘（AC-008）；`/resource/resource/record?accountId=x` 落到 `/resource/record`，账号条件为 x，时间为近 30 天（AC-009）；无效 accountId 回到默认列表；旧资源接入页内点「操作记录」「回收站」页签同样进入新页面 |
| 回收保留时间 | 选账号，改成 3 天并保存，离开再进入显示 3 天；模拟接口失败时值不变；清空账号后设置区消失（AC-003、AC-004） |
| 详情跳转 | 本 commit 及 ⑥ 之前：点主机/硬盘 ID 进入 `/resource/detail/<类型>`；⑥ 之后由 ⑥ 验证跳到新详情路由 |
| 权限 | 没有 `recycle_bin_find` 或资源审计查看权限的账号：菜单隐藏、直接访问显示框架内无权限页（AC-006、AC-007）；没有 `recycle_bin_manage` 时恢复/销毁按钮不可用或弹出申请（AC-S01） |
| `AUTH_BIZ_FIND_AUDIT` 不带 relation 时的语义 | 不传业务时后端走资源审计（`resource_audit_find`、任意账号），与 HEAD `store/common.ts` 的现有用法一致；本单实测无权限用户确实被拦截 |
| 转址生成的 `filter` 能否解析回来 | 转址后搜索栏的账号和时间回显正确，点「重置」回到默认条件 |

## 待确认

无。权限显隐（Q-001）和回收保留时间的位置（Q-002）已在需求文档中确认；其余未验证的点已在正文标「待核实」，由 ②、⑥ 的产物或本单挂载后实测确认。
