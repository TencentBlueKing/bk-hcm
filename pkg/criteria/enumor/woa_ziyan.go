/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package enumor

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"hcm/pkg/tools/converter"
)

// ObsProject is obs project.
type ObsProject string

const (
	// ObsProjectNormal is obs normal project.
	ObsProjectNormal ObsProject = "常规项目"
	// ObsProjectReuse is obs project reuse project.
	ObsProjectReuse ObsProject = "改造复用"
	// ObsProjectMigrate is obs migrate project.
	ObsProjectMigrate ObsProject = "轻量云徙"
	// ObsProjectRollServer is obs roll server project.
	ObsProjectRollServer ObsProject = "滚服项目"
	// ObsProjectShortLease is obs short lease project.
	ObsProjectShortLease ObsProject = "短租项目"
)

// GetObsProjectMembers get ObsProject's members.
func GetObsProjectMembers() []ObsProject {
	obsProjects := []ObsProject{ObsProjectNormal, ObsProjectReuse, ObsProjectMigrate, ObsProjectRollServer}
	obsProjects = append(obsProjects, getSpringObsProject(), getDissolveObsProject())

	return obsProjects
}

// GetObsProjectMembersForResPlan get ObsProject's members for resource plan.
// 顺序为： 常规项目、滚服项目、春节保障、机房裁撤、改造复用、轻量云徙
func GetObsProjectMembersForResPlan() []ObsProject {
	obsProjects := []ObsProject{ObsProjectNormal, ObsProjectShortLease, ObsProjectRollServer}
	obsProjects = append(obsProjects, getSpringObsProjectForResPlan()...)
	obsProjects = append(obsProjects, getDissolveObsProjectForResPlan()...)
	obsProjects = append(obsProjects, []ObsProject{ObsProjectReuse, ObsProjectMigrate}...)

	return obsProjects
}

// Validate ObsProject.
func (o ObsProject) Validate() error {
	obsProjects := GetObsProjectMembers()
	obsProjectMap := converter.SliceToMap(obsProjects, func(obj ObsProject) (ObsProject, struct{}) {
		return obj, struct{}{}
	})

	if _, ok := obsProjectMap[o]; !ok {
		return fmt.Errorf("unsupported obs project: %s", o)
	}

	return nil
}

// ValidateResPlan validate obs project used in resource plan.
func (o ObsProject) ValidateResPlan() error {
	// TODO 临时 支持同步短租项目到本地。待后续正式支持短租项目时去除此独立逻辑
	if o == ObsProjectShortLease {
		return nil
	}

	obsProjects := GetObsProjectMembersForResPlan()
	obsProjectMap := converter.SliceToMap(obsProjects, func(obj ObsProject) (ObsProject, struct{}) {
		return obj, struct{}{}
	})

	if _, ok := obsProjectMap[o]; !ok {
		return fmt.Errorf("unsupported obs project: %s", o)
	}

	return nil
}

// getSpringObsProject get spring obs project.
func getSpringObsProject() ObsProject {
	// 春保窗口期：12月1日～次年3月15日
	// 12月1日～12月31日提单的春保项目前缀为次年
	year := time.Now().Local().Year()
	if time.Now().Month() == time.December {
		year += 1
	}

	prefixYear := strconv.Itoa(year)
	project := ObsProject(prefixYear + "春节保障")

	return project
}

// getDissolveObsProject get dissolve obs project.
func getDissolveObsProject() ObsProject {
	// 按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	year := time.Now().Local().Year()
	prefixYear := strconv.Itoa(year)
	project := ObsProject(prefixYear + "机房裁撤")

	return project
}

