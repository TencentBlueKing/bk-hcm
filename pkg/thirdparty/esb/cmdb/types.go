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
	"hcm/pkg/thirdparty/esb/types"
)

// ----------------------------- biz -----------------------------

// SearchBizParams is esb search cmdb business parameter.
type esbSearchBizParams struct {
	*types.CommParams
	*SearchBizParams
}

// SearchBizParams is cmdb search business parameter.
type SearchBizParams struct {
	Fields            []string     `json:"fields"`
	Page              BasePage     `json:"page"`
	BizPropertyFilter *QueryFilter `json:"biz_property_filter,omitempty"`
}

const BizIDField = "bk_biz_id"

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
type Condition string

const (
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
type Operator string

var (
	OperatorEqual = Operator("equal")
	OperatorIn    = Operator("in")
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

// SearchCloudAreaParams is esb search cmdb cloud area parameter.
type esbSearchCloudAreaParams struct {
	*types.CommParams
	*SearchCloudAreaParams
}

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

// esbAddCloudHostToBizParams is esb add cmdb cloud host to biz parameter.
type esbAddCloudHostToBizParams struct {
	*types.CommParams
	*AddCloudHostToBizParams
}

// AddCloudHostToBizParams is esb add cloud host to biz parameter.
type AddCloudHostToBizParams struct {
	BizID    int64  `json:"bk_biz_id"`
	HostInfo []Host `json:"host_info"`
}

// esbDeleteCloudHostFromBizParams is esb delete cmdb cloud host from biz parameter.
type esbDeleteCloudHostFromBizParams struct {
	*types.CommParams
	*DeleteCloudHostFromBizParams
}

// DeleteCloudHostFromBizParams is esb delete cloud host from biz parameter.
type DeleteCloudHostFromBizParams struct {
	BizID   int64   `json:"bk_biz_id"`
	HostIDs []int64 `json:"bk_host_ids"`
}

// esbListBizHostParams is esb list cmdb host in biz parameter.
type esbListBizHostParams struct {
	*types.CommParams
	*ListBizHostParams
}

// ListBizHostParams is esb list cmdb host in biz parameter.
type ListBizHostParams struct {
	BizID              int64        `json:"bk_biz_id"`
	BkSetIDs           []int64      `json:"bk_set_ids"`
	BkModuleIDs        []int64      `json:"bk_module_ids"`
	Fields             []string     `json:"fields"`
	Page               BasePage     `json:"page"`
	HostPropertyFilter *QueryFilter `json:"host_property_filter,omitempty"`
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
	BkCloudVendor     CloudVendor     `json:"bk_cloud_vendor"`
	BkCloudInstID     string          `json:"bk_cloud_inst_id"`
	BkCloudHostStatus CloudHostStatus `json:"bk_cloud_host_status"`
	BkCloudID         int64           `json:"bk_cloud_id"`
	BkHostInnerIP     string          `json:"bk_host_innerip"`
	BkHostOuterIP     string          `json:"bk_host_outerip"`
	BkHostInnerIPv6   string          `json:"bk_host_innerip_v6"`
	BkHostOuterIPv6   string          `json:"bk_host_outerip_v6"`
	Operator          string          `json:"operator"`
	BkBakOperator     string          `json:"bk_bak_operator"`
	BkHostName        string          `json:"bk_host_name"`
	BkComment         *string         `json:"bk_comment,omitempty"`
}

// CloudVendor defines cmdb cloud vendor type.
type CloudVendor string

const (
	AwsCloudVendor    CloudVendor = "1"
	TCloudCloudVendor CloudVendor = "2"
	GcpCloudVendor    CloudVendor = "3"
	AzureCloudVendor  CloudVendor = "4"
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

// CloudHostStatus defines cmdb cloud host status type.
type CloudHostStatus string

const (
	UnknownCloudHostStatus    CloudHostStatus = "1"
	StartingCloudHostStatus   CloudHostStatus = "2"
	RunningCloudHostStatus    CloudHostStatus = "3"
	StoppingCloudHostStatus   CloudHostStatus = "4"
	StoppedCloudHostStatus    CloudHostStatus = "5"
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
