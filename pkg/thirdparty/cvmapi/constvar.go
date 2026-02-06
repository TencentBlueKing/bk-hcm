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

// Package cvmapi ...
package cvmapi

// CvmCbsPlanModityType 需求预测接口调整类型
var CvmCbsPlanModityType = map[int64]string{
	1: "add",
	2: "delete",
	3: "update",
}

const (
	// CvmId CVM请求ID
	CvmId = "1"
	// CvmJsonRpc CVM请求JSONRPC
	CvmJsonRpc = "2.0"

	// CvmDeptId CVM容量查询部门ID
	CvmDeptId = 1041
	// CvmLaunchDeptName CVM生产时部门名称
	CvmLaunchDeptName = "IEG技术运营部"
	// CvmLaunchProductName CVM运营产品名（项目名）
	CvmLaunchProductName = "互娱资源公共平台"
	// CvmLaunchBiz1Id CVM一级业务ID
	CvmLaunchBiz1Id = 656545
	// CvmLaunchBiz1Name CVM一级业务名
	CvmLaunchBiz1Name = "CC_资源运营服务"
	// CvmLaunchBiz2Id CVM二级业务ID
	CvmLaunchBiz2Id = 656560
	// CvmLaunchBiz2Name CVM二级业务名
	CvmLaunchBiz2Name = "CC_资源运营服务"
	// CvmLaunchBiz3Id CVM三级业务ID
	CvmLaunchBiz3Id = 1073015
	// CvmLaunchBiz3Name CVM三级业务名
	CvmLaunchBiz3Name = "CC_SCR_加工池"
	// CvmLaunchSystemDiskTypePremium CVM生产时系统盘类型，当前固定为高性能云盘
	CvmLaunchSystemDiskTypePremium = "CLOUD_PREMIUM"
	// CvmLaunchSystemDiskTypeBasic CVM生产时系统盘类型，对于固定为本地盘
	CvmLaunchSystemDiskTypeBasic = "LOCAL_BASIC"
	// CvmLaunchSystemDiskSizePremium CVM生产时系统盘大小，当前固定为100G
	CvmLaunchSystemDiskSizePremium = 100
	// CvmLaunchSystemDiskSizeBasic CVM生产时系统盘大小，对于IT设备固定为50G
	CvmLaunchSystemDiskSizeBasic = 50
	// CVM_LAUNCH_USETIME CVM生产时数据盘类型，当前固定为高性能云盘
	CVM_LAUNCH_USETIME = "0000-00-00 00:00:00" // CVM生产时必填项，yuti开发对该字段含义也未知，暂时写死为该固定值0000-00-00 00:00:00
	// CvmLaunchProjectId CVM项目ID，yuti开发对该字段含义也未知，暂时写死为固定值0
	CvmLaunchProjectId = 0
	// CvmOrderLinkPrefix CVM生产单据详情链接前缀
	CvmOrderLinkPrefix = "https://yunti.woa.com/orders/cvm/"
	// CvmReturnLinkPrefix CVM退回单据详情链接前缀
	CvmReturnLinkPrefix = "https://yunti.woa.com/orders/cvmreturn/"
	// CvmPlanLinkPrefix CVM&CBS需求单据详情链接前缀
	CvmPlanLinkPrefix = "https://yunti.woa.com/orders/iaasplan/"
	// CvmUpgradeLinkPrefix CVM升降配单据详情链接前缀
	CvmUpgradeLinkPrefix = "https://crp.woa.com/crp-outside/yunti/orders/cvmadjust/"

	// CvmSeparateCampus 分Campus
	CvmSeparateCampus = "cvm_separate_campus"
	// CvmZoneAll 所有可用区
	CvmZoneAll = "all"

	// CvmApiKey CVM API key
	CvmApiKey = "api_key"
	// CvmApiKeyVal CVM API key value
	CvmApiKeyVal = "octopuskg"

	// CvmCbsPlanQueryId 需求预测查询id
	CvmCbsPlanQueryId = "16318853269804145"
	// CvmCbsPlanAdjustId 需求预测调整id
	CvmCbsPlanAdjustId = "16319322822855177"

	// CvmCbsPlanQueryBgName 需求预测首页查询接口事业群名称
	CvmCbsPlanQueryBgName = "IEG互动娱乐事业群"
	// CvmCbsPlanDeptId 需求预测接口事业群ID
	CvmCbsPlanDeptId = 1041
	// DefaultPlanProductName 需求预测默认规划产品
	DefaultPlanProductName = "互娱运营支撑产品"
	// TransferPlanProductID 转移规划产品ID
	TransferPlanProductID = 4487
	// TransferPlanProductName 转移规划产品名称
	TransferPlanProductName = "IEG资源预测服务"
	// TransferOpProductID 转移运营产品ID
	TransferOpProductID = 9112
	// TransferOpProductName 转移运营产品名称
	TransferOpProductName = "IEG资源预测服务"

	// CvmLaunchMethod cvm methods
	// 创建CVM订单方法
	CvmLaunchMethod = "createCvmOrder"
	// CvmOrderStatusMethod CVM单据进度查询方法
	CvmOrderStatusMethod = "queryOrders"
	// CvmQueryApproveLogMethod CVM审批日志查询方法
	CvmQueryApproveLogMethod = "queryCvmApproveLog"
	// CvmRevokeOrderMethod  CVM撤销单据方法
	CvmRevokeOrderMethod = "revokeOrder"
	// CvmInstanceStatusMethod CVM实例状态查询方法
	CvmInstanceStatusMethod = "queryCVMInstances"
	// CvmCapacityMethod CVM容量查询方法
	CvmCapacityMethod = "queryApplyCapacity"
	// CvmVpcMethod CVM vpc信息查询方法
	CvmVpcMethod = "getVpcInfo"
	// CvmSubnetMethod CVM subnet信息查询方法
	CvmSubnetMethod = "getSubNetInfo"
	// CvmRealSubnetMethod CVM 获取私有子网
	CvmRealSubnetMethod = "getRealSubnetInfo"
	// CvmGetProcessMethod CVM流程查询方法
	CvmGetProcessMethod = "getCVMProcess"
	// GetErpProcessMethod ERP流程查询方法
	GetErpProcessMethod = "getERPProcess"
	// CvmReturnMethod CVM退回提单方法
	CvmReturnMethod = "createCvmReturnOrder"
	// CvmReturnStatusMethod CVM退回单据状态查询方法
	CvmReturnStatusMethod = "queryCvmReturnOrder"
	// CvmReturnDetailMethod 根据单号查询退回CVM方法
	CvmReturnDetailMethod = "queryReturnCvmByOrder"
	// CvmReturnPlanMethod 查询CVM退回计划方法
	CvmReturnPlanMethod = "queryReturnPlanItem"
	// QueryCvmInstanceType 查询CVM机型信息
	QueryCvmInstanceType = "queryCvmInstanceType"
	// GetInstanceTypeInfoMethod 获取可申领的实例机型以及实例机型参数
	GetInstanceTypeInfoMethod = "getInstanceTypeInfo"
	// GetApproveLogMethod 查询审批日志
	GetApproveLogMethod = "getApproveLog"
	// CvmMatchSwapGroupMethod 亲和性匹配请求方法
	CvmMatchSwapGroupMethod = "matchSwapGroup"
	// CvmQueryMatchTaskMethod 亲和性匹配任务查询方法
	CvmQueryMatchTaskMethod = "queryMatchTask"

	// CvmUpgradeMethod CVM升降配提单方法
	CvmUpgradeMethod = "createUpgradeOrder"
	// CvmUpgradeDetailMethod CVM升降配单据明细查询方法
	CvmUpgradeDetailMethod = "queryUpgradeOrderDetail"

	// CvmCbsPlanDefaultCvmDesc 需求预测单据的默认CVM备注，用于管理员判断需求来源
	CvmCbsPlanDefaultCvmDesc = "[From IEG HCM CVM]"
	// CvmCbsPlanDefaultCADesc 需求预测单据的默认CA备注
	CvmCbsPlanDefaultCADesc = "[From IEG HCM CA]"

	// DftImageID default image id of TencentOS Server 2.6 (TK4)
	DftImageID = "img-fjxtfi0n"

	// AdjustTypeAdjust 预测调整类型-常规修改
	AdjustTypeAdjust = "常规修改"
	// AdjustTypeDelay 预测调整类型-加急延期
	AdjustTypeDelay = "加急延期"
	// AdjustTypeCancel 预测调整类型-需求取消
	AdjustTypeCancel = "需求取消"
)