// getSpringObsProjectForResPlan get spring obs project for resource plan.
func getSpringObsProjectForResPlan() []ObsProject {
	projects := make([]ObsProject, 0)
	nowYear := time.Now().Year()
	// 春保窗口期：12月1日～次年3月25日

	// 因预测的提前性，1月1日～次年3月25日均允许提次年的春保项目预测单
	projects = append(projects, ObsProject(strconv.Itoa(nowYear+1)+"春节保障"))

	// 因预算提前一年追加，需要允许提后年的春保项目预测（在次年底即可使用，提前一年追加预测）
	projects = append(projects, ObsProject(strconv.Itoa(nowYear+2)+"春节保障"))

	// 3月25日前允许申请当年的春保项目预测单
	ddl := time.Date(nowYear, time.March, 25, 0, 0, 0, 0, time.Local)
	if time.Now().Before(ddl) {
		prefixYear := strconv.Itoa(nowYear)
		projects = append(projects, ObsProject(prefixYear+"春节保障"))
	}

	return projects
}

// getDissolveObsProjectForResPlan get dissolve obs project for resource plan.
func getDissolveObsProjectForResPlan() []ObsProject {
	projects := make([]ObsProject, 0)

	// 按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	nowYear := time.Now().Local().Year()
	prefixYear := strconv.Itoa(nowYear)
	projects = append(projects, ObsProject(prefixYear+"机房裁撤"))
	nextPrefixYear := strconv.Itoa(nowYear + 1)
	projects = append(projects, ObsProject(nextPrefixYear+"机房裁撤"))

	return projects
}

// IsDissolveObsProjectForResPlan is dissolve obs project contains.
func IsDissolveObsProjectForResPlan(obsProject ObsProject) bool {
	return slices.Contains(getDissolveObsProjectForResPlan(), obsProject)
}

// RequireType is resource apply require type.
type RequireType int64

const (
	// RequireTypeRegular 常规项目
	RequireTypeRegular RequireType = 1
	// RequireTypeSpring 春节保障
	RequireTypeSpring RequireType = 2
	// RequireTypeDissolve 机房裁撤
	RequireTypeDissolve RequireType = 3
	// RequireTypeRollServer 滚服项目
	RequireTypeRollServer RequireType = 6
	//	RequireTypeGreenChannel 小额绿通
	RequireTypeGreenChannel RequireType = 7
	// RequireTypeSpringResPool 春保资源池
	RequireTypeSpringResPool RequireType = 8
	// RequireTypeShortLease 短租项目
	RequireTypeShortLease RequireType = 9
)

var requireTypeNameMap = map[RequireType]string{
	RequireTypeRegular:       "常规项目",
	RequireTypeSpring:        "春节保障",
	RequireTypeDissolve:      "机房裁撤",
	RequireTypeRollServer:    "滚服项目",
	RequireTypeGreenChannel:  "小额绿通",
	RequireTypeSpringResPool: "春保资源池",
	RequireTypeShortLease:    "短租项目",
}

// GetName get name of RequireType.
func (t RequireType) GetName() string {
	if name, ok := requireTypeNameMap[t]; ok {
		return name
	}

	return "Unknown"
}

// GetRequireTypeMembers get members of RequireType.
func GetRequireTypeMembers() []RequireType {
	return []RequireType{
		RequireTypeRegular,
		RequireTypeSpring,
		RequireTypeDissolve,
		RequireTypeRollServer,
		RequireTypeGreenChannel,
		RequireTypeSpringResPool,
		RequireTypeShortLease,
	}
}

// Validate RequireType.
func (t RequireType) Validate() error {
	requireTypeMembers := GetRequireTypeMembers()
	requireTypeMemberMap := converter.SliceToMap(requireTypeMembers, func(member RequireType) (RequireType, struct{}) {
		return member, struct{}{}
	})
	if _, ok := requireTypeMemberMap[t]; !ok {
		return fmt.Errorf("unsupported require type: %d", t)
	}

	return nil
}

// NeedVerifyResPlan need verify resource plan.
func (t RequireType) NeedVerifyResPlan() bool {
	switch t {
	// 常规项目、春节保障、机房裁撤、春保资源池、短租项目需要校验预测
	case RequireTypeRegular, RequireTypeSpring, RequireTypeDissolve, RequireTypeSpringResPool, RequireTypeShortLease:
		return true
	default:
		return false
	}
}

// ToObsProject ObsProject.
func (t RequireType) ToObsProject() ObsProject {
	if obsProject, ok := RequireTypeObsProjectMap[t]; ok {
		return obsProject
	}

	// 默认是常规项目
	return ObsProjectNormal
}

