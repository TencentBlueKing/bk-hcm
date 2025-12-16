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

package dissolve

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	hostdefine "hcm/pkg/dal/table/dissolve/host"
	moduledefine "hcm/pkg/dal/table/dissolve/module"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/es"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
)

// -------------------------- Create --------------------------

// RecycleModuleCreateReq define recycle module create request.
type RecycleModuleCreateReq struct {
	Modules []moduledefine.RecycleModuleTable `json:"modules" validate:"required"`
}

// Validate recycle module create request.
func (req *RecycleModuleCreateReq) Validate() error {
	if len(req.Modules) == 0 {
		return errors.New("modules is required")
	}

	if len(req.Modules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("recycle module count should <= %d, but got: %d", constant.BatchOperationMaxLimit,
			len(req.Modules))
	}

	return nil
}

// RecycleModuleCreateResp define recycle module create response.
type RecycleModuleCreateResp struct {
	IDs []string `json:"ids"`
}

// RecycleHostCreateReq define recycle host create request.
type RecycleHostCreateReq struct {
	Hosts []hostdefine.RecycleHostTable `json:"hosts" validate:"required"`
}

// Validate recycle host create request.
func (req *RecycleHostCreateReq) Validate() error {
	if len(req.Hosts) == 0 {
		return errors.New("hosts is required")
	}

	if len(req.Hosts) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("recycle host count should <= %d, but got: %d", constant.BatchOperationMaxLimit,
			len(req.Hosts))
	}

	return nil
}

// RecycleHostCreateResp define recycle host create response.
type RecycleHostCreateResp struct {
	IDs []string `json:"ids"`
}

// -------------------------- Update --------------------------

// RecycleModuleUpdateReq define recycle module update request.
type RecycleModuleUpdateReq struct {
	moduledefine.RecycleModuleTable `json:",inline"`
}

// Validate recycle module update request.
func (req *RecycleModuleUpdateReq) Validate() error {
	if len(req.ID) == 0 {
		return errors.New("id is required")
	}

	return nil
}

// RecycleHostUpdateReq define recycle host update request.
type RecycleHostUpdateReq struct {
	hostdefine.RecycleHostTable `json:",inline"`
}

// Validate recycle host update request.
func (req *RecycleHostUpdateReq) Validate() error {
	if len(req.ID) == 0 {
		return errors.New("id is required")
	}

	return nil
}

// -------------------------- List --------------------------

// RecycleModuleListReq recycle module list req.
type RecycleModuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate recycle module list request.
func (req *RecycleModuleListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             core.DefaultMaxPageLimit,
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return err
	}

	return nil
}

// RecycleHostListReq recycle host list req.
type RecycleHostListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate recycle host list request.
func (req *RecycleHostListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             core.DefaultMaxPageLimit,
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return err
	}

	return nil
}

// HostListReq host list request.
type HostListReq struct {
	ResDissolveReq `json:",inline"`
	Page           *core.BasePage `json:"page" validate:"required"`
}

// Validate host list request.
func (req *HostListReq) Validate() error {
	if err := req.ResDissolveReq.Validate(); err != nil {
		return err
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             50000, // 由于前端需要导出数据，这里特殊调整限制值
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return err
	}

	return nil
}

// ResDissolveReq resource dissolve request.
type ResDissolveReq struct {
	GroupIDs    []int64  `json:"group_ids"`
	BizNames    []string `json:"bk_biz_names"`
	ModuleNames []string `json:"module_names"`
	Operators   []string `json:"operators"`
}

// Validate table list request.
func (req *ResDissolveReq) Validate() error {
	if len(req.ModuleNames) == 0 {
		return errf.Newf(errf.InvalidParameter, "module_names is required")
	}

	return nil
}

// GetESCond get elasticsearch condition
func (req *ResDissolveReq) GetESCond(assetIDs []string, bizIDName, blackBizIDName map[int64]string) (
	map[string][]interface{}, error) {

	cond := make(map[string][]interface{})

	for _, id := range req.GroupIDs {
		cond[es.GroupID] = append(cond[es.GroupID], id)
	}

	for id := range bizIDName {
		cond[es.BizID] = append(cond[es.BizID], id)
	}

	for id := range blackBizIDName {
		cond[es.BlackList] = append(cond[es.BlackList], id)
	}

	for _, v := range req.Operators {
		cond[es.Operator] = append(cond[es.Operator], v)
	}

	for _, assetID := range assetIDs {
		cond[es.AssetID] = append(cond[es.AssetID], assetID)
	}

	return cond, nil
}

