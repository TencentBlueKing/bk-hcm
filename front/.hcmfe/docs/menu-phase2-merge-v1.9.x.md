# 菜单整合二期（首批 + 收尾单）：合入 v1.9.x 的冲突处置指引

分支 `feat-menu-phase2-cloud-resource`，首批 8 个提交（① `943fd921e`、①修订 `8591ae2d1`、② `0fd5d853a`、③ `2f023c259`、④ `fd57dd924`、⑤ `3c1a4cb06`、⑥ `b083e2111`、⑦ `767bdd242`）+ 收尾单提交。分叉点 `01bca385c`（一期 #2085）。

2026-10-12 核对：`origin/v1.9.x` 分叉后新增 4 个提交，均只改后端，前端暂无冲突。合并前按下文重新核对。

## 1. 合并方式

- 建议在本分支上 `git merge origin/v1.9.x` 一次性解决冲突（不要 rebase：被删除 / 迁移的文件会在多个提交里反复冲突）。
- 合并前先跑 §2 的命令，按 §3 逐文件决定处置，再处理 §4 的「不冲突但会失效」项，最后跑 §5 验证。

## 2. 合并前：找出主干动过的「本分支已删除 / 迁移」文件

```bash
MB=01bca385c
git fetch origin v1.9.x
# 主干在分叉后改过的前端文件
git diff --name-only $MB origin/v1.9.x -- front/src > /tmp/trunk.txt
# 本分支删除或迁移的前端文件（旧路径）
git diff --name-status -M50% $MB HEAD -- front/src | awk '$1=="D"{print $2} $1~/^R/{print $2}' > /tmp/gone.txt
# 交集 = 必须人工搬运的主干改动
comm -12 <(sort /tmp/trunk.txt) <(sort /tmp/gone.txt)
# 本分支大改过、主干也改过的文件（文本冲突高发）
git diff --name-only $MB HEAD -- front/src | sort | comm -12 - <(sort /tmp/trunk.txt)
```

对交集里的每个文件，用 `git log -p $MB..origin/v1.9.x -- <旧路径>` 看主干改了什么，再按 §3 的「新位置」把逻辑搬过去。

## 3. 旧文件 → 新位置与处置

处置规则：
- **R（git 识别为重命名）**：git 通常会把主干改动跟到新路径；冲突时把主干的 hunk 应用到新路径。
- **D·替换**：旧文件删了，功能在新文件里重写。主干对旧文件的改动 git 会报 modify/delete 冲突，**不能直接保留旧文件**，要把改动意图搬到新位置后删除旧文件。
- **D·移除**：功能已不存在（PR #1897 同样去掉）。主干改动若是新增能力，需判断是否在新结构里补；否则直接删除。

### ① 页面布局与菜单服务化

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `router/header-config.ts` | D·替换 | `common/menu-service.ts`（顶栏一级菜单、各侧栏菜单树，含 `ENABLE_ACCOUNT_BILL` 等开关） |
| `views/home/index.tsx`、`index.scss` | D·替换 | `app.vue`（组装 `bk-navigation`）+ `components/layout/{header,menu,main-content,breadcrumb}.vue` + `style/app-shell.scss` |
| `views/home/breadcrumb.tsx`、`.scss` | D·替换 | `components/layout/breadcrumb.vue` + `hooks/use-breadcrumb`（动态标题 `setTitle`，右侧按钮 `Teleport` 到 `#breadcrumbExtra`） |
| `views/home/hooks/to-legacy-menus.ts`、`useChangeHeaderTab.ts` | D·替换 | `common/menu-service.ts` 的 `getMenuRoute` / 按 `route.matched[0].name` 切顶栏 |
| `views/home/business-selector/index.tsx` | R（① 移入）→ D·替换（收尾单） | `components/business-selector/global.vue` + `hooks/use-business-favorite.ts`，切换逻辑在 `components/layout/menu.vue` |
| `views/home/release-note/index.vue` | R | `components/release-note/index.vue` |
| `views/home/hooks/useCalcTopWithNotice.ts` | R | `hooks/use-calc-top-with-notice.ts` |

主干若在 `views/home/index.tsx` 里加了菜单项、顶栏入口、全局副作用（如新通知、新弹窗），对应搬到 `menu-service.ts` / `app.vue`；新路由的菜单可见性在路由 `meta`（`notMenu`、`activeKey`、`menu.relative`）与 `menu-service.ts` 声明。