// IsNeedQuotaManage 是否在主机申请时使用额度管理
func (t RequireType) IsNeedQuotaManage() bool {
	if t == RequireTypeRollServer || t == RequireTypeSpringResPool {
		return true
	}

	return false
}

// IsUseManageBizPlan 是否使用管理业务的运营产品去申请主机，扣减预测
func (t RequireType) IsUseManageBizPlan() bool {
	if t == RequireTypeRollServer || t == RequireTypeSpringResPool {
		return true
	}

	return false
}

// ToRequireTypeWhenGetDevice 查询机型时使用的需求类型
func (t RequireType) ToRequireTypeWhenGetDevice() RequireType {
	requireTypeMap := map[RequireType]RequireType{
		RequireTypeRegular:    RequireTypeRegular,
		RequireTypeRollServer: RequireTypeRollServer,
		// "小额绿通"使用"常规项目"类型查询机型
		RequireTypeGreenChannel: RequireTypeRegular,
		RequireTypeSpring:       RequireTypeSpring,
		RequireTypeDissolve:     RequireTypeDissolve,
		// "春保资源池"使用"常规项目"类型查询机型
		RequireTypeSpringResPool: RequireTypeRegular,
	}

	requireType, ok := requireTypeMap[t]
	if !ok {
		return t
	}
	return requireType
}

// RequireTypeObsProjectMap 需求类型与 obs project 的映射
var RequireTypeObsProjectMap = map[RequireType]ObsProject{
	RequireTypeRegular:    ObsProjectNormal,
	RequireTypeRollServer: ObsProjectRollServer,
	// "小额绿通"使用"常规项目"的 obs project
	RequireTypeGreenChannel: ObsProjectNormal,
	RequireTypeSpring:       getSpringObsProject(),
	RequireTypeDissolve:     getDissolveObsProject(),
	// "春保资源池"使用"常规项目"的 obs project
	RequireTypeSpringResPool: ObsProjectNormal,
	RequireTypeShortLease:    ObsProjectShortLease,
}

// CrpOrderStatus is crp order status.
type CrpOrderStatus int

const (
	// CrpOrderStatusDeptApprove 部门管理员审批
	CrpOrderStatusDeptApprove CrpOrderStatus = 0
	// CrpOrderStatusDirectorApprove 业务总监审批
	CrpOrderStatusDirectorApprove CrpOrderStatus = 1
	// CrpOrderStatusPlanApprove 规划经理审批
	CrpOrderStatusPlanApprove CrpOrderStatus = 2
	// CrpOrderStatusResourceApprove 资源经理审批
	CrpOrderStatusResourceApprove CrpOrderStatus = 3
	// CrpOrderStatusCloudApprove 等待云上审批
	CrpOrderStatusCloudApprove CrpOrderStatus = 14
	// CrpOrderStatusWaitDeliver 等待交付
	CrpOrderStatusWaitDeliver CrpOrderStatus = 4
	// CrpOrderStatusDelivering 交付队列中
	CrpOrderStatusDelivering CrpOrderStatus = 5
	// CrpOrderStatusResource 资源准备中
	CrpOrderStatusResource CrpOrderStatus = 6
	// CrpOrderStatusCvm CVM 生成中
	CrpOrderStatusCvm CrpOrderStatus = 7
	// CrpOrderStatusFinish 执行完成
	CrpOrderStatusFinish CrpOrderStatus = 8
	// CrpOrderStatusReject 驳回
	CrpOrderStatusReject CrpOrderStatus = 127
	// CrpOrderStatusFailed CVM创建失败
	CrpOrderStatusFailed CrpOrderStatus = 129
)

// StatusName CrpOrderStatus.
func (cs CrpOrderStatus) StatusName() string {
	switch cs {
	case CrpOrderStatusCvm:
		return "CRP-CVM生产中"
	case CrpOrderStatusFinish:
		return "CRP-生产成功"
	case CrpOrderStatusReject:
		return "CRP-驳回"
	case CrpOrderStatusFailed:
		return "CRP-CVM创建失败"
	default:
		return fmt.Sprintf("CRP-unsupported crp order status: %d", cs)
	}
}

