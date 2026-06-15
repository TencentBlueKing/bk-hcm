## ADDED Requirements

### Requirement: Biz-scoped operable flag on sub account secret list items

When `page.count` is false, each detail row in the business-scoped sub account secret join list response SHALL include a boolean field `operable` indicating whether the current business may operate on that secret for subsequent management actions.

The system SHALL compute `operable` using the secret row's `account_id` and the requesting path `bk_biz_id`:

- `operable` is true when the account record for that `account_id` exists and its `bk_biz_id` equals the current request `bk_biz_id`
- otherwise `operable` is false (including when the account record is missing)

The system SHALL batch-resolve account metadata for distinct `account_id` values from the list result and SHALL reuse the shared account logic under `cmd/cloud-server/logics/account` used for biz sub-account ext operable mapping (equivalent semantics to `BuildOperableMapByAccountMap`).

The system SHALL encapsulate conversion from the data-service join list result into the cloud-server response shape that includes `operable` in dedicated helper(s), consistent with the biz sub-account ext list pattern. The composite detail type SHALL be named with a `Biz` prefix (for example `BizSubAccountSecretJoinExtDetail`) to indicate business-scoped response fields.

#### Scenario: Operable true when account belongs to current biz

- **WHEN** a returned secret has `account_id=A1` and account `A1` has `bk_biz_id` equal to the request `bk_biz_id`
- **THEN** that row's `operable` is true

#### Scenario: Operable false when account belongs to another biz or is missing

- **WHEN** account `A1` exists but its `bk_biz_id` differs from the request `bk_biz_id`, or no account exists for `account_id`
- **THEN** that row's `operable` is false

#### Scenario: Count-only mode unchanged

- **WHEN** `page.count` is true
- **THEN** the response follows existing count semantics and does not require per-row `operable` fields
