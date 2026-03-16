# 资源预测（resource-plan）模块代码结构分析

> 分析日期：2026-03-14

## 1. 模块概述

"资源预测"模块是 HCM 平台中管理资源预测需求单据的功能模块，支持业务侧提交/调整预测需求，以及管理员侧查看/审批需求。模块涉及三个主要视角：

- **业务资源视角**（`/business/resource-plan`）：业务用户提交、查看、调整预测需求
- **资源运营视角**（`/service/resource-plan`）：管理员查看、审批预测需求
- **单据管理视角**（`/business/ticket` 和 `/service/ticket`）：资源预测单据的审批流管理

---

## 2. 文件全景

### 2.1 核心组件目录

```
front/src/components/resource-plan/
├── constants.ts                          # 需求状态常量（可申领/未到申领时间/已过期/额度用尽/变更中）
├── add/                                  # 新增预测需求
│   ├── index.tsx                         # 新增页面主入口
│   ├── index.module.scss
│   ├── basic/index.tsx                   # 基础信息表单
│   ├── basic/index.module.scss
│   ├── cvm/index.tsx                     # CVM（虚拟机）资源配置
│   ├── cvm/index.module.scss
│   ├── cbs/index.tsx                     # CBS（云盘）资源配置
│   ├── cbs/index.module.scss
│   └── type/index.tsx                    # 预测类型选择
├── applications/                         # 申请单详情
│   └── detail/
│       ├── approval/index.tsx            # 审批流展示
│       ├── approval/index.module.scss
│       ├── basic/index.vue               # 基本信息
│       ├── basic/index.tsx
│       ├── basic/index.module.scss
│       ├── header/index.tsx              # 详情头部（含权限校验 type: 'resource_plan'）
│       ├── header/index.module.scss
│       ├── list/index.vue                # 需求明细列表
│       ├── list/changed-text.vue         # 变更字段高亮
│       └── ticket-audit/index.vue        # 审批记录
└── resource-manage/                      # 资源管理（管理员视角）
    ├── mod/                              # 调整/修改需求
    │   ├── index.tsx
    │   ├── index.scss
    │   └── useModColumn.tsx              # 调整列定义
    ├── list/                             # 列表
    │   ├── table/index.tsx               # 表格主体（含权限校验 biz_resource_plan_operate）
    │   ├── table/index.module.scss
    │   ├── table/components/
    │   │   ├── batch-cancellation-dialog/ # 批量取消对话框
    │   │   └── batch-postpone-sideslider/ # 批量延期侧滑面板
    │   ├── search/index.tsx              # 搜索条件
    │   └── search/index.module.scss
    └── detail/                           # 详情
        ├── list/index.tsx                # 详情需求列表
        ├── basic/index.tsx               # 详情基本信息
        └── basic/index.module.scss
```

### 2.2 业务视角页面（views/business）

```
front/src/views/business/resource-plan/
├── list/                                 # 业务预测列表页
│   ├── index.tsx
│   └── index.module.scss
├── add/                                  # 业务新增预测
│   ├── index.tsx                         # 新增页面入口
│   ├── index.module.scss
│   ├── basic/index.tsx                   # 基础信息
│   ├── basic/index.module.scss
│   ├── header/index.tsx                  # 页头
│   ├── header/index.module.scss
│   ├── list/index.tsx                    # 需求列表编辑
│   ├── list/index.module.scss
│   ├── button/index.tsx                  # 操作按钮
│   ├── button/index.module.scss
│   ├── memo/index.tsx                    # 备注
│   ├── memo/index.module.scss
│   └── plan-remark.js                    # 预测说明
├── detail/                               # 业务预测详情
│   ├── index.tsx
│   └── index.module.scss
├── mod/index.tsx                         # 业务调整预测
└── children/
    └── obs-project-selector.vue          # OBS 项目选择器
```

### 2.3 运营管理视角页面（views/service）

```
front/src/views/service/resource-plan/
└── resource-manage/
    ├── list/                             # 运营预测列表
    │   ├── index.tsx
    │   └── index.module.scss
    ├── detail/                           # 运营预测详情
    │   ├── index.tsx
    │   └── index.module.scss
    └── mod/index.tsx                     # 运营调整预测
```

### 2.4 单据管理视角（views/ticket）

