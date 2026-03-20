## Context

资源计划模块已有完整的 GPU 需求提报主单（`res_plan_demand_gpu_order`）和子单（`res_plan_demand_gpu_suborder`）数据服务层。主单表中不存储聚合字段 `total_gpu_num` / `total_qpm_max`，这两个值需要从子单表 SUM 汇总。

现有 data-service 提供：
- `ListResPlanDemandGpuOrder`：标准列表查询
- `ListResPlanDemandGpuSubOrder`：子单列表查询

需要在 woa-server 新增两个对外接口（资源视角 + 业务视角），返回含聚合字段的主单列表。

## Goals / Non-Goals

**Goals:**
- 在 woa-server 暴露资源视角和业务视角两个列表查询接口
- 响应中包含 `total_gpu_num`、`total_qpm_max` 聚合字段
- 业务视角自动过滤 bk_biz_id，防止越权访问
- 支持标准 filter + page（含 count 模式）

**Non-Goals:**
- 不修改 data-service、DAO、数据库表结构
- 不按子单状态过滤（全部子单参与 SUM）
- 不支持按 total_gpu_num / total_qpm_max 排序

## Decisions

### 决策1：两阶段查询 vs JOIN 查询

**选择**：两阶段查询（woa-server 内存聚合）

**理由**：
- 不需要修改 data-service 和 DAO，改动范围极小
- 与现有代码模式一致（`batchUpdateSubOrderStatuses` 同样是两阶段）
- 主单 page 最大 500 条，子单数量可控，内存聚合开销可接受
- JOIN 方案需新增 data-service 接口类型、proto、client，成本更高

**两阶段流程**：
```
Step 1: ListResPlanDemandGpuOrder(filter, page)  → 主单列表 + orderIDs
Step 2: 循环翻页 ListResPlanDemandGpuSubOrder(order_id IN orderIDs) → 子单明细
Step 3: 内存 map[orderID]{gpuNum, qpmMax} SUM 聚合
Step 4: 拼装响应
```

### 决策2：count 模式短路

当 `page.count=true` 时，直接调用主单 count 接口返回，不查子单。聚合字段在 count 响应中无意义。

### 决策3：bk_biz_id 注入方式

业务视角接口从路径参数提取 `bk_biz_id`，通过 IAM `ListAuthorizedBiz` 校验权限后，将其作为额外 rule 注入 filter（`And` 组合），复用与 `BatchTerminateBizResPlanDemandGpuOrder` 相同的鉴权模式。

### 决策4：函数拆分

核心逻辑拆分为 3 个私有函数，保持每个函数不超过 80 行：
- `listResPlanDemandGpuOrder(kt, filter, page)` — 入口，协调两阶段
- `fetchSubOrderStats(kt, orderIDs)` — 查子单并返回聚合 map
- `assembleGpuOrderItems(orders, statsMap)` — 组装最终响应

## Risks / Trade-offs

- **子单量大时 N+1 开销** → 通过 `order_id IN (orderIDs)` 单次批量查询子单（循环翻页），不是逐条查询；每次 page.limit=500，整体 RPC 次数可控
- **filter 注入的边界** → 业务视角 filter 注入后若原始 filter 为空（op+rules 均空），需要确保 filter 合法；通过构造 `And(original, eq(bk_biz_id, x))` 保证结构合法
