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

import "fmt"

const (
	// TicketSvcNameResPlan 资源预测在ITSM的服务
	TicketSvcNameResPlan = "res_plan"
	// TicketNodeNameCrpAudit 资源预测在ITSM流程中的CRP审批节点
	TicketNodeNameCrpAudit = "crp_audit"
	// TicketOperatorNameCrpAudit 资源预测在ITSM流程中的CRP审批节点操作人
	TicketOperatorNameCrpAudit = "icr"
	// AuditFlowTimeoutDay 审批流超时时间，单位天
	AuditFlowTimeoutDay int = 28
	// PendingTicketTraceDay 带处理的单据历史追溯时间，单位天
	PendingTicketTraceDay int = 42
)

// RPTicketType is resource plan ticket type.
type RPTicketType string

const (
	// RPTicketTypeAdd is resource plan ticket status add.
	RPTicketTypeAdd RPTicketType = "add"
	// RPTicketTypeAdjust is resource plan ticket status adjust.
	RPTicketTypeAdjust RPTicketType = "adjust"
	// RPTicketTypeDelete is resource plan ticket status delete.
	RPTicketTypeDelete RPTicketType = "delete"
	// RPTicketTypeDelay is resource plan ticket status delay. Only used in sub_ticket.
	// 该类型不返回给前端展示，仅用于内部逻辑，汇总不影响预算，可以直接发起调整的需求
	RPTicketTypeDelay RPTicketType = "delay"
	// RPTicketTypeTransfer is resource plan ticket status transfer. Only used in sub_ticket.
	RPTicketTypeTransfer RPTicketType = "transfer"

	// RPTicketTypeTransferIN is resource plan ticket status transfer in.
	// Will be modified to RPTicketTypeTransfer in sub_ticket create stage.
	RPTicketTypeTransferIN RPTicketType = "transfer_in"
	// RPTicketTypeTransferOUT is resource plan ticket status transfer out.
	// Will be modified to RPTicketTypeTransfer in sub_ticket create stage.
	RPTicketTypeTransferOUT RPTicketType = "transfer_out"

	// RPTicketTypeAutomaticTransfer is resource plan ticket type automatic transfer.
	RPTicketTypeAutomaticTransfer RPTicketType = "automatic_transfer"
)

// Validate RPTicketType.
func (t RPTicketType) Validate() error {
	switch t {
	case RPTicketTypeAdd, RPTicketTypeAdjust, RPTicketTypeDelay, RPTicketTypeDelete, RPTicketTypeTransfer,
		RPTicketTypeAutomaticTransfer:
	default:
		return fmt.Errorf("unsupported resource plan type: %s", t)
	}

	return nil
}

// rdTicketTypeNameMap records RPTicketType's name.
var rdTicketTypeNameMap = map[RPTicketType]string{
	RPTicketTypeAdd:      "新增",
	RPTicketTypeAdjust:   "调整",
	RPTicketTypeDelay:    "延期",
	RPTicketTypeDelete:   "取消",
	RPTicketTypeTransfer: "转移",
}

// Name return RPTicketType's name.
func (t RPTicketType) Name() string {
	return rdTicketTypeNameMap[t]
}

// CanMerged returns whether the demand can be merged to a sub ticket.
func (t RPTicketType) CanMerged() bool {
	return t == RPTicketTypeAdd || t == RPTicketTypeDelete
}

// GetRPTicketTypeMembers get RPTicketType's members for root tickets.
func GetRPTicketTypeMembers() []RPTicketType {
	return []RPTicketType{
		RPTicketTypeAdd,
		RPTicketTypeAdjust,
		RPTicketTypeDelete,
	}
}

// GetPRSubTicketTypeMembers get RPTicketType's members for sub_tickets.
func GetPRSubTicketTypeMembers() []RPTicketType {
	return []RPTicketType{
		RPTicketTypeAdd,
		RPTicketTypeAdjust,
		RPTicketTypeDelay,
		RPTicketTypeDelete,
		RPTicketTypeTransfer,
	}
}

