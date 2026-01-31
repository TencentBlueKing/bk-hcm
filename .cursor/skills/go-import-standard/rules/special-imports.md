# 特殊导入规范

## 下划线导入（副作用导入）

使用 `_` 导入仅用于触发包的 `init()` 函数，**必须添加注释说明用途**：

```go
import (
	// import mysql driver, used to create conn.
	_ "github.com/go-sql-driver/mysql"
)
```

### 允许的下划线导入

| 包 | 用途 |
|----|-----|
| `github.com/go-sql-driver/mysql` | MySQL 驱动注册 |
| `net/http/pprof` | 性能分析端点注册 |

> **文件参考**: `pkg/dal/dao/dao.go`, `pkg/rest/handler.go`

### 使用规范

1. **必须添加注释**：说明为什么需要这个副作用导入
2. **限制使用场景**：仅用于驱动注册、性能分析等必要场景
3. **避免滥用**：不要用下划线导入来隐藏未使用的包

## 点导入（禁止）

**禁止使用点导入**（`. "package"`）。

### 为什么禁止

- 点导入会污染当前命名空间
- 降低代码可读性
- 难以追踪符号来源

### 项目符合度

- **下划线导入**: 仅 3 处使用，均有注释说明
- **点导入**: 0 处使用（完全符合规范）

## 常见错误示例

### ❌ 错误：下划线导入没有注释

```go
import (
	_ "github.com/go-sql-driver/mysql"
)
```

### ✅ 正确：下划线导入有注释说明

```go
import (
	// import mysql driver, used to create conn.
	_ "github.com/go-sql-driver/mysql"
)
```

### ❌ 错误：使用点导入

```go
import (
	. "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	Equal(t, 1, 1)  // 无法追踪 Equal 来自哪里
}
```

### ✅ 正确：使用常规导入

```go
import (
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert.Equal(t, 1, 1)  // 清晰地知道 Equal 来自 assert 包
}
```
