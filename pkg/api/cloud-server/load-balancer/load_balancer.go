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

package cslb

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// BatchBindLbSecurityGroupReq batch bind lb security group req.
type BatchBindLbSecurityGroupReq struct {
	ClbID            string   `json:"clb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate validate.
func (req *BatchBindLbSecurityGroupReq) Validate() error {
	if len(req.ClbID) == 0 {
		return errors.New("clb_id is required")
	}

	if len(req.SecurityGroupIDs) == 0 {
		return errors.New("security_group_ids is required")
	}

	if len(req.SecurityGroupIDs) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("security_group_ids should <= %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// UnBindClbSecurityGroupReq unbind clb security group req.
type UnBindClbSecurityGroupReq struct {
	ClbID           string `json:"clb_id" validate:"required"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
}

// Validate validate.
func (req *UnBindClbSecurityGroupReq) Validate() error {
	if len(req.ClbID) == 0 {
		return errors.New("clb_id is required")
	}

	if len(req.SecurityGroupID) == 0 {
		return errors.New("security_group_id is required")
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Associate --------------------------

// SecurityGroupAssociateClbReq define security group associate clb option.
type SecurityGroupAssociateClbReq struct {
	ClbID            string   `json:"clb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate security group associate clb request.
func (req *SecurityGroupAssociateClbReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AssignLbToBizReq define assign clb to biz req.
type AssignLbToBizReq struct {
	BkBizID int64    `json:"bk_biz_id" validate:"required,min=0"`
	LbIDs   []string `json:"lb_ids" validate:"required,min=1"`
}

// Validate assign clb to biz request.
func (req *AssignLbToBizReq) Validate() error {

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.LbIDs) == 0 {
		return errors.New("lb ids is required")
	}

	if len(req.LbIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("lb ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List Listener--------------------------

// ListListenerResult defines list listener result.
type ListListenerResult = core.ListResultT[ListListenerBase]

// ListListenerBase define list listener base.
type ListListenerBase struct {
	corelb.BaseListener
	TargetGroupID      string                        `json:"target_group_id"`
	Scheduler          string                        `json:"scheduler"`
	SessionType        string                        `json:"session_type"`
	SessionExpire      int64                         `json:"session_expire"`
	DomainNum          int64                         `json:"domain_num"`
	UrlNum             int64                         `json:"url_num"`
	HealthCheck        *corelb.TCloudHealthCheckInfo `json:"health_check"`
	Certificate        *corelb.TCloudCertificateInfo `json:"certificate"`
	RsWeightZeroNum    int64                         `json:"rs_weight_zero_num"`
	RsWeightNonZeroNum int64                         `json:"rs_weight_non_zero_num"`
	BindingStatus      enumor.BindingStatus          `json:"binding_status"`
}

// -------------------------- Get Listener --------------------------

// GetTCloudListenerDetail define get tcloud listener detail.
type GetTCloudListenerDetail struct {
	corelb.TCloudListener
	LblID              string                        `json:"lbl_id"`
	LblName            string                        `json:"lbl_name"`
	CloudLblID         string                        `json:"cloud_lbl_id"`
	TargetGroupID      string                        `json:"target_group_id"`
	TargetGroupName    string                        `json:"target_group_name"`
	CloudTargetGroupID string                        `json:"cloud_target_group_id"`
	Scheduler          string                        `json:"scheduler"`
	SessionType        string                        `json:"session_type"`
	SessionExpire      int64                         `json:"session_expire"`
	HealthCheck        *corelb.TCloudHealthCheckInfo `json:"health_check"`
	Certificate        *corelb.TCloudCertificateInfo `json:"certificate"`
	DomainNum          int64                         `json:"domain_num"`
	UrlNum             int64                         `json:"url_num"`
}

// -------------------------- List LoadBalancer Url Rule --------------------------

// ListLbUrlRuleResult defines list lb url rule result.
type ListLbUrlRuleResult = core.ListResultT[ListLbUrlRuleBase]

// ListLbUrlRuleBase define list lb url rule base.
type ListLbUrlRuleBase struct {
	corelb.TCloudLbUrlRule
	LblName              string              `json:"lbl_name"`
	LbName               string              `json:"lb_name"`
	PrivateIPv4Addresses []string            `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string            `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string            `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string            `json:"public_ipv6_addresses"`
	Protocol             enumor.ProtocolType `json:"protocol"`
	Port                 int64               `json:"port"`
	VpcID                string              `json:"vpc_id"`
	VpcName              string              `json:"vpc_name"`
	CloudVpcID           string              `json:"cloud_vpc_id"`
	InstType             enumor.InstType     `json:"inst_type"`
}

// -------------------------- List TargetGroup --------------------------

// ListTargetGroupResult defines list target group result.
type ListTargetGroupResult = core.ListResultT[ListTargetGroupSummary]

// ListTargetGroupSummary define list listener summary.
type ListTargetGroupSummary struct {
	corelb.BaseTargetGroup
	LbID                 string   `json:"lb_id"`
	LbName               string   `json:"lb_name"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
	ListenerNum          int64    `json:"listener_num"`
}

// -------------------------- Get TargetGroup --------------------------

// GetTargetGroupDetail define get target group detail.
type GetTargetGroupDetail struct {
	corelb.BaseTargetGroup
	TargetList []corelb.BaseTarget `json:"target_list"`
}

// GetListenerDomainResult 监听器下域名列表
type GetListenerDomainResult struct {
	DefaultDomain string       `json:"default_domain"`
	DomainList    []DomainInfo `json:"domain_list"`
}

// DomainInfo 七层监听器下的域名信息
type DomainInfo struct {
	Domain   string `json:"domain"`
	UrlCount int    `json:"url_count"`
}

// -------------------------- Create Target Group Listener Rel --------------------------

// TargetGroupListenerRelAssociateReq target group listener rel associate req.
type TargetGroupListenerRelAssociateReq struct {
	ListenerID     string `json:"listener_id" validate:"required"`
	ListenerRuleID string `json:"listener_rule_id" validate:"required"`
	TargetGroupID  string `json:"target_group_id" validate:"required"`
}

// Validate validate target group listener rel associate
func (req *TargetGroupListenerRelAssociateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------[批量添加RS]--------------------------

// TCloudTargetBatchCreateReq tcloud target batch create req.
type TCloudTargetBatchCreateReq struct {
	TargetGroups []*TCloudBatchAddTargetReq `json:"target_groups" validate:"required,min=1,max=10,dive"`
}

// Validate request.
func (req *TCloudTargetBatchCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudBatchAddTargetReq tcloud target batch operate req.
type TCloudBatchAddTargetReq struct {
	TargetGroupID string                 `json:"target_group_id" validate:"required"`
	Targets       []*cloud.TargetBaseReq `json:"targets" validate:"required,min=1,max=100,dive"`
}

// --------------------------[批量移除RS]--------------------------

// TCloudTargetBatchRemoveReq tcloud target batch remove req.
type TCloudTargetBatchRemoveReq struct {
	TargetGroups []*TCloudRemoveTargetReq `json:"target_groups" validate:"required,min=1,max=10,dive"`
}

// Validate request.
func (req *TCloudTargetBatchRemoveReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRemoveTargetReq tcloud remove target req.
type TCloudRemoveTargetReq struct {
	TargetGroupID string   `json:"target_group_id" validate:"required"`
	TargetIDs     []string `json:"target_ids" validate:"required,min=1,max=100,dive"`
}

// TCloudRuleBatchCreateReq tcloud lb url rule batch create req.
type TCloudRuleBatchCreateReq struct {
	Rules []TCloudRuleCreate `json:"rules" validate:"min=1,dive"`
}

// Validate request.
func (req *TCloudRuleBatchCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRuleCreate 腾讯云url规则创建
type TCloudRuleCreate struct {
	Url string `json:"url,omitempty" validate:"required"`

	TargetGroupID string `json:"target_group_id" validate:"required"`

	Domains           []string `json:"domains,omitempty"`
	SessionExpireTime *int64   `json:"session_expire_time,omitempty"`
	Scheduler         *string  `json:"scheduler,omitempty"`
	ForwardType       *string  `json:"forward_type,omitempty"`
	DefaultServer     *bool    `json:"default_server,omitempty"`
	Http2             *bool    `json:"http2,omitempty"`
	TargetType        *string  `json:"target_type,omitempty"`
	Quic              *bool    `json:"quic,omitempty"`
	TrpcFunc          *string  `json:"trpc_func,omitempty"`
	TrpcCallee        *string  `json:"trpc_callee,omitempty"`

	Certificates *corelb.TCloudCertificateInfo `json:"certificates,omitempty"`

	Memo *string `json:"memo,omitempty"`
}

// Validate request.
func (req *TCloudRuleCreate) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------[批量修改RS端口]--------------------------

// TCloudBatchModifyTargetPortReq tcloud batch modify target port req.
type TCloudBatchModifyTargetPortReq struct {
	TargetIDs []string `json:"target_ids" validate:"required,min=1,max=20"`
	NewPort   int64    `json:"new_port" validate:"required,min=1,max=65535"`
}

// Validate request.
func (req *TCloudBatchModifyTargetPortReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------[批量修改RS权重]--------------------------

// TCloudBatchModifyTargetWeightReq tcloud batch modify target weight req.
type TCloudBatchModifyTargetWeightReq struct {
	TargetIDs []string `json:"target_ids" validate:"required,min=1,max=100"`
	NewWeight *int64   `json:"new_weight" validate:"required,min=0,max=100"`
}

// Validate request.
func (req *TCloudBatchModifyTargetWeightReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update Target Group --------------------------

// TargetGroupUpdateReq ...
type TargetGroupUpdateReq struct {
	IDs        []string            `json:"ids" validate:"omitempty"`
	Name       string              `json:"name,omitempty"`
	VpcID      string              `json:"vpc_id,omitempty"`
	CloudVpcID string              `json:"cloud_vpc_id,omitempty"`
	Region     string              `json:"region,omitempty"`
	Protocol   enumor.ProtocolType `json:"protocol,omitempty"`
	Port       int64               `json:"port,omitempty"`
	Weight     *int64              `json:"weight,omitempty"`
}

// Validate ...
func (req *TargetGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------  Terminate Async Flow --------------------------

// AsyncFlowTerminateReq terminate async flow req.
type AsyncFlowTerminateReq struct {
	FlowID string `json:"flow_id" validate:"required"`
}

// Validate ...
func (req *AsyncFlowTerminateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------  Clone Async Flow --------------------------

// AsyncFlowCloneReq terminate async flow req.
type AsyncFlowCloneReq struct {
	FlowID string `json:"flow_id" validate:"required"`
}

// Validate ...
func (req *AsyncFlowCloneReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------  Retry Async Flow Task --------------------------

// AsyncTaskRetryReq retry async flow task req.
type AsyncTaskRetryReq struct {
	FlowID string `json:"flow_id" validate:"required"`
	TaskID string `json:"task_id" validate:"required"`
}

// Validate ...
func (req *AsyncTaskRetryReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------  Get Async Flow Result After Terminate --------------------------

// TerminatedAsyncFlowResultReq get terminated async flow result req.
type TerminatedAsyncFlowResultReq struct {
	FlowID  string   `json:"flow_id" validate:"required"`
	TaskIDs []string `json:"task_ids" validate:"dive,required"`
}

// Validate ...
func (req *TerminatedAsyncFlowResultReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ResourceFlowStatusResp define resource flow status response.
type ResourceFlowStatusResp struct {
	Status enumor.ResFlowStatus `json:"status"`
}

// --------------------------  Get Async Flow Result After Terminate --------------------------

// TerminatedAsyncFlowResult  terminated async flow result .
type TerminatedAsyncFlowResult struct {
	TaskID        string               `json:"task_id,omitempty"`
	TargetGroupID string               `json:"target_group_id,omitempty"`
	Targets       []TCloudResultTarget `json:"target_list,omitempty"`
}

// TCloudResultTarget ...
type TCloudResultTarget struct {
	InstType    enumor.InstType `json:"inst_type"`
	CloudInstID string          `json:"cloud_inst_id"`
	InstName    string          `json:"inst_name"`
	Port        int64           `json:"port"`
	Weight      *int64          `json:"weight"`
}
