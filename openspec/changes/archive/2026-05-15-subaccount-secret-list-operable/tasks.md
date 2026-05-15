## 1. API types (cloud-server)

- [x] 1.1 Add `BizSubAccountSecretJoinExtDetail` and `BizSubAccountSecretJoinExtListResult` in `pkg/api/cloud-server/sub-account-secret/sub_account_secret.go`

## 2. Service layer

- [x] 2.1 Wire `ListSubAccountSecret` TCloud path to convert data-service join result using `logicaccount.BatchBuildOperableAndNameMap` and encapsulated helpers in `cmd/cloud-server/service/subaccount-secret/list.go`

## 3. Documentation

- [x] 3.1 Align `operable` description in `docs/api-docs/web-server/docs/biz/sub_account_secret/list_sub_account_secret.md` with secret semantics
