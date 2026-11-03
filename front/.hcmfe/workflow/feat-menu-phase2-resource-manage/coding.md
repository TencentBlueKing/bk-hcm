# Coding — feat-menu-phase2-resource-manage

需求文档：仓库根 `docs/reqs/二期资源纳管主体.md`（TAPD [#1069995598138702496](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598138702496)，验收以文中 AC-001 ~ AC-020、AC-008b、AC-S01、AC-S02、AC-P01 为准；AC-013 随 ④⑤ 验收，AC-014 ~ AC-020 为第 8 轮由重组单移入）。设计参照 PR #1897（base `8fea9924df`，head `df8865b539`，只看 `25696ea07b`、`5c513eb559`、`60dac3a85a` 中非筛选部分；`2e535a0cd1` 全是筛选与收藏，不取）。

**目标**：在 `src/views/resource-manage/` 新建「资源纳管」模块——路由（含旧地址转址）、入口两栏（云账号树 + 内容）、资源列表（去掉四个顶部页签，资源类型进地址）、账号摘要条、只读「资源账号状态」侧滑（支持地址参数自动打开）、新建页挂到 `manage/service-apply/*`；**并在本单即上线新地址**（用户 2026-10-10 决定）：`/resource` 翻转为「资源运营」一级菜单根（默认进资源纳管），顶栏「资源接入」改为只跳转的入口，资源运营侧栏按最终菜单树的结构，只放已有最终页面的项（资源纳管、云账号管理三项；账号管理、回收管理、操作记录由 ④⑤ 上线时加入，用户 2026-10-10 决定：除资源纳管外均为无云账号树的独立页面，不用原页面临时顶替），账单挂到资源运营下（地址仍为 `/bill/*`，迁址留给 ⑦），兼容全部旧资源管理地址，改写站内指向资源管理的旧链接，资源纳管页显示迁移提示条。

首批 commit 顺序：① 布局 → ② 视图权限 → ③ 本单（资源纳管 + 资源运营翻转）→ ④ 回收管理 / 操作记录 → ⑤ 账号管理 → ⑥ 资源详情页迁移 → ⑦ 账单迁址与清理。③～⑥ 每单上线即注册新地址、兼容旧地址、改站内链接，并把自己的菜单项加入资源运营侧栏（用户 2026-10-10 决定）。

## 跨单约定

| 项 | 约定 |
|----|------|
| 菜单符号 | 自 ① 起 `src/constants/menu-symbol.ts` 为 PR #1897 的 `Symbol`。本单新增 `MENU_RESOURCE_MANAGE = Symbol('menu_resource_manage')`、`MENU_RESOURCE_RESOURCE_LIST = Symbol('menu_resource_resource_list')`、`MENU_RESOURCE_ACCESS = Symbol('menu_resource_access')`（顶栏「资源接入」菜单项 id，不对应路由，§10）；复用 `MENU_RESOURCE`（本单起指资源运营根）、`MENU_RESOURCE_HOST_APPLY / VPC_APPLY / DISK_APPLY / SUBNET_APPLY / LOAD_BALANCER_APPLY`。删除 `MENU_BILL`（不再是一级父路由，§11）。`MENU_RESOURCE_RESOURCE_MANAGEMENT` 保留在旧 `/resource/resource` 记录上（④⑤ 迁走其子页面前仍在），站内按该 name 的跳转本单全部改掉，⑦ 删除常量。本单所有 name 比较都直接用常量（`===`、`includes`），不做 `toString()` / 字符串拼接 |
| ① 依赖 | `components/layout/*`、`breadcrumb.vue`（自身不判断 `display`，由调用方决定渲染；返回时自动带 `?bizs=`，支持 `back-config`）、`Meta.layout.breadcrumb`；`router/utils/history-storage.ts` 已按 PR 支持 Symbol name 的序列化（2026-10-10 核实）；① 离开一级菜单时 `resourceAccountStore.clear()`，本单入口卸载时再清一次（§6），不冲突——**两处清空都是过渡**：PR 的新模块不用 `useResourceAccountStore`（账号 / 云厂商在地址里），无需清空；本单为兼容现网读 store 的列表组件暂沿用 store，由「资源列表筛选条件写入地址」单切换到 PR 版云账号树时一并删除（见「对照 PR 的旧文件与过渡逻辑退场」）。① 修订（2026-10-09）已去掉 5 个申请页与 11 个详情组件的 `DetailHeader`（申请页改 `sticky-bottom-container`，详情页 `setTitle` + `Teleport` 到 `#breadcrumbExtra`），本单入口右栏的面包屑即它们的标题 / 返回 / 操作按钮所在 |
| `Meta` 标题 | `Meta` 构造函数执行 `this.menu.i18n = this.title`，用 `new Meta` 时 `menu.i18n` 会被 `title` 覆盖（未传即空串）；本单路由一律写 `title`，`menu` 只放 `relative` |
| ② 权限 | 资源纳管**不加**页面级 `auth.view`（PR 的 `AUTH_FIND_IAAS_RESOURCE` 删掉），不进 `constants/view-auth.ts` 预鉴权；页内按钮沿用 `hcm-auth`。理由（用户 2026-10-10 确认）：资源纳管入口放开，权限控制在数据层——云账号树只列有权限的账号，资源列表按账号实例过滤。PR 的 `auth.view` 不带实例，后端走 `AuthorizeAny`（任一账号有 `resource_find` 即通过），只对零权限用户显示引导页（`permission.vue` 中 `AUTH_FIND_IAAS_RESOURCE` 文案）；不加后零权限用户看到空账号树，与现网一致 |
| 模块导出与注册 | `src/views/resource-manage/route-config.ts`：`export default resourceManageRoutes`（相对路径）+ `export const resourceManageLegacy`（5 个旧申请页地址的绝对路径转址，不带 name）+ `export const toResourceList(to)`（旧列表地址转址函数，§7）。**本单即注册**：`views/index.ts` 汇入 `resourceViews`（经 `injectStatusComponents`），`router/index.ts` 挂到资源运营根 `/resource` 下，`resourceManageLegacy` 放在 catch-all 404 之前（§10） |
| ⑤ 账号管理 | ⑤ 的旧地址转址（`/resource/resource/account?accountId=…`、`/resource/resource/account/resource?accountId=…`）落到 `{ name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'host' }, query: { accountId, showAccountStatus: '1' } }`，由本单 §3 自动打开侧滑；账号接入模块（原 `createAccount`）由本单按 PR 目录结构移到 `views/resource-manage/account/account-import/**` 并按 kebab-case 改名（`git mv`，内容为 HEAD 最新版，含 PR 副本缺的 `fc622471df`、`db6f42fc86`；原位不留副本，见 §2b），⑤ 如需接入弹窗从 `account-import` 引、`useSecretExtension` 从 `account-import/components/account-form/use-secret-extension` 引 |
| ⑥ 详情迁移 | ⑥ 往本单 `route-config.ts` 的 `MENU_RESOURCE_MANAGE.children` **追加** `:resourceType/details/:id`（`MENU_RESOURCE_DETAIL`），并改写所有进入详情的链接；本单列表链接保持现状（仍按 `'resourceDetail'` name 进旧详情页，旧详情在 ⑥ 前继续用旧资源接入过渡布局，§12）。本单入口为非列表子路由渲染面包屑（§6），⑥ 的详情路由 meta 需写 `title`（见「`Meta` 标题」），`menu.relative` 指向 `MENU_RESOURCE_RESOURCE_LIST` |
| ⑦ | 原「组装」内容中，注册新模块、资源运营翻转、菜单树、资源接入入口、提示条、资源管理相关站内链接**已移入本单**；⑦ 只剩：账单迁 `/resource/bill/*` 与 `/bill/*` 旧地址、`useWhereAmI` 一行、删除旧资源接入过渡布局（`main-content.vue` 中标 `TODO(菜单整合二期·组装)` 的代码）与旧文件、删除 `MENU_RESOURCE_RESOURCE_MANAGEMENT`、全量终扫 |
| ④⑤ | 旧 `/resource/resource` 记录保留到 ④⑤ 把 `record` / `recycle` / `account*` 子页面迁走为止（本单只给它加空路径子路由转址，§7）；④⑤ 各自注册新页面、转址这些子地址、把自己的菜单项加入资源运营侧栏（§10）。本单期间这些原页面只能经旧地址打开，侧栏无入口（同一 PR 合入，不单独发布） |
| 不动 | `views/resource/resource-manage/children/manage/*-manage.vue`、详情页不移动（PR 同样留在原位）；旧账号页、`resource-manage.vue`、旧云账号树、`router/module/*` 本单不删，按「对照 PR 的旧文件与过渡逻辑退场」表分批删除 |

## 执行顺序

1. 菜单符号（§1）；账号接入模块移动改名为 `account-import` 并改 5 个引用方（§2b），先跑一次 build
2. 账号上下文 hook + 云账号树副本（§2）→ 状态侧滑（§4）→ 摘要条（§3，含自动打开）
3. 资源列表 `resource-list.vue`（§5）
4. 入口 `resource-entry.vue`（§6）+ 迁移提示条（§13）
5. `route-config.ts`（§7）
6. 共享组件兼容修补（§8）
7. 资源运营翻转与注册（§10）、账单挂到资源运营下（§11）、旧资源接入过渡调整（§12）
8. 站内旧链接改写（§9）
9. `bkdevbuddy_lint` + `npm run build` + `tsc` 对比；`rg` 终扫（风险表）

> 排序依据：7 依赖 1～5（注册时新模块须完整）；8 依赖 7（新 name 已注册才能改链接）。

## 1. 菜单符号

**文件** `src/constants/menu-symbol.ts`

在 `MENU_RESOURCE` 一段追加三个 `Symbol` 常量、删除 `MENU_BILL`（见跨单约定；`MENU_BILL` 的引用随 §10 / §11 一并去掉），其余不动。

## 2. 账号上下文与云账号树

**文件** `src/views/resource-manage/account/use-resource-account.ts`（新）、`src/views/resource-manage/account/vendor-group/index.vue`（新）

- **本单暂不移植 PR 版 vendor-group（过渡，非放弃）**：PR 版是目标形态——按 `route.query.accountId / vendor` 驱动、不写 `useResourceAccountStore`、无需离开时清空。切换它必须同时改掉所有读 store 的消费方：9 个 `*-manage.vue` 经 `use-filter.ts` 的账号 / 云厂商筛选、`security-manage.vue`、VPC 申请页、新建按钮的 `hcm-auth`，以及 PR 漏改仍读 store 的 `HostOperations.tsx`（按钮鉴权 relation）、`template-dialog`（`account_id`）、`use-selection.ts`。PR 是把这些与 `useFilterFromRoute` 一起改的，因此整块归「资源列表筛选条件写入地址」单（用户 2026-10-10 确认）。本单以 **HEAD 的 `views/resource/resource-manage/account/vendor-group/index.vue` 为底**复制一份作过渡；该单把它替换为 PR 版（目录不变，即 `views/resource-manage/account/vendor-group/`）。
- `use-resource-account.ts`：复制 HEAD `account/accountList/useResourceAccount.tsx`（`watchEffect` 拉账号详情写 `setResourceAccount`、监听 `route.query.accountId`、`onBeforeMount` 写 `setCurrentAccountSimpleInfo({ id })`），只改 `setAccountId`，去掉写死的 `/resource/resource/` 与白名单：
  - 当前路由是 `MENU_RESOURCE_RESOURCE_LIST` 或 5 个新建页之一：留在原路由只改 query，`routerAction.redirect({ query: { ...route.query, accountId: val || undefined, id: undefined } })`（现网 CLB 新建页切账号会跳回列表，这里统一留在原页，与 PR `navigate` 一致）；
  - 其它子路由（即 ⑥ 追加的详情页）：`routerAction.redirect({ name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: route.params.resourceType }, query: { accountId } }, { replace: true })`，对齐 PR「详情页切账号回对应资源类型列表」，且不依赖 ⑥ 的常量。
- `vendor-group/index.vue`：改用上面的 hook；账号接入弹窗 `AccountImport` 从 `../account-import`（§2b 迁移改名后的位置）引；`vendorProperty` 从同目录 `./vendor.plugin` 引：按 PR 目录结构，把 `vendor.plugin.ts`、`vendor-property.ts` 从 `views/resource/resource-manage/account/vendor-group/` `git mv` 到 `views/resource-manage/account/vendor-group/`（内容不变），旧云账号树 `index.vue` 改为从新位置绝对路径引用（⑦ 删旧 `index.vue` 时不牵连新模块）。三种选择仍写 store（`clear` / `setCurrentVendor` / `setCurrentAccountSimpleInfo`），「接入」按钮 `hcm-auth AUTH_IMPORT_ACCOUNT` + `?dialog=create_account` 不变，`router.push` 换 `routerAction.redirect`。云厂商不进地址（与现网一致；F-001 只要求账号进地址）。

## 2b. 账号接入模块按 PR 目录结构迁移并改名

**文件** `src/views/resource/resource-manage/account/createAccount/**` → `src/views/resource-manage/account/account-import/**`、`src/views/bill/account/create-account/create-first-account/index.tsx`、`src/views/bill/account/account-detail/first-level-account-detail/index.tsx`、`src/views/resource/accountmanage/account-detail.tsx`、`src/views/resource/resource-manage/account/vendor-group/index.vue`、`src/views/resource/resource-manage/account/accountList/index.tsx`

- **职责与命名**（用户 2026-10-10 确认）：该模块是把已有云账号接入 HCM（弹窗「云账号接入」：录入账号 → 资源同步 → 结果；入口权限 `AUTH_IMPORT_ACCOUNT` / `account_import`），不是新建账号。位置取 PR（`views/resource-manage/account/`），目录名改为 `account-import`，与权限点一致；内部目录与文件按 kebab-case 一并改名（`git mv`，内容为 HEAD 最新版，含 PR 副本缺的 `fc622471df` 人员选择器、`db6f42fc86` 腾讯云禁用提示）：

  | 原 | 新 |
  |---|---|
  | `createAccount/index.tsx`、`index.scss` | `account-import/index.tsx`、`index.scss` |
  | `components/accountForm/`（含 `useSecretExtension.tsx`） | `components/account-form/`（`use-secret-extension.tsx`） |
  | `components/accountResource/` | `components/account-resource/` |
  | `components/resultPage/` | `components/result-page/` |

  组件内部的相对 import 随改名同步；导出的组件 / hook 名（`useSecretExtension`、`ValidateStatus` 等）不改，只改路径。原位不留副本（PR 原位标 `@deprecated` 保留，本分支直接移走）。
- **改 5 个引用方的 import**（路径改为 `@/views/resource-manage/account/account-import...`）：账单录入一级账号页、一级账号详情、旧账号详情（`account-detail.tsx`，⑤ / ⑦ 退场前仍在用）引 `components/account-form/use-secret-extension`；旧云账号树 `vendor-group/index.vue` 与旧账号列表 `accountList/index.tsx`（相对路径 `../createAccount`）引模块本身。旧文件只改 import 路径，局部变量名不动（随 ⑦ 删除）。
- **新模块内的命名**：§2 的云账号树副本用 `AccountImport` / `isAccountImportDialogShow`，「接入」按钮写入的地址参数改为 `dialog: 'account_import'`。打开判定沿用现网「`dialog` 有值即打开」，旧书签带 `dialog=create_account` 经转址进来同样能打开，不需要兼容分支。
- 迁移后 `rg "createAccount|accountForm/useSecretExtension" src` 只剩 `secondary-account.ts` 的 `createAccountSecret`、三级账号页的 `handleCreateAccount` 等无关同名标识；`npm run build` 通过（引用方都在已注册路由上，构建能覆盖）。

## 3. 账号摘要条（含自动打开侧滑）

**文件** `src/views/resource-manage/account/account-summary.vue`（新，取 PR UI/样式）

- 数据源改为 `useResourceAccountStore`（不另发详情请求，兼顾 AC-P01）：名称取 `resourceAccount.name` → `currentAccountSimpleInfo.name` → `useAccountSelectorStore().authorizedResourceAccountList` 按 id 查（满足「详情失败只显示名称」）；只选云厂商时显示 `VendorMap[currentVendor]`（**PR 缺**，F-004.1）；都没选显示「全部账号」。
- 扩展字段：`headerExtensionMap` 原样搬自 HEAD `resource-manage.vue`，`resourceAccount.extension && !isOtherVendor` 时显示。
- `selectedAccountId` 有值才显示「查看账号状态」（AC-008），打开 §4 侧滑，传 `:account-id="selectedAccountId"`。
- **自动打开（F-005.5 / AC-008b）**：地址参数 `showAccountStatus=1`。
  ```ts
  watch(() => route.query.showAccountStatus, (flag) => {
    if (flag !== '1') return;
    if (route.query.accountId) isShowAccountStatus.value = true;
    const { showAccountStatus, ...query } = route.query;
    routerAction.redirect({ query }, { replace: true });
  }, { immediate: true });
  ```
  - 打开后立即用 `replace` 去掉参数：刷新、后退、切页签都不会再弹；`replace` 不新增历史记录，后退直接回到转址前的页面。
  - 只认 `accountId`，侧滑数据随 `selectedAccountId`（§2 hook 在 `onBeforeMount` 已按地址写入）；带参数但没有 `accountId` 时只去掉参数、不打开。
  - 放在摘要条里（侧滑状态的唯一持有者），只在列表页生效；⑤ 的转址目标固定是列表页。去参数只改 query，`resource-list` 不重建，侧滑状态保留；§5 的非法类型回退若同时触发，其 query 只带 `accountId`，同样会去掉参数。

## 4. 资源账号状态侧滑

**文件** `src/views/resource-manage/account/account-status/index.vue`、`resource-status.tsx`、`resource-status.scss`（新，取 PR，scss 来自 HEAD `accountInfo/component/resourceStatus/index.scss`）

PR 版基础上的修正：

- `bk-sideslider` 加 `render-directive="if"`：关闭即卸载内容，`useTimeoutPoll` 随 scope 销毁停轮询（bkui 默认 `show` 会让轮询在后台继续）。
- 账号详情改为「打开时」拉取：`watch([() => props.isShow, () => props.accountId])`，`isShow && accountId` 才请求（PR 是 accountId 一变就请求，侧滑没开也发）。
- `resource-status.tsx` 保持 PR 写法（`getList` 内现取 `props.accountId`，`watch(accountId)` 里 `reset()` + `getList()` + `resume()`），即 `resource.md` 记录的防闭包写法，满足 AC-007；URL 去掉已废弃的 `BK_HCM_AJAX_URL_PREFIX`（`http` 已配 `baseURL`）；请求失败时显示「资源状态加载失败」。
- 基本信息只读：账号名称、云厂商（`CloudType`）、账号 ID + 有值的 7 个扩展字段，不展示密钥（AC-S02），无编辑、无账号详情入口（AC-005）。
- 「同步资源」：`bk-pop-confirm` 二次确认 → `accountStore.accountSync`，成功提示逐字同需求（AC-006）；失败由 `http` 的 `globalError` 弹后端信息；与现网账号详情页一致，不加按钮级鉴权。

## 5. 资源列表

**文件** `src/views/resource-manage/resource-list.vue`（新）

骨架取 PR `resource-list.vue`：资源类型 `bk-tab`（`card-grid`）、「分配状态」下拉放 `#setting`、11 个资源组件映射（CLB = `views/load-balancer/entry-rsc.vue`，证书 = `views/business/cert-manager`）、EIP/子网/安全组/GCP 防火墙侧滑表单、`TemplateDialog`、购买/新建按钮、`provide('securityType' | 'isOtherVendor')`、flex 自适应样式。顶部放 `<AccountSummary />` 与 `sync_failed_reason` 告警。列表里的资源链接不改（仍走 `'resourceDetail'`，由 ⑥ 改写）。

**筛选沿用 HEAD 机制（Q-001 的落地方案）**——PR 版列表不给子组件传 `filter`，子组件改由 `use-filter-from-route` 读地址；本单不移植那套，把 HEAD `resource-manage.vue` 的筛选状态整体搬进来：

| 搬入内容（来自 HEAD `resource-manage.vue`） | 说明 |
|---|---|
| `filter` / `accountId` / `status` ref，`filterData()`，四个 watch（`accountId`、`status`、`resourceAccountStore.selectedAccountId`、`resourceAccountStore.currentVendor`） | 逐字搬；`currentVendor` 的 watch 保持非 immediate（入口卸载会 `clear()`，初值必为空，与现网一致） |
| `onMounted(getResourceAccountList)` → `accountStore.updateAccountList` | `use-filter.ts` 的「云账号」搜索项选项来自 `accountStore.accountList`，PR 删了这步，不搬搜索下拉为空 |
| 子组件 props：`:filter`、`:where-am-i="activeTab"`、`:is-resource-page`（`!accountStore.bizs`）、`v-model:is-form-data-changed`、侧滑表单 `:filter` | HEAD `*-manage.vue` 的 `defineProps` 依赖这些；`@route-done` 不再需要（类型在 params，不回写 `query.type`） |
| `isOtherVendor`、`tabs` 用 `resourceAccountStore.vendorInResourcePage` | 替换 PR 的 `route.query.vendor` + `accountSelectorStore` 推导 |
| 新建按钮 `hcm-auth { type: AUTH_CREATE_IAAS_RESOURCE, relation: [resourceAccountStore.resourceAccount?.id] }` | 与现网一致（AC-S01） |

地址与跳转：

- `activeTab = computed(() => route.params.resourceType || 'host')`。
- 切页签：`routerAction.redirect({ name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType }, query: { accountId: accountId || undefined } }, { replace: true })`——与现网一样只保留账号、replace；不取 PR 的「保留其余 query + push」。
- 非法 / 当前云厂商不支持的类型（F-001 异常）：`watch([tabs, activeTab], immediate)`，不在 `tabs` 内时按上式 replace 到 `tabs[0]`（风险 4）。
- 新建：取 PR `handleAdd`，申请页 name 映射为 `Record<string, symbol>`，`routerAction.redirect({ name, query: route.query }, { history: true })`，返回时 `useBack` 从历史栈取回原列表（带 `resourceType` 与账号），满足 AC-009。
- `filter` 是组件内状态，切页签只换 params、组件复用，分配状态跨页签保留，与现网一致。

## 6. 入口

**文件** `src/views/resource-manage/resource-entry.vue`（新）

取 PR 两栏骨架与样式（树 240px 固定宽、右侧 `flex: 1` 纵向：面包屑 + `resource-entry-content`（`flex: 1; overflow-y: auto`，无内边距）内的 `RouterView`）。内容区不加内边距：列表页用 `page-container`、详情页 `detail-content-wrap`、申请页 `sticky-bottom-container` 各自提供 24px，且吸底组件需要父容器有确定高度。

- **面包屑**：`showBreadcrumb = route.name !== MENU_RESOURCE_RESOURCE_LIST`，即非列表子路由（新建页、⑥ 的详情页）都渲染，放在右栏顶部、不横跨云账号树（同 PR `resource-entry.vue` 的 `!isListPage`）。新建页的 `DetailHeader` 已由 ① 修订（2026-10-09）去掉并改用 `sticky-bottom-container`，标题取路由 `title`；新建页路由补 `title`（沿用原标题栏文案：购买主机 / 购买VPC / 购买硬盘 / 新建子网 / 购买负载均衡）与 `menu.relative`。① 资源接入过渡布局里同样的右栏面包屑（`main-content.vue` 的 `legacyBreadcrumbRoutes`）本单起只剩旧详情与旧操作记录详情在用，④⑥ 迁走后由 ⑥ 删除。
- 模块内所有路由都写 `layout.breadcrumb.show: false`，只为让 `main-content.vue` 不在整页顶部再画一条；`breadcrumb.vue` 本身不看 `display`，入口直接 `<HcmBreadcrumb v-if="!isListPage" />` 即可（同 PR），不改 ① 的组件。
- `onBeforeUnmount(() => resourceAccountStore.clear())`：离开资源纳管（去账号管理 / 回收管理等）时清空；列表、新建、详情同在本入口下，互相切换不卸载，账号保持。

## 7. 路由

**文件** `src/views/resource-manage/route-config.ts`（新）

结构取 PR，`meta` 一律 `...new Meta({ owner: MENU_RESOURCE, activeKey: MENU_RESOURCE_MANAGE, title, menu: { relative }, layout: { breadcrumb: { show: false } } })`（标题写 `title`，见跨单约定「`Meta` 标题」）：

| name | path（相对 `/resource`） | component | title / relative |
|------|------|-----------|------|
| `MENU_RESOURCE_MANAGE` | `manage`，`redirect: { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'host' } }`（对象 redirect 保留 query） | `resource-entry.vue` | `'资源纳管'` |
| `MENU_RESOURCE_RESOURCE_LIST` | children `:resourceType` | `resource-list.vue` | `'资源列表'` / `MENU_RESOURCE_MANAGE` |
| `MENU_RESOURCE_HOST_APPLY` 等 5 个 | children `service-apply/{cvm,vpc,disk,subnet,clb}` | HEAD `views/service/service-apply/*` 原样 | 沿用原标题栏文案「购买主机」「购买VPC」「购买硬盘」「新建子网」「购买负载均衡」（与 ① 修订后旧路由一致，不用 PR 的「申请主机」等） / `MENU_RESOURCE_RESOURCE_LIST`（无站内历史时回列表） |

- 去掉 PR 的 `auth.view`；详情子路由由 ⑥ 追加到同一 `children`（`:resourceType/details/:id` 两段，与 `:resourceType` 不冲突）。
- CLB 新建页保留 HEAD 的 `isFilterAccount: true`（`components/account-selector` 据此只列腾讯云账号；PR 漏了），写在展开之外：`meta: { ...new Meta({...}), isFilterAccount: true }`。

旧地址（redirect 只依赖地址、不依赖权限，可用函数式）：

```ts
// 旧资源列表 /resource/resource?type=&accountId= → 新列表；挂到旧 /resource/resource 记录的空路径子路由（§12），
// 旧记录在 ④⑤ 迁走其子页面前仍在，不能另建同路径记录
export const toResourceList = (to: RouteLocationNormalized) => {
  const { type, ...query } = to.query;
  return { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: (type as string) || 'host' }, query };
};

// 5 个旧申请页地址：旧同名路由本单撤掉，转址到新申请页，参数保留
export const resourceManageLegacy: RouteRecordRaw[] = applyPairs.map(([slug, name]) => ({
  path: `/resource/service-apply/${slug}`,
  redirect: (to) => ({ name, query: to.query }),
}));
```

⑦（旧 `/resource/resource` 记录删除时）把 `toResourceList` 改为独立的 `{ path: '/resource/resource', redirect: toResourceList }` 记录。

## 8. 共享组件兼容修补

**文件** `src/components/account-selector/filter.class.ts`、`src/components/account-selector/index.vue`

两处「证书只列腾讯云账号」判定写的是 `isResourcePage && route.query.type === 'certs'`；新地址里类型在 params，证书页上传证书的账号下拉（`index-new.vue` 默认走 `filter.plugin` → `filter.class`）会列出全部云厂商。改为 `route.query.type === 'certs' || route.params.resourceType === 'certs'`，旧页面没有 `params.resourceType`，行为不变。PR 也漏了这处。

**文件** `src/views/resource/resource-manage/hooks/use-filter.ts`、`src/views/resource/resource-manage/hooks/use-filter-host.ts`

两个 hook 末尾监听 `props.filter`，变化时置空 `searchValue` 以重算内部 `filter`（真正发请求的那份）；判断写的是旧路径 `/^\/resource\/resource/`，新地址 `/resource/manage/*` 下不触发，导致「全部账号 → 具体账号」、选云厂商、回到全部账号都不重新请求（账号之间切换能请求，只因外层原地改了同一条 `account_id` 规则对象）。照 PR 改为 `/\/resource\/(resource|manage)/`（用户 2026-10-10 实测发现）。过渡修补：筛选写入地址单切到 `useFilterFromRoute`（监听 `route.query` 原地重算 rules）后，这两个 hook 随之删除。

## 9. 站内旧链接改写（F-007.3，本单执行）

新路由本单即注册，链接本单改。按**路径**跳转的旧链接虽有转址兜底，也一并改成新 name；按 `MENU_RESOURCE_RESOURCE_MANAGEMENT` name 跳转的必须改（按 name 进旧记录会渲染旧列表页、不触发空路径子路由的转址）：

| 位置 | 现状 | 改为 |
|------|------|------|
| `views/service/service-apply/clb/index.tsx` `goBack` | `{ name: MENU_RESOURCE_RESOURCE_MANAGEMENT, query: { type: 'clb' } }`（丢账号，现网同） | `{ name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'clb' }, query: { accountId } }`（**必改**） |
| `views/service/service-apply/subnet/index.tsx` `handleCancel` 兜底 | `path: '/resource/resource', query: { type: 'subnet' }` | 新 name + params |
| `children/manage/security-manage.vue` 切 scene | 资源下 `router.replace({ path: '/resource/resource', query: { type: 'security', ... } })` | 资源下只改 query、留在当前路由 |
| `children/components/host/host-security.vue`（2 处 `routerAction.open`） | `path: '/resource/resource', query: { type, accountId }` | 新 name + params（与 ⑥ 改详情时一并处理亦可） |
| `service-apply/components/common/{vpc,subnet,security-group}-selector.tsx`、`cvm/children/VpcPreviewDialog` | `window.open('/#/resource/resource?type=...')` | `routerAction.open` 新 name |
| 旧资源接入页内跳旧列表（`resource-manage.vue` 顶部「资源」页签 `RESOURCE_TABS`、旧 `useResourceAccount.tsx` 的路径白名单） | `/resource/resource` 路径 | 不改：经空路径子路由转址到新列表；随旧页在 ⑦ 删除 |
| 旧详情 `resourceDetail` 路由的 `menu.relative`（面包屑返回） | `MENU_RESOURCE_RESOURCE_MANAGEMENT` | `MENU_RESOURCE_RESOURCE_LIST`，`main-content.vue` 的 `legacyBackConfig` 改传 `{ params: { resourceType: route.params.type } }`（§12） |
| `components/layout/header.vue` 切一级菜单清账号选择 | `nav.id !== MENU_RESOURCE` | `![MENU_RESOURCE, MENU_RESOURCE_ACCESS].includes(nav.id)`（过渡，随筛选写入地址单删除） |

改完 `rg "MENU_RESOURCE_RESOURCE_MANAGEMENT" src` 只剩 `menu-symbol.ts` 与旧 `resource.ts` 记录本身；`rg "/resource/resource" src` 只剩旧资源接入页内部与 `resource.ts`。

## 10. 资源运营翻转与注册

**文件** `src/constants/menu-symbol.ts`、`src/views/index.ts`、`src/router/index.ts`、`src/common/menu-service.ts`、`src/components/layout/header.vue`

- `views/index.ts`：`resourceViews = injectStatusComponents([...resourceManageRoutes, ...resource, ...resourceInside, ...billRoutes])`（`resource` / `resourceInside` 为旧资源接入记录，§12；`billRoutes` 见 §11），删除 `resourceAccessViews` / `billViews`。
- `router/index.ts`：
  - `{ name: MENU_RESOURCE, path: '/resource', redirect: { name: MENU_RESOURCE_MANAGE }, children: resourceViews }` 替换 ① 的「资源接入」过渡父路由与 `{ name: MENU_BILL, path: '/bill', … }`。
  - 顶栏「资源接入」不注册路由（用户 2026-10-10 决定，取代原 ⑦ 的 `/resource-access` 专用转址记录）：菜单项声明 `route: getMenuRoute(resourceViews, MENU_RESOURCE_MANAGE)`，`header.vue` 对声明了 `route` 的一级菜单（跳转入口）链到 `route.name`、`hasRoute` 也按它判断，并给该 `RouterLink` 传空的 `active-class` / `exact-active-class`，使其在资源纳管页不与资源运营同时高亮。「资源接入下线」单删菜单项即可。
  - `...resourceManageLegacy` 与 `{ path: '/bill', redirect: '/bill/account-manage' }`（保持现网裸 `/bill` 行为，§11）放在 catch-all 404 之前。
- `menu-service.ts` 一级菜单（顺序 = 顶栏顺序，与现网一致）：业务资源、服务请求、`{ id: MENU_RESOURCE_ACCESS, i18n: '资源接入', route: 资源纳管 }`（只跳转）、资源选型、`{ id: MENU_RESOURCE, i18n: '资源运营', menu: [...] }`（不设 visibility，始终显示；原资源运营的 `ENABLE_ACCOUNT_BILL` 开关下放到账单三项）。删除一期 `MENU_BILL` 一级项与 `billMenus`；去掉 ① 为资源接入写的 `MENU_RESOURCE` 无侧栏特例（`layout.sideMenu: false`，以 ① 实际位置为准）。
- 资源运营侧栏（最终菜单树，取原 ⑦ 方案）：除资源纳管外均为无云账号树的独立页面，按 PR 不用原资源接入页面临时顶替（用户 2026-10-10 决定）；各项由交付其独立页面的单加入：

  | 顺序 | 菜单 | id | 加入单 | group | 显隐 |
  |---|---|---|---|---|---|
  | 1 | 资源纳管 | `MENU_RESOURCE_MANAGE` | **本单** | 云资源（groupIcon `bkhcm-icon-vpc`） | 始终 |
  | 2 | 账号管理 | `MENU_RESOURCE_ACCOUNT_MANAGEMENT` | ⑤ | 云资源 | ⑤ 定（`hideMenu`） |
  | 3 | 回收管理 | `MENU_RESOURCE_RECYCLEBIN` | ④ | 云资源 | ④ 定（预鉴权） |
  | 4 | 一级账号 | `MENU_BILL_ROOT_ACCOUNT` | **本单** | 云账号管理（groupIcon `bkhcm-icon-account-manage`） | `ENABLE_ACCOUNT_BILL === 'true'` + ② `hideMenu` |
  | 5 | 二级账号 | `MENU_BILL_MAIN_ACCOUNT` | **本单** | 云账号管理 | 开关 |
  | 6 | 账单管理 | `MENU_BILL_MANAGE` | **本单** | 云账号管理 | 开关 + ② `hideMenu` |
  | 7 | 操作记录 | `MENU_RESOURCE_OPERATION_LOG` | ④ | 无（icon `bkhcm-icon-operation-record`） | ④ 定（预鉴权） |

  本单期间原资源接入的账号列表 / 页签、回收站、操作记录、资源详情只能经旧地址打开（侧栏无入口，仍为原云账号树布局，见 §12），随 ④⑤⑥ 撤掉；首批同一 PR 合入，这一中间状态不单独发布。`route` 一律 `getMenuRoute(resourceViews, id)`。分组图标取 PR 的 `groupIcon`；① 的 `menu-item.vue` 按菜单项 `icon` 渲染，各项沿用一期 / 业务侧图标（资源纳管 `bkhcm-icon-vpc`、账号管理与账单三项 `bkhcm-icon-account-manage` / `bkhcm-icon-bill-manage`、回收管理 `bkhcm-icon-recyclebin`），与侧栏其它菜单对齐。
- `app.vue` 进入 `MENU_RESOURCE` 置业务 ID 为 0 的逻辑不改，只去掉过渡 TODO：翻转后它覆盖整个资源运营（含账单，账单页不读 `accountStore.bizs`）。
- `header.vue`：跳转与高亮不改（`{ name: menu.id }` + `.router-link-active`）；切一级菜单清账号选择的条件见 §9。

## 11. 账单挂到资源运营下（地址不变）

**文件** `src/views/bill/route-config.ts`

- 各记录 `path` 改为绝对路径（`/bill/root-account`、`/bill/root-account/create`、`/bill/main-account`、`/bill/main-account/create`、`/bill/bill-manage`；账单管理的子路由仍相对），旧地址 `account-manage`、`account-manage/first-account`、`account-manage/second-account` 同样改绝对路径（`beforeEnter` 分流不变）。作为 `/resource` 根的子路由：地址不变，`matched[0]` 变为 `MENU_RESOURCE`，顶栏高亮资源运营、侧栏为资源运营菜单树。
- 账单管理 `title` 由「云账单管理」改为「账单管理」（与菜单一致，原 ⑦ 方案）。组件、name、`activeKey`、`menu.relative`、`layout.breadcrumb`、嵌套结构不变。
- 裸 `/bill` 保持现网（转 `/bill/account-manage` 分流）；迁到 `/resource/bill/*`、`/bill/*` 旧地址与裸 `/bill` 改 404 留给 ⑦。`useWhereAmI` 的 `/bill/` 判定不受影响。

## 12. 旧资源接入页面的过渡调整

**文件** `src/router/module/resource.ts`、`src/components/layout/main-content.vue`、`src/style/app-shell.scss`

④⑤⑥ 交付前仍注册的旧页面（只承接旧地址与站内旧链接，侧栏无入口）：账号列表 / 账号页签（⑤）、回收站 / 操作记录（④）、旧资源详情（⑥，资源纳管列表点资源仍进它）、`resourceInside`（以 ⑤ 为准）。

- `resource.ts`：
  - 删除 5 条 `/resource/service-apply/*`（与新申请页同名，§7 转址）。
  - 旧 `/resource/resource` 记录加子路由 `{ path: '', redirect: toResourceList }`：精确访问转到新列表；`record` / `recycle` / `account*` 子页面不受影响。
  - 账号列表去掉 `alias: ''`：裸 `#/resource` 改由资源运营根 redirect 到资源纳管（用户 2026-10-09 提出的预期）。
  - 外层 `resource-entry.vue` 记录（`path: '/resource'`）加 `redirect: { name: MENU_RESOURCE_MANAGE }`（它与根同路径、vue-router 优先匹配子记录，不加则裸 `/resource` 渲染空白外壳）与 `meta: { legacyResourceAccess: true }`。
  - `resourceDetail` 的 `menu.relative` 改为 `MENU_RESOURCE_RESOURCE_LIST`（§9）。
- `main-content.vue`：`isLegacyResourceAccess` 改为 `!!route.meta.legacyResourceAccess`（原按 `matched[0] === MENU_RESOURCE`，翻转后会套到新模块和账单上）；`legacyBreadcrumbRoutes` 去掉 5 个申请页 name（已是新模块路由，由入口右栏渲染面包屑）；`legacyBackConfig` 改传 `params.resourceType`（旧详情类型映射为页签：`eips → ip`、`route → routing`、`gcp` / `template → security`）并带 `query.accountId`：新列表的账号来自地址，不带会丢失账号选择。
- `app-shell.scss` 旧云账号树容器：原 `position: fixed; left: 0; width: 15%` 贴视口左边，翻转后资源运营有侧栏会被盖住；改为两栏 flex（树 `width: 240px; flex-shrink: 0`、内容区 `flex: 1; min-width: 0`），去掉 `margin-left: 15%` / `width: 85%`。只影响上列旧页面，⑦ 随过渡容器删除。

## 13. 迁移提示条

**文件** `src/views/resource-manage/children/migration-notice.vue`（新）、`src/views/resource-manage/constants.ts`（新）、`src/views/resource-manage/resource-entry.vue`

照原 ⑦ 方案移入本单：`bk-alert` closable，文案「资源接入已迁移至 资源运营 → 云资源 → 资源纳管，原菜单将于 {日期} 下线」；日期常量 `RESOURCE_ACCESS_OFFLINE_DATE` 放 `constants.ts`；关闭状态用 `localStorageActions` 写入 key `hcm-resource-access-migration-notice-closed`；挂在 `resource-entry.vue` 右侧内容区顶部，列表、新建、详情都显示，直到关闭。**下线日期与文案待定（资源运营重组需求 Q-001 / Q-002），未填不得发版。**

## 从 PR #1897 取 / 不取

| PR 文件 / 改动 | 取舍 |
|---|---|
| `views/resource-manage/route-config.ts` | 取结构与 Symbol name；不取 `auth.view`；详情子路由由 ⑥ 追加；补 `isFilterAccount` |
| `resource-entry.vue` | 取两栏骨架、样式与「非列表渲染面包屑」；补卸载清 store |
| `resource-list.vue` | 取页签 / 表单 / 侧滑 / 新建跳转；不取按地址读账号、云厂商、分配状态；补 HEAD 筛选、账号列表预取、子组件 props、非法类型回退 |
| `account/vendor-group/*`（PR 副本） | 本单过渡：以 HEAD 版为底复制 `index.vue`，取 PR「详情页切账号回列表」思路；`.plugin.ts` 与 `vendor-property.ts` 按 PR 移入新目录。`index.vue` 的 PR 版由「筛选条件写入地址」单替换 |
| `account/account-summary.vue` | 取 UI；数据源改 store；补云厂商名、`showAccountStatus` 自动打开 |
| `account/account-status/*` | 取；改按需渲染、打开时取详情、去前缀、加载失败提示 |
| `account/createAccount/**`（PR 副本） | 取 PR 位置，目录按职责改名为 `account-import`（kebab-case），内容用 HEAD 最新版整体移动（§2b） |
| `hooks/use-filter-from-route.ts`、`hooks/search-properties.ts`、`docs/filter-refactor-plan.md`、`*-manage.vue` / `use-filter*.ts` / `use-query-list-common.ts` 的改动、`2e535a0cd1` | 不取（筛选写入地址单） |
| `25696ea07b` 中 `service-apply/**` 去 `DetailHeader`、`sticky-bottom-container` | 已由 ① 修订（2026-10-09）取，本单不再处理 |
| `5c513eb559` 详情页布局、`MENU_RESOURCE_DETAIL` | 不属本单（⑥） |
| `60dac3a85a` 中 `account-manage/**`、`usersList` 删除 | 不属本单（⑤）；本单不提供用户列表替代 |
| `components/layout/breadcrumb.vue`、`main-content.vue`、收藏 | 不取（① / 用户决定）；不改 ① 的组件 |

## 对照 PR 的旧文件与过渡逻辑退场

PR #1897 对旧文件的处理是加 `@deprecated` 保留；本分支不留废弃副本，**无引用即删除**，按下表分批（用户 2026-10-10 确认：PR 为目标状态，本单不做的给出归属批次）：

| PR 中的状态 | 本分支现状 | 退场批次 |
|---|---|---|
| `views/resource-manage/account/createAccount/`（移入新目录） | 旧位置 | **本单**（§2b 整体移动并改名为 `account-import`） |
| `views/resource-manage/account/vendor-group/{vendor.plugin,vendor-property}.ts`（移入新目录） | 旧位置 | **本单**（§2 移动，旧云账号树改引新位置） |
| 新云账号树按地址驱动，不用 `useResourceAccountStore`、无离开清空 | 本单过渡用 store 版副本 + 入口卸载 `clear()`；① 顶栏切换时 `clear()` | 筛选条件写入地址单：替换为 PR 版云账号树，消费方改读地址（含 PR 漏改的 `HostOperations.tsx` / `template-dialog` / `use-selection.ts`），删除两处 `clear()` 与新模块对 store 的写入 |
| `hooks/use-filter.ts`、`use-filter-host.ts` `@deprecated`，`*-manage.vue` 改用 `useFilterFromRoute` | 在用 | 筛选条件写入地址单（切换后无引用即删） |
| `resource-manage.vue`、旧云账号树 `account/vendor-group/`、`account/accountList/**`（含 `useResourceAccount.tsx`）`@deprecated` | 旧资源接入路由在用 | ⑦（旧路由与 `main-content.vue` 过渡容器撤掉后无引用即删） |
| `accountInfo/index.tsx` 迁 `views/account-manage/` | 旧账号页在用 | ⑤ 新建、⑦ 删旧（见 ⑤ coding.md） |
| `views/resource/recyclebin-manager/` `@deprecated` | 旧回收站在用 | ④ 新建、⑦ 删旧 |
| `views/resource/route-config.ts`、`router/module/*` `@deprecated` | 在用 | 资源接入相关 3 个随 ⑦；其余见拆单总览「旧路由文件退场计划」 |
| `*-manage.vue`、详情页留在 `views/resource/resource-manage/children/` | 同 | 不迁（与 PR 一致） |

## HEAD 新增能力必须保留

本单复用以下 HEAD 文件且不改其内容，能力自动保留（`git log --oneline 8fea9924df..HEAD -- <path>` 已核对）：

| 路径 | HEAD 独有提交 | 保留方式 |
|------|--------------|----------|
| `views/resource/resource-manage/children/manage/` | `9c6ac9204f` searchselect async / filterRules、`4348ae1301` 主机批量分配、`0a7a62faff` bkui 升级 | 列表直接 import，不取 PR 改动 |
| `views/resource/resource-manage/children/dialog/` | `c5b61eff78` 安全组克隆选目标地域、`c1938521b9` select filterable；批量分配弹窗 `batch-distribution/` | 经 `*-manage.vue` 间接复用 |
| `utils/search.ts` | `9c6ac9204f` | 不动 |
| `views/resource/resource-manage/account/createAccount/`（本单移为 `views/resource-manage/account/account-import/`） | `fc622471df` 人员选择器、`db6f42fc86` 腾讯云禁用提示 | `git mv` 保留内容，单一来源 |
| `views/service/service-apply/` | `55961e6bc1` CLB 名称支持点号、`c1938521b9` | 新建页路由指向原组件 |
| `views/load-balancer/entry-rsc.vue` | 基线已有，PR 同样复用 | 原样 |

`resource.md` 记录的「账号同步状态轮询闭包修正」与 `load_balancer_exclusive_cluster` 译名**不在 HEAD**（提交 `8121345b7a` 只在别的分支，HEAD 也没有独占集群）；本单侧滑的 `resource-status.tsx` 本身就是修正后的写法，译名 map 无需补。

## 风险与验证

| # | 风险 | 怎么验 |
|---|------|--------|
| 1 | 旧筛选机制搬进新列表后行为偏差（核心） | 逐项对照现网：全部账号 / 选云厂商 / 选账号后请求 `filter.rules`（`vendor`、`account_id`）；分配状态三档；搜索栏云账号下拉有选项；`?cloud_id=` 进入时搜索框回填；切页签后筛选与分配状态保持 |
| 2 | `security-manage` 切 scene 仍 replace 到旧路径 | 本单改为只改 query（§9）；切「安全组 / 模板 / GCP 防火墙」地址留在 `/resource/manage/security`，`scene`、`accountId` 保留，无转址 |
| 3 | 新建页挂进三栏后的排版 | ① 修订已去掉固定定位的 `DetailHeader` 与按「树占 15%」写死的偏移，改为 `sticky-bottom-container`；本单只核对 5 个新建页在侧栏展开 / 折叠、1366px 下面包屑在右栏顶部、底部按钮吸底不错位 |
| 4 | 云厂商依赖账号详情异步返回，刷新 `/resource/manage/clb?accountId=<非腾讯云>` 先显示 CLB 再回退主机 | 确认回退为 replace、无报错；可接受（现网同样异步） |
| 5 | 侧滑轮询未停 / 打到旧账号 | AC-007：A 打开→关闭→切 B→打开，`sync_details/<id>` 全为 B；关闭后无新请求 |
| 6 | 翻转后路由匹配（同路径父子记录、空路径子路由、绝对路径子路由） | 浏览器逐条：裸 `#/resource`、`#/resource/` → 资源纳管主机；点顶栏「资源接入」→ 资源纳管，顶栏只高亮资源运营；`#/resource/resource?type=clb&accountId=x` → `/resource/manage/clb?accountId=x`；`#/resource/resource/record`、`/recycle`、`#/resource/account`、`#/resource/recyclebin` 仍为旧页面；5 个旧申请地址转到新申请页且 query 保留；`#/bill`、`#/bill/account-manage*` 分流不变；未知地址 404 |
| 6b | 资源运营侧栏与高亮 | 侧栏只有资源纳管与云账号管理三项；账单三项在 `ENABLE_ACCOUNT_BILL` 关 / 无权限时隐藏；进账单页侧栏高亮对应项、顶栏高亮资源运营；经旧地址打开的原资源接入页面顶栏高亮资源运营，旧云账号树在侧栏右侧两栏展示、不被盖住 |
| 7 | 1366px 下工具栏被截（侧栏 240 + 树 240 后内容约 880px） | 11 个页签逐一看搜索框可压缩、按钮不被吃掉（`resource.md` 的 `justify-content-end` 左溢问题） |
| 8 | `.plugin.ts` 可能被其它构建变体按路径替换 | 按 PR 移到新目录（PR 同样移动）；若存在按旧路径覆盖 `vendor.plugin.ts` 的构建变体，需同步其路径——提交前 `rg "vendor-group/vendor.plugin"` 全仓（含构建配置）确认无旧路径引用 |
| 9 | 证书页账号下拉 | 证书页签点「上传证书」，只列腾讯云；旧页面同样只列腾讯云 |
| 10 | 性能 AC-P01 | 摘要条不新增请求；首屏请求数与现网对比 |
| 11 | Symbol name 进 `HistoryStorage`（`JSON.stringify` 会丢 Symbol） | ① 已移植（2026-10-10 核实）；冒烟：主机列表→购买→取消 / 面包屑返回，回到原页签且账号不变（AC-009） |
| 12 | 自动打开侧滑 | AC-008b：打开 `/resource/manage/host?accountId=<id>&showAccountStatus=1`，侧滑打开、账号为该 id、地址里参数已去掉；刷新 / 后退不再弹；无 `accountId` 时不弹且参数被去掉 |
| 13 | 入口右栏面包屑 | 冒烟：新建页面包屑在右栏顶部、不横跨树，整页顶部无第二条，返回回到列表原页签；列表页无面包屑；⑥ 合入后详情页同样在右栏，详情的操作按钮出现在面包屑右侧 |

## 待确认

- 迁移提示条的下线日期与文案（§13，资源运营重组需求 Q-001 / Q-002）：可先开发，发版前必须填。

站内链接改写已移入本单（§9）；旧详情页过渡形态随 ⑥ 消除。
