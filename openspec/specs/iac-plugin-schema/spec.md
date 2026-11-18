# iac-plugin-schema Specification

## Purpose

提供海垒（HCM）IaC 插件 Schema 的权威存储、同步校验、防篡改幂等落库以及 `hc-service` 副本级懒加载缓存与按版本读取能力，支撑本地插件（`tf_provider`）与三方 HTTP 插件（`http`）的 Schema 生命周期管理。

## Requirements

### Requirement: Authoritative storage of plugin schemas by plugin name and version
The system SHALL persist plugin schemas in the system-level `iac_plugin_schema` table in `data-service` without `tenant_id`, indexed uniquely by `(plugin_name, plugin_version)`. The storage layer MUST store `id`, `plugin_name`, `plugin_version`, and `schema_hash` (`varchar(160)`, supporting `sha256`, `sha384`, and `sha512` prefixed digests) in standalone columns, serialize `SchemaDetail` into the `schema_detail` JSON column, and support querying full records or lightweight `schema_hash` projections.

#### Scenario: Create a new plugin schema record in data-service
- **WHEN** `data-service` receives a request to create a plugin schema for an unregistered `(plugin_name, plugin_version)` pair with valid `schema_hash` and `SchemaDetail`
- **THEN** `data-service` inserts the record into `iac_plugin_schema` with a generated primary key `id` and returns the created ID

#### Scenario: Query only lightweight schema_hash from data-service
- **WHEN** a caller queries `iac_plugin_schema` by `(plugin_name, plugin_version)` specifying `fields: ["schema_hash"]`
- **THEN** `data-service` returns the matching record containing `schema_hash` without transferring the multi-megabyte `schema_detail` JSON column

### Requirement: Unified PluginSchema data model with inlined SchemaDetail
The system SHALL represent a plugin schema using `PluginSchema`, which combines top-level fields (`plugin` of type `PluginInfo{Name, Version}`, `schema_hash`) with an inlined `SchemaDetail` (embedding `tfjson.ProviderSchema` plus `protocol_version` and `capabilities`) backed by standard `terraform-json` schema structures.

#### Scenario: Serialize and deserialize PluginSchema as flattened JSON
- **WHEN** `PluginSchema` is serialized to or deserialized from JSON
- **THEN** `plugin` (`name`, `version`), `schema_hash`, and all `SchemaDetail` fields appear at the same top-level JSON object hierarchy

#### Scenario: Persist only SchemaDetail into the database JSON column
- **WHEN** `PluginSchema` is persisted into the `iac_plugin_schema.schema_detail` column
- **THEN** only the `SchemaDetail` portion is serialized into the `schema_detail` JSON column while `plugin.name`, `plugin.version`, and `schema_hash` are stored in the `plugin_name`, `plugin_version`, and `schema_hash` table columns

### Requirement: Reusable exported schema validation against plugin metadata and protocol rules
The system MUST provide a single exported validation entry (`ValidatePluginSchema`) reusable across `SchemaManager.SyncSchema` and other modules, while `SchemaManager.SyncSchema` verifies that `(plugin.name, plugin.version)` exists in `iac_plugin`. `ValidatePluginSchema` MUST check `protocol_version` (`http` non-empty, `tf_provider` allowed empty), verify `schema_hash` matches `<algo>:<lowercase hex>` (`algo` in `sha256`, `sha384`, `sha512`) and equals the computed digest of `(plugin, SchemaDetail)`, validate `capabilities` via `(*SchemaDetail).ValidateCapabilities()` per southbound protocol rules (using a strongly typed `Capabilities` struct that ignores unknown JSON keys for forward-compatible protocol evolution), and verify `provider.block` and type schemas are valid.

#### Scenario: Reject schema when plugin name or version does not exist in iac_plugin
- **WHEN** `SchemaManager.SyncSchema` receives a `PluginSchema` whose `plugin.name` or `plugin.version` is empty or does not exist in `iac_plugin`
- **THEN** `SchemaManager.SyncSchema` aborts synchronization and returns a failed `SyncSchemaResult` without writing to `iac_plugin_schema`

#### Scenario: Reject HTTP plugin schema with empty protocol_version
- **WHEN** `SchemaManager.SyncSchema` validates a schema for a plugin whose `iac_plugin.protocol` is `http` and `protocol_version` is empty
- **THEN** `SchemaManager.SyncSchema` rejects the schema and returns a validation error

#### Scenario: Allow tf_provider plugin schema with empty protocol_version
- **WHEN** `SchemaManager.SyncSchema` validates a schema for a plugin whose `iac_plugin.protocol` is `tf_provider` and `protocol_version` is empty while all other fields are valid
- **THEN** `SchemaManager.SyncSchema` passes protocol version validation

#### Scenario: Reject schema when schema_hash format is invalid or does not match locally computed hash
- **WHEN** `SchemaManager.SyncSchema` receives a schema whose `schema_hash` is not a valid `sha256`/`sha384`/`sha512` `<algo>:<lowercase hex>` digest OR whose locally computed digest over canonical `(plugin, SchemaDetail)` JSON does not equal the returned `schema_hash`
- **THEN** `SchemaManager.SyncSchema` rejects the schema and returns a validation error

