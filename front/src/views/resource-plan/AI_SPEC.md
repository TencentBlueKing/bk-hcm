# 资源预测模块 - Tab + Component 架构规范

## 1. 架构理念

资源预测模块采用 **Tab + Component** 的入口模式（参考 `views/task/index.vue`），将 CVM 预测（原有功能）和 GPU 需求（新增功能）统一在同一个入口组件下，通过 Tab 切换渲染不同子组件。

核心原则：
- **入口统一**：业务和服务请求视角各有一个入口组件（`entry-biz.vue` / `entry-srv.vue`），通过 `<component :is="...">` 渲染 Tab 内容
- **路由分离**：CVM 路由保留在 `router/module/business.ts` 和 `service.ts` 中（仅添加 `/cvm` 后缀），GPU 路由通过 `route-config.ts` 独立定义并导入
- **静态 Tab 标识**：使用 `MENU_*_RESOURCE_PLAN_CVM` / `MENU_*_RESOURCE_PLAN_GPU` 路由名称作为 Tab 类型标识（替代 `:type` 路由参数）
- **原有文件不变**：CVM 预测的所有组件、页面文件保持原位不动

## 2. 目录结构

```
views/
├── resource-plan/                    # 共享入口层（本目录）
│   ├── entry-biz.vue                 # 业务视角入口（Tab + Component）
│   ├── entry-srv.vue                 # 服务请求视角入口（Tab + Component）
│   ├── route-config.ts               # GPU 需求路由定义（参考 operation-log/route-config.ts）
│   ├── AI_SPEC.md                    # 本文档
│   └── gpu/                          # GPU 需求子模块
│       ├── types.ts                  # 仅存放视图相关类型（如 ISearchCondition）
│       ├── list/                     # 列表页
│       │   ├── index.vue
│       │   └── children/
│       │       ├── search/           # 搜索组件
│       │       └── data-list/        # 表格组件
│       └── detail/                   # 详情页
│           └── index.vue
├── business/resource-plan/           # CVM 预测 - 业务视角页面（保持不变）
│   ├── list/                         # 列表页
│   ├── add/                          # 新增页
│   ├── detail/                       # 详情页
│   └── mod/                          # 调整页
└── service/resource-plan/            # CVM 预测 - 服务请求视角页面（保持不变）
    └── resource-manage/
        ├── list/                     # 列表页
        ├── detail/                   # 详情页
        └── mod/                      # 调整页

store/
└── resource-plan/
    ├── index.ts                      # CVM 需求 store（已有）
    └── gpu-demand.ts                 # GPU 需求 store + API 数据类型/常量
```

## 3. 路由结构

### 3.1 业务视角（business.ts）

CVM 路由在 `business.ts` 中直接定义，仅在原路径基础上添加 `/cvm` 后缀：

```
/business/resource-plan/cvm              → entry-biz.vue（CVM Tab 激活）
/business/resource-plan/cvm/add          → CVM 新增页
/business/resource-plan/cvm/detail       → CVM 详情页
```

GPU 路由在 `route-config.ts` 中定义，通过 `...gpuDemandBizRouteConfig` 展开导入：

```
/business/resource-plan/gpu              → entry-biz.vue（GPU Tab 激活）
```

### 3.2 服务请求视角（service.ts）

CVM 路由在 `service.ts` 中直接定义（保持原有扁平结构），添加 `/cvm` 后缀：

```
/service/resource-plan/cvm               → redirect → /service/resource-plan/cvm/home
/service/resource-plan/cvm/home          → entry-srv.vue（CVM Tab 激活）
/service/resource-plan/cvm/detail        → CVM 详情页
/service/resource-plan/cvm/mod           → CVM 调整页
```

GPU 路由同样通过 `...gpuDemandSrvRouteConfig` 展开导入：

```
/service/resource-plan/gpu               → entry-srv.vue（GPU Tab 激活）
```

### 3.3 路由名称常量

| 常量 | 值 | 用途 |
|------|-----|------|
| `MENU_BUSINESS_RESOURCE_PLAN_CVM` | `menu_business_resource_plan_cvm` | 业务 CVM 列表页 |
| `MENU_BUSINESS_RESOURCE_PLAN_GPU` | `menu_business_resource_plan_gpu` | 业务 GPU 列表页 |
| `MENU_SERVICE_RESOURCE_PLAN_CVM` | `menu_service_resource_plan_cvm` | 服务 CVM 列表页 |
| `MENU_SERVICE_RESOURCE_PLAN_GPU` | `menu_service_resource_plan_gpu` | 服务 GPU 列表页 |

CVM 子页面（add/detail/mod）沿用原路由名称（如 `BizResourcePlanAdd`、`opResourcePlanDetail`）。

## 4. 入口组件规范

### 4.1 Tab 切换机制

入口组件通过 `route.name` 判断当前激活的 Tab（类似 `task/index.vue` 通过 `route.params.resourceType` 判断）：

```ts
const tabActive = computed({
  get() {
    return (route.name as string) || tabPanels[0].name;
  },
  set(value) {
    router.push({ name: value });
  },
});
```

