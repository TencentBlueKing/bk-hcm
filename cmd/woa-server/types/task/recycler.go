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

// Package task defines task types
package task

import (
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/metadata"
)

// RecycleCheckReq resource recycle check request
type RecycleCheckReq struct {
	IPs      []string `json:"ips"`
	AssetIDs []string `json:"asset_ids"`
	HostIDs  []int64  `json:"bk_host_ids"`
}

// Validate whether RecycleCheckReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *RecycleCheckReq) Validate() (errKey string, err error) {
	if len(req.IPs) == 0 && len(req.AssetIDs) == 0 && len(req.HostIDs) == 0 {
		return "ips", fmt.Errorf("ips, asset_ids or bk_host_ids should be set")
	}

	if len(req.IPs) > pkg.BKMaxInstanceLimit {
		return "ips", fmt.Errorf("ips exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.AssetIDs) > pkg.BKMaxInstanceLimit {
		return "asset_ids", fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.HostIDs) > pkg.BKMaxInstanceLimit {
		return "bk_host_ids", fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return "", nil
}

// RecycleCheckRst resource recycle check result
type RecycleCheckRst struct {
	Count int64               `json:"count"`
	Info  []*RecycleCheckInfo `json:"info"`
}

// RecycleCheckInfo resource recycle check info
type RecycleCheckInfo struct {
	HostID           int64  `json:"bk_host_id"`
	AssetID          string `json:"asset_id"`
	IP               string `json:"ip"`
	BkHostOuterIP    string `json:"bk_host_outerip"`
	BizID            int64  `json:"bk_biz_id"`
	BizName          string `json:"bk_biz_name"`
	ModuleDefaultVal int64  `json:"module_default_val"`
	Operator         string `json:"operator"`
	BakOperator      string `json:"bak_operator"`
	DeviceType       string `json:"device_type"`
	State            string `json:"state"`
	InputTime        string `json:"input_time"`
	Recyclable       bool   `json:"recyclable"`
	Message          string `json:"message"`
	TopoModule       string `json:"topo_module"`
}

// ReturnPlan resource return plan specification
type ReturnPlan struct {
	CvmPlan table.RetPlanType `json:"cvm"`
	PmPlan  table.RetPlanType `json:"pm"`
}

// Validate whether ReturnPlan is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *ReturnPlan) Validate() (errKey string, err error) {
	switch param.CvmPlan {
	case table.RetPlanImmediate:
	case table.RetPlanDelay:
	default:
		return "cvm", fmt.Errorf("unknown return plan type %s", param.CvmPlan)
	}

	switch param.PmPlan {
	case table.RetPlanImmediate:
	case table.RetPlanDelay:
	default:
		return "pm", fmt.Errorf("unknown return plan type %s", param.PmPlan)
	}

	return "", nil
}

// PreviewRecycleReq preview recycle order request
type PreviewRecycleReq struct {
	IPs                 []string            `json:"ips"`
	AssetIDs            []string            `json:"asset_ids"`
	HostIDs             []int64             `json:"bk_host_ids"`
	ReturnPlan          *ReturnPlan         `json:"return_plan"`
	SkipConfirm         bool                `json:"skip_confirm"`
	RecycleTypeSequence []table.RecycleType `json:"recycle_type_sequence"`
	Remark              string              `json:"remark" bson:"remark"`
}

// Validate whether PreviewRecycleReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *PreviewRecycleReq) Validate() error {
	if len(req.IPs) == 0 && len(req.AssetIDs) == 0 && len(req.HostIDs) == 0 {
		return fmt.Errorf("ips, asset_ids or bk_host_ids should be set")
	}

	if len(req.IPs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ips exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	remarkLimit := 256
	if len(req.Remark) > remarkLimit {
		return fmt.Errorf("remark exceed size limit %d", remarkLimit)
	}

	// 校验回收顺序参数提供的回收类型是否正确
	for _, recycleT := range req.RecycleTypeSequence {
		if err := recycleT.Validate(); err != nil {
			return err
		}
		if recycleT.IsFixedType() {
			return fmt.Errorf("fixed recycle type: %s is not allowed", recycleT)
		}
	}

	if req.ReturnPlan == nil {
		return fmt.Errorf("return_plan should be set")
	}

	if _, err := req.ReturnPlan.Validate(); err != nil {
		return err
	}

	return nil
}