```
front/src/views/ticket/
├── route-config.ts                       # 资源预测单据路由（含服务请求和业务两侧）
├── constants.ts                          # 单据类型常量（含 resource_plan tab 定义）
├── entry-srv.vue                         # 服务请求入口（含"资源预测" tab）
├── entry-biz.vue                         # 业务资源入口（含"资源预测" tab）
└── children/resource-plan/
    ├── typings.ts                        # 类型定义
    ├── list/
    │   ├── list-srv.vue                  # 服务请求下的资源预测单据列表
    │   ├── list-biz.vue                  # 业务资源下的资源预测单据列表
    │   └── children/
    │       ├── search/
    │       │   ├── search.vue            # 搜索组件
    │       │   └── condition.ts          # 搜索条件定义
    │       └── data-list/
    │           ├── data-list.vue         # 数据列表组件
    │           └── column.ts             # 列定义
    ├── detail/index.vue                  # 资源预测单据详情
    └── sub-ticket/                       # 子单据
        ├── sub-ticket-list.vue           # 子单据列表
        ├── sub-ticket-detail.vue         # 子单据详情
        └── components/
            ├── status-text.vue           # 状态文本展示
            └── stage.vue                 # 审批阶段展示
```

### 2.5 关联页面

```
front/src/views/business/host-inventory/
├── index.tsx                             # 主机库存（引用 resourcePlan store）
└── resource-demands-result.vue           # 资源需求结果（直接调用预测需求 API）
```

---

## 3. 路由配置

### 3.1 业务资源路由（`router/module/business.ts`）

| 路由路径 | 路由名称 | 组件 | 说明 |
|---------|---------|------|------|
| `/business/resource-plan` | `BizResourcePlan` | - | 菜单入口，title: "资源预测" |
| `/business/resource-plan`（默认子路由） | `bizResourcePlanList` | `views/business/resource-plan/list` | 列表页 |
| `/business/resource-plan/add` | `BizResourcePlanAdd` | `views/business/resource-plan/add` | 新增页 |
| `/business/resource-plan/detail` | `BizResourcePlanDetail` | `views/business/resource-plan/detail` | 详情页 |
| `/business/service/resource-plan-mod` | `bizModPlanList` | `views/business/resource-plan/mod` | 调整页（notMenu） |
| `/business/ticket/resource-plan/detail` | `menu_business_ticket_resource_plan_details` | `views/ticket/children/resource-plan/detail` | 单据详情 |

### 3.2 资源运营路由（`router/module/service.ts`）

| 路由路径 | 路由名称 | 组件 | 说明 |
|---------|---------|------|------|
| `/service/resource-plan` | `opResourcePlan-redirect` | 重定向到 `/business/resource-plan` | 兼容旧宣传链接 |
| `/service/resource-plan/home` | `opResourcePlan` | `views/service/resource-plan/resource-manage/list` | 管理员列表页，需 `ziyan_resource_plan_manage` 权限 |
| `/service/resource-plan/detail` | `opResourcePlanDetail` | `views/service/resource-plan/resource-manage/detail` | 管理员详情页 |
| `/service/resource-plan/mod` | `modPlanList` | `views/service/resource-plan/resource-manage/mod` | 管理员调整页 |
| `/service/my-apply/resource-plan/detail` | - | 重定向到 `/service/ticket/resource-plan/detail` | 兼容旧路由 |
| `/service/ticket/resource-plan/detail` | `menu_service_ticket_resource_plan_details` | `views/ticket/children/resource-plan/detail` | 单据详情 |

### 3.3 独立路由（`router/module/resource-plan.ts`）

当前为空数组（内容已注释），历史上曾有独立的 `/resource-plan/manage` 路由。

### 3.4 Header Tab 切换（`views/home/hooks/useChangeHeaderTab.ts`）

`case 'resource-plan'` 分支加载 `resourcePlan` 菜单配置。

---

## 4. Store（状态管理）

### 4.1 `store/resourcePlan.ts` — 主 Store（Options API 风格）

- **Store ID**: `resourcePlanStore`
- **导出**: `useResourcePlanStore`
- **API 方法**:

