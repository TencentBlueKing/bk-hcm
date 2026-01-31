# 缩写约定

## 保持大写的缩写

以下缩写在命名中保持大写形式：

| 缩写 | 含义 | 正确示例 | 错误示例 |
|-----|-----|---------|---------|
| `ID` | Identifier | `UserID`, `CacheID` | `UserId`, `Cacheid` |
| `GPU` | Graphics Processing Unit | `GPUSMUsage`, `GPUTrainingUsage` | `GpuSMUsage` |
| `API` | Application Programming Interface | `APIGateway`, `APIPrefix` | `ApiGateway` |
| `URL` | Uniform Resource Locator | `BaseURL` | `BaseUrl` |
| `HTTP` | HyperText Transfer Protocol | `HTTPClient` | `HttpClient` |
| `SQL` | Structured Query Language | `SQLQuery` | `SqlQuery` |

---

## 项目通用缩写

| 缩写 | 全称 | 使用场景 | 示例 |
|-----|-----|---------|------|
| `kt` | Kit | 请求上下文 | `kt *kit.Kit` |
| `opt` | Option | 配置选项 | `opt *types.ListOption` |
| `req` | Request | 请求参数 | `req *CreateTicketReq` |
| `resp` | Response | 响应数据 | `resp *CacheResp` |
| `cfg` | Config | 配置 | `cfg *cc.Itsm` |
| `ctx` | Context | Go context | `ctx context.Context` |
| `tx` | Transaction | 数据库事务 | `tx *sqlx.Tx` |
| `err` | Error | 错误 | `err error` |

---

## 业务领域缩写

| 缩写 | 全称 | 使用场景 |
|-----|-----|---------|
| `obs` | OBS（账单系统） | `obsbill`, `ObsBill`, `ObsCostTrend` |
| `hc` | HC（公有云） | `HcRegionMeta`, `hcClient` |
| `ds` | DataService | `dsset`（包名） |
| `an` | Analysis | `anset`（包名） |
| `rel` | Relation | `ProductOrgRel`, `DeptParentRel` |
| `meta` | Metadata | `ObsCategoryMeta`, `RegionMeta` |
| `op` | Operation Product | `OpProductIDs`, `opProduct` |
| `biz` | Business | `bizService`, `biz`（包名） |

---

## 复合缩写

业务领域复合词使用小写连接：

| 复合词 | 来源 | 使用场景 |
|-------|-----|---------|
| `obsbill` | OBS + Bill | 包名 |
| `costoptim` | Cost + Optimization | `costoptimrecord` 包名 |
| `gpusmusage` | GPU + SM + Usage | 包名 |
| `dsset` | DataService + Set | 包名（客户端集合） |
| `anset` | Analysis + Set | 包名（客户端集合） |

---

## 命名示例

### 结构体字段

```go
type ObsBillReq struct {
    OpProductIDs []string `json:"op_product_ids"`  // ✅ ID 大写
    GPUType      string   `json:"gpu_type"`         // ✅ GPU 大写
    APIKey       string   `json:"api_key"`          // ✅ API 大写
}
```

### 函数参数

```go
func (s *service) GetObsBillTrend(
    kt *kit.Kit,           // ✅ kt = Kit
    opt *types.ListOption, // ✅ opt = Option
) (*types.Result, error)
```

### 局部变量

```go
func handleRequest(ctx context.Context, req *CreateReq) (*Resp, error) {
    cfg := getConfig()     // ✅ cfg = Config
    resp := &Resp{}        // ✅ resp = Response
    
    tx, err := db.Begin()  // ✅ tx = Transaction, err = Error
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()
    
    return resp, nil
}
```
