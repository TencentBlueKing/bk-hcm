# 函数和方法命名

## 基本规则

- 导出函数使用 **PascalCase**
- 私有函数使用 **camelCase**
- 使用标准前缀表明函数用途

## 常用前缀

| 前缀 | 用途 | 示例 |
|-----|-----|------|
| `New*` | 构造函数 | `NewClient`, `NewDaoSet` |
| `Get*` | 获取单个资源 | `GetCostTrend`, `GetTicketStatus` |
| `List*` | 获取资源列表 | `List`, `ListBrief` |
| `Batch*` | 批量操作 | `BatchCreate`, `BatchDelete` |
| `Init*` | 初始化 | `InitService`, `initInfrastructure` |
| `add*` | 私有路由注册 | `addObsApis`, `addMetaApis` |

---

## 构造函数 (New*)

```go
func NewClient(cfg *cc.Itsm, reg prometheus.Registerer) (Client, error)

// From: pkg/dal/dao/dao.go
func NewDaoSet(opt *cc.DataBase, esCfg *cc.ES) (*Set, error)

func NewOrgObsCostTrendExporter(cs clientset.ClientSet, req *anset.ExportOrgCostTrendReq) *OrgObsCostTrendExporterHandler
```

---

## 查询方法 (Get*/List*)

```go
// 获取单个资源或聚合结果
func (d *obsBillDao) GetCostTrend(kt *kit.Kit, opt *types.ObsBillDateRangeOption) ([]types.ObsBillCostTrendElement, error)

// 获取资源列表
func (d *obsBillDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListObsBillResult, error)

// 获取简要信息列表
func (d *obsBillDao) ListBrief(kt *kit.Kit, opt *types.ListOption) ([]types.Brief, error)
```

---

## 批量操作 (Batch*)

```go
func (d *obsBillDao) BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []table.ObsBill) ([]string, error)

func (d *obsBillDao) BatchDelete(kt *kit.Kit, tx *sqlx.Tx, expr filter.Expression) error

func (d *obsBillDao) BatchUpdate(kt *kit.Kit, tx *sqlx.Tx, models []table.ObsBill) error
```

---

## 初始化函数 (Init*)

```go
// 公开的服务初始化
func InitService(cap *options.Capability)
func InitBizService(cap *options.Capability)
func InitExcelService(c *options.Capability)

// 私有的基础设施初始化
func initInfrastructure(opt *cc.DataBase, esCfg *cc.ES, ...) error
```

---

## 路由注册函数 (add*)

私有路由注册函数使用 `add*` 前缀：

```go
func addObsApis(cap *options.Capability)
func addMetaApis(cap *options.Capability)
func addGPUApis(cap *options.Capability)
```

---

## 验证方法 (Validate)

请求结构体的验证方法统一命名为 `Validate`：

```go
func (req *ListDeptParentRelReq) Validate() error {
    if len(req.DeptIDs) == 0 {
        return errors.New("dept_ids is required")
    }
    return nil
}
```

---

## 特殊模式

### Parser 构造函数

```go
func NewGPUAmountParser(cs clientset.ClientSet) *GPUAmountParser
func NewObsBillParser(cs clientset.ClientSet) *ObsBillParser
```

### Audit 相关函数

```go
func New(ao audit.AuditOption) audit.Interface
func NewRetriever(ao audit.AuditOption) audit.Retriever
```
