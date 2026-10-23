# Coding — feat-menu-phase2-closing

需求文档：仓库根 `docs/reqs/二期收尾体验.md`（TAPD [#1069995598138828468](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138828468)，验收以 AC-001～AC-015、AC-011a、AC-P01、AC-S01 为准）。设计参照 PR #1897（base `8fea9924df`，head `df8865b539`）。基线 `767bdd242`（首批 ⑦）。

**目标**：补齐首批排除的三项（菜单收藏、业务地址 `/business/<业务ID>` 与全局业务选择器、分组折叠），业务资源路由按模块迁出 `router/module/business.ts`，旧 `/403/:id` 退场，并完成首批暂缓的业务侧对齐（F-010）。内里逻辑对齐 #1897；#1897 没有的现网能力（业务侧单据、业务侧云账号管理、② 的侧栏按视图权限显隐）保留。

## 跨单约定

| 项 | 约定 |
|----|------|
| 前置 | 首批 ①～⑦ 已提交（`943fd921e`～`767bdd242`）；本单在同一分支上做，与首批一起合入 v1.9.x |
| 后置 | 「资源列表筛选条件写入地址」等二期合入后从主干另开分支做；本单**不**动列表筛选：`business-manage.vue` 继续给列表组件传 `filter` / `isResourcePage` / `bkBizId`；资源纳管过渡版云账号树（含入口卸载时清空账号缓存）、旧筛选 hook 原样保留；`header.vue` 切换一级菜单时的清空已在本单删除（入口卸载已覆盖） |
| 菜单符号 | 新增与改名一律取 #1897 的名字和描述串（F-010）；HEAD 独有的 `MENU_BUSINESS_TICKET_*`、`MENU_BUSINESS_CLOUD_ACCOUNT*` 保留 |
| 业务 ID 来源 | 以路由参数 `bizId` 为准；拦截器把它同步到 `accountStore.bizs` 与 localStorage（#1897 做法）。`accountStore.bizs` 的收口程度见「已定」1 |
| ⑥ 留下的 `?bizs=` | ⑥ 内联业务详情跳转时额外带 `[GLOBAL_BIZS_KEY]`（11 个文件，需求 F-007 已列）；本单改为 `params.bizId`（或由当前路由继承）后去掉 |
| 合并指引 | 本单完成时更新 `front/.hcmfe/docs/menu-phase2-merge-v1.9.x.md`（§3 补本单删除 / 迁移，§4 补业务地址等约定，§6 删除），随本单提交 |

## 执行顺序

1. 业务路由骨架：`/business/:bizId(\d+)?` 根、`business-interceptor.ts` 按 #1897 补 `?bizs=` 兼容与默认业务（§1）
2. 业务路由按模块迁出 `business.ts`，菜单符号对齐（F-010 命名）、弹性 IP 路径改 `eip`、旧地址兼容（§2）
3. `business-manage.vue` 对齐（不含筛选传参）、业务回收站并入共享页面、回收后跳转（§3）
4. 站内 `?bizs=` 拼法迁移（§4，依赖 1、2：名字与路径定了才能改）
5. 全局业务选择器与业务收藏（§5）
6. 菜单收藏、分组折叠（§6，与 1～5 独立，可并行）
7. 旧 `/403/:id` 退场（§7）
8. `accountStore.bizs` 收口（§8，依赖 1、4）
9. 旧按钮鉴权退场（§10，第 16 轮补入，与 1～8 独立）
10. `bkdevbuddy_lint`、`npm run build`、`tsc` 对比、终扫；更新合并指引

> 1、2 是后面所有改动的地基；6、7 互不依赖，可穿插。提交拆分见「已定」2。

## 改动点

### 1. 业务路由骨架
**文件**：`src/router/index.ts`、`src/router/business-interceptor.ts`

- 根记录改为 `{ name: MENU_BUSINESS, path: '/business/:bizId(\\d+)?', redirect: { name: MENU_BUSINESS_HOST_MANAGEMENT }, children: businessViews }`，子路由全部写相对路径。业务 ID 可选：旧地址 `/business/host?bizs=1` 也能命中。
- 拦截器在 ② 已有的「无业务权限 / 业务不存在 / 业务无权限」三态基础上，按 #1897 补：
  1. `?bizs=` 且无路径业务 ID → `replace` 到同名路由 `params.bizId`，去掉 `bizs`，其余 query 保留（F-005、AC-006）
  2. 无业务 ID → localStorage → 第一个有权限业务 → 第一个业务，`replace` 补上；都没有 → `noBiz`
  3. 校验通过后写 localStorage、`accountStore.updateBizsId(bizId)`
- ② 的「`bizUnauthed` 只显示业务状态页、不弹申请弹窗」（AC-013）保持；`checkViewPermission` 的缓存键沿用 ② 写法。

### 2. 业务路由按模块迁出
**文件**：新增 `src/views/{host,drive,image,vpc,subnet,eip,network-interface,route-table,security-group,cert}/route-config.ts`（照 #1897）；改 `src/views/{recyclebin,operation-log,load-balancer,task,ticket,cloud-account-manage}/route-config.ts`、`src/views/index.ts`、`src/constants/menu-symbol.ts`、`src/common/menu-service.ts`；删除 `src/router/module/business.ts`

- 每个模块一份相对路径记录：列表 `<类型>`、详情 `<类型>/detail/:id?`（⑥ 已改的 Symbol 名与 `normalizeDetailId` 原样搬入）、申请页 `service/service-apply/<类型>`。`meta` 取 HEAD 写法（`title`、`activeKey`、`menu.relative`、`layout.breadcrumb`、② 的 `auth.view`），补 `owner: MENU_BUSINESS`。
- `views/index.ts` 的 `businessViews` 改为各模块数组拼接（#1897 顺序），并保留 HEAD 独有的 `ticketRoutesBiz`、`cloudAccountManageRouteConfig`；`injectStatusComponents(…, statusBusiness)` 保留。
- `load-balancer/route-config.ts` 里的绝对路径 `/business/load-balancer/device` 改为相对路径。
- 菜单符号（F-010）：`MENU_BUSINESS_RECYCLE_BIN_MANAGEMENT` → `MENU_BUSINESS_RECYCLEBIN`；申请页 `applyCvm` / `applyVPC` / `applyDisk` / `applySubnet` → `MENU_BUSINESS_APPLY_{CVM,VPC,DISK,SUBNET}`，`MENU_BUSINESS_LOAD_BALANCER_APPLY` → `MENU_BUSINESS_APPLY_CLB`；`MENU_BUSINESS_LOAD_BALANCER_OVERVIEW` / `MENU_BUSINESS_TARGET_GROUP_OVERVIEW` → `MENU_BUSINESS_LOAD_BALANCER_LB_VIEW` / `MENU_BUSINESS_LOAD_BALANCER_TG_VIEW`（先核对两处页面语义与 #1897 一致）；删除无引用的 `MENU_BUSINESS_HOST_APPLICATION`、`MENU_BUSINESS_HOST_RECYCLE`。引用处同步替换。
- 弹性 IP：列表 `eip`、详情 `eip/detail/:id?`（#1897 是 `ip`，本单按第 13 轮统一为 `eip`）；兼容记录 `/business/:bizId(\\d+)?/ip`、`…/ip/detail/:id?` → 对应新名字，query 保留（AC-011a）。路由表保持 `routing`。
- 回收站旧地址：`/business/recyclebin`、`/business/host/recyclebin/:type`、`/business/drive/recyclebin/:type` → `MENU_BUSINESS_RECYCLEBIN`，`type` 映射到页签。
- `menu-service.ts` 业务菜单项改按新名字取 `route`（`getMenuRoute(businessViews, …)`）。

### 3. 业务列表页与业务回收站（F-008、F-010）
**文件**：`src/views/business/business-manage.vue`、`src/views/recyclebin/{index.vue,route-config.ts}`、`src/views/resource/resource-manage/children/detail/{host,drive}-detail.vue`；删除 `src/views/resource/recyclebin-manager/`

- `business-manage.vue` 照 #1897：组件 / 表单按列表路由 name（`Map<symbol, Component>`）取，不再 `route.path.includes(...)`（改 `eip` 路径的前提）；新建按 `MENU_BUSINESS_APPLY_*` + `routerAction.redirect(…, { history: true })`；鉴权 `relation` 取 `getBizsId()`；去掉等 `accountStore.bizs` 的 `bk-loading`；移除内嵌回收站分支。**保留** `filter` / `isResourcePage` / `bkBizId` 传参（移入筛选写入地址单）。另照 #1897 删两处无接收方的传参（评审时补）：列表组件上的 `v-model:is-form-data-changed`（列表组件都不声明该 prop、不发该事件，只有侧栏表单用，表单那份保留）、侧栏表单上的 `:filter`（`forms/*` 都不声明 `filter` prop）。
- 业务回收站改用共享的 `views/recyclebin/index.vue`：`recyclebin/route-config.ts` 新增 `recyclebinBiz`（`MENU_BUSINESS_RECYCLEBIN`，相对路径 `recyclebin`）；页面按 `whereAmI` 区分业务 / 资源（#1897 已支持，④ 只用了资源侧，逐处核对业务侧接口路径、账号选择器隐藏、查看详情跳业务详情）。
- 回收后跳转：主机、硬盘详情回收成功后，业务侧 `MENU_BUSINESS_RECYCLEBIN`，资源侧 `MENU_RESOURCE_RECYCLEBIN`（修正现网资源侧主机跳业务回收站地址）。

### 4. 站内 `?bizs=` 迁移（F-007）
**文件**：约 59 个文件（`rg "GLOBAL_BIZS_KEY|bizs=|\\bbizs:|query\\.bizs|'/business/|\`/business/|/#/business" src` 的结果，实施时逐个列入本节）

- 按名字跳业务页时不再带 `[GLOBAL_BIZS_KEY]`，改为显式带 `params: { bizId }`（同一视图内跳转也要带）：`bizId` 是可选参数，vue-router 4.1.3 按 name 跳转时不会沿用可选参数，不传的话只能靠业务拦截器从 localStorage 补回。
- 字符串地址 `/business/xxx?bizs=` / `/#/business/xxx?bizs=` 改为 `/#/business/<业务ID>/xxx`，能改按名字的改按名字。
  - GCP 网卡详情「已关联到」（`network-interface-info-gcp.vue`）原来用 `link` 拼字符串地址（资源 `/#/resource/manage/host/details/<id>`、业务 `/#/business/<bk_biz_id>/host/detail/<id>`），#1897 也是这样写。本单改为 `txtBtn` + `routerAction.redirect(..., { history: true })` 按名字跳转：资源 `MENU_RESOURCE_DETAIL`（`resourceType: 'host'`，带 `accountId`、`vendor`），业务 `MENU_BUSINESS_HOST_DETAILS`（`bizId: getBizsId()`，组件两个视图共用）。主机详情的面包屑因此能返回网卡详情，写法同 `ip-info.vue`。资源详情「账号」字段原来用 `link` 拼 `/#/resource/account/details/<id>`，共 8 处：`network-interface-info{,-gcp,-huawei}.vue`、`ip-info.vue`、`route-info.vue`、`drive-detail.vue`、`subnet-detail.vue`、`vpc-detail.vue`。统一改为 `txtBtn` 跳 `MENU_RESOURCE_ACCOUNT_DETAIL`，传 `params.accountId` 并带 `history: true`，写法同 `account-manage/list.vue`。账号详情只挂在资源运营下，业务、资源两侧跳的是同一地址。
  - 申请页新窗口打开的列表：`security-group-selector.tsx`（安全组）、`vpc-selector.tsx`（VPC）和 `cvm/children/VpcPreviewDialog`（VPC，带 `cloud_id` query）原来用 `window.open` 打开拼好的字符串地址。现在改为 `routerAction.open` 按名字打开：业务侧为 `MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT` / `MENU_BUSINESS_VPC_MANAGEMENT`（`bizId: getBizsId()`），资源侧为 `MENU_RESOURCE_RESOURCE_LIST`（`resourceType: 'security' / 'vpc'`）。新窗口没有返回栈，所以不带 `history`。
- `menu-item.vue`、`breadcrumb.vue`、`header.vue` 里给业务链接补 `?bizs=` 的逻辑改为 `params.bizId`（#1897 的 `menu-item.vue` 写法）。
- ⑥ 内联的 11 处业务详情跳转去掉 `[GLOBAL_BIZS_KEY]`（需求 F-007 列表）。
- 遗漏的旧拼法由 §1 的兼容转址兜底，不出现空白页。

### 5. 全局业务选择器与业务收藏（F-006）
**文件**：新增 `src/components/business-selector/global.vue`、`src/hooks/use-business-favorite.ts`（照 #1897）；改 `src/components/layout/menu.vue`；删除 `src/components/layout/business-selector.tsx`、`src/hooks/useFavorite.ts`

- 选择器放侧栏顶部（`menu.vue` 的固定区，只在业务视图显示；侧栏收起时照常渲染，与 #1897 一致）。切换业务：目标为 `route.meta.menu?.relative ?? route.name`，`router.resolve` 后整页刷新（R-004、AC-008）。
- 保留首批在旧选择器上加的判断：当前业务不在有权限列表中时不查收藏（AC-013）；实测首次加载能否取到收藏，取不到补首次查询（AC-009）。
- 「新建业务」引导用现有 `BK_CMDB_CREATE_BIZ_URL`。

### 6. 菜单收藏与分组折叠（F-001～F-003）
**文件**：新增 `src/hooks/use-menu-favorite.ts`、`src/hooks/use-menu-collapse.ts`、`src/components/layout/menu-favorites.vue`（照 #1897）；改 `src/components/layout/menu.vue`、`menu-item.vue`、`src/common/menu-service.ts`

- 侧栏分为固定区（业务选择器 + 「我的收藏」）与滚动区（常规菜单）；分组由 `bk-menu-group` 改为可折叠的 `bk-submenu`，折叠状态按「一级菜单 × 分组」存 localStorage（按用户区分）；收藏同样按一级菜单存 localStorage。
- 收藏区只展示有权限且存在的菜单项，并清理失效收藏（F-002）。**过滤规则沿用 ② 的 `hideMenu` 判定**（`viewAuthConfig` 的 `hideMenu` + `isViewPermissionCached`），不取 #1897 的「所有 Symbol 菜单都按 `hasViewPermission` 过滤」（会把二级账号等不隐藏的菜单藏掉）。
- 点击来源弱化高亮（AC-004）照 #1897 的 `lastClickSource`。
- 菜单图标与分组：`bk-submenu` 需要 `groupIcon`，分组内菜单项不带图标。业务侧改为「云资源 / 操作管理」两组 + 独立的「云账号」「回收站」，资源运营「云账号管理」组改名「云账单管理」，见「已定」3、4。

### 7. 旧 `/403/:id` 退场（F-009）
**文件**：`src/router/index.ts`、`src/views/index.ts`；删除 `src/router/module/common.ts`、`src/views/error-pages/403.{tsx,scss}`

- 照 #1897 不再注册 `/403/:id`，旧书签落 catch-all 的 404 页（第 15 轮，AC-012）。
- 同时删除：全局守卫里的 `to.name === '403'` 分支、自指的 `{ path: '/403', redirect: '/403' }` 记录、`useWhereAmI` 的 `/403/` 判断与 `Senarios.unauthorized`（无其他引用时）。
- 守卫里只为旧 403 页服务的 `biz_access` 读取一并去掉；守卫里的旧按钮预取与「首次导航」标志一并删除，旧按钮鉴权整体退场见 §10（第 16 轮）。

### 8. `accountStore.bizs` 收口
**文件**：`src/router/business-interceptor.ts`、`src/hooks/useWhereAmI.ts`、`src/store/account.ts`、`src/store/business.ts`

#1897 的拦截器把路由 `bizId` 同步进 `accountStore.bizs`，并留 TODO「应该使用 `route.params.bizId`」。需求约束要求本单收口。现网 34 个文件、69 处读 `accountStore.bizs`（含 `useWhereAmI.getBizsId` 的回退链）。按「已定」1：单向同步保留，`getBizsId` 先读路由，store 注明只读镜像，69 处不逐个改。

`store/business.ts` 照 #1897（评审时补）：删掉文件内自带的 `getBusinessApiPath`（按 `location.href.includes('business')` 判业务、读 `accountStore.bizs` → localStorage），改为 setup 写法 `defineStore('business', () => {...})`（id 由 `businessStore` 改为 `business`），在 setup 里取 `useWhereAmI().getBusinessApiPath`（按路由判业务视图、`getBizsId` 先读路由）。方法体保留 HEAD 版本（如克隆安全组的 `target_region`）。setup store 在 pinia 2.3 + Vue 3.5 下经 `runWithContext` 执行，`useRoute()` 可取到路由；全部调用方都在路由安装后的组件 / hook 里，无 `this` 与 `$` API 使用。

### 10. 旧按钮鉴权退场（F-011，第 16 轮）
**文件**：`src/main.ts`、`src/store/common.ts`、`src/hooks/index.ts`、`src/components/index.ts`、`src/views/scheme/{scheme-list/index.tsx,scheme-detail/index.tsx,components/scheme-selector/index.tsx}`、`src/views/bill/account/account-detail/second-level-account-detail/index.tsx`、`src/views/resource/resource-manage/common/info-list/info-list.tsx`、`src/views/ticket/children/apply-detail/account-apply-detail/index.tsx`；删除 `src/hooks/useVerify.ts`、`src/components/permission-dialog/`、`src/assets/image/403.png`、`src/views/service/service-apply/components/{basic-resource,senior-resource,resource-card}/`

本分支 `store/common.ts` 的旧鉴权方法从未删除，旧按钮现网可用；迁移要求权限项、是否带实例与迁移前一致。

| 位置 | 旧（`permissionAction` id） | 新 |
|---|---|---|
| 方案列表「创建选型方案」 | `cloud_selection_recommend`（无权限时 `handleAuth('cloud_selection_create')`，id 对不上，申请弹窗打不开） | `hcm-auth` + `AUTH_CREATE_CLOUD_SELECTION_SCHEME`（照 #1897） |
| 方案列表名称旁编辑 | `cloud_selection_edit` | `AUTH_UPDATE_CLOUD_SELECTION_SCHEME` |
| 方案列表「删除」、方案详情「删除」 | `cloud_selection_delete` | `AUTH_DELETE_CLOUD_SELECTION_SCHEME` |
| 方案选择器「编辑」 | `cloud_selection_edit` | `AUTH_UPDATE_CLOUD_SELECTION_SCHEME` |
| 二级账号详情「编辑」 | `provide('authAction', main_account_edit)`，#1995 改整页弹窗后无字段消费，入口实际无鉴权 | 保留整页弹窗，入口包 `HcmAuth` + `AUTH_UPDATE_MAIN_ACCOUNT`（#1897 是字段级 `getAuthSign`，不取） |
| `info-list` | `inject('authAction')` / `isAuth` / `handleAuth` | 删除，只留 `getAuthSign`（照 #1897） |
| 账号申请单据详情 | `root_account_find` 决定是否拉一级账号列表 | 读 `authStore.isViewPermissionCached(MENU_BILL_ROOT_ACCOUNT) && hasViewPermission(...)`（启动已预取 `AUTH_FIND_ROOT_ACCOUNT`）；#1897「总是请求、吞掉失败」在本分支会触发接口 403 弹窗，不取 |
| 服务申请卡片 `basic-resource` / `senior-resource` / `resource-card` | `account_import` / `biz_iaas_resource_create` | 三个组件全仓无引用（路由名 `applyCvm` 等本单已改符号），删除 |

- `main.ts` 恢复 `preload().finally(() => {...})`（#1897 同），去掉 `useVerify`、`useCommonStore` 引入。
- `store/common.ts` 只留 `isNoticeAlert` / `setIsNoticeAlert`（`app.vue`、`views/notice`、`use-calc-top-with-notice` 在用），照 #1897 改为 setup 写法、id 由 `commonStore` 改为 `common`（无持久化、消费方只用 `storeToRefs` 与 action，不受影响）；删 `list`、`authVerifyData`、`authVerifyParams`、`pageAuthData`、`authVerify`、`authActionUrl`、`addAuthVerifyData`、`updatePageAuthData` 与 `BK_HCM_AJAX_URL_PREFIX`。
- `hooks/index.ts` 去掉 `useVerify` 导出；`components/index.ts` 去掉 `PermissionDialog` 全局注册。
- `common.scss` 的 `.hcm-no-permision-*` 保留（`useMoreActionDropdown.tsx` 仍用）。
- 一级账号详情编辑入口现网无鉴权、无对应符号，不在本单。
- **无权限优先于业务提示（2026-10-15 评审定）**：`hcm-auth` 在 `noPerm` 时显示锁形光标、并用透明覆盖层接收点击；如果 slot 里按钮的业务 tooltip 没有判断 `noPerm`，鼠标悬停时会同时出现锁形光标和业务提示。规则：`hcm-auth` slot 内的 `v-bk-tooltips` / `withDirectives([[bkTooltips, …]])` 一律写成 `noPerm ? { disabled: true } : 原选项`；原选项本身是三元的，写成 `(!noPerm && 原选项) || { disabled: true }`，避免嵌套三元。写法与 #1897 `load-balancer-table.vue` 的 `getSingleDeleteDisabledTooltips(row, noPerm)` 及本分支 `host-drive` / `host-ip` / `host-detail` / `vpc-detail` 一致；#1897 的 drive / ip / host / security 管理页没有处理，本单补上。没有抽公共 helper：同一行判断，抽函数收益不大。本轮修改的文件：负载均衡的 `device/main-content/listener-table.vue`、`listener/listener-table.vue`、`children/export/batch-button.vue`、`listener/layer7/rule-collapse-panel.vue`；`business/cert-manager/index.tsx`、`clb-detail/index.tsx`、`all-groups-manager/index.tsx`；`plugin-handler/business-host-manage.tsx`；资源的 `host-manage` / `ip-manage` / `drive-manage` / `security-manage` 操作列，以及 `HostOperations.tsx`、`security-info` / `security-rule` / `platform` / `collapse-data-list`、`ip-detail`；`recyclebin/index.vue`；权限模板列表和详情；三级账号详情侧栏。另外 `cert-manager` 删除按钮原来的 tooltip `disabled` 条件和按钮禁用条件相同，导致需要提示时反而不显示，已改为取反。下拉菜单项上的 tooltip（例如 `host-manage` 的更多操作）不受这个问题影响：无权限时覆盖层挡住点击，菜单展不开。
- 资源详情里残留的旧鉴权注释（评审时核对，均与 #1897 一致，本单不补鉴权）：`drive-detail.vue` 的挂载 / 卸载 / 回收自 #368（2023-11）起前端鉴权即被注释，只靠后端校验，无权限时由 `http/index.ts` 弹申请弹窗；`gcp-detail.vue` 的修改按钮在 `a5acb86a74`（2023-08「屏蔽旧修改button」）已去掉，注释是随按钮遗留的；`subnet-manage.vue` 的整列「操作」都已注释，没有删除按钮。另 `security-detail.vue` 的「同步」也无前端鉴权（#1897 同）。`drive-detail.vue` 的两段注释已删，另两处注释夹着其他已停用逻辑，不动。前端补鉴权另立单据。

## 从 PR #1897 取 / 不取

| PR 内容 | 取舍 |
|---|---|
| `/business/:bizId(\d+)?` 根、拦截器的 `?bizs=` 兼容与默认业务 | 取；状态页与权限缓存沿用 ② |
| 业务路由按模块 `views/*/route-config.ts` | 取；保留 HEAD 独有的业务侧单据、云账号管理路由与 ② 的 `auth.view` |
| 弹性 IP 路径 `ip` | 不取：按第 13 轮改 `eip`，旧 `ip` 兼容 |
| `business-manage.vue` 按路由 name 选组件、新建按符号跳转、去 loading、去内嵌回收站 | 取 |
| `business-manage.vue` 去列表组件的 `v-model:is-form-data-changed`、去侧栏表单的 `:filter` | 取：两处都无接收方，与筛选无关 |
| `business-manage.vue` 去 `filter` / `isResourcePage` / `bkBizId` 传参 | 不取：移入筛选写入地址单 |
| 业务回收站 `recyclebinBiz` 用共享页面 | 取 |
| `global.vue`、`use-business-favorite.ts` | 取；保留「无权限业务不查收藏」判断，补首次加载 |
| 菜单收藏、分组折叠、侧栏固定区 / 滚动区 | 取 |
| 侧栏按 `hasViewPermission` 过滤所有 Symbol 菜单 | 不取：沿用 ② 的 `hideMenu` 判定 |
| 业务分组改名「云资源 / 操作管理」 | 不取（分组名不在本单范围） |
| 不注册 `/403/:id` | 取 |
| `MENU_BUSINESS_*` 命名（F-010） | 取 |

## HEAD 新增能力必须保留

- 业务侧单据（`ticketRoutesBiz`、`MENU_BUSINESS_TICKET_*`）、业务侧云账号管理（`cloud-account-manage`、`MENU_BUSINESS_CLOUD_ACCOUNT*`）：#1897 没有，迁移时作为独立模块保留。
- ② 的视图权限：侧栏 `hideMenu`、`meta.auth.view`、业务状态三态、`statusBusiness` named view。
- ⑥ 的业务详情：Symbol 名、`detail/:id?`、`normalizeDetailId`、`useResourceDetailAuth` 按路由选签名（不能按「有无业务 ID」判断）。
- ⑦ 的账单 `/resource/bill/*`、`useWhereAmI` 的 `/resource/bill/` 优先判断。

## 风险与验证

| # | 风险 | 怎么验 |
|---|------|--------|
| 1 | 旧地址兼容（AC-006、AC-011a） | 带 `?bizs=` 打开业务各列表、详情（含 `?id=`）、申请页、回收站、负载均衡、任务、操作记录、单据、云账号管理；弹性 IP 旧 `ip` 路径 |
| 2 | 业务状态（AC-007、AC-013、AC-S01） | 无业务、业务不存在、业务无权限三种地址；无权限时只显示状态页、不弹申请弹窗、不返回该业务数据 |
| 3 | 路由迁移后页面一致（AC-011） | 逐项打开业务各菜单、详情、新建页：页面、侧栏高亮、面包屑、查看权限与迁移前一致；`router.getRoutes()` 无重复 name |
| 4 | 站内跳转（AC-010） | `rg "GLOBAL_BIZS_KEY\|bizs=" src` 只剩拦截器与兼容记录；业务视图内各页面跳转地址不含 `bizs` |
| 5 | 切换业务（AC-008、R-004） | 列表、详情、新建、GCP 防火墙 / 模板详情各切一次；整页刷新后落在同类页面 |
| 6 | 业务收藏（AC-009） | 搜索名称 / ID、星标置顶；刷新后首次展开仍置顶 |
| 7 | 菜单收藏与折叠（AC-001～AC-005） | 三个一级菜单各自收藏互不影响；权限收回后收藏不显示；折叠状态刷新与切一级菜单后保持 |
| 8 | 回收站（AC-014） | 业务主机 / 硬盘进回收站、详情回收后跳转、旧回收站地址；资源侧主机回收跳资源回收管理 |
| 9 | 旧 403（AC-012） | `#/403/xxx` 进 404 页 |
| 10 | 符号命名（AC-015） | `menu-symbol.ts` 与 #1897 逐个比对（HEAD 独有能力除外） |
| 11 | 性能（AC-P01） | 切换业务整页刷新耗时与现网对比 |
| 12 | 构建与类型 | `bkdevbuddy_lint`、`npm run build`、`tsc` 与 ⑦ 提交后对比无新增错误 |
| 13 | 旧按钮鉴权退场（AC-016～AC-019） | 方案列表 / 详情 / 选择器、二级账号详情「编辑」有无权限各一次；无一级账号权限打开账号申请单据详情不发请求、不弹窗；`rg "useVerify\|authVerifyData\|pageAuthData\|permission-dialog\|authAction" src` 无结果；启动只有一次 `/auth/verify` |

## 已定

1. **`accountStore.bizs` 收口**（选 A）：拦截器保持「路由 `bizId` → `accountStore.bizs`」单向同步（#1897 做法）；`useWhereAmI.getBizsId` 改为先读路由参数、业务视图外不再回退 localStorage；站内新写的代码一律读路由；现有 69 处读 `accountStore.bizs` 不逐个改，删掉 #1897 的 TODO 并在 store 上注明「只读镜像、由拦截器写入」。
   - **新代码取业务 ID 的写法**（2026-10-15 评审定）：组件**只在业务视图使用**，且已引入 `useRoute()` 或已有从路由取的业务 ID（如负载均衡 `entry-biz.vue` 由 `route.params.bizId` 派生后 provide 的 `currentGlobalBusinessId`）时，直接用路由的；否则（资源 / 业务共用的组件，或没引路由的组件）用 `getBizsId()`，不为此额外引入 `useRoute`。新代码不读 `accountStore.bizs`。已按此改：`tertiary-account/children/data-list/data-list.vue`、`cloud-secret/children/secret-detail-sideslider/index.vue`、`host-security.vue` 跳转里的 `accountStore.bizs` → `getBizsId()`。
     - 已有的 `route.params.bizId` 写法按「与 #1897 对齐，#1897 没改的这次也不改」处理，不追溯套用上面的规则。`?bizs=` 迁移中有 11 个资源 / 业务共用文件新加了 `params: { bizId: route.params.bizId }`：`useRouteLinkBtn.tsx`、`recyclebin/index.vue`、`host-drive`、`ip-info`、`route-info`、`drive-detail`、`host-detail`、`security-manage`、`use-columns.tsx`、`service-apply/clb/index.tsx`、`task/clb.vue`。#1897 在这些位置都不显式传业务 ID（`task/clb`、`service-apply/clb` 仍带 `?bizs=`），也没有改用 `getBizsId()`，所以前 10 个保留 `route.params.bizId`。`task/clb.vue` 在 #1897 用的是 `getBizsId()`，本文件其余接口也都用它，因此跳转改为 `getBizsId()`。不照 #1897 去掉显式 `bizId`，原因是 vue-router 4.1.3 的 `resolve` 按 name 跳转时只沿用必填参数（`matcher.keys.filter(k => !k.optional)`），而 `/business/:bizId(\\d+)?` 是可选参数，不传就会丢失；#1897 是靠业务拦截器从 localStorage 或默认业务补回来的。
2. **提交拆分**（多条）：首条用 TAPD 标题，后续按改动总结，均带 `--story`：业务路由与旧地址兼容（§1、§2、§7）→ 业务列表页与回收站（§3）→ 站内 `?bizs=` 迁移（§4、§8）→ 业务选择器（§5）→ 菜单收藏与折叠（§6）。
3. **菜单图标**（照 #1897 / 设计稿）：分组内的菜单项不带 `icon`，图标只用在分组标题（组内第一项的 `groupIcon`）和不在分组里的独立菜单（`icon`）。分组图标：「云资源」`bkhcm-icon-vpc`、「操作管理」`bkhcm-icon-operation-record`（#1897 映射），资源运营沿用。独立菜单：业务「回收站」照 #1897（去掉只有它一项的「回收站」分组，`bkhcm-icon-recyclebin`）、业务「云账号」（原「账号」分组下的「云账号管理」，`bkhcm-icon-account-manage`）、服务请求「单据管理」（`bkhcm-icon-my-apply`）与「权限策略库」（`bkhcm-icon-configuration`，去掉原「资源」分组）、资源运营「操作记录」。
4. **分组名**（2026-10-13 补充）：业务「资源」→「云资源」、「其他」→「操作管理」（均与 #1897 同名）；资源运营「云账号管理」→「云账单管理」。云账号页的面包屑标题仍是路由上的「云账号管理」。「操作管理」组内顺序保持现网（单据管理、任务管理、操作记录）。服务请求本单只去掉分组；#1897 的「工作台」（个人事务 / 公共事务，含资源选型、账号管理）需要整体迁移菜单，留到三期。

## 实施记录

### 与计划 / #1897 的差异

| 项 | 做法 | 原因 |
|---|---|---|
| `?bizs=` 兼容 | 与 #1897 相同：只在路径没有业务 ID 时转成路径参数并 `replace`；路径上已有业务 ID 时以路径为准 | 旧 query 不覆盖新路径参数 |
| 拦截器写法 | 与 #1897 相同：拦截器接收 `next`，转址和状态页自己调用，返回 `false` 表示已处理 | 返回地址再由守卫 `next(location)` 时，`false` 会被当成「继续导航」，和「已切状态页」混在一起 |
| 路由 `Meta` | `router/meta.ts` 照 #1897 重写：默认 `layout.breadcrumb = { show: true, back: true }`、`activeKey` / `owner` 默认 `Symbol.for('')`，不读 `title`；`use-breadcrumb` 标题只取 `menu.i18n`。类型里去掉 `title`（#1897 标了 deprecated 仍保留，但传了会被静默忽略）及无人读取的 `notMenu`、`icon`。9 个路由文件 22 处 `title` 改为 `menu.i18n`，同时去掉 `notMenu` / `icon` | 原 `Meta` 先把 `this.menu` 指向传入对象再写 `menu.i18n = title`，只写 `menu.i18n` 的路由（#1897 写法）标题被覆盖为空，业务侧各列表与回收站面包屑无标题。已核对默认值不会盖掉父路由配置：有子路由的账号详情、资源纳管子路由都写全了，账单管理子路由与资源选型不用 `Meta` |
| 默认业务与转址方式 | 与 #1897 相同：不读来源路由，`localStorage` → 第一个有权限业务 → 第一个业务，转址带 `replace: true`。`noBiz` 仍挂 `VIEW_AUTH_BIZ_ACCESS` 的 `permissionData`（#1897 的 noBiz 不可申请） | 先定路由上的业务再鉴权；`localStorage` 多标签共享是预期。站内跳转已显式传 `bizId`，`replace` 只作用于缺业务 ID 的中间地址。不在定业务 ID 之前用「没有任何业务权限」拦截 |
| 站内跳转带业务 ID | 菜单（`menu-item`、`menu-favorites`）照 #1897 经 provide / inject 显式传 `bizId`；约 60 处跳业务路由补 `params.bizId`（`route.params.bizId` / `currentGlobalBusinessId` / `getBizsId()` / `accountStore.bizs`） | 同上；#1897 只有菜单显式传 |
| `getBizsId` | 路由 `bizId` → `accountStore.bizs` → localStorage（#1897 只读路由） | 组件外（store、http）取不到路由；去掉了读 `?bizs=` 的一档 |
| 负载均衡 | `entry-biz.vue`、`clb/{index,details,load-balancer-list,load-balancer-table}.vue`、`route-config.ts` 取 #1897 版本；购买页保留 HEAD 的 `breadcrumb.show=false` | 原 `entry-biz` 按 `route.path.includes('/business/load-balancer/…')` 判断页签，路径插入业务 ID 后失效 |
| 面包屑返回栈（2026-10-15 评审补） | `routerAction.redirect` 的 `back` 分支由 `HistoryStorage.remove(当前页 name)` 改为 `HistoryStorage.removeLast()`（弹出栈顶，即返回的目标页）；删掉不再有调用方的 `remove(name)` | 当前页本身从不入栈，按 name 删除要么删不到，要么误删前面的同名记录。返回后栈顶残留目标页自身，目标页的 `from` 指向自己，点返回原地跳转、无反应。复现：主机列表→主机详情→硬盘详情→挂载主机（主机详情）→返回（正常）→硬盘详情再返回无反应；简单的三级路径（列表→详情→子详情，连退两次）同样会出现。#1897 与 `origin/v1.9.x` 写法相同，属于原有问题 |
| 详情页签高度（2026-10-15 评审补） | `common/tab/detail-tab.scss` 照 #1897（`5c513eb559`）去掉 `.detail-tab-main { height: 100% }` | 详情外层由 `i-detail-tap-wrap`（父级高度不确定，100% 实际按 auto 计算）改为 `.detail-content-wrap`，链路是 `resource-entry-content` 的 `flex: 1` 加 `min-height: 0`，再到 `height: 100%`，高度变成确定值。第一个页签因此撑满整个可视区，与第二个页签（如 VPC 详情的 CIDR / 子网 / 路由表）之间空出大段距离。去掉后页签高度由内容决定，超出时由 `.detail-content-wrap` 的 `overflow-y: auto` 滚动。影响所有 `detail-tab`（vpc、gcp、ip-info 等有两个页签的最明显），华为网卡 `huawei-network.vue` 自带固定高度，不受影响 |
| 证书托管 `isFilterAccount` | 放在 `meta` 顶层（#1897 放进 `extra`） | 账号选择器读 `route.meta.isFilterAccount`，放进 `extra` 读不到 |
| 申请页标题 | 沿用 HEAD「购买主机 / 购买VPC / 购买硬盘 / 新建子网」，GCP / 模板详情沿用「GCP防火墙详情 / 参数模板详情」 | 标题不在本单范围 |
| 业务回收站 | 共享 `views/recyclebin/index.vue` 按 `whereAmI` 区分：业务侧不显示账号选择与保留时间配置、不提供单行操作，详情跳业务详情 | 与旧 `recyclebin-manager` 业务侧行为一致 |
| 业务收藏 | `global.vue` 的 watch 加 `immediate` | #1897 首次加载不查收藏，星标不置顶（AC-009） |
| 菜单收藏 | 失效收藏按「菜单是否存在」清理，权限只决定显示；计数取实际显示的条数 | #1897 按权限过滤后的列表清理，权限接口失败或暂无权限时会永久删掉收藏 |
| 业务切换 | 去掉 `app.vue` 的 `reload` / `isRouterAlive`（只服务旧选择器） | 切换业务改为整页刷新 |
| 顶栏与面包屑 | `header.vue` 语言切换照 #1897 完整写法（`js-cookie` 删后写、366 天、有组件 API 才调接口、默认 `zh-cn`），只把 `BK_HCM_DOMAIN` 换成 `BK_DOMAIN`；删除切换一级菜单时清空资源账号缓存；`breadcrumb.vue` 去掉无人传入的 `backConfig` | #1897 的 `BK_HCM_DOMAIN` 已在 #1647 改名，照抄会写成当前域名 cookie；账号缓存由 `resource-entry.vue` 卸载时清空 |
| `app.vue` 布局状态 | 侧栏状态改用 #1897 的 `navigationState.collapse` / `handleCollapse`，只用于 `default-open`，侧栏不再接收（收起时照常渲染业务选择器，与 #1897 一致）；删除两个 `accountStore.bizs` 监听：切业务重拉 `pageAuthData` 的监听已无消费方（`biz_*` 权限项无人读取），「非业务页清零」移到 `router/index.ts` 的 `afterEach` | 布局入口只管布局，对齐 #1897 |
| 跑马灯 | 保留 `has-notice`（#1897 未处理，`<Notice>` 在导航之后且不减高度，会多出 40px 滚动条） | 旧外壳有此处理，#1897 删外壳时遗漏 |
| 页脚 | 删除 `business-manage.vue` 的 `is-host-page`（主机页去掉底部内边距，#1897 无）；不加 #1897 的 `has-footer` | `app-shell.scss` 的 flex 布局已给页脚留位，`has-footer` 在本分支无作用 |
| 主视图结构 | 不加 #1897 `main-content.vue` 的 `.main-layout` 包裹层；主视图删 `view-warp` 类与 `app-shell.scss` 的 `.view-warp` 规则（与 `.main-view` 自身样式重复，#1897 无此类），并删 HEAD 起就无使用方的 `.navigation-resource` 样式；`.main-view` 保留 `overflow-y: auto`（#1897 为 `hidden`） | `.main-layout` 只是用 `calc(100% - 52px)` 扣面包屑高度，本分支纵向 flex 已覆盖；保留外层滚动是因为还有旧页面不在 `page-container` 内自滚 |
| 收藏区图标 | 无收藏 `bkhcm-icon-not-favorited`、有收藏 `bkhcm-icon-collect` + 金色（#1897 只换颜色，一直是实心星） | 与菜单项星标同一套逻辑，符合设计稿 |
| 侧栏外观 | 删除 `app-shell.scss` 里旧外壳的侧栏 / 菜单 / 选择器底色覆盖与列表上内边距，用 bkui 默认暗色；选择器外边距照 #1897 写在 `app.vue` | 对齐 #1897 |
| 侧栏宽度 | 删 `app.vue` 的 `:hover-width="240"` 与 `menu.vue` 的 `.menu-warp { width: 240px }`，侧栏与菜单都用 bkui 默认 260px；根节点类名改为 #1897 的 `hcm-sidebar`；`app-shell.scss` 删已无引用、按 240 侧栏写的 `.action-bar` / `.button-group` 固定底栏样式 | 对齐 #1897（240 来自旧外壳 `NAV_WIDTH`） |
| 菜单项结构 | 照 #1897：文字直接放默认插槽，星标 `margin-left: auto; margin-right: 24px` | 对齐 #1897 |
| 业务选择器依赖 | `@blueking/app-select` 升到 `0.0.1-beta.19`（#1897 同） | `beta.6` 没有 `clearable` 参数，有值时总显示清除按钮，`:clearable="false"` 不生效 |
| `package.json` 其余差异 | `@types/cookie` 换成 `@types/js-cookie@3.0.6`（#1897 同）；`bkui-vue`、`less`、安全升级等保留主干更新的版本；不加 #1897 的 `packageManager: pnpm`（项目按 npm：`engines`、`.npmrc`、现有安装方式，是否切 pnpm 另行决定） | 项目只用 `js-cookie`，原类型包对不上；与 #1897 的其余差异都是主干在 #1897 基线之后的升级（#1887、#1911、#1920、#2005 等） |
| 旧 `/403` | `Senarios.unauthorized` 枚举保留（#1897 同） | 仅删 `useWhereAmI` 的 `/403/` 判断 |
| 状态页路由 | `statusRouters`（`/404`、`/error`）直接注册，不经 `injectStatusComponents`（#1897 同） | 它们自己就是状态页，不再套 permission / error / notFound named view |
| 转址 `hash` | 不传（#1897 同） | 应用是 hash 路由，`to.hash` 是路由后面的第二段 fragment，业务地址不用 |
| 旧按钮鉴权预取 | 删除（第 16 轮，见 §10）；守卫里的预取与 `isFirstNavigation` 一并去掉 | 原计划挪到 `main.ts` 保留；评审时确认剩余消费方只有 6 个文件且 #1897 已全部迁走，本单一次清掉 |
| `MENU_PLATFORM_MANAGEMENT` | 删除 | 外部版顶栏无平台管理，全仓无引用，#1897 也没有；`MENU_RESOURCE_ACCESS` 仍是顶栏「资源接入」入口，留给资源接入下线单 |

### 文件

- 路由：`router/index.ts`、`router/business-interceptor.ts`、`views/index.ts`；新增 `views/{host,drive,image,vpc,subnet,eip,network-interface,route-table,security-group,cert}/route-config.ts`；改 `views/{recyclebin,task,operation-log,ticket,cloud-account-manage,load-balancer}/route-config.ts`；删 `router/module/{business,common}.ts`、`views/error-pages/403.{tsx,scss}`
- 菜单与布局：`constants/menu-symbol.ts`、`common/menu-service.ts`、`components/layout/{menu,menu-item,header,breadcrumb,main-content}.vue`、`app.vue`、`style/app-shell.scss`；新增 `components/business-selector/global.vue`、`components/layout/menu-favorites.vue`、`hooks/{use-business-favorite,use-menu-favorite,use-menu-collapse}.ts`；删 `components/layout/business-selector.tsx`、`hooks/useFavorite.ts`
- 业务列表与回收站：`views/business/business-manage.vue`、`views/recyclebin/index.vue`、`plugin-handler/business-host-manage.tsx`、`views/business/host/children/host-operations/use-batch-operation.tsx`、`views/resource/resource-manage/children/detail/{host,drive}-detail.vue`；删 `views/resource/recyclebin-manager/`
- `?bizs=` 迁移（§4）：`hooks/{useWhereAmI,useRouteLinkBtn}.ts(x)`、`store/business.ts`、`utils/lb.ts`、`views/load-balancer/{entry-biz.vue,clb/*,listener/{add,children/info,children/batch-delete-dialog,layer4/target-group,layer7/rule-collapse-panel}.vue,device/main-content/{listener-table,url-rule-table,children/rs-ip-group,children/batch-rs-operation-dialog}.vue,clb/children/batch-import/index.vue}`、`views/business/load-balancer/{clb-view/components/TargetGroupSelector,group-view/components/AddOrUpdateTGSideslider,group-view/target-group-list}/index.tsx`、`views/operation-log/entry-{biz,rsc}.vue`、`views/task/{index,clb}.vue`、`views/task/details/children/rerun/rerun.vue`、`views/ticket/children/apply-detail/{index.tsx,components/status.vue}`、`views/cloud-account-manage/{index.vue,permission-policy/index.vue}`、`views/service/service-apply/{clb/index.tsx,components/common/{security-group-selector,vpc-selector}.tsx,cvm/children/VpcPreviewDialog/index.tsx}`、`views/resource/resource-manage/{hooks/use-columns.tsx,children/manage/security-manage.vue,children/detail/drive-detail.vue,children/components/{ip/ip-info,route/route-info,host/host-drive,host/host-security,network-interface/network-interface-info-gcp}.vue}`

### 提交拆分（调整为三条）

按「每条提交单独可运行」切：路由迁出不依赖地址形态；路径插入业务 ID 后读 `?bizs=` 的页面与旧选择器会同时失效，必须和 `?bizs=` 迁移、新选择器同一条。

1. **首条（TAPD 标题）：业务路由按模块迁出**（§2、§3、§7）— 地址仍是 `/business/xxx?bizs=`，拦截器与旧选择器保持 HEAD
   - 新增 10 个模块 `route-config.ts`、`recyclebinBiz` 与业务侧旧地址（`ip`、`host|drive/recyclebin/:type`）；改 `views/index.ts`、task / operation-log / ticket / cloud-account-manage 的 `owner`；删 `router/module/business.ts`
   - F-010 命名：`menu-symbol.ts`；引用处只改名字（负载均衡各文件、`ticket/.../status.vue`、`TargetGroupSelector`、`target-group-list`、`business-host-manage.tsx`、`use-batch-operation.tsx`、`menu-service.ts` 的回收站 id）
   - `business-manage.vue` 对齐（按路由名选组件、申请页按符号跳转、去 loading、去内嵌回收站）；`views/recyclebin/index.vue` 业务侧；删 `recyclebin-manager/`；`host-detail` / `drive-detail` 回收后跳转
   - 旧 `/403` 退场：`router/index.ts` 删 403 分支与记录、删 `router/module/common.ts` 与 `views/error-pages/`、`useWhereAmI` 删 `/403/` 判断
2. **业务地址改为 `/business/:bizId`**（§1、§4、§5、§8）
   - `router/index.ts` 根路径参数与守卫 `next(location)`；`business-interceptor.ts` 的 `?bizs=` 兼容与默认业务；`useWhereAmI.getBizsId`、`store/{account,business}.ts`
   - 负载均衡按 #1897 改结构（`entry-biz.vue` 改 `<router-view>`、`route-config.ts` 改相对路径与嵌套）——原写法按地址字符串判断页签，插入业务 ID 后失效
   - 全局业务选择器与业务收藏：新增 `global.vue`、`use-business-favorite.ts`，`menu.vue` 换选择器与切换逻辑，删 `business-selector.tsx`、`useFavorite.ts`、`app.vue` 的 `reload`、`main-content.vue` 的 `isRouterAlive`、`app-shell.scss` 旧选择器样式
   - 站内 `?bizs=` 迁移（上文「文件 · `?bizs=` 迁移」全部）；`menu-item.vue` / `header.vue` / `breadcrumb.vue` 去 `?bizs=`；`security-manage.vue` 与 `host-security.vue` 的绝对路径改按名字
3. **菜单收藏与分组折叠**（§6）：新增 `use-menu-{favorite,collapse}.ts`、`menu-favorites.vue`；`menu.vue` 固定区 / 滚动区与 `bk-submenu`；`menu-item.vue` 星标；`menu-service.ts` 的 `groupIcon`
4. **旧按钮鉴权退场**（§10）：§10「文件」全部；`main.ts` 的预取删除与守卫里的预取删除（第 1 条）合起来看，`main.ts` 只在本条改

跨条的文件（`router/index.ts`、`menu-symbol` 引用处的负载均衡与单据文件、`menu.vue`、`menu-item.vue`、`menu-service.ts`、`drive-detail.vue`、`useWhereAmI.ts`）提交时按块分批暂存；每条提交前在临时工作树单独跑 `build` 与 `tsc` 确认可运行。

### 留给后续单据

- **去掉 `accountStore.bizs`**（2026-10-13 定）：二期合入 v1.9.x 后从主干另开一单。34 个文件、69 处读取多为「为空即非业务」判断，逐处换成 `useWhereAmI()` 的 `whereAmI` / `getBizsId()`，换完删除 `accountStore.bizs`、`business-interceptor.ts` 的写入与 `router/index.ts` 的 `afterEach` 清零。

### 验证

- `eslint` / `stylelint` / `prettier`：改动文件通过（`listener-table.vue`、`rerun.vue` 的存量类名加了 `stylelint-disable-next-line`）
- `tsc --noEmit`：与 `767bdd242` 对比无新增错误
- `npm run build`：通过