### ② 视图权限体系

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `components/global-permission-dialog/index.tsx`、`store/useGlobalPermissionDialog.ts` | D·替换 | 接口 403 弹窗在 `http/index.ts`；按钮级用 `hcm-auth`（`components/auth/auth.vue`） |
| `store/usePagePermissionStore.ts`、`views/resource/NoPermission.tsx` | D·替换 | 页面查看权限：路由 `meta.auth.view` + `constants/view-auth.ts`（侧栏显隐 `hideMenu`）+ `store/auth.ts` 预鉴权 + `views/status/permission.vue`（框架内无权限页）；业务访问权限在 `router/business-interceptor.ts` |

主干若新增页面并在旧 `pageAuthData` / `usePagePermissionStore` 里加查看权限，改为在新路由上声明 `meta.auth.view`（需要侧栏隐藏时同时加 `view-auth.ts`）。

### ③ 资源纳管主体

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `views/resource/resource-manage/account/createAccount/**` | R（目录改 kebab-case） | `views/resource-manage/account/account-import/**` |
| `views/resource/resource-manage/account/vendor-group/{vendor-property.ts,vendor.plugin.ts}` | R | `views/resource-manage/account/vendor-group/` |

### ⑤ 账号管理独立菜单

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `router/module/resource-inside.ts`、`service-inside.ts` | D·替换 | `views/account-manage/route-config.ts`（`accountManageRsc` + 旧地址 `accountManageLegacy`） |
| `views/resource/accountmanage/index.vue` | R | `views/account-manage/list.vue` |
| `views/resource/accountmanage/account-detail.{tsx,scss}` | R | `views/account-manage/details/basic.{tsx,scss}` |
| `views/resource/resource-manage/accountInfo/index.scss` | R | `views/account-manage/details/index.scss` |
| `views/resource/resource-manage/accountInfo/index.tsx` | D·替换 | `views/account-manage/details/index.tsx`（只有「基本信息」页签） |
| `…/accountInfo/component/resourceStatus/**` | D·替换 | `views/resource-manage/account/account-status/`（资源纳管的账号状态侧滑，③ 新建） |
| `…/accountInfo/component/usersList/**` | D·移除 | 账号详情不再有「用户列表」页签（PR 同） |
| `views/resource/accountmanage/account-add.{tsx,scss}` | D·移除 | 账号登记复用 `views/service/service-apply/account-add`（`/resource/account/check-in`） |
| `views/resource/resource-manage/common/header/detail-header.{tsx,scss}` | D·替换 | 全局面包屑：`useBreadcrumb().setTitle()` + `Teleport defer to="#breadcrumbExtra"` |

### ⑦ 资源运营一级视图重组

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `router/module/resource.ts` | D·替换 | `views/resource-manage/route-config.ts`（资源纳管、资源详情、旧地址 `resourceManageLegacy` / `resourceDetailLegacy`）；其余资源运营页面在各自 `views/<模块>/route-config.ts`，由 `views/index.ts` 的 `resourceViews` 汇总 |
| `views/resource/resource-manage/resource-manage.vue` | D·替换 | `views/resource-manage/resource-list.vue`（资源类型在 `params.resourceType`） |
| `views/resource/resource-manage/resource-entry.vue` | D·替换 | `views/resource-manage/resource-entry.vue`（左云账号树 + 右栏面包屑） |
| `views/resource/resource-manage/account/vendor-group/index.vue` | D·替换（③ 复制、⑦ 删旧） | `views/resource-manage/account/vendor-group/index.vue` + `use-resource-account.ts` |
| `views/resource/resource-manage/account/accountList/**` | D·替换 | 同上；`useResourceAccount` → `views/resource-manage/account/use-resource-account.ts` |

### 收尾单：业务地址、业务路由、菜单收藏与折叠

