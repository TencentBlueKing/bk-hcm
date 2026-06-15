## ADDED Requirements

### Requirement: Biz SubAccount Ext Operable Calculation
The system MAY refactor `ListBizSubAccountExt` and `listSubAccountExt` implementation, but SHALL preserve equivalent behavior for auth scope, filter composition, vendor routing, and response semantics, then perform post-processing to calculate per-item operability.

The post-processing SHALL use each sub account's `account_id` and current `bk_biz_id` to determine operability:
- operable is `true` when current `bk_biz_id` equals the corresponding account record's biz id
- otherwise operable is `false`

#### Scenario: Keep equivalent list behavior after refactor
- **WHEN** `ListBizSubAccountExt` handles request
- **THEN** request validation, auth scope, and filter behavior remain equivalent to baseline behavior
- **THEN** vendor-specific list dispatch behavior remains equivalent
- **THEN** operable calculation happens after list result is obtained

#### Scenario: Operable is true for matching biz
- **WHEN** sub account item has `account_id=A1` and account `A1` belongs to biz `B100`
- **AND** current request `bk_biz_id` is `B100`
- **THEN** response item operable is `true`

#### Scenario: Operable is false for non-matching or missing account
- **WHEN** account biz does not equal current `bk_biz_id` or account record is missing
- **THEN** response item operable is `false`

### Requirement: Account Operable Public Function
The system SHALL provide a reusable service-layer public function under `cmd/cloud-server/logics/account` that supports batch input of `bk_biz_id` and `account_id[]`, and returns account-to-boolean mapping for operability.

The function SHALL support:
- batch account id input
- map-style output keyed by account id
- deterministic behavior for missing accounts (`false`)

#### Scenario: Batch calculation returns map
- **WHEN** input contains `bk_biz_id=B100` and account ids `[A1, A2, A3]`
- **THEN** function returns mapping that includes `A1/A2/A3` with boolean operable values

### Requirement: Extended Response Structure for Biz SubAccount Ext
The system SHALL NOT add fields directly to `SubAccount` core structure, and SHALL define a new response structure that contains:
- original `SubAccount`
- `operable` boolean
- `account_name` string

#### Scenario: Response uses new composite structure
- **WHEN** list response is built
- **THEN** each item contains all original SubAccount fields plus `operable` and `account_name`

### Requirement: Account Name Fill by AccountID
After operability calculation, the system SHALL query account names by `account_id` and fill `account_name` for each response item.

#### Scenario: Account name is filled
- **WHEN** account `A1` exists and its name is `prod-main-account`
- **THEN** response item with `account_id=A1` includes `account_name="prod-main-account"`

#### Scenario: Missing account yields empty name
- **WHEN** account record does not exist for item account_id
- **THEN** response item includes empty `account_name`

### Requirement: Conversion Functions Must Be Encapsulated
The system SHALL encapsulate transformation from original sub account list to the new composite response list in dedicated conversion functions.

#### Scenario: Centralized conversion
- **WHEN** service builds final list response
- **THEN** it uses dedicated conversion helper(s) instead of inlining field assembly in handler logic
