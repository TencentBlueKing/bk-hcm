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

package resplandemand

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	cvt "hcm/pkg/tools/converter"
)

// ResPlanDemandColumns defines all the resource plan demand table's columns.
var ResPlanDemandColumns = utils.MergeColumns(nil, ResPlanDemandColumnDescriptor)

// ResPlanDemandColumnDescriptor is ResPlanDemandTable's column descriptors.
var ResPlanDemandColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "locked", NamedC: "locked", Type: enumor.Numeric},
	{Column: "locked_cpu_core", NamedC: "locked_cpu_core", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "demand_class", NamedC: "demand_class", Type: enumor.String},
	{Column: "demand_res_type", NamedC: "demand_res_type", Type: enumor.String},
	{Column: "res_mode", NamedC: "res_mode", Type: enumor.String},
	{Column: "obs_project", NamedC: "obs_project", Type: enumor.String},
	{Column: "expect_time", NamedC: "expect_time", Type: enumor.Numeric},
	{Column: "return_plan_time", NamedC: "return_plan_time", Type: enumor.Numeric},
	{Column: "plan_type", NamedC: "plan_type", Type: enumor.String},
	{Column: "area_id", NamedC: "area_id", Type: enumor.String},
	{Column: "area_name", NamedC: "area_name", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "zone_id", NamedC: "zone_id", Type: enumor.String},
	{Column: "zone_name", NamedC: "zone_name", Type: enumor.String},
	{Column: "technical_class", NamedC: "technical_class", Type: enumor.String},
	{Column: "ticket_id", NamedC: "ticket_id", Type: enumor.String},
	{Column: "device_family", NamedC: "device_family", Type: enumor.String},
	{Column: "device_class", NamedC: "device_class", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "core_type", NamedC: "core_type", Type: enumor.String},
	{Column: "disk_type", NamedC: "disk_type", Type: enumor.String},
	{Column: "disk_type_name", NamedC: "disk_type_name", Type: enumor.String},
	{Column: "os", NamedC: "os", Type: enumor.Numeric},
	{Column: "cpu_core", NamedC: "cpu_core", Type: enumor.Numeric},
	{Column: "memory", NamedC: "memory", Type: enumor.Numeric},
	{Column: "disk_size", NamedC: "disk_size", Type: enumor.Numeric},
	{Column: "disk_io", NamedC: "disk_io", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanDemandTable is used to save resource's resource plan demand information.
type ResPlanDemandTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Locked 是否锁定
	Locked *enumor.CrpDemandLockStatus `db:"locked" json:"locked"`
	// LockedCPUCore 锁定的CPU核数
	LockedCPUCore *int64 `db:"locked_cpu_core" json:"locked_cpu_core"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BkBizName 业务名称
	BkBizName string `db:"bk_biz_name" json:"bk_biz_name" validate:"lte=64"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// OpProductName 运营产品名称
	OpProductName string `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	// PlanProductID 规划产品ID
	PlanProductID int64 `db:"plan_product_id" json:"plan_product_id"`
	// PlanProductName 规划产品名称
	PlanProductName string `db:"plan_product_name" json:"plan_product_name" validate:"lte=64"`
	// VirtualDeptID 虚拟部门ID
	VirtualDeptID int64 `db:"virtual_dept_id" json:"virtual_dept_id"`
	// VirtualDeptName 虚拟部门名称
	VirtualDeptName string `db:"virtual_dept_name" json:"virtual_dept_name" validate:"lte=64"`
	// DemandClass 需求类型
	DemandClass enumor.DemandClass `db:"demand_class" json:"demand_class" validate:"lte=16"`
	// DemandResType 需求资源类型
	DemandResType enumor.DemandResType `db:"demand_res_type" json:"demand_res_type" validate:"lte=8"`
	// ResMode 资源模式
	ResMode enumor.ResModeCode `db:"res_mode" json:"res_mode" validate:"lte=16"`
	// ObsProject 项目类型
	ObsProject enumor.ObsProject `db:"obs_project" json:"obs_project" validate:"lte=64"`
	// ExpectTime 期望交付时间，YYYYMMDD
	ExpectTime int `db:"expect_time" json:"expect_time"`
	// ReturnPlanTime 期望退回时间，YYYYMMDD
	ReturnPlanTime int `db:"return_plan_time" json:"return_plan_time"`
	// PlanType 预测内外
	PlanType enumor.PlanTypeCode `db:"plan_type" json:"plan_type" validate:"lte=16"`
	// AreaID 地域ID
	AreaID string `db:"area_id" json:"area_id" validate:"lte=64"`
	// AreaName 地域名称
	AreaName string `db:"area_name" json:"area_name" validate:"lte=64"`
	// RegionID 地区/城市ID
	RegionID string `db:"region_id" json:"region_id" validate:"lte=64"`
	// RegionName 地区/城市名称
	RegionName string `db:"region_name" json:"region_name" validate:"lte=64"`
	// ZoneID 可用区ID
	ZoneID string `db:"zone_id" json:"zone_id" validate:"lte=64"`
	// ZoneName 可用区名称
	ZoneName string `db:"zone_name" json:"zone_name" validate:"lte=64"`
	// TechnicalClass 技术分类
	TechnicalClass string `db:"technical_class" json:"technical_class" validate:"lte=64"`
	// TicketID 关联的res_plan_ticket ID
	TicketID string `db:"ticket_id" json:"ticket_id" validate:"lte=64"`
	// DeviceFamily 机型族
	DeviceFamily string `db:"device_family" json:"device_family" validate:"lte=64"`
	// DeviceClass 机型类型
	DeviceClass string `db:"device_class" json:"device_class" validate:"lte=64"`
	// DeviceType 机型类型
	DeviceType string `db:"device_type" json:"device_type" validate:"lte=64"`
	// CoreType 核心类型
	CoreType enumor.CoreType `db:"core_type" json:"core_type" validate:"lte=64"`
	// DiskType 磁盘类型
	DiskType enumor.DiskType `db:"disk_type" json:"disk_type" validate:"lte=64"`
	// DiskTypeName 磁盘类型名称
	DiskTypeName string `db:"disk_type_name" json:"disk_type_name" validate:"lte=64"`
	// OS 预测实例数
	OS *types.Decimal `db:"os" json:"os"`
	// CpuCore 预测CPU核数
	CpuCore *int64 `db:"cpu_core" json:"cpu_core"`
	// Memory 预测内存数
	Memory *int64 `db:"memory" json:"memory"`
	// DiskSize 磁盘大小
	DiskSize *int64 `db:"disk_size" json:"disk_size"`
	// DiskIO 磁盘IO
	DiskIO int64 `db:"disk_io" json:"disk_io"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r ResPlanDemandTable) TableName() table.Name {
	return table.ResPlanDemandTable
}

