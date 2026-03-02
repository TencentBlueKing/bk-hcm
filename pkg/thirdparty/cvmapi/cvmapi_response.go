/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cvmapi

import (
	"hcm/pkg/criteria/enumor"

	"github.com/shopspring/decimal"
)

// RespMeta cvm response meta info
type RespMeta struct {
	Id      string    `json:"id"`
	JsonRpc string    `json:"jsonrpc"`
	TraceId string    `json:"x_trace_id"`
	Error   RespError `json:"error"`
}

// RespError cvm response error info
type RespError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OrderCreateResp cvm create order response
type OrderCreateResp struct {
	RespMeta `json:",inline"`
	Result   OrderCreateRst `json:"result"`
}

// OrderCreateRst cvm create order result
type OrderCreateRst struct {
	OrderId string `json:"orderId"`
	Status  int    `json:"status"`
}

// OrderQueryResp cvm order query response
type OrderQueryResp struct {
	RespMeta `json:",inline"`
	Result   *OrderQueryRst `json:"result"`
}

// OrderQueryRst cvm order query result
type OrderQueryRst struct {
	Total int          `json:"total"`
	Data  []*OrderItem `json:"data"`
}

// FailInstanceInfo cvm order fail instance info
// 由于 CRP 接口协议混乱，目前两种命名方式都有，详细请联系 crp 确认，目前使用下划线命名法的数据
type FailInstanceInfo struct {
	ErrorMsgTypeEn string `json:"errorMsgTypeEn"`
	ErrorMsg1      string `json:"errorMsg"`
	ErrorMsg       string `json:"error_msg"`
	ErrorType1     string `json:"errorType"`
	ErrorType      string `json:"error_type"`
	ErrorMsgTypeCn string `json:"errorMsgTypeCn"`
	RequestId      string `json:"requestId"`
	ErrorCount     int    `json:"error_count"`
	Operator       string `json:"operator"`
	ErrorCount1    int    `json:"errorCount"`
}

// OrderItem cvm order info
type OrderItem struct {
	OrderId string `json:"orderId"`
	// 单据状态：
	// 8完成
	// 0待部门管理员审批,1待业务总监审批,2待规划经理审批,3待资源审批,4待生成CDH宿主机,
	// 5CDH宿主机生成中,6待生成CVM,7CVM生成中,127驳回,129下发生产失败
	Status            int                `json:"status"`
	StatusDesc        string             `json:"statusDesc"`
	ProductId         int64              `json:"productId"`
	ProductName       string             `json:"productName"`
	FailInstanceInfos []FailInstanceInfo `json:"failInstanceInfo"`
	SucInstanceCount  int                `json:"sucInstanceCount"`
	CreateTime        string             `json:"createTime"` // 提单时间，示例：2023-02-28 11:16:47
}

// InstanceQueryResp cvm instance query response
type InstanceQueryResp struct {
	RespMeta `json:",inline"`
	Result   *InstanceQueryRst `json:"result"`
}

// InstanceQueryRst cvm instance query result
type InstanceQueryRst struct {
	Total int             `json:"total"`
	Data  []*InstanceItem `json:"data"`
}

// InstanceItem cvm instance info
type InstanceItem struct {
	InstanceId      string                `json:"instanceId"`
	InstanceStatus  enumor.InstanceStatus `json:"instanceStatus"` // PENDING:创建中 RUNNING:成功创建 SHUTDOWN:关机待回收
	AssetId         string                `json:"instanceAssetId"`
	LanIp           string                `json:"lanIp"`
	WanIp           string                `json:"wanIp"`
	OwnerLanIp      string                `json:"ownerLanIp"`
	CloudCampus     string                `json:"cloudCampus"`
	SecurityGroupId string                `json:"securityGroupId"`
	ImageName       string                `json:"imageName"`
	PrivateVpcId    string                `json:"privateVpcId"`
	// TODO CRP返回该字段为空，需确认
	CloudRegion     string `json:"cloudRegion"`
	PrivateSubnetId string `json:"privateSubnetId"`
	CreateTime      string `json:"createTime"`
	Pool            int    `json:"pool"`
	ObsProject      string `json:"obsProject"`
}

// PlanOrderChangeResp cvm plan order change response
type PlanOrderChangeResp struct {
	RespMeta `json:",inline"`
	Result   *PlanOrderChangeRst `json:"result"`
}

// PlanOrderChangeRst cvm plan order change result
type PlanOrderChangeRst struct {
	Total int                    `json:"total"`
	Data  []*PlanOrderChangeItem `json:"data"`
}

