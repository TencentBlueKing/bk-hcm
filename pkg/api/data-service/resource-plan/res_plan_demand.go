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

// Package resourceplan ...
package resourceplan

import (
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"

	"github.com/shopspring/decimal"
)

// ResPlanDemandBatchCreateReq create request
type ResPlanDemandBatchCreateReq struct {
	Demands []ResPlanDemandCreateReq `json:"demands" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanDemandBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Demands {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandCreateReq create request
type ResPlanDemandCreateReq struct {
	BkBizID         int64                `json:"bk_biz_id" validate:"required"`
	BkBizName       string               `json:"bk_biz_name" validate:"required"`
	OpProductID     int64                `json:"op_product_id" validate:"required"`
	OpProductName   string               `json:"op_product_name" validate:"required"`
	PlanProductID   int64                `json:"plan_product_id" validate:"required"`
	PlanProductName string               `json:"plan_product_name" validate:"required"`
	VirtualDeptID   int64                `json:"virtual_dept_id" validate:"required"`
	VirtualDeptName string               `json:"virtual_dept_name" validate:"required"`
	DemandClass     enumor.DemandClass   `json:"demand_class" validate:"required"`
	DemandResType   enumor.DemandResType `json:"demand_res_type" validate:"required"`
	ResMode         enumor.ResModeCode   `json:"res_mode" validate:"required"`
	ObsProject      enumor.ObsProject    `json:"obs_project" validate:"required"`
	ExpectTime      string               `json:"expect_time" validate:"required"`
	ReturnPlanTime  string               `json:"return_plan_time" validate:"omitempty"`
	PlanType        enumor.PlanTypeCode  `json:"plan_type" validate:"required"`
	AreaID          string               `json:"area_id" validate:"required"`
	AreaName        string               `json:"area_name" validate:"required"`
	RegionID        string               `json:"region_id" validate:"required"`
	RegionName      string               `json:"region_name" validate:"required"`
	ZoneID          string               `json:"zone_id" validate:"omitempty"`
	ZoneName        string               `json:"zone_name" validate:"omitempty"`
	TechnicalClass  string               `json:"technical_class" validate:"required"`
	DeviceFamily    string               `json:"device_family" validate:"required"`
	DeviceClass     string               `json:"device_class" validate:"required"`
	DeviceType      string               `json:"device_type" validate:"required"`
	CoreType        string               `json:"core_type" validate:"required"`
	DiskType        enumor.DiskType      `json:"disk_type" validate:"required"`
	DiskTypeName    string               `json:"disk_type_name" validate:"required"`
	OS              *decimal.Decimal     `json:"os" validate:"required"`
	CpuCore         *int64               `json:"cpu_core" validate:"required"`
	Memory          *int64               `json:"memory" validate:"required"`
	DiskSize        *int64               `json:"disk_size" validate:"required"`
	DiskIO          int64                `json:"disk_io" validate:"required"`
	Creator         string               `json:"creator" validate:"omitempty"`
}

// Validate validate
func (r *ResPlanDemandCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, r.ExpectTime)
	if err != nil {
		return err
	}

	if r.ObsProject == enumor.ObsProjectShortLease {
		_, err = time.Parse(constant.DateLayout, r.ReturnPlanTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// ResPlanDemandBatchUpdateReq batch update request
type ResPlanDemandBatchUpdateReq struct {
	Demands []ResPlanDemandUpdateReq `json:"demands" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanDemandBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Demands {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandUpdateReq batch update request
type ResPlanDemandUpdateReq struct {
	ID              string           `json:"id" validate:"required"`
	OpProductID     int64            `json:"op_product_id"`
	OpProductName   string           `json:"op_product_name"`
	PlanProductID   int64            `json:"plan_product_id"`
	PlanProductName string           `json:"plan_product_name"`
	VirtualDeptID   int64            `json:"virtual_dept_id"`
	VirtualDeptName string           `json:"virtual_dept_name"`
	CoreType        string           `json:"core_type"`
	OS              *decimal.Decimal `json:"os"`
	CpuCore         *int64           `json:"cpu_core"`
	Memory          *int64           `json:"memory"`
	DiskSize        *int64           `json:"disk_size"`
	Reviser         string           `json:"reviser"`
	TechnicalClass  string           `json:"technical_class"`
	IsConfirmed     bool             `json:"is_confirmed"`
}

// Validate validate
func (r *ResPlanDemandUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResPlanDemandLockOpReq lock operation request
type ResPlanDemandLockOpReq struct {
	LockedItems []ResPlanDemandLockOpItem `json:"locked_objs" validate:"required,min=1,max=100"`
}

// Validate validate
func (r ResPlanDemandLockOpReq) Validate(isLock bool) error {
	for _, item := range r.LockedItems {
		if err := item.Validate(isLock); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(r)
}

// NewResPlanDemandLockOpReqBatch return ResPlanDemandLockOpReq has same locked_cpu_core
func NewResPlanDemandLockOpReqBatch(demandIDs []string, lockedCPUCore int64) *ResPlanDemandLockOpReq {
	items := make([]ResPlanDemandLockOpItem, len(demandIDs))
	for i, id := range demandIDs {
		items[i] = ResPlanDemandLockOpItem{
			ID:            id,
			LockedCPUCore: lockedCPUCore,
		}
	}

	return &ResPlanDemandLockOpReq{
		LockedItems: items,
	}
}

// ResPlanDemandLockOpItem lock operation item
type ResPlanDemandLockOpItem struct {
	ID            string `json:"id" validate:"required"`
	TicketID      string `json:"ticket_id"`
	LockedCPUCore int64  `json:"locked_cpu_core"`
}

// Validate validate
func (r ResPlanDemandLockOpItem) Validate(isLock bool) error {
	if isLock {
		if r.TicketID == "" {
			return errf.New(errf.InvalidParameter, "ticket_id can not be empty")
		}
	}
	return validator.Validate.Struct(r)
}

// ResPlanDemandBatchUpsertReq batch upsert request
type ResPlanDemandBatchUpsertReq struct {
	CreateDemands []ResPlanDemandCreateReq `json:"create_demands" validate:"omitempty"`
	UpdateDemands []ResPlanDemandUpdateReq `json:"update_demands" validate:"omitempty"`
}

// Validate validate
func (r ResPlanDemandBatchUpsertReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.CreateDemands) == 0 && len(r.UpdateDemands) == 0 {
		return errf.New(errf.InvalidParameter, "create demands and update demands can not be both empty")
	}

	for _, c := range r.CreateDemands {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	for _, c := range r.UpdateDemands {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandListResult list result
type ResPlanDemandListResult types.ListResult[tablers.ResPlanDemandTable]

// ResPlanDemandListReq list request
type ResPlanDemandListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *ResPlanDemandListReq) Validate() error {
	return r.ListReq.Validate()
}
