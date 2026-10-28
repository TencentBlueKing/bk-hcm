# 单据/工单

> status: stub · kind: module
> globs: `src/views/ticket/**`

工单流程，含业务 / 工作台两套入口。

## 职责

`route-config.ts` 注册业务与工作台两套单据入口（`entry-biz.vue` / `entry-srv.vue`）；各列表页签共用 `children/common-table.vue`，通过 `hooks/useTable/useTable.tsx` 生成搜索与表格。旧“我的申请”路由重定向到单据入口。

三期起工作台侧单据挂在「工作台 → 个人事务」下（一级菜单仍是 `MENU_SERVICE`，地址仍为 `/service/ticket`）；「服务请求」降为顶栏跳转壳（`MENU_SERVICE_REQUEST`），自身无路由、无侧栏。

## 关键流程 / 注意事项

- 列表搜索是共享 `SearchSelect`，字段为单号、申请类型、状态、申请人（`constants.ts`），不是 task 的多列 Grid 搜索。工具栏搜索框允许收缩；不直接开启工具栏换行，因为 `useTable` 按固定 48px 扣除其高度。
- 申请详情按类型分派；`children/apply-detail/clb.vue` 使用弹性两列 Grid，`common-apply-detail/index.vue` 展示通用申请的基本信息及参数。长文本在值区换行；已有参数 overflow tooltip 保留。
- 账号申请详情还有 `AccountApplyDetail → DetailInfo → InfoList` 独立链路，不应只改 Grid 后假定覆盖所有详情。列表固定列与表格自身横滚保持各自职责。
- `entry-srv.vue` **不再** `provide('isServicePage')`：公开仓无任何 `inject` 消费方；内部版的消费方（`views/ticket/children/resource-plan/**`）由它自己的 `list-srv.vue` / `resource-plan/entry-srv.vue` 提供，不依赖本入口。
- 列表默认只带页签的 `operation` 条件。**不要**在本模块私自追加 `applicant` 等个人过滤——「只看个人单据 / 我的审批」属于三期工作台个人数据权限方案（与后端统一后实施）。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
