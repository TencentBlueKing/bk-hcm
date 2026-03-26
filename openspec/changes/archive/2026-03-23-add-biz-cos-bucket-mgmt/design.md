## Context

当前 COS 存储桶管理仅有资源视角的代理接口（`cmd/cloud-server/service/cos/`），直接透传到 hc-service 再调 adaptor。无业务逻辑、无标签管理、无业务隔离。

项目中已有成熟的业务标签管理模式：
- **标签生成**：`cmd/cloud-server/logics/ziyan/tags.go` 的 `GenTagsForBizsWithManager` 从 CMDB/global_config 获取业务元数据生成标签
- **标签体系**：`pkg/ziyan/resource_tag.go` 定义了运营产品、一二级业务、运营部门、负责人、备份负责人等标签
- **biz 路由**：`rest.NewHandler()` + `bizH.Path("/bizs/{bk_biz_id}")` 模式（参见安全组、CLB 等实现）
- **AppID 获取**：`pkg/adaptor/tcloud/account.go` 的 `GetAccountInfoBySecret` 已调 CAM `GetUserAppId`，但未保存 AppId 返回值

COS adaptor 层（`pkg/adaptor/tcloud/cos.go`）已支持：
- `CreateBucket`：通过 `XCosTagging` 字段支持创建时打标签
- `ListBuckets`：通过 `TagKey`/`TagValue` 字段支持按标签过滤
- `ClientOpt.BucketNameAppID`：接受 `name-APPID` 格式的存储桶名称

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 实现业务视角的 COS 创建/删除/列表三个接口
- 通过腾讯云标签实现业务归属管理，复用自研云标签体系
- 在 hc-service 实现 BucketName 与 APPID 的智能拼接
- 复用现有的 adaptor 层接口，不改变原有接口语意

**Non-Goals:**
- 不写入本地 DB，不做 COS 资源同步
- 不涉及前端 UI 变更
- 不支持 TCloud（公有云）厂商，仅支持 TCloudZiyan（自研云）
- 不实现 COS 对象（Object）级别的操作

## Decisions

### 1. 业务逻辑分层：cloud-server 处理业务逻辑，hc-service 处理 adaptor 逻辑

**选择**：标签生成、归属校验等业务逻辑放在 cloud-server；AppID 拼接逻辑放在 hc-service。

**理由**：
- cloud-server 是业务编排层，能访问 DataService 和 CMDB 获取业务元数据，适合处理标签生成
- hc-service 是资源操作层，直接持有 adaptor 实例，适合处理 AppID 获取和 BucketName 拼接
- 与安全组等资源的现有模式一致（`createTCloudZiyanSecurityGroup` 在 cloud-server 生成标签）

**备选**：全部放 cloud-server → 但 AppID 需要调 CAM 接口，cloud-server 不直接持有 adaptor，需额外透传

### 2. 标签转 XCosTagging 格式

**选择**：在 cloud-server 的 biz create handler 中，将 `[]apicore.TagPair` 转为 COS SDK 要求的 `XCosTagging` HTTP Header 格式。

**格式说明**：COS 的 `x-cos-tagging` header 格式为 URL-encoded key=value 对，用 `&` 分隔。例如：
```
key1=value1&key2=value2
```

由于标签 key 包含中文（如"二级业务"），需要进行 URL 编码。

**实现**：在 cloud-server 新增 `tagsToXCosTagging(tags []apicore.TagPair) string` 工具函数，遍历 TagPair 将 Key 和 Value 进行 URL 编码后拼接。

### 3. 删除前的归属校验策略

**选择**：在 cloud-server 的 biz delete handler 中，先调 hc-service 的 ListBuckets（带业务标签过滤），检查目标 bucket 是否在返回列表中。

**流程**：
```
cloud-server DeleteBizCosBucket:
  1. 根据 bk_biz_id 获取业务元数据 → 生成二级业务标签
  2. 调 hc-service ListBuckets(TagKey="二级业务", TagValue="xxx_123", Region=req.Region)
  3. 遍历返回的 bucket 列表，匹配 req.Name
  4. 匹配成功 → 调 hc-service DeleteBucket
  5. 匹配失败 → 返回错误
```

