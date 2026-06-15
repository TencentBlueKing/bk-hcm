## Why

审批流程表（`approval_process`）中 `application_type` 与审批中心 `service_id` 为 1:1 关系，导致一个 `type` 只能对应一条审批流。但现实业务中存在多种操作共用同一审批流的场景（如 `operate_sub_account` 下同时包含创建子账号、更新子账号等操作），而前端通过 `type` 字段对申请单分类查询，无法区分同一 `type` 下的不同操作，影响了申请单的精细化管理和展示。

## What Changes

- **新增 `operation` 字段**：在 `application` 表中新增 `operation` 字段，代表细粒度操作类型；后端查询接口支持以 `operation` 作为过滤条件，供前端按细粒度操作类型查询申请单（不涉及前端代码修改，属于接口语义上的调整）。
- **定义 `ApplicationOperation` 枚举**：在 `enumor` 包中完善 `ApplicationOperation` 类型，既兼容现有 `ApplicationType` 的值（对于一个 type 只有一种操作的情况，`operation = type`），也扩展细粒度的新操作值（如 `create_sub_account`、`update_sub_account`）。
- **`ApplicationHandler` 接口扩展**：新增 `GetOperation()` 方法，各业务 handler 实现自己的 operation 值。
- **`BaseApplicationHandler` 扩展**：新增 `operation` 字段，`NewBaseApplicationHandler` 支持传入 operation。
- **创建申请单流程调整**：`createApplication` 函数写入 `operation` 字段；`bkBizIDs` 记录逻辑从 `type` 判断改为 `operation` 判断。
- **查询接口调整**：`ListBizApplications` 等查询接口支持以 `operation` 字段过滤，data-service 侧响应体新增 `operation` 字段返回。
- **数据迁移**：为存量数据写入 `operation`（与其 `type` 值相同）以保证向后兼容。

## Capabilities

### New Capabilities

- `application-operation-field`：application 申请单细粒度操作字段，定义操作枚举、handler 接口扩展、创建/查询流程改造。

### Modified Capabilities

## Impact

- **数据层**：`application` 表新增 `operation` 列（需 DDL 变更及数据迁移脚本）。
- **API 层**：`data-service` 的 `ApplicationCreateReq` / `ApplicationResp` / `ApplicationListResult` 新增 `operation` 字段；cloud-server 侧查询接口 `ListBizApplications` 支持按 `operation` 过滤。
- **枚举定义**：`pkg/criteria/enumor/application.go` 中 `ApplicationOperation` 类型从空完善为包含全量常量定义。
- **Handler 层**：所有实现 `ApplicationHandler` 接口的 handler 需实现新增的 `GetOperation()` 方法（约 15+ 个 handler 文件）。
- **`bkBizIDs` 判断逻辑**：`create.go` 中判断是否记录业务 ID 的条件从对比 `applicationType` 改为对比 `operation`，避免 `type` 粗粒度导致漏记或多记。
- **接口语义**：后端对外暴露 `operation` 字段，前端可通过该字段对申请单进行细粒度分类查询（不涉及前端代码变更，由前端自行选择时机接入）。
