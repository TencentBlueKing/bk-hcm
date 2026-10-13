# 权限控制

> status: stub · kind: module
> globs: `src/common/auth-service.ts`, `src/constants/auth-symbols.ts`, `src/components/auth/**`, `src/components/permission/**`, `src/hooks/useVerify.ts`, `src/store/common.ts`, `src/views/error-pages/403.tsx`

视图级权限与操作级权限。旧 useVerify/PermissionDialog 已废弃。

## 职责

两层权限：视图级（能否进入页面，路由 `meta.auth.view` + 路由守卫 + `views/status/permission.vue`）与操作级（能否执行某个操作，按钮禁用 + 权限申请弹窗）。操作级的关键文件：

| 文件 | 职责 |
| --- | --- |
| `constants/auth-symbols.ts` | 权限点 symbol 常量，业务代码只认 symbol |
| `common/auth-service.ts` | symbol → 权限定义（IAM action id、hcm action/resourceType、transform）；`getAuthResources` 把 `IAuthSign.relation` 转成鉴权请求项 |
| `components/auth/auth.vue`（`hcm-auth`） | 包裹按钮：合并鉴权请求、无权限置灰并接管点击 |
| `hooks/use-permission-dialog.ts` | 全局弹窗上下文，由 `app.vue` 提供，`show()` 触发 |
| `components/permission/apply-dialog.vue` | 无权限弹窗：权限清单表格 + 去申请 |
| `store/auth.ts` | `verify` 鉴权、`getApplyPermUrl` 取 IAM 申请链接 |

旧链路（`hooks/useVerify.ts`、`components/permission-dialog/`、`store/useGlobalPermissionDialog.ts`）已废弃，新代码禁止使用，迁移口径见 rule `fe-auth-migration`。

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

### 现存缺口

`app.vue` 把弹窗实例挂到 `window.hcmPermissionDialog`、`http/index.ts` 在 403 时 `bus.$emit('show-forbidden')`，但两者都没有消费方——接口返回无权限时并没有全局兜底弹窗，只会落到各自的错误提示。因此新增操作必须自己做预鉴权，不能指望兜底。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