// RPTicketStatus is resource plan ticket status.
type RPTicketStatus string

const (
	// RPTicketStatusInit is resource plan ticket status init.
	RPTicketStatusInit RPTicketStatus = "init"
	// RPTicketStatusAuditing is resource plan ticket status auditing.
	RPTicketStatusAuditing RPTicketStatus = "auditing"
	// RPTicketStatusRejected is resource plan ticket status rejected.
	RPTicketStatusRejected RPTicketStatus = "rejected"
	// RPTicketStatusPartialRejected is resource plan ticket status partial_rejected.
	RPTicketStatusPartialRejected RPTicketStatus = "partial_rejected"
	// RPTicketStatusDone is resource plan ticket status done.
	RPTicketStatusDone RPTicketStatus = "done"
	// RPTicketStatusFailed is resource plan ticket status failed.
	RPTicketStatusFailed RPTicketStatus = "failed"
	// RPTicketStatusPartialFailed is resource plan ticket status partial_failed.
	RPTicketStatusPartialFailed RPTicketStatus = "partial_failed"
	// RPTicketStatusRevoked is resource plan ticket status revoked.
	RPTicketStatusRevoked RPTicketStatus = "revoked"
	// RPTicketStatusTerminated is resource plan ticket status terminated.
	RPTicketStatusTerminated RPTicketStatus = "terminated"
)

// Validate RPTicketStatus.
func (s RPTicketStatus) Validate() error {
	switch s {
	case RPTicketStatusInit:
	case RPTicketStatusAuditing:
	case RPTicketStatusRejected:
	case RPTicketStatusPartialRejected:
	case RPTicketStatusDone:
	case RPTicketStatusFailed:
	case RPTicketStatusPartialFailed:
	case RPTicketStatusRevoked:
	case RPTicketStatusTerminated:
	default:
		return fmt.Errorf("unsupported resource plan status: %s", s)
	}

	return nil
}

// IsUnfinished returns whether the status is unfinished.
func (s RPTicketStatus) IsUnfinished() bool {
	switch s {
	case RPTicketStatusInit:
	case RPTicketStatusAuditing:
	default:
		return false
	}
	return true
}

// rdTicketStatusNameMap records RPTicketStatus's name.
var rdTicketStatusNameMap = map[RPTicketStatus]string{
	RPTicketStatusInit:            "待审批",
	RPTicketStatusAuditing:        "审批中",
	RPTicketStatusRejected:        "审批拒绝",
	RPTicketStatusPartialRejected: "部分拒绝",
	RPTicketStatusDone:            "成功",
	RPTicketStatusFailed:          "失败",
	RPTicketStatusPartialFailed:   "部分失败",
	RPTicketStatusRevoked:         "已撤销",
	RPTicketStatusTerminated:      "已终止",
}

// Name return RPTicketStatus's name.
func (s RPTicketStatus) Name() string {
	return rdTicketStatusNameMap[s]
}

// GetRPTicketStatusMembers get RPTicketStatus's members.
func GetRPTicketStatusMembers() []RPTicketStatus {
	return []RPTicketStatus{
		RPTicketStatusInit,
		RPTicketStatusAuditing,
		RPTicketStatusRejected,
		RPTicketStatusPartialRejected,
		RPTicketStatusDone,
		RPTicketStatusFailed,
		RPTicketStatusPartialFailed,
		RPTicketStatusRevoked,
		RPTicketStatusTerminated,
	}
}

// RPSubTicketStatus is resource plan sub_ticket status.
type RPSubTicketStatus string