// GetCCHostCond get cc host condition
func (req *ResDissolveReq) GetCCHostCond(assetIDs []string) []*cmdb.QueryFilter {
	andRules := make([]querybuilder.Rule, 0)
	cloudIDRule := querybuilder.AtomRule{
		Field:    pkg.BKCloudIDField,
		Operator: querybuilder.OperatorEqual,
		Value:    0, // 只需要查询管控区域为0的公司的机器
	}
	andRules = append(andRules, cloudIDRule)

	if len(req.Operators) != 0 {
		operatorRule := querybuilder.CombinedRule{
			Condition: querybuilder.ConditionOr,
			Rules: []querybuilder.Rule{
				querybuilder.AtomRule{
					Field: pkg.BKOperatorField, Operator: querybuilder.OperatorIn, Value: req.Operators,
				},
				querybuilder.AtomRule{
					Field: pkg.BKBakOperatorField, Operator: querybuilder.OperatorIn, Value: req.Operators,
				},
			},
		}
		andRules = append(andRules, operatorRule)
	}

	result := make([]*cmdb.QueryFilter, 0)
	batchSize := pkg.BKMaxPageSize
	if len(assetIDs) != 0 {
		for _, batch := range slice.Split(assetIDs, batchSize) {
			assetIDRule := querybuilder.AtomRule{
				Field: pkg.BKAssetIDField, Operator: querybuilder.OperatorIn, Value: batch,
			}
			rules := make([]querybuilder.Rule, len(andRules))
			copy(rules, andRules)
			rules = append(rules, assetIDRule)
			cond := &cmdb.QueryFilter{Rule: querybuilder.CombinedRule{Condition: querybuilder.ConditionAnd,
				Rules: rules}}
			result = append(result, cond)
		}
	}

	return result
}

// ListCurHostCond list current host condition
type ListCurHostCond struct {
	Organizations []string `json:"organizations"`
	BizIDs        []int    `json:"bk_biz_ids"`
	ModuleNames   []string `json:"module_names"`
	Operators     []string `json:"operators"`
}

// ListHostDetails list elasticsearch host details.
type ListHostDetails struct {
	Count   int64  `json:"count,omitempty"`
	Details []Host `json:"details,omitempty"`
}

// Host host data
type Host struct {
	ServerAssetID        string  `json:"server_asset_id"`
	InnerIP              string  `json:"ip"`
	OuterIP              string  `json:"outer_ip"`
	AppName              string  `json:"app_name"`
	BizID                int64   `json:"bk_biz_id"`
	Module               string  `json:"module"`
	DeviceType           string  `json:"device_type"`
	SvrTypeName          string  `json:"svr_type_name"`
	ModuleName           string  `json:"module_name"`
	IdcUnitName          string  `json:"idc_unit_name"`
	SfwNameVersion       string  `json:"sfw_name_version"`
	GoUpDate             string  `json:"go_up_date"`
	RaidName             string  `json:"raid_name"`
	LogicArea            string  `json:"logic_area"`
	ServerBakOperator    string  `json:"server_bak_operator"`
	ServerOperator       string  `json:"server_operator"`
	DeviceLayer          string  `json:"device_layer"`
	CPUScore             float64 `json:"cpu_score"`
	MemScore             float64 `json:"mem_score"`
	InnerNetTrafficScore float64 `json:"inner_net_traffic_score"`
	DiskIoScore          float64 `json:"disk_io_score"`
	DiskUtilScore        float64 `json:"disk_util_score"`
	IsPass               bool    `json:"is_pass"`
	Mem4linux            float64 `json:"mem4linux"`
	InnerNetTraffic      float64 `json:"inner_net_traffic"`
	OuterNetTraffic      float64 `json:"outer_net_traffic"`
	DiskIo               float64 `json:"disk_io"`
	DiskUtil             float64 `json:"disk_util"`
	DiskTotal            float64 `json:"disk_total"`
	MaxCPUCoreAmount     int64   `json:"max_cpu_core_amount"`
	GroupName            string  `json:"group_name"`
	Center               string  `json:"center"`
}

