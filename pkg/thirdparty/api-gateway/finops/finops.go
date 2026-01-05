package finops

import (
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// OpProductBgIEG BG ID for IEG
const OpProductBgIEG = 4

// MaxOpProductIDsPerRequest OpProductIds 参数最大限制数量
const MaxOpProductIDsPerRequest = 3000

// ListOpProductParam ...
type ListOpProductParam struct {
	// 筛选事业群id列表，不传筛选全部
	BgIds []int64 `json:"bg_ids"`
	// 要查询部门 id 列表(不传则筛选全部)
	DeptIds []int64 `json:"dept_ids"`
	// 要查询运营产品 id 列表(不传则筛选全部)
	OpProductIds []int64 `json:"op_product_ids"`
	// 要查询运营产品名称(支持全模糊匹配，不传则筛选全部)
	OpProductNames string        `json:"op_product_name"`
	Page           core.BasePage `json:"page"`
}

// OperationProduct 运营产品
type OperationProduct struct {
	BgId          int64  `json:"bg_id"`
	BgName        string `json:"bg_name"`
	BgShortName   string `json:"bg_short_name"`
	DeptId        int64  `json:"dept_id"`
	DeptName      string `json:"dept_name"`
	PlProductId   int64  `json:"pl_product_id"`
	PlProductName string `json:"pl_product_name"`
	OpProductId   int64  `json:"op_product_id"`
	OpProductName string `json:"op_product_name"`
	PrincipalName string `json:"principal_name"`
}

// ListOpProductResult ...
type ListOpProductResult struct {
	Count uint64             `json:"count"`
	Items []OperationProduct `json:"items"`
}

// GetDeviceLoadComplianceParam 查询业务指定日期的设备利用率达标情况参数
type GetDeviceLoadComplianceParam struct {
	BizID int64  `json:"biz_id"`
	Date  string `json:"date"`
}

// DeviceLoadComplianceResult 设备利用率达标情况
type DeviceLoadComplianceResult struct {
	CPUUsage float64 `json:"cpu_usage"`
	// AchievedKPI 是否达标
	AchievedKPI      bool    `json:"achieved_kpi"`
	EmptyLoadCPUCore float64 `json:"empty_load_cpu_core"`
	EmptyLoadOS      float64 `json:"empty_load_os"`
	LowLoadCPUCore   float64 `json:"low_load_cpu_core"`
	LowLoadOS        float64 `json:"low_load_os"`
}

// GetDeviceCPUUsageTrendParam 查询业务的CPU利用率趋势参数
type GetDeviceCPUUsageTrendParam struct {
	// 要查询的业务ID
	BizID int64 `json:"biz_id"`
	// 日期时间粒度，枚举值：day(每天)、week(周日)、month(月末)
	TimeGranularity enumor.LoadUsageTimeGranularity `json:"time_granularity"`
	// 要查询时间范围。当粒度为day时，不能超过90天；当粒度为week或month时，不能超过25个月
	DateRange *DateRange `json:"date_range"`
	// 设备环境列表，不传时查询全部，枚举值：idc(国内环境)、sg(SG环境)
	Envs []enumor.LoadUsageDeviceENV `json:"envs,omitempty"`
	// 设备类型列表，不传时查询全部，枚举值：cvm(虚拟机)、bareMetal(物理机)
	DevTypes []enumor.FinOpsDeviceType `json:"dev_types,omitempty"`
	// 包含DB设备列表，不传时查询全部，传true时查询DB设备，传false时查询非DB设备
	IncludeDBA []bool `json:"include_dba,omitempty"`
	// 包含考核设备列表，不传时查询全部，传true时查询考核设备，传false时查询无考核设备
	IncludeExamine []bool `json:"include_examine,omitempty"`
	// 包含上报设备列表，不传时查询全部，传true时查询有上报数据，传false时查询无上报数据
	IncludeReport []bool `json:"include_report,omitempty"`
}

// DateRange 时间范围
type DateRange struct {
	// 起始时间，格式为"年-月-日"，如"2023-06-01"
	Start string `json:"start"`
	// 结束时间，不能晚于当前日期，格式同 start
	End string `json:"end"`
}

// CPUUsageTrendData CPU利用率趋势数据
type CPUUsageTrendData struct {
	Date     string  `json:"date"`
	CPUUsage float64 `json:"cpu_usage"`
}

// GetDeviceCPUUsageTrendResult 查询业务的CPU利用率趋势响应结果
type GetDeviceCPUUsageTrendResult struct {
	Trend []CPUUsageTrendData `json:"trend"`
}

// GetBudgetDeclarationOperatorParam 获取预算申报单据运营产品和操作人对应关系参数
type GetBudgetDeclarationOperatorParam struct {
	// 年份列表，最大20，不传时查询全部
	Years []int `json:"years,omitempty"`
	// 运营产品ID列表，最大3000，不传时查询全部
	OpProductIDs []int64 `json:"op_product_ids,omitempty"`
}

// BudgetDeclarationOperatorItem 预算申报单据运营产品和操作人对应关系项
type BudgetDeclarationOperatorItem struct {
	// 年份
	Year int `json:"year"`
	// 组成
	Composition []BudgetDeclarationComposition `json:"composition"`
}

// BudgetDeclarationComposition 预算申报组成
type BudgetDeclarationComposition struct {
	// 运营产品ID
	OpProductID int64 `json:"op_product_id"`
	// 创建人列表
	Creators []string `json:"creators"`
	// 提交人列表
	Committers []string `json:"committers"`
}

// GetBudgetDeclarationOperatorResult 获取预算申报单据运营产品和操作人对应关系结果
type GetBudgetDeclarationOperatorResult struct {
	// 总记录条数
	Count int `json:"count"`
	// 关系列表
	Items []BudgetDeclarationOperatorItem `json:"items"`
}

// Client FinOps Client
type Client interface {
	// ListOpProduct 查询全内部事业群运营产品
	ListOpProduct(kt *kit.Kit, params *ListOpProductParam) (*ListOpProductResult, error)
	// GetDeviceLoadCompliance 查询业务指定日期的设备利用率达标情况
	GetDeviceLoadCompliance(kt *kit.Kit, params *GetDeviceLoadComplianceParam) (*DeviceLoadComplianceResult, error)
	// GetDeviceCPUUsageTrend 查询业务的CPU利用率趋势
	GetDeviceCPUUsageTrend(kt *kit.Kit, params *GetDeviceCPUUsageTrendParam) (*GetDeviceCPUUsageTrendResult, error)
	// GetBudgetDeclarationOperator 查询预算提报人
	GetBudgetDeclarationOperator(kt *kit.Kit, params *GetBudgetDeclarationOperatorParam) (*GetBudgetDeclarationOperatorResult, error)
}

// NewClient initialize a new FinOps client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &apigateway.Discovery{
			Name:    "fineOps",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/v1")
	systemRestCli := rest.NewClient(c, "/api/system/v1")
	return &finOps{
		config:    cfg,
		client:    restCli,
		systemCli: systemRestCli,
	}, nil
}

