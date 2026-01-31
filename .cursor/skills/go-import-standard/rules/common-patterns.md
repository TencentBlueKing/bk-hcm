# 常见导入模式

本文档展示项目中不同类型文件的典型导入模式。

## Service 层导入模式

```go
import (
	"net/http"

	"hcm/cmd/analysis-server/options"
	"hcm/cmd/analysis-server/service/logics"
	clientset "hcm/pkg/client"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/iris"
)
```

### 特点
- 引入 cmd 下同服务的 options 和 logics
- 使用 `clientset` 别名导入客户端集合
- 引入 HTTP 相关工具包

## DAO 层导入模式

```go
import (
	"time"

	dsset "hcm/pkg/client/data-service"
	ctypes "hcm/pkg/client/types"
	"hcm/pkg/criteria/enumor"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/share"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"

	"github.com/bluele/gcache"
	"github.com/jmoiron/sqlx"
)
```

### 特点
- 大量使用别名（`dsset`, `ctypes`, `idgen`）
- 引入 DAO 层内部共享包（`orm`, `share`, `types`）
- 引入 `filter` 用于查询条件构建
- 使用 `sqlx` 和 `gcache` 第三方库

## Handler 层导入模式

```go
import (
	"net/http"

	"hcm/cmd/data-service/service/dao"
	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"

	restful "github.com/emicklei/go-restful/v3"
)
```

### 特点
- 引入 HTTP 框架 `go-restful`
- 引入日志、配置、监控等基础设施包
- 引入同服务的 DAO 层

## 工具包导入模式

```go
// pkg/tools/times/time.go
import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
)
```

### 特点
- 大量标准库导入
- 少量内部基础包（constant, errf）
- 通常不依赖第三方包

## 测试文件导入模式

```go
import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

### 特点
- 引入 `testing` 标准库
- 使用 `testify` 的 `assert` 和 `require` 包
- 可能引入 `io`, `os`, `bytes` 等用于测试数据准备

## 常用第三方库

| 库 | 用途 | 导入语句 |
|----|-----|---------|
| sqlx | SQL 扩展 | `"github.com/jmoiron/sqlx"` |
| excelize | Excel 处理 | `"github.com/xuri/excelize/v2"` |
| go-restful | REST 框架 | `"github.com/emicklei/go-restful/v3"` |
| testify | 测试断言 | `"github.com/stretchr/testify/assert"` |
| prometheus | 监控指标 | `"github.com/prometheus/client_golang/prometheus"` |
| gcache | 缓存 | `"github.com/bluele/gcache"` |
| elastic | Elasticsearch | `"gopkg.in/olivere/elastic.v6"` |
| jsonschema | JSON Schema | `"github.com/invopop/jsonschema"` |
| cron | 定时任务 | `"github.com/robfig/cron/v3"` |

## 常用内部包

```go
// 基础设施
"hcm/pkg/cc"                    // 配置
"hcm/pkg/logs"                  // 日志
"hcm/pkg/kit"                   // 请求上下文
"hcm/pkg/metrics"               // 监控指标
"hcm/pkg/serviced"              // 服务发现

// 错误与常量
"hcm/pkg/criteria/constant"     // 常量定义
"hcm/pkg/criteria/errf"         // 错误定义
"hcm/pkg/criteria/enumor"       // 枚举定义

// HTTP
"hcm/pkg/rest"             // REST 客户端/服务
"hcm/pkg/handler"          // HTTP 处理器

// 数据层
"hcm/pkg/dal/dao"               // DAO 接口
"hcm/pkg/dal/table"             // 表结构定义
"hcm/pkg/runtime/filter"        // 过滤器

// 工具函数
"hcm/pkg/tools/times"           // 时间处理
"hcm/pkg/tools/converter"            // 类型转换
"hcm/pkg/tools/xlsx"            // Excel 处理
"hcm/pkg/tools/uuid"            // UUID 生成
"hcm/pkg/tools/ssl"             // SSL/TLS
```