| 方法 | API 路径 | 说明 |
|------|---------|------|
| `getDiskTypes` | `GET /api/v1/woa/meta/disk_type/list` | 磁盘类型列表 |
| `getObsProjects` | `GET /api/v1/woa/meta/obs_project/list` | OBS 项目列表 |
| `getBizResourcesTicketsList` | `POST /api/v1/woa/bizs/{bizId}/plans/resources/tickets/list` | 业务预测单据列表 |
| `getOpResourcesTicketsList` | `POST /api/v1/woa/plans/resources/tickets/list` | 管理员预测单据列表 |
| `getBizResourcesTicketsById` | `GET /api/v1/woa/bizs/{bizId}/plans/resources/tickets/{id}` | 业务单据详情 |
| `getOpResourcesTicketsById` | `GET /api/v1/woa/plans/resources/tickets/{id}` | 管理员单据详情 |
| `getBizResourcesTicketsAuditById` | `GET /api/v1/woa/bizs/{bizId}/plans/resources/tickets/{id}/audit` | 业务审批流 |
| `getOpResourcesTicketsAuditById` | `GET /api/v1/woa/plans/resources/tickets/{id}/audit` | 管理员审批流 |
| `createBizPlan` | `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/tickets/create` | 创建预测单据 |
| `getBizOrgRelation` | `GET /api/v1/woa/bizs/{bizId}/org/relation` | 业务组织关系 |
| `getDemandClasses` | `GET /api/v1/woa/plan/demand_class/list` | 预测类型列表 |
| `getRegions` | `GET /api/v1/woa/meta/region/list` | 城市列表 |
| `getZones` | `POST /api/v1/woa/meta/zone/list` | 可用区列表 |
| `getSources` | `GET /api/v1/woa/plan/demand_source/list` | 变更来源列表 |
| `getDeviceClasses` | `GET /api/v1/woa/meta/device_class/list` | 机型规格列表 |
| `getDeviceTypes` | `POST /api/v1/woa/meta/device_type/list` | 机型类型列表 |
| `getOpProductsList` | `POST /api/v1/woa/metas/op_products/list` | 运营产品列表 |
| `getPlanProductsList` | `POST /api/v1/woa/metas/plan_products/list` | 规划产品列表 |
| `getBizsByOpProductList` | `POST /api/v1/woa/metas/bizs/by/op_product/list` | 按运营产品查业务 |
| `getPlanTypes` | `POST /api/v1/woa/metas/plan_types/list` | 预测类型列表 |
| `getResourcesDemandsList` | `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/demands/list` | 业务需求列表 |
| `getPlanDemand` | `GET /api/v1/woa/bizs/{bk_biz_id}/plans/demands/{demand_id}` | 业务需求详情 |
| `getListChangeLogs` | `POST /api/v1/woa/bizs/{bk_biz_id}/plans/demands/change_logs/list` | 变更历史 |
| `getResourcesDemandsListByOrg` | `POST /api/v1/woa/plans/resources/demands/list` | 管理员需求列表 |
| `getPlanDemandByOrg` | `GET /api/v1/woa/plans/demands/{demand_id}` | 管理员需求详情 |
| `getListChangeLogsByOrg` | `POST /api/v1/woa/plans/demands/change_logs/list` | 管理员变更历史 |
| `cancelResourcesDemands` | `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/demands/cancel` | 批量取消需求 |

### 4.2 `store/resource-plan/index.ts` — 需求列表 Store（Composition API 风格）

- **Store ID**: `resource-plan`
- **导出**: `useResourcePlanStore`（注意与 4.1 同名导出，使用时需注意导入路径）
- **功能**: 提供 `getDemandList`、`getOpProductList`、`getPlanProductList`
- **类型定义**: 包含 `ResourcesDemandStatus`、`IResourcesDemandItem`、`IListResourcesDemandsParams` 等

### 4.3 `store/ticket/resource-plan.ts` — 单据 Store

- **Store ID**: `ticket/resource-plan`
- **导出**: `useResourcePlanTicketStore`
- **功能**: 提供 `getTicketList`、`getTicketStatusList`、`getTicketTypeList`

### 4.4 `store/ticket/res-sub-ticket.ts` — 子单据 Store

- **Store ID**: `resSubTicketStore`
- **导出**: `useResSubTicketStore`
- **功能**: 子单列表、审批流、子单详情、重试、部门审批、额度查询、终止单据

### 4.5 `store/usePlanStore.ts` — 预测调整 Store

- **Store ID**: `planStore`
- **功能**: 需求查询、需求校验、批量调整、交付时间范围查询、数据转换工具函数

### 4.6 `store/index.ts` — Store 统一导出

- 第 18 行：`export * from './resourcePlan'`

---

## 5. 类型定义

### 5.1 `typings/resourcePlan.ts`

