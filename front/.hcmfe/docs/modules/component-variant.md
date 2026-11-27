# 组件级多云与变体差异化

> status: drafted · kind: module
> globs: `src/**/*.plugin.ts`, `src/**/*-internal.plugin.ts`

单组件同时应对多云与构建变体。典范：account-selector。

## 职责

账号选择器 `src/components/account-selector/index-new.vue` 使用 `data-factory.ts` 返回的数据源与 `vendorProperty`。账号响应先按该厂商配置过滤，再传入默认或调用方的 `filter`；厂商标签、账号计数与自动选中均基于过滤后的列表，避免外部版本展示内部厂商账号。

## 关键流程 / 注意事项

默认筛选在 `account-selector/filter.class.ts` 中按资源类型编排，例如负载均衡通过 `filter.plugin.ts` 只保留腾讯云。子网支持多云，不应直接复用负载均衡的腾讯云单厂商限制。指定厂商时，同步弹窗提供的自定义筛选也必须接收已排除当前版本不支持厂商的账号列表。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
