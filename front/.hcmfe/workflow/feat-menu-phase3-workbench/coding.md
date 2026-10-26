# Coding — feat-menu-phase3-workbench

需求文档：[`docs/reqs/三期工作台骨架.md`](../../../../docs/reqs/三期工作台骨架.md)（TAPD [1069995598139342268](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598139342268)，验收以 AC-001～AC-007、AC-P01、AC-S01 为准）。分支 `feat-menu-phase3-workbench`，基线 `origin/v1.9.x`（`74e8d5a3c`）。

## 目标

顶栏由「业务资源 / 服务请求 / 资源接入 / 资源选型 / 资源运营」改为「业务资源 / 工作台 / 服务请求 / 资源接入 / 资源选型 / 资源运营」，其中：

- **工作台**复用 service 视图：一级菜单 id 仍是 `MENU_SERVICE`，只改显示名并前移到业务资源之后，`/service/*` 地址不变（不新建 `workbenchViews` / `entry-wb.vue`）。
- 工作台侧栏按 PR #1897 建分组：个人事务（单据管理）、公共事务（本期先放权限策略库，资源选型 / 账号管理 / 机房裁撤由后续子单加入）。
- **服务请求**改为顶栏跳转壳（与二期「资源接入」同范式），自身不注册路由、没有侧栏；点击跳「工作台」（落到工作台默认页）并提示迁移。提示文案只到一级模块，不写具体功能。

## 关键决策（2026-10-26 澄清）

侧栏是按 `route.matched[0].name` 反查的（`components/layout/menu.vue:24`、`common/menu-service.ts` 的 `getCurrentTopMenu`），没有路由的一级菜单不会渲染侧栏。因此：

- 服务请求不能既复用 service 视图又保留侧栏，只能做顶栏跳转壳；
- 原 Q4「权限策略库留在服务请求」修正为：**权限策略库过渡放在「工作台 → 公共事务」**，四期随服务请求一起迁走，用户已确认。

## 改动清单

| 文件 | 改动 |
|---|---|
| `src/constants/menu-symbol.ts` | 注释由「运营（服务请求）」改为「工作台」；新增 `MENU_SERVICE_REQUEST`（顶栏过渡入口，四期删除） |
| `src/common/menu-service.ts` | `serviceMenus` 改为工作台侧栏：单据管理 `group: 个人事务`、权限策略库 `group: 公共事务`，`groupIcon: bkhcm-icon-user-8`；`menus` 中 `MENU_SERVICE` 显示名改「工作台」并前移；新增 `MENU_SERVICE_REQUEST` 顶栏跳转项（`route` 指向 `MENU_SERVICE_TICKET_MANAGEMENT`） |
| `src/components/layout/header.vue` | 迁移提示由「只认 `MENU_RESOURCE_ACCESS`」改为 `migrationTips` 映射表，新增服务请求的迁移文案：只写到一级模块「工作台」，不写具体功能；下线时间用「上线后 2 周」占位，具体日期上线前补 |
| `src/components/layout/header.vue`（样式） | 顶栏一级菜单改自适应：`.header-nav` 加 `min-width: 0 / overflow-x: auto` 并隐藏滚动条，链接 `flex: 0 0 auto` + `white-space: nowrap`，右侧工具区 `flex: 0 0 auto`；≤1280px 紧凑模式（间距 8→4、内边距 16→10）。修复「一级菜单从 4/5 项增加到 6 项后文字被压到换行」 |
| `src/style/common.scss` | 仅改注释：说明 `min-width: 980px` 是主区可读下限，顶栏不再依赖该数值容纳菜单 |

未改：`views/index.ts`（仍导出 `serviceViews`）、`router/index.ts`（`MENU_SERVICE` 顶层记录与 `/service` 前缀保留）、`store/auth.ts`、`constants/view-auth.ts`（本期无符号改名，权限配置无需迁移）。

## 与后续子单的接口

- 「单据管理入工作台个人事务与服务请求跳转」：单据管理已在个人事务分组，本单的 `MENU_SERVICE_REQUEST` 顶栏壳即其跳转入口（文案只提「工作台」，不点具体功能）；该单不再新增侧栏跳转项。
- 「资源选型并入」「账号管理迁入」：把菜单项加到公共事务分组即可，本单已建好分组与图标。
- 「收尾」：确认 `MENU_SERVICE_REQUEST` 与迁移提示在工作台化完成后仍正确，下线时连同提示一并删除。

## 验证

| 项 | 结果 |
|---|---|
| `npx eslint src/common/menu-service.ts src/components/layout/header.vue src/constants/menu-symbol.ts` | 通过（0 error / 0 warning） |
| `npx stylelint src/components/layout/header.vue` | 通过 |
| 小屏顶栏（现象回归） | 980px / 1024px / 1280px 下：一级菜单不换行、不被裁切；超出宽度时顶栏内部横向滚动（滚动条隐藏），右侧工具区不受挤压 |
| `npx tsc --noEmit -p tsconfig.json` | 改动文件无新增错误（仓内既有错误与本单无关：`views/service/service-apply/vpc`、`views/ticket/children/apply-detail/**`、`node_modules/uuid`） |
| `getMenus()` 顺序 | 业务资源 → 工作台 → 服务请求 → 资源接入 → 资源选型 → 资源运营（`MENU_SCHEME` 待「资源选型并入」子单删除） |
| 服务请求壳 | `getMenuRoute(serviceViews, MENU_SERVICE_TICKET_MANAGEMENT)` 命中 /service/ticket，`router.hasRoute()` 通过，顶栏可见 |
| 工作台侧栏 | `/service/ticket` 命中 `matched[0].name = MENU_SERVICE` → 侧栏显示 个人事务 / 公共事务 |
| 收藏与折叠 | `MENU_SERVICE` 未改名，收藏键与折叠 key 不变；新增/移入分组的菜单项收藏键为各自 symbol，未改名不失效 |
| 权限 | 未改 `view-auth.ts`；权限策略库仍按原配置显示/隐藏 |

## 遗留

1. 资源选型仍是一级菜单（`MENU_SCHEME`），删除动作归「资源选型并入工作台公共事务」子单。
2. 服务请求迁移提示的下线日期用「上线后 2 周」占位，具体日期上线前补。
3. `docs/reqs/三期工作台骨架.md`、`三期拆单总览.md` 已同步「服务请求为跳转壳、权限策略库过渡放公共事务」的结论。
