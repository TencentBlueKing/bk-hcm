# 包和文件命名

## 包命名规则

### 基本规则
- 使用**小写字母**，不使用下划线或连字符
- 目录名可用 kebab-case，但包名必须是小写连接
- 优先使用**单数名词**
- 工具类包可使用**复数形式**（参照 Go 标准库）

### ✅ 正确示例

```go
// 目录 global-config → 包名 globalconfig
// From: pkg/dal/dao/global-config/global-config.go
package daogconf

// 目录 data-service → 包名 dsset（缩写）
// From: pkg/client/data-service/client.go
package dataservice

// 工具包使用复数形式
// From: pkg/tools/times/time.go
package times

// From: pkg/tools/converter/converter.go
package converter

// From: pkg/criteria/enumor/view.go
package enumor
```

### ❌ 错误示例

```go
package obs_bill    // 错误：不应使用下划线
package DataService // 错误：不应使用大写字母
package obsService  // 错误：不应使用 camelCase
```

### 工具包复数形式例外

以下工具包使用复数形式，这是参照 Go 标准库的惯例：

| 包名 | 路径 |
|-----|-----|
| `times` | `pkg/tools/times/` |
| `maths` | `pkg/tools/maths/` |
| `strings` | `pkg/tools/strings/` |
| `maps` | `pkg/tools/maps/` |
| `sorts` | `pkg/tools/sorts/` |
| `runtimes` | `pkg/runtime/runtimes/` |

---

## 文件命名规则

### 基本规则
- 使用 **snake_case**
- 文件名应反映主要内容或资源类型

### 常用后缀

| 后缀 | 用途 | 示例 |
|-----|-----|------|
| `_test.go` | 测试文件 | `obs_bill_test.go` |
| `_types.go` | 类型定义 | `plat_types.go` |
| `_impl.go` | 实现细节 | `obs_bill_impl.go` |

### ✅ 正确示例

```
pkg/dal/dao/obs-bill/obs_bill.go         # 主实现
pkg/dal/dao/obs-bill/obs_bill_impl.go    # 扩展实现
pkg/client-set/data-service/plat_types.go # 类型定义
pkg/tools/times/time_test.go             # 测试文件
pkg/thirdparty/itsm/itsm.go              # 第三方客户端
cmd/web-server/logics/excel/parser/gpu_amount_parser.go  # 解析器
```

### 文件组织模式

```
pkg/dal/dao/obs-bill/
├── obs_bill.go           # 主文件：接口定义 + 构造函数
├── obs_bill_impl.go      # 实现文件：接口方法实现
└── obs_bill_test.go      # 测试文件
```