// PlanOrderChangeItem cvm plan order change item
type PlanOrderChangeItem struct {
	UseTime           string            `json:"useTime"`
	BgName            string            `json:"bgName"`
	DeptName          string            `json:"deptName"`
	PlanProductName   string            `json:"planProductName"`
	ProductName       string            `json:"productName"`
	ProjectName       enumor.ObsProject `json:"projectName"`
	CityName          string            `json:"cityName"`
	ZoneName          string            `json:"zoneName"`
	TechnicalClass    string            `json:"technicalClass"`
	InstanceFamily    string            `json:"instanceFamily"`
	InstanceType      string            `json:"instanceType"`
	InstanceModel     string            `json:"instanceModel"`
	CoreTypeName      string            `json:"coreTypeName"`
	DiskTypeName      string            `json:"diskTypeName"`
	PlanType          enumor.PlanType   `json:"planType"`
	ChangeCvmAmount   decimal.Decimal   `json:"changeCvmAmount"`
	ChangeCoreAmount  int64             `json:"changeCoreAmount"`
	ChangeRamAmount   int64             `json:"changeRamAmount"`
	ChangedDiskAmount int64             `json:"changedDiskAmount"`
	InstanceIO        int64             `json:"instanceIO"`
	SourceType        string            `json:"sourceType"`
	OrderId           string            `json:"orderId"`
	ResourceMode      enumor.ResMode    `json:"resourceMode"`
	ReturnPlanTime    string            `json:"returnPlanTime"`
	Desc              string            `json:"desc"`
}

// DemandChangeLogQueryResp cvm demand change log query response
type DemandChangeLogQueryResp struct {
	RespMeta  `json:",inline"`
	Result    *DemandChangeLogQueryRst `json:"result"`
	Errorinfo interface{}              `json:"errorinfo"`
}

// DemandChangeLogQueryRst cvm demand change log query result
type DemandChangeLogQueryRst struct {
	Total int                               `json:"total"`
	Data  []*DemandChangeLogQueryDemandItem `json:"data"`
}

// DemandChangeLogQueryDemandItem cvm demand change log query demand item
type DemandChangeLogQueryDemandItem struct {
	DemandId int                            `json:"demandId"`
	Info     []*DemandChangeLogQueryLogItem `json:"info"`
}

// DemandChangeLogQueryLogItem cvm demand change log query log item
type DemandChangeLogQueryLogItem struct {
	DemandId            int     `json:"demandId"`
	UseTime             string  `json:"useTime"`
	BgName              string  `json:"bgName"`
	DeptName            string  `json:"deptName"`
	PlanProductName     string  `json:"planProductName"`
	ProductName         string  `json:"productName"`
	ProjectName         string  `json:"projectName"`
	CityName            string  `json:"cityName"`
	ZoneName            string  `json:"zoneName"`
	RequirementWeekType string  `json:"requirementWeekType"`
	ResourcePoolType    int     `json:"resourcePoolType"`
	InstanceType        string  `json:"instanceType"`
	InstanceModel       string  `json:"instanceModel"`
	ChangeCvmAmount     float32 `json:"changeCvmAmount"`
	AfterCvmAmount      float32 `json:"afterCvmAmount"`
	ChangeCoreAmount    float32 `json:"changeCoreAmount"`
	AfterCoreAmount     float32 `json:"afterCoreAmount"`
	ChangeRamAmount     float32 `json:"changeRamAmount"`
	AfterRamAmount      float32 `json:"afterRamAmount"`
	DiskTypeName        string  `json:"diskTypeName"`
	InstanceIO          int     `json:"instanceIO"`
	ChangedDiskAmount   float32 `json:"changedDiskAmount"`
	AfterDiskAmount     float32 `json:"afterDiskAmount"`
	SourceType          string  `json:"sourceType"`
	OrderId             string  `json:"orderId"`
	CreateTime          string  `json:"createTime"`
	Desc                string  `json:"desc"`
	ResourcePoolName    string  `json:"resourcePoolName"`
}

// CvmCbsPlanPenaltyRatioReportResp cvm and cbs plan penalty ratio report response
type CvmCbsPlanPenaltyRatioReportResp struct {
	RespMeta  `json:",inline"`
	Result    *CvmCbsPlanPenaltyRatioReportRst `json:"result"`
	Errorinfo interface{}                      `json:"errorinfo"`
}

// CvmCbsPlanPenaltyRatioReportRst cvm and cbs plan penalty ratio report result
type CvmCbsPlanPenaltyRatioReportRst struct {
	Message string `json:"message"`
}

// CvmCbsPlanQueryResp cvm and cbs plan query response
type CvmCbsPlanQueryResp struct {
	RespMeta  `json:",inline"`
	Result    CvmCbsPlanQueryRst `json:"result"`
	Errorinfo interface{}        `json:"errorinfo"`
}

// CvmCbsPlanQueryRst cvm and cbs plan query result
type CvmCbsPlanQueryRst struct {
	Total         int                    `json:"total"`
	Data          []*CvmCbsPlanQueryItem `json:"data"`
	AllCvmAmount  float64                `json:"allCvmAmount"`
	AllCoreAmount int64                  `json:"allCoreAmount"`
}