// CrpOrderStatusCanRevoke 目前 crp 只支持单据处于下列状态时可以发起撤单
var CrpOrderStatusCanRevoke = []CrpOrderStatus{
	CrpOrderStatusDeptApprove,
	CrpOrderStatusDirectorApprove,
	CrpOrderStatusPlanApprove,
	CrpOrderStatusResourceApprove,
	CrpOrderStatusWaitDeliver,
	CrpOrderStatusDelivering,
}

var crpOrderStatusApprovalStateMap = map[CrpOrderStatus]ApprovalState{
	CrpOrderStatusDeptApprove: HcmAdminApproval,
	CrpOrderStatusPlanApprove: CrpAdminApproval,
}

// IsAdminApproval is admin approval.
func (cs CrpOrderStatus) IsAdminApproval() bool {
	_, ok := crpOrderStatusApprovalStateMap[cs]
	return ok
}

// GetApprovalState get approval state.
func (cs CrpOrderStatus) GetApprovalState() ApprovalState {
	return crpOrderStatusApprovalStateMap[cs]
}

// CrpUpgradeOrderStatus is crp upgrade order status.
type CrpUpgradeOrderStatus int

const (
	// CrpUpgradeOrderDeptApprove 部门管理员审批
	CrpUpgradeOrderDeptApprove CrpUpgradeOrderStatus = 0
	// CrpUpgradeOrderPlanApprove 规划经理审批
	CrpUpgradeOrderPlanApprove CrpUpgradeOrderStatus = 1
	// CrpUpgradeOrderResourceApprove 资源经理审批
	CrpUpgradeOrderResourceApprove CrpUpgradeOrderStatus = 2
	// CrpUpgradeOrderWaitProcess 等待执行
	CrpUpgradeOrderWaitProcess CrpUpgradeOrderStatus = 9
	// CrpUpgradeOrderProcessing 执行中
	CrpUpgradeOrderProcessing CrpUpgradeOrderStatus = 10
	// CrpUpgradeOrderFinish 执行完成
	CrpUpgradeOrderFinish CrpUpgradeOrderStatus = 20
	// CrpUpgradeOrderReject 驳回
	CrpUpgradeOrderReject CrpUpgradeOrderStatus = 127
	// CrpUpgradeOrderFailed 订单失败
	CrpUpgradeOrderFailed CrpUpgradeOrderStatus = 128
)

// StatusName CrpUpgradeOrderStatus.
func (cs CrpUpgradeOrderStatus) StatusName() string {
	switch cs {
	case CrpUpgradeOrderProcessing:
		return "CRP-升降配执行中"
	case CrpUpgradeOrderFinish:
		return "CRP-升降配完成"
	case CrpUpgradeOrderReject:
		return "CRP-驳回"
	case CrpUpgradeOrderFailed:
		return "CRP-升降配失败"
	default:
		return fmt.Sprintf("CRP-unsupported crp order status: %d", cs)
	}
}

// CrpUpgradeCVMStatus is crp upgrade cvm status.
type CrpUpgradeCVMStatus string

const (
	// CrpUpgradeCVMWaiting 待操作
	CrpUpgradeCVMWaiting CrpUpgradeCVMStatus = "WAITING"
	// CrpUpgradeCVMOperating 操作中
	CrpUpgradeCVMOperating CrpUpgradeCVMStatus = "OPERATING"
	// CrpUpgradeCVMSuccess 成功
	CrpUpgradeCVMSuccess CrpUpgradeCVMStatus = "SUCCESS"
	// CrpUpgradeCVMFailed 失败
	CrpUpgradeCVMFailed CrpUpgradeCVMStatus = "FAILED"
)

// CoreType 核心类型
type CoreType string

const (
	// CoreTypeBig 大核心
	CoreTypeBig CoreType = "大核心"
	// CoreTypeMedium 中核心
	CoreTypeMedium CoreType = "中核心"
	// CoreTypeSmall 小核心
	CoreTypeSmall CoreType = "小核心"
)