const (
	// RPSubTicketStatusInit is resource plan sub_ticket status init.
	RPSubTicketStatusInit RPSubTicketStatus = "init"
	// RPSubTicketStatusWaiting is resource plan sub_ticket status waiting.
	RPSubTicketStatusWaiting RPSubTicketStatus = "waiting"
	// RPSubTicketStatusAuditing is resource plan sub_ticket status auditing.
	RPSubTicketStatusAuditing RPSubTicketStatus = "auditing"
	// RPSubTicketStatusRejected is resource plan sub_ticket status rejected.
	RPSubTicketStatusRejected RPSubTicketStatus = "rejected"
	// RPSubTicketStatusDone is resource plan sub_ticket status done.
	RPSubTicketStatusDone RPSubTicketStatus = "done"
	// RPSubTicketStatusFailed is resource plan sub_ticket status failed.
	RPSubTicketStatusFailed RPSubTicketStatus = "failed"
	// RPSubTicketStatusInvalid is resource plan sub_ticket status invalid.
	RPSubTicketStatusInvalid RPSubTicketStatus = "invalid"
)

// Validate RPSubTicketStatus.
func (s RPSubTicketStatus) Validate() error {
	switch s {
	case RPSubTicketStatusInit:
	case RPSubTicketStatusWaiting:
	case RPSubTicketStatusAuditing:
	case RPSubTicketStatusRejected:
	case RPSubTicketStatusDone:
	case RPSubTicketStatusFailed:
	case RPSubTicketStatusInvalid:
	default:
		return fmt.Errorf("unsupported resource plan sub_ticket status: %s", s)
	}

	return nil
}

// IsUnfinished return true if RPSubTicketStatus is unfinished.
func (s RPSubTicketStatus) IsUnfinished() bool {
	switch s {
	case RPSubTicketStatusInit:
	case RPSubTicketStatusWaiting:
	case RPSubTicketStatusAuditing:
	default:
		return false
	}
	return true
}

// rpSubTicketStatusNameMap records RPSubTicketStatus's name.
var rpSubTicketStatusNameMap = map[RPSubTicketStatus]string{
	RPSubTicketStatusInit:     "待审批",
	RPSubTicketStatusWaiting:  "等待中",
	RPSubTicketStatusAuditing: "审批中",
	RPSubTicketStatusRejected: "审批拒绝",
	RPSubTicketStatusDone:     "成功",
	RPSubTicketStatusFailed:   "失败",
	RPSubTicketStatusInvalid:  "已失效",
}

// Name return RPTicketStatus's name.
func (s RPSubTicketStatus) Name() string {
	return rpSubTicketStatusNameMap[s]
}

// GetRPSubTicketStatusMembers get RPSubTicketStatus's members.
func GetRPSubTicketStatusMembers() []RPSubTicketStatus {
	return []RPSubTicketStatus{
		RPSubTicketStatusInit,
		// RPSubTicketStatusWaiting 仅用于内部状态流转，不对外暴露
		RPSubTicketStatusAuditing,
		RPSubTicketStatusRejected,
		RPSubTicketStatusDone,
		RPSubTicketStatusFailed,
		RPSubTicketStatusInvalid,
	}
}

// RPSubTicketStage is resource plan sub_ticket stage.
type RPSubTicketStage string

const (
	// RPSubTicketStageInit is resource plan sub_ticket stage init.
	RPSubTicketStageInit RPSubTicketStage = "init"
	// RPSubTicketStageAdminAudit is resource plan sub_ticket stage admin audit.
	RPSubTicketStageAdminAudit RPSubTicketStage = "admin_audit"
	// RPSubTicketStageCRPAudit is resource plan sub_ticket stage crp audit.
	RPSubTicketStageCRPAudit RPSubTicketStage = "crp_audit"
)

// Validate RPSubTicketStage.
func (s RPSubTicketStage) Validate() error {
	switch s {
	case RPSubTicketStageInit:
	case RPSubTicketStageAdminAudit:
	case RPSubTicketStageCRPAudit:
	default:
		return fmt.Errorf("unsupported resource plan sub_ticket stage: %s", s)
	}

	return nil
}

// RPAdminAuditStatus is resource plan admin audit status.
type RPAdminAuditStatus string

const (
	// RPAdminAuditStatusAuditing is resource plan admin audit status auditing.
	RPAdminAuditStatusAuditing RPAdminAuditStatus = "auditing"
	// RPAdminAuditStatusRejected is resource plan admin audit status rejected.
	RPAdminAuditStatusRejected RPAdminAuditStatus = "rejected"
	// RPAdminAuditStatusDone is resource plan admin audit status done.
	RPAdminAuditStatusDone RPAdminAuditStatus = "done"
	// RPAdminAuditStatusSkip is resource plan admin audit status skip.
	RPAdminAuditStatusSkip RPAdminAuditStatus = "skip"
)

