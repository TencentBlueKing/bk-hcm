## ADDED Requirements

### Requirement: Get TCloud access key last used info via hc-service
The system SHALL provide an HTTP POST endpoint to query the last usage information for a list of TCloud access key IDs. The request MUST include `account_id` and `secret_id_list` (1-10 items). The response SHALL return an array of objects containing `secret_id`, `last_used_date`, and `last_secret_used_date`.

#### Scenario: Successful query
- **WHEN** caller sends POST to `/vendors/tcloud/sub_accounts/secrets/last_used` with valid `account_id` and `secret_id_list` containing 1-10 items
- **THEN** system queries TCloud CAM GetSecurityLastUsed API and returns the last usage information for each key

#### Scenario: Empty or oversized secret_id_list
- **WHEN** caller sends request with `secret_id_list` containing 0 or more than 10 items
- **THEN** system returns `InvalidParameter` error