**备选**：先获取单个 bucket 的标签再比对 → COS SDK 无直接获取单个 bucket 标签的 Service-level API，需用 Bucket-level API（需额外构造 BucketURL），不如 list 方案简洁

### 4. AppID 自动拼接逻辑位置与实现

**选择**：在 hc-service 的 `CreateTCloudZiyanCosBucket` 中实现。

**算法**：
```
1. 获取 adaptor 实例后，调 tcloud.GetAccountInfoBySecret(kt) 获取 AppId
2. appIdStr = strconv.FormatInt(appId, 10)
3. 如果 req.Name 以 "-" + appIdStr 结尾 → 已拼接，不处理
4. 否则 → req.Name = req.Name + "-" + appIdStr
5. 继续原有创建流程
```

**理由**：hc-service 直接持有 adaptor 实例，能调 `GetAccountInfoBySecret` 获取 AppId，无需额外的 RPC 调用。

### 5. biz 路由注册方式

**选择**：在 `cmd/cloud-server/service/cos/cos.go` 的 `InitService` 中新增 `bizH` handler，与资源路由并存。

**实现**：
```go
bizH := rest.NewHandler()
bizH.Path("/bizs/{bk_biz_id}")
bizH.Add("CreateBizCosBucket", http.MethodPost, "/cos/buckets/create", svc.CreateBizCosBucket)
bizH.Add("DeleteBizCosBucket", http.MethodDelete, "/cos/buckets/delete", svc.DeleteBizCosBucket)
bizH.Add("ListBizCosBucket", http.MethodPost, "/cos/buckets/list", svc.ListBizCosBucket)
bizH.Load(c.WebService)
```

与 load-balancer、security-group 等服务的 biz 路由注册模式一致。

### 6. 请求结构体设计

**选择**：在 `pkg/api/cloud-server/` 下新增 COS biz 请求结构体，与 hc-service 的请求结构体分离。

**新增结构体**：
- `TCloudBizCreateCosBucketReq`：包含 `account_id`、`region`、`name`、`manager`、`bak_manager` 及原有的 ACL/Configuration 字段
- `TCloudBizDeleteCosBucketReq`：包含 `account_id`、`region`、`name`
- `TCloudBizListCosBucketReq`：包含 `account_id`、`region` 及可选的分页参数

**理由**：biz 接口的 create 请求需要额外的 `manager`/`bak_manager` 字段，不能直接复用 hc-service 的 `TCloudCreateBucketReq`。delete 和 list 的结构虽然相似，但为保持一致性和独立演进能力，单独定义。

### 7. List 标签过滤使用二级业务标签

**选择**：使用 `TagKey="二级业务"` + `TagValue="业务名_业务ID"` 作为 ListBuckets 的过滤条件。

**理由**：
- 二级业务标签是自研云资源归属的核心标识，一个业务 ID 对应唯一的二级业务标签
- `ResourceMeta.GetSyncFilterTag()` 已封装了该逻辑，可直接复用
- 腾讯云 ListBuckets API 的 `TagKey`/`TagValue` 参数天然支持此过滤方式

## Risks / Trade-offs

### [风险] 标签被人为修改导致业务归属错乱
→ **缓解**：COS 标签管理依赖腾讯云 API，云上标签被修改后会影响本系统的 list/delete 判断。当前不额外处理，与其他自研云资源（安全组、CLB）的风险一致。

### [风险] ListBuckets 按标签过滤的性能
→ **缓解**：腾讯云 ListBuckets 单次最多返回 2000 条，支持分页。按标签过滤在腾讯云侧完成，不影响本系统性能。

### [风险] AppID 获取增加一次 CAM API 调用
→ **缓解**：仅在 hc-service 的 create 流程中增加一次 `GetUserAppId` 调用，delete/list 不受影响。后续可考虑缓存 AppId。

### [Trade-off] 不写本地 DB
→ **优势**：不增加存储和同步复杂度，COS 资源以腾讯云为唯一数据源
→ **劣势**：无法支持本地搜索、高级筛选等功能，每次查询都走腾讯云 API

### [Trade-off] 修改 TCloudInfoBySecret 结构体
→ 该结构体被多处引用（账号校验、账号创建等流程），新增字段使用默认零值不影响已有调用方。但需确认所有使用 `GetAccountInfoBySecret` 的地方不会因多返回一个字段而出现问题。
