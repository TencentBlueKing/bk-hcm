# 权限控制

> status: stub · kind: module
> globs: `src/common/auth-service.ts`, `src/constants/auth-symbols.ts`, `src/components/auth/**`, `src/components/permission/**`, `src/store/auth.ts`, `src/router/business-interceptor.ts`

视图级权限与操作级权限。旧 useVerify / PermissionDialog / pageAuthData 已删除。

## 职责

两层权限：视图级（能否进入页面）与操作级（能否执行某个操作，按钮禁用 + 权限申请弹窗）。

### 视图级

页面查看权限 = 路由声明 + 启动预鉴权 + 守卫判定 + 框架内无权限页（named view），对齐 PR #1897，按现网拦截范围裁剪。

- **路由拦截** — 路由 `meta.auth.view`（`IAuthSign`，只用静态 sign），由 `router/index.ts` 的 `beforeEach` 调 `authStore.checkViewPermission(to.name, sign)` 判定；无权限把 `to.meta.view` 置为 `permission` 并挂 `permissionData`，地址不变。`views/index.ts` 的 `injectStatusComponents` 给各视图集合的带组件记录注入 `permission` / `error` / `notFound` 组件，`main-content.vue` 的出口 `RouterView :name="view"` 渲染对应页面。路由数组必须经它汇入，否则无权限时白屏。
- **业务资源** — 不按路由声明，由 `router/business-interceptor.ts` 在业务 ID 确定之后统一判（取值顺序见 menu-route）：解析不出任何业务 `noBiz`、业务 ID 不在全量列表 `bizNotFound`、不在授权列表 `bizUnauthed`。不在定业务 ID 之前用「没有任何业务权限」拦下带具体业务的地址。渲染 `views/status/business.vue`（`noBiz` / `bizUnauthed` 可申请，不弹全局申请弹窗）。
- **预鉴权与缓存** — `constants/view-auth.ts` 的 `viewAuthConfig` 在 `store/index.ts` 的 `preload` 中合成一次 `/auth/verify`，结果按 `viewId`（= 路由 name，Symbol 或字符串原样作键）缓存在 `store/auth.ts` 的 `viewPermissions`。未预鉴权的声明在首次访问时动态查一次后缓存。接口失败**不写缓存**：守卫再请求一次，仍失败进 error 页。
- **侧栏显隐** — `viewAuthConfig` 中 `hideMenu: true` 的项，`components/layout/menu.vue` 只在「已缓存且有权限」时显示同 id 菜单（接口失败也不显示）。菜单显隐与页面拦截解耦：云账单管理是「菜单按权限隐藏、直接访问可打开」（预鉴权 + `hideMenu`，路由不声明）；二级账号是「菜单常显、点进去拦截」（路由声明，不配 `hideMenu`）；一级账号、权限策略库、账号管理（`AUTH_FIND_ACCOUNT`）、资源侧回收管理（`AUTH_FIND_RECYCLE_BIN`）与操作记录（`AUTH_BIZ_FIND_AUDIT`，不传业务即资源审计）两处都配。
- **无权限页** — `views/status/permission.vue` 按 sign 取 `descriptions` 文案；先「去申请」，点过后换「已申请，刷新页面」。需由管理员主动授权的权限加进 `NO_SELF_APPLY_SIGNS`（当前为一级账号）：只提示联系管理员、不显示申请按钮。后端未返回 `permissionData` 时也不显示按钮。
- **旧链路已退场** — 旧 `/403/:id` 页已删除，旧书签落 404。启动阶段只有 `preload` 里视图权限那一次 `/auth/verify`，不再预取旧按钮权限（`store/common.ts` 只剩跑马灯状态）。
- **只读「是否有权限」、不渲染按钮** — 有对应视图预鉴权的，读缓存：`authStore.isViewPermissionCached(viewId) && authStore.hasViewPermission(viewId)`（例：账号申请单据详情按 `MENU_BILL_ROOT_ACCOUNT` 决定是否拉一级账号列表）。不要直接请求受保护接口再吞掉失败：`http/index.ts` 遇接口无权限会自动弹申请弹窗。

新增页面声明查看权限：路由 name 用 `menu-symbol.ts` 常量 → `meta.auth.view` → 有菜单且需按权限隐藏时在 `viewAuthConfig` 加 `{ viewId: 路由 name, authSign, hideMenu: true }` → 需要说明文案的在 `permission.vue` 的 `descriptions` 加一项。

### 操作级

关键文件：

| 文件 | 职责 |
| --- | --- |
| `constants/auth-symbols.ts` | 权限点 symbol 常量，业务代码只认 symbol |
| `common/auth-service.ts` | symbol → 权限定义（IAM action id、hcm action/resourceType、transform）；`getAuthResources` 把 `IAuthSign.relation` 转成鉴权请求项 |
| `components/auth/auth.vue`（`hcm-auth`） | 包裹按钮：合并鉴权请求、无权限置灰并接管点击 |
| `hooks/use-permission-dialog.ts` | 全局弹窗上下文，由 `app.vue` 提供，`show()` 触发 |
| `components/permission/apply-dialog.vue` | 无权限弹窗：权限清单表格 + 去申请 |
| `store/auth.ts` | `verify` 鉴权、`getApplyPermUrl` 取 IAM 申请链接 |

