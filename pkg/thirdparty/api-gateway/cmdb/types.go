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

package cmdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/thirdparty/esb/types"
)

// ----------------------------- biz -----------------------------

// SearchBizParams is cmdb search business parameter.
type SearchBizParams struct {
	Fields            []string     `json:"fields"`
	Page              BasePage     `json:"page"`
	BizPropertyFilter *QueryFilter `json:"biz_property_filter,omitempty"`
}

// QueryFilter is cmdb common query filter.
type QueryFilter struct {
	Rule `json:",inline"`
}

// Rule is cmdb common query rule type.
type Rule interface {
	GetDeep() int
}

// CombinedRule is cmdb query rule that is combined by multiple AtomRule.
type CombinedRule struct {
	Condition Condition `json:"condition"`
	Rules     []Rule    `json:"rules"`
}

// Condition cmdb condition
type Condition string

const (
	// ConditionAnd and
	ConditionAnd = Condition("AND")
)

// GetDeep get query rule depth.
func (r CombinedRule) GetDeep() int {
	maxChildDeep := 1
	for _, child := range r.Rules {
		childDeep := child.GetDeep()
		if childDeep > maxChildDeep {
			maxChildDeep = childDeep
		}
	}
	return maxChildDeep + 1
}

// Combined ...
func Combined(cond Condition, rules ...Rule) CombinedRule {
	return CombinedRule{
		Condition: cond,
		Rules:     rules,
	}
}

// Equal ...
func Equal(field string, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: OperatorEqual,
		Value:    value,
	}
}

// In ...
func In(field string, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: OperatorIn,
		Value:    value,
	}
}

// Atomic ...
func Atomic(field string, op Operator, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: op,
		Value:    value,
	}
}

// AtomRule is cmdb atomic query rule.
type AtomRule struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// Operator cmdb operator
type Operator string

var (
	// OperatorEqual ...
	OperatorEqual = Operator("equal")
	// OperatorNotEqual ...
	OperatorNotEqual = Operator("not_equal")
	// OperatorIn ...
	OperatorIn = Operator("in")
)

// GetDeep get query rule depth.
func (r AtomRule) GetDeep() int {
	return 1
}

// MarshalJSON marshal QueryFilter to json.
func (qf *QueryFilter) MarshalJSON() ([]byte, error) {
	if qf.Rule != nil {
		return json.Marshal(qf.Rule)
	}
	return make([]byte, 0), nil
}