// PreviewRecycleOrderRst preview recycle order result
type PreviewRecycleOrderRst struct {
	Info []*table.RecycleOrder `json:"info"`
}

// PreviewRecycleOrderCpuRst preview recycle order cpu result
type PreviewRecycleOrderCpuRst struct {
	Info []*RecycleOrderCpuInfo `json:"info"`
}

// RecycleOrderCpuInfo recycle order cpu info
type RecycleOrderCpuInfo struct {
	*table.RecycleOrder `json:",inline"`
	SumCpuCore          int64 `json:"sum_cpu_core"`
}

// AuditRecycleReq audit recycle order request
type AuditRecycleReq struct {
	SuborderID []string `json:"suborder_id"`
	Operator   string   `json:"operator"`
	Approval   bool     `json:"approval"`
	Remark     string   `json:"remark"`
}

// Validate whether AuditRecycleReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *AuditRecycleReq) Validate() (errKey string, err error) {
	if len(req.SuborderID) == 0 {
		return "suborder_id", fmt.Errorf("suborder_id should be set")
	}

	arrayLimit := 20
	if len(req.SuborderID) > arrayLimit {
		return "suborder_id", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(req.Operator) == 0 {
		return "operator", fmt.Errorf("operator should be set")
	}

	remarkLimit := 256
	if len(req.Remark) > remarkLimit {
		return "remark", fmt.Errorf("exceed size limit %d", remarkLimit)
	}

	return "", nil
}

// CreateRecycleReq create recycle order request
type CreateRecycleReq struct {
	IPs                 []string            `json:"ips"`
	AssetIDs            []string            `json:"asset_ids"`
	HostIDs             []int64             `json:"bk_host_ids"`
	ReturnPlan          *ReturnPlan         `json:"return_plan"`
	SkipConfirm         bool                `json:"skip_confirm"`
	RecycleTypeSequence []table.RecycleType `json:"recycle_type_sequence"`
	Remark              string              `json:"remark" bson:"remark"`
}

