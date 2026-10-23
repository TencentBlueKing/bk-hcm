# Spec Delta

## Purpose

提供海垒（HCM）IaC 插件实例（`iac_plugin_instance`）的多租户持久化存储、对照活跃插件 `Schema.provider.block` 的动态属性递归校验与敏感字段同构拆分、基于 `data-service` `Cipher` 的敏感叶子值加解密，以及面向内部服务的完整实例生命周期管理（CRUD）与明文 `properties` 深度合并还原能力。

## ADDED Requirements

### Requirement: Multi-tenant storage of plugin instances in data-service
The system SHALL persist plugin instances in the tenant-enabled `iac_plugin_instance` table in `data-service`, storing `id`, `plugin_name`, `name`, `attrs` (JSON), `secret_attrs` (JSON), `extension` (JSON), `memo`, `tenant_id`, and audit fields (`creator`, `reviser`, `created_at`, `updated_at`), with immutable `plugin_name` after creation.

#### Scenario: Create plugin instance records in data-service
- **WHEN** `data-service` receives a valid batch request to create plugin instances with `plugin_name`, `name`, `attrs`, `secret_attrs`, `extension`, and optional `memo`
- **THEN** `data-service` generates primary key `id`s, persists the records scoped to the caller's `tenant_id`, and returns the created `ids`

#### Scenario: Prevent updating immutable plugin_name on an existing instance
- **WHEN** a caller attempts to update `plugin_name` on an existing `iac_plugin_instance` record
- **THEN** the storage validation rejects the update and returns an invalid parameter error

#### Scenario: Delete existing plugin instance records
- **WHEN** `data-service` receives a batch delete request with a filter expression within the caller's tenant scope
- **THEN** `data-service` deletes the matching records from `iac_plugin_instance`

### Requirement: Leaf-level encryption and decryption of secret_attrs in data-service
`data-service` MUST encrypt all scalar leaf values inside `secret_attrs` using `cipher.EncryptToBase64(string(json.Marshal(leafVal)))` before database insertion or update while preserving all nested map keys, slice structures, and `{}` index placeholders in plaintext. On list queries, `data-service` MUST decrypt all scalar leaf ciphertexts in `secret_attrs` and unmarshal them back to their original scalar types before returning.

#### Scenario: Encrypt nested scalar leaf values of secret_attrs before persistence
- **WHEN** `data-service` creates or updates a plugin instance whose `secret_attrs` contains top-level scalars, nested maps, or arrays of objects
- **THEN** `data-service` preserves all map keys, array structures, and `{}` placeholders in plaintext in the database and stores each scalar leaf value as `cipher.EncryptToBase64(string(json.Marshal(leafVal)))`

#### Scenario: Decrypt nested scalar leaf values of secret_attrs on query
- **WHEN** `data-service` retrieves plugin instance records via list
- **THEN** `data-service` decrypts each encrypted leaf string in `secret_attrs` via `cipher.DecryptFromBase64` and `json.Unmarshal`, returning the restored scalar values (such as `string`, `number`, or `bool`) in the response

### Requirement: Active plugin availability and provider schema validation before instance mutation
When creating a plugin instance or updating its `properties`, `hc-service` MUST verify that the target `plugin_name` exists in the plugin registry (`iac_plugin`) with `state = "enabled"`, and MUST load the active version's `PluginSchema` via schema query (`GetSchema`) to obtain `provider.block` for property validation.

#### Scenario: Reject instance creation when plugin is unregistered or not enabled
- **WHEN** `CreatePluginInstance` is called for a `plugin_name` that does not exist in the plugin registry or whose `state` is not `enabled`
- **THEN** `hc-service` rejects the request and returns an invalid parameter error without creating an instance

#### Scenario: Reject instance creation when active plugin schema cannot be loaded
- **WHEN** `CreatePluginInstance` is called for an enabled plugin whose active version schema is missing or lacks `provider.block`
- **THEN** `hc-service` rejects the request and returns an error without creating an instance

### Requirement: Recursive schema validation of instance properties against provider.block
`hc-service` MUST recursively validate `properties` top-down against `provider.block` (`attributes` and `block_types`). It MUST reject unknown attributes or blocks, reject pure `computed` attributes (`computed = true` and `optional = false` and `required = false`), validate required fields and scalar/collection/object value types against the schema.

#### Scenario: Reject unknown attribute or block with full property path
- **WHEN** `CreatePluginInstance` or `UpdatePluginInstance` receives a property key not declared in the corresponding `block.attributes` or `block.block_types` (at top level or inside a nested block/object)
- **THEN** the system rejects the request with `errf.InvalidParameter` containing the full property path (e.g., `assume_role[0].unknown_field`)

