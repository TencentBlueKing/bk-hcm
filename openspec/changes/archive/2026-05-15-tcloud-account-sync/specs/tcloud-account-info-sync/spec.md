## ADDED Requirements

### Requirement: Account sync method added to Interface
The system SHALL expose an `Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error)` method on the
`cmd/hc-service/logics/res-sync/tcloud.Interface`, enabling callers to trigger synchronization of a single
TCloud second-level account (资源账号) from cloud to DB.

#### Scenario: Valid resource account is synced
- **WHEN** `Account` is called with a valid `accountID` for a TCloud ResourceAccount
- **THEN** the system fetches cloud data, compares with DB, and updates only changed fields

#### Scenario: Account has empty CloudSubAccountID
- **WHEN** `Account` is called for an account whose extension has an empty `cloud_sub_account_id`
- **THEN** the system skips cloud API calls and returns an empty `SyncResult` without error

---

### Requirement: Email and CloudCreatedAt synced from cloud
The system SHALL call TCloud `DescribeSubAccounts` API using the account's `cloud_sub_account_id` UIN, and
update the account's `email` and `cloud_created_at` fields in DB to match the cloud values.

#### Scenario: Cloud returns updated email
- **WHEN** the cloud `email` differs from the DB value
- **THEN** the account's `email` field is updated to the cloud value

#### Scenario: Cloud and DB email are identical
- **WHEN** the cloud `email` is the same as the DB value
- **THEN** no update call is made for the account

#### Scenario: Cloud returns null for CreateTime
- **WHEN** `DescribeSubAccounts` returns `create_time` as nil
- **THEN** `cloud_created_at` is left unchanged in DB

---

### Requirement: LoginFlag and ActionFlag synced from cloud
The system SHALL call TCloud `DescribeSafeAuthFlag` API using the account's sub-account UIN, and update
the `login_flag` and `action_flag` fields in the account's extension to match the cloud values.
A nil cloud result MUST clear the corresponding DB value (nil = no protection set).

#### Scenario: Cloud returns a LoginFlag value
- **WHEN** `DescribeSafeAuthFlag` returns a non-nil `LoginFlag`
- **THEN** `extension.login_flag` is set to the mapped `AccountProtectionFlag` value

#### Scenario: Cloud returns nil LoginFlag (no protection set)
- **WHEN** `DescribeSafeAuthFlag` returns nil for `LoginFlag`
- **THEN** `extension.login_flag` is set to null in DB, overwriting any previously stored value

#### Scenario: CloudSubAccountID equals CloudMainAccountID
- **WHEN** the account's `cloud_sub_account_id` matches `cloud_main_account_id`
- **THEN** `DescribeSafeAuthFlag` is skipped; LoginFlag and ActionFlag remain unchanged in DB

---

### Requirement: Extension patch via SyncExtensionPatch to support null writes
The system SHALL update LoginFlag/ActionFlag using a `SyncExtensionPatch *json.RawMessage` field on
`AccountUpdateReq`. The patch JSON is merged into the existing DB extension via `json.UpdateMerge` (gjson
@join), so only the specified fields are affected; all other extension fields (including encrypted
`cloud_secret_key`) are left unchanged. A nil value in the patch MUST be serialized as JSON `null` (no
`omitempty`) so that @join correctly overwrites the DB field.

#### Scenario: LoginFlag cleared to null via SyncExtensionPatch
- **WHEN** `SyncExtensionPatch` is `{"login_flag": null, "action_flag": null}`
- **THEN** data-service merges the patch into the DB extension, resulting in `login_flag: null` while all other extension fields are preserved

#### Scenario: Normal cloud-server extension updates are unaffected
- **WHEN** cloud-server calls `AccountUpdateReq` with the existing `Extension` field (not `SyncExtensionPatch`)
- **THEN** data-service applies `UpdateMerge` as before; LoginFlag/ActionFlag in DB are preserved

---

### Requirement: AccountSecret status synced from cloud
The system SHALL call TCloud `ListAccessKeys` using the account's sub-account UIN, match the result against
the account secret's `cloud_secret_id` (stored in account_secret extension), and update the secret's
`status` field (`normal` or `invalid`) to reflect the cloud key status (`Active`/`Inactive`).

#### Scenario: Cloud key status is Active
- **WHEN** `ListAccessKeys` returns `Status: Active` for the matching `cloud_secret_id`
- **THEN** the account_secret `status` is set to `normal`

#### Scenario: Cloud key status is Inactive
- **WHEN** `ListAccessKeys` returns `Status: Inactive` for the matching `cloud_secret_id`
- **THEN** the account_secret `status` is set to `invalid`

#### Scenario: No matching key found in cloud
- **WHEN** no key in `ListAccessKeys` result matches the account secret's `cloud_secret_id`
- **THEN** the secret status is left unchanged and a warning is logged

#### Scenario: ListAccessKeys call fails due to insufficient permission
- **WHEN** `ListAccessKeys` returns a permission error
- **THEN** the error is logged and secret status sync is skipped; the overall sync does not fail