// Validate whether CreateRecycleReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *CreateRecycleReq) Validate() error {
	if len(req.IPs) == 0 && len(req.AssetIDs) == 0 && len(req.HostIDs) == 0 {
		return fmt.Errorf("ips, asset_ids or bk_host_ids should be set")
	}

	if len(req.IPs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ips exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(req.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	remarkLimit := 256
	if len(req.Remark) > remarkLimit {
		return fmt.Errorf("remark exceed size limit %d", remarkLimit)
	}

	if req.ReturnPlan == nil {
		return fmt.Errorf("return_plan should be set")
	}

	if _, err := req.ReturnPlan.Validate(); err != nil {
		return err
	}

	return nil
}

// ToPreviewParam convert to preview param
func (req *CreateRecycleReq) ToPreviewParam() *PreviewRecycleReq {
	reviewReq := &PreviewRecycleReq{
		IPs:                 req.IPs,
		AssetIDs:            req.AssetIDs,
		HostIDs:             req.HostIDs,
		ReturnPlan:          req.ReturnPlan,
		SkipConfirm:         req.SkipConfirm,
		RecycleTypeSequence: req.RecycleTypeSequence,
		Remark:              req.Remark,
	}
	// 默认顺序：滚服 -> 短租
	if len(reviewReq.RecycleTypeSequence) == 0 {
		reviewReq.RecycleTypeSequence = []table.RecycleType{
			table.RecycleTypeRollServer,
			table.RecycleTypeShortRental,
		}
	}
	return reviewReq
}

// CreateRecycleOrderRst create recycle order result
type CreateRecycleOrderRst struct {
	Info []*table.RecycleOrder `json:"info"`
}

// GetRecycleOrderReq get recycle order request
type GetRecycleOrderReq struct {
	OrderID      []uint64              `json:"order_id"`
	SuborderID   []string              `json:"suborder_id"`
	BizID        []int64               `json:"bk_biz_id"`
	ResourceType []table.ResourceType  `json:"resource_type"`
	RecycleType  []table.RecycleType   `json:"recycle_type"`
	ReturnPlan   []table.RetPlanType   `json:"return_plan"`
	Stage        []table.RecycleStage  `json:"stage"`
	Status       []table.RecycleStatus `json:"status"`
	User         []string              `json:"bk_username"`
	Start        string                `json:"start"`
	End          string                `json:"end"`
	Page         metadata.BasePage     `json:"page"`
}

// Validate whether GetRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecycleOrderReq) Validate() error {
	arrayLimit := 20
	if len(param.OrderID) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	if len(param.BizID) <= 0 {
		return fmt.Errorf("bk_biz_id is required")
	}

	if len(param.ResourceType) > arrayLimit {
		return fmt.Errorf("resource_type exceed limit %d", arrayLimit)
	}

	if len(param.RecycleType) > arrayLimit {
		return fmt.Errorf("recycle_type exceed limit %d", arrayLimit)
	}

	if len(param.Stage) > arrayLimit {
		return fmt.Errorf("stage exceed limit %d", arrayLimit)
	}

	if len(param.Status) > arrayLimit {
		return fmt.Errorf("status exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return fmt.Errorf("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return fmt.Errorf("invalid page.limit < 0")
	}

	if param.Page.Limit > 200 {
		return fmt.Errorf("exceed page.limit 200")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetRecycleOrderReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		}
	}

	if len(param.BizID) > 0 {
		filter["bk_biz_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.BizID,
		}
	}

	if len(param.ResourceType) > 0 {
		filter["resource_type"] = mapstr.MapStr{
			pkg.BKDBIN: param.ResourceType,
		}
	}

	if len(param.RecycleType) > 0 {
		filter["recycle_type"] = mapstr.MapStr{
			pkg.BKDBIN: param.RecycleType,
		}
	}

	if len(param.ReturnPlan) > 0 {
		filter["return_plan"] = mapstr.MapStr{
			pkg.BKDBIN: param.ReturnPlan,
		}
	}

	if len(param.Stage) == 0 {
		filter["stage"] = mapstr.MapStr{
			pkg.BKDBNE: table.RecycleStageCommit,
		}
	} else {
		stages := make([]table.RecycleStage, 0)
		for _, stage := range param.Stage {
			if stage == table.RecycleStageCommit {
				continue
			}
			stages = append(stages, stage)
		}
		filter["stage"] = mapstr.MapStr{
			pkg.BKDBIN: stages,
		}
	}

	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			pkg.BKDBIN: param.Status,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetRecycleOrderRst get recycle order result
type GetRecycleOrderRst struct {
	Count int64                 `json:"count"`
	Info  []*table.RecycleOrder `json:"info"`
}

// GetBizRecycleReq get business recycle order request
type GetBizRecycleReq struct {
	BkBizID int64             `json:"bk_biz_id" bson:"bk_biz_id"`
	Start   string            `json:"start" bson:"start"`
	End     string            `json:"end" bson:"end"`
	Page    metadata.BasePage `json:"page" bson:"page"`
}

// Validate whether GetBizRecycleReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetBizRecycleReq) Validate() (errKey string, err error) {
	if param.BkBizID <= 0 {
		return "bk_biz_id", errors.New("invalid bk_biz_id <= 0")
	}

	if param.Start != "" {
		if _, err := time.Parse(dateLayout, param.Start); err != nil {
			return "start", fmt.Errorf("start should be in format like \"%s\"", dateLayout)
		}
	}

	if param.End != "" {
		if _, err := time.Parse(dateLayout, param.End); err != nil {
			return "end", fmt.Errorf("end should be in format like \"%s\"", dateLayout)
		}
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", fmt.Errorf("invalid start < 0")
	}

	if param.Page.Limit <= 0 {
		return "page.limit", fmt.Errorf("invalid limit <= 0")
	}

	if param.Page.Limit > 100 {
		return "page.limit", fmt.Errorf("exceed limit 100")
	}

	return "", nil
}

