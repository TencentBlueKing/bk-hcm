## ADDED Requirements

### Requirement: Create TCloud access key via hc-service
The system SHALL provide an HTTP POST endpoint to create an access key for a specified TCloud CAM sub-user. The request MUST include `account_id` and `target_uin`. An optional `description` field MAY be provided. The response SHALL return the `AccessKeyId`, `SecretAccessKey`, `Status`, and `CreateTime`.

#### Scenario: Successful access key creation
- **WHEN** caller sends POST to `/vendors/tcloud/sub_accounts/secrets/create` with valid `account_id` and `target_uin`
- **THEN** system creates an access key via TCloud CAM CreateAccessKey API and returns the key details including `AccessKeyId` and `SecretAccessKey`

#### Scenario: Invalid parameters
- **WHEN** caller sends request with missing `account_id` or `target_uin`
- **THEN** system returns `InvalidParameter` error

#### Scenario: Access key limit exceeded
- **WHEN** the target sub-user already has two access keys
- **THEN** system returns the TCloud error `OperationDenied.AccessKeyOverLimit`

### Requirement: Delete TCloud access key via hc-service
The system SHALL provide an HTTP POST endpoint to delete an access key for a specified TCloud CAM sub-user. The request MUST include `account_id`, `access_key_id`, and `target_uin`.

#### Scenario: Successful access key deletion
- **WHEN** caller sends POST to `/vendors/tcloud/sub_accounts/secrets/delete` with valid `account_id`, `access_key_id`, and `target_uin`
- **THEN** system deletes the access key via TCloud CAM DeleteAccessKey API and returns success

#### Scenario: Access key not found
- **WHEN** the specified `access_key_id` does not exist
- **THEN** system returns the TCloud error `ResourceNotFound.SecretNotExist`

### Requirement: Update TCloud access key status via hc-service
The system SHALL provide an HTTP POST endpoint to update (activate/deactivate) an access key for a specified TCloud CAM sub-user. The request MUST include `account_id`, `access_key_id`, `status` (Active or Inactive), and `target_uin`.

#### Scenario: Successful access key status update
- **WHEN** caller sends POST to `/vendors/tcloud/sub_accounts/secrets/update` with valid parameters and `status` set to `Inactive`
- **THEN** system updates the access key status via TCloud CAM UpdateAccessKey API and returns success

#### Scenario: Invalid status value
- **WHEN** caller sends request with `status` not in `[Active, Inactive]`
- **THEN** system returns `InvalidParameter` error