// InsertValidate validate resource plan demand on insertion.
func (r ResPlanDemandTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if err := r.Locked.Validate(); err != nil {
		return err
	}

	if err := r.bizInsertValidate(); err != nil {
		return err
	}

	if err := r.DemandClass.Validate(); err != nil {
		return err
	}

	if err := r.DemandResType.Validate(); err != nil {
		return err
	}

	if err := r.ResMode.Validate(); err != nil {
		return err
	}

	if err := r.ObsProject.ValidateResPlan(); err != nil {
		return err
	}

	if r.ExpectTime <= 0 {
		return errors.New("expect time can not be empty")
	}

	if err := r.PlanType.Validate(); err != nil {
		return err
	}

	// NOTE: zone can be empty.

	if len(r.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(r.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(r.DeviceType) == 0 {
		return errors.New("device type can not be empty")
	}

	// NOTE: disk type can be empty.
	if len(r.DiskType) > 0 {
		if err := r.DiskType.Validate(); err != nil {
			return err
		}
	}

	if err := r.resourceInsertValidate(); err != nil {
		return err
	}

	if r.DiskIO <= 0 {
		return errors.New("disk io should be > 0")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

func (r ResPlanDemandTable) bizInsertValidate() error {
	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(r.BkBizName) == 0 {
		return errors.New("bk biz name can not be empty")
	}

	if r.OpProductID <= 0 {
		return errors.New("op product id should be > 0")
	}

	if len(r.OpProductName) == 0 {
		return errors.New("op product name can not be empty")
	}

	if r.PlanProductID <= 0 {
		return errors.New("plan product id should be > 0")
	}

	if len(r.PlanProductName) == 0 {
		return errors.New("plan product name can not be empty")
	}

	if r.VirtualDeptID <= 0 {
		return errors.New("virtual dept id should be > 0")
	}

	if len(r.VirtualDeptName) == 0 {
		return errors.New("virtual dept name can not be empty")
	}

	return nil
}

func (r ResPlanDemandTable) resourceInsertValidate() error {
	if r.LockedCPUCore == nil {
		return errors.New("locked cpu core can not be nil")
	}

	if cvt.PtrToVal(r.LockedCPUCore) < 0 {
		return errors.New("locked cpu core should be >= 0")
	}

	if r.OS.Sign() < 0 {
		return errors.New("os should be >= 0")
	}

	if r.CpuCore == nil {
		return errors.New("cpu core can not be nil")
	}

	if cvt.PtrToVal(r.CpuCore) < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if r.Memory == nil {
		return errors.New("memory can not be nil")
	}

	if cvt.PtrToVal(r.Memory) < 0 {
		return errors.New("memory should be >= 0")
	}

	if r.DiskSize == nil {
		return errors.New("disk size can not be nil")
	}

	if cvt.PtrToVal(r.DiskSize) < 0 {
		return errors.New("disk size should be >= 0")
	}

	return nil
}

// UpdateValidate validate resource plan demand on update.
func (r ResPlanDemandTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.Locked != nil {
		if err := r.Locked.Validate(); err != nil {
			return err
		}
	}

	if err := r.bizUpdateValidate(); err != nil {
		return err
	}

	if err := r.demandResourceMetaValidate(); err != nil {
		return err
	}

	if len(r.PlanType) > 0 {
		if err := r.PlanType.Validate(); err != nil {
			return err
		}
	}

	if len(r.DiskType) > 0 {
		if err := r.DiskType.Validate(); err != nil {
			return err
		}
	}

	if err := r.resourceUpdateValidate(); err != nil {
		return err
	}

	if r.DiskIO < 0 {
		return errors.New("disk io should be >= 0")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}

// 抽出部分检查，降低圈复杂度
func (r ResPlanDemandTable) demandResourceMetaValidate() error {
	if len(r.DemandClass) > 0 {
		if err := r.DemandClass.Validate(); err != nil {
			return err
		}
	}

	if len(r.DemandResType) > 0 {
		if err := r.DemandResType.Validate(); err != nil {
			return err
		}
	}

	if len(r.ResMode) > 0 {
		if err := r.ResMode.Validate(); err != nil {
			return err
		}
	}

	if len(r.ObsProject) > 0 {
		if err := r.ObsProject.ValidateResPlan(); err != nil {
			return err
		}
	}

	return nil
}

func (r ResPlanDemandTable) bizUpdateValidate() error {
	if r.BkBizID < 0 {
		return errors.New("bk biz id should be >= 0")
	}

	if r.OpProductID < 0 {
		return errors.New("op product id should be >= 0")
	}

	if r.PlanProductID < 0 {
		return errors.New("plan product id should be >= 0")
	}

	if r.VirtualDeptID < 0 {
		return errors.New("virtual dept id should be >= 0")
	}

	return nil
}

func (r ResPlanDemandTable) resourceUpdateValidate() error {
	if cvt.PtrToVal(r.LockedCPUCore) < 0 {
		return errors.New("locked cpu core should be >= 0")
	}

	if r.OS != nil && r.OS.Sign() < 0 {
		return errors.New("os should be >= 0")
	}

	if cvt.PtrToVal(r.CpuCore) < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if cvt.PtrToVal(r.Memory) < 0 {
		return errors.New("memory should be >= 0")
	}

	if cvt.PtrToVal(r.DiskSize) < 0 {
		return errors.New("disk size should be >= 0")
	}

	return nil
}