// CvmCbsPlanQueryItem cvm and cbs plan query item
type CvmCbsPlanQueryItem struct {
	BaseCoreAmount     int                `json:"baseCoreAmount"`
	BaseCvmAmount      float64            `json:"baseCvmAmount"`
	SliceId            string             `json:"sliceId"`
	YearMonth          string             `json:"yearMonth"`
	Year               int                `json:"year"`
	Month              int                `json:"month"`
	Week               int                `json:"week"`
	YearMonthWeek      string             `json:"yearMonthWeek"`
	ExpectStartDate    string             `json:"expectStartDate"`
	ExpectEndDate      string             `json:"expectEndDate"`
	UseTime            string             `json:"useTime"`
	BgId               int                `json:"bgId"`
	BgName             string             `json:"bgName"`
	DeptId             int                `json:"deptId"`
	DeptName           string             `json:"deptName"`
	PlanProductId      int                `json:"planProductId"`
	PlanProductName    string             `json:"planProductName"`
	ProductId          int                `json:"productId"`
	ProductName        string             `json:"productName"`
	ProjectName        enumor.ObsProject  `json:"projectName"`
	OrderId            string             `json:"orderId"`
	CityId             int                `json:"cityId"`
	CityName           string             `json:"cityName"`
	ZoneId             int                `json:"zoneId"`
	ZoneName           string             `json:"zoneName"`
	InPlan             string             `json:"inPlan"`
	PlanWeek           int                `json:"planWeek"`
	ExpeditedPostponed string             `json:"expeditedPostponed"`
	CoreType           int                `json:"coreType"`
	CoreTypeName       string             `json:"coreTypeName"`
	InstanceFamily     string             `json:"instanceFamily"`
	InstanceType       string             `json:"instanceType"`
	InstanceModel      string             `json:"instanceModel"`
	InstanceIO         int                `json:"instanceIO"`
	DiskType           enumor.CRPDiskType `json:"diskType"`
	DiskTypeName       string             `json:"diskTypeName"`
	// CvmAmount 未执行需求数
	CvmAmount     float64 `json:"cvmAmount"`
	RamAmount     float64 `json:"ramAmount"` // CRP 格式定义有问题，实际一定是int，可以按照int64处理
	CoreAmount    int64   `json:"coreAmount"`
	AllDiskAmount int64   `json:"allDiskAmount"`
	// ApplyCvmAmount 已申领数
	ApplyCvmAmount  float64 `json:"applyCvmAmount"`
	ApplyRamAmount  float64 `json:"applyRamAmount"`
	ApplyCoreAmount int64   `json:"applyCoreAmount"`
	ApplyDiskAmount int64   `json:"applyDiskAmount"`
	// PlanCvmAmount 总需求数
	PlanCvmAmount  float64 `json:"planCvmAmount"`
	PlanRamAmount  float64 `json:"planRamAmount"`
	PlanCoreAmount int64   `json:"planCoreAmount"`
	PlanDiskAmount int64   `json:"planDiskAmount"`
	// ExpiredCvmAmount 已过期数
	ExpiredCvmAmount  float64 `json:"expiredCvmAmount"`
	ExpiredRamAmount  float64 `json:"expiredRamAmount"`
	ExpiredCoreAmount int64   `json:"expiredCoreAmount"`
	ExpiredDiskAmount int64   `json:"expiredDiskAmount"`
	// RealCvmAmount 未过期的未执行数
	RealCvmAmount         float64                    `json:"realCvmAmount"`
	RealRamAmount         float64                    `json:"realRamAmount"`
	RealCoreAmount        int64                      `json:"realCoreAmount"`
	RealDiskAmount        int64                      `json:"realDiskAmount"`
	MjOrderId             string                     `json:"mjOrderId"`
	RequirementStatus     int                        `json:"requirementStatus"`
	RequirementStatusName string                     `json:"requirementStatusName"`
	RequirementWeekType   string                     `json:"requirementWeekType"`
	IsManualWeekType      int                        `json:"isManualWeekType"`
	IsInProcessing        int                        `json:"isInProcessing"`
	ProcessingOrderId     string                     `json:"processingOrderId"`
	DemandId              string                     `json:"demandId"`
	ResourcePoolType      int                        `json:"resourcePoolType"`
	ResourcePoolName      string                     `json:"resourcePoolName"`
	ResourceMode          string                     `json:"resourceMode"`
	StatisticalClass      string                     `json:"statisticalClass"`
	TechnicalClass        string                     `json:"technicalClass"`
	VagueStatus           int                        `json:"vagueStatus"`
	ReviewStatus          enumor.ResPlanReviewStatus `json:"reviewStatus"`
	ForecastType          string                     `json:"forecastType"` // 需求类型（常规需求、年度预算）
	GenerationType        string                     `json:"generation_type"`
	// 短租退回时间相关参数
	IsAutoReturnPlan bool   `json:"isAutoReturnPlan"`
	ReturnPlanTime   string `json:"returnPlanTime"`
}