// ConvertHost convert host
func ConvertHost(origin *es.Host) (*Host, error) {
	if origin == nil {
		return nil, nil
	}

	var isPass bool
	var err error
	if origin.IsPass != "" {
		isPass, err = strconv.ParseBool(origin.IsPass)
		if err != nil {
			return nil, err
		}
	}

	return &Host{
		ServerAssetID:        origin.ServerAssetID,
		InnerIP:              origin.InnerIP,
		OuterIP:              origin.OuterIP,
		AppName:              origin.AppName,
		BizID:                origin.BizID,
		Module:               origin.Module,
		DeviceType:           origin.DeviceType,
		ModuleName:           origin.ModuleName,
		IdcUnitName:          origin.IdcUnitName,
		SfwNameVersion:       origin.SfwNameVersion,
		GoUpDate:             origin.GoUpDate,
		RaidName:             origin.RaidName,
		LogicArea:            origin.LogicArea,
		ServerBakOperator:    origin.ServerBakOperator,
		ServerOperator:       origin.ServerOperator,
		DeviceLayer:          origin.DeviceLayer,
		CPUScore:             origin.CPUScore,
		MemScore:             origin.MemScore,
		InnerNetTrafficScore: origin.InnerNetTrafficScore,
		DiskIoScore:          origin.DiskIoScore,
		DiskUtilScore:        origin.DiskUtilScore,
		IsPass:               isPass,
		Mem4linux:            origin.Mem4linux,
		InnerNetTraffic:      origin.InnerNetTraffic,
		OuterNetTraffic:      origin.OuterNetTraffic,
		DiskIo:               origin.DiskIo,
		DiskUtil:             origin.DiskUtil,
		DiskTotal:            origin.DiskTotal,
		MaxCPUCoreAmount:     origin.MaxCPUCoreAmount,
		GroupName:            origin.GroupName,
		Center:               origin.Center,
	}, nil
}

// ResDissolveTable resource dissolve table
type ResDissolveTable struct {
	Items []BizDetail `json:"items"`
}

// BizDetail business detail
type BizDetail struct {
	BizID           interface{}    `json:"bk_biz_id"`
	BizName         string         `json:"bk_biz_name"`
	ModuleHostCount map[string]int `json:"module_host_count"`
	Total           Total          `json:"total"`
	Progress        string         `json:"progress"`
}

// Total statistical data of hosts under business
type Total struct {
	Origin        TotalData `json:"origin"`
	Current       TotalData `json:"current"`
	DeliveredCore int64     `json:"delivered_core"`
}

// TotalData statistical data of host under business
type TotalData struct {
	HostCount interface{} `json:"host_count"`
	CpuCount  int64       `json:"cpu_count"`
}

// -------------------------- Delete --------------------------

// RecycleModuleDeleteReq recycle module delete request.
type RecycleModuleDeleteReq struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Validate recycle module delete request.
func (req *RecycleModuleDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch delete limit is %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// RecycleHostDeleteReq recycle host delete request.
type RecycleHostDeleteReq struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Validate recycle host delete request.
func (req *RecycleHostDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch delete limit is %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HostDissolveStatusCheckReq check host dissolve status request.
type HostDissolveStatusCheckReq struct {
	HostIDs []int64 `json:"bk_host_ids" validate:"required,min=1"`
}

// Validate ...
func (req *HostDissolveStatusCheckReq) Validate() error {
	if len(req.HostIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("check dissolve host status limit is %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HostDissolveStatusCheckResp check host dissolve status response.
type HostDissolveStatusCheckResp struct {
	Info []HostDissolveStatusCheckInfo `json:"info"`
}

// HostDissolveStatusCheckInfo check host dissolve status info.
type HostDissolveStatusCheckInfo struct {
	HostID int64 `json:"bk_host_id"`
	Status bool  `json:"status"`
}

// ListDissolveCpuCoreSummaryReq list dissolve cpu core summary request
type ListDissolveCpuCoreSummaryReq struct {
	BizID int64 `json:"bk_biz_id" validate:"required"`
}

// Validate ...
func (l *ListDissolveCpuCoreSummaryReq) Validate() error {
	return validator.Validate.Struct(l)
}

// CpuCoreSummary dissolve cpu core summary
type CpuCoreSummary struct {
	TotalCore     int64     `json:"total_core"`
	DeliveredCore int64     `json:"delivered_core"`
	HostApplyTime time.Time `json:"host_apply_time"`
}

// Config dissolve config
type Config struct {
	HostApplyTime *time.Time `json:"host_apply_time"`
	ApprovalLimit *float64   `json:"approval_limit"`
}

// UpsertConfigReq upsert config request
type UpsertConfigReq struct {
	HostApplyTime *time.Time `json:"host_apply_time" validate:"omitempty"`
	ApprovalLimit *float64   `json:"approval_limit" validate:"omitempty"`
}

// Validate ...
func (u *UpsertConfigReq) Validate() error {
	if err := validator.Validate.Struct(u); err != nil {
		return err
	}

	if u.ApprovalLimit == nil {
		return nil
	}
	if cvt.PtrToVal(u.ApprovalLimit) < 0 || cvt.PtrToVal(u.ApprovalLimit) > 100 {
		return errors.New("approval_limit must between 0 and 100")
	}

	return nil
}