旧链路（`hooks/useVerify.ts`、`components/permission-dialog/`、`provide('authAction')`）已删除（菜单整合二期收尾单），站内操作级权限只剩 `hcm-auth` 一套。详情字段级编辑鉴权用 `info-list` 的 `getAuthSign`。

## 关键流程 / 注意事项

### 操作级鉴权链路

`hcm-auth` 的 `sign` → `getAuthResources` 按定义的 `transform` 产出 `{ action, resource_type, resource_id?, bk_biz_id? }` → `POST /api/v1/web/auth/verify`（相同 sign 经 CombineRequest 合并成一次请求）→ `noPerm` 通过插槽交给按钮控制禁用态 → 无权限时的点击由 `.hcm-auth.disabled::after` 覆盖层接管 → `usePermissionDialog().show(permission)` → `apply-dialog` 渲染清单 → 「去申请」用 `POST /api/v1/web/auth/find/apply_perm_url` 换 IAM 申请页链接。

后端按 `(resourceType, action)` 映射出真正的 IAM action（`cmd/auth-server/service/auth/gen_id.go`），前端定义里的 `id` 必须与 IAM 侧 action id 一致——弹窗就是靠它显示权限名。

### 分配（resource_assign）是唯一的双资源实例权限

- 后端所有资源的分配都收敛到同一个 IAM action `resource_assign`：`genIaaSResourceResource` 命中 `meta.Assign` 时直接转 `genCloudResResource`，主机、CLB、VPC、证书、独占集群概不例外。
- 权限定义里的 `resourceType` 只决定后端用哪个生成函数（`cmd/auth-server/service/auth/adaptor.go` 的 `genResourceFuncMap`），不参与 IAM 判定。分配统一用 `cloud_resource`——后端 `meta.CloudResource` 是「涵盖全部云资源」的特殊类型，直连 `genCloudResResource`。写具体类型（如 `cvm`）也能通，但要靠 `genCvmResource` 的 `default` 兜回 `genIaaSResourceResource`，绕且容易让人误以为只管主机。
- 它在 IAM 注册的 `RelatedResourceTypes` 是**云账号 + CMDB 业务**两个实例（`pkg/iam/sys/initial_actions.go`），所以鉴权必须同时给出账号与**目标业务**：`relation: [accountId, targetBizId]`，transform 产出 `resource_id` + `bk_biz_id`。
- 由此时机上有硬约束：目标业务是在分配弹窗里选的，**在打开弹窗前的按钮上鉴权必然不通过**（BizID 为 0，IAM 侧按业务实例 "0" 判定）。鉴权点应落在弹窗的「确定」上，未选业务时用 `hcm-auth` 的 `ignore` 跳过校验、由按钮禁用兜住。实现见 `views/resource/resource-manage/children/dialog/batch-distribution/index.tsx`，该组件是负载均衡、VPC、子网、云硬盘、弹性IP、参数模板、GCP 防火墙、证书、独占集群共用的分配入口。
- 主机与安全组不走这个组件（各有自研弹窗和专属接口）。安全组分配不选目标业务，取的是安全组自身的管理业务，接入鉴权时 relation 里的业务来源不同。

### 无权限弹窗的展示口径

- 表格一行对应一个 action；「关联的资源实例」格把该 action 下**全部** `related_resource_types` 的 instances 打平去重后逐行展示，分配场景即【账号】与【业务】两行。
- `instances` 是形如 A-B/C-D 的多层级路径结构，当前统一打平；注册的资源类型都还没有真实父链，故多层级未支持（代码内留 TODO）。
- 多个资源鉴同一权限时后端会返回重复实例，按 `type + id` 去重。
- 「去申请」提交的是未裁剪的整个 `permission` payload，展示层的裁剪不影响申请链接的内容。

### 接口无权限兜底

`http/index.ts` 在 HTTP 403 或业务码 `2030403`（后端 `PermissionDenied`，响应体 `permission` 带缺失权限）时调 `window.hcmPermissionDialog.show(permission)` 弹统一申请弹窗；无 `permission` 或弹窗尚未挂载（`preload` 阶段）时退回错误提示。弹窗单实例，后到的覆盖先到的。兜底只解决「漏了预鉴权」时的提示，按钮置灰仍要靠 `hcm-auth` 预鉴权。

旧全局弹窗（`components/global-permission-dialog`、`store/useGlobalPermissionDialog.ts`）与整页无权限（`views/resource/NoPermission.tsx`、`store/usePagePermissionStore.ts`）已删除。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
