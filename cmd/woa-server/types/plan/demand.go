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

// Package plan ...
package plan

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	mtypes "hcm/pkg/dal/dao/types/meta"
	wdttablers "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// ListResPlanDemandReq is list resource plan demand request.
type ListResPlanDemandReq struct {
	BkBizIDs        []int64               `json:"bk_biz_ids" validate:"omitempty,max=100"`
	OpProductIDs    []int64               `json:"op_product_ids" validate:"omitempty,max=100"`
	PlanProductIDs  []int64               `json:"plan_product_ids" validate:"omitempty,max=100"`
	DemandIDs       []string              `json:"demand_ids" validate:"omitempty,max=100"`
	ObsProjects     []enumor.ObsProject   `json:"obs_projects" validate:"omitempty,max=100"`
	DemandClasses   []enumor.DemandClass  `json:"demand_classes" validate:"omitempty,max=100"`
	CoreTypes       []enumor.CoreType     `json:"core_types" validate:"omitempty,max=100"`
	DeviceFamilies  []string              `json:"device_families" validate:"omitempty,max=100"`
	DeviceClasses   []string              `json:"device_classes" validate:"omitempty,max=100"`
	DeviceTypes     []string              `json:"device_types" validate:"omitempty,max=100"`
	RegionIDs       []string              `json:"region_ids" validate:"omitempty,max=100"`
	ZoneIDs         []string              `json:"zone_ids" validate:"omitempty,max=100"`
	PlanTypes       []enumor.PlanType     `json:"plan_types" validate:"omitempty,max=100"`
	ExpiringOnly    bool                  `json:"expiring_only" validate:"omitempty"`
	ExpectTimeRange *times.DateRange      `json:"expect_time_range" validate:"required"`
	Statuses        []enumor.DemandStatus `json:"statuses" validate:"omitempty,max=5"`
	Page            *core.BasePage        `json:"page" validate:"required"`
}

