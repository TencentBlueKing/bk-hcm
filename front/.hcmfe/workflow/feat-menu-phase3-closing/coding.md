# Coding — feat-menu-phase3-closing

需求文档：[`docs/reqs/三期收尾体验.md`](../../../../docs/reqs/三期收尾体验.md)（TAPD [1069995598139342337](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598139342337)）。分支 `feat-menu-phase3-workbench`，起始 HEAD `90f1a72ed`。

## 范围说明

原计划依赖 A2–A5，但 **A5（服务请求单据管理下线）只动顶栏壳入口，不影响 `router/module/` 退场**，因此本单先做路由文件退场与终扫，A5 之后无需回改。

## 改动

| 文件 | 改动 |
|---|---|
| `src/views/index.ts` | 工作台视图改为**视图级内联汇总**：`serviceViews = [...ticketRoutes, ...permissionPolicyRoutes, ...scheme, ...accountManageRsc]`，新增 `ticketRoutes` / `permissionPolicyRoutes` 两个 import |
| `src/router/module/service.ts` | 删除（薄壳：`/service` 父记录已在 `router/index.ts`，`meta.groupTitle` 全仓无消费方）；`src/router/module/` 目录随之为空并删除 |
| `src/views/service/my-apply/`（13 个文件） | 删除：无任何路由加载该页面组件，`/service/my-apply` 只剩 redirect 到 `/service/ticket`，目录内只有自引用 |
| `src/views/service/my-approval/`（2 个文件） | 删除：全仓 0 引用（ITSM `my_approval` 接口封装同样未被调用），从未接入路由 |

**没有新建 `views/service/route-config.ts`**：`views/<模块>/route-config.ts` 的约定只给真实页面模块（ticket / scheme / account-manage 各自都有）；工作台是视图级聚合，按 PR #1897 的原则直接在 `views/index.ts` 组装，不再多一层中间文件。

## 终扫结果（编码后复跑）

| 检查 | 结果 |
|---|---|
| `rg "MENU_RESOURCE_ACCOUNT_" src` | 0 |
| `rg "router/module" src` | 0 |
| `rg "workbenchViews\|entry-wb" src` | 0 |
| 旧字符串路由名 `'scheme-recommendation'` / `'scheme-list'` / `'scheme-detail'` | 0（只剩 CSS 类名 `class='scheme-list'`） |
| `'/scheme` 硬编码地址 | 0（PR 口径不做旧地址兼容） |
| `src/router/module/` | 已删除 |
| `views/service/route-config.ts` | 未创建（视图级汇总内联在 `views/index.ts`，符合 PR 原则） |
| `views/service/my-apply`、`views/service/my-approval` | 已删除；`/service/my-apply` 的两条 redirect 仍保留 |

## 发现但未改（留给合入内部版时决定）

`src/hooks/useWhereAmI.ts:23` 仍保留 `/^\/scheme\/.+/` → `Senarios.scheme` 的判断。A3 把资源选型迁到 `/service/scheme/*` 后该分支已成为死代码，`isSchemePage` 在全仓 0 消费方。

不在本单删除的原因：

1. PR #1897 同样保留了这个分支，删掉会与 PR 产生无谓差异；
2. `isSchemePage` 是 hook 的公开字段，内部版可能有消费方，删除会让合入冲突面变大；
3. 收尾单的边界约定是「命中项若属二期遗留或内部版独有，记录在文档而非强行修改」。

合入内部版（A7）时按内部版实际消费方决定：无消费则连同 `Senarios.scheme` / `isSchemePage` 一起删，有消费则改为匹配 `/^\/service\/scheme/`。

## 合入内部版注意：内部版的 `router/module/service.ts` 远不止本仓这几行

本仓删掉的 `service.ts` 只有 12 行（ticket + 权限策略库），但**内部版的同名文件有 296 行**，承载大量 BCC 独有路由。合并到内部版时，不能按「本仓删除 → 内部版直接删除」处理，否则这些页面会全部失去路由。

核对 `bcc/bcc/v1.9.x:front/src/router/module/service.ts` 得到的清单，合并时需逐项迁到 `views/service/route-config.ts`（或对应视图的 route-config）：