// Clone return a clone CvmCbsPlanQueryItem.
func (i *CvmCbsPlanQueryItem) Clone() *CvmCbsPlanQueryItem {
	return &CvmCbsPlanQueryItem{
		BaseCoreAmount:        i.BaseCoreAmount,
		BaseCvmAmount:         i.BaseCvmAmount,
		SliceId:               i.SliceId,
		YearMonth:             i.YearMonth,
		Year:                  i.Year,
		Month:                 i.Month,
		Week:                  i.Week,
		YearMonthWeek:         i.YearMonthWeek,
		ExpectStartDate:       i.ExpectStartDate,
		ExpectEndDate:         i.ExpectEndDate,
		UseTime:               i.UseTime,
		ReturnPlanTime:        i.ReturnPlanTime,
		BgId:                  i.BgId,
		BgName:                i.BgName,
		DeptId:                i.DeptId,
		DeptName:              i.DeptName,
		PlanProductId:         i.PlanProductId,
		PlanProductName:       i.PlanProductName,
		ProductId:             i.ProductId,
		ProductName:           i.ProductName,
		ProjectName:           i.ProjectName,
		OrderId:               i.OrderId,
		CityId:                i.CityId,
		CityName:              i.CityName,
		ZoneId:                i.ZoneId,
		ZoneName:              i.ZoneName,
		InPlan:                i.InPlan,
		PlanWeek:              i.PlanWeek,
		ExpeditedPostponed:    i.ExpeditedPostponed,
		CoreType:              i.CoreType,
		CoreTypeName:          i.CoreTypeName,
		InstanceFamily:        i.InstanceFamily,
		InstanceType:          i.InstanceType,
		InstanceModel:         i.InstanceModel,
		InstanceIO:            i.InstanceIO,
		DiskType:              i.DiskType,
		DiskTypeName:          i.DiskTypeName,
		CvmAmount:             i.CvmAmount,
		RamAmount:             i.RamAmount,
		CoreAmount:            i.CoreAmount,
		AllDiskAmount:         i.AllDiskAmount,
		ApplyCvmAmount:        i.ApplyCvmAmount,
		ApplyRamAmount:        i.ApplyRamAmount,
		ApplyCoreAmount:       i.ApplyCoreAmount,
		ApplyDiskAmount:       i.ApplyDiskAmount,
		PlanCvmAmount:         i.PlanCvmAmount,
		PlanRamAmount:         i.PlanRamAmount,
		PlanCoreAmount:        i.PlanCoreAmount,
		PlanDiskAmount:        i.PlanDiskAmount,
		ExpiredCvmAmount:      i.ExpiredCvmAmount,
		ExpiredRamAmount:      i.ExpiredRamAmount,
		ExpiredCoreAmount:     i.ExpiredCoreAmount,
		ExpiredDiskAmount:     i.ExpiredDiskAmount,
		RealCvmAmount:         i.RealCvmAmount,
		RealRamAmount:         i.RealRamAmount,
		RealCoreAmount:        i.RealCoreAmount,
		RealDiskAmount:        i.RealDiskAmount,
		MjOrderId:             i.MjOrderId,
		RequirementStatus:     i.RequirementStatus,
		RequirementStatusName: i.RequirementStatusName,
		RequirementWeekType:   i.RequirementWeekType,
		IsManualWeekType:      i.IsManualWeekType,
		IsInProcessing:        i.IsInProcessing,
		ProcessingOrderId:     i.ProcessingOrderId,
		DemandId:              i.DemandId,
		ResourcePoolType:      i.ResourcePoolType,
		ResourcePoolName:      i.ResourcePoolName,
		ResourceMode:          i.ResourceMode,
		StatisticalClass:      i.StatisticalClass,
		VagueStatus:           i.VagueStatus,
		ReviewStatus:          i.ReviewStatus,
		ForecastType:          i.ForecastType,
		GenerationType:        i.GenerationType,
	}
}

// CvmCbsPlanAdjustResp cvm and cbs plan adjust response
type CvmCbsPlanAdjustResp struct {
	RespMeta  `json:",inline"`
	Result    *CvmCbsPlanAdjustRst `json:"result"`
	Errorinfo interface{}          `json:"errorinfo"`
}

// CvmCbsPlanAdjustRst cvm and cbs plan adjust result
// adjustOrder 和 submitAutoAdjustOrder 返回格式不同，需注意
type CvmCbsPlanAdjustRst struct {
	Status  int    `json:"status"`
	OrderId string `json:"orderId"`
}

// AddCvmCbsPlanResp add cvm and cbs plan order response
type AddCvmCbsPlanResp struct {
	RespMeta `json:",inline"`
	Result   *AddCvmCbsPlanRst `json:"result"`
}

// AddCvmCbsPlanRst add cvm and cbs plan order result
type AddCvmCbsPlanRst struct {
	Status  int    `json:"status"`
	OrderId string `json:"orderId"`
}

// QueryPlanOrderResp query cvm and cbs plan order response
type QueryPlanOrderResp struct {
	RespMeta `json:",inline"`
	Result   map[string]*QueryPlanOrderRst `json:"result"`
}

// QueryPlanOrderRst query cvm and cbs plan order result
type QueryPlanOrderRst struct {
	Code int           `json:"code"`
	Data PlanOrderData `json:"data"`
}

// PlanOrderData query cvm and cbs plan order data
type PlanOrderData struct {
	BaseInfo PlanOrderBaseInfo `json:"baseInfo"`
}

// PlanOrderStatus cvm and cbs plan order status
type PlanOrderStatus int

