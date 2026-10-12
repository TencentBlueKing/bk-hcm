# 资源运营一级视图（容器·待拆解）

> status: stub · kind: module
> globs: `src/views/resource/**`

一级视图 /resource/ 容器，聚合账号/回收站/资源纳管，待拆解。

## 职责

一级视图 `/resource/` 容器。资源 tab「负载均衡」挂 `src/views/load-balancer/entry-rsc.vue`（双入口壳）；CLB 列表在 `load-balancer-manage.vue`，独占集群列表在负载均衡模块 `exclusive-cluster/`。

资源分配弹窗 `src/views/resource/resource-manage/children/dialog/batch-distribution/`：工具栏「批量分配」读表格勾选；`open(rows)` 用入参行打开同一弹窗（不改勾选）。一条时标题为「{资源名}分配」。独占集群类型 `exclusive_clusters`，body key `exclusive_cluster_ids`。

## 关键流程 / 注意事项

- 资源接入各 `children/manage/*-manage.vue` 的 toolbar 在 `isResourcePage` 时取 `justify-content-end`：内容溢出会往**左**溢并被裁掉（先没的是 `toolbar-prefix` 里的 slot 前缀），不是常见的右侧溢出，也无法滚动看到。所以这一行必须留一个可压缩项——右侧搜索框容器给 `min-width: 0`、搜索框 `max-width: 100%` 加可用下限；同时给非搜索的直接子元素 `flex-shrink: 0`，否则收缩量会按 flex-basis 摊到按钮上把文字压掉。注意 slot 内容带的是父组件 scope id，子组件的 `> *` 选不中，需要在各自组件里自管（如 `resource-subtype-switch` 自带 `flex-shrink: 0`）。
- 账号「资源状态」页 `accountInfo/component/resourceStatus/`：表格「资源名称」列直接用 `RESOURCE_TYPES_MAP` 转译 `sync_details` 返回的 `res_name`，后端上新资源类型时前端只需在该 map 补一行（独占集群是 `load_balancer_exclusive_cluster`）。轮询回调**不能**吃 watch 回调的账号形参：轮询只创建一次，账号会被闭包永久固定，而 `resourceAccount` 是 `useResourceAccount` 异步取详情后才写进 store、离开 resource 页前又不会 `clear()`，结果切账号后首次请求对、后续轮询全打到上一个账号。账号只能在 `getList` 内现取；轮询用 `useTimeoutPoll`，watch 里 `reset()` + `resume()` 重置轮次。
- 资源侧 CLB 同步复用业务侧同一套 `use-clb-sync-feedback`（成功 Toast、空 id、`2000002`、跳任务详情）。
- 跳转进的是业务任务管理详情；`bizs` 取最近一次业务选择，没有单独的资源任务入口。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