// Validate CoreType.
func (r CoreType) Validate() error {
	switch r {
	case CoreTypeBig:
	case CoreTypeMedium:
	case CoreTypeSmall:
	default:
		return fmt.Errorf("unsupported verify core type result: %s", r)
	}

	return nil
}

// CRPCoreTypeMap crp core type map
// key为crp侧的值，1.2.3 分别标识，小核心，中核心，大核心
var CRPCoreTypeMap = map[int]CoreType{
	1: CoreTypeSmall,
	2: CoreTypeMedium,
	3: CoreTypeBig,
}

// GetCoreTypeByCRPCoreTypeID 根据 crp 的 coreTypeID 获取 coreType
func GetCoreTypeByCRPCoreTypeID(coreTypeID int) CoreType {
	return CRPCoreTypeMap[coreTypeID]
}

// ItsmServiceNameApply ITSM中的资源申请流程在 HCM 中的名称（此处名称并非与 ITSM 中的流程名称完全一致）
const ItsmServiceNameApply = "资源申领流程"

const (
	// ResourcePoolBiz 资源池业务
	ResourcePoolBiz = 931
	// ResourcePlanRollServerBiz 资源预测提报滚服项目的业务
	ResourcePlanRollServerBiz = 931
)

// AbolishPhase 裁撤阶段
type AbolishPhase string

const (
	// Incomplete 裁撤未完成
	Incomplete AbolishPhase = "incomplete"
	// Complete 裁撤完成
	Complete AbolishPhase = "complete"
	// BsiComplete 业务退回
	BsiComplete AbolishPhase = "bsiComplete"
	// Retain 保留暂不裁撤
	Retain AbolishPhase = "retain"
)

// XrayFaultTicketIsEnd xray故障单是否结单
type XrayFaultTicketIsEnd int

const (
	XrayFaultTicketNotEnd XrayFaultTicketIsEnd = 0
	XrayFaultTicketHasEnd XrayFaultTicketIsEnd = 1
)

// InstanceStatus cvm instance status
type InstanceStatus string

const (
	// CvmInstanceStatusRunning CVM实例-运行中
	CvmInstanceStatusRunning = "RUNNING"
)

// CvmModifyRecordStatus is cvm modify record status
type CvmModifyRecordStatus int

const (
	// ApprovalPendingCvmModifyStatus CVM变更记录-待审批
	ApprovalPendingCvmModifyStatus CvmModifyRecordStatus = 0
	// ApprovedCvmModifyStatus CVM变更记录-审批通过
	ApprovedCvmModifyStatus CvmModifyRecordStatus = 1
	// ApprovalFailedCvmModifyStatus CVM变更记录-审批失败
	ApprovalFailedCvmModifyStatus CvmModifyRecordStatus = 2
	// ApprovalRejectedCvmModifyStatus CVM变更记录-审批拒绝
	ApprovalRejectedCvmModifyStatus CvmModifyRecordStatus = 3
	// ApprovalTimeoutCvmModifyStatus CVM变更记录-审批超时/作废
	ApprovalTimeoutCvmModifyStatus CvmModifyRecordStatus = 4
)

// CvmModifyRecordStatusMap cvm modify record status map
var CvmModifyRecordStatusMap = map[CvmModifyRecordStatus]string{
	ApprovalPendingCvmModifyStatus:  "待审批",
	ApprovedCvmModifyStatus:         "审批通过",
	ApprovalFailedCvmModifyStatus:   "审批失败",
	ApprovalRejectedCvmModifyStatus: "审批拒绝",
	ApprovalTimeoutCvmModifyStatus:  "审批超时",
}

// ResAssign 资源分配方式（1表示“有资源区域优先”、2表示“分Campus生产”）
type ResAssign uint

const (
	// ResPriorityResAssign 有资源区域优先
	ResPriorityResAssign ResAssign = 1
	// CampusResAssign 分Campus生产
	CampusResAssign ResAssign = 2
)

// Validate ResAssign.
func (r ResAssign) Validate() error {
	switch r {
	case ResPriorityResAssign:
	case CampusResAssign:
	default:
		return fmt.Errorf("unsupported verify res assign result: %d", r)
	}

	return nil
}

