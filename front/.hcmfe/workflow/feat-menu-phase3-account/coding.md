# Coding — feat-menu-phase3-account

需求文档：[`docs/reqs/三期账号管理迁入.md`](../../../../docs/reqs/三期账号管理迁入.md)（TAPD [1069995598139342308](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598139342308)）。分支 `feat-menu-phase3-workbench`，起始 HEAD `f5a9500e2`。

## 目标与 PR 对齐

账号管理从「资源运营 → 云资源」迁到「工作台 → 公共事务」，符号按 PR 改名为 `MENU_SERVICE_ACCOUNT_*`，路由从 `resourceViews` 迁到 `serviceViews`（地址 `/resource/account*` → `/service/account*`，子路径不变），列表不再展示资源账号。

与 PR #1897 逐项核对（**按源码核对，不只依据需求文档**）：

| 项 | PR 做法 | 本单 |
|---|---|---|
| 符号命名 | `MENU_SERVICE_ACCOUNT_MANAGEMENT/_CHECK_IN/_DETAIL/_DETAIL_BASIC`，description 为 `menu_service_account_*` | 一致 |
| 路由前缀 | 挂在 `MENU_SERVICE` 下，`path: 'account'` / `'account/check-in'` / `'account/details/:accountId'` | 一致 |
| 视图汇总 | 并进 `serviceViews` | 一致 |
| 菜单归属 | PR 放 `group: 公共事务` | **按产品定义放 `个人事务`**（与单据管理同组；已确认偏离 PR） |
| 筛选去资源账号 | `ACCOUNT_TYPES.filter(v => v.id !== ACCOUNT_TYPE_ENUM.RESOURCE)`，**不动全局 `ACCOUNT_TYPES`** | 一致 |
| 请求过滤 | 规则重建后 `push({ field: 'type', op: 'neq', value: RESOURCE })`，不用默认值 | 一致（落在本仓的 `watchEffect` 重建之后） |
| 列表列兜底 | 列 `option` 仍用全量 `ACCOUNT_TYPES`，历史资源账号行仍可渲染 | 一致 |
| 目录结构 | PR 重组为 `accountInfo/` + `accountmanage/` | **不照搬**：按现仓库 `list.vue` + `details/` 结构实现（用户已确认；重组无功能收益，留待收尾单） |

## 改动清单

| 文件 | 改动 |
|---|---|
| `src/constants/menu-symbol.ts` | 4 个符号 `MENU_RESOURCE_ACCOUNT_*` → `MENU_SERVICE_ACCOUNT_*`，description 同步改为 `menu_service_account_*`；定义从「资源运营」区移到新增的「账号管理（工作台 → 公共事务）」区 |
| `src/views/account-manage/route-config.ts` | `meta.owner` 由 `MENU_RESOURCE` 改 `MENU_SERVICE`（列表 / 登记 / 详情三处，含详情共用 meta）；新增旧地址转址 `/resource/account`、`/resource/account/details/:accountId`、`/resource/account/check-in`；原有 7 条 legacy 因按 name 跳转自动跟随 |
| `src/views/index.ts` | `accountManageRsc` 从 `resourceViews` 移到 `serviceViews` |
| `src/common/menu-service.ts` | 删除「资源运营 → 云资源 → 账号管理」过渡项；在工作台「个人事务」新增「账号管理」（位于单据管理之后，产品定义），`route` 从 `serviceViews` 取 |
| `src/views/account-manage/list.vue` | 账号类型筛选项去掉资源账号；`watchEffect` 重建 rules 后固定追加 `type neq resource` |
| `src/constants/view-auth.ts` | 配置项 `viewId` 跟随改名（`MENU_SERVICE_ACCOUNT_MANAGEMENT`），`hideMenu` 语义不变 |
| 其它 9 个引用文件 | `useRouteLinkBtn.tsx`、`account-manage/details/basic.tsx`、`views/resource/resource-manage/**`（vpc / subnet / drive / ip / 网络接口 / 路由表 / 安全组等 7 处跳账号详情）随符号改名自动跟随，共 50 处引用 |

## 验证

| 项 | 结果 |
|---|---|
| `npx eslint`（主要 7 个文件） | 通过 |
| `npx tsc --noEmit -p tsconfig.json` | 改动文件无新增错误 |
| `rg "MENU_RESOURCE_ACCOUNT_" src` | 0 处残留 |
| `rg "getMenuRoute(resourceViews, MENU_SERVICE_ACCOUNT" src` | 0 处 |
| `ACCOUNT_TYPES` | 未改动全局常量（其它入口仍可用） |

## 遗留

1. 目录结构未按 PR 重组（`accountInfo/` + `accountmanage/`）。**这是二期已评估并最终确定的决策**，不是本期遗留：现结构 `list.vue` + `details/` 与 PR 两套目录功能等价，重组无收益，后续任何一期都不再按 PR 拆分。
2. 符号改名后旧收藏键（`menu_resource_account_*`）成为孤儿，由 `getFavoriteMenus` 自动清理；未做旧键→新键映射。
3. `/resource/account` 系列旧地址已转址，但站内如果有第三方系统硬编码该地址，需要另行确认。
