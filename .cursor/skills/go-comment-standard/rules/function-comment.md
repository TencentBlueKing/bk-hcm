# 函数和方法注释规范

## 导出函数

### 规则

1. 注释**必须**以函数名开头
2. 使用第三人称单数形式（creates, returns, validates）
3. 简洁描述函数的主要功能

### 标准格式

```go
// FuncName verb + description.
func FuncName() {}
```

### 代码示例

```34:40:pkg/client/account-server/client.go
// NewClient create a new Account server client instance.
func NewClient(c *client.Capability, version string) *Client {
```

```
// ImportGPUProductDataExcel imports GPU product data from Excel file.
func (e *excelSvc) ImportGPUProductDataExcel(cts *rest.Contexts) (interface{}, error) {
```

## 非导出函数

### 规则

1. 注释格式与导出函数相同
2. 简单明确的私有函数**可省略**注释
3. 复杂的私有函数**必须**有注释

### 代码示例

```pkg/dal/dao/dao.go
// Cvm return cvm dao.
func (s *set) Cvm() cvm.Interface {
    return &cvm.Dao{
        Orm:   s.orm,
        IDGen: s.idGen,
        Audit: s.audit,
    }
}
```

```pkg/dal/dao/cloud/cvm/cvm.go
// Interface only used for cvm.
type Interface interface {
    BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecvm.Table) ([]string, error)
    Update(kt *kit.Kit, expr *filter.Expression, model *tablecvm.Table) error
    UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecvm.Table) error
    List(kt *kit.Kit, opt *types.ListOption) (*types.ListCvmDetails, error)
    ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListCvmDetails, error)
    DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Interface = new(Dao)

// Dao cvm dao.
type Dao struct {
    Orm   orm.Interface
    IDGen idgenerator.IDGenInterface
    Audit audit.Interface
}

// BatchCreateWithTx cvm.
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecvm.Table) ([]string, error) {
```

## 参数说明

### 规则

1. 简单参数通过命名自解释，不需单独说明
2. 重要参数使用 `The xxx parameter ...` 格式单独说明
3. 每个参数说明占一行

### 代码示例

// SortByOrderedList sorts the slice items according to their position in the ordered list.
// The reverse parameter controls whether to sort in ascending (false) or descending (true) order.
// The notFoundToEnd parameter controls whether items not found in the ordered list should be placed at the end (true) or at the beginning (false).
func SortByOrderedList[T any, K comparable](items []T, orderedList []K, reverse bool, notFoundToEnd bool, extract func(T) K) {
```

## 返回值说明

### 规则

1. 简单返回值在主注释中隐含说明
2. 多种返回情况使用**缩进列表**格式

### 代码示例

// Compare returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
//
// The "less" here is the partial order.
func Compare[T cmp.Ordered](x, y T, reverse bool) int {
```

## 错误说明

### 规则

1. 通常不单独说明错误，通过函数签名中的 `error` 返回值体现
2. 特殊错误情况可在注释中补充说明

### 代码示例

```go
// LoadConfig loads configuration from the given path.
// Returns ErrNotFound if the config file does not exist.
func LoadConfig(path string) (*Config, error) {
```

## 常见动词

| 动词 | 使用场景 |
|-----|---------|
| creates/returns | 构造函数 |
| gets/retrieves | 获取数据 |
| sets/updates | 设置/更新数据 |
| validates | 验证数据 |
| parses | 解析数据 |
| converts | 转换数据 |
| handles | 处理请求/事件 |
| initializes | 初始化 |