| 内部版路由 | 说明 |
|---|---|
| `...dissolveRoutes`（`views/dissolve/route-config`） | 机房裁撤，B 线要迁入工作台·公共事务 |
| `/service/resource-plan/cvm/home`、`/cvm/detail`、`/gpu` 等 | 资源预测（CVM / GPU），含 `beforeEnter` 里的 `useVerify` 权限分支（`ziyan_resource_plan_manage` / `ziyan_resource_plan_gpu_demands`） |
| `/service/resource-plan`、`/resource-plan/home`、`/resource-plan/detail` | 资源预测旧地址 redirect |
| `/service/hostInventory`、`/service/hostApplication`（含 detail / apply / modify）、`/service/hostRecycling` | 主机申请与回收，name 多为字符串（`'host-application-detail'`、`'提交主机申请'` 等） |
| `/service/cvm-model`、`/service/cvm-subnet`、`/service/resource-manage` | 机型 / 子网 / 资源纳管旧入口 |
| `/service/my-apply/resource-plan/detail` | 旧地址 → `/service/ticket/resource-plan/detail` 的兼容跳转 |

合并时的三个坑：

1. **旧 meta 字段**：内部版这些记录用 `title` / `isShowBreadcrumb` / `notMenu` / `icon`，二期的 `Meta` 已改为 `menu.i18n` / `layout.breadcrumb` / `activeKey`，必须逐条换算，不能原样搬。
2. **旧鉴权**：`useVerify()` + `authVerifyData.permissionAction` 是二期的旧鉴权写法，迁入时要换成路由 `meta.auth.view` + `views` 层的 `hcm-auth`，或至少在 A7 里明确保留原因。
3. **name 是字符串**：这些页面仍用字符串 name（含中文 name，如 `'提交主机申请'`），与三期「name 用菜单符号」的约定不一致；A7 需判断哪些要改符号、哪些保留（改符号会影响收藏键与站内 `router.push({ name })`）。

结论：**A7 的 `service.ts` 退场不是删除，而是「按视图拆分 + 老写法换算」**，工作量主要在内部版独有路由上，本单只保证本仓侧已清空。

### 内部版的 `my-apply` 有正在使用的组件，不能跟着删

本仓删掉 `views/service/my-apply/`（13 个文件）与 `views/service/my-approval/`（2 个文件）后，合入内部版时**不能直接接受这两处删除**：

| 目录 | PR | 本仓 | 内部版 | 合并处理 |
|---|---|---|---|---|
| `views/service/my-apply/` | 保留（13 个文件，与本仓一致） | 删除 | **15 个文件**，多出 `components/bpass-apply-detail/`（`index.tsx` + `index.scss`） | 只删「页面级」文件；`bpass-apply-detail/` 必须保留或迁到 `views/ticket/children/apply-detail/components/` 之类的位置 |
| `views/service/my-approval/` | 保留（2 个文件） | 删除 | 2 个文件（与本仓一致） | 内部版同样 0 引用，可直接跟着删 |

内部版还有两处**正在 import** `bpass-apply-detail` 的地方，删目录会直接编译失败：

- `front/src/views/ticket/children/apply-detail/apply-content-render-internal.plugin.tsx:5`
- `front/src/views/ticket/children/apply-detail/index.tsx:11`（用到 `BpaasEndStatus`）

即：`my-apply` 在内部版不是纯死代码，它是「单据详情的 BPAAS 结束状态 / 内容渲染插件」的宿主目录，只是公共仓没有这两个文件。

## 其他收尾项

- **收藏与折叠**：不改代码。账号管理符号改名后旧收藏键（`menu_resource_account_*`）由 `getFavoriteMenus` 自动清理；分组折叠按 `topMenuKey`（`MENU_SERVICE` 未变）保留。由 test.md 覆盖验证。
- **服务请求壳兜底**：A5 才会处理入口去留（Q-002），本单不动。
- **文档**：`menu-route` 中「`router/module/*` 尚未退场」的表述与 `views/service/route-config.ts` 的位置，随 test 阶段文档收口一并更新。

## 验证

| 项 | 结果 |
|---|---|
| `npx eslint src/views/service/route-config.ts src/views/index.ts` | 通过 |
| `npx tsc --noEmit -p tsconfig.json` | 改动文件无新增错误 |
| 终扫 6 项 | 见上表，全部符合预期 |
