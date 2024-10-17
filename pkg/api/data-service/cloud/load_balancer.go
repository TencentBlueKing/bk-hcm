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

package cloud

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// LoadBalancerBatchCreateReq load balancer create req.
type LoadBalancerBatchCreateReq[Extension corelb.Extension] struct {
	Lbs []LbBatchCreate[Extension] `json:"lbs" validate:"required,min=1"`
}

// TCloudCLBCreateReq batch create load balancer
type TCloudCLBCreateReq = LoadBalancerBatchCreateReq[corelb.TCloudClbExtension]

// TCloudCLBCreate create load balancer
type TCloudCLBCreate = LbBatchCreate[corelb.TCloudClbExtension]

// LbBatchCreate define load balancer batch create.
type LbBatchCreate[Extension corelb.Extension] struct {
	CloudID          string               `json:"cloud_id" validate:"required"`
	Name             string               `json:"name" validate:"required"`
	Vendor           enumor.Vendor        `json:"vendor" validate:"required"`
	AccountID        string               `json:"account_id" validate:"required"`
	BkBizID          int64                `json:"bk_biz_id" validate:"omitempty"`
	LoadBalancerType string               `json:"load_balancer_type" validate:"required"`
	IPVersion        enumor.IPAddressType `json:"ip_version" validate:"required"`

	Region               string   `json:"region" validate:"omitempty"`
	Zones                []string `json:"zones" `
	BackupZones          []string `json:"backup_zones"`
	VpcID                string   `json:"vpc_id" validate:"omitempty"`
	CloudVpcID           string   `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID             string   `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID        string   `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
	Domain               string   `json:"domain"`
	Status               string   `json:"status"`
	CloudCreatedTime     string   `json:"cloud_created_time"`
	CloudStatusTime      string   `json:"cloud_status_time"`
	CloudExpiredTime     string   `json:"cloud_expired_time"`

	Memo      *string    `json:"memo"`
	Extension *Extension `json:"extension"`
}

// Validate load balancer create request.
func (req *LoadBalancerBatchCreateReq[T]) Validate() error {
	if len(req.Lbs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("lbs count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// LoadBalancerExtUpdateReq ...
type LoadBalancerExtUpdateReq[T corelb.Extension] struct {
	ID      string `json:"id" validate:"required"`
	Name    string `json:"name,omitempty"`
	BkBizID int64  `json:"bk_biz_id,omitempty"`

	IPVersion enumor.IPAddressType `json:"ip_version"`

	VpcID                string   `json:"vpc_id"`
	CloudVpcID           string   `json:"cloud_vpc_id"`
	SubnetID             string   `json:"subnet_id"`
	CloudSubnetID        string   `json:"cloud_subnet_id"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
	Domain               string   `json:"domain"`
	Status               string   `json:"status"`
	CloudCreatedTime     string   `json:"cloud_created_time"`
	CloudStatusTime      string   `json:"cloud_status_time"`
	CloudExpiredTime     string   `json:"cloud_expired_time"`
	Memo                 *string  `json:"memo"`

	*core.Revision `json:",inline"`
	Extension      *T `json:"extension"`
}

