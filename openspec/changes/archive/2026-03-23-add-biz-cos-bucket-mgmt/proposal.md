## Why

当前 COS 存储桶管理仅提供资源视角的代理接口，缺少业务逻辑，不支持按业务维度管理。自研云场景下需要 COS 资源像 CLB、安全组等资源一样，支持业务归属管理，实现业务隔离和权限控制。

## What Changes

- 在 cloud-server 新增业务视角的 COS 接口（`/bizs/{bk_biz_id}/cos/buckets/{create,delete,list}`），主要业务逻辑在 cloud-server 层
- 创建存储桶时，根据业务 ID、负责人、备份负责人生成自研云标签，通过 `XCosTagging` 传递给腾讯云
- 删除存储桶时，先按业务标签查询腾讯云的存储桶列表，校验存储桶确实属于当前业务后才允许删除
- 查询存储桶列表时，按业务 ID 生成标签过滤条件，仅返回当前业务的存储桶，防止越权
- COS 资源不写入本地 DB，完全通过腾讯云标签实现业务归属
- hc-service 的 `CreateTCloudZiyanCosBucket` 新增 AppID 自动拼接逻辑：判断 BucketName 后缀是否已包含正确的 APPID，未包含则自动拼接
- adaptor 层 `GetAccountInfoBySecret` 补充返回 `AppId` 字段（已有 CAM 调用，当前未保存该值）

## Capabilities

### New Capabilities
- `biz-cos-bucket-create`: 业务视角下创建 COS 存储桶，包含业务标签生成与标签注入
- `biz-cos-bucket-delete`: 业务视角下删除 COS 存储桶，包含业务归属校验
- `biz-cos-bucket-list`: 业务视角下查询 COS 存储桶列表，按业务标签过滤
- `cos-appid-auto-concat`: hc-service 层 BucketName 自动拼接 APPID 逻辑

### Modified Capabilities
- `tcloud-account-info`: `GetAccountInfoBySecret` 返回结构新增 `AppId` 字段

## Impact

- **cloud-server**：`cmd/cloud-server/service/cos/` 新增 biz 路由和 handler 文件
- **hc-service**：`cmd/hc-service/service/cos/tcloud_ziyan.go` 修改 `CreateTCloudZiyanCosBucket` 增加 AppID 拼接逻辑
- **adaptor**：`pkg/adaptor/tcloud/account.go` 修改 `GetAccountInfoBySecret` 返回值
- **API 类型**：`pkg/api/core/cloud/account_by_secret.go` 修改 `TCloudInfoBySecret`；`pkg/api/cloud-server/` 或 `pkg/api/hc-service/cos/` 新增业务级请求结构体
- **接口文档**：已有 `docs/api-docs/api-server/docs/zh/` 下的三个接口文档
- **不涉及 DB 变更**：COS 资源不写本地 DB