// QueryOrderInfoStatus 根据销毁单据查询预测返还信息接口状态
type QueryOrderInfoStatus int

const (
	// QueryOrderInfoStatusSuccess 根据销毁单据查询预测返还信息接口 - 成功状态
	QueryOrderInfoStatusSuccess = 0
)

// HostTransitStatus 主机转移状态
type HostTransitStatus string

const (
	// HostTransitStatusCompleted 转移完成
	HostTransitStatusCompleted HostTransitStatus = "completed"
	// HostTransitStatusTransitToOrigin 需要转移到原始业务
	HostTransitStatusTransitToOrigin HostTransitStatus = "transit_to_origin"
	// HostTransitStatusTransitToReborn 需要转移到reborn业务
	HostTransitStatusTransitToReborn HostTransitStatus = "transit_to_reborn"
	// HostTransitStatusAbnormal 异常情况
	HostTransitStatusAbnormal HostTransitStatus = "abnormal"
)

// ApprovalState defines the approval state
type ApprovalState string

const (
	// LeaderApproval defines the leader approval state
	LeaderApproval ApprovalState = "leader_approval"
	// HcmAdminApproval defines the hcm admin approval state
	HcmAdminApproval ApprovalState = "hcm_admin_approval"
	// CrpAdminApproval defines the crp admin approval state
	CrpAdminApproval ApprovalState = "crp_admin_approval"
)

// Validate validates the approve state
func (a ApprovalState) Validate() error {
	if _, ok := approveStateMap[a]; !ok {
		return fmt.Errorf("invalid approval state: %s", a)
	}

	return nil
}

var approveStateMap = map[ApprovalState]string{
	LeaderApproval:   "直属leader审批",
	HcmAdminApproval: "HCM系统管理员审批",
	CrpAdminApproval: "CRP系统管理员审批",
}

// HostApplyItsmStepName defines the host apply itsm step name
type HostApplyItsmStepName string

const (
	// HostApplyItsmStepNameSpecifiedUserApproval defines the host apply specified user approval step name
	HostApplyItsmStepNameSpecifiedUserApproval HostApplyItsmStepName = "指定审批人审批"
	// HostApplyItsmStepNameLeaderApproval defines the host apply itsm leader approval step name
	HostApplyItsmStepNameLeaderApproval HostApplyItsmStepName = "直属Leader审批"
	// HostApplyItsmStepNameAdminApproval defines the host apply itsm admin approval step name
	HostApplyItsmStepNameAdminApproval HostApplyItsmStepName = "管理员审批"
)

// Validate validates the host apply itsm step name
func (h HostApplyItsmStepName) Validate() error {
	if _, ok := hostApplyItsmStepNameApproveStateMap[h]; !ok {
		return fmt.Errorf("invalid host apply itsm step name: %s", h)
	}

	return nil
}

var hostApplyItsmStepNameApproveStateMap = map[HostApplyItsmStepName]ApprovalState{
	HostApplyItsmStepNameSpecifiedUserApproval: LeaderApproval,
	HostApplyItsmStepNameLeaderApproval:        LeaderApproval,
	HostApplyItsmStepNameAdminApproval:         HcmAdminApproval,
}

// GetApprovalState get the approval state
func (h HostApplyItsmStepName) GetApprovalState() ApprovalState {
	return hostApplyItsmStepNameApproveStateMap[h]
}

// ApplyTicketSource apply ticket source
type ApplyTicketSource string

const (
	// ApplyTicketSrcBusiness business
	ApplyTicketSrcBusiness ApplyTicketSource = "business"
	// ApplyTicketSrcPurchaseToResPool purchase to resource pool
	ApplyTicketSrcPurchaseToResPool ApplyTicketSource = "purchase_to_resource_pool"
)

// GetAllApplyTicketSource get all apply ticket source
func GetAllApplyTicketSource() []ApplyTicketSource {
	return []ApplyTicketSource{ApplyTicketSrcBusiness, ApplyTicketSrcPurchaseToResPool}
}
