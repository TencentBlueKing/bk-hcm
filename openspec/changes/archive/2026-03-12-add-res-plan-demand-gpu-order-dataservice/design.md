## Context

`res_plan_demand_gpu_order` 是 GPU 需求提报场景下的主单表，当前仅有建表 SQL，尚无对应的 data-service 层接口。项目中已有同类资源（如 `res_plan_demand`）的完整实现，本次工作参照该模式进行。

## Goals / Non-Goals

**Goals:**
- 新增 `res_plan_demand_gpu_order` 表对应的 Go table 结构体、DAO 层、API 协议层和 data-service service 层
- 提供批量创建（BatchCreate）、批量更新（BatchUpdate）、分页查询（List）、批量删除（Delete）四个接口
- 路由注册到 data-service 的 resource-plan 分组下

**Non-Goals:**
- 不涉及 cloud-server / woa-server 等上层服务的入口接口
- 不涉及前端页面开发
- 不涉及其他表的改动

## Decisions

### 1. 目录结构完全复用 res-plan-demand 模式

**决策**：`res_plan_demand_gpu_order` 的四层实现（table / dao / api / service）与 `res_plan_demand` 保持一致的目录结构和代码分层。

**理由**：项目已有成熟的 resource-plan 分层模式，保持一致性便于后续维护，减少额外学习成本。

### 2. 批量更新采用逐条 filter 模式

**决策**：批量更新接口循环对每条记录执行 `UpdateWithTx`，以 `id = ?` 为 filter 条件。

**理由**：不同记录需要更新不同字段值，无法用单一 SQL 表达，与项目已有模式一致。

### 3. status 字段定义为枚举类型

**决策**：为 `status` 字段在 `pkg/criteria/enumor/` 中定义枚举类型 `ResPlanDemandGpuOrderStatus`，并实现 `Validate()` 方法。

**理由**：项目规范要求可枚举字段必须使用枚举类型，并需要校验方法。

### 4. op_product_name 字段类型修正

**决策**：建表 SQL 中 `op_product_name` 误写为 `bigint`，应为 `varchar(64)`，在 Go 结构体和业务逻辑中按字符串处理。

**理由**：从业务语义和字段命名（name 后缀）判断，该字段应为字符串类型，长度与同类字段（如 `op_product_name` 在其他表中的定义）一致，取 64。

## Risks / Trade-offs

- [风险] op_product_name 字段 SQL 定义为 bigint，与业务语义不符 → 缓解：Go 代码按 string 处理，实际建表需修正为 varchar(64)；任务中明确说明需修正该字段类型
- [风险] 新增 DAO 注册到 `dao.go` 时需同时更新接口定义和实现，遗漏会导致编译失败 → 缓解：tasks 中列为独立步骤，编译检查可发现