| 旧路径 | 类型 | 新位置 |
|---|---|---|
| `router/module/business.ts` | D·替换 | 各模块 `views/{host,drive,image,vpc,subnet,eip,network-interface,route-table,security-group,cert}/route-config.ts`（列表、详情、申请页），负载均衡 / 单据 / 操作记录 / 任务 / 云账号管理 / 回收站在各自原有 `route-config.ts`；`views/index.ts` 的 `businessViews` 拼接。根路由 `/business/:bizId(\d+)?`，子路由相对路径 |
| `router/module/common.ts`、`views/error-pages/403.{tsx,scss}` | D·移除 | 旧 `/403/:id` 不再注册，落 404（PR 同）；页面查看权限见 ② |
| `views/resource/recyclebin-manager/recyclebin-manager.vue` | D·替换 | `views/recyclebin/index.vue`（业务 / 资源共用，按 `whereAmI` 区分），路由 `recyclebinBiz` / `recyclebinRsc` |
| `components/layout/business-selector.tsx`、`hooks/useFavorite.ts` | D·替换 | `components/business-selector/global.vue`、`hooks/use-business-favorite.ts` |
| `hooks/useVerify.ts`、`components/permission-dialog/{index.tsx,index.scss}`、`assets/image/403.png` | D·替换 | 按钮级鉴权统一 `hcm-auth`（`components/auth/auth.vue`），申请弹窗由 `app.vue` 全局提供（PR 同） |
| `store/common.ts` 的 `pageAuthData`、`authVerifyData`、`authVerify` 等 | D·移除 | 只剩 `isNoticeAlert`；`main.ts` 不再在挂载前预取旧按钮权限（PR 同） |
| `views/service/service-apply/components/{basic-resource,senior-resource,resource-card}/` | D·移除 | 全仓无引用的旧服务申请卡片（PR 同） |

主干若在 `business.ts` 加了业务页面：在对应模块 `route-config.ts` 加相对路径记录（`name` 用菜单符号，`meta.owner = MENU_BUSINESS`），并入 `businessViews`，菜单项加在 `menu-service.ts`。主干若改了 `recyclebin-manager.vue`：业务侧行为搬到 `views/recyclebin/index.vue` 的 `isBusiness` 分支。主干若改了旧选择器：搬到 `global.vue`（展示 / 收藏）或 `menu.vue` 的 `handleChangeBusiness`（切换）。主干若新增 `useVerify()` / `authVerifyData.permissionAction.xxx` / `<PermissionDialog>` 用法：在 `pageAuthData` 里查到 `xxx` 对应的 `type` + `action`，在 `constants/auth-symbols.ts` / `common/auth-service.ts` 找到或补上同一权限项的符号，改成 `hcm-auth`（业务侧带 `relation: [bizId]`）。

## 4. 不会文本冲突、但合并后会失效的约定变化

主干新代码若按旧约定写，合并能通过、运行时才出错。合并后逐项排查主干新增代码：

