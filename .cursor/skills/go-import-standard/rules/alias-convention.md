# 别名命名规范

## 何时使用别名

在以下情况**必须**使用导入别名：

| 场景 | 说明 | 示例 |
|-----|------|-----|
| **路径含连字符** | 目录名包含 `-`，与 Go 包名不一致 | `client-set` → `clientset` |
| **包名冲突** | 两个包的默认名称相同 | `ctypes` vs `cstypes` |
| **语义不清晰** | 默认包名过于通用或不明确 | `types` → `ctypes` |
| **简化长包名** | 包路径层级过深导致默认名过长 | `cost-optimization-ticket` → `costoptimticket` |

## 内部包别名规范

| 模式 | 命名规范 | 示例 |
|------|---------|-----|
| **client-set 子包** | 使用缩写 + `set` 后缀 | `dsset`, `accset`, `hcset`, `tsset`, `webset` |
| **types 包** | 使用前缀 + `types` | `ctypes`, `cstypes`, `daotypes` |
| **dao 子包** | 去掉连字符，使用驼峰缩写 | `obsbill`, `costoptimticket`, `idgenerator` |
| **audit 包** | 实体名 + `audit` | `obsaudit`, `prodaudit` |
| **thirdparty 子包** | 使用简短缩写 | `obscli`, `dmmg`, `iegatp` |

## 常用别名对照表

### client-set 相关

```go
clientset   "hcm/pkg/client"
accset       "hcm/pkg/client/account-server"
dsset       "hcm/pkg/client/data-service"
hcset       "hcm/pkg/client/hc-service"
tsset       "hcm/pkg/client/task-server"
apiset       "hcm/pkg/client/api-server"
authset     "hcm/pkg/client/auth-server"
webset      "hcm/pkg/client/web-server"
```

### types 相关

```go
cstypes     "hcm/pkg/client/types"    // 推荐使用
ctypes      "hcm/pkg/client/types"    // 历史遗留
daotypes    "hcm/pkg/dal/dao/types"
```

> **注意**: `cstypes` 和 `ctypes` 都指向同一个包，`cstypes` 使用更广泛（277 次 vs 226 次），建议新代码统一使用 `cstypes`

### dao 相关

```go
idgenerator "hcm/pkg/dal/dao/id-generator"   // 完整形式（更常用，70次）
idgen       "hcm/pkg/dal/dao/id-generator"   // 简写形式（31次）
```

### thirdparty 相关

```go
obscli      "hcm/pkg/thirdparty/obs"
accreport   "hcm/pkg/thirdparty/obs/account-report"
iegatp      "hcm/pkg/thirdparty/ieg-atp"
```

> **文件参考**: `pkg/dal/dao/dao.go`, `cmd/etl-service/service/service.go`

## 第三方包别名规范

第三方包一般**不使用别名**，除非：
- 包名与项目内部包冲突
- 包名过于通用需要区分
- 官方文档推荐使用别名

### 常用第三方包别名

```go
// Prometheus
prm         "github.com/prometheus/client_golang/prometheus"

// Validator
gvalidator  "github.com/go-playground/validator/v10"
ut          "github.com/go-playground/universal-translator"
zhTranslations "github.com/go-playground/validator/v10/translations/zh"

// JSON
jsoniter    "github.com/json-iterator/go"

// OpenTelemetry
sdktrace    "go.opentelemetry.io/otel/sdk/trace"
semconv     "go.opentelemetry.io/otel/semconv/v1.26.0"

// 腾讯云 SDK
cos         "github.com/tencentyun/cos-go-sdk-v5"
sts         "github.com/tencentyun/qcloud-cos-sts-sdk/go"

// UUID (用于性能对比测试)
guuid       "github.com/google/uuid"
puuid       "github.com/pborman/uuid"
```

> **文件参考**: `pkg/traces/traces.go`, `pkg/tools/uuid/uuid_test.go`

## 别名使用统计

| 类别 | 数量 | 占比 |
|-----|------|-----|
| 使用别名的内部包导入 | ~2,064 | ~18% |
| 不使用别名的内部包导入 | ~9,442 | ~82% |
| 第三方包使用别名 | ~17 | <1% |

> **结论**: 仅在必要时使用别名，大多数情况下使用默认包名