// fineOps is an esb client to request fineOps.
type finOps struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
	// http client instance for system requests
	systemCli rest.ClientInterface
}

// ListOpProduct 查询全内部事业群运营产品 get_op_product_meta
func (c *finOps) ListOpProduct(kt *kit.Kit, params *ListOpProductParam) (*ListOpProductResult, error) {

	return apigateway.ApiGatewayCall[ListOpProductParam, ListOpProductResult](c.client, c.config, rest.POST,
		kt, params, "/analysis/dm/meta/get/op_product/info")
}

// GetDeviceLoadCompliance 查询业务指定日期的设备利用率达标情况 get_dm_device_load_compliance
func (c *finOps) GetDeviceLoadCompliance(kt *kit.Kit, params *GetDeviceLoadComplianceParam) (
	*DeviceLoadComplianceResult, error) {

	return apigateway.ApiGatewayCall[GetDeviceLoadComplianceParam, DeviceLoadComplianceResult](c.systemCli, c.config,
		rest.POST, kt, params, "/analysis/dm/device/get/usage/compliance")
}

// GetDeviceCPUUsageTrend 查询业务的CPU利用率趋势 get_dm_device_cpu_usage_trend
func (c *finOps) GetDeviceCPUUsageTrend(kt *kit.Kit, params *GetDeviceCPUUsageTrendParam) (
	*GetDeviceCPUUsageTrendResult, error) {

	return apigateway.ApiGatewayCall[GetDeviceCPUUsageTrendParam, GetDeviceCPUUsageTrendResult](c.systemCli, c.config,
		rest.POST, kt, params, "/analysis/dm/device/get/cpu/usage/trend")
}

// GetBudgetDeclarationOperator 查询预算提报人
func (c *finOps) GetBudgetDeclarationOperator(kt *kit.Kit, params *GetBudgetDeclarationOperatorParam) (
	*GetBudgetDeclarationOperatorResult, error) {

	return apigateway.ApiGatewayCall[GetBudgetDeclarationOperatorParam, GetBudgetDeclarationOperatorResult](c.systemCli, c.config,
		rest.POST, kt, params, "/analysis/dm/budget/get/declaration/operator")
}

// 其他请求使用esb 接口