主要类型定义（约 494 行）：
- `IBizResourcesTicketsParam` / `IOpResourcesTicketsParam` — 查询参数
- `IResourcesTicketItem` — 单据列表项
- `TicketStatus` — 单据状态枚举（init/auditing/rejected/partial_rejected/done/failed/partial_failed/revoked/terminated）
- `TicketByIdResult` / `TicketBaseInfo` / `TicketDemands` — 单据详情
- `IPlanTicketAudit` / `IPlanTicketItsmAudit` / `IPlanTicketCrpAudit` — 审批流类型
- `IPlanTicket` / `IPlanTicketDemand` — 创建单据参数
- `IListResourcesDemandsParam` / `IListResourcesDemandsResult` — 需求列表
- `ResourcesDemandsStatus` — 需求状态枚举（can_apply/not_ready/expired/spent_all/locked）
- `IPlanDemandResult` / `IListChangeLogsResult` — 需求详情和变更历史
- `IOpProductsResult` / `IPlanProductsResult` / `IBizsByOpProductResult` — 运营/规划产品
- `ResourceDemandResultStatus` — 预测结果状态

### 5.2 `typings/plan.ts`

- `IDemandListDetail` — 需求明细完整字段（约 43 个字段）
- `AdjustType` — 调整类型枚举（update/delay/none）
- `IAdjust` / `IAdjustParams` / `AdjustInfo` — 调整相关
- `IVerifyResourceDemandParams` / `IVerifyResourceDemandData` — 需求校验
- `IDemandSpec` / `IDemandSuborder` — 需求规格
- `IExceptTimeRange` / `ITimeRange` — 时间范围
- `ChargeType` / `ChargeTypeMap` — 计费模式

### 5.3 `store/resource-plan/index.ts` 中的类型

- `ResourcesDemandStatus` — 需求状态（与 typings 中重复定义）
- `IResourcesDemandOverview` — 需求概览统计
- `IResourcesDemandItem` — 需求列表项
- `IListResourcesDemandsParams` — 查询参数

### 5.4 `store/ticket/resource-plan.ts` 中的类型

- `IResourcePlanTicketItem` — 单据列表项
- `IResourcePlanTicketStatusItem` / `IResourcePlanTicketTypeItem` — 状态和类型

### 5.5 `store/ticket/res-sub-ticket.ts` 中的类型

- `SubTicketItem` / `SubTicketParam` — 子单据
- `SubTicketAudit` / `AdminAudit` — 审批流
- `SubTicketDetail` / `SubTicketDemand` — 子单详情
- `TransferQuotas` / `TransferQuotasConfigs` — 额度配置
- `STATUS_ENUM` / `STAGE_ENUM` — 状态和阶段枚举

---

## 6. 权限配置

### 6.1 权限定义（`store/common.ts`，第 82-86 行）

| 权限 type | action | 权限 ID | 说明 |
|-----------|--------|---------|------|
| `ziyan_resource_plan` | `access` | `ziyan_resource_plan_manage` | 服务请求-管理员查看及操作 |
| `resource_plan` | `create` | `biz_resource_plan_operate` | 业务-资源预测-新增（bk_biz_id: 0） |
| `resource_plan` | `update` | `biz_resource_plan_operate` | 业务-资源预测-修改（bk_biz_id: 0） |
| `resource_plan` | `delete` | `biz_resource_plan_operate` | 业务-资源预测-删除（bk_biz_id: 0） |

### 6.2 权限使用位置

- `router/module/service.ts:67` — `checkAuth: 'ziyan_resource_plan_manage'`（运营列表页路由守卫）
- `components/resource-plan/resource-manage/list/table/index.tsx` — 多处 `biz_resource_plan_operate` 校验
- `components/resource-plan/applications/detail/header/index.tsx:29` — `type: 'resource_plan'`
- `views/ticket/children/resource-plan/list/list-srv.vue:66` — `type: 'resource_plan'`
- `views/ticket/children/resource-plan/list/list-biz.vue:76` — `type: 'resource_plan'`

### 6.3 AuthActionType（`common/auth-service.ts`，第 9-18 行）

资源预测使用的 action 类型：`create`、`update`、`delete`、`access`

---

## 7. 常量定义

### 7.1 `components/resource-plan/constants.ts`

- `RESOURCE_DEMANDS_STATUS_NAME` — 需求状态中文名称映射
- `RESOURCE_DEMANDS_STATUS_CLASSES` — 需求状态样式类映射