const (
	// PlanOrderStatusDeptAdmin 部门管理员审批
	PlanOrderStatusDeptAdmin PlanOrderStatus = 1
	// PlanOrderStatusPlanManager 规划经理审批
	PlanOrderStatusPlanManager PlanOrderStatus = 2
	// PlanOrderStatusResManager 资源经理审批
	PlanOrderStatusResManager PlanOrderStatus = 3
	// PlanOrderStatusFinished 申请结束
	PlanOrderStatusFinished PlanOrderStatus = 4
	// PlanOrderStatusArchPlat 架平审批
	PlanOrderStatusArchPlat PlanOrderStatus = 6
	// PlanOrderStatusResGM 资源总监审批
	PlanOrderStatusResGM PlanOrderStatus = 10
	// PlanOrderStatusRejected 审批驳回
	PlanOrderStatusRejected PlanOrderStatus = 127
	// PlanOrderStatusApproved 审批通过
	PlanOrderStatusApproved PlanOrderStatus = 20
)

// PlanOrderBaseInfo query cvm and cbs plan order base info
type PlanOrderBaseInfo struct {
	Status           PlanOrderStatus `json:"status"`
	StatusMsg        string          `json:"statusMsg"`
	StatusDesc       string          `json:"statusDesc"`
	CurrentProcessor string          `json:"currentProcessor"`
}

// QueryReturnPlanResp query cvm return plan response
type QueryReturnPlanResp struct {
	RespMeta `json:",inline"`
	Result   *QueryReturnPlanRst `json:"result"`
}

// QueryReturnPlanRst query cvm return plan result
type QueryReturnPlanRst struct {
	Total                  int               `json:"total"`
	Data                   []*ReturnPlanItem `json:"data"`
	TotalCoreAmount        float64           `json:"totalCoreAmount"`
	TotalAppliedCoreAmount float64           `json:"totalAppliedCoreAmount"`
	TotalLeftCoreAmount    float64           `json:"totalLeftCoreAmount"`
	TotalExpiredCoreAmount float64           `json:"totalExpiredCoreAmount"`
}

// ReturnPlanItem cvm return plan item
type ReturnPlanItem struct {
	ID                 int64             `json:"id"`
	BGID               int64             `json:"bgId"`
	BGName             string            `json:"bgName"`
	DeptID             int64             `json:"deptID"`
	DeptName           string            `json:"deptName"`
	PlanProductID      int64             `json:"planProductID"`
	PlanProductName    string            `json:"planProductName"`
	ProductID          int64             `json:"productID"`
	ProductName        string            `json:"productName"`
	ProjectName        enumor.ObsProject `json:"projectName"`
	PlanTime           string            `json:"planTime"` // YYYY-MM-DD
	GenerationType     int64             `json:"generationType"`
	GenerationTypeName string            `json:"generationTypeName"` // 采购代次
	CoreType           int64             `json:"coreType"`
	CoreTypeName       string            `json:"coreTypeName"`
	DeviceFamilyName   string            `json:"deviceFamilyName"` // 物理机机型族：云上计算标准
	InstanceModel      string            `json:"instanceModel"`    // CVM机型
	InstanceType       string            `json:"instanceType"`
	CityName           string            `json:"cityName"`
	ZoneName           string            `json:"zoneName"`
	CvmAmount          float64           `json:"cvmAmount"`
	CoreAmount         decimal.Decimal   `json:"coreAmount"` // 退回计划可能有小数核心
}

// CapacityResp cvm apply capacity query response
type CapacityResp struct {
	RespMeta `json:",inline"`
	Result   *CapacityRst `json:"result"`
}

// CapacityRst cvm apply capacity query result
type CapacityRst struct {
	MaxNum  int             `json:"maxNum"`
	MaxInfo []*CapacityInfo `json:"maxInfo"`
	Ret     int             `json:"ret"`
	Msg     string          `json:"msg"`
}

// CapacityInfo cvm apply capacity into
type CapacityInfo struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

// VpcResp cvm vpc query response
type VpcResp struct {
	RespMeta `json:",inline"`
	Result   []*VpcInfo `json:"result"`
}

// VpcInfo cvm vpc query result
type VpcInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// SubnetResp cvm subnet query response
type SubnetResp struct {
	RespMeta `json:",inline"`
	Result   []*SubnetInfo `json:"result"`
}

// SubnetInfo cvm subnet query result
type SubnetInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	LeftIpNum int    `json:"leftIpNum"`
}

// ReturnQueryResp cvm return order query response
type ReturnQueryResp struct {
	RespMeta `json:",inline"`
	Result   *ReturnQueryRst `json:"result"`
}

// ReturnQueryRst cvm return order query result
type ReturnQueryRst struct {
	Total int                `json:"total"`
	Data  []*ReturnOrderItem `json:"data"`
}

// ReturnOrderItem cvm return order info
type ReturnOrderItem struct {
	Status      int               `json:"status"`
	Description string            `json:"statusDesc"`
	Message     string            `json:"statusMsg"`
	ReturnCnt   int               `json:"returnCount"`
	FinishCnt   int               `json:"finishCount"`
	Instances   []*ReturnInstance `json:"returnInstances"`
}

// ReturnInstance cvm return instance info
type ReturnInstance struct {
	InstanceId   string `json:"instanceId"`
	Status       int    `json:"status"`
	Summary      string `json:"summaryStatus"`
	Description  string `json:"statusDesc"`
	ReturnBudget bool   `json:"returnBudget"`
	FinishTime   string `json:"finishTime"`
}

