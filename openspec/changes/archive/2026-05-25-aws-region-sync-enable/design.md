## Context

当前 AWS 地域同步逻辑位于 `cmd/cloud-server/service/sync/aws/region.go`，通过 `ListRegion` 函数获取本地 `aws_region` 表中的地域列表，再逐一调用云厂商 API 同步资源。

现有问题：
- `aws_region` 表仅包含地域基础信息（region_id、region_name、status、endpoint），缺少同步开关控制字段；
- 当账号在某些地域无权限（如 opt-in regions 未启用）时，API 调用失败会阻断整个同步流程；
- 运维人员无法通过配置或接口灵活控制哪些地域参与同步。

已有数据结构：
- 表结构定义：`pkg/dal/table/cloud/region/aws.go` - `AwsRegionTable`、`AwsRegionColumnDescriptor`
- Core 结构体：`pkg/api/core/cloud/region/aws.go` - `AwsRegion`
- DS 结构体：`pkg/api/data-service/cloud/region/aws.go` - `AwsRegionBatchCreate`、`AwsRegionBatchUpdate`
- DS 服务层：`cmd/data-service/service/cloud/region/aws.go` - `BatchCreateAwsRegion`、`BatchUpdateAwsRegion`、`convertAwsBaseRegion`
- Cloud-Server 层：`cmd/cloud-server/service/region/region.go` - `RegionSvc`、`InitRegionService`
- 同步逻辑：`cmd/cloud-server/service/sync/aws/region.go` - `SyncRegion`、`ListRegion`

## Goals / Non-Goals

**Goals:**
- 在 `aws_region` 表新增 `sync_enable` 字段，支持按地域控制同步开关。
- 修改 AWS 地域查询逻辑，支持按 `sync_enable` 字段过滤，同步任务仅处理启用的地域。
- 提供 API 接口管理地域同步状态，支持批量启用/禁用。
- 在同步日志中输出被过滤的地域信息，便于运维排查。
- 默认所有地域启用同步（sync_enable=1），保持向后兼容。

**Non-Goals:**
- 不改变现有地域同步的核心逻辑（增量/全量比对机制保持不变）。
- 本期不扩展到其他云厂商（TCloud、HuaWei、GCP、Azure），后续按需扩展。
- 不修改前端代码（仅后端接口扩展，前端按需接入）。
- 不做地域自动探测与自动禁用（由运维人员主动管理）。

## Decisions

### D1: 字段设计使用 TINYINT(1) 而非 BOOLEAN

**选择：** 使用 `sync_enable TINYINT(1) NOT NULL DEFAULT 1` 作为字段定义。

**理由：**
- 与项目中现有布尔语义字段（如 `status`）保持一致的数据类型风格；
- MySQL 中 BOOLEAN 实际为 TINYINT(1) 的别名，两者等效；
- 1 表示启用，0 表示禁用，语义清晰。

### D2: 默认启用同步，保持向后兼容

**选择：** `sync_enable` 字段默认值为 1（启用），存量数据通过迁移脚本补填为 1。

**理由：**
- 保证现有系统升级后行为不变，不会因新增字段导致地域被意外跳过；
- 运维人员按需禁用特定地域，而非逐个启用。

### D3: 同步过滤在 DAO 查询层实现

**选择：** 在资源同步调用 region 查询接口时，通过 filter 条件增加 `sync_enable = 1` 过滤，而非在同步逻辑层做内存过滤。

**理由：**
- 在数据库层过滤可减少无效数据传输；
- 复用现有通用 filter 机制，无需改动查询框架；
- 查询接口支持 `sync_enable` 字段作为过滤条件，调用方按需使用。

### D4: 管理接口采用批量更新模式

**选择：** 提供批量更新接口 `PATCH /api/v1/cloud/vendors/{vendor}/regions/sync_enable/batch`，支持同时指定多个 region ID 及目标状态。

**理由：**
- 运维场景常需批量操作（如一次性禁用多个异常地域）；
- 批量接口减少请求次数，提升操作效率；
- 保持 RESTful 风格，与现有接口设计一致；
- 路由中包含 `{vendor}` 参数，便于后续扩展到其他云厂商。

### D5: 新地域入库时默认启用

**选择：** 地域同步从云端获取新地域后，入库时 `sync_enable` 设为 1（启用）。

**理由：**
- 新地域默认参与同步是合理预期；
- 若新地域有问题，运维人员可通过管理接口手动禁用。

## Risks / Trade-offs

- **[风险] 运维人员误操作禁用关键地域导致资源不同步**
  → 管理接口需做权限控制，仅管理员可操作；接口返回中明确展示受影响的地域列表。

- **[风险] 存量数据迁移遗漏导致地域被跳过**
  → 迁移脚本需在服务升级前执行，且需验证 `sync_enable` 全部为 1。

- **[权衡] 本期仅支持 AWS，其他云厂商暂不支持**
  → 后续可按相同模式扩展，表结构和接口设计已考虑扩展性。

## Migration Plan

1. **数据库迁移**：执行 DDL 新增 `sync_enable` 字段，存量数据默认填充为 1：
   ```sql
   ALTER TABLE aws_region ADD COLUMN sync_enable TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用同步：1-启用，0-禁用';
   ```

2. **服务部署**：部署新版本 `data-service` 和 `cloud-server`。

3. **验证**：
   - 检查 `aws_region` 表所有记录 `sync_enable = 1`；
   - 触发一次地域同步任务，确认所有地域正常同步；
   - 调用管理接口禁用某地域，验证下次同步时该地域被跳过。

4. **回滚策略**：
   - 新增字段向后兼容，旧版本代码不读取 `sync_enable` 字段，不影响运行；
   - 若需回滚，可保留字段或执行 `ALTER TABLE aws_region DROP COLUMN sync_enable`。
