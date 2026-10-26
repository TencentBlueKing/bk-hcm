# Coding — feat-menu-phase3-scheme

需求文档：[`docs/reqs/三期资源选型并入.md`](../../../../docs/reqs/三期资源选型并入.md)（TAPD [1069995598139342301](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598139342301)）。分支 `feat-menu-phase3-workbench`，起始 HEAD `701ddf483`。

## 目标与 PR 对齐结论

资源选型从顶栏一级菜单移入「工作台 → 公共事务」，路由迁到 `views/scheme/route-config.ts` 并挂 `serviceViews`，name 由字符串改为菜单符号，**地址随之变为 `/service/scheme/*`**。

与 PR #1897 逐项核对：

| 项 | PR 做法 | 本单 |
|---|---|---|
| 菜单 | `id: MENU_SCHEME`、`group: 公共事务`、`groupIcon: bkhcm-icon-user-8`、`visibility: ENABLE_CLOUD_SELECTION === 'true'` | 一致 |
| 路由 | `views/scheme/route-config.ts`：父 `name: MENU_SCHEME`、`path: 'scheme'`；子 `recommendation` / `deployment/list` / `deployment/detail` | 一致 |
| 视图汇总 | `serviceViews = [...ticketRoutes, ...scheme, ...accountManage]` | 一致（本期只并入 `scheme`） |
| 顶层路由 | 删除 `MENU_SCHEME` 顶层记录与 `schemeViews` | 一致 |
| 旧地址 | PR **不做** `/scheme/*` 兼容转址 | 一致：未加转址（站内 0 处链接，已 grep 确认） |

与原计划的唯一差异：原计划补 `/scheme/*` 兼容转址，核对 PR 后按 PR 做法取消。

## 改动清单

| 文件 | 改动 |
|---|---|
| `src/constants/menu-symbol.ts` | 新增 `MENU_SCHEME_RECOMMENDATION` / `MENU_SCHEME_LIST` / `MENU_SCHEME_DETAIL` |
| `src/views/scheme/route-config.ts` | 新增：资源选型路由，`meta.owner = MENU_SERVICE`、`activeKey = MENU_SCHEME`；推荐页与方案列表带 `auth.view`，三页 `breadcrumb.show = false`（沿用现网） |
| `src/views/index.ts` | `scheme` 改为从 `views/scheme/route-config` 引入并并入 `serviceViews`；删除 `schemeViews` 导出 |
| `src/router/index.ts` | 删除 `MENU_SCHEME` 顶层路由与 `schemeViews` 引用；`afterEach` 清 `accountStore.bizs` 的判断去掉 `MENU_SCHEME`（资源选型页面现在的 `matched[0]` 是 `MENU_SERVICE`） |
| `src/common/menu-service.ts` | 顶栏删除「资源选型」一级菜单；工作台「公共事务」新增「资源选型」，`groupIcon` 挂在该组第一项上 |
| `src/router/module/scheme.ts` | 删除（`router/module/` 只剩 `service.ts`，归 A2 后续/收尾处理） |
| `src/views/scheme/index.tsx` | 页签路由名改为符号；`isActived` / `handleTabChange` 类型改 `symbol`；`key` 用 `label`（符号不适合做 JSX key） |
| `src/views/scheme/scheme-list/index.tsx` | `goToCreate` / `goToDetail` 改符号跳转 |
| `src/views/scheme/scheme-detail/index.tsx` | 删除后返回列表改符号跳转 |
| `src/views/scheme/components/scheme-selector/index.tsx` | 返回列表 / 打开详情改符号跳转 |

## 验证

| 项 | 结果 |
|---|---|
| `npx eslint`（10 个改动文件） | 通过 |
| `npx tsc --noEmit -p tsconfig.json` | 改动文件无新增错误 |
| `rg "scheme-recommendation\|'scheme-list'\|'scheme-detail'\|schemeViews"` | 无路由名/视图残留（仅剩 CSS class 与文件目录名） |
| `router/module/` 现状 | 只剩 `service.ts` |

## 遗留

1. `/scheme/*` 旧地址未做兼容（与 PR 一致）；如果后续确认有外部书签，补三条 redirect 即可。
2. `router/module/service.ts` 尚未退场（三期收尾单处理）。
3. 资源选型的页面实现未改，仅路由装配与归属调整。