### 7.2 `common/constant.ts`（第 640-676 行）

- `RESOURCE_PLAN_STATUSES_type` — 单据状态枚举值（init/auditing/rejected/done/failed/revoked）
- `RESOURCE_PLAN_STATUSES_MAP` — 单据状态图标和颜色映射

### 7.3 `constants/menu-symbol.ts`（第 91-92 行）

- `MENU_SERVICE_TICKET_RESOURCE_PLAN_DETAILS` — 服务请求下单据详情菜单标识
- `MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS` — 业务资源下单据详情菜单标识

### 7.4 `views/ticket/constants.ts`

- `APPLY_TYPES` — 单据类型 tab 列表中包含 `resource_plan` 项

### 7.5 `store/ticket/res-sub-ticket.ts`

- `STATUS_ENUM` — 子单据状态（待审批/审批中/审批拒绝/失败/成功/已失效）
- `STAGE_ENUM` — 审批阶段（部门审批/公司审批）

---

## 8. 菜单配置

| 位置 | 菜单名称 | 图标 |
|------|---------|------|
| 业务资源侧边菜单 | 资源预测 | `hcm-icon bkhcm-icon-resource-plan` |
| 资源运营侧边菜单 | 资源预测 | `hcm-icon bkhcm-icon-resource-plan` |
| 单据管理 Tab（服务请求） | 资源预测 | - |
| 单据管理 Tab（业务资源） | 资源预测 | - |

---

## 9. 文件清单汇总（按类型）

### 组件文件（33 个）
所有文件位于 `front/src/components/resource-plan/` 下（见 2.1 节）

### 页面文件（36 个）
- `views/business/resource-plan/` — 19 个文件
- `views/service/resource-plan/` — 5 个文件
- `views/ticket/children/resource-plan/` — 12 个文件

### Store 文件（5 个）
- `store/resourcePlan.ts`
- `store/resource-plan/index.ts`
- `store/ticket/resource-plan.ts`
- `store/ticket/res-sub-ticket.ts`
- `store/usePlanStore.ts`

### 类型文件（2 个）
- `typings/resourcePlan.ts`
- `typings/plan.ts`

### 路由文件（4 个）
- `router/module/resource-plan.ts`
- `router/module/business.ts`（第 354-387, 557-565 行）
- `router/module/service.ts`（第 17-87 行）
- `views/ticket/route-config.ts`（第 51-62, 101-112 行）

### 常量/配置文件（5 个）
- `components/resource-plan/constants.ts`
- `common/constant.ts`（第 640-676 行）
- `constants/menu-symbol.ts`（第 91-92 行）
- `views/ticket/constants.ts`（第 91-95 行）
- `router/header-config.ts`（第 42-44 行，已注释）

### 权限文件（1 个）
- `store/common.ts`（第 82-86 行）

### 入口文件（2 个）
- `views/ticket/entry-srv.vue`（第 62-63 行）
- `views/ticket/entry-biz.vue`（第 88-89 行）

### Hook 文件（1 个）
- `views/home/hooks/useChangeHeaderTab.ts`（第 5, 53-54 行）

### 关联页面（2 个）
- `views/business/host-inventory/index.tsx`
- `views/business/host-inventory/resource-demands-result.vue`

### Store 导出（1 个）
- `store/index.ts`（第 18 行）

---

## 10. 架构特点与注意事项

1. **Store 命名冲突**：`store/resourcePlan.ts` 和 `store/resource-plan/index.ts` 都导出 `useResourcePlanStore`，使用时需注意导入路径区分。

2. **双视角设计**：几乎所有 API 都分为 `bizs/{bizId}/...` （业务视角）和无 bizId 前缀（管理员视角）两组，通过 `useWhereAmI` hook 或 `resolveBizApiPath` 工具函数动态拼接路径。

3. **类型定义分散**：`ResourcesDemandStatus` 枚举在 `typings/resourcePlan.ts` 和 `store/resource-plan/index.ts` 中重复定义。

4. **路由兼容**：多处 redirect 处理旧路由兼容，如 `/service/resource-plan` → `/business/resource-plan`，`/service/my-apply/resource-plan/detail` → `/service/ticket/resource-plan/detail`。

5. **API 前缀**：统一使用 `/api/v1/woa/` 前缀，符合项目 API 规范中 woa-server 入口的定义。
