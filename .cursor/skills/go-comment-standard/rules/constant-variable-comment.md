# 常量和变量注释规范

## 常量注释

### 规则

1. 以常量名开头
2. 常量组内每个常量单独注释
3. 类型定义前添加注释

### 单个常量

```go
// MaxRetries is the maximum number of retry attempts.
const MaxRetries = 3
```

### 常量组

```go
const (
	// ApiPrefix is the prefix of all API.
	ApiPrefix = "api"
	// ApiVersion is the version of current API.
	ApiVersion = "version"
)
```

### 代码示例

```63:98:pkg/criteria/constant/clb.go
const (
	// CLBFilePrefix 负载均衡文件名前缀
	CLBFilePrefix = "hcm-clb"
	// Layer4ListenerFilePrefix 四层监听器文件名前缀
	Layer4ListenerFilePrefix = "tcp_udp监听器"
	// Layer7ListenerFilePrefix 七层监听器文件名前缀
	Layer7ListenerFilePrefix = "http_https监听器"
	// RuleFilePrefix 规则文件名前缀
	RuleFilePrefix = "http_https规则URL"
	// Layer4RsFilePrefix 四层RS文件名前缀
	Layer4RsFilePrefix = "tcp_udp绑定的RS"
	// Layer7RsFilePrefix 七层RS文件名前缀
	Layer7RsFilePrefix = "http_https绑定的RS"
	// Layer4ListenerSheetName 四层监听器sheet名
	Layer4ListenerSheetName = "批量创建监听器-TCP-UDP"
	// Layer7ListenerSheetName 七层监听器sheet名
	Layer7ListenerSheetName = "批量创建监听器-HTTP-HTTPS"
	// RuleSheetName 规则sheet名
	RuleSheetName = "批量创建URL规则-HTTP-HTTPS"
	// Layer4RsSheetName 四层RS sheet名
	Layer4RsSheetName = "绑定RS-TCP-UDP"
	// Layer7RsSheetName 七层RS sheet名
	Layer7RsSheetName = "绑定RS-HTTP-HTTPS"

	// CLBExcelHeaderVendor excel表头云厂商字段值
	CLBExcelHeaderVendor = "vendor(云厂商)"
	// CLBExcelHeaderTCloud excel表头腾讯云字段值
	CLBExcelHeaderTCloud = "tencent_cloud_public(腾讯云-公有云)"
)
```

## 类型常量（枚举）

### 规则

1. 类型定义前添加注释（如 `// TypeName is description.`）
2. 每个枚举值单独注释
3. 相关常量使用 `const ()` 分组

### 标准格式

```go
// TypeName is the type description.
type TypeName string

const (
	// TypeNameValue1 is description.
	TypeNameValue1 TypeName = "value1"
	// TypeNameValue2 is description.
	TypeNameValue2 TypeName = "value2"
)
```

### 代码示例

```26:41:pkg/criteria/enumor/account.go
// AccountType is account type.
type AccountType string

// Validate the AccountType is valid or not
func (a AccountType) Validate() error {
	switch a {
	case ResourceAccount:
	case RegistrationAccount:
	case SecurityAuditAccount:
	default:
		return fmt.Errorf("unsupported account type: %s", a)
	}

	return nil
}
```

```20:32:pkg/criteria/constant/tenant.go
// Package constant constant 多租户相关的常量
package constant

const (
	// DefaultTenantID 默认的租户id，使用场景：兼容不开启多租户的场景，上下游调用默认传递default租户
	DefaultTenantID = "default"
	// SystemTenantID 运营租户id
	SystemTenantID = "system"
	// TenantIDField 租户id字段
	TenantIDField = "tenant_id"
	// TenantIDTableField 租户id对应的table里的字段
	TenantIDTableField = "TenantID"
)
```

## 变量注释

### 规则

1. 以变量名开头
2. 说明变量的数据内容或用途

### 单个变量

```go
// defaultTimeout is the default HTTP request timeout.
var defaultTimeout = 30 * time.Second
```

### 变量组

```go
var (
	// YearTimeGrans 年时间粒度列表
	YearTimeGrans = []TimeGran{TimeGranYear}
	// MonthTimeGrans 月时间粒度列表
	MonthTimeGrans = []TimeGran{TimeGranMonth}
)
```

### 代码示例

```26:39:pkg/criteria/constant/time.go
const (
	// TimeStdFormat is the system's standard time format to store or to query, equal to time.RFC3339
	TimeStdFormat = "2006-01-02T15:04:05Z07:00"
	// DateLayout is the date layout with '%Y-%m-%d'
	DateLayout = "2006-01-02"
	// DateTimeLayout is the date layout with '%Y-%m-%d %H:%M:%S'
	DateTimeLayout = "2006-01-02 15:04:05"
)

// TimeStdRegexp is a regular expression to match the TimeStdFormat(RFC3339) with millisecond
var TimeStdRegexp = regexp.
MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.\d{3})?[Z+]([0-9]{2}:[0-9]{2})*$`)
```

## 映射表变量

### 规则

1. 说明映射关系
2. 业务数据值可使用中文

### 代码示例

```go
// objectTypeNameMap is the map from DetObjectType to its name.
var objectTypeNameMap = map[DetObjectType]string{
	DetObjectTypeBiz:       "业务",
	DetObjectTypeOpProduct: "运营产品",
	DetObjectTypeDept:      "部门",
}

// DiskTypeNameMap disk type name map.
var (
	DiskTypeNameMap = map[typecvm.AwsVolumeType]string{
		typecvm.GP3:      "通用型SSD卷(gp3)",
		typecvm.GP2:      "通用型SSD卷(gp2)",
		typecvm.IO1:      "预置IOPS SSD卷(io1)",
		typecvm.IO2:      "预置IOPS SSD卷(io2)",
		typecvm.ST1:      "吞吐量优化型HDD卷(st1)",
		typecvm.SC1:      "Cold HDD卷(sc1)",
		typecvm.Standard: "上一代磁介质卷(standard)",
	}
)
```

## 私有常量/变量

### 规则

1. 私有常量/变量同样建议添加注释
2. 简单明确的私有常量可省略注释

### 代码示例

```go
const (
	// defaultPageSize is the default pagination size.
	defaultPageSize = 20
	// maxPageSize is the maximum allowed pagination size.
	maxPageSize = 1000
)

var (
	// instance is the singleton instance.
	instance *Client
	// once ensures instance is created only once.
	once sync.Once
)
```

## 表名常量

### 规则

1. 使用 `*Table` 后缀命名
2. 说明对应的数据库表

### 代码示例

```go
const (
	// OpProductTable is the database table name for operational products.
	OpProductTable = "op_product"
	// UserTable is the database table name for users.
	UserTable = "user"
)
```