// Validate whether ListResPlanDemandReq is valid.
func (r ListResPlanDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, bkBizID := range r.BkBizIDs {
		if bkBizID <= 0 {
			return errors.New("bk biz id should be > 0")
		}
	}
	for _, opProductID := range r.OpProductIDs {
		if opProductID <= 0 {
			return errors.New("op product id should be > 0")
		}
	}
	for _, planProductID := range r.PlanProductIDs {
		if planProductID <= 0 {
			return errors.New("plan product id should be > 0")
		}
	}

	for _, projectName := range r.ObsProjects {
		if err := projectName.ValidateResPlan(); err != nil {
			return err
		}
	}

	for _, class := range r.DemandClasses {
		if err := class.Validate(); err != nil {
			return err
		}
	}

	for _, planType := range r.PlanTypes {
		if err := planType.Validate(); err != nil {
			return err
		}
	}

	if r.ExpectTimeRange != nil {
		if err := r.ExpectTimeRange.Validate(); err != nil {
			return err
		}
	}

	for _, status := range r.Statuses {
		if err := status.Validate(); err != nil {
			return err
		}
	}

	if r.Page != nil {
		if err := r.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CheckDemandIDs check whether demand id contains.
func (r ListResPlanDemandReq) CheckDemandIDs(demandID string) bool {
	if len(r.DemandIDs) > 0 {
		return slices.Contains(r.DemandIDs, demandID)
	}
	return true
}

// CheckObsProjects check whether obs project contains.
func (r ListResPlanDemandReq) CheckObsProjects(obsProject enumor.ObsProject) bool {
	if len(r.ObsProjects) > 0 {
		return slices.Contains(r.ObsProjects, obsProject)
	}
	return true
}

// CheckDemandClasses check whether demand class contains.
func (r ListResPlanDemandReq) CheckDemandClasses(demandClass enumor.DemandClass) bool {
	if len(r.DemandClasses) > 0 {
		return slices.Contains(r.DemandClasses, demandClass)
	}
	return true
}

// CheckDeviceClasses check whether device class contains.
func (r ListResPlanDemandReq) CheckDeviceClasses(deviceClass string) bool {
	if len(r.DeviceClasses) > 0 {
		return slices.Contains(r.DeviceClasses, deviceClass)
	}
	return true
}

// CheckDeviceTypes check whether device type contains.
func (r ListResPlanDemandReq) CheckDeviceTypes(deviceType string) bool {
	if len(r.DeviceTypes) > 0 {
		return slices.Contains(r.DeviceTypes, deviceType)
	}
	return true
}

// CheckRegionIDs check whether region id contains.
func (r ListResPlanDemandReq) CheckRegionIDs(regionID string) bool {
	if len(r.RegionIDs) > 0 {
		return slices.Contains(r.RegionIDs, regionID)
	}
	return true
}

// CheckZoneIDs check whether zone id contains.
func (r ListResPlanDemandReq) CheckZoneIDs(zoneID string) bool {
	if len(r.ZoneIDs) > 0 {
		return slices.Contains(r.ZoneIDs, zoneID)
	}
	return true
}

// CheckPlanTypes check whether plan type contains.
func (r ListResPlanDemandReq) CheckPlanTypes(planType enumor.PlanType) bool {
	if len(r.PlanTypes) > 0 {
		return slices.Contains(r.PlanTypes, planType)
	}
	return true
}

// ListResPlanDemandResp is list resource plan demand response.
type ListResPlanDemandResp struct {
	Overview *ListResPlanDemandOverview `json:"overview" validate:"omitempty"`
	Count    uint64                     `json:"count"`
	Details  []*ListResPlanDemandItem   `json:"details"`
}

// ListDemandIds list crp demand ids
func (l *ListResPlanDemandResp) ListDemandIds() []string {
	res := make([]string, 0, len(l.Details))
	for _, item := range l.Details {
		res = append(res, item.DemandID)
	}
	return res
}

// ListResPlanDemandOverview is list resource plan demand overview
type ListResPlanDemandOverview struct {
	TotalCpuCore          int64 `json:"total_cpu_core"`
	TotalAppliedCore      int64 `json:"total_applied_core"`
	InPlanCpuCore         int64 `json:"in_plan_cpu_core"`
	InPlanAppliedCpuCore  int64 `json:"in_plan_applied_cpu_core"`
	OutPlanCpuCore        int64 `json:"out_plan_cpu_core"`
	OutPlanAppliedCpuCore int64 `json:"out_plan_applied_cpu_core"`
	ExpiringCpuCore       int64 `json:"expiring_cpu_core"`
}

// ListResPlanDemandItem is list resource plan demand detail's item
type ListResPlanDemandItem struct {
	DemandID         string               `json:"demand_id"`
	BkBizID          int64                `json:"bk_biz_id"`
	BkBizName        string               `json:"bk_biz_name"`
	OpProductID      int64                `json:"op_product_id"`
	OpProductName    string               `json:"op_product_name"`
	PlanProductID    int64                `json:"plan_product_id"`
	PlanProductName  string               `json:"plan_product_name"`
	Status           enumor.DemandStatus  `json:"status"`
	StatusName       string               `json:"status_name"`
	DemandClass      enumor.DemandClass   `json:"demand_class"`
	DemandResType    enumor.DemandResType `json:"demand_res_type"`
	ExpectTime       string               `json:"expect_time"`
	ReturnPlanTime   *string              `json:"return_plan_time"`
	CanApplyTime     string               `json:"can_apply_time"`
	ExpiredTime      string               `json:"expired_time"`
	DeviceClass      string               `json:"device_class"`
	DeviceType       string               `json:"device_type"`
	TotalOS          decimal.Decimal      `json:"total_os"`
	AppliedOS        decimal.Decimal      `json:"applied_os"`
	RemainedOS       decimal.Decimal      `json:"remained_os"`
	TotalCpuCore     int64                `json:"total_cpu_core"`
	AppliedCpuCore   int64                `json:"applied_cpu_core"`
	RemainedCpuCore  int64                `json:"remained_cpu_core"`
	ExpiringCpuCore  int64                `json:"-"` // ExpiringCpuCore 即将过期核心数，目前仅用于计算overview
	TotalMemory      int64                `json:"total_memory"`
	AppliedMemory    int64                `json:"applied_memory"`
	RemainedMemory   int64                `json:"remained_memory"`
	TotalDiskSize    int64                `json:"total_disk_size"`
	AppliedDiskSize  int64                `json:"applied_disk_size"`
	RemainedDiskSize int64                `json:"remained_disk_size"`
	RegionID         string               `json:"region_id"`
	RegionName       string               `json:"region_name"`
	AreaName         string               `json:"area_name"`
	ResMode          enumor.ResModeCode   `json:"res_mode"`
	ZoneID           string               `json:"zone_id"`
	ZoneName         string               `json:"zone_name"`
	PlanType         enumor.PlanType      `json:"plan_type"`
	ObsProject       enumor.ObsProject    `json:"obs_project"`
	TechnicalClass   string               `json:"technical_class"`
	TicketID         string               `json:"ticket_id"`
	DeviceFamily     string               `json:"device_family"`
	CoreType         enumor.CoreType      `json:"core_type"`
	DiskType         enumor.DiskType      `json:"disk_type"`
	DiskTypeName     string               `json:"disk_type_name"`
	DiskIO           int64                `json:"disk_io"`
	Creator          string               `json:"creator"`
	Reviser          string               `json:"reviser"`
}

// SetStatus set demand status
func (l *ListResPlanDemandItem) SetStatus(status enumor.DemandStatus) {
	l.Status = status
	// spent_all（已耗尽）优先级更高
	if l.AppliedCpuCore == l.TotalCpuCore {
		l.Status = enumor.DemandStatusSpentAll
	}
}

// SetRegionAndZoneID set region and zone id
func (l *ListResPlanDemandItem) SetRegionAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[l.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", l.RegionName)
	}
	l.RegionID = regionArea.RegionID

	zoneID, exists := zoneNameMap[l.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", l.ZoneName)
	}
	l.ZoneID = zoneID
	return nil
}

// PlanDemandDetail crp demand detail的本地格式化
type PlanDemandDetail struct {
	GetPlanDemandDetailResp `json:",inline"`
	Year                    int     `json:"year"`
	Month                   int     `json:"month"`
	Week                    int     `json:"week"`
	TotalOS                 float32 `json:"total_os"`
	AppliedOS               float32 `json:"applied_os"`
	RemainedOS              float32 `json:"remained_os"`
	TotalCpuCore            float32 `json:"total_cpu_core"`
	AppliedCpuCore          float32 `json:"applied_cpu_core"`
	RemainedCpuCore         float32 `json:"remained_cpu_core"`
	ExpiringCpuCore         float32 `json:"expiring_cpu_core"`
	TotalMemory             float32 `json:"total_memory"`
	AppliedMemory           float32 `json:"applied_memory"`
	RemainedMemory          float32 `json:"remained_memory"`
	TotalDiskSize           float32 `json:"total_disk_size"`
	AppliedDiskSize         float32 `json:"applied_disk_size"`
	RemainedDiskSize        float32 `json:"remained_disk_size"`
}

// GetPlanDemandDetailResp get plan demand detail response
type GetPlanDemandDetailResp struct {
	DemandID        string            `json:"demand_id"`
	ExpectTime      string            `json:"expect_time"`
	ReturnPlanTime  *string           `json:"return_plan_time"`
	BkBizID         int64             `json:"bk_biz_id"`
	BkBizName       string            `json:"bk_biz_name"`
	DeptID          int64             `json:"dept_id"`
	DeptName        string            `json:"dept_name"`
	PlanProductID   int64             `json:"plan_product_id"`
	PlanProductName string            `json:"plan_product_name"`
	OpProductID     int64             `json:"op_product_id"`
	OpProductName   string            `json:"op_product_name"`
	ObsProject      enumor.ObsProject `json:"obs_project"`
	AreaName        string            `json:"area_name"`
	RegionID        string            `json:"region_id"`
	RegionName      string            `json:"region_name"`
	ZoneID          string            `json:"zone_id"`
	ZoneName        string            `json:"zone_name"`
	PlanType        enumor.PlanType   `json:"plan_type"`
	CoreType        enumor.CoreType   `json:"core_type"`
	DeviceFamily    string            `json:"device_family"`
	DeviceClass     string            `json:"device_class"`
	DeviceType      string            `json:"device_type"`
	OS              decimal.Decimal   `json:"os"`
	Memory          int64             `json:"memory"`
	CpuCore         int64             `json:"cpu_core"`
	DiskSize        int64             `json:"disk_size"`
	DiskIO          int64             `json:"disk_io"`
	DiskType        enumor.DiskType   `json:"disk_type"`
	DiskTypeName    string            `json:"disk_type_name"`
	ResMode         enumor.ResMode    `json:"res_mode"`
}

// SetDiskType set disk type
func (g *GetPlanDemandDetailResp) SetDiskType() error {
	diskTypes := enumor.GetDiskTypeMembers()
	for _, diskType := range diskTypes {
		if g.DiskTypeName == diskType.Name() {
			g.DiskType = diskType
			return nil
		}
	}
	return fmt.Errorf("invalid disk type name: %s", g.DiskTypeName)
}

// SetRegionAreaAndZoneID set region/area and zone id
func (g *GetPlanDemandDetailResp) SetRegionAreaAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[g.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", g.RegionName)
	}
	g.RegionID = regionArea.RegionID
	g.AreaName = regionArea.AreaName

	zoneID, exists := zoneNameMap[g.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", g.ZoneName)
	}
	g.ZoneID = zoneID
	return nil
}

