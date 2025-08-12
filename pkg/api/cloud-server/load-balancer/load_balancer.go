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
	"encoding/json"
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
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
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
type ListListenerResult = core.ListResultT[*ListenerListInfo]

// ListenerListInfo define list listener base.
type ListenerListInfo struct {
	corelb.BaseListener
	EndPort       *int64                        `json:"end_port"`
	TargetGroupID string                        `json:"target_group_id"`
	Scheduler     string                        `json:"scheduler"`
	SessionType   string                        `json:"session_type"`
	SessionExpire int64                         `json:"session_expire"`
	DomainNum     int64                         `json:"domain_num"`
	UrlNum        int64                         `json:"url_num"`
	HealthCheck   *corelb.TCloudHealthCheckInfo `json:"health_check"`
	Certificate   *corelb.TCloudCertificateInfo `json:"certificate"`
	BindingStatus enumor.BindingStatus          `json:"binding_status"`
}

// -------------------------- Get Listener --------------------------

// GetTCloudListenerDetail define get tcloud listener detail.
type GetTCloudListenerDetail struct {
	corelb.TCloudListener
	EndPort            *int64                        `json:"end_port"`
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

// ListTargetWeightNumReq ...
type ListTargetWeightNumReq struct {
	TargetGroupIDs []string `json:"target_group_ids" validate:"required,min=1"`
}

// Validate ...
func (r ListTargetWeightNumReq) Validate() error {
	return validator.Validate.Struct(r)
}

// TargetGroupRsWeightNum 目标组下rs权重统计
type TargetGroupRsWeightNum struct {
	TargetGroupID      string `json:"target_group_id"`
	RsWeightZeroNum    int64  `json:"rs_weight_zero_num"`
	RsWeightNonZeroNum int64  `json:"rs_weight_non_zero_num"`
}

// -------------------------- List LoadBalancer Url Rule --------------------------

// ListLbUrlRuleResult defines list lb url rule result.
type ListLbUrlRuleResult = core.ListResultT[ListLbUrlRuleBase]

// ListLbUrlRuleBase define list lb url rule base.
type ListLbUrlRuleBase struct {
	corelb.TCloudLbUrlRule
	LbName               string   `json:"lb_name"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`

	VpcID      string          `json:"vpc_id"`
	VpcName    string          `json:"vpc_name"`
	CloudVpcID string          `json:"cloud_vpc_id"`
	InstType   enumor.InstType `json:"inst_type"`
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

// --------------------------[批量增加规则]--------------------------

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

// --------------------------[批量移除规则]--------------------------

// TcloudBatchDeleteRuleReq 批量移除规则的请求体，包含七层（url_rule）和四层（listener）规则
type TcloudBatchDeleteRuleReq struct {
	// 要删除的七层规则id列表（url_rule_ids）
	URLRuleIDs []string `json:"url_rule_ids" validate:"required,min=1,max=100"`
	// 要删除的四层规则id列表（listener_ids）
	ListenerIDs []string `json:"listener_ids" validate:"required,min=1,max=100"`
}

// Validate validate request
func (req TcloudBatchDeleteRuleReq) Validate() error {
	if len(req.URLRuleIDs)+len(req.ListenerIDs) == 0 {
		return fmt.Errorf("url_rule_ids and listener_ids cannot both be empty")
	}

	return validator.Validate.Struct(req)
}

// TcloudBatchDeleteRuleIDs delete rule ids
type TcloudBatchDeleteRuleIDs struct {
	// 要删除的七层规则id列表（url_rule_ids）
	URLRuleIDs []string `json:"url_rule_ids" validate:"required,min=1,max=100"`
	// 要删除的四层规则id列表（listener_ids）
	ListenerIDs []string `json:"listener_ids" validate:"required,min=1,max=100"`
}

// Validate request.
func (req *TcloudBatchDeleteRuleIDs) Validate() error {
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

// TgIDAndTCloudBatchModifyTargetWeightReq tcloud batch modify target weight req.
type TgIDAndTCloudBatchModifyTargetWeightReq struct {
	TgID string                           `json:"tg_id" validate:"required"`
	Req  TCloudBatchModifyTargetWeightReq `json:"req" validate:"required"`
}

// Validate request.
func (req *TgIDAndTCloudBatchModifyTargetWeightReq) Validate() error {
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

// --------------------------[标准运维-批量添加RS]--------------------------

// TCloudSopsTargetBatchCreateReq tcloud sops target batch create req.
type TCloudSopsTargetBatchCreateReq struct {
	RuleQueryList []TargetGroupQueryItemForRsOnline `json:"rule_query_list" validate:"required,min=1,max=50,dive"`
	RsIP          []string                          `json:"rs_ip" validate:"required,min=1,max=50"`
	RsPort        []int                             `json:"rs_port" validate:"required,min=1,max=50"`
	RsWeight      int64                             `json:"rs_weight" validate:"required,min=0,max=100"`
	RsType        enumor.InstType                   `json:"rs_type" validate:"required"`
}

// Validate request.
func (req *TCloudSopsTargetBatchCreateReq) Validate() error {
	if len(req.RsIP) != len(req.RsPort) {
		return fmt.Errorf("length of rs_ip and rs_port should be equal")
	}

	if req.RsType != "CVM" && req.RsType != "ENI" {
		return fmt.Errorf("unspoort rs type: %s", req.RsType)
	}

	return validator.Validate.Struct(req)
}

// TargetGroupQueryItemForRsOnline 规则查询结构体-RS上线专属
type TargetGroupQueryItemForRsOnline struct {
	Region   string              `json:"region" validate:"required"`
	Vip      []string            `json:"vip" validate:"max=50"`
	VPort    []int               `json:"vport" validate:"max=50"`
	RsIP     []string            `json:"rs_ip" validate:"max=50"`
	RsType   string              `json:"rs_type"`
	Protocol enumor.ProtocolType `json:"protocol" validate:"required"`
	Domain   []string            `json:"domain" validate:"max=50"`
	Url      []string            `json:"url" validate:"max=50"`
}

// Validate req
func (req *TargetGroupQueryItemForRsOnline) Validate() error {
	if err := req.validateProtocol(); err != nil {
		return err
	}
	if len(req.RsType) != 0 {
		if req.RsType != "CVM" && req.RsType != "ENI" {
			return fmt.Errorf("unspoort rs type: %s", req.RsType)
		}
	}

	//  RSIP 与 VIP 至少填一个
	if len(req.RsIP) == 0 && len(req.Vip) == 0 {
		return fmt.Errorf("rs ip and VIP, fill in at least one")
	}

	// 填写了VIP则必填Vport
	if len(req.Vip) != 0 && len(req.VPort) == 0 {
		return fmt.Errorf("if you fill in VIP, you must fill in Vport")
	}

	// RSIP与RSType同时填或不填
	if (len(req.RsIP) == 0 && len(req.RsType) != 0) || (len(req.RsIP) != 0 && len(req.RsType) == 0) {
		return fmt.Errorf("rs ip and RSType can be filled in at the same time or not")
	}

	// 七层下必填Domain、Url
	if !req.Protocol.IsLayer7Protocol() && (len(req.Domain) != 0 || len(req.Url) != 0) {
		return fmt.Errorf("is not a seven-layer rule, the domain name cannot be filled in")
	}
	if req.Protocol.IsLayer7Protocol() && (len(req.Domain) == 0 || len(req.Url) == 0) {
		return fmt.Errorf("using the seven-layer rule, the domain name must be filled in")
	}

	return validator.Validate.Struct(req)
}

func (req *TargetGroupQueryItemForRsOnline) validateProtocol() error {
	if req.Protocol != enumor.HttpProtocol && req.Protocol != enumor.HttpsProtocol &&
		req.Protocol != enumor.TcpProtocol && req.Protocol != enumor.UdpProtocol {
		return fmt.Errorf("unspoort protocol: %s", req.Protocol)
	}
	return nil
}

// --------------------------[标准运维-批量移除RS]--------------------------

// TCloudSopsTargetBatchRemoveReq tcloud sops target batch remove req.
type TCloudSopsTargetBatchRemoveReq struct {
	RuleQueryList []TargetGroupRuleQueryItem `json:"rule_query_list" validate:"required,min=1,max=10,dive"`
}

// Validate request.
func (req *TCloudSopsTargetBatchRemoveReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TargetGroupRuleQueryItem 目标组规则查询结构体
type TargetGroupRuleQueryItem struct {
	Region   string              `json:"region" validate:"required"`
	Vip      []string            `json:"vip" validate:"max=50"`
	VPort    []int               `json:"vport" validate:"max=50"`
	RsIP     []string            `json:"rs_ip" validate:"required,minx=1,max=50"`
	RsType   string              `json:"rs_type" validate:"required"`
	Protocol enumor.ProtocolType `json:"protocol"`
	Domain   string              `json:"domain"`
}

// Validate ...
func (req *TargetGroupRuleQueryItem) Validate() error {
	if req.RsType != "CVM" && req.RsType != "ENI" {
		return fmt.Errorf("unspoort rs type: %s", req.RsType)
	}

	if len(req.Protocol) != 0 {
		if req.Protocol != enumor.HttpProtocol && req.Protocol != enumor.HttpsProtocol &&
			req.Protocol != enumor.TcpProtocol && req.Protocol != enumor.UdpProtocol {
			return fmt.Errorf("unspoort protocol: %s", req.Protocol)
		}
	}

	if req.Protocol.IsLayer7Protocol() && len(req.Domain) == 0 {
		return fmt.Errorf("using the seven-layer rule, the domain name must be filled in")
	}
	if !req.Protocol.IsLayer7Protocol() && len(req.Domain) != 0 {
		return fmt.Errorf("is not a seven-layer rule, the domain name cannot be filled in")
	}

	return validator.Validate.Struct(req)
}

// --------------------------[标准运维-批量修改权重]--------------------------

// TCloudSopsTargetBatchModifyWeightReq tcloud sops target batch modify weight req.
type TCloudSopsTargetBatchModifyWeightReq struct {
	RuleQueryList []TargetGroupRuleQueryItem `json:"rule_query_list" validate:"required,min=1,max=10,dive"`
	RsWeight      int64                      `json:"rs_weight" validate:"required,min=0,max=100"`
}

// Validate request.
func (req *TCloudSopsTargetBatchModifyWeightReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------[标准运维-批量添加规则]--------------------------
// TODO 依赖于clb excel导入功能，暂时注释，需clb excel导入功能确定后修改
// TCloudSopsRuleBatchCreateReq tloud sops rule batch create request
//type TCloudSopsRuleBatchCreateReq struct {
//	BindRSRecords []*BindRSRecordForSops `json:"bind_rs_records" validate:"required,dive,required"`
//}
//
//// Validate validate req data
//func (req *TCloudSopsRuleBatchCreateReq) Validate() error {
//	return validator.Validate.Struct(req)
//}
//
//// BindRSRecordForSops ...
//type BindRSRecordForSops struct {
//	//Action enumor.BatchOperationActionType `json:"action"`
//
//	ListenerName string              `json:"name"`
//	Protocol     enumor.ProtocolType `json:"protocol"`
//indRSRecordForSops struct {
//	TODO 这里的内容依赖clb excel导入的代码，暂时注释，等待clb excel导入合并后再
//	IPDomainType string `json:"ip_domain_type"`
//	VIP          string `json:"vip"`
//	VPorts       []int  `json:"vports"`
//	HaveEndPort  bool   `json:"have_end_port"` // 是否是端口端
//
//	Domain     string   `json:"domain"`         // 域名
//	URLPath    string   `json:"url"`            // URL路径
//	ServerCert []string `json:"cert_cloud_ids"` // ref
//	ClientCert string   `json:"ca_cloud_id"`    // 客户端证书
//
//	InstType enumor.InstType `json:"inst_type"` // 后端类型 CVM、ENI
//	RSIPs    []string        `json:"rs_ips"`
//	RSPorts  []int           `json:"rs_ports"`
//	Weight   []int           `json:"weight"`
//	//RSInfos  []*lblogic.RSInfo `json:"rs_info"` // 后端实例信息
//
//	Scheduler      string `json:"scheduler"`       // 均衡方式
//	SessionExpired int64  `json:"session_expired"` // 会话保持时间，单位秒
//	HealthCheck    bool   `json:"health_check"`    // 是否开启健康检查
//}

// --------------------------[标准运维-批量移除规则]--------------------------

// TCloudSopsRuleBatchDeleteReq tloud sops rule batch delete request
type TCloudSopsRuleBatchDeleteReq struct {
	RuleQueryList []RuleQueryItemForRuleOffline `json:"rule_query_list" validate:"required,min=1,max=10,dive,required"`
}

// Validate request.
func (req *TCloudSopsRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RuleQueryItemForRuleOffline 规则查询结构体-规则下线专属
type RuleQueryItemForRuleOffline struct {
	Region   string                `json:"region" validate:"required"`
	Vip      []string              `json:"vip" validate:"max=50"`
	VPort    []int                 `json:"vport" validate:"max=50"`
	RsIP     []string              `json:"rs_ip" validate:"max=50"`
	RsType   string                `json:"rs_type" validate:"max=50"`
	Protocol []enumor.ProtocolType `json:"protocol" validate:"max=50"`
	Domain   []string              `json:"domain" validate:"max=50"`
	Url      []string              `json:"url" validate:"max=50"`
}

// Validate ...
func (r *RuleQueryItemForRuleOffline) Validate() error {
	if len(r.RsType) != 0 {
		if r.RsType != "CVM" && r.RsType != "ENI" {
			return fmt.Errorf("unspoort rs type: %s", r.RsType)
		}
	}
	for _, protocol := range r.Protocol {
		if protocol != enumor.HttpProtocol && protocol != enumor.HttpsProtocol &&
			protocol != enumor.TcpProtocol && protocol != enumor.UdpProtocol {
			return fmt.Errorf("unspoort protocol: %s", protocol)
		}
	}

	if len(r.Vip) == 0 && len(r.RsIP) == 0 {
		return fmt.Errorf("vip and RsIP must specify at least one")
	}
	if (len(r.RsIP) == 0 && len(r.RsType) != 0) || len(r.RsIP) != 0 && len(r.RsType) == 0 {
		return fmt.Errorf("rsIP and RsType must be specified at the same time or not at the same time")
	}

	return validator.Validate.Struct(r)
}

// TCloudCreateSnatIpReq ...
type TCloudCreateSnatIpReq struct {
	SnatIps []*corelb.SnatIp `json:"snat_ips" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *TCloudCreateSnatIpReq) Validate() error {
	return validator.Validate.Struct(r)
}

// TCloudDeleteSnatIpReq ...
type TCloudDeleteSnatIpReq struct {
	DeleteIps []string `json:"delete_ips" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *TCloudDeleteSnatIpReq) Validate() error {
	return validator.Validate.Struct(r)
}

// UploadExcelFileBaseResp ...
type UploadExcelFileBaseResp struct {
	Details interface{} `json:"details"`
}

// ImportExcelReq ...
type ImportExcelReq struct {
	AccountID string                      `json:"account_id" validate:"required"`
	RegionIDs []string                    `json:"region_ids" validate:"required,min=1,dive,required"`
	Source    enumor.TaskManagementSource `json:"source"`
	Details   json.RawMessage             `json:"details"`
}

// Validate ...
func (i *ImportExcelReq) Validate() error {
	if err := i.Source.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(i)
}

// ImportValidateReq ...
type ImportValidateReq struct {
	AccountID string          `json:"account_id" validate:"required"`
	RegionIDs []string        `json:"region_ids" validate:"required,min=1,dive,required"`
	Details   json.RawMessage `json:"details"`
}

// Validate ...
func (i *ImportValidateReq) Validate() error {
	return validator.Validate.Struct(i)
}

// RuleBindingStatusListReq ...
type RuleBindingStatusListReq struct {
	RuleIDs []string `json:"rule_ids" validate:"required,min=1,max=100"`
}

// Validate ...
func (i *RuleBindingStatusListReq) Validate() error {
	return validator.Validate.Struct(i)
}

// RuleBindingStatusListResp ...
type RuleBindingStatusListResp struct {
	Details []RuleBindingStatus `json:"details"`
}

// RuleBindingStatus ...
type RuleBindingStatus struct {
	RuleID     string               `json:"rule_id"`
	BindStatus enumor.BindingStatus `json:"binding_status"`
}

// ListListenerTargetsStatReq ...
type ListListenerTargetsStatReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate ...
func (r *ListListenerTargetsStatReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ListenerTargetsStat 监听器 绑定的RS权重情况统计
type ListenerTargetsStat struct {
	NonZeroWeightCount int `json:"non_zero_weight_count"`
	ZeroWeightCount    int `json:"zero_weight_count"`
	TotalCount         int `json:"total_count"`
}

// ExportListenerReq 导出业务下监听器及其下面的资源
type ExportListenerReq struct {
	Listeners []ExportListener `json:"listeners"`
}

// Validate ...
func (r *ExportListenerReq) Validate() error {
	if len(r.Listeners) == 0 {
		return errors.New("listeners required")
	}
	if len(r.Listeners) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("listeners count should <= %d", constant.BatchOperationMaxLimit)
	}

	for _, l := range r.Listeners {
		if err := l.Validate(); err != nil {
			return err
		}
	}

	_, lblIDs := r.GetPartLbAndLblIDs()
	if len(lblIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("lbl_ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// GetAllLbIDs 获取所有负载均衡id
func (r *ExportListenerReq) GetAllLbIDs() []string {
	lbIDMap := make(map[string]struct{})
	for _, l := range r.Listeners {
		lbIDMap[l.LbID] = struct{}{}
	}

	return maps.Keys(lbIDMap)
}

// GetPartLbAndLblIDs 获取负载均衡id和监听器id，当参数传了监听器id, 不返回对应的负载均衡的id
func (r *ExportListenerReq) GetPartLbAndLblIDs() ([]string, []string) {
	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	for _, l := range r.Listeners {
		if len(l.LblIDs) != 0 {
			lblIDs = append(lblIDs, l.LblIDs...)
			continue
		}
		lbIDs = append(lbIDs, l.LbID)
	}

	return slice.Unique(lbIDs), slice.Unique(lblIDs)
}

// ExportListener ...
type ExportListener struct {
	LbID   string   `json:"lb_id"`
	LblIDs []string `json:"lbl_ids"`
}

// Validate ...
func (r *ExportListener) Validate() error {
	if len(r.LbID) == 0 {
		return errors.New("lb_id required")
	}

	if len(r.LblIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("lbl_ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// ExportListenerResp ...
type ExportListenerResp struct {
	Pass   bool   `json:"pass"`
	Reason string `json:"reason"`
}
