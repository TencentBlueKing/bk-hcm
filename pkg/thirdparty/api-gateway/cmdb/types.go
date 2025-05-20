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
	BizID              int64        `json:"bk_biz_id" validate:"required"`
	BkSetIDs           []int64      `json:"bk_set_ids"`
	BkModuleIDs        []int64      `json:"bk_module_ids"`
	Fields             []string     `json:"fields"`
	Page               *BasePage    `json:"page" validate:"required"`
	HostPropertyFilter *QueryFilter `json:"host_property_filter,omitempty"`
}

// Validate validate ListBizHostParams
func (p *ListBizHostParams) Validate() error {
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