| 变化 | 旧写法（需改） | 新写法 |
|---|---|---|
| 路由 name 改为菜单符号 | `'hostBusinessDetail'` 等 `${type}BusinessDetail`、`'resourceDetail'`、`'accountDetail'` | `MENU_BUSINESS_*_DETAILS`、`MENU_RESOURCE_DETAIL`、`MENU_RESOURCE_ACCOUNT_DETAIL`（`constants/menu-symbol.ts`） |
| 详情 ID 进路径 | `query: { id }`、`route.query.id` | `params: { id }`；详情组件读 `useResourceDetailRoute().detailId`（旧 `?id=` 由 `normalizeDetailId` / 旧地址数组转换） |
| 跳详情的云厂商键 | `query: { type: vendor }` | `query: { vendor }`；详情组件读 `route.query.vendor` 或详情返回的 `vendor` |
| 资源类型名 | 列表 / 详情类型 `ip`、`eips`、`route` | `eip`、`routing`（`RESOURCE_TYPES`、资源纳管页签与详情地址）；旧名只做兼容。`eips` 仍是接口路径与列配置 key |
| 地址变化 | `/resource/resource*`、`/resource/detail/:type`、`/resource/account/detail`、`/bill/*` | `/resource/manage/<类型>[/details/<id>]`、`/resource/account[/details/<id>]`、`/resource/recyclebin`、`/resource/record`、`/resource/bill/*`；站内一律按 name 跳 |
| 详情组件按钮鉴权 | `inject('authVerifyData')`、`useVerify()`、`bus.$emit('auth')` | `hcm-auth` + `useResourceDetailAuth`（资源侧按账号、业务侧按业务） |
| 页面标题栏 | `<detail-header>` / 自画标题 | `useBreadcrumb().setTitle` + `#breadcrumbExtra` |
| 页面查看权限 | `store/common.ts` 的 `pageAuthData` 加查看项 | 路由 `meta.auth.view` |
| 按钮级鉴权（非资源页） | `useVerify()`、`authVerifyData.permissionAction.xxx`、`handleAuth('xxx')`、`<PermissionDialog>`、`provide('authAction')` | `hcm-auth` / `HcmAuth`（`sign={{ type: AUTH_* }}`），无权限置灰、点击弹全局申请弹窗；只想读「是否有权限」用 `useAuthStore().verify` 或视图权限缓存 |
| 旧权限弹窗全局注册已去掉 | 模板里直接写 `<permission-dialog>` / `<PermissionDialog>` 而不 import | 编译能过，运行时只报「Failed to resolve component」、弹窗不出现。改 `hcm-auth` |
| `store/common.ts` 字段删除 | `useCommonStore().authVerify / authVerifyData / pageAuthData / list / updatePageAuthData` | 只剩 `isNoticeAlert` / `setIsNoticeAlert`；TS 文件会报错，`.vue` 里 `any` 断言的写法不会报错，按 §4 命令搜 |
| `@/hooks` 导出减少 | `import { useVerify } from '@/hooks'` | 编译报错，按上一行改 `hcm-auth` |
| 菜单 | 改 `header-config.ts` / `views/home` | 改 `common/menu-service.ts` |
| 业务 ID 进路径 | `query: { [GLOBAL_BIZS_KEY]: … }`、`?bizs=`、`route.query.bizs`、`'/business/xxx'` | 跳业务路由一律显式传 `params: { bizId }`（可选路径参数按名字跳转不会被沿用）；读用 `route.params.bizId` / `getBizsId()`；字符串地址 `/#/business/${bizId}/xxx`。漏传或旧 `?bizs=` 由拦截器转址兜底，不会白屏但多一次跳转 |
| 业务路由 name | `'businessRecyclebin'`、`'applyCvm'` / `'applyVPC'` / `'applyDisk'` / `'applySubnet'`、`MENU_BUSINESS_LOAD_BALANCER_{OVERVIEW,APPLY}`、`MENU_BUSINESS_TARGET_GROUP_OVERVIEW`、`MENU_BUSINESS_RECYCLE_BIN_MANAGEMENT` | `MENU_BUSINESS_RECYCLEBIN`、`MENU_BUSINESS_APPLY_{CVM,VPC,DISK,SUBNET,CLB}`、`MENU_BUSINESS_LOAD_BALANCER_{LB,TG}_VIEW` |
| 业务地址 | `/business/ip`、`/business/host/recyclebin/:type`、`/business/load-balancer/device`（绝对路径） | `/business/<业务ID>/eip`、`/business/<业务ID>/recyclebin?type=`；旧地址有转址 |
| 业务页按路径判断 | `route.path.includes('/business/host')`、`route.path === '/business/recyclebin'` | 按 `route.name` 判断（路径里多了业务 ID） |
| 等业务 ID 再渲染 | `v-if="accountStore.bizs"` / `bk-loading` 等选择器补参数 | 拦截器在进入页面前写好 `accountStore.bizs`，不需要等 |

检查命令（2026-10-12 在本分支实测结果见注释）：

```bash
# 旧路由 name：期望 0
rg "BusinessDetail'|'resourceDetail'|'accountDetail'" front/src
# 旧详情鉴权：期望 0（gcp-detail.vue 里是原有注释，已排除）
rg "^\s*[^/\s].*(inject\('authVerifyData'|\\\$emit\('auth')" front/src/views/resource/resource-manage front/src/views/business
# 旧标题栏组件：期望 0（资源选型自有的 scheme-detail/components/detail-header 不相关）
rg "common/header/detail-header" front/src
# 旧地址字面量：期望只剩各 route-config.ts 的旧地址数组
rg "['\`]/(#/)?resource/resource|['\`]/(#/)?resource/detail/|['\`]/(#/)?bill/" front/src | grep -v route-config.ts
# 业务 ID 旧拼法：期望只剩 common/constant.ts（定义）、router/business-interceptor.ts、hooks/useWhereAmI.ts（后两者只用作 localStorage 键或兼容转址）
rg -l "GLOBAL_BIZS_KEY|query\.bizs|bizs=\\\$|\bbizs: (accountStore|getBizsId|currentGlobal|route)" front/src
# 不带业务 ID 的业务地址字面量：期望 0
rg "['\`\"]/(#/)?business/[a-z]" front/src
# 已删除的业务路由 name：期望 0
rg "'businessRecyclebin'|'apply(Cvm|VPC|Disk|Subnet)'|MENU_BUSINESS_(LOAD_BALANCER_(OVERVIEW|APPLY)|TARGET_GROUP_OVERVIEW|RECYCLE_BIN_MANAGEMENT)" front/src
# 旧按钮鉴权：期望只剩 subnet-manage.vue、gcp-detail.vue 里的原有注释
rg "useVerify|authVerifyData|pageAuthData|components/permission-dialog|'authAction'|<permission-dialog|<PermissionDialog|commonStore\.(authVerify|list|updatePageAuthData)" front/src
```