// BasePage is cmdb paging parameter.
type BasePage struct {
	Sort        string `json:"sort,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Start       int64  `json:"start"`
	EnableCount bool   `json:"enable_count,omitempty"`
}

// SearchBizResp is cmdb search business response.
type SearchBizResp struct {
	types.BaseResponse
	SearchBizResult `json:"data"`
}

// SearchBizResult is cmdb search business response.
type SearchBizResult struct {
	Count int64 `json:"count"`
	Info  []Biz `json:"info"`
}

// Biz is cmdb biz info.
type Biz struct {
	BizID   int64  `json:"bk_biz_id"`
	BizName string `json:"bk_biz_name"`
	// 运营产品信息
	BkProductID   int64  `json:"bk_product_id"`
	BkProductName string `json:"bk_product_name"`
	// 二级业务id
	BsName2ID int64 `json:"bs2_name_id"`
	// 运维负责人
	BkBizMaintainer string `json:"bk_biz_maintainer"`
	BkOperGrpNameID int64  `json:"bk_oper_grp_name_id"`
}

// -------------------------- cloud area --------------------------

// SearchCloudAreaParams is cmdb search cloud area parameter.
type SearchCloudAreaParams struct {
	Fields    []string               `json:"fields"`
	Page      BasePage               `json:"page"`
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// SearchCloudAreaResp is cmdb search cloud area response.
type SearchCloudAreaResp struct {
	types.BaseResponse `json:",inline"`
	Data               *SearchCloudAreaResult `json:"data"`
}

// SearchCloudAreaResult is cmdb search cloud area result.
type SearchCloudAreaResult struct {
	Count int64       `json:"count"`
	Info  []CloudArea `json:"info"`
}

// CloudArea is cmdb cloud area info.
type CloudArea struct {
	CloudID   int64  `json:"bk_cloud_id"`
	CloudName string `json:"bk_cloud_name"`
}

// ---------------------------- create ----------------------------

// BatchCreateResp cmdb's basic batch create resource response.
type BatchCreateResp struct {
	types.BaseResponse `json:",inline"`
	Data               *BatchCreateResult `json:"data"`
}

// BatchCreateResult cmdb's basic batch create resource result.
type BatchCreateResult struct {
	IDs []int64 `json:"ids"`
}

// ----------------------------- host -----------------------------

// AddCloudHostToBizParams is esb add cloud host to biz parameter.
type AddCloudHostToBizParams struct {
	BizID    int64             `json:"bk_biz_id" validate:"required"`
	HostInfo []HostCreateParam `json:"host_info" validate:"required,min=1,max=200,dive"`
}

// Validate validate AddCloudHostToBizParams
func (p *AddCloudHostToBizParams) Validate() error {
	return validator.Validate.Struct(p)
}

// HostCreateParam is cmdb host create parameter.
type HostCreateParam struct {
	BkHostID          int64           `json:"bk_host_id"`
	BkCloudVendor     CloudVendor     `json:"bk_cloud_vendor" validate:"required"`
	BkCloudInstID     string          `json:"bk_cloud_inst_id" validate:"required"`
	BkCloudHostStatus CloudHostStatus `json:"bk_cloud_host_status,omitempty"`
	BkCloudID         int64           `json:"bk_cloud_id" validate:"required"`
	// 云上地域，如 "ap-guangzhou"
	BkCloudRegion   string  `json:"bk_cloud_region"`
	BkHostInnerIP   string  `json:"bk_host_innerip" validate:"required"`
	BkHostOuterIP   string  `json:"bk_host_outerip"`
	BkHostInnerIPv6 string  `json:"bk_host_innerip_v6"`
	BkHostOuterIPv6 string  `json:"bk_host_outerip_v6"`
	Operator        string  `json:"operator"`
	BkBakOperator   string  `json:"bk_bak_operator"`
	BkHostName      string  `json:"bk_host_name"`
	BkComment       *string `json:"bk_comment,omitempty"`
}

// DeleteCloudHostFromBizParams is esb delete cloud host from biz parameter.
type DeleteCloudHostFromBizParams struct {
	BizID   int64   `json:"bk_biz_id" validate:"required"`
	HostIDs []int64 `json:"bk_host_ids" validate:"required,min=1,max=200"`
}

// Validate validate DeleteCloudHostFromBizParams
func (p *DeleteCloudHostFromBizParams) Validate() error {
	return validator.Validate.Struct(p)
}

// ListBizHostParams is esb list cmdb host in biz parameter.
type ListBizHostParams struct {
	BizID              int64           `json:"bk_biz_id" validate:"required"`
	BkSetIDs           []int64         `json:"bk_set_ids"`
	BkModuleIDs        []int64         `json:"bk_module_ids"`
	ModuleCond         []ConditionItem `json:"module_cond"`
	Fields             []string        `json:"fields"`
	Page               *BasePage       `json:"page" validate:"required"`
	HostPropertyFilter *QueryFilter    `json:"host_property_filter,omitempty"`
}

// Validate validate ListBizHostParams
func (p *ListBizHostParams) Validate() error {
	if len(p.BkModuleIDs) > 0 && len(p.ModuleCond) > 0 {
		return fmt.Errorf("bk_module_ids and module_cond can not be used at the same time")
	}
	return validator.Validate.Struct(p)
}

// ListBizHostResp is cmdb list cmdb host in biz response.
type ListBizHostResp struct {
	types.BaseResponse
	*ListBizHostResult `json:"data"`
}

// ListBizHostResult is cmdb list cmdb host in biz result.
type ListBizHostResult struct {
	Count int64  `json:"count"`
	Info  []Host `json:"info"`
}

// Host defines cmdb host info.
type Host struct {
	BkHostID          int64           `json:"bk_host_id"`
	BkCloudVendor     CloudVendor     `json:"bk_cloud_vendor" validate:"required"`
	BkCloudInstID     string          `json:"bk_cloud_inst_id" validate:"required"`
	BkCloudHostStatus CloudHostStatus `json:"bk_cloud_host_status,omitempty"`
	BkCloudID         int64           `json:"bk_cloud_id" validate:"required"`
	// 云上地域，如 "ap-guangzhou"
	BkCloudRegion   string  `json:"bk_cloud_region"`
	BkHostInnerIP   string  `json:"bk_host_innerip" validate:"required"`
	BkHostOuterIP   string  `json:"bk_host_outerip"`
	BkHostInnerIPv6 string  `json:"bk_host_innerip_v6"`
	BkHostOuterIPv6 string  `json:"bk_host_outerip_v6"`
	Operator        string  `json:"operator"`
	BkBakOperator   string  `json:"bk_bak_operator"`
	BkHostName      string  `json:"bk_host_name"`
	BkComment       *string `json:"bk_comment,omitempty"`
	BkOSName        string  `json:"bk_os_name,omitempty"`
	BkMac           string  `json:"bk_mac,omitempty"`
	CreateTime      string  `json:"create_time,omitempty"`

	// 以下字段仅内部版支持，由cc从云梯获取
	SvrSourceTypeID    SvrSourceTypeID `json:"bk_svr_source_type_id,omitempty"`
	BkAssetID          string          `json:"bk_asset_id,omitempty"`
	SvrDeviceClassName string          `json:"bk_svr_device_cls_name,omitempty"`
	BkCloudZone        string          `json:"bk_cloud_zone,omitempty"`
	BkCloudVpcID       string          `json:"bk_cloud_vpc_id,omitempty"`
	BkCloudSubnetID    string          `json:"bk_cloud_subnet_id,omitempty"`
	// 外网运营商
	BkIpOerName string `json:"bk_ip_oper_name,omitempty"`
	// 机型
	SvrDeviceClass string `json:"svr_device_class,omitempty"`
	// 服务器类型名称
	SvrTypeName string `json:"svr_type_name,omitempty"`
	// 操作系统类型
	BkOsType OsType `json:"bk_os_type,omitempty"`
	// 操作系统版本
	BkOsVersion string `json:"bk_os_version,omitempty"`
	// IDC区域
	BkIdcArea string `json:"bk_idc_area,omitempty"`
	// 地域
	BkZoneName string `json:"bk_zone_name,omitempty"`
	// 可用区(子Zone)
	SubZone string `json:"sub_zone,omitempty"`
	// 子ZoneID
	SubZoneId  string `json:"sub_zone_id,omitempty"`
	ModuleName string `json:"module_name,omitempty"`
	// 机架号
	RackId      string `json:"rack_id,omitempty"`
	IdcUnitName string `json:"idc_unit_name,omitempty"`
	// 逻辑区域
	LogicDomain string `json:"logic_domain,omitempty"`
	RaidName    string `json:"raid_name,omitempty"`
	// 机器上架时间，格式如"2018-05-07T00:00:00+08:00"
	SvrInputTime string `json:"svr_input_time,omitempty"`
	// 状态
	SrvStatus string `json:"srv_status,omitempty"`
	// 磁盘容量
	BkDisk float64 `json:"bk_disk,omitempty"`
	// CPU逻辑核心数
	BkCpu int64 `json:"bk_cpu,omitempty"`
	// 实例计费模式
	InstanceChargeType string `json:"instance_charge_type,omitempty"`
	// 套餐计费起始时间
	BillingStartTime time.Time `json:"billing_start_time,omitempty"`
	// 套餐计费过期时间
	BillingExpireTime time.Time `json:"billing_expire_time,omitempty"`
	// 运维部门
	DeptName string `json:"dept_name,omitempty"`
	// 母机固资号
	BKSvrOwnerAssetID string `json:"bk_svr_owner_asset_id,omitempty"`
}

// HostWithCloudID defines cmdb host with cloud id.
type HostWithCloudID struct {
	Host
	BizID   int64  `json:"bk_biz_id"`
	CloudID string `json:"cloud_id"`
}

// GetCloudID ...
func (h HostWithCloudID) GetCloudID() string {
	return h.CloudID
}

// HostFields cmdb common fields
var HostFields = []string{
	"bk_cloud_inst_id",
	"bk_host_id",
	"bk_asset_id",
	// 云地域
	"bk_cloud_region",
	// 云厂商
	"bk_cloud_vendor",
	"bk_host_innerip",
	"bk_host_outerip",
	"bk_host_innerip_v6",
	"bk_host_outerip_v6",
	"bk_cloud_host_status",

	"bk_host_name",
	"bk_cloud_id",
	"bk_os_name",
	"bk_svr_source_type_id",
	"bk_svr_device_cls_name",
	"svr_source_type_id", // 服务器来源类型ID
	"srv_status",         // CC的运营状态
	"svr_device_class",   // 机型
	"bk_disk",            // 磁盘容量(GB)
	"bk_cpu",             // CPU逻辑核心数
	"operator",           // 主要维护人
	"bk_bak_operator",    // 备份维护人

	// 以下字段仅内部版支持，由cc从云梯获取
	"bk_cloud_vpc_id",
	"bk_cloud_subnet_id",
	"bk_cloud_zone",
}

// FindHostTopoRelationParams cmdb find host topo request params
type FindHostTopoRelationParams struct {
	BizID       int64     `json:"bk_biz_id" validate:"required"`
	BkSetIDs    []int64   `json:"bk_set_ids,omitempty"`
	BkModuleIDs []int64   `json:"bk_module_ids,omitempty"`
	HostIDs     []int64   `json:"bk_host_ids"`
	Page        *BasePage `json:"page" validate:"required"`
}

// Validate validate FindHostTopoRelationParams
func (p *FindHostTopoRelationParams) Validate() error {
	return validator.Validate.Struct(p)
}

// HostTopoRelationResult cmdb host topo relation result warp
type HostTopoRelationResult struct {
	Count int64              `json:"count"`
	Page  BasePage           `json:"page"`
	Data  []HostTopoRelation `json:"data"`
}

// HostTopoRelation cmdb host topo relation
type HostTopoRelation struct {
	BizID             int64  `json:"bk_biz_id"`
	BkSetID           int64  `json:"bk_set_id"`
	BkModuleID        int64  `json:"bk_module_id"`
	HostID            int64  `json:"bk_host_id"`
	BkSupplierAccount string `json:"bk_supplier_account"`
}

// SearchModuleParams cmdb module search parameter.
type SearchModuleParams struct {
	BizID             int64  `json:"bk_biz_id" validate:"required"`
	BkSetID           int64  `json:"bk_set_id,omitempty"`
	BkSupplierAccount string `json:"bk_supplier_account,omitempty"`

	Fields    []string               `json:"fields"`
	Page      BasePage               `json:"page"`
	Condition map[string]interface{} `json:"condition"`
}

// Validate validate SearchModuleParams
func (s *SearchModuleParams) Validate() error {
	return validator.Validate.Struct(s)
}

// ModuleInfoResult cmdb module info list result
type ModuleInfoResult struct {
	Count int64         `json:"count"`
	Info  []*ModuleInfo `json:"info"`
}

// ModuleInfo cmdb module info
type ModuleInfo struct {
	BkSetID      int64  `json:"bk_set_id"`
	BkModuleName string `json:"bk_module_name"`
	Default      int64  `json:"default"`

	BkModuleId int64 `json:"bk_module_id"`
	Bs3NameID  int   `json:"bs3_name_id"`
}

// CloudVendor defines cmdb cloud vendor type.
type CloudVendor string

const (
	// AwsCloudVendor cmdb aws vendor
	AwsCloudVendor CloudVendor = "1"
	// TCloudCloudVendor cmdb cloud vendor
	TCloudCloudVendor CloudVendor = "2"
	// GcpCloudVendor cmdb gcp vendor
	GcpCloudVendor CloudVendor = "3"
	// AzureCloudVendor cmdb azure vendor
	AzureCloudVendor CloudVendor = "4"
	// HuaWeiCloudVendor cmdb huawei vendor
	HuaWeiCloudVendor CloudVendor = "15"
)

// HcmCmdbVendorMap is hcm vendor to cmdb cloud vendor map.
var HcmCmdbVendorMap = map[enumor.Vendor]CloudVendor{
	enumor.Aws:    AwsCloudVendor,
	enumor.TCloud: TCloudCloudVendor,
	enumor.Gcp:    GcpCloudVendor,
	enumor.Azure:  AzureCloudVendor,
	enumor.HuaWei: HuaWeiCloudVendor,
}

// CmdbHcmVendorMap cmdb vendor to hcm vendor
var CmdbHcmVendorMap = map[CloudVendor]enumor.Vendor{
	AwsCloudVendor:    enumor.Aws,
	TCloudCloudVendor: enumor.TCloud,
	GcpCloudVendor:    enumor.Gcp,
	AzureCloudVendor:  enumor.Azure,
	HuaWeiCloudVendor: enumor.HuaWei,
}

// CloudHostStatus defines cmdb cloud host status type.
type CloudHostStatus string

const (
	// UnknownCloudHostStatus ...
	UnknownCloudHostStatus CloudHostStatus = "1"
	// StartingCloudHostStatus ...
	StartingCloudHostStatus CloudHostStatus = "2"
	// RunningCloudHostStatus ...
	RunningCloudHostStatus CloudHostStatus = "3"
	// StoppingCloudHostStatus ...
	StoppingCloudHostStatus CloudHostStatus = "4"
	// StoppedCloudHostStatus ...
	StoppedCloudHostStatus CloudHostStatus = "5"
	// TerminatedCloudHostStatus ...
	TerminatedCloudHostStatus CloudHostStatus = "6"
)

// HcmCmdbHostStatusMap is hcm vendor to cmdb cloud host status map.
var HcmCmdbHostStatusMap = map[enumor.Vendor]map[string]CloudHostStatus{
	enumor.TCloud: TCloudCmdbStatusMap,
	enumor.Aws:    AwsCmdbStatusMap,
	enumor.Gcp:    GcpCmdbStatusMap,
	enumor.Azure:  AzureCmdbStatusMap,
	enumor.HuaWei: HuaWeiCmdbStatusMap,
}

// TCloudCmdbStatusMap is tcloud status to cmdb cloud host status map.
var TCloudCmdbStatusMap = map[string]CloudHostStatus{
	"PENDING":       UnknownCloudHostStatus,
	"LAUNCH_FAILED": UnknownCloudHostStatus,
	"RUNNING":       RunningCloudHostStatus,
	"STOPPED":       StoppedCloudHostStatus,
	"STARTING":      StartingCloudHostStatus,
	"STOPPING":      StoppingCloudHostStatus,
	"REBOOTING":     UnknownCloudHostStatus,
	"SHUTDOWN":      StoppedCloudHostStatus,
	"TERMINATING":   TerminatedCloudHostStatus,
}

// AwsCmdbStatusMap is aws status to cmdb cloud host status map.
var AwsCmdbStatusMap = map[string]CloudHostStatus{
	"pending":       UnknownCloudHostStatus,
	"running":       RunningCloudHostStatus,
	"shutting-down": StoppingCloudHostStatus,
	"terminated":    TerminatedCloudHostStatus,
	"stopping":      StoppingCloudHostStatus,
	"stopped":       StoppedCloudHostStatus,
}

// GcpCmdbStatusMap is gcp status to cmdb cloud host status map.
var GcpCmdbStatusMap = map[string]CloudHostStatus{
	"PROVISIONING": UnknownCloudHostStatus,
	"STAGING":      StartingCloudHostStatus,
	"RUNNING":      RunningCloudHostStatus,
	"STOPPING":     StoppingCloudHostStatus,
	"SUSPENDING":   StoppingCloudHostStatus,
	"SUSPENDED":    StoppedCloudHostStatus,
	"REPAIRING":    UnknownCloudHostStatus,
	"TERMINATED":   TerminatedCloudHostStatus,
}

// AzureCmdbStatusMap is azure status to cmdb cloud host status map.
var AzureCmdbStatusMap = map[string]CloudHostStatus{
	"PowerState/running":      RunningCloudHostStatus,
	"PowerState/stopped":      StoppedCloudHostStatus,
	"PowerState/deallocating": StoppingCloudHostStatus,
	"PowerState/deallocated":  StoppedCloudHostStatus,
}

// HuaWeiCmdbStatusMap is huawei status to cmdb cloud host status map.
var HuaWeiCmdbStatusMap = map[string]CloudHostStatus{
	"BUILD":             UnknownCloudHostStatus,
	"REBOOT":            UnknownCloudHostStatus,
	"HARD_REBOOT":       UnknownCloudHostStatus,
	"REBUILD":           UnknownCloudHostStatus,
	"MIGRATING":         UnknownCloudHostStatus,
	"RESIZE":            UnknownCloudHostStatus,
	"ACTIVE":            RunningCloudHostStatus,
	"SHUTOFF":           StoppedCloudHostStatus,
	"REVERT_RESIZE":     UnknownCloudHostStatus,
	"VERIFY_RESIZE":     UnknownCloudHostStatus,
	"ERROR":             UnknownCloudHostStatus,
	"DELETED":           TerminatedCloudHostStatus,
	"SHELVED":           UnknownCloudHostStatus,
	"SHELVED_OFFLOADED": UnknownCloudHostStatus,
	"UNKNOWN":           UnknownCloudHostStatus,
}

// EventType is cmdb watch event type.
type EventType string

const (
	// Create is cmdb watch event create type.
	Create EventType = "create"
	// Update is cmdb watch event update type.
	Update EventType = "update"
	// Delete is cmdb watch event delete type.
	Delete EventType = "delete"
)

// CursorType is cmdb watch event cursor type.
type CursorType string

const (
	// HostType is cmdb watch event host cursor type.
	HostType CursorType = "host"
	// HostRelation is cmdb watch event host relation cursor type.
	HostRelation CursorType = "host_relation"
)

// WatchEventParams is esb watch cmdb event parameter.
type WatchEventParams struct {
	// event types you want to care, empty means all.
	EventTypes []EventType `json:"bk_event_types"`
	// the fields you only care, if nil, means all.
	Fields []string `json:"bk_fields"`
	// unix seconds timesss to where you want to watch from.
	// it's like Cursor, but StartFrom and Cursor can not use at the same time.
	StartFrom int64 `json:"bk_start_from"`
	// the cursor you hold previous, means you want to watch event form here.
	Cursor string `json:"bk_cursor"`
	// the resource kind you want to watch
	Resource CursorType       `json:"bk_resource" validate:"required"`
	Filter   WatchEventFilter `json:"bk_filter"`
}

// Validate validate WatchEventParams
func (p *WatchEventParams) Validate() error {
	return validator.Validate.Struct(p)
}

// WatchEventFilter watch event filter
type WatchEventFilter struct {
	// SubResource the sub resource you want to watch, e.g. object ID of the instance resource, watch all if not set
	SubResource string `json:"bk_sub_resource,omitempty"`
}

// CCErrEventChainNodeNotExist 如果事件节点不存在，cc会返回该错误码
var CCErrEventChainNodeNotExist = "1103007"

// WatchEventResult is cmdb watch event result.
type WatchEventResult struct {
	// watched events or not
	Watched bool               `json:"bk_watched"`
	Events  []WatchEventDetail `json:"bk_events"`
}

// WatchEventDetail is cmdb watch event detail.
type WatchEventDetail struct {
	Cursor    string          `json:"bk_cursor"`
	Resource  CursorType      `json:"bk_resource"`
	EventType EventType       `json:"bk_event_type"`
	Detail    json.RawMessage `json:"bk_detail"`
}

// HostModuleRelationParams get host and module relation parameter
type HostModuleRelationParams struct {
	BizID  int64   `json:"bk_biz_id,omitempty"`
	HostID []int64 `json:"bk_host_id" validate:"required,min=1,max=500"`
}

// Validate validate HostModuleRelationParams
func (p *HostModuleRelationParams) Validate() error {
	return validator.Validate.Struct(p)
}

// GetBizBriefCacheTopoParams define get biz brief cache topo params.
type GetBizBriefCacheTopoParams struct {
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
}

// Validate get biz brief cache topo params.
func (p *GetBizBriefCacheTopoParams) Validate() error {
	return validator.Validate.Struct(p)
}

// GetBizBriefCacheTopoResult define get biz brief cache topo result.
type GetBizBriefCacheTopoResult struct {
	// basic business info
	Biz *BizBase `json:"biz"`
	// the idle set nodes info
	Idle []Node `json:"idle"`
	// the other common nodes
	Nodes []Node `json:"nds"`
}

// Node define node info.
type Node struct {
	// the object of this node, like set or module
	Object string `json:"object_id"`
	// the node's instance id, like set id or module id
	ID int64 `json:"id"`
	// the node's name, like set name or module name
	Name string `json:"name"`
	// only set, module has this field.
	// describe what kind of set or module this node is.
	// 0: normal module or set.
	// >1: special set or module
	Default *int `json:"type,omitempty"`
	// the sub-nodes of current node
	SubNodes []Node `json:"nds"`
}

// BizBase define biz base.
type BizBase struct {
	// business id
	ID int64 `json:"id" bson:"bk_biz_id"`
	// business name
	Name string `json:"name" bson:"bk_biz_name"`
	// describe it's a resource pool business or normal business.
	// 0: normal business
	// >0: special business, like resource pool business.
	Default int `json:"type" bson:"default"`

	OwnerID string `json:"bk_supplier_account" bson:"bk_supplier_account"`
}

// ListHostWithoutBizParams is esb list cmdb host without biz parameter.
type ListHostWithoutBizParams struct {
	Fields             []string     `json:"fields"`
	Page               *BasePage    `json:"page" validate:"required"`
	HostPropertyFilter *QueryFilter `json:"host_property_filter,omitempty"`
}

// Validate validate ListHostReq
func (req *ListHostWithoutBizParams) Validate() error {
	return validator.Validate.Struct(req)
}

// ListHostWithoutBizResult is cmdb list cmdb host without biz result.
type ListHostWithoutBizResult struct {
	Count int64  `json:"count"`
	Info  []Host `json:"info"`
}

// BkAddressing cc主机寻址方式.
type BkAddressing string

const (
	// StaticAddressing 静态寻址
	StaticAddressing BkAddressing = "static"
	// DynamicAddressing 动态寻址
	DynamicAddressing BkAddressing = "dynamic"
)

// ListResourcePoolHostsParams list resource pool hosts parameter
type ListResourcePoolHostsParams struct {
	Fields             []string     `json:"fields"`
	Page               *BasePage    `json:"page"`
	HostPropertyFilter *QueryFilter `json:"host_property_filter,omitempty"`
}

// ListResourcePoolHostsResult list resource pool hosts result
type ListResourcePoolHostsResult struct {
	Count int64  `json:"count"`
	Info  []Host `json:"info"`
}

// UpdateCvmOSReq ...
type UpdateCvmOSReq struct {
	BkAssetId string             `json:"bk_asset_id"`
	Data      UpdateCvmOSReqData `json:"data"`
}

// UpdateCvmOSReqData ...
type UpdateCvmOSReqData struct {
	BkOsName    string `json:"bk_os_name"`
	BkOsVersion string `json:"bk_os_version"`
	SrvStatus   string `json:"srv_status"`
}

// Validate ...
func (u *UpdateCvmOSReq) Validate() error {
	if len(u.BkAssetId) == 0 {
		return fmt.Errorf("bk_asset_id is required")
	}
	return nil
}

// AddHostReq add host to cc request
type AddHostReq struct {
	// to be added hosts' asset id list, max length is 10
	// SvrIDs 要转移的公司cmdb主机ID数组
	SvrIDs []int64 `json:"svr_ids" validate:"omitempty,max=10"`
	// AssetIDs 要新增的公司cmdb固资编号数组
	AssetIDs []string `json:"asset_ids" validate:"omitempty,max=10"`
	// InnerIps 要新增的公司cmdb内网ip数组
	InnerIps []string `json:"inner_ips" validate:"omitempty,max=10"`
}

// Validate validate AddHostReq
func (a *AddHostReq) Validate() error {
	// svr_ids,asset_ids,和inner_ips只能选择其中一个作为接口参数
	sum := 0
	if len(a.SvrIDs) > 0 {
		sum++
	}
	if len(a.AssetIDs) > 0 {
		sum++
	}
	if len(a.InnerIps) > 0 {
		sum++
	}
	if sum > 1 {
		return fmt.Errorf("svr_ids,asset_ids, and inner_ips can only choose one")
	}
	return validator.Validate.Struct(a)
}

// OsType 操作系统类型
type OsType string

const (
	// LinuxOsType 操作系统类型-Linux
	LinuxOsType OsType = "1"
	// WindowsOsType 操作系统类型-Windows
	WindowsOsType OsType = "2"
)

// SvrSourceTypeID 服务器来源类型
type SvrSourceTypeID string

const (
	// SvrSourceTypeIDOwn 服务器来源类型ID-自有, 物理机
	SvrSourceTypeIDOwn SvrSourceTypeID = "1"
	// SvrSourceTypeIDHosting 服务器来源类型ID-托管, 物理机
	SvrSourceTypeIDHosting SvrSourceTypeID = "2"
	// SvrSourceTypeIDRent 服务器来源类型ID-租用, 物理机
	SvrSourceTypeIDRent SvrSourceTypeID = "3"
	// SvrSourceTypeIDCVM 服务器来源类型ID-虚拟机
	SvrSourceTypeIDCVM SvrSourceTypeID = "4"
	// SvrSourceTypeIDContainer 服务器来源类型ID-容器
	SvrSourceTypeIDContainer SvrSourceTypeID = "5"
)

// IsPhysicalMachine 检查是否物理机
func IsPhysicalMachine(svrSourceTypeID SvrSourceTypeID) bool {
	// 服务器来源类型ID(未知(0, 默认值) 自有(1) 托管(2) 租用(3) 虚拟机(4) 容器(5))
	switch svrSourceTypeID {
	case SvrSourceTypeIDOwn, SvrSourceTypeIDHosting, SvrSourceTypeIDRent:
		return true
	default:
		return false
	}
}

// SearchBizCompanyCmdbInfoParams is search cmdb business belonging parameter.
type SearchBizCompanyCmdbInfoParams struct {
	BizIDs   []int64   `json:"bk_biz_ids,omitempty" validate:"omitempty,max=20"`
	BizNames []string  `json:"bk_biz_names,omitempty" validate:"omitempty,max=20"`
	Page     *BasePage `json:"page,omitempty" validate:"omitempty"`
}

// Validate validate SearchBizCompanyCmdbInfoParams
func (p *SearchBizCompanyCmdbInfoParams) Validate() error {
	if len(p.BizIDs) == 0 && len(p.BizNames) == 0 && p.Page == nil {
		return fmt.Errorf("page is required, when biz_ids and biz_names are empty")
	}
	return validator.Validate.Struct(p)
}

// CompanyCmdbInfo is search cmdb business belonging element of result.
type CompanyCmdbInfo struct {
	BkBizID          int64  `json:"bk_biz_id"`
	BizName          string `json:"bk_biz_name"`
	BkProductID      int64  `json:"bsi_product_id"`
	BkProductName    string `json:"bsi_product_name"`
	PlanProductID    int64  `json:"plan_product_id"`
	PlanProductName  string `json:"plan_product_name"`
	BusinessDeptID   int64  `json:"business_dept_id"`
	BusinessDeptName string `json:"business_dept_name"`
	Bs1Name          string `json:"bs1_name"`
	Bs1NameID        int64  `json:"bs1_name_id"`
	Bs2Name          string `json:"bs2_name"`
	Bs2NameID        int64  `json:"bs2_name_id"`
	VirtualDeptID    int64  `json:"virtual_dept_id"`
	VirtualDeptName  string `json:"virtual_dept_name"`
}

// ListHostReq list host request
type ListHostReq struct {
	HostPropertyFilter *QueryFilter `json:"host_property_filter"`
	Fields             []string     `json:"fields"`
	Page               BasePage     `json:"page" validate:"required"`
}

// Validate validate ListHostReq
func (req *ListHostReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListHostResult ...
type ListHostResult struct {
	Count int64  `json:"count"`
	Info  []Host `json:"info"`
}

// DeviceTopoInfo topo info
type DeviceTopoInfo struct {
	InnerIP      string `json:"innerIP"`
	AssetID      string `json:"assetID"`
	DeviceClass  string `json:"deviceClass"`
	Raid         string `json:"raid"`
	OSName       string `json:"osName"`
	OSVersion    string `json:"osVersion"`
	IdcArea      string `json:"idcArea"`
	SZone        string `json:"sZone"`
	ModuleName   string `json:"moduleName"`
	Equipment    string `json:"equipment"`
	IdcLogicArea string `json:"idcLogicArea"`
}

// GetUniqIp get CC host unique inner ip
func (h *Host) GetUniqIp() string {
	// when CC host has multiple inner ips, bk_host_innerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(h.BkHostInnerIP, ",")
	if len(multiIps) == 0 {
		return ""
	}
	return multiIps[0]
}

// CrTransitReq transfer host to CR transit module request
type CrTransitReq struct {
	From CrTransitSrcInfo `json:"bk_from" validate:"required"`
	To   CrTransitDstInfo `json:"bk_to" validate:"required"`
}

// Validate validate CrTransitReq
func (req *CrTransitReq) Validate() error {
	if err := req.From.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(req)
}

// CrTransitSrcInfo transfer host source info
type CrTransitSrcInfo struct {
	FromBizID    int64 `json:"bk_biz_id" validate:"required"`
	FromModuleID int64 `json:"bk_module_id" validate:"required"`
	// AssetIDs 要转移的公司cmdb固资编号数组
	AssetIDs []string `json:"asset_ids" validate:"omitempty,max=100"`
	// InnerIps 要新增的公司cmdb内网ip数组
	InnerIps []string `json:"inner_ips" validate:"omitempty,max=100"`
	// SvrIDs 要转移的公司cmdb主机ID数组
	SvrIDs []int64 `json:"svr_ids" validate:"omitempty,max=100"`
}

// Validate validate CrTransitSrcInfo
func (req *CrTransitSrcInfo) Validate() error {
	// svr_ids,asset_ids,和inner_ips只能选择其中一个作为接口参数
	sum := 0
	if len(req.SvrIDs) > 0 {
		sum++
	}
	if len(req.AssetIDs) > 0 {
		sum++
	}
	if len(req.InnerIps) > 0 {
		sum++
	}
	if sum > 1 {
		return fmt.Errorf("svr_ids,asset_ids, and inner_ips can only choose one")
	}
	return validator.Validate.Struct(req)
}

// CrTransitDstInfo transfer host destination info
type CrTransitDstInfo struct {
	ToBizID int64 `json:"bk_biz_id" validate:"required"`
}

// CrTransitRst transfer host to CR transit module result
type CrTransitRst struct {
	AssetIds []string `json:"asset_ids"`
}

// TransferHostReq transfer host to another business request
type TransferHostReq struct {
	From TransferHostSrcInfo `json:"bk_from" validate:"required"`
	To   TransferHostDstInfo `json:"bk_to" validate:"required"`
}

// Validate ...
func (req *TransferHostReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TransferHostSrcInfo transfer host source info
type TransferHostSrcInfo struct {
	FromBizID int64   `json:"bk_biz_id" validate:"required"`
	HostIDs   []int64 `json:"bk_host_ids" validate:"required,min=1,max=500"`
}

// TransferHostDstInfo transfer host destination info
type TransferHostDstInfo struct {
	ToBizID    int64 `json:"bk_biz_id" validate:"required"`
	ToModuleID int64 `json:"bk_module_id,omitempty"` // 主机要转移到的模块ID，如果不传，则将主机转移到该业务的空闲机模块下
}

// CrTransitIdleReq transfer host from CR transit module back to idle module request
type CrTransitIdleReq struct {
	BkBizId int64 `json:"bk_biz_id" validate:"required"`
	// AssetIDs 要转移的公司cmdb固资编号数组
	AssetIDs []string `json:"asset_ids" validate:"omitempty,max=10"`
	// InnerIps 要新增的公司cmdb内网ip数组
	InnerIps []string `json:"inner_ips" validate:"omitempty,max=10"`
	// SvrIDs 要转移的公司cmdb主机ID数组
	SvrIDs []int64 `json:"svr_ids" validate:"omitempty,max=10"`
}

// Validate validate CrTransitIdleReq
func (req *CrTransitIdleReq) Validate() error {
	// svr_ids,asset_ids,和inner_ips只能选择其中一个作为接口参数
	sum := 0
	if len(req.SvrIDs) > 0 {
		sum++
	}
	if len(req.AssetIDs) > 0 {
		sum++
	}
	if len(req.InnerIps) > 0 {
		sum++
	}
	if sum > 1 {
		return fmt.Errorf("svr_ids,asset_ids, and inner_ips can only choose one")
	}
	return validator.Validate.Struct(req)
}

// IsPmAndOuterIPDevice 检查是否物理机，是否有外网IP
func (h *Host) IsPmAndOuterIPDevice() bool {
	// 服务器来源类型ID(未知(0, 默认值) 自有(1) 托管(2) 租用(3) 虚拟机(4) 容器(5))
	if !IsPhysicalMachine(h.SvrSourceTypeID) {
		return false
	}

	if h.BkHostOuterIP == "" && h.BkHostOuterIPv6 == "" {
		return false
	}

	return true
}

// UpdateHostsReq update hosts request
type UpdateHostsReq struct {
	Update []*UpdateHostProperty `json:"update" validate:"required,min=1,max=500,dive"`
}

// Validate validate UpdateHostsReq
func (req *UpdateHostsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// UpdateHostProperty update hosts property
type UpdateHostProperty struct {
	HostID     int64                  `json:"bk_host_id" validate:"required"`
	Properties map[string]interface{} `json:"properties" validate:"required"`
}

// ModuleHost host module relation result
type ModuleHost struct {
	AppID    int64  `json:"bk_biz_id,omitempty" bson:"bk_biz_id"`
	HostID   int64  `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	ModuleID int64  `json:"bk_module_id,omitempty" bson:"bk_module_id"`
	SetID    int64  `json:"bk_set_id,omitempty" bson:"bk_set_id"`
	OwnerID  string `json:"bk_supplier_account,omitempty" bson:"bk_supplier_account"`
}