// ReturnDetailResp cvm return order detail query response
type ReturnDetailResp struct {
	RespMeta `json:",inline"`
	Result   *ReturnDetailRst `json:"result"`
}

// ReturnDetailRst cvm return order detail query result
type ReturnDetailRst struct {
	Total int             `json:"total"`
	Data  []*ReturnDetail `json:"data"`
}

// ReturnDetail cvm return order detail info
type ReturnDetail struct {
	InstanceId string `json:"instanceId"`
	AssetId    string `json:"instanceAssetId"`
	Ip         string `json:"lanIp"`
	// instance return status:
	// 1	回收站
	// 2	云销毁中
	// 10	停用并回收IP
	// 15	从CMDB删除
	// 20	销毁完成
	// 127  审批驳回
	// 128	异常终止
	Status int `json:"status"`
	// Tag return plan tag
	Tag string `json:"tag"`
	// Partition cost sharing ratio
	Partition float64 `json:"partition"`
	// RetPlanMsg return plan and cost sharing remark
	RetPlanMsg string `json:"returnPlanMessage"`
	FinishTime string `json:"finishTime"`
}

// UpgradeDetailResp cvm upgrade order detail query response
type UpgradeDetailResp struct {
	RespMeta `json:",inline"`
	Result   *UpgradeDetail `json:"result"`
}

// UpgradeDetail cvm upgrade order detail info
type UpgradeDetail struct {
	OrderID string `json:"orderId"`
	// Status 订单状态：
	// 0 - 部门管理员审批
	// 1 - 规划经理审批 （免审流程不存在）
	// 2 - 资源经理审批（免审流程不存在）
	// 9 - 待确认执行
	// 10 - 执行中
	// 20 - 执行完成
	// 127 - 驳回 （免审流程不存在）
	// 128 - 订单失败
	Status          enumor.CrpUpgradeOrderStatus `json:"status"`
	StatusMsg       string                       `json:"statusMsg"`
	InstanceCount   int                          `json:"instanceCount"`
	BG              string                       `json:"bg"`
	VirtualDeptID   int                          `json:"virtualDeptId"`
	VirtualDept     string                       `json:"virtualDept"`
	Region          string                       `json:"region"`
	RegionName      string                       `json:"regionName"`
	Creator         string                       `json:"creator"`
	ProjectName     string                       `json:"projectName"`
	Reason          string                       `json:"reason"`
	CurrentApprover string                       `json:"currentApprover"`
	CreateTime      string                       `json:"createTime"`
	UpdateTime      string                       `json:"updateTime"`
	FinishTime      string                       `json:"finishTime"`
	AuditTime       string                       `json:"auditTime"`
	DetailList      []UpgradeDetailInstance      `json:"detailList"`
}

// UpgradeDetailInstance cvm upgrade order detail instance
type UpgradeDetailInstance struct {
	// InstanceID 实例云上ID
	InstanceID string `json:"instanceId"`
	// InstanceAssetID 实例固资号
	InstanceAssetID string `json:"instanceAssetId"`
	Zone            string `json:"zone"`
	ZoneName        string `json:"zoneName"`
	// Status 实例状态
	// WAITING - 待操作
	// OPERATING - 操作中
	// SUCCESS - 成功
	// FAILED - 失败
	Status             enumor.CrpUpgradeCVMStatus `json:"status"`
	StatusDesc         string                     `json:"statusDesc"`
	ReqID              string                     `json:"reqId"`
	SourceInstanceType string                     `json:"sourceInstanceType"`
	TargetInstanceType string                     `json:"targetInstanceType"`
	ResetType          string                     `json:"resetType"`
	CreateTime         string                     `json:"createTime"`
	UpdateTime         string                     `json:"updateTime"`
	FinishTime         string                     `json:"finishTime"`
	CloudFinishTime    string                     `json:"cloudFinishTime"`
	Business1ID        int                        `json:"business1Id"`
	Business1Name      string                     `json:"business1Name"`
	Business2ID        int                        `json:"business2Id"`
	Business2Name      string                     `json:"business2Name"`
	Business3ID        int                        `json:"business3Id"`
	Business3Name      string                     `json:"business3Name"`
	ProductID          int                        `json:"productId"`
	ProductName        string                     `json:"productName"`
}

// GetCvmProcessResp get cvm process response
type GetCvmProcessResp struct {
	RespMeta `json:",inline"`
	Result   *GetCvmProcessRst `json:"result"`
}

// GetCvmProcessRst get cvm process result
type GetCvmProcessRst struct {
	Total int               `json:"total"`
	Data  []*CvmProcessItem `json:"data"`
}

// CvmProcessItem cvm process item
type CvmProcessItem struct {
	InstanceId string `json:"instanceId"`
	AssetId    string `json:"instanceAssetId"`
	Ip         string `json:"lanIp"`
	OrderId    string `json:"orderId"`
	// StatusDesc cvm process status description
	// OTHERS(-1, "未定义的流程"),
	// EMPTY(0, ""),
	// UPGRADE(1, "升降配中"),
	// MIGRATE(2, "迁移中"),
	// EXCHANGE(8, "置换中"),
	// RETURN(9, "销毁中")
	StatusDesc string `json:"statusDesc"`
}

