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

package resplanticket

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanTicketColumns defines all the resource plan ticket table's columns.
var ResPlanTicketColumns = utils.MergeColumns(nil, ResPlanTicketColumnDescriptor)

// ResPlanTicketColumnDescriptor is ResPlanTicketTable's column descriptors.
var ResPlanTicketColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "demands", NamedC: "demands", Type: enumor.Json},
	{Column: "applicant", NamedC: "applicant", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "demand_class", NamedC: "demand_class", Type: enumor.String},
	{Column: "original_os", NamedC: "original_os", Type: enumor.Numeric},
	{Column: "original_cpu_core", NamedC: "original_cpu_core", Type: enumor.Numeric},
	{Column: "original_memory", NamedC: "original_memory", Type: enumor.Numeric},
	{Column: "original_disk_size", NamedC: "original_disk_size", Type: enumor.Numeric},
	{Column: "updated_os", NamedC: "updated_os", Type: enumor.Numeric},
	{Column: "updated_cpu_core", NamedC: "updated_cpu_core", Type: enumor.Numeric},
	{Column: "updated_memory", NamedC: "updated_memory", Type: enumor.Numeric},
	{Column: "updated_disk_size", NamedC: "updated_disk_size", Type: enumor.Numeric},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "submitted_at", NamedC: "submitted_at", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanTicketTable is used to save resource's resource plan ticket information.
type ResPlanTicketTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Type 单据类型
	Type enumor.RPTicketType `db:"type" json:"ticket_type" validate:"lte=64"`
	// Demands 需求列表，每个需求包括：original、updated两个部分
	Demands types.JsonField `db:"demands" json:"demands"`
	// Applicant 申请人
	Applicant string `db:"applicant" json:"applicant" validate:"lte=64"`
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
	// DemandClass 预测的需求类型
	DemandClass enumor.DemandClass `db:"demand_class" json:"demand_class" validate:"lte=16"`
	// TODO OS需要调整为decimal；cpu、memory、disk_size需要调整为int64
	// OriginalOS 原始OS数，单位：台
	OriginalOS float64 `db:"original_os" json:"original_os"`
	// OriginalCpuCore 原始CPU核心数，单位：台
	OriginalCpuCore int64 `db:"original_cpu_core" json:"original_cpu_core"`
	// OriginalMemory 原始内存大小，单位：GB
	OriginalMemory int64 `db:"original_memory" json:"original_memory"`
	// OriginalDiskSize 原始云盘大小，单位：GB
	OriginalDiskSize int64 `db:"original_disk_size" json:"original_disk_size"`
	// UpdatedOS 更新OS数，单位：台
	UpdatedOS float64 `db:"updated_os" json:"updated_os"`
	// UpdatedCpuCore 更新CPU核心数，单位：台
	UpdatedCpuCore int64 `db:"updated_cpu_core" json:"updated_cpu_core"`
	// UpdatedMemory 更新内存大小，单位：GB
	UpdatedMemory int64 `db:"updated_memory" json:"updated_memory"`
	// UpdatedDiskSize 更新云盘大小，单位：GB
	UpdatedDiskSize int64 `db:"updated_disk_size" json:"updated_disk_size"`
	// Remark 预测说明，最短20，最长1024
	Remark string `db:"remark" json:"remark" validate:"lte=1024"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// SubmittedAt 提单或改单的时间
	SubmittedAt string `db:"submitted_at" json:"submitted_at"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// ResPlanDemands is Demands struct of ResPlanTicketTable.
type ResPlanDemands []ResPlanDemand

// ResPlanDemand is used to save resource's resource plan demand information.
type ResPlanDemand struct {
	// DemandClass 预测的需求类型
	DemandClass enumor.DemandClass `json:"demand_class" validate:"lte=16"`
	// Original 原始需求
	Original *OriginalRPDemandItem `json:"original"`
	// Updated 更新需求
	Updated *UpdatedRPDemandItem `json:"updated"`
}

