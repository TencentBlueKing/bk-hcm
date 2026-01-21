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

// Package config ...
package config

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
)

// Requirement resource requirement type config
type Requirement struct {
	BkInstId    int64              `json:"id" bson:"id"`
	RequireType enumor.RequireType `json:"require_type" bson:"require_type"`
	RequireName string             `json:"require_name" bson:"require_name"`
	Position    int64              `json:"position" bson:"position"`
}

// GetRequirementResult get requirement type list result
type GetRequirementResult struct {
	Count int64          `json:"count"`
	Info  []*Requirement `json:"info"`
}

// Region qcloud resource region config
type Region struct {
	Region         string `json:"region" bson:"region"`
	RegionCn       string `json:"region_cn" bson:"region_cn"`
	CmdbRegionName string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// GetRegionResult get region list result
type GetRegionResult struct {
	Count int64     `json:"count"`
	Info  []*Region `json:"info"`
}

// Zone qcloud resource zone config
type Zone struct {
	Zone           string `json:"zone" bson:"zone"`
	ZoneCn         string `json:"zone_cn" bson:"zone_cn"`
	Region         string `json:"region" bson:"region"`
	RegionCn       string `json:"region_cn" bson:"region_cn"`
	CmdbRegionName string `json:"cmdb_region_name" bson:"cmdb_region_name"`
	CmdbZoneName   string `json:"cmdb_zone_name" bson:"cmdb_zone_name"`
}

// GetZoneParam get zone list request param
type GetZoneParam struct {
	Region     []string `json:"region" bson:"region"`
	Zone       []string `json:"zone" bson:"zone"`
	CmdbRegion []string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// GetZoneResult get zone list result
type GetZoneResult struct {
	Count int64   `json:"count"`
	Info  []*Zone `json:"info"`
}

// Vpc cvm vpc config
type Vpc struct {
	BkInstId string `json:"id" bson:"id"`
	Region   string `json:"region" bson:"region"`
	VpcId    string `json:"vpc_id" bson:"vpc_id"`
	VpcName  string `json:"vpc_name" bson:"vpc_name"`
}

// GetVpcParam get vpc list request param
type GetVpcParam struct {
	Region string `json:"region" bson:"region"`
}

// GetVpcResult get vpc list result
type GetVpcResult struct {
	Count int64  `json:"count"`
	Info  []*Vpc `json:"info"`
}

// GetVpcListParam get vpc id list request param
type GetVpcListParam struct {
	Regions []string `json:"regions" bson:"regions"`
}

// GetVpcListRst get vpc id list result
type GetVpcListRst struct {
	Info []interface{} `json:"info"`
}

// Subnet cvm subnet config
type Subnet struct {
	BkInstId   string `json:"id" bson:"id"`
	Region     string `json:"region" bson:"region"`
	Zone       string `json:"zone" bson:"zone"`
	VpcId      string `json:"vpc_id" bson:"vpc_id"`
	VpcName    string `json:"vpc_name" bson:"vpc_name"`
	SubnetId   string `json:"subnet_id" bson:"subnet_id"`
	SubnetName string `json:"subnet_name" bson:"subnet_name"`
	Enable     bool   `json:"enable" bson:"enable"`
	Comment    string `json:"comment"`
}

// GetSubnetParam get subnet list request param
type GetSubnetParam struct {
	Region string `json:"region" bson:"region"`
	Zone   string `json:"zone" bson:"zone"`
	Vpc    string `json:"vpc" bson:"vpc"`
}

// GetSubnetResult get subnet list result
type GetSubnetResult struct {
	Count int64     `json:"count"`
	Info  []*Subnet `json:"info"`
}

// GetSubnetListParam get subnet detail list request param
type GetSubnetListParam struct {
	Filter *filter.Expression `json:"filter"`
	Page   *core.BasePage     `json:"page"`
}

// Validate whether GetSubnetListParam is valid
func (param *GetSubnetListParam) Validate() error {
	if param.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	if param.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if err := param.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// UpdateSubnetPropertyParam update subnet property request param
type UpdateSubnetPropertyParam struct {
	Ids      []string               `json:"ids" bson:"ids"`
	Property map[string]interface{} `json:"properties"`
}

// Validate whether UpdateSubnetPropertyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateSubnetPropertyParam) Validate() error {
	limit := 200
	if len(param.Ids) > limit {
		return fmt.Errorf("ids exceed limit %d", limit)
	}

	return nil
}

// GetIdcZoneParam get idc zone list request param
type GetIdcZoneParam struct {
	Region []string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// IdcZone idc resource zone config
type IdcZone struct {
	BkInstId int64  `json:"id" bson:"id"`
	Region   string `json:"cmdb_region_name" bson:"cmdb_region_name"`
	ZoneName string `json:"cmdb_zone_name" bson:"cmdb_zone_name"`
	ZoneId   int64  `json:"cmdb_zone_id" bson:"cmdb_zone_id"`
}

// GetIdcRegionRst get idc region list result
type GetIdcRegionRst struct {
	Info []interface{} `json:"info"`
}

// GetIdcZoneRst get idc zone list result
type GetIdcZoneRst struct {
	Count int64      `json:"count"`
	Info  []*IdcZone `json:"info"`
}

// DeviceRestrict device restrict config
type DeviceRestrict struct {
	BkInstId int64   `json:"id" bson:"id"`
	Cpu      []int64 `json:"cpu" bson:"cpu"`
	Mem      []int64 `json:"mem" bson:"mem"`
	Disk     []int64 `json:"disk" bson:"disk"`
}

// GetDeviceRestrictResult get device info result
type GetDeviceRestrictResult struct {
	Cpu  []int64 `json:"cpu" bson:"cpu"`
	Mem  []int64 `json:"mem" bson:"mem"`
	Disk []int64 `json:"disk" bson:"disk"`
}

// CvmImage cvm image config
type CvmImage struct {
	BkInstId  int64  `json:"id" bson:"id"`
	Region    string `json:"region" bson:"region"`
	ImageId   string `json:"image_id" bson:"image_id"`
	ImageName string `json:"image_name" bson:"image_name"`
}

// GetCvmImageParam get cvm image list request param
type GetCvmImageParam struct {
	Region []string `json:"region" bson:"region"`
}

// GetCvmImageResult get zone list result
type GetCvmImageResult struct {
	Count int64       `json:"count"`
	Info  []*CvmImage `json:"info"`
}

// DeviceInfo cvm device info
type DeviceInfo struct {
	BkInstId       int64              `json:"id" bson:"id"`
	RequireType    enumor.RequireType `json:"require_type" bson:"require_type"`
	Region         string             `json:"region" bson:"region"`
	Zone           string             `json:"zone" bson:"zone"`
	DeviceType     string             `json:"device_type" bson:"device_type"`
	Cpu            int64              `json:"cpu" bson:"cpu"`
	Mem            int64              `json:"mem" bson:"mem"`
	Disk           int64              `json:"disk" bson:"disk"`
	Remark         string             `json:"remark" bson:"remark"`
	Label          mapstr.MapStr      `json:"label" bson:"label"`
	CapacityFlag   int                `json:"capacity_flag" bson:"capacity_flag"`
	EnableCapacity bool               `json:"enable_capacity" bson:"enable_capacity"`
	EnableApply    bool               `json:"enable_apply" bson:"enable_apply"`
	Score          float64            `json:"score" bson:"score"`
	Comment        string             `json:"comment" bson:"comment"`
	// DeviceTypeClass 通/专用机型，SpecialType专用，CommonType通用
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class" bson:"device_type_class"`
}

const (
	// CapLevelUndefined undefined capacity flag
	CapLevelUndefined int = 0
	// CapLevelEmpty "无库存" - 0
	CapLevelEmpty int = 1
	// CapLevelLow "库存紧张" - [1~10]
	CapLevelLow int = 2
	// CapLevelMedium "少量库存" - [11~50]
	CapLevelMedium int = 3
	// CapLevelHigh "库存充足" - [51~oo]
	CapLevelHigh int = 4
)

// GetDeviceParam get device list request param
type GetDeviceParam struct {
	Filter *querybuilder.QueryFilter `json:"filter" bson:"filter"`
	Page   metadata.BasePage         `json:"page" bson:"page"`
}

// Validate whether GetDeviceParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetDeviceParam) Validate() (errKey string, err error) {
	if key, err := param.Page.Validate(false); err != nil {
		return fmt.Sprintf("page.%s", key), err
	}

	if param.Filter != nil {
		if key, err := param.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if param.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetDeviceParam) GetFilter() (map[string]interface{}, error) {
	if param.Filter != nil {
		mgoFilter, key, err := param.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		return mgoFilter, nil
	}
	return make(map[string]interface{}), nil
}

// GetDeviceInfoResult get device info result
type GetDeviceInfoResult struct {
	Count int64         `json:"count"`
	Info  []*DeviceInfo `json:"info"`
}

// GetDeviceTypeResult get device type result
type GetDeviceTypeResult struct {
	Count int64            `json:"count"`
	Info  []DeviceTypeItem `json:"info"`
}

// DeviceTypeItem device type item
type DeviceTypeItem struct {
	DeviceType      string                   `json:"device_type"`       // 机型
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class"` // 通/专用机型，SpecialType专用，CommonType通用
	DeviceGroup     string                   `json:"device_group"`      // 机型族
	CPUAmount       float64                  `json:"cpu_amount"`        // CPU数量
	RamAmount       float64                  `json:"ram_amount"`        // 内存容量
	CoreType        int                      `json:"core_type"`         // 1.2.3 分别标识，小核心，中核心，大核心
	DeviceClass     string                   `json:"device_class"`      // 实例类型
	TechnicalClass  string                   `json:"technical_class"`   // 技术分类
}

// DeviceTypeInfo cvm device type info
type DeviceTypeInfo struct {
	DeviceType  string `json:"device_type"`
	DeviceGroup string `json:"device_group"`
}

// GetDeviceTypeDetailResult get device type detail result
type GetDeviceTypeDetailResult struct {
	Count int64             `json:"count"`
	Info  []*DeviceTypeInfo `json:"info"`
}

// CreateManyDeviceParam create device config in batch request param
type CreateManyDeviceParam struct {
	RequireType     []enumor.RequireType     `json:"require_type" validate:"required,max=20,dive"`
	Zone            []string                 `json:"zone" validate:"required,max=100,dive"`
	DeviceGroup     string                   `json:"device_group" validate:"required"`
	DeviceSize      enumor.CoreType          `json:"device_size" validate:"required"`
	DeviceType      string                   `json:"device_type" validate:"required"`
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class" validate:"omitempty"`
	Cpu             int64                    `json:"cpu" validate:"required,min=1"`
	Mem             int64                    `json:"mem" validate:"required,min=1"`
	Remark          string                   `json:"remark"`
	// ForceCreate 当机型在CRP中不存在时是否仍然创建
	ForceCreate    bool   `json:"force_create" validate:"omitempty"`
	TechnicalClass string `json:"technical_class"` // 技术分类
}

// Validate whether GetDeviceParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *CreateManyDeviceParam) Validate() error {
	if len(param.RequireType) == 0 {
		return errors.New("require_type empty or non-exist")
	}

	if len(param.RequireType) > 20 {
		return errors.New("require_type exceed limit 20")
	}

	for _, rt := range param.RequireType {
		if err := rt.Validate(); err != nil {
			return fmt.Errorf("require_type: %v", err)
		}
	}

	if len(param.Zone) == 0 {
		return errors.New("zone empty or non-exist")
	}

	if len(param.Zone) > 100 {
		return errors.New("zone exceed limit 100")
	}

	if param.DeviceGroup == "" {
		return errors.New("device_group empty or non-exist")
	}

	if param.DeviceSize == "" {
		return errors.New("device_size empty or non-exist")
	}

	if err := param.DeviceSize.Validate(); err != nil {
		return err
	}

	if param.DeviceType == "" {
		return errors.New("device_type empty or non-exist")
	}

	if param.Cpu <= 0 {
		return errors.New("cpu should be positive")
	}

	if param.Mem <= 0 {
		return errors.New("mem should be positive")
	}

	return validator.Validate.Struct(param)
}

// UpdateDevicePropertyParam update device property request param
type UpdateDevicePropertyParam struct {
	Ids      []int64                `json:"ids" bson:"ids"`
	Property map[string]interface{} `json:"properties"`
}

// Validate whether UpdateDevicePropertyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateDevicePropertyParam) Validate() error {
	limit := 200
	if len(param.Ids) > limit {
		return fmt.Errorf("ids exceed limit %d", limit)
	}

	return nil
}

// DvmDeviceInfo dvm device info
type DvmDeviceInfo struct {
	BkInstId    int64         `json:"id" bson:"id"`
	DeviceType  string        `json:"device_type" bson:"device_type"`
	Cpu         int64         `json:"cpu" bson:"cpu"`
	Mem         int64         `json:"mem" bson:"mem"`
	Disk        int64         `json:"disk" bson:"disk"`
	NetWork     string        `json:"network" bson:"network"`
	CpuProvider string        `json:"cpu_provider" bson:"cpu_provider"`
	Remark      string        `json:"remark" bson:"remark"`
	Label       mapstr.MapStr `json:"label" bson:"label"`
}

// GetDvmDeviceRst get dvm device result
type GetDvmDeviceRst struct {
	Count int64            `json:"count"`
	Info  []*DvmDeviceInfo `json:"info"`
}

// PmDeviceInfo physical machine device info
type PmDeviceInfo struct {
	BkInstId   int64         `json:"id" bson:"id"`
	DeviceType string        `json:"device_type" bson:"device_type"`
	Cpu        int64         `json:"cpu" bson:"cpu"`
	Mem        int64         `json:"mem" bson:"mem"`
	Raid       string        `json:"raid" bson:"raid"`
	NetWork    string        `json:"network" bson:"network"`
	Remark     string        `json:"remark" bson:"remark"`
	Label      mapstr.MapStr `json:"label" bson:"label"`
}

// GetPmDeviceRst get dvm device result
type GetPmDeviceRst struct {
	Count int64           `json:"count"`
	Info  []*PmDeviceInfo `json:"info"`
}

// GetCapacityParam get resource apply capacity request param
type GetCapacityParam struct {
	RequireType enumor.RequireType `json:"require_type"`
	DeviceType  string             `json:"device_type"`
	Region      string             `json:"region"`
	Zone        string             `json:"zone"`
	Vpc         string             `json:"vpc"`
	Subnet      string             `json:"subnet"`
	// 计费模式(计费模式：PREPAID包年包月，POSTPAID_BY_HOUR按量计费，默认为：PREPAID)
	ChargeType cvmapi.ChargeType `json:"charge_type"`
	// IgnorePrediction 获取容量时，是否忽略预测
	IgnorePrediction bool  `json:"ignore_prediction"`
	BizID            int64 `json:"bk_biz_id"`
	// 多可用区
	Zones []string `json:"zones"`
	// 是否插入或更新DB数据
	DisableUpsertDB bool `json:"disable_upsert_db"`
}

// Validate whether GetCapacityParam is valid
// err: detail reason why parameter is invalid
func (param *GetCapacityParam) Validate() error {
	if err := param.RequireType.Validate(); err != nil {
		return fmt.Errorf("require_type validation failed: %v", err)
	}

	if len(param.DeviceType) == 0 {
		return fmt.Errorf("device_type cannot be empty")
	}

	if len(param.Region) == 0 {
		return fmt.Errorf("region cannot be empty")
	}

	if len(param.Zone) == 0 {
		return fmt.Errorf("zone cannot be empty")
	}

	return nil
}

// GetCapacityRst get resource apply capacity result
type GetCapacityRst struct {
	Count int64           `json:"count"`
	Info  []*CapacityInfo `json:"info"`
}

// CapacityInfo resource apply capacity info
type CapacityInfo struct {
	Region  string             `json:"region"`
	Zone    string             `json:"zone"`
	Vpc     string             `json:"vpc"`
	Subnet  string             `json:"subnet"`
	MaxNum  int64              `json:"max_num"`
	MaxInfo []*CapacityMaxInfo `json:"max_info"`
}

// CapacityMaxInfo resource apply capacity max info
type CapacityMaxInfo struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

// BatchGetCapacityParam 批量获取资源申请容量请求参数
type BatchGetCapacityParam struct {
	RequireType enumor.RequireType `json:"require_type"`
	DeviceTypes []string           `json:"device_types"`
	Region      string             `json:"region"`
	Zones       []string           `json:"zones"`
	Vpc         string             `json:"vpc"`
	Subnet      string             `json:"subnet"`
	// 计费模式
	ChargeType cvmapi.ChargeType `json:"charge_type"`
	// 是否忽略预测
	IgnorePrediction bool  `json:"ignore_prediction"`
	BizID            int64 `json:"bk_biz_id"`
	// 是否禁用插入或更新DB数据
	DisableUpsertDB bool `json:"disable_upsert_db"`
}

// Validate 验证批量容量查询参数
func (param *BatchGetCapacityParam) Validate() error {
	if err := param.RequireType.Validate(); err != nil {
		return fmt.Errorf("require_type validation failed: %w", err)
	}

	if len(param.DeviceTypes) == 0 {
		return fmt.Errorf("device_types cannot be empty")
	}
	if len(param.DeviceTypes) > 100 {
		return fmt.Errorf("device_types length cannot exceed 100")
	}

	if len(param.Region) == 0 {
		return fmt.Errorf("region cannot be empty")
	}

	if len(param.Zones) == 0 {
		return fmt.Errorf("zones cannot be empty, must specify at least one zone to query capacity")
	}
	if len(param.Zones) > 100 {
		return fmt.Errorf("zones length cannot exceed 100")
	}

	if param.ChargeType != "" {
		if err := param.ChargeType.Validate(); err != nil {
			return fmt.Errorf("charge_type validation failed: %v", err)
		}
	}

	return nil
}

// BatchGetCapacityRst 批量获取资源申请容量结果
type BatchGetCapacityRst struct {
	Count int64                `json:"count"`
	Info  []*BatchCapacityInfo `json:"info"`
}

// BatchCapacityInfo 批量容量信息
type BatchCapacityInfo struct {
	DeviceType string             `json:"device_type"`
	Region     string             `json:"region"`
	Zone       string             `json:"zone"`
	Vpc        string             `json:"vpc"`
	Subnet     string             `json:"subnet"`
	MaxNum     int64              `json:"max_num"`
	MaxInfo    []*CapacityMaxInfo `json:"max_info"`
}

// UpdateCapacityParam update resource capacity info param
type UpdateCapacityParam struct {
	RequireType enumor.RequireType `json:"require_type"`
	DeviceType  string             `json:"device_type"`
	Region      string             `json:"region"`
	Zone        string             `json:"zone"`
}

// Validate whether UpdateCapacityParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateCapacityParam) Validate() (string, error) {
	if err := param.RequireType.Validate(); err != nil {
		key := "require_type"
		return key, err
	}

	if len(param.DeviceType) == 0 {
		key := "device_type"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Region) == 0 {
		key := "region"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Zone) == 0 {
		key := "zone"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	return "", nil
}

// GetAffinityParam get affinity request parameter
type GetAffinityParam struct {
	ResourceType string `json:"resource_type" bson:"resource_type"`
	HasZone      bool   `json:"has_zone" bson:"has_zone"`
}

// GetAffinityRst get affinity result
type GetAffinityRst struct {
	Info []*AffinityInfo `json:"info"`
}

// AffinityInfo affinity info
type AffinityInfo struct {
	Level       string `json:"level" bson:"level"`
	Description string `json:"description" bson:"description"`
}

// 资源类型
const (
	ResourceTypePm        string = "IDCPM"
	ResourceTypeCvm       string = "QCLOUDCVM"
	ResourceTypeIdcDvm    string = "IDCDVM"
	ResourceTypeQcloudDvm string = "QCLOUDDVM"
)

// Anti类型
const (
	AntiNone   string = "ANTI_NONE"   // 无要求
	AntiRack   string = "ANTI_RACK"   // 分机架
	AntiModule string = "ANTI_MODULE" // 分Module
	AntiCampus string = "ANTI_CAMPUS" // 分Campus
)

// Description description of terms
var Description = map[string]string{
	AntiNone:   "无要求",
	AntiRack:   "分机架",
	AntiModule: "分Module",
	AntiCampus: "分Campus",
}

// DeviceTypeCpuItem device type cpu item
type DeviceTypeCpuItem struct {
	DeviceType     string          `json:"device_type"`     // 机型
	CPUAmount      int64           `json:"cpu_amount"`      // CPU数量
	DeviceGroup    string          `json:"device_group"`    // 机型族
	CoreType       enumor.CoreType `json:"core_type"`       // 机型核心类型
	TechnicalClass string          `json:"technical_class"` // 技术分类
}

// UpsertRegionDftVpcReq upsert region default vpc request.
type UpsertRegionDftVpcReq struct {
	RegionDftVpcInfos []RegionDftVpc `json:"region_dft_vpc_infos" validate:"min=1,max=100"`
}

// Validate ...
func (u *UpsertRegionDftVpcReq) Validate() error {
	if err := validator.Validate.Struct(u); err != nil {
		return err
	}

	for _, v := range u.RegionDftVpcInfos {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DftVpc default vpc.
type DftVpc struct {
	VpcID string `json:"vpc_id" validate:"required"`
}

// RegionDftVpc region default vpc.
type RegionDftVpc struct {
	Region string `json:"region" validate:"required"`
	DftVpc
}

// Validate ...
func (req RegionDftVpc) Validate() error {
	if err := validator.Validate.Struct(req.DftVpc); err != nil {
		return err
	}

	return validator.Validate.Struct(req)
}

// UpsertRegionDftSgReq upsert region default security group request.
type UpsertRegionDftSgReq struct {
	RegionDftSgInfos []RegionDftSg `json:"region_dft_sg_infos" validate:"min=1,max=100"`
}

// Validate ...
func (u *UpsertRegionDftSgReq) Validate() error {
	if err := validator.Validate.Struct(u); err != nil {
		return err
	}

	for _, v := range u.RegionDftSgInfos {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DftSecurityGroup default security group.
type DftSecurityGroup struct {
	SgID   string `json:"security_group_id" validate:"required"`
	SgName string `json:"security_group_name" validate:"omitempty"`
	SgDesc string `json:"security_group_desc" validate:"omitempty"`
}

// RegionDftSg region default security group.
type RegionDftSg struct {
	Region string `json:"region" validate:"required"`
	DftSecurityGroup
}

// Validate ...
func (req RegionDftSg) Validate() error {
	if err := validator.Validate.Struct(req.DftSecurityGroup); err != nil {
		return err
	}

	return validator.Validate.Struct(req)
}

// DeviceFuzzyMatchInfo 设备模糊匹配信息
type DeviceFuzzyMatchInfo struct {
	TechnicalClass string          `json:"technical_class"` // 技术分类
	CoreType       enumor.CoreType `json:"core_type"`       // 机型核心类型
}

// GetAllSubnetReq get all subnet request param
type GetAllSubnetReq struct {
	Region     string   `json:"region"`
	Zones      []string `json:"zones"`
	CloudVpcID string   `json:"cloud_vpc_id"`
	CloudID    string   `json:"cloud_id"`
	Name       string   `json:"name"`
}

// UpsertDeviceCapacityItem 需要更新或创建的库存项
type UpsertDeviceCapacityItem struct {
	RequireType enumor.RequireType `json:"require_type"`
	DeviceType  string             `json:"device_type"`
	Region      string             `json:"region"`
	Zone        string             `json:"zone"`
	MaxNum      int64              `json:"max_num"`
	MaxInfo     []*CapacityMaxInfo `json:"max_info"`
}

// Validate ...
func (u *UpsertDeviceCapacityItem) Validate() error {
	if err := u.RequireType.Validate(); err != nil {
		return fmt.Errorf("require_type validation failed: %v", err)
	}

	if len(u.DeviceType) == 0 {
		return fmt.Errorf("device_type cannot be empty")
	}

	if len(u.Region) == 0 {
		return fmt.Errorf("region cannot be empty")
	}

	if len(u.Zone) == 0 {
		return fmt.Errorf("zone cannot be empty")
	}

	if u.MaxNum < 0 {
		return fmt.Errorf("max_num cannot be less than 0")
	}

	return nil
}

// UpsertSpringResPoolChargeTypeReq upsert spring resource pool charge type request
type UpsertSpringResPoolChargeTypeReq struct {
	BizID      *int64            `json:"bk_biz_id"`                       // 业务ID，如果为空则更新全局配置
	ChargeType cvmapi.ChargeType `json:"charge_type" validate:"required"` // 计费模式
}

// Validate ...
func (u *UpsertSpringResPoolChargeTypeReq) Validate() error {
	if err := validator.Validate.Struct(u); err != nil {
		return err
	}

	if err := u.ChargeType.Validate(); err != nil {
		return fmt.Errorf("charge_type validation failed: %w", err)
	}

	return nil
}

// DelSpringResPoolChargeTypeReq delete spring resource pool charge type request
type DelSpringResPoolChargeTypeReq struct {
	BizID *int64 `json:"bk_biz_id"` // 业务ID，如果为空则删除全局配置
}

// Validate ...
func (d *DelSpringResPoolChargeTypeReq) Validate() error {
	return validator.Validate.Struct(d)
}
