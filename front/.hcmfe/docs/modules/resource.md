# 资源运营一级视图（容器·待拆解）

> status: stub · kind: module
> globs: `src/views/resource/**`, `src/views/resource-manage/**`, `src/views/recyclebin/**`, `src/views/account-manage/**`

一级视图 /resource/ 容器，聚合账号/回收站/资源纳管，待拆解。

## 职责

一级视图 `/resource/` 容器（资源运营，路由 name `MENU_RESOURCE`，默认进资源纳管）。资源 tab「负载均衡」挂 `src/views/load-balancer/entry-rsc.vue`（双入口壳）；CLB 列表在 `load-balancer-manage.vue`，独占集群列表在负载均衡模块 `exclusive-cluster/`。

资源分配弹窗 `src/views/resource/resource-manage/children/dialog/batch-distribution/`：工具栏「批量分配」读表格勾选；`open(rows)` 用入参行打开同一弹窗（不改勾选）。一条时标题为「{资源名}分配」。独占集群类型 `exclusive_clusters`，body key `exclusive_cluster_ids`。

## 资源纳管（`views/resource-manage/`）

- 路由 `route-config.ts`：`manage`（`MENU_RESOURCE_MANAGE`，入口 `resource-entry.vue`）下挂列表 `:resourceType`（`MENU_RESOURCE_RESOURCE_LIST`，资源类型 = 页签 key，在 params 里）与 5 个新建页 `service-apply/{cvm,vpc,disk,subnet,clb}`。旧地址：`toResourceList`（`/resource/resource?type=` → 列表，挂在原资源接入记录的空路径子路由上）、`resourceManageLegacy`（旧新建页地址）。
- 入口两栏：左侧云账号树（240px），右侧迁移提示条 + 非列表页的面包屑 + 内容区；内容区无内边距，列表 / 详情 / 新建页各自留白。入口卸载时清空 `useResourceAccountStore`。
- 列表 `resource-list.vue`：筛选沿用原资源接入机制——账号、云厂商由云账号树写入 `useResourceAccountStore`，列表把它们拼进 `filter` 传给共用的 `children/manage/*-manage.vue`；分配状态为页内状态。切页签只保留 `accountId`；地址里的类型不可用时 replace 到第一个可用页签。
- 云账号树 `account/vendor-group/` 是原资源接入树的副本（过渡，改用 `account/use-resource-account.ts`：列表 / 新建页切账号只改 query，其它子页回对应类型列表）；筛选写入地址单会替换为 PR 版地址驱动的树。
- 账号摘要条 `account/account-summary.vue` 只读 store、不发请求；「查看账号状态」打开只读侧滑 `account/account-status/`（打开时才拉详情，`render-directive="if"` 关闭即停轮询）。地址带 `showAccountStatus=1` 且有 `accountId` 时自动打开并立刻去掉该参数。
- 账号接入弹窗在 `account/account-import/`（原 `createAccount`），`use-secret-extension` 也被账单录入一级账号等页面引用。
- 资源详情 `:resourceType/details/:id`（`MENU_RESOURCE_DETAIL`，排在列表之前）渲染共用的 `views/resource/resource-manage/resource-detail.vue`；列表页签与详情地址同一套类型名（`RESOURCE_TYPES`：弹性 IP 为 `eip`、路由表为 `routing`；GCP 防火墙、安全组模板详情另为 `gcp`、`template`）。旧名只做兼容：`resource-detail.vue` 的 `resourceTypeToComponentKey`（`ip` / `eips → eip`、`route → routing`），列表 `resource-list.vue` 的 `DETAIL_TYPE_TO_TAB`（同上，另把 gcp / template 转到安全组对应 `scene`）。注意 `eips` 仍是接口路径与 `useColumns` 的列配置 key，与路由类型名无关。站内跳详情按 PR 内联路由对象，云厂商 query 键统一 `vendor`（详情组件读 `query.vendor || detail.vendor`）。旧地址 `/resource/detail/:resourceType?id=&type=` 由 `resourceDetailLegacy` 转址（`type` 转 `vendor`，缺 id 回对应列表）。
- 详情组件与业务详情共用：取 ID、按钮鉴权都经 `hooks/use-resource-detail-route.ts`，两侧都是 `hcm-auth`——按路由选签名：资源纳管详情 `AUTH_*_IAAS_RESOURCE` + 账号 ID，业务详情 `AUTH_BIZ_*_IAAS_RESOURCE` + 业务 ID（不能按「有无业务 ID」判断，`getBizsId()` 会回退 localStorage）；关联 ID 未就绪时不鉴权、按钮禁用。
- 业务详情（`router/module/business.ts`，组件 `views/business/business-detail.vue` 按路由 name 取详情组件）：name 为 `MENU_BUSINESS_*_DETAILS`，路径 `<类型>/detail/:id?`，`normalizeDetailId` 把旧 `?id=` 转成路径参数；站内跳转按 PR 内联 `{ name, params: { id } }`，另带当前 `?bizs=`（`/business/:bizId` 前业务 ID 在 query，按 name 跳转不会自动带上）；切换业务时 `business-selector.tsx` 按业务详情 name 列表回 `menu.relative` 列表。