// Validate whether ResPlanDemand is valid.
func (d ResPlanDemand) Validate() error {
	if err := validator.Validate.Struct(d); err != nil {
		return err
	}

	if err := d.DemandClass.Validate(); err != nil {
		return err
	}

	if d.Original != nil {
		if err := d.Original.Validate(); err != nil {
			return err
		}
	}

	if d.Updated != nil {
		if err := d.Updated.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a string representation of ResPlanDemand.
func (d ResPlanDemand) String() string {
	return fmt.Sprintf("origin: %+v, update: %+v", d.Original, d.Updated)
}

// Clone return a clone ResPlanDemand.
func (d *ResPlanDemand) Clone() *ResPlanDemand {
	newD := &ResPlanDemand{
		DemandClass: d.DemandClass,
		Original:    nil,
		Updated:     nil,
	}

	if d.Original != nil {
		newD.Original = d.Original.Clone()
	}
	if d.Updated != nil {
		newD.Updated = d.Updated.Clone()
	}

	return newD
}

// CloneOriginal return a clone ResPlanDemand with only original demand.
func (d *ResPlanDemand) CloneOriginal() *ResPlanDemand {
	newD := &ResPlanDemand{
		DemandClass: d.DemandClass,
		Original:    nil,
		Updated:     nil,
	}

	if d.Original != nil {
		newD.Original = d.Original.Clone()
	}

	return newD
}

// CloneUpdated return a clone ResPlanDemand with only updated demand.
func (d *ResPlanDemand) CloneUpdated() *ResPlanDemand {
	newD := &ResPlanDemand{
		DemandClass: d.DemandClass,
		Original:    nil,
		Updated:     nil,
	}

	if d.Updated != nil {
		newD.Updated = d.Updated.Clone()
	}

	return newD
}

// ExtractCbsDemand 提取 CBS 需求（克隆需求，保留 Cbs，清空 Cvm）
func (d *ResPlanDemand) ExtractCbsDemand() *ResPlanDemand {
	cbsDemand := d.Clone()
	if cbsDemand.Original != nil {
		// 清空 Cvm 信息，只保留 CBS
		cbsDemand.Original.Cvm = Cvm{}
	}

	if cbsDemand.Updated != nil {
		// 清空 Cvm 信息，只保留 CBS
		cbsDemand.Updated.Cvm = Cvm{}
	}
	return cbsDemand
}

// ClearCbsDemand 清空 Cbs 需求量
func (d *ResPlanDemand) ClearCbsDemand() {
	if d.Original != nil {
		d.Original.Cbs.DiskSize = 0
	}
	if d.Updated != nil {
		d.Updated.Cbs.DiskSize = 0
	}
}

// OriginalRPDemandItem is original resource plan demand item.
type OriginalRPDemandItem struct {
	// DemandID HCM res plan demand ID
	DemandID string `json:"demand_id"`
	// CrpDemandID crp demand id, 兼容旧版本json
	CrpDemandID int64 `json:"crp_demand_id"`
	// ObsProject OBS项目类型
	ObsProject enumor.ObsProject `json:"obs_project" validate:"lte=64"`
	// ExpectTime 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01
	ExpectTime string `json:"expect_time" validate:"lte=64"`
	// ReturnPlanTime 期望退回时间，格式为YYYY-MM-DD，例如2024-01-01
	ReturnPlanTime string `json:"return_plan_time" validate:"lte=64"`
	// ZoneID 可用区ID
	ZoneID string `json:"zone_id" validate:"lte=64"`
	// ZoneName 可用区名称
	ZoneName string `json:"zone_name" validate:"lte=64"`
	// RegionID 地区/城市ID
	RegionID string `json:"region_id" validate:"lte=64"`
	// RegionName 地区/城市名称
	RegionName string `json:"region_name" validate:"lte=64"`
	// AreaName 地域名称
	AreaName string `json:"area_name" validate:"lte=64"`
	// Remark 需求备注
	Remark string `json:"remark" validate:"lte=255"`
	// Cvm cvm信息
	Cvm Cvm `json:"cvm"`
	// Cbs cbs信息
	Cbs Cbs `json:"cbs"`
}

// Validate whether OriginalRPDemandItem is valid.
func (i *OriginalRPDemandItem) Validate() error {
	if err := validator.Validate.Struct(i); err != nil {
		return err
	}

	if len(i.DemandID) <= 0 && i.CrpDemandID <= 0 {
		return errors.New("demand id can not be empty")
	}

	if err := i.ObsProject.ValidateResPlan(); err != nil {
		return err
	}

	if len(i.ExpectTime) == 0 {
		return errors.New("expect time can not be empty")
	}

	// NOTE: zone can be empty.

	if len(i.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(i.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(i.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	if err := i.Cvm.Validate(); err != nil {
		return err
	}

	if err := i.Cbs.Validate(); err != nil {
		return err
	}

	return nil
}

// Clone return a clone OriginalRPDemandItem.
func (i *OriginalRPDemandItem) Clone() *OriginalRPDemandItem {
	return &OriginalRPDemandItem{
		DemandID:       i.DemandID,
		CrpDemandID:    i.CrpDemandID,
		ObsProject:     i.ObsProject,
		ExpectTime:     i.ExpectTime,
		ReturnPlanTime: i.ReturnPlanTime,
		ZoneID:         i.ZoneID,
		ZoneName:       i.ZoneName,
		RegionID:       i.RegionID,
		RegionName:     i.RegionName,
		AreaName:       i.AreaName,
		Remark:         i.Remark,
		Cvm:            i.Cvm,
		Cbs:            i.Cbs,
	}
}

// UpdatedRPDemandItem is updated resource plan demand item.
type UpdatedRPDemandItem struct {
	// ObsProject OBS项目类型
	ObsProject enumor.ObsProject `json:"obs_project" validate:"lte=64"`
	// ExpectTime 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01
	ExpectTime string `json:"expect_time" validate:"lte=64"`
	// ReturnPlanTime 期望退回时间，格式为YYYY-MM-DD，例如2024-01-01
	ReturnPlanTime string `json:"return_plan_time" validate:"lte=64"`
	// ZoneID 可用区ID
	ZoneID string `json:"zone_id" validate:"lte=64"`
	// ZoneName 可用区名称
	ZoneName string `json:"zone_name" validate:"lte=64"`
	// RegionID 地区/城市ID
	RegionID string `json:"region_id" validate:"lte=64"`
	// RegionName 地区/城市名称
	RegionName string `json:"region_name" validate:"lte=64"`
	// AreaName 地域名称
	AreaName string `json:"area_name" validate:"lte=64"`
	// DemandSource 需求分类/变更原因
	DemandSource enumor.DemandSource `json:"demand_source" validate:"lte=64"`
	// Remark 需求备注
	Remark string `json:"remark" validate:"lte=255"`
	// Cvm cvm信息
	Cvm Cvm `json:"cvm"`
	// Cbs cbs信息
	Cbs Cbs `json:"cbs"`
}

// Validate whether UpdatedRPDemandItem is valid.
func (i *UpdatedRPDemandItem) Validate() error {
	if err := validator.Validate.Struct(i); err != nil {
		return err
	}

	if err := i.ObsProject.ValidateResPlan(); err != nil {
		return err
	}

	if len(i.ExpectTime) == 0 {
		return errors.New("expect time can not be empty")
	}

	// NOTE: zone can be empty.

	if len(i.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(i.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(i.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	if err := i.Cvm.Validate(); err != nil {
		return err
	}

	if err := i.Cbs.Validate(); err != nil {
		return err
	}

	return nil
}

// Clone return a clone UpdatedRPDemandItem.
func (i *UpdatedRPDemandItem) Clone() *UpdatedRPDemandItem {
	return &UpdatedRPDemandItem{
		ObsProject:     i.ObsProject,
		ExpectTime:     i.ExpectTime,
		ReturnPlanTime: i.ReturnPlanTime,
		ZoneID:         i.ZoneID,
		ZoneName:       i.ZoneName,
		RegionID:       i.RegionID,
		RegionName:     i.RegionName,
		AreaName:       i.AreaName,
		DemandSource:   i.DemandSource,
		Remark:         i.Remark,
		Cvm:            i.Cvm,
		Cbs:            i.Cbs,
	}
}

// Cvm is struct of ResPlanDemandTable's Cvm.
type Cvm struct {
	ResMode        enumor.ResMode `json:"res_mode"`
	DeviceType     string         `json:"device_type"`
	DeviceClass    string         `json:"device_class" validate:"lte=64"`
	DeviceFamily   string         `json:"device_family" validate:"lte=64"`
	TechnicalClass string         `json:"technical_class" validate:"lte=64"`
	CoreType       string         `json:"core_type" validate:"lte=64"`
	Os             types.Decimal  `json:"os"`
	CpuCore        int64          `json:"cpu_core"`
	Memory         int64          `json:"memory"`
}

// Validate whether Cvm is valid.
func (c *Cvm) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	if c.Os.IsNegative() {
		return errors.New("os should be >= 0")
	}

	if c.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if c.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	return nil
}

// IsEmpty checks if Cvm is empty (all fields are zero values).
// This is used to identify CBS-only demands.
func (c *Cvm) IsEmpty() bool {
	return c.DeviceType == "" && c.CpuCore == 0 && c.Memory == 0 && c.Os.IsZero()
}

// Cbs is struct of ResPlanDemandTable's Cbs.
type Cbs struct {
	DiskType     enumor.DiskType `json:"disk_type"`
	DiskTypeName string          `json:"disk_type_name"`
	DiskIo       int64           `json:"disk_io"`
	DiskSize     int64           `json:"disk_size"`
}

// Validate whether Cbs is valid.
func (c *Cbs) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	if c.DiskIo < 0 {
		return errors.New("disk io should be >= 0")
	}

	if c.DiskSize < 0 {
		return errors.New("disk size should be >= 0")
	}

	return nil
}

// TableName is the recycleRecord's database table name.
func (r ResPlanTicketTable) TableName() table.Name {
	return table.ResPlanTicketTable
}

// InsertValidate validate resource plan ticket on insertion.
func (r ResPlanTicketTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.Type) == 0 {
		return errors.New("type can not be empty")
	}

	if len(r.Demands) == 0 {
		return errors.New("demands can not be empty")
	}

	if len(r.Applicant) == 0 {
		return errors.New("applicant can not be empty")
	}

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

	if err := r.DemandClass.Validate(); err != nil {
		return err
	}

	if err := r.validateValue(); err != nil {
		return err
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if len(r.SubmittedAt) == 0 {
		return errors.New("submitted can not be empty")
	}

	return nil
}

func (r ResPlanTicketTable) validateValue() error {
	if r.OriginalOS < 0 {
		return errors.New("original os should be >= 0")
	}

	if r.OriginalCpuCore < 0 {
		return errors.New("original cpu core should be >= 0")
	}

	if r.OriginalMemory < 0 {
		return errors.New("original memory should be >= 0")
	}

	if r.OriginalDiskSize < 0 {
		return errors.New("original disk size should be >= 0")
	}

	if r.UpdatedOS < 0 {
		return errors.New("updated os should be >= 0")
	}

	if r.UpdatedCpuCore < 0 {
		return errors.New("updated cpu core should be >= 0")
	}

	if r.UpdatedMemory < 0 {
		return errors.New("updated memory should be >= 0")
	}

	if r.UpdatedDiskSize < 0 {
		return errors.New("updated disk size should be >= 0")
	}

	return nil
}

// UpdateValidate validate resource plan ticket on update.
func (r ResPlanTicketTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

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

	if len(r.DemandClass) > 0 {
		if err := r.DemandClass.Validate(); err != nil {
			return err
		}
	}

	if err := r.validateValue(); err != nil {
		return err
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
