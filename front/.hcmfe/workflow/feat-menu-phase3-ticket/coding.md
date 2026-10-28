# Coding — feat-menu-phase3-ticket

需求文档：[`docs/reqs/三期单据管理迁移.md`](../../../../docs/reqs/三期单据管理迁移.md)（TAPD [1069995598139342281](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598139342281)）。分支 `feat-menu-phase3-workbench`，起始 HEAD `2e2d24fb6`。

## 目标

单据管理在工作台下的「菜单归属 + 页面上下文 + 站内跳转」收口。**不含数据权限**（已移出到 `1069995598139399484` 工作台个人数据权限与后端方案对齐）。

## 起点：A1 已完成的部分（本单不重复做）

- 菜单：`MENU_SERVICE_TICKET_MANAGEMENT` 已归入工作台「个人事务」分组（`menu-service.ts`），分组图标 `bkhcm-icon-user-8`
- 顶栏：`MENU_SERVICE` 显示名为「工作台」并前移到业务资源之后；「服务请求」已是顶栏跳转壳 `MENU_SERVICE_REQUEST`（无侧栏、点击跳工作台并提示迁移）
- 地址：`/service/ticket`、`/service/ticket/detail` 未变；`/service/my-apply` 等旧地址仍由 `ticket/route-config.ts` 转址

## 本单改动

| 文件 | 改动 | 原因 |
|---|---|---|
| `src/views/service/service-apply/hooks/use-vpc-form-data.ts` | 提单成功后跳「我的单据」由 `router.push({ path: '/service/my-apply' })` 改为 `routerAction.redirect({ name: MENU_SERVICE_TICKET_MANAGEMENT }, { history: true })` | 原来依赖旧地址转址，且直接调用 vue-router 实例方法，违反 `fe-router-action` 规范 |
| `src/views/service/service-apply/hooks/use-disk-form-data.ts` | 同上（硬盘提单） | 同上 |
| `src/views/ticket/entry-srv.vue` | 删除 `provide('isServicePage', true)`，改为注释说明工作台按路径判定归属 | 该 provide 全仓无消费方（只有 `isBusinessPage` 被 `common-table.vue` 注入），保留会误导后续以为存在「服务请求」上下文 |

## 未改动（有意，均已与 PR #1897 对齐核对）

- **地址与路由表**：PR #1897 `front/src/router/index.ts` 的 `MENU_SERVICE` 顶层记录同样是 `path: '/service'`，`views/ticket/route-config.ts` 子路由为相对路径 `ticket` / `ticket/detail`，即 `/service/ticket` 不变。我们的实现与 PR 一致，无需改地址或加转址。
- **页面归属判定**：PR #1897 的 `front/src/hooks/useWhereAmI.ts` 同样保留 `Senarios.service` 与 `/^\/service\/.+/` 正则、`isServicePage` 字段；我们保持相同实现（多出的 `accountStore.bizs` 回退是二期遗留，与本单无关）。
- **数据范围**：列表仍只带页签 `operation` 条件，个人过滤与审批数据由权限单统一设计
- **面包屑**：明细页 `menu.i18n` 为「申请单详情」，`menu.relative` 指回列表，符合二期约定

## 验证

| 项 | 结果 |
|---|---|
| `npx eslint use-vpc-form-data.ts use-disk-form-data.ts entry-srv.vue` | 通过 |
| `npx tsc --noEmit -p tsconfig.json` | 改动文件无新增错误（仓内既有错误与本单无关） |
| 提单后跳转 | 云主机 / 硬盘提单成功后进入 `/service/ticket`；直接调用不会被旧地址重定向多一跳 |
| 明细返回 | 单据明细面包屑为「申请单详情」，返回回到列表，页签 `type` 保留 |
| 顶栏与侧栏高亮 | 在单据页面顶栏高亮工作台、侧栏高亮「个人事务 → 单据管理」 |

## `provide('isServicePage')` 删除依据（含内部版核对）

- 公开仓：`views/ticket/entry-srv.vue` 的 `provide('isServicePage', true)` **没有任何 `inject` 消费方**；`common-table.vue` 只注入 `isBusinessPage`。
- 内部版（hcm 目录）：确有使用，但不依赖 ticket 入口的 provide——
  - `views/resource-plan/entry-srv.vue:33`、`views/ticket/children/resource-plan/list/list-srv.vue:75` 各自 provide；
  - 消费方 `views/ticket/children/resource-plan/list/children/data-list/data-list.vue:21` 的最近父级就是 `list-srv.vue`，删掉 ticket 入口的 provide 后仍能取到 `true`；
  - 唯一用途是 `max-height` 的三元表达式，两侧取值为同一数值。
- 结论：公开仓删除安全；合入内部版时该文件按本分支版本覆盖，消费方不受影响。合入后若内部版另有新增消费方，再按需补回。

## 追加修复：权限策略库操作列 hover 出现 tooltip（2026-10-26）

**现象**：权限策略库列表与「已应用」表格的表级 `bk-table` 都开了 `show-overflow-tooltip`，操作列内的按钮被当作溢出内容，hover 时弹出 tooltip，即使文字并未截断。

**项目惯例**：其他列表在操作列上显式关闭，如 `permission-template/children/list/data-list/data-list.vue:115`、`load-balancer/clb/load-balancer-table.vue:421`、`load-balancer/exclusive-cluster/data-list.vue:110`、`load-balancer/listener/layer7/rule-collapse-panel.vue:345` 都是 `:show-overflow-tooltip="false"`。

**改动**：

| 文件 | 改动 |
|---|---|
| `src/views/cloud-account-manage/permission-policy/children/data-list/data-list.vue` | 操作列加 `:show-overflow-tooltip="false"` |
| `src/views/cloud-account-manage/permission-policy/children/apply-sideslider/update-applied-table.vue` | 同上（「已应用」表格的操作列） |

验证：`eslint`、`stylelint` 两个文件均通过；hover 操作列不再出现 tooltip，其他列仍保留溢出提示。

## 遗留

1. 数据权限（个人申请 / 我的审批）见 `1069995598139399484`，方案确认后实施。
2. 服务请求迁移提示的具体下线日期仍为「上线后 2 周」占位。
3. 合入内部版时确认 `ticket/children/resource-plan` 的 `inject('isServicePage')` 仍由 `list-srv.vue` 提供。