#### Scenario: Reject schema when capabilities violate southbound protocol rules while ignoring unknown keys
- **WHEN** `SchemaManager.SyncSchema` receives a schema with empty `capabilities`, `resource_schemas` non-empty without `apply`/`plan`/`refresh` all set to `true`, `resource_schemas` empty with any resource capability set to `true`, `data_source_schemas` non-empty without `read_data: true`, `data_source_schemas` empty with `read_data: true`, or `stop: true` without `async_apply: true`
- **THEN** `SchemaManager.SyncSchema` rejects the schema and returns a validation error identifying the violated capability rule, whereas unknown capability keys in raw JSON are ignored without error during deserialization

#### Scenario: Reject schema with missing provider block
- **WHEN** `SchemaManager.SyncSchema` receives a schema whose `provider` or `provider.block` is nil
- **THEN** `SchemaManager.SyncSchema` rejects the schema and returns a validation error

### Requirement: Idempotent and tamper-resistant Schema synchronization
After passing validation, `SchemaManager.SyncSchema` MUST check `iac_plugin_schema` by `(plugin_name, plugin_version)`. If no record exists, it MUST insert the schema and return `sync_result: "success"` with an empty `sync_message` and the schema, without populating the in-memory cache. If a record exists with the exact same `schema_hash`, it MUST skip writing and return `sync_result: "success"` with an empty `sync_message` and the schema. If a record exists with a different `schema_hash`, or if fetching or validation fails, it MUST log an error, refuse to overwrite any record, and return `sync_result: "failed"` with `sync_message` set to the reason and `schema` nil.

#### Scenario: First-time sync for a new plugin version
- **WHEN** `SchemaManager.SyncSchema` is called for a `(plugin_name, plugin_version)` that does not exist in `iac_plugin_schema`
- **THEN** the system inserts the record into `iac_plugin_schema`, does not write to `SchemaCache`, and returns `SyncSchemaResult` with `sync_result: "success"`, an empty `sync_message`, and the schema

#### Scenario: Repeat sync with identical schema_hash
- **WHEN** `SchemaManager.SyncSchema` is called for an existing `(plugin_name, plugin_version)` and the incoming `schema_hash` equals the stored `schema_hash`
- **THEN** the system skips database insertion and returns `SyncSchemaResult` with `sync_result: "success"`, an empty `sync_message`, and the validated `PluginSchema`

#### Scenario: Conflicting sync with different schema_hash on the same version
- **WHEN** `SchemaManager.SyncSchema` is called for an existing `(plugin_name, plugin_version)` and the incoming `schema_hash` differs from the stored `schema_hash`
- **THEN** the system logs an error, preserves the existing database record unchanged, and returns `sync_result: "failed"` with the mismatch reason in `sync_message` and `schema` nil

### Requirement: Per-replica lazy caching with lightweight hash validation and expiration eviction
`SchemaManager.GetSchema` MUST use a per-replica `SchemaCache` (`schemaCache` + `expireMap`), and `SchemaManager.Close()` MUST stop the background eviction goroutine. On a non-expired hit, it MUST query only `schema_hash` from `data-service`, extend TTL on match, and return the cached read-only `*PluginSchema`. Whenever a read misses, hits an expired entry, or detects a `schema_hash` mismatch, it MUST query the full schema from `data-service` via `singleflight`, synchronously update `schemaCache` and `expireMap`, and return the schema to the caller.

#### Scenario: Cache hit within TTL with matching database schema_hash
- **WHEN** `SchemaManager.GetSchema` finds a non-expired entry in `SchemaCache` and the lightweight `schema_hash` queried from `data-service` equals the cached `SchemaHash`
- **THEN** the system refreshes the entry's expiration timestamp in `expireMap` by adding the configured TTL and returns the cached `*PluginSchema` without fetching the full JSON column from `data-service`

#### Scenario: Cache hit within TTL when database record was deleted
- **WHEN** `SchemaManager.GetSchema` finds a non-expired entry in `SchemaCache` but `data-service` returns no record for `(plugin_name, plugin_version)`
- **THEN** the system removes the entry from both `schemaCache` and `expireMap` and returns a not-found error

#### Scenario: Cache expired or schema_hash mismatch queries DB and updates cache synchronously
- **WHEN** `SchemaManager.GetSchema` encounters an expired cache entry OR a non-expired entry whose `SchemaHash` differs from the `schema_hash` in `data-service`
- **THEN** the system queries the full schema from `data-service` via `singleflight`, updates `schemaCache` and `expireMap` synchronously (and signals `evictCh` on expiration to clean up any other expired cold keys), and returns the fetched schema to the caller

### Requirement: Concurrent back-to-source merging and synchronous cache population
When `SchemaCache` misses, expires, or detects a `schema_hash` mismatch, the system MUST use `singleflight` keyed by `pluginName:pluginVersion` to coalesce concurrent full-schema queries to `data-service`. Once the full schema is fetched from `data-service`, the system MUST update `schemaCache` and `expireMap` synchronously before returning the schema to all waiting callers.

#### Scenario: Concurrent cache misses for the same plugin and version
- **WHEN** multiple goroutines concurrently call `SchemaManager.GetSchema` for the same uncached `(plugin_name, plugin_version)`
- **THEN** only one full-schema query is sent to `data-service` via `singleflight`, the cache is populated once before returning, and the fetched schema is returned to all callers

#### Scenario: Query non-existent plugin schema on cache miss
- **WHEN** `SchemaManager.GetSchema` misses the cache and `data-service` has no record for `(plugin_name, plugin_version)`
- **THEN** the system returns a not-found error without populating `SchemaCache`
