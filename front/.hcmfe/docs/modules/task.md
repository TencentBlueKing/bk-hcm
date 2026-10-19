# 任务管理

> status: stub · kind: module
> globs: `src/views/task/**`

异步任务（cvm/clb 等）列表与详情。

## 职责

异步任务列表与详情（cvm / clb 等）。CLB 同步任务类型枚举 `sync_load_balancer`，列表展示「同步」（`src/views/task/constants.ts`）。

## 关键流程 / 注意事项

- 列表筛「同步」→ 查询 `operations=sync_load_balancer`。
- 账号条件：`account_ids` 的 op 为 `json_overlaps`（不是 `in`），模型在 `src/model/task/search.view.ts`。
- 同步任务详情操作列（`details/children/action-list/fields.ts`）：开始/结束时间、类别（新增/修改/删除）、CLB ID（`cloud_lb_id`）、CLB VIP/域名、任务状态、失败原因。无监听器类字段。
- `sync_load_balancer` 不支持失败重执行（入口禁用）。
- 资源侧 Toast 跳进的是**业务**任务详情，业务 id 取最近一次业务选择（`getBizsId()`）。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
