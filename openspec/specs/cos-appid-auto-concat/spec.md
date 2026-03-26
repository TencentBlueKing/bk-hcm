# cos-appid-auto-concat

## Requirements

### Requirement: hc-service 自动拼接 BucketName 与 APPID
hc-service 的 `CreateTCloudZiyanCosBucket` MUST 在创建存储桶前检查 BucketName 是否已包含正确的 APPID 后缀。如果未包含，MUST 通过 CAM `GetUserAppId` 接口获取 APPID 并自动拼接。

#### Scenario: BucketName 已包含正确的 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket-1250000000`，且通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统判断后缀匹配，不做额外处理，直接使用原始 BucketName 创建存储桶

#### Scenario: BucketName 未包含 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket`，通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统自动拼接为 `my-bucket-1250000000`，使用拼接后的名称创建存储桶

#### Scenario: BucketName 包含错误的 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket-9999999999`，但通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统判断后缀不匹配，自动拼接正确的 APPID，使用 `my-bucket-9999999999-1250000000` 创建存储桶

#### Scenario: 获取 APPID 失败
- **WHEN** 调用 CAM `GetUserAppId` 接口失败
- **THEN** 系统返回错误，不创建存储桶