#### Scenario: Reject pure computed read-only attribute
- **WHEN** `CreatePluginInstance` or `UpdatePluginInstance` receives a property whose schema attribute has `computed: true`, `optional: false`, and `required: false`
- **THEN** the system rejects the request with `errf.InvalidParameter` indicating that read-only computed attributes cannot be preset

#### Scenario: Reject property with mismatched value type
- **WHEN** `CreatePluginInstance` or `UpdatePluginInstance` receives a property whose JSON value does not match the declared `cty.Type` or `nesting_mode` (for example, passing a string where a number or list block is required)
- **THEN** the system rejects the request with `errf.InvalidParameter` identifying the mismatched property path

### Requirement: Top-down isomorphic splitting of properties into attrs and secret_attrs
During property validation, the system MUST split `properties` into two isomorphic JSON trees (`attrs` and `secret_attrs`). Once a node along a path has `sensitive = true`, its entire subtree MUST be placed into `secret_attrs`. Non-sensitive composite blocks or objects MUST be traversed recursively, placing non-sensitive leaves in `attrs` and sensitive subtrees in `secret_attrs`, using `{}` placeholders in sparse array elements to keep array indices aligned.

#### Scenario: Split top-level sensitive scalar and composite attributes into secret_attrs
- **WHEN** `properties` contains a non-sensitive attribute (e.g., `endpoint`), a sensitive scalar attribute (e.g., `secret_key`), and a top-level sensitive map attribute (e.g., `secret_headers: {"X-Token": "tk-99"}`)
- **THEN** `attrs` contains only `{"endpoint": ...}` and `secret_attrs` contains `{"secret_key": ..., "secret_headers": {"X-Token": ...}}` preserving nested keys under `secret_headers`

#### Scenario: Split nested block containing both sensitive and non-sensitive sibling attributes
- **WHEN** `properties` contains a non-sensitive list block `auth` where each element has non-sensitive `username` and sensitive `password`
- **THEN** `attrs` stores `{"auth": [{"username": "bot-01"}, {"username": "bot-02"}]}` and `secret_attrs` stores `{"auth": [{"password": ...}, {"password": ...}]}`

#### Scenario: Preserve array index alignment with empty object placeholder on sparse sensitive elements
- **WHEN** `properties` contains a list block `auth` where index `0` has only `{"username": "bot-01"}` (omitting optional `password`) and index `1` has `{"username": "bot-02", "password": "s3cret-02"}`
- **THEN** `attrs` stores `{"auth": [{"username": "bot-01"}, {"username": "bot-02"}]}` and `secret_attrs` stores `{"auth": [{}, {"password": ...}]}` so index `1` remains aligned at position `1`

### Requirement: Full replacement of attrs and secret_attrs on properties update
When `UpdatePluginInstance` is called with a non-nil `properties` map, `hc-service` MUST re-validate and re-split `properties` against the active `provider.block` and overwrite both `attrs` and `secret_attrs` in full. When `properties` is nil (only updating `name`, `extension`, or `memo`), existing `attrs` and `secret_attrs` MUST remain unchanged.

#### Scenario: Full overwrite of attrs and secret_attrs when updating properties
- **WHEN** `UpdatePluginInstance` is called with a new `properties` object that removes a previously stored key and adds a new nested block
- **THEN** `hc-service` validates the new `properties` against the active schema and completely replaces `attrs` and `secret_attrs` in `iac_plugin_instance` without retaining removed keys

#### Scenario: Partial metadata update without modifying properties
- **WHEN** `UpdatePluginInstance` is called with `name`, `extension`, or `memo` while `properties` is nil
- **THEN** `hc-service` updates only the specified metadata fields and keeps the stored `attrs` and `secret_attrs` intact

### Requirement: Deep merge reconstruction of full plaintext properties on Get and List
When `GetPluginInstance` or `ListPluginInstance` is called on `hc-service` (including its HTTP API), the system MUST deep-merge `attrs` and the decrypted `secret_attrs` by object key and matching array index `i` into a single plaintext `properties` map, and return `properties` and `extension` without exposing `attrs` or `secret_attrs` in the response.

#### Scenario: Reconstruct complete plaintext properties from sparse array and nested sensitive fields
- **WHEN** `GetPluginInstance` or `ListPluginInstance` reads an instance whose `attrs` has `{"auth": [{"username": "bot-01"}, {"username": "bot-02"}]}` and decrypted `secret_attrs` has `{"auth": [{}, {"password": "s3cret-02"}]}`
- **THEN** the returned `PluginInstance` contains `properties: {"auth": [{"username": "bot-01"}, {"username": "bot-02", "password": "s3cret-02"}]}` with index `0` and index `1` accurately merged

#### Scenario: Filter plugin instances by extension fields such as account_id
- **WHEN** `ListPluginInstance` is called with a filter expression matching `plugin_name` and `extension.account_id` (JSON extract filter)
- **THEN** the system returns the matching plugin instances with their reconstructed plaintext `properties` and `extension`