// Validate RPSubTicketStage.
func (s RPAdminAuditStatus) Validate() error {
	switch s {
	case RPAdminAuditStatusAuditing:
	case RPAdminAuditStatusRejected:
	case RPAdminAuditStatusDone:
	case RPAdminAuditStatusSkip:
	default:
		return fmt.Errorf("unsupported resource plan admin audit status: %s", s)
	}

	return nil
}

// IsFinished return true if RPAdminAuditStatus is finished.
func (s RPAdminAuditStatus) IsFinished() bool {
	switch s {
	case RPAdminAuditStatusSkip:
	case RPAdminAuditStatusRejected:
	case RPAdminAuditStatusDone:
	default:
		return false
	}
	return true
}

// DemandClass is resource plan demand class.
type DemandClass string

const (
	// DemandClassCVM is demand class cvm.
	DemandClassCVM DemandClass = "CVM"
	// DemandClassCA is demand class ca.
	DemandClassCA DemandClass = "CA"
)

// Validate DemandClass.
func (c DemandClass) Validate() error {
	switch c {
	case DemandClassCVM:
	case DemandClassCA:
	default:
		return fmt.Errorf("unsupported demand class: %s", c)
	}

	return nil
}

// GetDemandClassMembers get DemandClass's members.
func GetDemandClassMembers() []DemandClass {
	return []DemandClass{DemandClassCVM, DemandClassCA}
}

// DemandResType is resource plan demand resource type.
type DemandResType string

const (
	// DemandResTypeCVM is demand resource type cvm.
	DemandResTypeCVM DemandResType = "CVM"
	// DemandResTypeCBS is demand resource type cbs.
	DemandResTypeCBS DemandResType = "CBS"
)

// Validate DemandResType.
func (t DemandResType) Validate() error {
	switch t {
	case DemandResTypeCVM:
	case DemandResTypeCBS:
	default:
		return fmt.Errorf("unsupported demand resource type: %s", t)
	}

	return nil
}

// ResModeCode is resource plan res mode code.
type ResModeCode string

const (
	ResModeCodeByDeviceType   ResModeCode = "device_type"
	ResModeCodeByDeviceFamily ResModeCode = "device_family"
)

// Validate ResModeCode
func (r ResModeCode) Validate() error {
	switch r {
	case ResModeCodeByDeviceType:
	case ResModeCodeByDeviceFamily:
	default:
		return fmt.Errorf("unsupported res mode code: %s", r)
	}

	return nil
}

// ResMode is resource plan resource mode.
type ResMode string

const (
	// ResModeByDeviceType is resource mode of by device type.
	ResModeByDeviceType ResMode = "按机型"
	// ResModeByDeviceFamily is resource mode of by device family.
	ResModeByDeviceFamily ResMode = "按机型族"
)

// Validate ResMode.
func (r ResMode) Validate() error {
	switch r {
	case ResModeByDeviceType:
	case ResModeByDeviceFamily:
	default:
		return fmt.Errorf("unsupported res mode: %s", r)
	}

	return nil
}

// GetResModeMembers get ResMode's members.
func GetResModeMembers() []ResMode {
	return []ResMode{ResModeByDeviceType, ResModeByDeviceFamily}
}

var resModeNameMap = map[ResModeCode]ResMode{
	ResModeCodeByDeviceType:   ResModeByDeviceType,
	ResModeCodeByDeviceFamily: ResModeByDeviceFamily,
}

// Name get ResModeCode's name.
func (r ResModeCode) Name() ResMode {
	return resModeNameMap[r]
}

// Code get ResMode's code.
func (r ResMode) Code() (ResModeCode, error) {
	for code, n := range resModeNameMap {
		if n == r {
			return code, nil
		}
	}
	return "", fmt.Errorf("unsupported res mode: %s", r)
}

