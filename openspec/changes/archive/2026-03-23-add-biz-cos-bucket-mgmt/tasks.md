## 1. API 类型与 Adaptor 层变更（基础依赖）

- [x] 1.1 在 `pkg/api/core/cloud/account_by_secret.go` 的 `TCloudInfoBySecret` 结构体中新增 `AppId int64` 字段
- [x] 1.2 修改 `pkg/adaptor/tcloud/account.go` 的 `GetAccountInfoBySecret` 方法，从 `resp.Response.AppId` 中读取 AppId 并赋值到返回结构体，增加 AppId 为空的错误校验
- [x] 1.3 在 `pkg/api/cloud-server/` 下新增 COS biz 请求结构体文件，定义 `TCloudBizCreateCosBucketReq`（含 account_id、region、name、manager、bak_manager 及 ACL/Configuration 可选字段）、`TCloudBizDeleteCosBucketReq`（含 account_id、region、name）、`TCloudBizListCosBucketReq`（含 account_id、region 及可选的 max_keys、marker、range、create_time 分页参数），编写 Validate 方法

## 2. hc-service 层变更（AppID 自动拼接）

- [x] 2.1 修改 `cmd/hc-service/service/cos/tcloud_ziyan.go` 的 `CreateTCloudZiyanCosBucket` 方法：在构造 adaptor option 前，调用 `tcloud.GetAccountInfoBySecret(kt)` 获取 AppId，将 AppId 转为字符串
- [x] 2.2 在 `CreateTCloudZiyanCosBucket` 中实现 BucketName 后缀判断逻辑：如果 `req.Name` 以 `-{appIdStr}` 结尾则不处理，否则拼接 `-{appIdStr}` 到 `req.Name`，使用拼接后的名称传入 adaptor

## 3. cloud-server 层变更（biz 路由与标签工具）

- [x] 3.1 在 `cmd/cloud-server/service/cos/cos.go` 的 `InitService` 中新增 biz 路由注册：创建 `bizH := rest.NewHandler()`，设置 `bizH.Path("/bizs/{bk_biz_id}")`，添加 CreateBizCosBucket、DeleteBizCosBucket、ListBizCosBucket 三个路由，调用 `bizH.Load(c.WebService)`
- [x] 3.2 新增标签转换工具函数 `tagsToXCosTagging(tags []apicore.TagPair) string`：遍历 TagPair，对 Key 和 Value 进行 URL 编码，用 `key=value` 格式拼接，多个标签间用 `&` 分隔

## 4. cloud-server 业务 Handler 实现

- [x] 4.1 实现 `CreateBizCosBucket` handler：解码请求为 `TCloudBizCreateCosBucketReq` → 从 URL 提取 `bk_biz_id` → IAM 鉴权 → 获取账号信息并校验厂商为 TCloudZiyan → 调用 `GenTagsForBizsWithManager(kt, dataCli, cmdbCli, bizID, manager, bakManager)` 生成业务标签 → 调用 `tagsToXCosTagging` 转换标签格式注入 `XCosTagging` 字段 → 构造 hc-service 请求调用 `HCService().TCloudZiyan.Cos.CreateCosBucket`
- [x] 4.2 实现 `ListBizCosBucket` handler：解码请求为 `TCloudBizListCosBucketReq` → 从 URL 提取 `bk_biz_id` → IAM 鉴权 → 获取账号信息并校验厂商 → 调用 `ziyan.GetResourceMetaByBiz(kt, dataCli, cmdbCli, bizID)` 获取业务元数据 → 调用 `meta.GetSyncFilterTag()` 获取二级业务标签 → 将 TagKey/TagValue 注入 hc-service 的 list 请求 → 调用 `HCService().TCloudZiyan.Cos.ListCosBucket` 返回结果
- [x] 4.3 实现 `DeleteBizCosBucket` handler：解码请求为 `TCloudBizDeleteCosBucketReq` → 从 URL 提取 `bk_biz_id` → IAM 鉴权 → 获取账号信息并校验厂商 → 先调用与 4.2 相同的标签过滤 list 逻辑获取当前业务的 bucket 列表 → 遍历匹配 `req.Name` → 匹配成功调用 `HCService().TCloudZiyan.Cos.DeleteCosBucket`，匹配失败返回错误

## 5. hc-service Client 补充（如需要）

- [x] 5.1 检查 `pkg/client/hc-service/tcloud-ziyan/cos.go` 中是否已有 CreateCosBucket、DeleteCosBucket、ListCosBucket 的 client 方法，如缺少则补充，确保 cloud-server 能通过 `HCService().TCloudZiyan.Cos.*` 调用 hc-service