// GetErpProcessResp get erp process response
type GetErpProcessResp struct {
	RespMeta `json:",inline"`
	Result   *GetErpProcessRst `json:"result"`
}

// GetErpProcessRst get erp process result
type GetErpProcessRst struct {
	Total int               `json:"total"`
	Data  []*ErpProcessItem `json:"data"`
}

// ErpProcessItem erp process item
type ErpProcessItem struct {
	AssetId    string `json:"logicPcCode"`
	OrderId    string `json:"orderCode"`
	ActionType string `json:"actionType"`
}

// QueryCvmInstanceTypeResp query cvm instance type response
type QueryCvmInstanceTypeResp struct {
	RespMeta `json:",inline"`
	Result   *QueryCvmInstanceTypeRst `json:"result"`
}

// QueryCvmInstanceTypeRst query cvm instance type result
type QueryCvmInstanceTypeRst struct {
	Data []QueryCvmInstanceTypeItem `json:"data"`
}

// InstanceTypeClass 通/专用机型，SpecialType专用，CommonType通用
type InstanceTypeClass string

const (
	// SpecialType 专用机型
	SpecialType InstanceTypeClass = "SpecialType"
	// CommonType 通用机型
	CommonType InstanceTypeClass = "CommonType"
)

// QueryCvmInstanceTypeItem query cvm instance type item
type QueryCvmInstanceTypeItem struct {
	InstanceClassDesc     string            `json:"instanceClassDesc"`     // 实例类型
	InstanceType          string            `json:"instanceType"`          // 实例规格
	InstanceTypeClass     InstanceTypeClass `json:"instanceTypeClass"`     // 通/专用机型，SpecialType专用，CommonType通用
	InstanceTypeClassDesc string            `json:"instanceTypeClassDesc"` // // 通/专用机型
	RamAmount             float64           `json:"ramAmount"`             // 内存
	GPUType               string            `json:"gpuType"`               // GPU类型
	FirmName              string            `json:"firmName"`              // 厂商
	InstanceGroup         string            `json:"instanceGroup"`         // 机型族
	CPUAmount             float64           `json:"cpuAmount"`             // CPU数量
	GPUAmount             float64           `json:"gpuAmount"`             // GPU卡数量
	InstanceClass         string            `json:"instanceClass"`         // 实例类型
	CoreType              int               `json:"coreType"`              // 1.2.3 分别标识，小核心，中核心，大核心
	CvmInstanceTypeClass  string            `json:"cvmInstanceTypeClass"`  // 技术分类
}

// GetApproveLogResp get approve log response
type GetApproveLogResp struct {
	RespMeta  `json:",inline"`
	Result    map[string]GetApproveLogOrderRst `json:"result"`
	Errorinfo interface{}                      `json:"errorinfo"`
}

// GetApproveLogOrderRst get approve log result
type GetApproveLogOrderRst []*GetApproveLogItem

// GetApproveLogItem get approve log item
type GetApproveLogItem struct {
	TodoOrderID   string `json:"todoOrderId"`
	OperateTime   string `json:"operateTime"`
	OperateResult string `json:"operateResult"`
	OperateInfo   string `json:"operateInfo"`
	Activity      string `json:"activity"`
	Operator      string `json:"operator"`
	Memo          string `json:"memo"`
	Platform      string `json:"platform"`
	OrderID       string `json:"orderId"`
}

// GetCvmApproveLogsResp get cvm approve logs response
type GetCvmApproveLogsResp struct {
	RespMeta `json:",inline"`
	Result   *CvmApproveLogsRst `json:"result"`
}

// CvmApprovalLog cvm approve log result
type CvmApprovalLog struct {
	TaskNo        int64  `json:"taskNo"`
	TaskName      string `json:"taskName"`
	OperateResult string `json:"operateResult"`
	Operator      string `json:"operator"`
	OperateInfo   string `json:"operateInfo"`
	OperateTime   string `json:"operateTime"`
}

// CvmApproveLogsRst cvm approve log result
type CvmApproveLogsRst struct {
	Data            []CvmApprovalLog `json:"data"`
	CurrentTaskNo   int              `json:"currentTaskNo"`
	CurrentTaskName string           `json:"currentTaskName"`
}

// RevokeCvmOrderResp ...
type RevokeCvmOrderResp struct {
	RespMeta `json:",inline"`
}

// QueryOrderListResp ...
type QueryOrderListResp struct {
	RespMeta `json:",inline"`
	Result   *QueryOrderListRst `json:"result"`
}

// QueryOrderListRst cvm plan order change response
type QueryOrderListRst struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    *QueryOrderListRstData `json:"data"`
}

// QueryOrderListRstData cvm plan order change response data
type QueryOrderListRstData struct {
	Total int               `json:"total"`
	Data  []*QueryOrderInfo `json:"data"`
}

