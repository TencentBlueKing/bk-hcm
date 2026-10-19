# 模型驱动与页面模式

> status: stub · kind: module
> globs: `src/model/**`

装饰器字段元数据模型与多云 Factory、展示/表单/列表三场景模式。

## 职责

字段元数据模型。本轮任务相关：`src/model/task/search.view.ts`、`src/model/task/detail.view.ts`。

## 关键流程 / 注意事项

- 任务列表账号字段：`account_ids` 查询 op 为 `json_overlaps`（数组字段重叠，不是 `in`）。
- 同步任务详情列由 detail 模型 / action-list fields 对齐 `cloud_lb_id` 等协议字段，不混用其它任务类型的 `cloud_clb_id`。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
