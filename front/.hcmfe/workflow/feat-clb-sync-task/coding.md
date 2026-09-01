# Coding — feat-clb-sync-task

lite：无独立 PRD/Design/API 阶段。需求以 TAPD [#1069995598137677539](https://<TAPD_HOST>/tapd_fe/69995598/story/detail/1069995598137677539) 为准；契约以用户提供的《CLB 条件同步 — 前端接口协议》为准。UI 复用现网【任务管理-负载均衡】，不新开页面。

适用：`vendor=tcloud` / `tcloud-ziyan`，`res=load_balancer`。列表/详情 filter、分页与现网其它 CLB 任务相同。

## 执行顺序

1. 提交后提示与跳转 — `sync_by_cond` 成功读 `task_management_id`，可点进详情
2. 任务类型「同步」 — `TaskClbType.SYNC_LOAD_BALANCER = 'sync_load_balancer'`
3. 操作详情列 — 严格按协议「行字段 → 页面列」（7 列，`reason` 仅 failed）；VIP/域名换行拼接

> 排序依据：先改提交入口，再扩任务管理枚举与详情列。

## 共享改动 / 提交策略

- 跳转：`routerAction.open` 新窗口打开 `MENU_BUSINESS_TASK_MANAGEMENT_DETAILS`，params `{ resourceType: clb, id: task_management_id }`，query 带当前 `bizs`
- CLB 列表同步与监听器页「同步当前负载均衡」走同一弹窗/同一接口，成功态一并改，避免一处仍显示「已同步成功」
- 提交与关单：本迭代单 TAPD，一次提交

## 需求对照

| 编号 | 要落地的行为 |
|------|----------------|
| F-004 | 提交后提示「同步任务已创建，可在【任务管理-负载均衡】查看进度」，【任务管理-负载均衡】可点击跳到该任务详情 |
| F-002 | 【任务管理-负载均衡】展示并筛选「同步」 |
| F-003 | 操作详情列按协议「行字段 → 页面列」；有 `param.domain` 时与 VIP **换行**展示 |
| 边界 | `task_management_id === ''` → 「没有可处理的负载均衡」，不跳转 |
| 边界 | `reason` 仅 `state=failed` 有值 |
| 边界 | 未填地域：表单必填 + 后端 `2000001` 用 `message` |
| 边界 | `2000002` → 「该账号地域同步任务进行中」 |
| 不包含 | 百分比进度条；后端编排；改列表/详情分页协议 |

与 TAPD 原文差异（以接口协议为准）：

- TAPD「非自研云提示仅支持自研云」：协议无独立错误码。现网 CLB 同步弹窗传 `resourceType=CLB`，账号下拉默认 `filter.plugin` 只留 `vendor === tcloud`，AWS/Azure 等不会出现；选中账号后用其 `vendor` 拼 `sync_by_cond` 路径。前端没有 `tcloud-ziyan` 枚举。**本轮不在提交时再拦 vendor**，后端若拒则走 `message`
- TAPD「6 字段」：以协议「行字段 → 页面列」7 列为准；失败原因任意状态为空都显示 `--`
- 现网 `init` 文案是「待执行」，协议写「未执行」：**本轮不改全局文案**

## 落码入口（复用现网，不走 page-* 新建页）

| 能力 | 现网入口 | 本轮改动 |
|------|----------|----------|
| 条件同步弹窗 | `src/components/sync-account-resource/index.vue` | **只做通用提交**。默认成功 Toast「已同步成功」。调用方按需传 `successHandler` / `errorHandler`（有 `errorHandler` 才 `globalError: false`）。**禁止**在弹窗内按 `resourceName === 'load_balancer'` 写 CLB 文案/跳转。成功仍 `emit('success', res)` |
| CLB 成功/失败反馈 | `src/views/load-balancer/use-clb-sync-feedback.ts` | 业务与资源共用：`2000002` → 「该账号地域同步任务进行中」；其它错误用 `message`；成功按 `task_management_id` 分支（空 id → 「没有可处理的负载均衡」；有 id → 可点 Toast + `routerAction.open`）。跳转 `bizs`：调用方传入 `getBizId`，否则 `useWhereAmI().getBizsId()` |
| 列表 / 单 CLB / 资源同步 | `src/views/load-balancer/clb/load-balancer-table.vue`、`src/views/load-balancer/listener/listener-table.vue`、`src/views/resource/resource-manage/children/manage/load-balancer-manage.vue` | 三处都挂同一弹窗，并传入 hook 的 `successHandler` / `errorHandler`。业务两处另传 `:business-id`（账号过滤）；资源页不传，跳转用 `getBizsId()` |
| 跳转样例 | `src/views/load-balancer/device/main-content/listener-table.vue`、`src/views/load-balancer/clb/children/batch-import/index.vue` | 对齐 `task_management_id` → 任务详情；提示文案按 F-004 |
| 任务列表 / 搜索 / 详情基本信息 | 共享模型，**不按操作类型拆列** | 只往 `TaskClbType` + `TASK_CLB_TYPE_NAME` 加 `sync_load_balancer` → 「同步」。列表列来自 `ListView`（`getColumns` 忽略 resource）；CLB 搜索多一个「任务类型」来自 `SearchClbView.operations`（已是 `json_overlaps`）；详情基本信息来自 `Properties`，同一套 `TASK_TYPE_NAME`。不加新列表列 |
| 操作详情列 | `src/views/task/details/children/action-list/fields.ts` | **按 resource + operation 定制**：`fieldIdMap.get(CLB)[sync_load_balancer]` 只挂协议点名的字段 id（见下）。未登记会掉到 `baseFieldIds`。列名以协议为准，不要拿 `cloud_clb_id` / `region_id` 等其它类型字段去猜 |
| 重执行 | `src/views/task/details/clb.vue` + `fieldRerunIdMap` | 现网只给创建监听器 / 创建规则 / 绑 RS 配了 rerun 列；SOPS 四类（删监听器、权重、解绑）按钮「暂不支持」。协议与 TAPD **都没提** rerun。建议：`sync_load_balancer` **不支持**，并入禁用名单（否则未配列会误用 `clbBaseRerunParamFieldIds`：`cloud_clb_id` / protocol / listener_port，和同步 param 对不上） |
| 字段定义 | `src/model/task/detail.view.ts` | 补 `param.op`（只给同步列用）。VIP/域名：现网 `param.clb_vip_domain` 是**原样字符串**，列名「CLB VIP/域名」不是 `x/y` 展示格式。协议同步回包该字段**只有 VIP**，域名在 `param.domain`。其它类型往往另有「域名」列，**不能**给共享 `param.clb_vip_domain` 加全局拼接。同步要拼时单独做展示，格式按已拍板「换行」 |

禁止：`router.push` / `useVerify` / `BK_HCM_AJAX_URL_PREFIX`。

### 提交成功 / 失败（弹窗）

| 条件 | 前端 |
|------|------|
| `code === 0` 且 `data.task_management_id` 非空 | 关弹窗。Toast：「同步任务已创建，可在【任务管理-负载均衡】查看进度」。可点击段跳该 id 详情。回包只有该字段，不要读 total/create/update/delete |
| `code === 0` 且 id 为 `""` | 「没有可处理的负载均衡」，不跳转 |
| `2000001` | 用接口 `message`（未填地域等） |
| `2000002` | 「该账号地域同步任务进行中」 |
| 其它 `code != 0` | 用 `message`，不当成任务已创建 |

可点击 Toast：优先 `Message` 的 `message` 传 VNode（「【任务管理-负载均衡】」为 text 按钮）；若组件只吃 string，则用 `actions` 加「查看进度」。点击用 `routerAction.open` 新窗口。跳转前关弹窗。

入参与现网一致：`regions` 必填（单选包成数组即可）、可选 `cloud_ids` / `tag_filters`。协议 `regions` 最多 20；现网列表是单地域，不改选择器上限。

### 操作详情列（`operation=sync_load_balancer`）

**以协议「行字段 → 页面列」为唯一列清单**（不是 TAPD「6 字段」的自行裁剪）：

| 页面列 | 取值 | 说明 |
|--------|------|------|
| 开始时间 | `created_at` | 已有 |
| 结束时间 | `updated_at` | 已有：`init`/`running` 显示 `--` |
| 类别 | `param.op` | 新增：create/update/delete → 新增/修改/删除 |
| CLB ID | `param.cloud_lb_id` | 协议字段名；**不是**部分类型用的 `param.cloud_clb_id` |
| CLB VIP/域名 | `param.clb_vip_domain` + 可选 `param.domain` | 协议：VIP 单独、域名单独。有 domain 时换行两行（同步专用展示，不改共享列） |
| 任务状态 | `state` | 现网 `TASK_DETAIL_STATUS_NAME` |
| 失败原因 | `reason` | 任意状态：有值原样展示，空则 `--` |

不展示：`param.account_id`、`param.region`。`param` 无 `bk_biz_id`、无 `cloud_id`。

进度统计继续用现网 `task_details/state/count`，不要用提交回包。

## API 契约（摘自协议）

### 1. 提交条件同步

| 场景 | 方法 | 路径 |
|------|------|------|
| 资源 | POST | `/api/v1/cloud/vendors/{vendor}/accounts/{account_id}/resources/load_balancer/sync_by_cond` |
| 业务 | POST | `/api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/accounts/{account_id}/resources/load_balancer/sync_by_cond` |

现网 `syncResource` 已拼该路径。body：`regions`（必填，最多 20）、`cloud_ids`（可选，最多 20）、`tag_filters`（可选，最多 10）。

成功：`data.task_management_id` 字符串，无统计字段。

### 2. 任务列表

`POST /api/v1/cloud/bizs/{bk_biz_id}/task_managements/list`

筛「同步」：`operations` + `json_overlaps` + `["sync_load_balancer"]`（搜索模型已是该 op）。`state=running` 展示「执行中」（现网 `TASK_STATUS_NAME`，协议口语「运行中」，**不改全局列表文案**）。

### 3. 操作详情

`POST /api/v1/cloud/bizs/{bk_biz_id}/task_details/list`，按 `task_management_id` 滤。行 `param` 示例字段：`op` / `account_id` / `cloud_lb_id` / `clb_vip_domain` / `domain` / `region`。

明细 `state`：`init` / `running` / `success` / `failed` / `cancel`。

## 文件

`src/components/sync-account-resource/index.vue` `src/views/load-balancer/use-clb-sync-feedback.ts` `src/views/load-balancer/clb/load-balancer-table.vue` `src/views/load-balancer/listener/listener-table.vue` `src/views/resource/resource-manage/children/manage/load-balancer-manage.vue` `src/views/task/typings.ts` `src/views/task/constants.ts` `src/views/task/details/clb.vue` `src/views/task/details/children/action-list/fields.ts` `src/views/task/details/children/action-list/action-list.vue` `src/model/task/detail.view.ts`

## 已拍板

1. VIP/域名：协议字段各取各的；有 `param.domain` 才拼，样式**换行**。现网该列不是 `x/y` 格式，不能当复用依据
2. Toast 点击：**新窗口** `routerAction.open`
3. 详情列：按协议「行字段 → 页面列」；字段名以协议为准（`cloud_lb_id` / `clb_vip_domain` / `domain` / `op`），不按其它操作类型推测
4. 重执行：协议未给契约，**默认不支持**（待你若要支持再补 API）
5. 公共同步弹窗不承载具体资源的成功/失败文案；CLB 通过 `successHandler` / `errorHandler` 注入（与安全组 `errorHandler` 同一扩展点）
6. CLB 反馈 hook 放在 `views/load-balancer/`，不进全局 `src/hooks`（`/views` 独立模块）