切换 Tab 时，通过 `router.push({ name: value })` 导航到对应路由，触发 Vue Router 组件切换。由于 CVM 和 GPU 路由都指向同一入口组件，组件实例复用，仅 `route.name` 改变，computed 自动更新 activeTab。

### 4.2 组件渲染

通过 `<component :is="...">` 动态渲染 Tab 内容组件：

```ts
const tabComps: Record<string, any> = {
  [MENU_BUSINESS_RESOURCE_PLAN_CVM]: CvmPrediction,
  [MENU_BUSINESS_RESOURCE_PLAN_GPU]: GpuDemand,
};
```

CVM 列表组件直接静态导入（如 `import CvmPrediction from '@/views/business/resource-plan/list'`），GPU 占位组件从当前目录导入。

## 5. 路由配置模式

### 5.1 CVM 路由（旧模式，保留在 router/module 中）

CVM 路由保持在 `business.ts` / `service.ts` 中直接定义，遵循项目原有惯例，仅添加 `/cvm` 路径前缀：

```diff
- path: '/business/resource-plan',
+ path: '/business/resource-plan/cvm',
```

### 5.2 GPU 路由（新模式，模块自身的 route-config.ts）

GPU 路由定义在 `views/resource-plan/route-config.ts` 中，参考 `views/operation-log/route-config.ts`：
- 分别导出 `gpuDemandBiz` 和 `gpuDemandSrv`
- 在 `business.ts` / `service.ts` 中通过别名导入并展开
- 路由使用相对路径（如 `resource-plan/gpu`），由父级 `/business` 或 `/service` 解析

### 5.3 为什么混用两种模式

保留老的方式（CVM 路由在 router/module 中），同时新模块（GPU）采用新的模式（route-config 定义在模块目录中），这样：
- 不影响现有 CVM 路由的代码结构，减少风险
- 新模块遵循更好的代码组织实践（路由定义与视图代码同一目录）
- 渐进式迁移，未来可以逐步将老模块的路由也迁移到新模式

## 6. 类型与常量的分层规范

### 6.1 核心原则

**views 目录只放视图相关的类型/常量，API 数据相关的全部定义到 store 文件中。**

| 类别 | 存放位置 | 示例 |
|------|---------|------|
| API 响应数据结构 | `store/resource-plan/gpu-demand.ts` | `IGpuDemandItem` |
| API 请求参数类型 | `store/resource-plan/gpu-demand.ts` | `IGpuDemandListParams` |
| API 枚举值/状态常量 | `store/resource-plan/gpu-demand.ts` | `GPU_DEMAND_STATUS`、`GpuDemandStatus`、`GPU_DEMAND_STATUS_MAP` |
| 视图组件的 props/form 类型 | `views/resource-plan/gpu/types.ts` | `ISearchCondition` |
| @Model/@Column 装饰器类 | `views/.../children/` 目录 | `SearchCondition`、`TableColumn` |

### 6.2 为什么这样分

- Store 是数据层的 single source of truth，API 返回什么字段、状态枚举有哪些值，这些与后端接口强绑定，理应和 API 调用放在一起
- Views 中的 `@Model`/`@Column` 装饰器类虽然引用了 store 中的类型/常量，但它们定义的是"如何在 UI 上展示"，属于视图层关注点
- 视图特有的类型（如搜索表单的值结构 `ISearchCondition`）仅服务于 UI 交互，不应出现在 store 中

## 7. API 接入模式

### 7.1 Store 结构

GPU 需求的 store 文件 `store/resource-plan/gpu-demand.ts` 包含：
- API 数据相关的类型和常量导出
- Pinia store（`useGpuDemandStore`）提供 API 调用方法和加载状态

### 7.2 API 路径自动区分视角

使用 `useWhereAmI` 的 `getBusinessApiPath()` 自动拼接路径前缀：

```ts
const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/orders/list`;
// 业务视角 → /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/orders/list
// 服务请求视角 → /api/v1/woa/plans/resources/gpu/demands/orders/list
```

### 7.3 并发获取数据和总数

后端接口的 `page.count` 参数决定返回数据详情还是总数，两者互斥。使用 `enableCount` + `Promise.all` 并发请求：

```ts
const [listRes, countRes] = await Promise.all([
  http.post(api, enableCount(params, false)),  // 返回 details
  http.post(api, enableCount(params, true)),   // 返回 count
]);
```

### 7.4 类型复用

请求参数的 `filter` 和 `page` 直接复用项目已有类型（`QueryFilterType`、`IPageQuery`），不要自定义重复结构。

## 8. 列表页交互模式

### 8.1 表格勾选（useTableSelection）

使用项目提供的 `useTableSelection` hook 管理表格行选择，而非自行处理 `selection-change` 事件。

```ts
import useTableSelection from '@/hooks/use-table-selection';

const isRowSelectEnable = ({ row }: { row: IGpuDemandItem }) => {
  return row.status === GPU_DEMAND_STATUS.INIT;
};

const { selections, handleSelectAll, handleSelectChange } = useTableSelection({
  isRowSelectable: isRowSelectEnable,
});