// 资源预测相关方法
const (
	// CvmCbsDemandChangeLogQueryMethod 预测需求的变更记录查询接口
	CvmCbsDemandChangeLogQueryMethod = "queryDemandChangeLogForIEG"
	// CvmCbsPlanOrderChangeMethod 预测需求的订单变更流水接口
	CvmCbsPlanOrderChangeMethod = "queryOrderChangeForIEG"
	// CvmCbsPlanQueryMethod 需求预测首页查询接口
	CvmCbsPlanQueryMethod = "queryCvmCbsInfoForIEG"
	// CvmCbsAdjustAblePlanQueryMethod 可调整预测列表查询接口
	CvmCbsAdjustAblePlanQueryMethod = "queryAdjustAbleDemandList"
	// CvmCbsPlanAdjustMethod 需求预测首页调整接口
	CvmCbsPlanAdjustMethod = "adjustOrder"
	// CvmCbsPlanAutoAdjustMethod 需求预测细粒度调整接口
	CvmCbsPlanAutoAdjustMethod = "submitAutoAdjustOrder"
	// CvmCbsPlanAddMethod 需求预测追加接口
	CvmCbsPlanAddMethod = "addYuntiOrder"
	// CvmCbsPlanOrderQueryMethod 需求单据查询接口
	CvmCbsPlanOrderQueryMethod = "queryYuntiOrder"
	// CvmCbsPlanPenaltyRatioReportMethod 需求预测罚金分摊比例上报接口
	CvmCbsPlanPenaltyRatioReportMethod = "reportForecastPartition"
	// CvmQueryOrderList 根据销毁单据查询预测返还信息
	CvmQueryOrderList = "queryOrderList"
	// CvmCbsPlanTransOrderMethod 需求转移接口
	CvmCbsPlanTransOrderMethod = "transOrder"
)

