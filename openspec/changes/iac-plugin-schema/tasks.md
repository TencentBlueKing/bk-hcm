# Tasks

## 1. Dependencies, SQL Scripts, and Core Models

- [x] 1.1 Add `github.com/hashicorp/terraform-json` to `go.mod` / `go.sum` and verify `go mod tidy` succeeds
- [x] 1.2 Create the `iac_plugin_schema` DDL migration script in `scripts/sql/`, and verify the SQL syntax matches the schema specification
- [x] 1.3 在 `pkg/api/core/plugin/schema.go` 和 `pkg/api/core/plugin/schema_validate.go` 中定义 `PluginSchema`、`SchemaDetail` 及可复用的导出校验与哈希函数（`ValidatePluginSchema`、`ValidateCapabilities`、`ComputeSchemaHash`）；data-service 列表响应直接复用 `IacPluginSchemaTable`，并通过单元测试验证 `json:",inline"` 序列化、反序列化及全部校验规则

## 2. DAL Table and DAO Layer (`pkg/dal`)

- [x] 2.1 Implement `IacPluginSchemaTable` in `pkg/dal/table/plugin/schema.go`, register it in `pkg/dal/table/table.go` with `EnableTenant: false`, and verify table validation and column descriptors pass unit tests
- [x] 2.2 Implement `IacPluginSchema` DAO in `pkg/dal/dao/plugin/schema.go`, register it in `pkg/dal/dao/dao.go` and `pkg/dal/dao/types/`, and verify `go build ./pkg/dal/...` succeeds

## 3. Data-Service REST API and Global SDK Client

- [x] 3.1 在 `pkg/api/data-service/plugin/schema.go` 中定义 data-service 请求与响应类型，其中 Schema 列表结果直接复用 `IacPluginSchemaTable`，并通过单元测试验证请求校验
- [x] 3.2 Implement `data-service` REST handlers for `iac_plugin_schema` (Create, List supporting `fields` projection for `schema_hash` and full record queries) in `cmd/data-service/service/plugin/schema.go`, wire them into `cmd/data-service/service/service.go`, and verify `go build ./cmd/data-service/...` succeeds
- [x] 3.3 Implement Global SDK Client methods for `iac_plugin_schema` in `pkg/client/data-service/global/plugin_schema.go`, wire them into `pkg/client/data-service/global/client.go`, and verify `go build ./pkg/client/...` succeeds

## 4. HC-Service SchemaManager and SchemaCache

- [x] 4.1 Define `SchemaGetter`, `SchemaManager`, and `SyncSchemaResult` in `cmd/hc-service/logics/plugin/schema/types.go` and implement `SchemaCache` (when cache misses, expires, or `schema_hash` mismatches DB, query full schema from DB via `singleflight`, spawn a goroutine to update `schemaCache` + `expireMap`, and concurrently return schema to caller; include structured English logging via `hcm/pkg/logs` with `kt.Rid`) in `cmd/hc-service/logics/plugin/schema/cache.go`, and verify cache hit/miss/TTL expiration/hash-mismatch/concurrent-singleflight scenarios pass unit tests in `cmd/hc-service/logics/plugin/schema/cache_test.go`
- [x] 4.2 Implement `SchemaManager.SyncSchema` and `SchemaManager.GetSchema` in `cmd/hc-service/logics/plugin/schema/manager.go` (calling the exported `ValidatePluginSchema` function to validate against `iac_plugin`, `protocol_version`, `schema_hash` regex + local `SchemaDetail` SHA-256 recomputation match, southbound protocol `capabilities` rules, and `provider.block`, followed by idempotent write, tamper-resistant conflict rejection, and English logging via `hcm/pkg/logs`), and verify all `SyncSchema` and `GetSchema` scenarios pass unit tests in `cmd/hc-service/logics/plugin/schema/manager_test.go`

## Workflow follow-up

- Archive the change with `/opsx-archive` after review and implementation verification are complete.