## 5. 合并后验证

1. `npm run build`、`npx tsc --noEmit -p tsconfig.json`（与合并前对比无新增错误）、`bkdevbuddy lint`。
2. `front/.hcmfe/workflow/feat-menu-phase2-resource-ops/coding.md` 的「终扫」表全部满足；§4 的检查命令结果符合预期。
3. 跑 `front/.hcmfe/workflow/feat-menu-phase2-*/test.md` 的 P0（重点：资源纳管、资源详情两侧、账号管理、回收管理 / 操作记录、账单、旧地址转址；业务侧带 `?bizs=` 的旧地址、切换业务、业务回收站、菜单收藏与折叠）。

## 6. 留给后续单据的过渡代码

`accountStore.bizs`：只是地址业务 ID 的镜像（业务拦截器写入、`router/index.ts` 的 `afterEach` 在资源运营 / 资源选型 / 无归属页面清零），合并时保持现状，由「去掉 accountStore.bizs」单据逐处替换后删除。主干新代码不要再读它，改用 `route.params.bizId` / `getBizsId()`。

## 7. 合入 bcc/v1.9.x：本文清单不完整，必须在 bcc 分支上重新扫一遍

§3、§4 只覆盖 GitHub 仓库（`origin/v1.9.x`）里的文件和写法。`bcc/v1.9.x`（bcc 侧仓库，非 GitHub 的 `origin/v1.9.x`）有 GitHub 没有的模块和页面，也可能自己加了旧写法，合进去时**一定还有本文没列出来的适配点**，不能只按本文的清单改完就算完成。

1. **以 `bcc/v1.9.x` 为主干重跑 §2**：把命令里的 `origin/v1.9.x` 换成 `bcc/v1.9.x`。「主干改过 ∩ 本分支已删除 / 迁移」的交集会比 GitHub 多，逐个按 §3 搬。
2. **在合并后的 bcc 分支上重跑 §4 全部检查命令**，期望值以本文注释为准；多出来的命中全部是 bcc 独有的旧写法，按 §4 表格的「新写法」逐处改。重点关注：
   - 旧鉴权：`useVerify`、`pageAuthData`、`authVerifyData.permissionAction`、`<permission-dialog>`（全局注册已去掉，运行时才会发现）、`provide/inject('authAction')`、`bus.$emit('auth')`
   - 旧路由 name、旧地址字面量、`?bizs=` / `GLOBAL_BIZS_KEY`、`'/business/xxx'` 绝对路径、按 `route.path` 判断业务页
   - 往 `router/module/business.ts`、`router/module/common.ts`、`header-config.ts` 加的菜单 / 路由（这些文件已删，要搬到各模块 `route-config.ts` 和 `menu-service.ts`）
   - 页面查看权限往 `pageAuthData` 加的项（改 `meta.auth.view`，需要隐藏菜单的加 `viewAuthConfig`）
3. **bcc 独有的权限点**：`pageAuthData` 里如有 GitHub 没有的 `type` + `action`，先在 `constants/auth-symbols.ts` / `common/auth-service.ts` 补符号和定义，再改 `hcm-auth`；业务侧带 `relation: [bizId]`。改完逐个核对权限项（IAM action、是否带实例）与旧写法一致。
4. **bcc 独有的菜单 / 一级视图**：按 rule `fe-menu-route-architecture` 归到业务 / 工作台 / 资源运营视图，`meta.owner`、菜单分组、面包屑、`activeKey` 都要配。
5. **验证**：除 §5 外，bcc 独有页面各走一遍「有权限 / 无权限 / 旧书签地址」三种情况，并看启动阶段 Network 只有视图权限那一次 `/auth/verify`。
6. 合入 `bcc/v1.9.x` 时发现的新适配点，补回本文 §3 / §4（标「bcc」），方便下次合并。

另有一批过渡代码留给「资源列表筛选条件写入地址」（二期合入 v1.9.x 后从主干另开分支做），合并时同样保持现状：资源纳管的过渡版云账号树（`views/resource-manage/account/vendor-group/index.vue` + `use-resource-account.ts`，读写全局账号缓存 `useResourceAccountStore`）、`resource-entry.vue` 卸载时清空账号缓存、旧筛选 hook `use-filter.ts` / `use-filter-host.ts`。