// DemandSource is demand source.
// TODO this enum will be changed to get from obs api.
type DemandSource string

const (
	// DemandSourceIndChg is demand source indicator changes.
	DemandSourceIndChg DemandSource = "指标变化"
	// DemandSourceArchAdj is demand source architecture adjustment.
	DemandSourceArchAdj DemandSource = "架构调整"
	// DemandSourceOpt is demand source cost optimization and increased usage rate.
	DemandSourceOpt DemandSource = "成本优化&利用率提升"
	// DemandSourceBufferAdj is demand source buffer adjustment.
	DemandSourceBufferAdj DemandSource = "Buffer调整"
	// DemandSourceForced is demand source forced.
	DemandSourceForced DemandSource = "不可抗力(政策/合规)"
	// DemandSourceSupply is demand source supply issues.
	DemandSourceSupply DemandSource = "供应问题"
	// DemandSourceSysProcess is demand source system process issues.
	DemandSourceSysProcess DemandSource = "系统流程问题"
)

// Validate DemandSource.
func (d DemandSource) Validate() error {
	switch d {
	case DemandSourceIndChg:
	case DemandSourceArchAdj:
	case DemandSourceOpt:
	case DemandSourceBufferAdj:
	case DemandSourceForced:
	case DemandSourceSupply:
	case DemandSourceSysProcess:
	default:
		return fmt.Errorf("unsupported demand source: %s", d)
	}

	return nil
}

// GetDemandSourceMembers get DemandSource's members.
func GetDemandSourceMembers() []DemandSource {
	return []DemandSource{
		DemandSourceIndChg,
		DemandSourceArchAdj,
		DemandSourceOpt,
		DemandSourceBufferAdj,
		DemandSourceForced,
		DemandSourceSupply,
		DemandSourceSysProcess,
	}
}

// CrpDemandLockStatus is resource plan crp demand lock status.
type CrpDemandLockStatus int8

const (
	// CrpDemandUnLocked is resource plan crp demand unlocked.
	CrpDemandUnLocked CrpDemandLockStatus = 0
	// CrpDemandLocked is resource plan crp demand locked.
	CrpDemandLocked CrpDemandLockStatus = 1
)

// Validate CrpDemandLockStatus.
func (s CrpDemandLockStatus) Validate() error {
	switch s {
	case CrpDemandUnLocked:
	case CrpDemandLocked:
	default:
		return fmt.Errorf("unsupported crp demand lock status: %d", s)
	}

	return nil
}

// RPDemandAdjustType is resource plan demand adjust type.
type RPDemandAdjustType string

const (
	// RPDemandAdjustTypeUpdate is resource plan demand adjust type update.
	RPDemandAdjustTypeUpdate RPDemandAdjustType = "update"
	// RPDemandAdjustTypeDelay is resource plan demand adjust type delay.
	RPDemandAdjustTypeDelay RPDemandAdjustType = "delay"
)

// Validate RPDemandAdjustType.
func (t RPDemandAdjustType) Validate() error {
	switch t {
	case RPDemandAdjustTypeUpdate:
	case RPDemandAdjustTypeDelay:
	default:
		return fmt.Errorf("unsupported resource plan demand adjust type: %s", t)
	}

	return nil
}

// CrpAdjustType crp adjust type.
type CrpAdjustType string

const (
	// CrpAdjustTypeUpdate is crp adjust type update.
	CrpAdjustTypeUpdate CrpAdjustType = "常规修改"
	// CrpAdjustTypeDelay is crp adjust type delay.
	CrpAdjustTypeDelay CrpAdjustType = "加急延期"
	// CrpAdjustTypeCancel is crp adjust type cancel.
	CrpAdjustTypeCancel CrpAdjustType = "需求取消"
	// CrpAdjustTypeTransfer is crp adjust type transfer.
	CrpAdjustTypeTransfer CrpAdjustType = "转移"
)

// DemandStatus is resource plan demand status.
type DemandStatus string