// Validate ...
func (req *LoadBalancerExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// LbExtBatchUpdateReq ...
type LbExtBatchUpdateReq[T corelb.Extension] struct {
	Lbs []*LoadBalancerExtUpdateReq[T] `json:"lbs" validate:"min=1"`
}

// Validate ...
func (req *LbExtBatchUpdateReq[T]) Validate() error {
	if len(req.Lbs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("lbs length count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// TCloudClbBatchUpdateReq ...
type TCloudClbBatchUpdateReq = LbExtBatchUpdateReq[corelb.TCloudClbExtension]

// BizBatchUpdateReq 批量更新业务id
type BizBatchUpdateReq struct {
	IDs     []string `json:"ids" validate:"required"`
	BkBizID int64    `json:"bk_biz_id" validate:"required"`
}

// Validate ...
func (req *BizBatchUpdateReq) Validate() error {

	if len(req.IDs) == 0 {
		return errors.New("ids required")
	}

	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// LbListResult define lb list result.
type LbListResult = core.ListResultT[corelb.BaseLoadBalancer]

// LbExtListResult define lb with extension list result.
type LbExtListResult[T corelb.Extension] struct {
	Count   uint64                   `json:"count,omitempty"`
	Details []corelb.LoadBalancer[T] `json:"details,omitempty"`
}

// -------------------------- ListRaw --------------------------

// LbRawListResult define lb with extension list result.
type LbRawListResult = core.ListResultT[corelb.LoadBalancerRaw]

// -------------------------- Delete --------------------------

// LoadBalancerBatchDeleteReq delete request.
type LoadBalancerBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *LoadBalancerBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List Listener --------------------------

// ListenerListResult define listener list result.
type ListenerListResult = core.ListResultT[corelb.BaseListener]

// TCloudListenerListResult ...
type TCloudListenerListResult = core.ListResultT[corelb.Listener[corelb.TCloudListenerExtension]]

// -------------------------- List Count Listener By LbIDs --------------------------

// ListListenerCountByLbIDsReq define list listener count by lbIDs req.
type ListListenerCountByLbIDsReq struct {
	LbIDs []string `json:"lb_ids" validate:"required,min=1,max=100,dive"`
}

// Validate request.
func (req *ListListenerCountByLbIDsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListListenerCountResp define list listener count resp.
type ListListenerCountResp struct {
	Details []*ListListenerCountResult `json:"details"`
}

// ListListenerCountResult define list listener count result.
type ListListenerCountResult struct {
	LbID string `json:"lb_id" db:"lb_id"`
	Num  int64  `json:"num" db:"num"`
}

// -------------------------- Get Listener --------------------------

// TCloudListenerDetailResult ...
type TCloudListenerDetailResult = corelb.Listener[corelb.TCloudListenerExtension]

// -------------------------- List Target --------------------------

// TargetListResult define target list result.
type TargetListResult = core.ListResultT[corelb.BaseTarget]

// -------------------------- List TCloud Url Rule --------------------------

// TCloudURLRuleListResult define tcloud url rule list result.
type TCloudURLRuleListResult = core.ListResultT[corelb.TCloudLbUrlRule]

// TCloudUrlRuleBatchCreateReq ...
type TCloudUrlRuleBatchCreateReq struct {
	UrlRules []TCloudUrlRuleCreate `json:"url_rules" validate:"required,min=1"`
}

// Validate ...
func (r *TCloudUrlRuleBatchCreateReq) Validate() error {
	if len(r.UrlRules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("url_rules length count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(r)
}

// TCloudUrlRuleCreate tcloud url rule create.
type TCloudUrlRuleCreate struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`

	LbID       string `json:"lb_id" validate:"required,lte=255"`
	CloudLbID  string `json:"cloud_lb_id" validate:"required,lte=255"`
	LblID      string `json:"lbl_id" validate:"required,lte=255"`
	CloudLBLID string `json:"cloud_lbl_id" validate:"required,lte=255"`

	CloudID            string                        `json:"cloud_id" validate:"required,lte=255"`
	Name               string                        `json:"name" validate:"lte=255"`
	RuleType           enumor.RuleType               `json:"rule_type" validate:"required,lte=64"`
	TargetGroupID      string                        `json:"target_group_id" validate:"lte=255"`
	CloudTargetGroupID string                        `json:"cloud_target_group_id" validate:"lte=255"`
	Domain             string                        `json:"domain"`
	URL                string                        `json:"url"`
	Scheduler          string                        `json:"scheduler"`
	SessionType        string                        `json:"session_type"`
	SessionExpire      int64                         `json:"session_expire"`
	HealthCheck        *corelb.TCloudHealthCheckInfo `json:"health_check" validate:"required"`
	Certificate        *corelb.TCloudCertificateInfo `json:"certificate" validate:"required"`
	Memo               *string                       `json:"memo" validate:"lte=255"`
}

// Validate ...
func (req *TCloudUrlRuleCreate) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudUrlRuleBatchUpdateReq 批量更新url规则
type TCloudUrlRuleBatchUpdateReq struct {
	UrlRules []*TCloudUrlRuleUpdate `json:"url_rules" validate:"required,min=1,dive"`
}

// Validate ...
func (r *TCloudUrlRuleBatchUpdateReq) Validate() error {
	if len(r.UrlRules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("url_rules length count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(r)
}

// TCloudUrlRuleUpdate tcloud url rule update.
type TCloudUrlRuleUpdate struct {
	ID string `json:"id" validate:"required,lte=255"`

	Name               string                        `json:"name" validate:"lte=255"`
	TargetGroupID      string                        `json:"target_group_id" validate:"omitempty,lte=255"`
	CloudTargetGroupID string                        `json:"cloud_target_group_id" validate:"omitempty,lte=255"`
	Domain             string                        `json:"domain"`
	URL                string                        `json:"url"`
	Scheduler          string                        `json:"scheduler"`
	SessionType        string                        `json:"session_type"`
	SessionExpire      *int64                        `json:"session_expire"`
	HealthCheck        *corelb.TCloudHealthCheckInfo `json:"health_check" validate:"omitempty"`
	Certificate        *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
	Memo               *string                       `json:"memo" validate:"omitempty,lte=255"`
}

// Validate ...
func (req *TCloudUrlRuleUpdate) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create Res Flow Lock --------------------------

// ResFlowLockCreateReq res flow lock create req.
type ResFlowLockCreateReq struct {
	ResID   string                   `json:"res_id" validate:"required"`
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
	Owner   string                   `json:"owner" validate:"omitempty"`
}

// Validate validate res flow lock create
func (req *ResFlowLockCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Batch Create Res Flow Rel --------------------------

// ResFlowRelBatchCreateReq res flow rel batch create req.
type ResFlowRelBatchCreateReq struct {
	ResFlowRels []ResFlowRelCreateReq `json:"res_flow_rels" validate:"required,min=1"`
}

// Validate validate listener batch create
func (req *ResFlowRelBatchCreateReq) Validate() error {
	for _, item := range req.ResFlowRels {
		if err := item.Validate(); err != nil {
			return errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	return validator.Validate.Struct(req)
}

// ResFlowRelCreateReq res flow rel create req.
type ResFlowRelCreateReq struct {
	ResID    string                   `json:"res_id" validate:"required"`
	ResType  enumor.CloudResourceType `json:"res_type" validate:"required"`
	FlowID   string                   `json:"flow_id" validate:"required"`
	TaskType enumor.TaskType          `json:"task_type" validate:"omitempty"`
	Status   enumor.ResFlowStatus     `json:"status" validate:"omitempty"`
}

// Validate validate rel flow rel create
func (req *ResFlowRelCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update Res Flow Rel --------------------------

// ResFlowRelBatchUpdateReq res flow rel batch update req.
type ResFlowRelBatchUpdateReq struct {
	ResFlowRels []ResFlowRelUpdateReq `json:"res_flow_rels" validate:"required,min=1"`
}

// Validate validate res flow rel batch update
func (req *ResFlowRelBatchUpdateReq) Validate() error {
	for _, item := range req.ResFlowRels {
		if err := item.Validate(); err != nil {
			return errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	return validator.Validate.Struct(req)
}

// ResFlowRelUpdateReq res flow rel update req.
type ResFlowRelUpdateReq struct {
	ID       string                   `json:"id" validate:"required"`
	ResID    string                   `json:"res_id" validate:"required"`
	ResType  enumor.CloudResourceType `json:"res_type" validate:"required"`
	FlowID   string                   `json:"flow_id" validate:"required"`
	TaskType enumor.TaskType          `json:"task_type" validate:"omitempty"`
	Status   enumor.ResFlowStatus     `json:"status" validate:"omitempty"`
}

// Validate validate res flow rel update
func (req *ResFlowRelUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List Res Flow Lock --------------------------

// ResFlowLockListResult define res flow lock list result.
type ResFlowLockListResult = core.ListResultT[corelb.BaseResFlowLock]

// -------------------------- List Res Flow Rel --------------------------

// ResFlowRelListResult define res flow rel list result.
type ResFlowRelListResult = core.ListResultT[corelb.BaseResFlowRel]

// -------------------------- Delete Res Flow Lock --------------------------

// ResFlowLockDeleteReq delete res flow lock request.
type ResFlowLockDeleteReq struct {
	ResID   string                   `json:"res_id" validate:"required"`
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
	Owner   string                   `json:"owner" validate:"required"`
}

// Validate delete request.
func (req *ResFlowLockDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Lock Res Flow --------------------------

// ResFlowLockReq res flow lock req.
type ResFlowLockReq struct {
	ResID    string                   `json:"res_id" validate:"required"`
	ResType  enumor.CloudResourceType `json:"res_type" validate:"required"`
	FlowID   string                   `json:"flow_id" validate:"required"`
	Status   enumor.ResFlowStatus     `json:"status" validate:"required"`
	TaskType enumor.TaskType          `json:"task_type" validate:"omitempty"`
}

// Validate validate res flow lock
func (req *ResFlowLockReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRuleBatchCreateResult ...
type TCloudRuleBatchCreateResult struct {
	Details []TCloudRuleCreateResult `json:"details"`
}

// TCloudRuleCreateResult ...
type TCloudRuleCreateResult struct {
	RuleID string `json:"rule_id"`
	RelID  string `json:"rel_id"`
}

// -------------------------- List Listener With Targets --------------------------

// ListListenerWithTargetsReq define list listener with targets req.
type ListListenerWithTargetsReq struct {
	BkBizID           int64               `json:"bk_biz_id" validate:"required,min=1"`
	Vendor            enumor.Vendor       `json:"vendor" validate:"required,min=1"`
	AccountID         string              `json:"account_id" validate:"required,min=1"`
	ListenerQueryList []ListenerQueryItem `json:"rule_query_list" validate:"required,min=1,max=20"`
	NewRsWeight       int64               `json:"new_rs_weight" validate:"omitempty"`
}

// Validate request.
func (req *ListListenerWithTargetsReq) Validate() error {
	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id is illegal")
	}
	if err := req.Vendor.Validate(); err != nil {
		return err
	}
	if len(req.ListenerQueryList) == 0 {
		return errors.New("rule_query_list is empty")
	}
	for _, item := range req.ListenerQueryList {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(req)
}

// ListenerQueryItem 监听器查询列表
type ListenerQueryItem struct {
	Region        string              `json:"region" validate:"required,min=1"`
	ClbVipDomains []string            `json:"clb_vip_domains" validate:"required,min=1"`
	CloudLbIDs    []string            `json:"cloud_lb_ids" validate:"required,min=1"`
	Protocol      enumor.ProtocolType `json:"protocol" validate:"required,min=1"`
	Ports         []int64             `json:"ports" validate:"omitempty"`
	RuleType      enumor.RuleType     `json:"rule_type" validate:"required,min=1"`
	Domain        string              `json:"domain" validate:"omitempty"`
	Url           string              `json:"url" validate:"omitempty"`
	InstType      enumor.InstType     `json:"inst_type" validate:"required,min=1"`
	RsIPs         []string            `json:"rs_ips" validate:"required,min=1"`
	RsPorts       []int64             `json:"rs_ports" validate:"omitempty"`
	RsWeights     []int64             `json:"rs_weights" validate:"omitempty"`
}

// Validate request.
func (req *ListenerQueryItem) Validate() error {
	if !req.Protocol.IsLayer4Protocol() && !req.Protocol.IsLayer7Protocol() {
		return errors.New("protocol is illegal")
	}
	if err := req.RuleType.Validate(); err != nil {
		return err
	}
	if len(req.ClbVipDomains) != len(req.CloudLbIDs) {
		return errors.New("clb_vip_domains and cloud_lb_ids num must be equal")
	}
	// 传入的负载均衡ID数量不能超过50个
	if len(req.CloudLbIDs) > 50 {
		return errors.New("cloud_lb_ids num must be less than 50")
	}
	// RSPORT填写多个的话，必须和RSIP数量一致
	if len(req.RsPorts) > 0 && len(req.RsPorts) != len(req.RsIPs) {
		return errors.New("rs_port and rs_ip num must be equal")
	}
	// 权重填写多个的话，必须和RSIP数量一致
	if len(req.RsWeights) > 0 && len(req.RsWeights) != len(req.RsIPs) {
		return errors.New("rs_weight and rs_ip num must be equal")
	}

	return validator.Validate.Struct(req)
}

// ListListenerWithTargetsResp define list listener with targets resp.
type ListListenerWithTargetsResp struct {
	Details []*ListBatchListenerResult `json:"details"`
}

// ListBatchListenerResult define list batch listener result.
type ListBatchListenerResult struct {
	ClbID        string                      `json:"clb_id"`
	CloudClbID   string                      `json:"cloud_lb_id"`
	ClbVipDomain string                      `json:"clb_vip_domain"`
	BkBizID      int64                       `json:"bk_biz_id"`
	Region       string                      `json:"region"`
	Vendor       enumor.Vendor               `json:"vendor"`
	LblID        string                      `json:"lbl_id"`
	CloudLblID   string                      `json:"cloud_lbl_id"`
	Protocol     enumor.ProtocolType         `json:"protocol"`
	Port         int64                       `json:"port"`
	RsList       []*LoadBalancerTargetRsList `json:"rs_list"`
}

// LoadBalancerTargetRsList 负载均衡下的RS列表
type LoadBalancerTargetRsList struct {
	corelb.BaseTarget `json:",inline"`
	RuleID            string          `json:"rule_id"`
	CloudRuleID       string          `json:"cloud_rule_id"`
	RuleType          enumor.RuleType `json:"rule_type"`
	Domain            string          `json:"domain,omitempty"`
	Url               string          `json:"url,omitempty"`
}

// LoadBalancerUrlRuleResult 负载均衡四层/七层规则信息
type LoadBalancerUrlRuleResult struct {
	LbID              string                       `json:"lb_id"`
	CloudClbID        string                       `json:"cloud_lb_id"`
	LblID             string                       `json:"lbl_id"`
	CloudLblID        string                       `json:"cloud_lbl_id"`
	TargetGroupIDs    []string                     `json:"target_group_ids"`
	TargetGrouRuleMap map[string]DomainUrlRuleInfo `json:"target_group_rule_map"`
}

// DomainUrlRuleInfo 负载均衡四层/七层规则信息
type DomainUrlRuleInfo struct {
	TargetGroupID string          `json:"target_group_id"`
	RuleID        string          `json:"rule_id"`
	CloudRuleID   string          `json:"cloud_rule_id"`
	RuleType      enumor.RuleType `json:"rule_type"`
	Domain        string          `json:"domain"`
	Url           string          `json:"url"`
}

// BatchDeleteListenerReq listener batch delete request.
type BatchDeleteListenerReq struct {
	BkBizID           int64                `json:"bk_biz_id" validate:"required"`
	Vendor            enumor.Vendor        `json:"vendor" validate:"required"`
	AccountID         string               `json:"account_id" validate:"required"`
	ListenerQueryList []*ListenerDeleteReq `json:"rule_query_list" validate:"required"`
}

// Validate validate
func (req *BatchDeleteListenerReq) Validate() error {
	if len(req.Vendor) == 0 {
		return errors.New("vendor不能为空")
	}
	if err := req.Vendor.Validate(); err != nil {
		return err
	}
	if len(req.AccountID) == 0 {
		return errors.New("account_id不能为空")
	}
	if len(req.ListenerQueryList) == 0 {
		return errors.New("rule_query_list不能为空")
	}
	for _, item := range req.ListenerQueryList {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(req)
}

// ListenerDeleteReq listener delete request.
type ListenerDeleteReq struct {
	Region        string              `json:"region" validate:"required"`
	ClbVipDomains []string            `json:"clb_vip_domains" validate:"required"`
	CloudLbIDs    []string            `json:"cloud_lb_ids" validate:"required"`
	Protocol      enumor.ProtocolType `json:"protocol" validate:"required"`
	Ports         []int64             `json:"ports" validate:"required"`
}

// Validate request.
func (req *ListenerDeleteReq) Validate() error {
	if !req.Protocol.IsLayer4Protocol() && !req.Protocol.IsLayer7Protocol() {
		return errors.New("protocol is illegal")
	}
	if len(req.ClbVipDomains) != len(req.CloudLbIDs) {
		return errors.New("clb_vip_domains and cloud_lb_ids num must be equal")
	}
	if len(req.CloudLbIDs) > 50 {
		return errors.New("cloud_lb_ids num must be less than 50")
	}
	if len(req.Ports) == 0 {
		return fmt.Errorf("ports不能为空")
	}
	for _, port := range req.Ports {
		if port < 0 {
			return fmt.Errorf("ports must be greater than 0")
		}
	}

	return validator.Validate.Struct(req)
}

// BatchListListenerResp define batch list listener resp.
type BatchListListenerResp struct {
	Details []*corelb.BaseListener `json:"details"`
}