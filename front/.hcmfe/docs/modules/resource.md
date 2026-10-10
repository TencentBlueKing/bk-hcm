# 资源运营一级视图（容器·待拆解）

> status: stub · kind: module
> globs: `src/views/resource/**`, `src/views/resource-manage/**`, `src/views/recyclebin/**`

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

## 回收管理（`views/recyclebin/`）

资源侧回收站独立页面（`/resource/recyclebin`，`MENU_RESOURCE_RECYCLEBIN`），无云账号树：页签右侧的账号下拉只改地址 `accountId`，列表条件与回收保留时间都跟随它（账号详情失败视为无效账号，回到全部账号）；`type`（cvm / disk）、`cvm`（所属主机跳转的搜索值）也在地址里。原资源接入「回收站」页签地址 `/resource/resource/recycle` 经 `recyclebinRscLegacy` 原样带参转来。业务回收站仍用 `views/resource/recyclebin-manager/`。`useQueryCommonList` 只深度监听 `filter`、不立即请求，页面靠之后注册的 `immediate` 监听改写规则触发首屏请求。

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