// GetRecycleDetectReq get recycle detection task request
type GetRecycleDetectReq struct {
	OrderID    []uint64             `json:"order_id"`
	SuborderID []string             `json:"suborder_id"`
	BizID      []int64              `json:"bk_biz_id"`
	IP         []string             `json:"ip"`
	Status     []table.DetectStatus `json:"status"`
	User       []string             `json:"bk_username"`
	Start      string               `json:"start"`
	End        string               `json:"end"`
	Page       metadata.BasePage    `json:"page"`
}

// Validate whether GetRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecycleDetectReq) Validate() error {
	arrayLimit := 20
	if len(param.OrderID) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	if len(param.BizID) > arrayLimit {
		return fmt.Errorf("bk_biz_id exceed limit %d", arrayLimit)
	}

	if len(param.IP) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ip exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(param.Status) > arrayLimit {
		return fmt.Errorf("status exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return fmt.Errorf("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return fmt.Errorf("invalid page.limit < 0")
	}

	if param.Page.Limit > 500 {
		return fmt.Errorf("exceed page.limit 500")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetRecycleDetectReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		}
	}

	if len(param.BizID) > 0 {
		filter["bk_biz_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.BizID,
		}
	}

	if len(param.IP) > 0 {
		filter["ip"] = mapstr.MapStr{
			pkg.BKDBIN: param.IP,
		}
	}

	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			pkg.BKDBIN: param.Status,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetDetectTaskRst get recycle detection task result
type GetDetectTaskRst struct {
	Count int64               `json:"count"`
	Info  []*table.DetectTask `json:"info"`
}

// ListDetectHostRst get recycle detection host list result
type ListDetectHostRst struct {
	Info []interface{} `json:"info"`
}

// GetDetectStepReq get recycle detection step request
type GetDetectStepReq struct {
	OrderID    []uint64             `json:"order_id"`
	SuborderID []string             `json:"suborder_id"`
	BizID      []int64              `json:"bk_biz_id"`
	IP         []string             `json:"ip"`
	StepName   []string             `json:"step_name"`
	Status     []table.DetectStatus `json:"status"`
	User       []string             `json:"bk_username"`
	Start      string               `json:"start"`
	End        string               `json:"end"`
	TaskID     []string             `json:"task_id"`
	Page       metadata.BasePage    `json:"page"`
}

// Validate whether GetDetectStepReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetDetectStepReq) Validate() error {
	arrayLimit := 20
	if len(param.OrderID) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	if len(param.BizID) > arrayLimit {
		return fmt.Errorf("bk_biz_id exceed limit %d", arrayLimit)
	}

	if len(param.IP) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ip exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(param.StepName) > arrayLimit {
		return fmt.Errorf("step_name exceed limit %d", arrayLimit)
	}

	if len(param.Status) > arrayLimit {
		return fmt.Errorf("status exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return fmt.Errorf("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return fmt.Errorf("invalid page.limit < 0")
	}

	if param.Page.Limit > 200 {
		return fmt.Errorf("exceed page.limit 200")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetDetectStepReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		}
	}

	if len(param.BizID) > 0 {
		filter["bk_biz_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.BizID,
		}
	}

	if len(param.IP) > 0 {
		filter["ip"] = mapstr.MapStr{
			pkg.BKDBIN: param.IP,
		}
	}

	if len(param.StepName) > 0 {
		filter["step_name"] = mapstr.MapStr{
			pkg.BKDBIN: param.StepName,
		}
	}

	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			pkg.BKDBIN: param.Status,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	if len(param.TaskID) > 0 {
		filter["task_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.TaskID,
		}
	}

	// 防止前端不传任何参数的情况
	if len(filter) == 0 {
		return nil, errors.New("no filter condition provided")
	}

	return filter, nil
}

