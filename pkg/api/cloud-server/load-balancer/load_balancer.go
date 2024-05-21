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

	"hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
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

	Certificate *corelb.TCloudCertificateInfo `json:"certificate,omitempty"`

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
	ResID   string                   `json:"res_id,omitempty"`
	ResType enumor.CloudResourceType `json:"res_type,omitempty"`
	FlowID  string                   `json:"flow_id,omitempty"`
	Status  enumor.ResFlowStatus     `json:"status"`
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

// TCloudBatchCreateReq tcloud batch create req.
type TCloudBatchCreateReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	Region    string  `json:"region" validate:"required"`
	Name      *string `json:"name" validate:"required,max=60"`

	LoadBalancerType loadbalancer.TCloudLoadBalancerType   `json:"load_balancer_type" validate:"required"`
	AddressIPVersion loadbalancer.TCloudIPVersionForCreate `json:"address_ip_version" validate:"omitempty"`

	// 公网	单可用区		传递zones（单元素数组）
	// 公网	主备可用区	传递zones（单元素数组），以及backup_zones
	Zones                   []string `json:"zones" validate:"omitempty"`
	BackupZones             []string `json:"backup_zones" validate:"omitempty"`
	CloudVpcID              *string  `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           *string  `json:"cloud_subnet_id" validate:"omitempty"`
	Vip                     *string  `json:"vip" validate:"omitempty"`
	CloudEipID              *string  `json:"cloud_eip_id" validate:"omitempty"`
	VipIsp                  *string  `json:"vip_isp" validate:"omitempty"`
	InternetChargeType      *string  `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64   `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthPackageID      *string  `json:"bandwidth_package_id" validate:"omitempty"`
	SlaType                 *string  `json:"sla_type" validate:"omitempty"`
	AutoRenew               *bool    `json:"auto_renew" validate:"omitempty"`
	RequireCount            *uint64  `json:"require_count" validate:"omitempty"`
	Memo                    string   `json:"memo" validate:"omitempty"`
}

// Validate request.
func (req *TCloudBatchCreateReq) Validate() error {
	switch req.LoadBalancerType {
	case loadbalancer.InternalLoadBalancerType:
		// 内网校验
		if converter.PtrToVal(req.CloudSubnetID) == "" {
			return errors.New("subnet id is required for load balancer type 'INTERNAL'")
		}
	case loadbalancer.OpenLoadBalancerType:
		if converter.PtrToVal(req.CloudEipID) != "" {
			return errors.New("eip id only support load balancer type 'INTERNAL'")
		}
	default:
		return fmt.Errorf("unknown load balancer type: '%s'", req.LoadBalancerType)
	}

	return validator.Validate.Struct(req)
}

// TargetGroupCreateReq define target group create.
type TargetGroupCreateReq struct {
	Name            string                       `json:"name" validate:"required"`
	AccountID       string                       `json:"account_id" validate:"required"`
	BkBizID         int64                        `json:"bk_biz_id" validate:"omitempty"`
	Region          string                       `json:"region" validate:"required"`
	Protocol        enumor.ProtocolType          `json:"protocol" validate:"required"`
	Port            int64                        `json:"port" validate:"required"`
	VpcID           string                       `json:"vpc_id" validate:"omitempty"`
	CloudVpcID      string                       `json:"cloud_vpc_id" validate:"required"`
	TargetGroupType enumor.TargetGroupType       `json:"target_group_type" validate:"omitempty"`
	Weight          int64                        `json:"weight" validate:"omitempty"`
	HealthCheck     corelb.TCloudHealthCheckInfo `json:"health_check" validate:"omitempty"`
	Memo            *string                      `json:"memo"`
	RsList          []*cloud.TargetBaseReq       `json:"rs_list" validate:"omitempty,dive,required"`
}

// Validate 验证目标组创建参数
func (req *TargetGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}