// ListDemandChangeLogReq is list demand change log request.
type ListDemandChangeLogReq struct {
	DemandID string         `json:"demand_id" validate:"required"`
	Page     *core.BasePage `json:"page" validate:"required"`
}

// Validate whether ListDemandChangeLogReq is valid.
func (r *ListDemandChangeLogReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// ListDemandChangeLogResp is list demand change log response
type ListDemandChangeLogResp struct {
	Count   uint64                     `json:"count"`
	Details []*ListDemandChangeLogItem `json:"details"`
}

// Page 对details结果分页
func (l *ListDemandChangeLogResp) Page(page *core.BasePage) {
	// start超出范围
	if int(page.Start) >= len(l.Details) {
		l.Details = l.Details[:0]
		return
	}

	end := int(page.Start) + int(page.Limit)
	// end超出范围
	if end > len(l.Details) {
		end = len(l.Details)
	}
	// 按page需求截断slice
	l.Details = l.Details[int(page.Start):end]
	return
}

// ListDemandChangeLogItem is list demand change log detail's item
type ListDemandChangeLogItem struct {
	ID                string            `json:"id"`
	DemandId          string            `json:"demand_id"`
	ExpectTime        string            `json:"expect_time"`
	ObsProject        enumor.ObsProject `json:"obs_project"`
	RegionName        string            `json:"region_name"`
	ZoneName          string            `json:"zone_name"`
	DeviceType        string            `json:"device_type"`
	ChangeCvmAmount   decimal.Decimal   `json:"change_cvm_amount"`
	ChangeCoreAmount  int64             `json:"change_core_amount"`
	ChangeRamAmount   int64             `json:"change_ram_amount"`
	ChangedDiskAmount int64             `json:"changed_disk_amount"`
	DemandSource      string            `json:"demand_source"`
	TicketID          string            `json:"ticket_id"`
	CrpSn             string            `json:"crp_sn"`
	SuborderID        string            `json:"suborder_id"`
	CreateTime        string            `json:"create_time"`
	Remark            string            `json:"remark"`
}

// AdjustAbleDemandsReq is the request query demands that can be adjusted.
type AdjustAbleDemandsReq struct {
	RegionName      string            `json:"region_name" validate:"omitempty"`
	DeviceFamily    string            `json:"device_family" validate:"omitempty"`
	DeviceType      string            `json:"device_type" validate:"omitempty"`
	ExpectTime      string            `json:"expect_time" validate:"omitempty"`
	PlanProductName string            `json:"plan_product_name" validate:"omitempty"`
	OpProductName   string            `json:"op_product_name" validate:"omitempty"`
	ObsProject      enumor.ObsProject `json:"obs_project" validate:"omitempty"`
	DiskType        enumor.DiskType   `json:"disk_type" validate:"omitempty"`
	ResMode         enumor.ResMode    `json:"res_mode" validate:"omitempty"`
}

// Validate whether AdjustAbleDemandsReq is valid.
func (r AdjustAbleDemandsReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ExpectTime) > 0 {
		if _, err := time.Parse(constant.DateLayout, r.ExpectTime); err != nil {
			return err
		}
	}

	if len(r.ObsProject) > 0 {
		if err := r.ObsProject.ValidateResPlan(); err != nil {
			return err
		}
	}

	if len(r.DiskType) > 0 {
		if err := r.DiskType.Validate(); err != nil {
			return err
		}
	}

	if len(r.ResMode) > 0 {
		if err := r.ResMode.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AdjustRPDemandReq is adjust resource plan demand request.
type AdjustRPDemandReq struct {
	Adjusts []AdjustRPDemandReqElem `json:"adjusts" validate:"required,max=100"`
}

// Validate whether AdjustRPDemandReq is valid.
func (r *AdjustRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, adjust := range r.Adjusts {
		if err := adjust.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AdjustRPDemandReqElem is adjust resource plan demand request element.
type AdjustRPDemandReqElem struct {
	DemandID     string                    `json:"demand_id" validate:"required"`
	CrpDemandID  int64                     `json:"crp_demand_id" validate:"omitempty"`
	AdjustType   enumor.RPDemandAdjustType `json:"adjust_type" validate:"required"`
	DemandSource enumor.DemandSource       `json:"demand_source" validate:"omitempty"`
	OriginalInfo *CreateResPlanDemandReq   `json:"original_info" validate:"omitempty"`
	UpdatedInfo  *CreateResPlanDemandReq   `json:"updated_info" validate:"omitempty"`
	ExpectTime   string                    `json:"expect_time" validate:"omitempty"`
	// DelayOs 用于部分延期，此时仅指定部分的 OS 会被调整
	DelayOs *string `json:"delay_os" validate:"omitempty"`
}

// Validate whether AdjustRPDemandReqElem is valid.
func (e *AdjustRPDemandReqElem) Validate() error {
	if err := validator.Validate.Struct(e); err != nil {
		return err
	}

	if len(e.DemandID) <= 0 {
		return errors.New("invalid demand id, should be > 0")
	}

	// 常规修改和加急延期的通用参数校验
	if e.OriginalInfo == nil {
		return errors.New("original info of update demand can not be empty")
	}

	if err := e.OriginalInfo.Validate(); err != nil {
		return err
	}

	if e.UpdatedInfo == nil {
		return errors.New("updated info of update demand can not be empty")
	}

	if err := e.UpdatedInfo.Validate(); err != nil {
		return err
	}

	switch e.AdjustType {
	case enumor.RPDemandAdjustTypeUpdate:
		if e.DemandSource != "" {
			if err := e.DemandSource.Validate(); err != nil {
				return err
			}
		}
	case enumor.RPDemandAdjustTypeDelay:
		if len(e.ExpectTime) == 0 {
			return errors.New("expect time of delay demand can not be empty")
		}

		// 全部延期时不需要指定DelayOs，因此不对DelayOs做校验
	default:
		return fmt.Errorf("unsupported resource plan demand adjust type: %s", e.AdjustType)
	}

	return nil
}

// CancelRPDemandReq is cancel resource plan demand request.
type CancelRPDemandReq struct {
	CancelDemands []CancelRPDemandReqElem `json:"cancel_demands" validate:"required,max=100"`
}

// Validate whether CancelRPDemandReq is valid.
func (r *CancelRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, demand := range r.CancelDemands {
		if err := demand.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CancelRPDemandReqElem is cancel resource plan demand request element.
type CancelRPDemandReqElem struct {
	DemandID        string `json:"demand_id" validate:"required"`
	RemainedCpuCore int64  `json:"remained_cpu_core" validate:"required"`
}

// Validate whether CancelRPDemandReqElem is valid.
func (r CancelRPDemandReqElem) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.RemainedCpuCore <= 0 {
		return errors.New("remained cpu core should be > 0")
	}

	return nil
}

// RepairRPDemandReq is repair resource plan demand request.
type RepairRPDemandReq struct {
	BkBizIDs          []int64         `json:"bk_biz_ids" validate:"omitempty,max=100,dive,gt=0"`
	RepairTicketRange times.DateRange `json:"repair_ticket_range" validate:"required"`
}

// Validate whether RepairRPDemandReq is valid.
func (c RepairRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	if err := c.RepairTicketRange.Validate(); err != nil {
		return err
	}

	return nil
}

// SyncCRPDemandReq is sync crp demand request.
type SyncCRPDemandReq struct {
	CrpSN            string           `json:"crp_sn" validate:"required"`
	PriorBizIDs      []int64          `json:"prior_biz_ids" validate:"omitempty"`
	OpProductToBizID map[string]int64 `json:"op_product_to_biz_id" validate:"omitempty"`
}

// Validate whether SyncCRPDemandReq is valid.
func (s SyncCRPDemandReq) Validate() error {
	return validator.Validate.Struct(s)
}

// CalcPenaltyBaseReq is request of calc penalty base.
type CalcPenaltyBaseReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"omitempty,max=100"`
	// PenaltyBaseDay is any day of the penalty base week. Format is YYYY-MM-DD.
	PenaltyBaseDay string `json:"penalty_base_day" validate:"required"`
}

// Validate whether CalcPenaltyBaseReq is valid.
func (c CalcPenaltyBaseReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, c.PenaltyBaseDay)
	if err != nil {
		return err
	}

	return nil
}

// CalcAndPushPenaltyRatioReq is request of calc and push penalty ratio.
type CalcAndPushPenaltyRatioReq struct {
	// PenaltyTime is any day of the month which penalty ratio be calculated.
	// Note that the first week may not be part of the month.
	// Format is YYYY-MM-DD.
	PenaltyTime string `json:"penalty_time" validate:"required"`
}

// Validate whether CalcAndPushPenaltyRatioReq is valid.
func (c *CalcAndPushPenaltyRatioReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, c.PenaltyTime)
	if err != nil {
		return err
	}

	return nil
}

// PushExpireNoticeReq is request of push expire notice.
type PushExpireNoticeReq struct {
	BkBizIDs  []int64  `json:"bk_biz_ids" validate:"omitempty,max=100"`
	Receivers []string `json:"receivers" validate:"omitempty,max=10"`
}

// Validate whether PushExpireNoticeReq is valid.
func (p PushExpireNoticeReq) Validate() error {
	return validator.Validate.Struct(p)
}

// DemandResource is demand resource.
// Include device type, cpu core and disk size now.
// The OS and memory can be calculated by cpu core and device type.
type DemandResource struct {
	DeviceType string
	CpuCore    int64
	DiskSize   int64
}

// AutoTransferBizResPlanDemandReq is auto transfer biz res plan demand request.
type AutoTransferBizResPlanDemandReq struct {
	DemandIDs []string `json:"demand_ids" validate:"required,min=1,max=100"`
}

// Validate whether AutoTransferBizResPlanDemandReq is valid.
func (r AutoTransferBizResPlanDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}
	if len(r.DemandIDs) == 0 {
		return errors.New("demand_ids is required")
	}
	return nil
}

// AutoTransferBizResPlanDemandResp is auto transfer biz res plan demand response.
type AutoTransferBizResPlanDemandResp struct {
	TicketIDs []string `json:"ticket_ids"`
}

// CrpOrderChangeInfo is response of crp order change info.
type CrpOrderChangeInfo struct {
	OrderID         string               `json:"order_id"`
	VirtualDeptName string               `json:"virtual_dept_name"`
	PlanProductName string               `json:"plan_product_name"`
	OpProductName   string               `json:"op_product_name"`
	ExpectTime      string               `json:"expect_time"`
	ReturnPlanTime  string               `json:"return_plan_time"`
	ObsProject      enumor.ObsProject    `json:"obs_project"`
	DemandResType   enumor.DemandResType `json:"demand_res_type"`
	ResMode         enumor.ResModeCode   `json:"res_mode"`
	PlanType        enumor.PlanTypeCode  `json:"plan_type"`
	AreaName        string               `json:"area_name"`
	RegionID        string               `json:"region_id"`
	RegionName      string               `json:"region_name"`
	ZoneID          string               `json:"zone_id"`
	ZoneName        string               `json:"zone_name"`
	TechnicalClass  string               `json:"technical_class"`
	DeviceFamily    string               `json:"device_family"`
	DeviceClass     string               `json:"device_class"`
	DeviceType      string               `json:"device_type"`
	CoreType        string               `json:"core_type"`
	DiskType        enumor.DiskType      `json:"disk_type"`
	DiskTypeName    string               `json:"disk_type_name"`
	DiskIO          int64                `json:"disk_io"`

	ChangeOs       decimal.Decimal `json:"change_os"`
	ChangeCpuCore  int64           `json:"change_cpu_core"`
	ChangeMemory   int64           `json:"change_memory"`
	ChangeDiskSize int64           `json:"change_disk_size"`
}

// SetRegionAreaAndZoneID set region area and zone id.
func (c *CrpOrderChangeInfo) SetRegionAreaAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[c.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", c.RegionName)
	}
	c.RegionID = regionArea.RegionID
	c.AreaName = regionArea.AreaName

	// CRP的底层逻辑里空字符串会存储为"-"
	// 实际操作中，数据为空时CRP理论上会返回一个默认的可用区，这里仅为兜底
	if c.ZoneName == "-" {
		c.ZoneID = ""
		return nil
	}

	zoneID, exists := zoneNameMap[c.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", c.ZoneName)
	}
	c.ZoneID = zoneID
	return nil
}

// GetKey get key of crp order change info.
// 只能用于追加预测场景，修改场景使用默认diskType可能产生预期外行为
func (c *CrpOrderChangeInfo) GetKey(bkBizID int64, demandClass enumor.DemandClass) ResPlanDemandKey {
	key := ResPlanDemandKey{
		BkBizID:       bkBizID,
		DemandClass:   demandClass,
		DemandResType: c.DemandResType,
		ResMode:       c.ResMode,
		ObsProject:    c.ObsProject,
		ExpectTime:    c.ExpectTime,
		PlanType:      c.PlanType,
		RegionID:      c.RegionID,
		ZoneID:        c.ZoneID,
		DeviceType:    c.DeviceType,
		DiskIO:        c.DiskIO,
	}
	key.DiskType = key.DiskType.GetWithDefault()

	return key
}

// GetAggregateKey get aggregate key of crp order change info.
func (c *CrpOrderChangeInfo) GetAggregateKey(bkBizID int64, deviceTypes map[string]wdttablers.WoaDeviceTypeTable,
	expectTimeRange times.DateRange) (ResPlanDemandAggregateKey, error) {

	deviceInfo, ok := deviceTypes[c.DeviceType]
	if !ok {
		return ResPlanDemandAggregateKey{}, fmt.Errorf("device type: %s not found", c.DeviceType)
	}

	key := ResPlanDemandAggregateKey{
		BkBizID:         bkBizID,
		RegionID:        c.RegionID,
		ExpectTimeRange: expectTimeRange,
		DeviceFamily:    deviceInfo.DeviceFamily,
		CoreType:        deviceInfo.CoreType,
		PlanType:        c.PlanType,
		ObsProject:      c.ObsProject,
		ResType:         c.DemandResType,
		DiskType:        c.DiskType,
	}

	return key, nil
}

// ResPlanDemandKey is key of res plan demand. Used to uniquely identify a resource plan demand.
type ResPlanDemandKey struct {
	BkBizID       int64
	DemandClass   enumor.DemandClass
	DemandResType enumor.DemandResType
	ResMode       enumor.ResModeCode
	ObsProject    enumor.ObsProject
	ExpectTime    string
	PlanType      enumor.PlanTypeCode
	RegionID      string
	ZoneID        string
	DeviceType    string
	DiskType      enumor.DiskType
	DiskIO        int64
}

// ResPlanDemandAggregateKey 聚合key
// 为解决CRP模糊调整导致数据出现负数的问题，demandKey需要按照模糊范围查找多条进行调整，避免负数出现
// 模糊范围：城市、可用范围（当前是整个月，未来可能精确到周）、机型族、核心类型、预测内外、项目类型、资源类型、云盘类型
// Note: 云盘类型比较特殊，当云盘类型为 enumor.DiskUnknown 时，模糊范围排除掉云盘类型
type ResPlanDemandAggregateKey struct {
	BkBizID         int64
	RegionID        string
	ExpectTimeRange times.DateRange
	DeviceFamily    string
	CoreType        string
	PlanType        enumor.PlanTypeCode
	ObsProject      enumor.ObsProject
	ResType         enumor.DemandResType
	DiskType        enumor.DiskType
}

// DemandPenaltyBaseKey is key of demand penalty.
// bk_biz_id / area_name / device_family
type DemandPenaltyBaseKey struct {
	BkBizID      int64
	AreaName     string
	DeviceFamily string
}

// ResPlanDemandExpendKey is key of res plan demand expend.
type ResPlanDemandExpendKey struct {
	DemandClass enumor.DemandClass
	// DiskType      enumor.DiskType
	BkBizID       int64
	PlanType      enumor.PlanTypeCode
	AvailableTime AvailableMonth
	// 为了快速处理通配的情况，这里通过机型族和大小核心作为key
	// DeviceType    string
	DeviceFamily string
	CoreType     string
	ObsProject   enumor.ObsProject
	RegionID     string
}

// AvailableMonth available time.
type AvailableMonth string

// NewAvailableMonth new an available time.
// TODO: 目前只关注年和月，未来会添加周
func NewAvailableMonth(year int, month time.Month) AvailableMonth {
	return AvailableMonth(fmt.Sprintf("%04d-%02d", year, month))
}

// PushResPlanConfirmNoticeReq 手动触发资源计划确认通知请求
type PushResPlanConfirmNoticeReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"omitempty"`
}

// Validate whether PushResPlanConfirmNoticeReq is valid.
func (r *PushResPlanConfirmNoticeReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, bizID := range r.BkBizIDs {
		if bizID <= 0 {
			return errors.New("bk_biz_id must be greater than 0")
		}
	}
	return nil
}

// PushResPlanConfirmNoticeResp 手动触发资源计划确认通知响应
type PushResPlanConfirmNoticeResp struct {
	SuccessIDs []int64 `json:"success_ids"`
	FailedIDs  []int64 `json:"failed_ids"`
}

// ConfirmResPlanDemandsReq 确认资源计划需求请求
type ConfirmResPlanDemandsReq struct {
	BkBizID   int64    `json:"bk_biz_id" validate:"required"`
	DemandIDs []string `json:"demand_ids" validate:"required,min=1,max=100"`
}

// Validate whether ConfirmResPlanDemandsReq is valid.
func (r *ConfirmResPlanDemandsReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ConfirmBizResPlanDemandsReq 确认业务资源计划需求请求
type ConfirmBizResPlanDemandsReq struct {
	DemandIDs []string `json:"demand_ids" validate:"required,min=1,max=100"`
}

// Validate whether ConfirmBizResPlanDemandsReq is valid.
func (r *ConfirmBizResPlanDemandsReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ConfirmResPlanDemandsResp 确认资源计划需求响应
type ConfirmResPlanDemandsResp struct {
	SuccessIDs []string `json:"success_ids"`
	FailedIDs  []string `json:"failed_ids"`
}