// CVMCli yunti client options
type CVMCli struct {
	// CvmApiAddr yunti api address
	CvmApiAddr        string `yaml:"host"`
	CvmLaunchPassword string `yaml:"launch_password"`
}

// NewCvmQueryApproveLogReq CVM审批日志查询请求元数据
func NewCvmQueryApproveLogReq(params *GetCvmApproveLogParams) *GetCvmApproveLogReq {
	return &GetCvmApproveLogReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmQueryApproveLogMethod,
		},
		Params: params,
	}
}

// NewOrderQueryReq CVM审批日志查询请求元数据
func NewOrderQueryReq(params *OrderQueryParam) *OrderQueryReq {
	return &OrderQueryReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmOrderStatusMethod,
		},
		Params: params,
	}
}

// NewRevokeCvmOrderReq CVM撤回订单请求元数据
func NewRevokeCvmOrderReq(params *RevokeCvmOrderParams) *RevokeCvmOrderReq {
	return &RevokeCvmOrderReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmRevokeOrderMethod,
		},
		Params: params,
	}
}

// NewCvmUpgradeOrderReq CVM升降配订单创建请求元数据
func NewCvmUpgradeOrderReq(params *UpgradeParam) *UpgradeReq {
	return &UpgradeReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmUpgradeMethod,
		},
		Params: params,
	}
}

// NewCvmUpgradeDetailReq CVM升降配订单查询请求元数据
func NewCvmUpgradeDetailReq(params *UpgradeDetailParam) *UpgradeDetailReq {
	return &UpgradeDetailReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmUpgradeDetailMethod,
		},
		Params: params,
	}
}

// NewQueryReturnPlanReq CVM退回计划查询请求元数据
func NewQueryReturnPlanReq(params *QueryReturnPlanParam) *QueryReturnPlanReq {
	return &QueryReturnPlanReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmReturnPlanMethod,
		},
		Params: params,
	}
}