// QueryOrderInfo cvm plan order change response data
type QueryOrderInfo struct {
	OrderID           string                      `json:"orderId"`
	SourceType        int                         `json:"sourceType"`
	SourceTypeName    string                      `json:"sourceTypeName"`
	BgName            string                      `json:"bgName"`
	DeptID            int                         `json:"deptId"`
	DeptName          string                      `json:"deptName"`
	PlanProductName   string                      `json:"planProductName"`
	ToDeptName        string                      `json:"toDeptName"`
	ToPlanProductName string                      `json:"toPlanProductName"`
	UseTime           string                      `json:"useTime"`
	OrderDesc         string                      `json:"orderDesc"`
	Operator          string                      `json:"operator"`
	CreateTime        string                      `json:"createTime"`
	Status            enumor.QueryOrderInfoStatus `json:"status"`
	StatusMsg         string                      `json:"statusMsg"`
	StatusDesc        string                      `json:"statusDesc"`
	CurrentProcessor  string                      `json:"currentProcessor"`
	CvmData           []CvmData                   `json:"cvmData"`
	CbsData           []struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
		Unit  string `json:"unit"`
	} `json:"cbsData"`
	TechData []struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"techData"`
	AllDiskAmount   int    `json:"allDiskAmount"`
	AllCoreAmount   int    `json:"allCoreAmount"`
	DeltaCoreAmount int    `json:"deltaCoreAmount"`
	DeltaDiskAmount int    `json:"deltaDiskAmount"`
	OrderType       int    `json:"orderType"`
	OrderTypeName   string `json:"orderTypeName"`
}

// CvmData cvm plan order change response data
type CvmData struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// TransOrderResp ...
type TransOrderResp struct {
	RespMeta `json:",inline"`
	Result   OrderCreateRst `json:"result"`
}

// TransOrderRst ...
type TransOrderRst struct {
	OrderId string `json:"orderId"`
	Status  int    `json:"status"`
}

// MatchSwapGroupReq CRP匹配交换组请求
type MatchSwapGroupReq struct {
	ReqMeta `json:",inline"`
	Params  MatchSwapGroupParams `json:"params"`
}

// MatchSwapGroupParams ...
type MatchSwapGroupParams struct {
	DeptID       int64  `json:"deptId"`       // 部门ID
	ApplyNum     int    `json:"applyNum"`     // 申领量
	Zone         string `json:"zone"`         // 可用区
	InstanceType string `json:"instanceType"` // 机型
}

// MatchSwapGroupResp CRP匹配交换组响应
type MatchSwapGroupResp struct {
	RespMeta `json:",inline"`
	Result   MatchSwapGroupResult `json:"result"`
}

// MatchSwapGroupResult ...
type MatchSwapGroupResult struct {
	MatchID string `json:"matchId"` // 匹配成功时的单号
	Matched bool   `json:"matched"` // true: 匹配成功, false: 匹配失败
	Msg     string `json:"msg"`     // 消息
}

// QueryMatchTaskReq CRP查询匹配任务请求
type QueryMatchTaskReq struct {
	ReqMeta `json:",inline"`
	Params  QueryMatchTaskParams `json:"params"`
}

// QueryMatchTaskParams ...
type QueryMatchTaskParams struct {
	OrderID string `json:"orderId"` // 匹配单号
}

// QueryMatchTaskResp CRP查询匹配任务响应
type QueryMatchTaskResp struct {
	RespMeta `json:",inline"`
	Result   QueryMatchTaskResult `json:"result"`
}

// QueryMatchTaskResult ...
type QueryMatchTaskResult struct {
	Status       bool     `json:"status"`       // 匹配状态
	InstanceType string   `json:"instanceType"` // 规格
	MaxCutNum    int      `json:"maxCutNum"`    // 最大切片数量
	IPs          []string `json:"ips"`          // 母机IP列表
}

// GetInstanceTypeInfoResp get instance type info response
type GetInstanceTypeInfoResp struct {
	RespMeta `json:",inline"`
	Result   *GetInstanceTypeInfoRst `json:"result"`
}

// GetInstanceTypeInfoRst get instance type info result
type GetInstanceTypeInfoRst struct {
	InstanceTypes []InstanceTypeInfoItem `json:"instanceTypes"` // 可用机型列表
	CpuConditions []interface{}          `json:"cpuConditions"` // CPU核心支持的筛选选项
	RamConditions []interface{}          `json:"ramConditions"` // 内存支持的筛选选项
	CvmConditions []interface{}          `json:"cvmConditions"` // 机型类型的筛选选项
}

// InstanceTypeInfoItem instance type info item
type InstanceTypeInfoItem struct {
	CpuAmount        int     `json:"CpuAmount"`        // CPU核数
	CvmInstanceGroup string  `json:"CvmInstanceGroup"` // 机型大类
	CvmInstanceModel string  `json:"CvmInstanceModel"` // 具体规格
	CvmInstanceType  string  `json:"CvmInstanceType"`  // 机型类别
	Price            float64 `json:"Price"`            // 参考价格
	RamAmount        int     `json:"RamAmount"`        // 内存，单位G
	SysDiskAmount    int     `json:"SysDiskAmount"`    // 系统盘大小
	SellStatus       int     `json:"sellStatus"`       // 是否售卖
	InstanceFamily   string  `json:"instanceFamily"`   // 机型族
}

// LocalDiskTypeInfo local disk type info
type LocalDiskTypeInfo struct {
	Type string `json:"type"` // 类型，ROOT表示系统盘，DATA表示数据盘
	Size int    `json:"size"` // 大小
}
