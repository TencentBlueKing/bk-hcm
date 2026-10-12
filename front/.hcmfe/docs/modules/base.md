# 基础/共享接缝

> status: stub · kind: base
> globs: `src/api/**`, `src/assets/**`, `src/common/**`, `src/components/**`, `src/constants/**`, `src/css/**`, `src/decorator/**`, `src/directive/**`, `src/hooks/**`, `src/http/**`, `src/language/**`, `src/store/**`, `src/style/**`, `src/types/**`, `src/typings/**`, `src/utils/**`, `src/vendor/**`

跨模块共享层：HTTP 客户端、API 封装、全局组件、hooks、store、常量/类型、样式与工具函数等。已拆出 model/menu-route/auth 三个基础能力模块。

## 职责

跨模块共享层。本轮触及的接缝：`src/components/sync-account-resource`（账号+地域条件同步弹窗）；资源接入列表筛选用 `src/components/resource-search-select`。有 `fields`（字段模型）时由组件映射成搜索项；否则仍按 `ResourceTypeEnum` 从 `option-common.ts` 取字段（CVM 等存量）。新的资源接入列表不要在 `option-common` 再维护一份字段表。从 URL 回填时异步字段（业务/云账号）只有 id，展示名由组件用 `getOptionMenu` 取一次并缓存，调用方不再各写一遍名称映射。

## 关键流程 / 注意事项

- 弹窗只做通用提交、默认成功提示和关窗。可选 `successHandler` / `errorHandler` 给调用方。
- **禁止**按 `resourceName` 在弹窗内写死某一业务的成功/失败文案、错误码或跳转（CLB 反馈在 `views/load-balancer/use-clb-sync-feedback.ts`；安全组走默认「已同步成功」+ 已有 `errorHandler`）。
- `src/hooks` 只放跨模块 hook。
- `resource-search-select` 自带宽度：`width: 500px` 是上限，配 `max-width: 100%` + `min-width: 240px`。要它在窄屏收缩，**调用方容器必须给 `min-width: 0`**——flex 项的自动最小宽度等于内容宽（500px），不放开根本压不下去；未放开的使用方仍是固定 500px。
- 轮询统一用 `src/hooks/use-timeout-poll`（`resume`/`pause`/`reset`，`onScopeDispose` 自动停）。`src/utils/interval` 已 `@deprecated`：回调在创建时被闭包固定，取不到后续变化的参数（账号/ID 切换场景会一直打到上一个 id），也没有次数重置与 scope 销毁。
- 全局组件样式覆盖放 `src/style/override/`（`index.scss` 汇总）。`bkpopover.scss` 的 `.hcm-tooltips-popover` 用来让组件式 `bk-popover` 对齐 `v-bk-tooltips` 指令的紧凑暗色外观（两者默认 placement、arrow、内边距、背景都不同），表格单元格内应用组件而非指令。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