// GetDetectStepRst get recycle detection step result
type GetDetectStepRst struct {
	Count int64               `json:"count"`
	Info  []*table.DetectStep `json:"info"`
}

// GetRecycleHostReq get recycle host info request
type GetRecycleHostReq struct {
	OrderID    []uint64              `json:"order_id"`
	SuborderID []string              `json:"suborder_id"`
	BizID      []int64               `json:"bk_biz_id"`
	BkAssetID  []string              `json:"bk_asset_id"`
	DeviceType []string              `json:"device_type"`
	Zone       []string              `json:"bk_zone_name"`
	SubZone    []string              `json:"sub_zone"`
	Stage      []table.RecycleStage  `json:"stage"`
	Status     []table.RecycleStatus `json:"status"`
	User       []string              `json:"bk_username"`
	IP         []string              `json:"ip"`
	Start      string                `json:"start"`
	End        string                `json:"end"`
	Page       metadata.BasePage     `json:"page"`
}

// Validate whether GetRecycleHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecycleHostReq) Validate() error {
	arrayLimit := 20
	if len(param.OrderID) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	if len(param.BizID) == 0 {
		return errors.New("bk_biz_id is required")
	}

	if len(param.BkAssetID) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_asset_id exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(param.DeviceType) > arrayLimit {
		return fmt.Errorf("device_type exceed limit %d", arrayLimit)
	}

	if len(param.Zone) > arrayLimit {
		return fmt.Errorf("bk_zone_name exceed limit %d", arrayLimit)
	}

	if len(param.SubZone) > arrayLimit {
		return fmt.Errorf("sub_zone exceed limit %d", arrayLimit)
	}

	if len(param.Stage) > arrayLimit {
		return fmt.Errorf("stage exceed limit %d", arrayLimit)
	}

	if len(param.Status) > arrayLimit {
		return fmt.Errorf("status exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.IP) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ip exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if param.Page.EnableCount {
		if param.Page.Start > 0 || param.Page.Limit > 0 || param.Page.Sort != "" {
			return fmt.Errorf("params page can not be set")
		}
		return nil
	}

	if param.Page.Start < 0 {
		return fmt.Errorf("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return fmt.Errorf("invalid page.limit < 0")
	}

	if param.Page.Limit > 5000 {
		return fmt.Errorf("exceed page.limit 5000")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetRecycleHostReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		}
	}

	if len(param.BizID) > 0 {
		filter["bk_biz_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.BizID,
		}
	}

	if len(param.BkAssetID) > 0 {
		filter["asset_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.BkAssetID,
		}
	}

	if len(param.DeviceType) > 0 {
		filter["device_type"] = mapstr.MapStr{
			pkg.BKDBIN: param.DeviceType,
		}
	}

	if len(param.Zone) > 0 {
		filter["bk_zone_name"] = mapstr.MapStr{
			pkg.BKDBIN: param.Zone,
		}
	}

	if len(param.SubZone) > 0 {
		filter["sub_zone"] = mapstr.MapStr{
			pkg.BKDBIN: param.SubZone,
		}
	}

	if len(param.Stage) > 0 {
		filter["stage"] = mapstr.MapStr{
			pkg.BKDBIN: param.Stage,
		}
	}

	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			pkg.BKDBIN: param.Status,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	if len(param.IP) > 0 {
		filter["ip"] = mapstr.MapStr{
			pkg.BKDBIN: param.IP,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetRecycleHostRst get recycle host info result
type GetRecycleHostRst struct {
	Count int64                `json:"count"`
	Info  []*table.RecycleHost `json:"info"`
}

// GetDetectStepCfgRst get recycle detection step config result
type GetDetectStepCfgRst struct {
	Info []*table.DetectStepCfg `json:"info"`
}

// StartRecycleOrderReq start recycle order request
type StartRecycleOrderReq struct {
	OrderID               []uint64                            `json:"order_id"`
	SuborderID            []string                            `json:"suborder_id"`
	ReturnForecastConfigs []*RecycleOrderReturnForecastConfig `json:"return_forecast_configs"`
}

// Validate whether StartRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *StartRecycleOrderReq) Validate() error {
	if len(param.OrderID) == 0 && len(param.SuborderID) == 0 {
		return fmt.Errorf("order_id or suborder_id should be set")
	}

	if len(param.OrderID) > 0 && len(param.SuborderID) > 0 {
		return fmt.Errorf("order_id and suborder_id cannot be set as input at the same time")
	}

	if len(param.OrderID) > 0 {
		for _, orderID := range param.OrderID {
			if orderID == 0 {
				return fmt.Errorf("order_id should not be empty")
			}
		}
	}

	if len(param.SuborderID) > 0 {
		for _, subOrderID := range param.SuborderID {
			if len(subOrderID) == 0 {
				return fmt.Errorf("suborder_id should not be empty")
			}
		}
	}

	arrayLimit := 20
	if len(param.OrderID) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	return nil
}

// RecycleOrderReturnForecastConfig update recycle order return forecast config
type RecycleOrderReturnForecastConfig struct {
	SuborderID         string `json:"suborder_id"`
	ReturnForecast     bool   `json:"return_forecast"`
	ReturnForecastTime string `json:"return_forecast_time"`
}

// TerminateRecycleOrderReq terminate recycle order request
type TerminateRecycleOrderReq struct {
	SuborderID []string `json:"suborder_id"`
	Force      bool     `json:"force"`
}

// Validate whether TerminateRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *TerminateRecycleOrderReq) Validate() error {
	if len(param.SuborderID) == 0 {
		return fmt.Errorf("suborder_id should be set")
	}

	for _, subOrderID := range param.SuborderID {
		if len(subOrderID) == 0 {
			return fmt.Errorf("suborder_id should not be empty")
		}
	}

	arrayLimit := 20

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	return nil
}

// ResumeRecycleOrderReq resume recycle order request
type ResumeRecycleOrderReq struct {
	SuborderID []string `json:"suborder_id"`
}

// Validate whether ResumeRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *ResumeRecycleOrderReq) Validate() error {
	if len(param.SuborderID) == 0 {
		return fmt.Errorf("suborder_id should be set")
	}

	for _, subOrderID := range param.SuborderID {
		if len(subOrderID) == 0 {
			return fmt.Errorf("suborder_id should not be empty")
		}
	}

	arrayLimit := 20

	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	return nil
}