const (
	// DemandStatusCanApply 预测需求可申领.
	DemandStatusCanApply DemandStatus = "can_apply"
	// DemandStatusNotReady 预测需求未到申领时间.
	DemandStatusNotReady DemandStatus = "not_ready"
	// DemandStatusExpired 预测需求已过期.
	DemandStatusExpired DemandStatus = "expired"
	// DemandStatusSpentAll 预测需求已耗尽.
	DemandStatusSpentAll DemandStatus = "spent_all"
	// DemandStatusLocked 预测需求变更中.
	DemandStatusLocked DemandStatus = "locked"
)

// Validate DemandStatus.
func (d DemandStatus) Validate() error {
	switch d {
	case DemandStatusCanApply:
	case DemandStatusNotReady:
	case DemandStatusExpired:
	case DemandStatusSpentAll:
	case DemandStatusLocked:
	default:
		return fmt.Errorf("unsupported demand status: %s", d)
	}

	return nil
}

var demandStatusNameMaps = map[DemandStatus]string{
	DemandStatusCanApply: "可申领",
	DemandStatusNotReady: "未到申领时间",
	DemandStatusExpired:  "已过期",
	DemandStatusSpentAll: "已耗尽",
	DemandStatusLocked:   "变更中",
}

// Name return the name of DemandStatus.
func (d DemandStatus) Name() string {
	return demandStatusNameMaps[d]
}

// PlanTypeCode is resource plan type code.
type PlanTypeCode string

const (
	// PlanTypeCodeInPlan is in plan.
	PlanTypeCodeInPlan PlanTypeCode = "in_plan"
	// PlanTypeCodeOutPlan is out plan.
	PlanTypeCodeOutPlan PlanTypeCode = "out_plan"
	// PlanTypeCodeIgnore will ignore plan type code when matching.
	PlanTypeCodeIgnore PlanTypeCode = ""
)

// Validate PlanTypeCode.
func (p PlanTypeCode) Validate() error {
	switch p {
	case PlanTypeCodeInPlan:
	case PlanTypeCodeOutPlan:
	default:
		return fmt.Errorf("unsupported plan type code: %s", p)
	}

	return nil
}

// GetPlanTypeCodeHcmMembers get hcm PlanTypeCode's members.
func GetPlanTypeCodeHcmMembers() []PlanTypeCode {
	return []PlanTypeCode{PlanTypeCodeOutPlan, PlanTypeCodeInPlan}
}

// PlanTypeMaps is plan type maps.
var PlanTypeMaps = map[PlanTypeCode]PlanType{
	PlanTypeCodeInPlan:  PlanTypeHcmInPlan,
	PlanTypeCodeOutPlan: PlanTypeHcmOutPlan,
}

// Name return the name of PlanTypeCode.
func (p PlanTypeCode) Name() PlanType {
	return PlanTypeMaps[p]
}

// InPlan return true if the plan type is in plan.
func (p PlanTypeCode) InPlan() bool {
	switch p {
	case PlanTypeCodeInPlan:
		return true
	case PlanTypeCodeOutPlan:
		return false
	default:
		return false
	}
}

// PlanType is resource plan type.
type PlanType string

const (
	// PlanTypeCrpInPlan is crp in plan.
	PlanTypeCrpInPlan PlanType = "计划内"
	// PlanTypeCrpOutPlan is crp out plan.
	PlanTypeCrpOutPlan PlanType = "计划外"
	// PlanTypeHcmInPlan is hcm in plan.
	PlanTypeHcmInPlan PlanType = "预测内"
	// PlanTypeHcmOutPlan is hcm out plan.
	PlanTypeHcmOutPlan PlanType = "预测外"
)

// Validate PlanType.
func (p PlanType) Validate() error {
	switch p {
	case PlanTypeCrpInPlan:
	case PlanTypeCrpOutPlan:
	case PlanTypeHcmInPlan:
	case PlanTypeHcmOutPlan:
	default:
		return fmt.Errorf("unsupported plan type: %s", p)
	}

	return nil
}

