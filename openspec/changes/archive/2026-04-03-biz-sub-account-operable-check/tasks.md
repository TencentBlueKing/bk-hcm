## 1. Public Account Logic Functions

- [x] 1.1 Add a batch query helper in `cmd/cloud-server/logics/account` to fetch account basic info by `account_id[]` (at least account id, biz id, account name).
- [x] 1.2 Add a public operable check function in `cmd/cloud-server/logics/account` with input `bk_biz_id + account_id[]`, and output map keyed by `account_id` with boolean operable value.
- [x] 1.3 Add a helper to build `account_id -> account_name` map for response fill.

## 2. Extended Response Structure and Conversion

- [x] 2.1 Define a new composite response structure (without changing core `SubAccount`) containing `SubAccount`, `operable`, and `account_name`.
- [x] 2.2 Add conversion helper function(s) to transform original sub-account list and map data (`operable`, `account_name`) into the new composite response list.

## 3. ListBizSubAccountExt Post-Processing Integration

- [x] 3.1 Refactor list flow if needed, while keeping behavior equivalent (auth scope, filter composition, vendor routing, response semantics).
- [x] 3.2 Integrate operable/account_name enrichment in the chosen flow stage (post-processing or dedicated branch) with clear helper boundaries.
- [x] 3.3 Verify compatibility in no-permission, filtered query, multi-vendor, and count-query scenarios.

## 4. Verification

- [x] 4.1 Add/update tests for the new account logic helper functions (batch operable map and account name map behavior).
- [x] 4.2 Add/update tests for `ListBizSubAccountExt` to verify `operable` and `account_name` fields are correctly populated while original behavior remains compatible.