// StartDetectTaskReq start recycle detection task request
type StartDetectTaskReq struct {
	SuborderID []string `json:"suborder_id"`
}

// Validate whether StartDetectTaskReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *StartDetectTaskReq) Validate() error {
	if len(param.SuborderID) <= 0 {
		return fmt.Errorf("suborder_id empty or not set")
	}

	for _, subOrderID := range param.SuborderID {
		if len(subOrderID) == 0 {
			return fmt.Errorf("suborder_id should not be empty")
		}
	}

	arrayLimit := 20
	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	return nil
}

// ReviseRecycleOrderReq revise recycle order request
type ReviseRecycleOrderReq struct {
	SuborderID []string `json:"suborder_id"`
}

// Validate whether ReviseRecycleOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *ReviseRecycleOrderReq) Validate() error {
	if len(param.SuborderID) <= 0 {
		return fmt.Errorf("suborder_id empty or not set")
	}

	for _, subOrderID := range param.SuborderID {
		if len(subOrderID) == 0 {
			return fmt.Errorf("suborder_id should not be empty")
		}
	}

	arrayLimit := 20
	if len(param.SuborderID) > arrayLimit {
		return fmt.Errorf("suborder_id exceed limit %d", arrayLimit)
	}

	return nil
}

// GetRecycleRecordDevTypeRst get recycle record device type list result
type GetRecycleRecordDevTypeRst struct {
	Info []interface{} `json:"info"`
}