// PlanTypeNameMaps is plan type name maps.
var PlanTypeNameMaps = map[PlanType]PlanTypeCode{
	PlanTypeCrpInPlan:  PlanTypeCodeInPlan,
	PlanTypeCrpOutPlan: PlanTypeCodeOutPlan,
	PlanTypeHcmInPlan:  PlanTypeCodeInPlan,
	PlanTypeHcmOutPlan: PlanTypeCodeOutPlan,
}

// GetCode get plan type code.
func (p PlanType) GetCode() PlanTypeCode {
	return PlanTypeNameMaps[p]
}

// GetPlanTypeHcmMembers get hcm PlanType's members.
func GetPlanTypeHcmMembers() []PlanType {
	return []PlanType{PlanTypeHcmInPlan, PlanTypeHcmOutPlan}
}

// ToAnotherPlanType the plan type of crp to the plan type of hcm, or vice versa.
// TODO: 这个方法的功能是不明确的，使用者难以确定什么情况下使用该方法
func (p PlanType) ToAnotherPlanType() PlanType {
	switch p {
	case PlanTypeCrpInPlan:
		return PlanTypeHcmInPlan
	case PlanTypeCrpOutPlan:
		return PlanTypeHcmOutPlan
	case PlanTypeHcmInPlan:
		return PlanTypeCrpInPlan
	case PlanTypeHcmOutPlan:
		return PlanTypeCrpOutPlan
	default:
		return p
	}
}

// InPlan return the plan type in plan or not.
func (p PlanType) InPlan() bool {
	switch p {
	case PlanTypeCrpInPlan, PlanTypeHcmInPlan:
		return true
	case PlanTypeCrpOutPlan, PlanTypeHcmOutPlan:
		return false
	default:
		return false
	}
}

// GetCrpConsumeResPlanSourceTypes get crp system source types of consuming resource plan.
func GetCrpConsumeResPlanSourceTypes() []string {
	return []string{"申领划扣", "申领划扣退回"}
}

// VerifyResPlanRst is verify resource plan result.
type VerifyResPlanRst string

const (
	// VerifyResPlanRstPass is resource plan result pass.
	VerifyResPlanRstPass VerifyResPlanRst = "PASS"
	// VerifyResPlanRstFailed is resource plan result failed.
	VerifyResPlanRstFailed VerifyResPlanRst = "FAILED"
	// VerifyResPlanRstNotInvolved is resource plan result not involved.
	VerifyResPlanRstNotInvolved VerifyResPlanRst = "NOT_INVOLVED"
)

// Validate VerifyResPlanRst.
func (r VerifyResPlanRst) Validate() error {
	switch r {
	case VerifyResPlanRstPass:
	case VerifyResPlanRstFailed:
	case VerifyResPlanRstNotInvolved:
	default:
		return fmt.Errorf("unsupported verify resource plan result: %s", r)
	}

	return nil
}

// DemandPenaltyBaseSource is demand penalty base source.
type DemandPenaltyBaseSource string

const (
	DemandPenaltyBaseSourceLocal DemandPenaltyBaseSource = "local"
	DemandPenaltyBaseSourceCrp   DemandPenaltyBaseSource = "crp"
)

// Validate DemandPenaltyBaseSource.
func (d DemandPenaltyBaseSource) Validate() error {
	switch d {
	case DemandPenaltyBaseSourceLocal:
	case DemandPenaltyBaseSourceCrp:
	default:
		return fmt.Errorf("unsupported demand penalty base source: %s", d)
	}

	return nil
}

type DemandChangelogType string

const (
	DemandChangelogTypeAppend DemandChangelogType = "append"
	DemandChangelogTypeAdjust DemandChangelogType = "adjust"
	DemandChangelogTypeDelete DemandChangelogType = "delete"
	DemandChangelogTypeExpend DemandChangelogType = "expend"
)

// Validate DemandChangelogType.
func (d DemandChangelogType) Validate() error {
	switch d {
	case DemandChangelogTypeAppend:
	case DemandChangelogTypeAdjust:
	case DemandChangelogTypeDelete:
	case DemandChangelogTypeExpend:
	default:
		return fmt.Errorf("unsupported demand changelog type: %s", d)
	}

	return nil
}

