# 错误码定义和使用

## 1. 错误码设计

项目使用 7 位错误码（21 号段 + 5 位错误码）：

```go
const (
    OK               int32 = 0        // 成功
    PermissionDenied int32 = 2030403  // 权限不足
    Unknown          int32 = 2000000  // 未知错误
    InvalidParameter int32 = 2000001  // 参数无效
    TooManyRequest   int32 = 2000002  // 请求过多
    RecordNotFound   int32 = 2000003  // 记录不存在
    DecodeRequestFailed int32 = 2000004 // 解码请求失败
    UnHealthy        int32 = 2000005  // 服务不健康
    Aborted          int32 = 2000006  // 请求中止
    DoAuthorizeFailed int32 = 2000007 // 鉴权失败
    PartialFailed    int32 = 2000008  // 部分失败
    UserNoAppAccess  int32 = 2000009  // 用户没有访问权限
    RecordNotUpdate  int32 = 2000010  // DB更新影响行数为0
    RecordDuplicated int32 = 2100011 // 数据重复
    CloudVendorError        int32 = 2100013  // 云端报错
)
```

**代码位置**：`pkg/criteria/errf/code.go`

## 2. 错误码速查表

| 错误码 | 常量名 | 含义 | 使用场景 |
|-------|-------|------|---------|
| 0 | OK | 成功 | 正常响应 |
| 2030403 | PermissionDenied | 权限不足 | IAM 鉴权失败 |
| 2000000 | Unknown | 未知错误 | 未分类的错误 |
| 2000001 | InvalidParameter | 参数无效 | 请求参数校验失败 |
| 2000002 | TooManyRequest | 请求过多 | 限流场景 |
| 2000003 | RecordNotFound | 记录不存在 | 查询无结果 |
| 2000004 | DecodeRequestFailed | 解码失败 | JSON 解析失败 |
| 2000005 | UnHealthy | 服务不健康 | 健康检查失败 |
| 2000006 | Aborted | 请求中止 | 业务处理异常 |
| 2000007 | DoAuthorizeFailed | 鉴权失败 | 鉴权接口异常 |
| 2000008 | PartialFailed | 部分失败 | 批量操作部分失败 |
| 2000009 | UserNoAppAccess | 用户没有访问权限 | 用户鉴权失败 |
| 2000010 | RecordNotUpdate | DB更新影响行数为0 | DB更新失败 |
| 2000011 | RecordDuplicated | 数据重复 | DB执行失败 |
| 2000013 | CloudVendorError | 云端报错 | 云接口失败 |

## 3. Handler 常用错误码

| 处理阶段 | 错误码 | 示例 |
|---------|--------|------|
| 请求解码 | `DecodeRequestFailed` | `cts.DecodeInto(req)` 失败 |
| 参数校验 | `InvalidParameter` | `req.Validate()` 失败 |
| 业务处理 | `Aborted` | 业务逻辑异常 |
| 记录不存在 | `RecordNotFound` | 查询结果为空 |
| 权限校验 | `PermissionDenied` | IAM 鉴权失败 |

## 4. 使用示例

### Handler 标准处理

```go
func (svc *service) GetAreaItem(cts *rest.Contexts) (interface{}, error) {
    // 1. 解码请求
    req := new(types.BizDateRangeKindOption)
    if err := cts.DecodeInto(req); err != nil {
        logs.Errorf("decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
        return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
    }

    // 2. 参数校验
    if err := req.Validate(opt); err != nil {
        logs.Errorf("validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
        return nil, errf.NewFromErr(errf.InvalidParameter, err)
    }

    // 3. 业务处理
    items, err := svc.cs.DataService().GetAreaItem(cts.Kit, req)
    if err != nil {
        logs.Errorf("get area item failed, err: %v, rid: %s", err, cts.Kit.Rid)
        return nil, errf.NewFromErr(errf.Aborted, err)
    }

    return items, nil
}
```

### 记录不存在处理

```go
data, err := svc.dao.Get(kt, id)
if err != nil {
    if errors.Is(err, orm.ErrRecordNotFound) {
        return nil, errf.New(errf.RecordNotFound, "record does not exist")
    }
    return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)
}
```

### 批量操作部分失败

```go
var failedIDs []string
for _, id := range ids {
    if err := process(id); err != nil {
        failedIDs = append(failedIDs, id)
        continue
    }
}

if len(failedIDs) > 0 {
    return errf.Newf(errf.PartialFailed, "failed ids: %v", failedIDs)
}
```

## 5. 错误码选择原则

| 场景 | 推荐错误码 |
|-----|-----------|
| 必填参数缺失 | `InvalidParameter` |
| 参数格式错误 | `InvalidParameter` |
| 参数值非法 | `InvalidParameter` |
| JSON 解析失败 | `DecodeRequestFailed` |
| 数据库查询无结果 | `RecordNotFound` |
| SQL 执行失败 | `DBExecCmdFailed` |
| 缓存操作失败 | `InvalidCache` |
| 外部服务调用失败 | `Aborted` |
| 业务逻辑异常 | `Aborted` |
| 无权限 | `PermissionDenied` |
| 限流 | `TooManyRequest` |