// GetRecycleRecordRegionRst get recycle record region list result
type GetRecycleRecordRegionRst struct {
	Info []interface{} `json:"info"`
}

// GetRecycleRecordZoneRst get recycle record zone list result
type GetRecycleRecordZoneRst struct {
	Info []interface{} `json:"info"`
}

// GetRecycleBizHostReq get business hosts in recycle module request
type GetRecycleBizHostReq struct {
	BizID int64             `json:"bk_biz_id"`
	Page  metadata.BasePage `json:"page" bson:"page"`
}

// Validate whether GetRecycleBizHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecycleBizHostReq) Validate() error {
	if param.BizID <= 0 {
		return fmt.Errorf("invalid bk_biz_id %d <= 0", param.BizID)
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}
	if param.Page.Start < 0 {
		return fmt.Errorf("invalid page.start < 0")
	}
	if param.Page.Limit < 0 {
		return fmt.Errorf("invalid page.limit < 0")
	}
	if param.Page.Limit > 500 {
		return fmt.Errorf("exceed page.limit 500")
	}

	return nil
}

// GetRecycleBizHostRst get business hosts in recycle module result
type GetRecycleBizHostRst struct {
	Count int64             `json:"count"`
	Info  []*RecycleBizHost `json:"info"`
}

// RecycleBizHost business host info in recycle module
type RecycleBizHost struct {
	HostID      int64  `json:"bk_host_id"`
	AssetID     string `json:"asset_id"`
	IP          string `json:"ip"`
	Operator    string `json:"operator"`
	BakOperator string `json:"bak_operator"`
	DeviceType  string `json:"device_type"`
	SubZone     string `json:"sub_zone"`
	State       string `json:"state"`
	InputTime   string `json:"input_time"`
}

// StartRecycleOrderByRecycleTypeReq start recycle order by recycle type request
type StartRecycleOrderByRecycleTypeReq struct {
	SubOrderIDTypes []StartRecycleOrderByRecycleTypeItem `json:"suborder_id_types" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *StartRecycleOrderByRecycleTypeReq) Validate() error {
	if len(r.SubOrderIDTypes) == 0 {
		return fmt.Errorf("suborder_id_types is required")
	}
	if len(r.SubOrderIDTypes) > 100 {
		return fmt.Errorf("suborder_id_types length should <= 100")
	}
	for _, item := range r.SubOrderIDTypes {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(r)
}

// StartRecycleOrderByRecycleTypeItem start recycle order by recycle type item
type StartRecycleOrderByRecycleTypeItem struct {
	SuborderID         string            `json:"suborder_id"`
	RecycleType        table.RecycleType `json:"recycle_type"`
	ReturnForecast     bool              `json:"return_forecast"`
	ReturnForecastTime string            `json:"return_forecast_time"`
}

// Validate validate
func (r *StartRecycleOrderByRecycleTypeItem) Validate() error {
	if len(r.SuborderID) == 0 {
		return fmt.Errorf("suborder_id is required")
	}

	if len(r.RecycleType) == 0 {
		return fmt.Errorf("recycle_type is required")
	}

	if err := r.RecycleType.Validate(); err != nil {
		return err
	}

	return nil
}

// CheckHostUworkTicketReq check host uwork ticket request
type CheckHostUworkTicketReq struct {
	BkHostIDs []int64 `json:"bk_host_ids" validate:"required,min=1"`
}

// Validate CheckHostUworkTicketReq validate
func (r CheckHostUworkTicketReq) Validate() error {
	if len(r.BkHostIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch check max limit is %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(r)
}

// CheckHostUworkTicketResp check host uwork ticket resp
type CheckHostUworkTicketResp struct {
	Details []CheckHostUworkTicketItem `json:"details"`
}

// CheckHostUworkTicketItem check host uwork ticket item
type CheckHostUworkTicketItem struct {
	BkHostID       int64    `json:"bk_host_id"`
	HasOpenTickets bool     `json:"has_open_tickets"`
	OpenTicketIDs  []string `json:"open_ticket_ids"`
}