## 回收管理（`views/recyclebin/`）

资源侧回收站独立页面（`/resource/recyclebin`，`MENU_RESOURCE_RECYCLEBIN`），无云账号树：页签右侧的账号下拉只改地址 `accountId`，列表条件与回收保留时间都跟随它（账号详情失败视为无效账号，回到全部账号）；`type`（cvm / disk）、`cvm`（所属主机跳转的搜索值）也在地址里。原资源接入「回收站」页签地址 `/resource/resource/recycle` 经 `recyclebinRscLegacy` 原样带参转来。业务回收站仍用 `views/resource/recyclebin-manager/`。`useQueryCommonList` 只深度监听 `filter`、不立即请求，页面靠之后注册的 `immediate` 监听改写规则触发首屏请求。

## 账号管理（`views/account-manage/`）

资源运营 → 云资源 → 账号管理（过渡位置，三期迁工作台）：列表 `/resource/account`（`MENU_RESOURCE_ACCOUNT_MANAGEMENT`，三类账号都展示）、登记 `/resource/account/check-in`（复用 `views/service/service-apply/account-add`）、详情 `/resource/account/details/<ID>` 自动进 `/basic`（只有基本信息页签，账号 ID 在 params）。列表与详情声明 `auth.view: AUTH_FIND_ACCOUNT`，详情子路由要重复 `auth.view` / `owner` / `activeKey`（meta 浅合并，子路由 `Meta` 的空 `auth` 会盖掉父路由）。编辑 / 删除按实例鉴权（`AUTH_UPDATE_ACCOUNT` / `AUTH_DELETE_ACCOUNT` + `relation: [账号 ID]`），`RenderDetailEdit` 传 `authSign` 即用 `hcm-auth` 包编辑按钮。旧地址（原账号页签、资源状态、旧详情、两个旧登记入口）由 `accountManageLegacy` 转址：带账号的「账号信息 / 资源状态」进资源纳管并带 `showAccountStatus=1` 打开账号状态侧滑。站内跳账号详情一律按 `MENU_RESOURCE_ACCOUNT_DETAIL`（资源详情里的 `link` 用字符串 `/#/resource/account/details/<ID>`）。

## 关键流程 / 注意事项

- 资源接入各 `children/manage/*-manage.vue` 的 toolbar 在 `isResourcePage` 时取 `justify-content-end`：内容溢出会往**左**溢并被裁掉（先没的是 `toolbar-prefix` 里的 slot 前缀），不是常见的右侧溢出，也无法滚动看到。所以这一行必须留一个可压缩项——右侧搜索框容器给 `min-width: 0`、搜索框 `max-width: 100%` 加可用下限；同时给非搜索的直接子元素 `flex-shrink: 0`，否则收缩量会按 flex-basis 摊到按钮上把文字压掉。注意 slot 内容带的是父组件 scope id，子组件的 `> *` 选不中，需要在各自组件里自管（如 `resource-subtype-switch` 自带 `flex-shrink: 0`）。
- 共用的 `*-manage.vue` 在资源纳管下不能再按 `route.query.type` 判断资源类型（类型在 `route.params.resourceType`）；`security-manage.vue` 切子场景只改 query 留在当前路由。`use-filter.ts` / `use-filter-host.ts` 靠「`props.filter` 变化 → 置空 `searchValue` → 重算内部 filter」把云账号树的账号 / 云厂商变化传给真正发请求的那份 filter，这一步按地址正则 `/\/resource\/(resource|manage)/` 放行；地址再变时必须同步，否则只剩「原地改同一条规则对象」的情况会请求（如账号之间切换），全部账号 → 账号、选云厂商都不请求。
- 账号「资源状态」（`account/account-status/resource-status.tsx`，原 `accountInfo/component/resourceStatus/`）：表格「资源名称」列直接用 `RESOURCE_TYPES_MAP` 转译 `sync_details` 返回的 `res_name`，后端上新资源类型时前端只需在该 map 补一行（独占集群是 `load_balancer_exclusive_cluster`）。轮询回调**不能**吃 watch 回调的账号形参：轮询只创建一次，账号会被闭包永久固定，切账号后后续轮询全打到上一个账号。账号只能在 `getList` 内现取；轮询用 `useTimeoutPoll`，watch 里 `reset()` + `resume()` 重置轮次。
- 资源侧 CLB 同步复用业务侧同一套 `use-clb-sync-feedback`（成功 Toast、空 id、`2000002`、跳任务详情）。
- 跳转进的是业务任务管理详情；`bizs` 取最近一次业务选择，没有单独的资源任务入口。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
