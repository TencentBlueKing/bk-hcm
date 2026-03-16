# HTTP/API 错误处理规范

## 1. 统一 HTTP 响应结构

API 响应使用 `rest.Response` 结构，包含 code、message、data 字段。

```go
// pkg/http/rest/response.go
type Response struct {
    Result      bool                `json:"result"`
    Code        int32               `json:"code"`
    Message     string              `json:"message"`
    Permissions *meta.IamPermission `json:"permission,omitempty"`
    Data        interface{}         `json:"data"`
}
```

## 2. Handler 函数签名

**规则**：Handler 函数统一返回 `(interface{}, error)`，框架自动处理响应。

```go
func handler(cts *rest.Contexts) (interface{}, error)
```

## 3. 标准 Handler 错误处理模式

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

## 4. 第三方 API 响应检查

**规则**：调用第三方 API 后需要检查响应状态。

```go
resp := new(BaseResp[*ListAccountRst])
err := h.client.Post().
    SubResourcef("/cloud/accounts/list").
    WithContext(ctx).
    Body(req).
    Do().Into(resp)
if err != nil {
    logs.Errorf("list hcm account failed, req: %+v, err: %v", req, err)
    return nil, err
}
if resp.IsFailed() {
    logs.Errorf("list hcm account failed, req: %+v, resp: %+v", req, resp)
    return nil, fmt.Errorf("list account failed, code: %d, msg: %s", resp.Code, resp.Message)
}
return resp, nil
```

**代码示例**（`pkg/thirdparty/api-gateway/cmdb/cmdb.go:167-177`）：

```go
// FindHostTopoRelation ...
func (c *cmdbApiGateWay) FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (
*HostTopoRelationResult, error) {

    if err := params.Validate(); err != nil {
        return nil, err
    }

    return apigateway.ApiGatewayCall[FindHostTopoRelationParams, HostTopoRelationResult](c.client, c.config, rest.POST,
    kt, params, "/host/topo/relation/read")
}
```

## 5. 错误响应格式

成功响应：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": { ... }
}
```

错误响应：

```json
{
    "result": false,
    "code": 2100001,
    "message": "invalid parameter: name is required",
    "data": null
}
```

## 6. HTTP 状态码映射

| 业务错误码 | HTTP 状态码 | 场景 |
|-----------|------------|------|
| OK (0) | 200 | 成功 |
| InvalidParameter | 400 | 参数错误 |
| PermissionDenied | 403 | 权限不足 |
| RecordNotFound | 404 | 资源不存在 |
| TooManyRequest | 429 | 限流 |
| Aborted | 500 | 服务端错误 |
| DBExecCmdFailed | 500 | 数据库错误 |

## 7. 自定义错误类型

### 业务错误类型 ErrorF

```go
type ErrorF struct {
    Code    int32                  `json:"code"`
    Message string                 `json:"message"`
    Values  map[string]interface{} `json:"values"`
}

func (e *ErrorF) Error() string {
    return fmt.Sprintf(`{"code": %d, "message": "%s"}`, e.Code, e.Message)
}
```

### 领域特定错误类型

```go
// 解析错误 - pkg/criteria/errf/error.go
// ErrorF defines an error with error code and message.
type ErrorF struct {
    // Code is hcm errCode
    Code int32 `json:"code"`
    // Message is error detail
    Message string `json:"message"`
    // Permissions is no permission error related permission.
    Permissions *meta.IamPermission `json:"permission,omitempty"`
}

// 认证错误 - pkg/iam/client/types.go
type AuthError struct {
    RequestID string
    Reason    error
}

// API 错误 - pkg/thirdparty/api-gateway/cmdb/types.go
// SearchBizResp is cmdb search business response.
type SearchBizResp struct {
    types.BaseResponse
    SearchBizResult `json:"data"`
}

// SearchBizResult is cmdb search business response.
type SearchBizResult struct {
    Count int64 `json:"count"`
    Info  []Biz `json:"info"`
}

// Biz is cmdb biz info.
type Biz struct {
    BizID   int64  `json:"bk_biz_id"`
    BizName string `json:"bk_biz_name"`
}
```

## 最佳实践

### DO

```go
// ✓ 统一 Handler 签名
func (svc *service) Handler(cts *rest.Contexts) (interface{}, error)

// ✓ 检查第三方 API 响应
if resp.IsFailed() {
    return nil, fmt.Errorf("api failed: %s", resp.Message)
}

// ✓ 分层处理
// Handler: 解码 → 校验 → 业务 → 返回
```

### DON'T

```go
// ✗ 直接写 HTTP 响应
func (svc *service) Handler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(500)
    w.Write([]byte("error"))
}

// ✗ 忽略第三方 API 响应状态
resp, err := client.Call()
if err == nil {
    return resp.Data, nil  // 没检查 IsFailed
}

// ✗ 暴露内部错误给客户端
return nil, fmt.Errorf("mysql error: %v", err)  // 不应暴露数据库错误
```
