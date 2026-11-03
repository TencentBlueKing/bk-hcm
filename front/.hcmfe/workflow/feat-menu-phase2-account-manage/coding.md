# Coding — feat-menu-phase2-account-manage

需求文档：仓库根 `docs/reqs/二期账号管理菜单.md`（TAPD [#1069995598138702746](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702746)，验收以其 AC-001～AC-013、AC-P01、AC-S01 为准）。拆单与 commit 顺序见 `docs/reqs/二期拆单总览.md`；菜单机制见 `.hcmfe/docs/modules/menu-route.md`。

**目标**：交付独立的「账号管理」模块——账号列表、账号详情（仅基本信息页签，可编辑）、账号登记，以及全部账号旧地址的转址表，并**本单即上线**（用户 2026-10-10 决定：每单上线即用新地址）：新路由挂到 ③ 的资源运营根下，撤掉全部旧账号路由，旧地址转址到新页面，「账号管理」加入资源运营侧栏（③ 只放了已有最终页面的项），站内指向账号页面的链接本单全部改写。首批 PR 共 7 个 commit：① 布局 → ② 视图权限 → ③ 资源纳管 + 资源运营翻转 → ④ 回收管理 / 操作记录 → **⑤ 本单** → ⑥ 资源详情页迁移 → ⑦ 账单迁址与清理。

## 跨单约定

| 项 | 约定 |
|----|------|
| 菜单符号 | 自 ① 起 `src/constants/menu-symbol.ts` 为 PR #1897 的 `Symbol` 常量（① 负责转换与消费方适配）。本单新增 `MENU_RESOURCE_ACCOUNT_MANAGEMENT = Symbol('menu_resource_account_management')`、`MENU_RESOURCE_ACCOUNT_CHECK_IN = Symbol('menu_resource_account_check_in')`、`MENU_RESOURCE_ACCOUNT_DETAIL = Symbol('menu_resource_account_detail')`、`MENU_RESOURCE_ACCOUNT_DETAIL_BASIC = Symbol('menu_resource_account_detail_basic')`（PR 工作台版为 `MENU_SERVICE_ACCOUNT_*`，取同样的命名与写法） |
| 模块位置 | `src/views/account-manage/`（与 PR 同位，三期迁工作台时目录不动，只改路由 owner 与地址）。目录内文件按 `fe-conventions` §6 用 kebab-case，不沿用 PR 的 `accountmanage/`、`accountInfo/` |
| 导出 | `src/views/account-manage/route-config.ts` 导出 `accountManageRsc`（相对路径、`name` 用上表符号、`meta` 用 `...new Meta({ owner: MENU_RESOURCE, activeKey, title, menu.relative, auth.view, layout })`）与 `accountManageLegacy`（绝对路径旧地址转址，无 component、无 name） |
| 依赖 ① 布局 | 新布局读取 `owner`、`activeKey`、`menu.relative`、`layout.breadcrumb.show`、`title`。② 不改 `Meta` 构造函数的面包屑默认值，本单每条路由显式写 `layout.breadcrumb.show` |
| 依赖 ② 视图权限 | 已对齐 ② 的 coding.md：`Auth.view` 改为 `IAuthSign`（直接接 `AUTH_*` symbol）；本单在 ② 新建的 `src/constants/view-auth.ts` 加 `{ viewId: MENU_RESOURCE_ACCOUNT_MANAGEMENT, authSign: { type: AUTH_FIND_ACCOUNT }, hideMenu: true }`（无权限隐藏菜单）；列表、详情路由声明 `auth.view`（直接访问显示框架内无权限页）。旧 `/resource/account` 的 `pageAuthData.path` 由 ② 删除并改为路由 `auth.view`，本单不处理 |
| 依赖 ③ 资源纳管 | 已对齐 ③ 的 coding.md：资源列表路由 `MENU_RESOURCE_RESOURCE_LIST`（`manage/:resourceType`），按 `query.accountId` 选中账号，识别 `query.showAccountStatus=1` 时自动打开账号状态侧滑（③ §3）；③ 已把 `/resource` 翻成资源运营根，侧栏只有资源纳管与云账号管理三项（④ 加回收管理、操作记录） |
| 本单注册与撤旧 | `views/index.ts` 的 `resourceViews` 加 `...accountManageRsc`；`router/index.ts` 在 catch-all 前注册 `...accountManageLegacy`。同一 commit 撤掉：`router/module/resource.ts` 中旧 `/resource/account` 列表（与新列表同路径）与旧 `/resource/resource` 下的 `account*` 子路由；`router/module/resource-inside.ts`（只有 `accountAdd` / `accountDetail` 两条）与 `router/module/service-inside.ts`（只有 `applyAccount`）整个删除，`router/index.ts` / `views/index.ts` 去掉对应引用 |
| 加入菜单 | `common/menu-service.ts` 资源运营菜单树「云资源」组插入 `{ id: MENU_RESOURCE_ACCOUNT_MANAGEMENT, i18n: '账号管理', icon: 'bkhcm-icon-account-manage', group: '云资源', route: getMenuRoute(resourceViews, MENU_RESOURCE_ACCOUNT_MANAGEMENT) }`（位置：资源纳管 → 账号管理 → 回收管理），删掉 ③ 留的 TODO |
| 本单改的链接 | 撤掉旧路由后按 `'accountDetail'` / `'accountAdd'` / `'applyAccount'` name 跳转会报 No match，必须本单改；按路径的旧链接有转址兜底，也一并改（见改动点 6） |

## 执行顺序

1. 共享符号与权限定义（改动点 1）
2. `RenderDetailEdit` 增加 `authSign`（改动点 2）
3. 模块页面：列表、详情容器、基本信息（改动点 3、4）
4. `route-config.ts`：页面路由 + 旧地址转址（改动点 5）
5. `view-auth.ts` 预鉴权登记（改动点 1 末条）
6. 注册与撤旧、加入菜单（跨单约定）
7. 站内链接改写与旧文件删除（改动点 6）
8. `bkdevbuddy_lint` + `npm run build` + `tsc` 对比；浏览器冒烟

> 1 → 2 → 3：页面依赖新增的权限符号和组件属性；4、5 依赖 1 的菜单符号与 3 的组件路径；7 依赖 6（旧 name 撤掉后才能确认引用清零）。

## 改动点

### 1. 菜单符号、账号权限定义、预鉴权

**文件**：`src/constants/menu-symbol.ts`、`src/constants/auth-symbols.ts`、`src/common/auth-service.ts`、`src/constants/view-auth.ts`

- `menu-symbol.ts` 加上表 4 个 `Symbol`。
- `auth-symbols.ts` 新增 `AUTH_DELETE_ACCOUNT = Symbol.for('auth_delete_account')`（取自 PR；后端 IAM 已有 `account_delete` 动作）。
- `auth-service.ts` 只改两条，不整文件取 PR（② 明确这两条不在 ②）：
  - `AUTH_UPDATE_ACCOUNT`：HEAD 误写为 `action: 'import'` 且无 `transform`，改为 `action: 'update'` + `transform: basicTransform(definition, { resource_id: relation[0] })`（与 PR 一致）。全仓当前无调用方（已 Grep），修正不影响现网。
  - 新增 `AUTH_DELETE_ACCOUNT`：`{ id: 'account_delete', action: 'delete', resourceType: 'account', transform 同上 }`。
- `view-auth.ts` 加跨单约定中的一条。启动批量鉴权多一项，无可见变化。

### 2. `RenderDetailEdit` 支持按钮权限

**文件**：`src/components/RenderDetailEdit/index.tsx`

只取 PR 的两段：新增可选 prop `authSign`（`IAuthSign | IAuthSign[]`，默认 `undefined`）；编辑图标渲染抽成 `renderEditIcon`，有 `authSign` 时用 `hcm-auth` 包住文字按钮、无权限置灰。未传 `authSign` 的调用方渲染不变。保留 HEAD 在 fc622471df 中对该文件的改动。

### 3. 账号列表

**文件**：`src/views/account-manage/list.vue`

以 **HEAD** 的 `views/resource/accountmanage/index.vue` 为底（单一事实来源，含三类账号搜索、列、业务筛选、分页），叠加：

- 删除 `useVerify`、`permission-dialog` 及相关解构（`fe-deprecated` §4，现网「编辑」因此永远不可用）。
- 名称列点击与「编辑」都跳详情：`routerAction.redirect({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId: row.id } }, { history: true })`，不再用 `router.push` 和 `query.isDetail`。
- 操作列：「编辑」包 `<hcm-auth :sign="{ type: AUTH_UPDATE_ACCOUNT, relation: [row.id] }" tag="span" v-slot="{ noPerm }">`，「删除」包 `AUTH_DELETE_ACCOUNT` 同形态（PR 只留删除，本单按 F-002 两个都留）。`HcmAuth` 内部合批，整页只发一次鉴权请求。
- 删除流程保持 HEAD：先 `accountDeleteValidate`，失败时由 http 层全局错误提示展示后端原因、不打开确认框；通过后二次确认 → `accountDelete` → 刷新列表。
- **不加** `type neq resource` 过滤，搜索「账号类型」保留三类；在 `searchData` 旁加一行注释说明三期改为不展示资源账号。
- 外层取 PR 的独立页样式（`.account-manage-page` 留白 + `.account-manage-content` 白底容器），因为列表不再嵌在资源接入的页签里。

### 4. 账号详情（单页签容器 + 基本信息）

**文件**：`src/views/account-manage/details/index.tsx`、`src/views/account-manage/details/index.scss`、`src/views/account-manage/details/basic.tsx`、`src/views/account-manage/details/basic.scss`

- `details/index.tsx`：取 PR `accountInfo/index.tsx`，`card-grid` 的 `Tab` 只有「基本信息」一个面板，内含 `<RouterView />`。根节点沿用 PR 的 `page-container`：① 的 `main-content.vue` 已为它统一提供 24px 留白与 `height: 100%`。
- `details/basic.tsx`：以 **HEAD** 的 `views/resource/accountmanage/account-detail.tsx` 为底，叠加 PR 的改动：
  - `accountId = computed(() => route.params.accountId)`，`getDetail`、`watch` 都改读它；
  - 「编辑账号」弹窗 PATCH 用 `accountId.value`，不再用 `useResourceAccountStore().resourceAccount?.id`（HEAD 从列表进详情时 store 为空，会请求 `/accounts/undefined`）；
  - 删除 `DetailHeader` 与「同步」按钮（与原资源接入「基本信息」页签一致；资源账号同步在 ③ 的账号状态侧滑里）、删除 `isDetail` / `hideEdit={!!isDetail}` 和 `route.path.includes(...)` 旧路径判断；这是 `DetailHeader` 的最后一个使用方（其余 19 个已由 ① 修订改用全局面包屑），删后 `rg "common/header/detail-header" src` 为空即删除 `views/resource/resource-manage/common/header/detail-header.tsx` 与 `detail-header.scss`；
  - 名称字段传 `authSign={accountEditSign}`，「账号归属」「密钥信息」两个「编辑」入口包 `hcm-auth`（`AUTH_UPDATE_ACCOUNT` + `relation: [accountId]`）；
  - 详情加载后 `useBreadcrumb().setTitle(`${t('账号详情')}：ID（${accountId}）`)`；
  - `useSecretExtension` 改用 `@/views/resource-manage/account/account-import/components/account-form/use-secret-extension` 绝对路径（③ 已按 PR 目录结构把账号接入模块移到 `views/resource-manage/account/` 并改名为 `account-import`，2026-10-10）。
- **账号不存在**（F-003）：`getDetail` 抛错或返回无数据时显示 `bk-exception`（404，「账号不存在」）+「返回账号列表」按钮（`routerAction.redirect({ name: MENU_RESOURCE_ACCOUNT_MANAGEMENT })`）。后端对不存在 ID 的返回码 待核实（两种情况都按不存在处理）。
- 「账号归属」「密钥信息」的「编辑」无权限时置灰（`is-disabled` 类，不响应点击），替代 PR 的行内样式。
- `basic.scss` 取 HEAD `account-detail.scss`，删掉已无用的 `.h16`（同 PR）；类名 `formItem-cls` 改为 `form-item-cls`（新文件过 stylelint 的 kebab-case 规则，tsx 同步改）。
- 负责人选择器保留 HEAD 的 `hcm-form-user`，`useSecretExtension` 用 `account-import` 新路径：PR 的 `account-detail.tsx` 在这两处是旧代码，不取。

### 5. 路由与旧地址转址

**文件**：`src/views/account-manage/route-config.ts`

```ts
const accountManageRsc: RouteRecordRaw[] = [
  { name: MENU_RESOURCE_ACCOUNT_MANAGEMENT, path: 'account', component: () => import('@/views/account-manage/list.vue'),
    meta: { ...new Meta({ owner: MENU_RESOURCE, title: '账号管理', activeKey: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
      auth: { view: { type: AUTH_FIND_ACCOUNT } }, layout: { breadcrumb: { show: true } } }) } },
  { name: MENU_RESOURCE_ACCOUNT_CHECK_IN, path: 'account/check-in',
    component: () => import('@/views/service/service-apply/account-add/index'),
    meta: { ...new Meta({ owner: MENU_RESOURCE, title: '账号登记', activeKey: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
      menu: { relative: MENU_RESOURCE_ACCOUNT_MANAGEMENT }, layout: { breadcrumb: { show: false } } }) } },
  { name: MENU_RESOURCE_ACCOUNT_DETAIL, path: 'account/details/:accountId',
    redirect: { name: MENU_RESOURCE_ACCOUNT_DETAIL_BASIC },
    component: () => import('@/views/account-manage/details/index'), meta: detailMeta,
    children: [{ name: MENU_RESOURCE_ACCOUNT_DETAIL_BASIC, path: 'basic',
      component: () => import('@/views/account-manage/details/basic'), meta: detailMeta }] },
];
// detailMeta = new Meta({ owner, title: '账号详情', activeKey: 列表, auth.view 同列表, menu.relative: 列表, breadcrumb.show: true })
```

- 详情地址对齐 PR：`/resource/account/details/<ID>` 自动跳到 `/resource/account/details/<ID>/basic`（对象式 redirect 带上 params 与 query）。站内跳详情一律按 `MENU_RESOURCE_ACCOUNT_DETAIL` 跳。
- **子路由必须重复 `auth.view` / `owner` / `activeKey`**：vue-router 按 matched 顺序浅合并 meta，子路由 `new Meta()` 自带的空 `auth: {}` 会覆盖父路由的 `auth`。
- 账号登记直接复用 `views/service/service-apply/account-add`（与 PR 相同，不复制），提交后跳单据管理的逻辑不变；不加查看权限（现网也没有）。
- `accountManageLegacy`（绝对路径；除被搬进 params 的 `accountId` / `id` 和已废弃的 `isDetail` 外，其余 query 一律保留）：

| 旧地址 | 新目标 |
|--------|--------|
| `/resource/resource/account`、`/resource/resource/account/resource`（原「账号信息」页签、原「资源状态」） | `toAccountStatus`：有 `accountId` → `{ name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'host' }, query: { ...其余, accountId, showAccountStatus: '1' } }`；无 → `{ name: MENU_RESOURCE_ACCOUNT_MANAGEMENT, query }` |
| `/resource/resource/account/detail`、`/resource/resource/account/manage`、`/resource/account/detail` | `toDetail`：取 `query.accountId ?? query.id` → `{ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId }, query: 其余 }`（再跳 `/basic`）；取不到 → 账号列表 |
| `/resource/account/add`、`/service/service-apply/account-add` | `{ name: MENU_RESOURCE_ACCOUNT_CHECK_IN }`（对象式 redirect，query 默认保留） |

转址目标与权限无关，用 `redirect`（函数形式也安全）。query 值可能是数组，取第一个非空值。

### 6. 站内链接改写与旧文件删除（本单执行）

**文件**：`src/views/resource/resource-manage/children/detail/{vpc,subnet,drive}-detail.vue`、`src/views/resource/resource-manage/children/components/{ip/ip-info,route/route-info,network-interface/network-interface-info,network-interface/network-interface-info-gcp,network-interface/network-interface-info-huawei}.vue`、`src/hooks/useRouteLinkBtn.tsx`、`src/views/resource/accountmanage/*`、`src/views/resource/resource-manage/accountInfo/*`

- **资源详情里的账号链接**（8 个文件）：`/#/resource/account/detail/?accountId=${val}&id=${val}` 改为 `/#/resource/account/details/${val}`（`link` 渲染只接受字符串地址；能改走 `routerAction.open({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId } })` 的改走 name）。原计划由 ⑥ 经链接工具 `resource-detail-link.ts` 中转、⑦ 再切目标，随「本单即上线」取消：本单直接改到最终目标，⑥ 不再处理账号链接。
- **`useRouteLinkBtn`**：账号分支 `name: 'accountDetail'`（资源侧、业务侧两处）改为提前返回 `router.push({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId } })`，不再传 `query.accountId` / `query.id`；资源分支的 `'resourceDetail'` 归 ⑥。
- **`applyAccount` 引用**：`service-apply/components/resource-card/index.vue` 只按字符串比较 `item.routeName === 'applyAccount'` 判断权限类型，数据源 `basic-resource/index.vue` 本就无人引用（③ 前已核实），二者不改；删前 `rg "applyAccount" src` 确认没有按 name 跳转的调用。
- **删除旧文件**（本分支不留废弃副本，无引用即删；已逐个 `rg` 确认无外部引用）：`views/resource/accountmanage/` 整个目录（列表、详情、旧直接录入页 `account-add`）、`views/resource/resource-manage/accountInfo/` 整个目录（含 `component/resourceStatus`——③ 的账号状态侧滑是独立副本，不复用——与 `component/usersList`）、`views/resource/resource-manage/common/header/detail-header.tsx` / `.scss`（最后一个使用方是旧账号详情）。`resource-manage.vue` 不引用它们，随 ⑦ 删除。PR 合入时若 squash，git 能把旧文件删除和新文件新增识别为重命名。
- 旧 `/resource/resource` 记录撤掉 `account*` 子路由后只剩 ③ 的空路径转址子路由，按 ⑦ 计划改为独立转址记录；本单不提前动。

## 从 PR #1897 取 / 不取

| 内容 | 处理 | 说明 |
|------|------|------|
| `views/account-manage/` 模块、列表 / 登记 / 详情→basic 路由结构、登记复用服务请求组件 | 取 | owner 改为 `MENU_RESOURCE`，地址挂在 `/resource` 下 |
| 详情：params 取 ID、PATCH 用 accountId、去头部与同步、HcmAuth 编辑、面包屑标题 | 取 | 叠加在 HEAD 文件上 |
| `RenderDetailEdit` 的 `authSign`、`AUTH_DELETE_ACCOUNT`、`AUTH_UPDATE_ACCOUNT` 定义修正 | 取 | 只取相关段落 |
| 列表去 `useVerify`、删除包 HcmAuth、按名称跳详情、独立页外层样式 | 取 | 跳转改用 `routerAction` |
| 60dac3a85a 列表过滤资源账号、账号类型去掉资源账号 | 不取 | 三期 |
| 列表去掉「编辑」按钮 | 不取 | F-002.4 要求保留 |
| 工作台位置、`MENU_SERVICE_ACCOUNT_*`、「公共事务」分组 | 不取 | 三期 |
| `route-config` 内写相对路径的 `service-apply/account-add` 转址 | 不取 | 改为 `accountManageLegacy` 绝对路径 |
| PR 旧地址转址（丢 query；缺 ID 时生成 `/details/undefined`；`/resource/resource/account` 一律去列表） | 不取 | 本单保留 query、缺 ID 回列表、带账号去资源纳管并打开账号状态侧滑 |
| 不注册 `/resource/account/add`（落 404） | 不取 | 本单转到账号登记 |
| 账号登记页样式改动（`.page-container`、标题间距、scss 重排） | 不取 | 依赖 PR 布局；本期登记页外观与现网一致 |
| PR 的 `account-detail.tsx`、`auth-service.ts`、`RenderDetailEdit` 整文件 | 不取 | 基于旧基线，会回退 HEAD 改动 |
| `@/views/resource-manage/...` 导入路径 | 不取 | PR 搬目录后的路径，HEAD 不存在 |

## HEAD 新增能力必须保留

- fc622471df：「编辑账号」弹窗负责人用 `hcm-form-user`。PR 的 `account-detail.tsx` 仍 `import MemberSelect`，**不能拿 PR 文件做底**。
- fc622471df：`RenderDetailEdit`、`components/form/user.vue` 的人员选择器改动，只叠加 `authSign` 段落。
- 30fbd8552a、fae68a4196：`auth-service.ts` / `components/auth/auth.vue` 的云账号管理权限调整与弹窗实例去重，只改两条账号定义。
- 01bca385c3：`use-breadcrumb.ts` 的一期改动不动。
- HEAD 列表的搜索项、列（所属业务、管理业务、业务筛选）、排序、精确匹配，原样保留。
- 账号登记组件自 PR 基线后 HEAD 无改动（已核对 `git log`），原样复用。

## 三期待办

- 账号管理迁至工作台：`owner` → `MENU_SERVICE`，地址改为 `/service/account*`，菜单进「公共事务」；本单的 `/resource/account*` 新地址加入转址，本单 `accountManageLegacy` 的目标同步改指工作台（避免二跳）。
- 列表改为 PR 形态：搜索「账号类型」去掉资源账号、请求追加 `type neq resource`（对齐 60dac3a85a）；同时确认操作列是否按 PR 只留「删除」。
- 登记页按工作台布局取 PR 的样式改动。

## 风险与验证

| 风险 | 怎么验 |
|------|--------|
| 旧路由与新路由路径冲突：`/resource/account` 与旧列表同路径；`/resource/account/{detail,add}`、`/service/service-apply/account-add`、`/resource/resource/account/*` 原为旧路由 | 同一 commit 撤旧再注册；`router.getRoutes()` 无同路径重复；逐条验 AC-007～AC-011、AC-013，地址与 query 符合预期 |
| `/resource/resource/account*` 与 ③ 挂在旧 `/resource/resource` 记录上的空路径转址并存 | 撤掉 `account*` 子路由后，本单的绝对路径记录更具体、优先匹配；验 AC-008、AC-010 进资源纳管、选中账号并自动打开侧滑 |
| 旧 name 残留 | `rg "'accountDetail'\|'accountAdd'" src` 为 0；`'applyAccount'` 只剩 `resource-card` 的字符串比较 |
| 布局 | 新页面不带 ③ 的 `legacyResourceAccess`，不显示旧云账号树（AC-001）；侧栏「账号管理」高亮、顶栏高亮资源运营；详情页面包屑可返回列表 |
| 子路由 meta 覆盖导致详情不拦截 | 无 `account_find` 的账号直接开 `/resource/account/details/<ID>/basic`，应显示框架内无权限页（AC-012） |
| 详情加了查看拦截后，业务侧主机详情里的账号链接对无 `account_find` 的用户变为无权限页（旧详情页无页面级拦截） | 有意变化（用户决定），本单自测时确认无权限页可申请 |
| `AUTH_UPDATE_ACCOUNT` / `AUTH_DELETE_ACCOUNT` 按实例鉴权 | 分别用有 / 无编辑、删除权限的账号验 AC-003、AC-004、AC-005、AC-S01；无权限时按钮置灰，点击弹申请 |
| 列表每行两个权限标记 | 网络面板确认整页只有一次鉴权请求；AC-P01 对比打开耗时 |

## 已确认（2026-10-08）

1. **详情地址**：对齐 PR，`/resource/account/details/<ID>` 自动跳到 `/basic`（需求 F-003、AC-003、AC-009 已同步）。
2. **原「账号信息」页签与「资源状态」旧地址**：不带 `accountId` 进账号列表；带 `accountId` 进资源纳管主机列表，带上 `accountId` 与 `showAccountStatus=1`，由 ③ 自动打开账号状态侧滑。「基本信息」「用户列表」旧地址与 `/resource/account/detail` 进账号详情。其余 query 保留（需求 F-005、AC-008、AC-010 已同步）。
3. **菜单符号**：自 ① 起为 PR #1897 的 `Symbol` 常量，本单新增 4 个 `Symbol`。
4. **提交顺序**：「资源详情页迁移」并入首批，作为 ⑥；⑦ 收缩为账单迁址与清理。
5. **2026-10-10 调整**：每单上线即用新地址。本单注册路由、撤旧、把账号管理加入侧栏（③ 不放临时入口）；资源详情页里的账号详情链接与 `useRouteLinkBtn` 的账号分支改由本单直接改到最终目标（取代原「⑥ 经链接工具中转、⑦ 切目标」的协调结论）。
