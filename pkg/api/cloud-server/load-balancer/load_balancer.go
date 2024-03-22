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
	TargetGroupID string                        `json:"target_group_id"`
	Scheduler     string                        `json:"scheduler"`
	SessionType   string                        `json:"session_type"`
	SessionExpire int64                         `json:"session_expire"`
	DomainNum     int64                         `json:"domain_num"`
	UrlNum        int64                         `json:"url_num"`
	HealthCheck   *corelb.TCloudHealthCheckInfo `json:"health_check"`
	Certificate   *corelb.TCloudCertificateInfo `json:"certificate"`
}

// -------------------------- Get Listener --------------------------

// GetListenerDetail define get listener detail.
type GetListenerDetail struct {
	corelb.BaseListener
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
	corelb.BaseTCloudLbUrlRule
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
	InstType             string              `json:"inst_type"`
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

func (req *TargetGroupListenerRelAssociateReq) Validate() error {
	return validator.Validate.Struct(req)
}
