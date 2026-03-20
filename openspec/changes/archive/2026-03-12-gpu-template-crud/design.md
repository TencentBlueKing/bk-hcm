## Context

GPU需求提报功能需要模版配置表 `res_plan_demand_gpu_template` 的完整数据访问层支持。该表存储Excel模版的结构定义（JSON格式），属于资源规划（resource-plan）域。项目已有成熟的resource-plan CRUD模式（如 res_plan_week、res_plan_demand 等），本次变更严格遵循现有模式。

表结构：
- `id` VARCHAR(64) - 模版ID（主键，由IDGen生成）
- `schema` JSON - 模版内容（一个Excel对应一条记录）
- `remark` VARCHAR(255) - 备注（可选）
- `creator` VARCHAR(64) - 创建人
- `reviser` VARCHAR(64) - 修改人
- `created_at` DATETIME - 创建时间
- `updated_at` DATETIME - 更新时间

## Goals / Non-Goals

**Goals:**
- 提供 `res_plan_demand_gpu_template` 表的完整CRUD数据访问能力
- 遵循项目现有的 resource-plan 域代码模式，保持一致性
- 通过data-service暴露REST API供上层服务调用

**Non-Goals:**
- 不涉及上层业务服务（cloud-server/woa-server）的接口开发
- 不涉及前端页面开发
- 不涉及模版内容（schema JSON）的结构校验逻辑
- 不涉及权限控制和审计日志

## Decisions

### 1. 遵循现有 resource-plan CRUD 模式
**选择**: 复用 res_plan_week 的完整代码结构（table → dao → api → service）
**理由**: 项目已有统一且成熟的模式，保持一致性降低维护成本。`schema` 字段使用 `types.JsonField` 类型，与项目中其他JSON字段（如 res_plan_ticket 的 demands 字段）处理方式一致。

### 2. schema 字段使用 types.JsonField
**选择**: 使用项目已有的 `types.JsonField` 类型
**替代方案**: 定义强类型结构体解析JSON
**理由**: 模版schema结构由业务层定义且可变，data-service层不应耦合具体结构，使用 `JsonField` 透传即可。

### 3. API路径设计
**选择**: `/res_plans/demand_gpu_templates/{action}` 风格
**理由**: 与现有 resource-plan 路径风格一致（如 `/res_plans/res_plan_weeks/list`）。

## Risks / Trade-offs

- **[schema字段无校验]** → data-service层不校验JSON结构，由上层服务负责内容校验。这是项目统一模式，JsonField本身已处理数据库读写。
- **[无审计日志]** → 模版配置为管理型数据，非云资源操作，当前阶段不需要审计。后续如需添加可参考其他DAO的Audit接口扩展。