// FindManyHostsByAssetIDReq bkcc by_asset_id 请求
type FindManyHostsByAssetIDReq struct {
	BkAssetIDs []string `json:"bk_asset_ids" validate:"required,min=1"`
	Fields     []string `json:"fields,omitempty"`
}

// Validate  FindManyHostsByAssetIDReq
func (req *FindManyHostsByAssetIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HostByAssetID by_asset_id 响应的主机信息
type HostByAssetID struct {
	BkAssetID     string `json:"bk_asset_id"`
	BkHostInnerIP string `json:"bk_host_innerip"`
}

// FindManyHostsByAssetIDResp by_asset_id
type FindManyHostsByAssetIDResp struct {
	types.BaseResponse
	Data []HostByAssetID `json:"data"`
}

// GetBizInternalModuleReq get business's internal module request
type GetBizInternalModuleReq struct {
	BkBizID int64 `json:"bk_biz_id"`
}

// Validate validate GetBizInternalModuleReq
func (p *GetBizInternalModuleReq) Validate() error {
	if p.BkBizID <= 0 {
		return fmt.Errorf("bk_biz_id is required")
	}
	return nil
}

// BizInternalModuleRespRst search business's internal module result
type BizInternalModuleRespRst struct {
	BkSetID   int64         `json:"bk_set_id"`
	BkSetName string        `json:"bk_set_name"`
	Module    []*ModuleInfo `json:"module"`
}

// GetUniqOuterIp get CC host unique outer ip
func (h *Host) GetUniqOuterIp() string {
	// when CC host has multiple outer ips, bk_host_outerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(h.BkHostOuterIP, ",")
	if len(multiIps) == 0 {
		return ""
	}

	return multiIps[0]
}

// ConditionItem cc query condition item
type ConditionItem struct {
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}