var demandChangelogTypeNameMap = map[DemandChangelogType]string{
	DemandChangelogTypeAppend: "追加",
	DemandChangelogTypeAdjust: "调整",
	DemandChangelogTypeDelete: "删除",
	DemandChangelogTypeExpend: "消耗",
}

// Name get demand changelog type name.
func (d DemandChangelogType) Name() string {
	return demandChangelogTypeNameMap[d]
}

// ResPlanWeekHolidayStatus is resource plan week holiday status.
type ResPlanWeekHolidayStatus int8

const (
	// ResPlanWeekIsHoliday resource plan week is holiday.
	ResPlanWeekIsHoliday ResPlanWeekHolidayStatus = 0
	// ResPlanWeekIsNotHoliday resource plan week is not holiday.
	ResPlanWeekIsNotHoliday ResPlanWeekHolidayStatus = 1
)

// Validate ResPlanWeekHolidayStatus.
func (s ResPlanWeekHolidayStatus) Validate() error {
	switch s {
	case ResPlanWeekIsHoliday:
	case ResPlanWeekIsNotHoliday:
	default:
		return fmt.Errorf("unsupported res plan week holiday status: %d", s)
	}

	return nil
}

// ResPlanNotMatchReason is resource plan not match reason.
type ResPlanNotMatchReason string

const (
	// DemandClassIsNotMatch 需求类型不匹配时的提示语
	DemandClassIsNotMatch ResPlanNotMatchReason = "需求类型不匹配"
	// DiskTypeIsNotMatch 磁盘类型不匹配时的提示语
	DiskTypeIsNotMatch ResPlanNotMatchReason = "磁盘类型不匹配"
)

// GenerateMsg generate res plan not match reason msg.
func (r ResPlanNotMatchReason) GenerateMsg(applyType string, planType string) string {
	return fmt.Sprintf("%s：申请单为(%s)，预测单为(%s)", r, applyType, planType)
}

// CRPDiskType is crp disk type.
type CRPDiskType int

const (
	// CRPDiskTypePREMIUM CLOUD_PREMIUM
	CRPDiskTypePREMIUM CRPDiskType = 606
	// CRPDiskTypeSSD CLOUD_SSD
	CRPDiskTypeSSD CRPDiskType = 607
)

// Validate CRPDiskType.
func (c CRPDiskType) Validate() error {
	switch c {
	case CRPDiskTypePREMIUM:
	case CRPDiskTypeSSD:
	default:
		return fmt.Errorf("unsupported crp disk type: %d", c)
	}
	return nil
}

// CRPDiskTypeNameMap crp disk type name map.
var CRPDiskTypeNameMap = map[CRPDiskType]string{
	CRPDiskTypePREMIUM: "高性能云硬盘",
	CRPDiskTypeSSD:     "SSD云硬盘",
}

// Name return disk type name.
func (c CRPDiskType) Name() string {
	return CRPDiskTypeNameMap[c]
}

// GetCRPDiskTypeFromCRPName get CRPDiskType from crp disk type name.
func GetCRPDiskTypeFromCRPName(name string) (CRPDiskType, error) {
	for k, v := range CRPDiskTypeNameMap {
		if v == name {
			return k, nil
		}
	}

	return 0, fmt.Errorf("unsupported crp disk type name: %s", name)
}

// ResPlanReviewStatus is res plan review status.
type ResPlanReviewStatus string

const (
	// ResPlanReviewStatusPass 已评审
	ResPlanReviewStatusPass ResPlanReviewStatus = "已评审"
	// ResPlanReviewStatusPending 待评审
	ResPlanReviewStatusPending ResPlanReviewStatus = "待评审"
)

// Validate ResPlanReviewStatus.
func (r ResPlanReviewStatus) Validate() error {
	switch r {
	case ResPlanReviewStatusPass:
	case ResPlanReviewStatusPending:
	default:
		return fmt.Errorf("unsupported res plan review status: %s", r)
	}
	return nil
}