watch(selections, (val) => emit('select', val), { deep: true });
```

表格需同时绑定三个属性/事件：
- `:is-row-select-enable` — 控制哪些行的 checkbox 可点击
- `@select-all` → `handleSelectAll`
- `@selection-change` → `handleSelectChange`

子组件通过 `emit('select', selections)` 将选中行传递给父级，父级用 `selectedRows` 管理状态。

### 8.2 Toolbar + 表格面板

Toolbar 和表格用 `.table-panel` 包裹，提供白色背景、圆角和阴影，使之成为独立的视觉区块：

```vue
<div class="table-panel">
  <div class="toolbar">
    <bk-button v-if="isBusinessPage" theme="primary" @click="handleCreate">
      <Plus style="font-size: 22px" />
      新增需求
    </bk-button>
    <bk-button v-if="isServicePage" theme="primary" :disabled="!selectedRows.length" @click="handleBatchPending">
      批量更新状态
    </bk-button>
  </div>
  <data-list ... />
</div>
```

```scss
.table-panel {
  background: #fff;
  border-radius: 2px;
  box-shadow: 0 2px 4px 0 #1919290d;
  padding: 16px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
```

规范：
- 搜索区域在 `.table-panel` 外部，toolbar + 表格在内部
- 批量操作按钮在无勾选项时 `disabled`
- 图标组件从 `bkui-vue/lib/icon` 导入（如 `Plus`）
- 操作完成后调用 `fetchList()` 刷新列表

### 8.3 表格列配置

通过 `@Column` 装饰器的 `minWidth` 和 `fixed` 属性控制列宽和固定行为：

```ts
@Column('string', { name: '需求ID', fixed: 'left', minWidth: 120, index: 0 })
id: string;
```

- 每列都应设置 `minWidth`，避免内容挤压
- 关键标识列（如需求ID）设置 `fixed: 'left'`
- selection 列使用 `min-width="30"` + `fixed="left"`
- `data-list.vue` 模板中绑定 `:min-width="column.minWidth"` 和 `:fixed="column.fixed"`
- 仅服务请求视角独有的列（运营产品、业务）通过 `SERVICE_ONLY_COLUMNS` 在父级过滤

### 8.4 操作列按钮状态

操作列按钮的启用/禁用基于行的 `status` 字段：

| 操作 | 视角 | 可操作状态 | 说明 |
|------|------|-----------|------|
| 调整 | 业务 | 所有状态 | 进入详情/编辑页 |
| 终止 | 业务 | INIT、REJECT_ALL | 终止整单 |
| 评审 | 服务请求 | 非 INIT | 进入详情逐行评审（INIT 时需先批量更新状态） |
| 驳回 | 服务请求 | PENDING | 整单驳回为全部已驳回 |
| 终止 | 服务请求 | PENDING | 整单终止 |

### 8.5 事件驱动的操作流

表格子组件（`data-list.vue`）只负责 UI 渲染和按钮状态判断，所有业务操作通过 `emit` 事件向外抛出，由父级页面（`list/index.vue`）统一处理：

```vue
<!-- data-list.vue：只 emit 事件 -->
<bk-button @click="emit('view-details', row)">评审</bk-button>
<bk-button @click="emit('reject', row)">驳回</bk-button>
<bk-button @click="emit('terminate', row)">终止</bk-button>

<!-- list/index.vue：统一处理 -->
<data-list
  @view-details="handleViewDetails"
  @reject="handleReject"
  @terminate="handleTerminate"
  @select="handleSelect"
/>
```

原则：
- **子组件不调用 store / API**，不持有业务逻辑，只做展示和事件转发
- **父级集中管理**所有操作：路由跳转、API 调用、列表刷新、错误处理
- 操作完成后统一调用 `fetchList()` 刷新数据
- 勾选状态同样通过 `emit('select')` 上报，父级用 `selectedRows` 维护，供 toolbar 批量按钮使用

这样做的好处：
- 子组件可复用，不耦合具体的 store 或路由
- 操作流集中在一处，便于添加二次确认弹窗、loading 状态、错误提示等统一处理
- 测试时可独立验证子组件的 UI 逻辑和父级的业务逻辑

## 9. 新增功能模块开发指引

### 9.1 新增更多 Tab

如需新增 Tab（如 NPU 需求等）：

1. 在 `menu-symbol.ts` 中定义新的 `MENU_*_RESOURCE_PLAN_XXX` 常量
2. 在 `route-config.ts` 中添加新的路由定义
3. 在 entry 组件的 `tabPanels` 数组中添加新 Tab
4. 在 `tabComps` 中添加新组件映射
5. 在 `store/resource-plan/` 中新建对应 store 文件，包含 API 数据类型/常量和 API 调用

## 10. 权限相关

- 服务请求视角：`checkAuth: 'ziyan_resource_plan_manage'` 设置在 CVM home 路由和 GPU 路由上
- 业务视角：无路由级权限检查，组件内通过 `biz_resource_plan_operate` 校验
- GPU 独立权限待需求确认后添加
