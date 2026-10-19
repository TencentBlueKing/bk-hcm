# 操作记录

> status: drafted · kind: module
> globs: `src/views/operation-log/**`

操作审计日志。

## 职责

业务侧与资源侧两套入口共用列表、搜索、详情组件：`entry-biz.vue`（业务资源 → 操作记录）、`entry-rsc.vue`（资源运营 → 操作记录，`/resource/record`），详情 `details/flow-task`。路由在 `route-config.ts`：`operationLogBiz` 挂业务资源；`operationLogRsc`（`MENU_RESOURCE_OPERATION_LOG(_DETAILS)`）挂资源运营根，列表声明 `auth.view: AUTH_BIZ_FIND_AUDIT`（不传业务时后端按资源审计、任一账号判定），侧栏按同一权限 `hideMenu`；`operationLogRscLegacy` 承接原资源接入「操作记录」页签地址。

## 关键流程 / 注意事项

- 列表条件全在地址 `filter`（`useSearchQs`，`qs` 逗号格式、不编码）。`entry-rsc` / `entry-biz` 读到地址上有 `filter` 就不再套默认「近 30 天」，所以任何拼 `filter` 的跳转（如旧地址 `/resource/resource/record?accountId=` 的转址 `toRecordList`）都要把 `created_at` 一起写进去。
- 两个入口都自带 `padding: 24px; height: 100%`（主视图 `.main-view` 无内边距）。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
