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
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create Target Group --------------------------

// TargetGroupCreateReq define target group create.
type TargetGroupCreateReq struct {
	Name            string                 `json:"name" validate:"required"`
	AccountID       string                 `json:"account_id" validate:"required"`
	BkBizID         int64                  `json:"bk_biz_id" validate:"omitempty"`
	Region          string                 `json:"region" validate:"required"`
	Protocol        enumor.ProtocolType    `json:"protocol" validate:"required"`
	Port            int64                  `json:"port" validate:"required"`
	VpcID           string                 `json:"vpc_id" validate:"omitempty"`
	CloudVpcID      string                 `json:"cloud_vpc_id" validate:"required"`
	TargetGroupType enumor.TargetGroupType `json:"target_group_type" validate:"omitempty"`
	Weight          int64                  `json:"weight" validate:"omitempty"`
	HealthCheck     types.JsonField        `json:"health_check" validate:"omitempty"`
	Memo            *string                `json:"memo"`
	RsList          []*TargetBaseReq       `json:"rs_list" validate:"omitempty"`
}

// Validate 验证目标组创建参数
func (req *TargetGroupCreateReq) Validate() error {
	if req.RsList != nil {
		for _, item := range req.RsList {
			if err := item.Validate(); err != nil {
				return err
			}
		}
	}
	return validator.Validate.Struct(req)
}

// TargetBaseReq Target基本参数
type TargetBaseReq struct {
	ID               string          `json:"id" validate:"omitempty"`
	IP               string          `json:"ip" validate:"omitempty"`
	InstType         enumor.InstType `json:"inst_type" validate:"required"`
	Port             int64           `json:"port" validate:"required"`
	Weight           *int64          `json:"weight" validate:"required"`
	AccountID        string          `json:"account_id,omitempty" validate:"omitempty"`
	TargetGroupID    string          `json:"target_group_id,omitempty" validate:"omitempty"`
	CloudInstID      string          `json:"cloud_inst_id" validate:"omitempty"`
	InstName         string          `json:"inst_name,omitempty" validate:"omitempty"`
	PrivateIPAddress []string        `json:"private_ip_address,omitempty" validate:"omitempty"`
	PublicIPAddress  []string        `json:"public_ip_address,omitempty" validate:"omitempty"`
	CloudVpcIDs      []string        `json:"cloud_vpc_ids,omitempty" validate:"omitempty"`
	Zone             string          `json:"zone,omitempty" validate:"omitempty"`
	NewPort          *int64          `json:"new_port,omitempty" validate:"omitempty"`
	NewWeight        *int64          `json:"new_weight,omitempty" validate:"omitempty"`
}

// Validate ...
func (req *TargetBaseReq) Validate() error {
	switch req.InstType {
	case enumor.CvmInstType, enumor.EniInstType, enumor.CcnInstType:
	default:
		return errf.Newf(errf.InvalidParameter, "inst_type not supportted %s", req.InstType)
	}

	return validator.Validate.Struct(req)
}

// TargetGroupBatchCreateReq target group create req.
type TargetGroupBatchCreateReq[Extension corelb.TargetGroupExtension] struct {
	TargetGroups []TargetGroupBatchCreate[Extension] `json:"target_groups" validate:"required,min=1,dive"`
}

// Validate target group create request.
func (req *TargetGroupBatchCreateReq[T]) Validate() error {
	if len(req.TargetGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("target_groups count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudTargetGroupCreateReq ...
type TCloudTargetGroupCreateReq = TargetGroupBatchCreateReq[corelb.TCloudTargetGroupExtension]

// TargetGroupBatchCreate define target group batch create.
type TargetGroupBatchCreate[Extension corelb.TargetGroupExtension] struct {
	Name            string                 `json:"name" validate:"required"`
	Vendor          enumor.Vendor          `json:"vendor" validate:"required"`
	AccountID       string                 `json:"account_id" validate:"required"`
	BkBizID         int64                  `json:"bk_biz_id" validate:"required"`
	Region          string                 `json:"region" validate:"required"`
	Protocol        enumor.ProtocolType    `json:"protocol" validate:"required"`
	Port            int64                  `json:"port" validate:"required"`
	VpcID           string                 `json:"vpc_id" validate:"omitempty"`
	CloudVpcID      string                 `json:"cloud_vpc_id" validate:"required"`
	TargetGroupType enumor.TargetGroupType `json:"target_group_type" validate:"omitempty"`
	Weight          int64                  `json:"weight" validate:"omitempty"`
	HealthCheck     types.JsonField        `json:"health_check" validate:"omitempty"`
	Memo            *string                `json:"memo"`
	Extension       *Extension             `json:"extension"`
	RsList          []*TargetBaseReq       `json:"rs_list" validate:"omitempty"`
}

// Validate 验证目标组创建参数
func (req *TargetGroupBatchCreate[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchCreateTgWithRelReq 创建目标组并绑定监听器/规则.
type BatchCreateTgWithRelReq[Extension corelb.TargetGroupExtension] struct {
	TargetGroups []CreateTargetGroupWithRel[Extension] `json:"target_groups" validate:"required,min=1"`
}

// Validate target group create request.
func (req *BatchCreateTgWithRelReq[T]) Validate() error {
	if len(req.TargetGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("target_groups count should <= %d", constant.BatchOperationMaxLimit)
	}

	for _, item := range req.TargetGroups {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(req)
}

// TCloudBatchCreateTgWithRelReq ...
type TCloudBatchCreateTgWithRelReq = BatchCreateTgWithRelReq[corelb.TCloudTargetGroupExtension]

// CreateTargetGroupWithRel define target group batch create.
type CreateTargetGroupWithRel[Extension corelb.TargetGroupExtension] struct {
	TargetGroup TargetGroupBatchCreate[Extension] `json:"target_group" validate:"required"`

	ListenerRuleID      string               `json:"listener_rule_id" validate:"required"`
	CloudListenerRuleID string               `json:"cloud_listener_rule_id" validate:"required"`
	ListenerRuleType    enumor.RuleType      `json:"listener_rule_type" validate:"required"`
	LbID                string               `json:"lb_id" validate:"required"`
	CloudLbID           string               `json:"cloud_lb_id" validate:"required"`
	LblID               string               `json:"lbl_id" validate:"required"`
	CloudLblID          string               `json:"cloud_lbl_id" validate:"required"`
	BindingStatus       enumor.BindingStatus `json:"binding_status" validate:"omitempty"`
	Detail              types.JsonField      `json:"detail" validate:"omitempty"`
}

// Validate ...
func (r *CreateTargetGroupWithRel[Extension]) Validate() error {
	return validator.Validate.Struct(r)
}

// -------------------------- Update Target Group --------------------------

// TargetGroupUpdateReq ...
type TargetGroupUpdateReq struct {
	IDs             []string                      `json:"ids" validate:"omitempty"`
	BkBizID         int64                         `json:"bk_biz_id"`
	Name            string                        `json:"name"`
	TargetGroupType enumor.TargetGroupType        `json:"target_group_type"`
	VpcID           string                        `json:"vpc_id"`
	CloudVpcID      string                        `json:"cloud_vpc_id"`
	Region          string                        `json:"region"`
	Protocol        enumor.ProtocolType           `json:"protocol"`
	Port            int64                         `json:"port"`
	Weight          *int64                        `json:"weight,omitempty"`
	HealthCheck     *corelb.TCloudHealthCheckInfo `json:"health_check"`
}

// Validate ...
func (req *TargetGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update Target Group Expr --------------------------

// TargetGroupExtUpdateReq ...
type TargetGroupExtUpdateReq[T corelb.TargetGroupExtension] struct {
	ID        string `json:"id" validate:"required"`
	Name      string `json:"name"`
	Vendor    string `json:"vendor"`
	AccountID string `json:"account_id"`
	BkBizID   int64  `json:"bk_biz_id"`

	TargetGroupType enumor.TargetGroupType `json:"target_group_type"`
	VpcID           string                 `json:"vpc_id"`
	CloudVpcID      string                 `json:"cloud_vpc_id"`
	Region          string                 `json:"region"`
	Protocol        enumor.ProtocolType    `json:"protocol"`
	Port            int64                  `json:"port"`
	Weight          int64                  `json:"weight"`
	HealthCheck     types.JsonField        `json:"health_check"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
	Extension      *T `json:"extension"`
}

// Validate ...
func (req *TargetGroupExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// TargetGroupBatchUpdateReq 目标组批量更新参数
type TargetGroupBatchUpdateReq[T corelb.TargetGroupExtension] []*TargetGroupExtUpdateReq[T]

// Validate ...
func (req *TargetGroupBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- List Target Group --------------------------

// TargetGroupListResult define target group list result.
type TargetGroupListResult = core.ListResultT[corelb.BaseTargetGroup]

// TargetGroupExtListResult define clb with extension list result.
type TargetGroupExtListResult[T corelb.TargetGroupExtension] struct {
	Count   uint64                  `json:"count,omitempty"`
	Details []corelb.TargetGroup[T] `json:"details,omitempty"`
}

// -------------------------- Delete Target Group --------------------------

// TargetGroupBatchDeleteReq delete request.
type TargetGroupBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *TargetGroupBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List Target Listener Rule Rel --------------------------

// TargetListenerRuleRelListResult define target listener rule rel list result.
type TargetListenerRuleRelListResult = core.ListResultT[corelb.BaseTargetListenerRuleRel]

// -------------------------- Create Target Group Listener Rel --------------------------

// TargetGroupListenerRelCreateReq target group listener rel create req.
type TargetGroupListenerRelCreateReq struct {
	ListenerRuleID      string               `json:"listener_rule_id" validate:"required"`
	CloudListenerRuleID string               `json:"cloud_listener_rule_id" validate:"required"`
	ListenerRuleType    enumor.RuleType      `json:"listener_rule_type" validate:"required"`
	TargetGroupID       string               `json:"target_group_id" validate:"required"`
	CloudTargetGroupID  string               `json:"cloud_target_group_id" validate:"required"`
	LbID                string               `json:"lb_id" validate:"required"`
	CloudLbID           string               `json:"cloud_lb_id" validate:"required"`
	LblID               string               `json:"lbl_id" validate:"required"`
	CloudLblID          string               `json:"cloud_lbl_id" validate:"required"`
	BindingStatus       enumor.BindingStatus `json:"binding_status" validate:"omitempty"`
	Detail              types.JsonField      `json:"detail" validate:"omitempty"`
}

// Validate 验证目标组与监听器关系接口的参数
func (req *TargetGroupListenerRelCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TGListenerRelStatusUpdateReq 状态 详情更新
type TGListenerRelStatusUpdateReq struct {
	BindingStatus enumor.BindingStatus `json:"binding_status" validate:"omitempty"`
	Detail        types.JsonField      `json:"detail" validate:"omitempty"`
}

// Validate ...
func (req *TGListenerRelStatusUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create Listener --------------------------

// TCloudListenerBatchCreateReq ...
type TCloudListenerBatchCreateReq = ListenerBatchCreateReq[corelb.TCloudListenerExtension]

// ListenerBatchCreateReq listener batch create req.
type ListenerBatchCreateReq[T corelb.ListenerExtension] struct {
	Listeners []ListenersCreateReq[T] `json:"listeners" validate:"required,min=1,dive,required"`
}

// Validate 验证监听器批量创建的参数
func (req *ListenerBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ListenersCreateReq listener create req.
type ListenersCreateReq[Extension corelb.ListenerExtension] struct {
	CloudID       string              `json:"cloud_id" validate:"required"`
	Name          string              `json:"name" validate:"required"`
	Vendor        enumor.Vendor       `json:"vendor" validate:"required"`
	AccountID     string              `json:"account_id" validate:"required"`
	BkBizID       int64               `json:"bk_biz_id" validate:"omitempty"`
	LbID          string              `json:"lb_id" validate:"required"`
	CloudLbID     string              `json:"cloud_lb_id" validate:"required"`
	Protocol      enumor.ProtocolType `json:"protocol" validate:"required"`
	Port          int64               `json:"port" validate:"required"`
	DefaultDomain string              `json:"default_domain" validate:"omitempty"`
	Extension     *Extension          `json:"extension"`
}

// Validate 验证监听器创建参数
func (req *ListenersCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create Listener && Rule --------------------------

// ListenerWithRuleBatchCreateReq listener with rule batch create req.
type ListenerWithRuleBatchCreateReq struct {
	ListenerWithRules []ListenerWithRuleCreateReq `json:"listener_with_rules" validate:"required,min=1"`
}

// Validate 验证监听器跟规则批量创建的参数
func (req *ListenerWithRuleBatchCreateReq) Validate() error {
	for _, item := range req.ListenerWithRules {
		if err := item.Validate(); err != nil {
			return errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	return validator.Validate.Struct(req)
}

// ListenerWithRuleCreateReq listener with rule create req.
type ListenerWithRuleCreateReq struct {
	CloudID   string              `json:"cloud_id" validate:"required"`
	Name      string              `json:"name" validate:"required"`
	Vendor    enumor.Vendor       `json:"vendor" validate:"required"`
	AccountID string              `json:"account_id" validate:"required"`
	BkBizID   int64               `json:"bk_biz_id" validate:"omitempty"`
	LbID      string              `json:"lb_id" validate:"required"`
	CloudLbID string              `json:"cloud_lb_id" validate:"required"`
	Protocol  enumor.ProtocolType `json:"protocol" validate:"required"`
	Port      int64               `json:"port" validate:"required"`

	CloudRuleID        string                        `json:"cloud_rule_id" validate:"required"`
	Scheduler          string                        `json:"scheduler" validate:"required"`
	RuleType           enumor.RuleType               `json:"rule_type" validate:"required"`
	SessionType        string                        `json:"session_type" validate:"required"`
	SessionExpire      int64                         `json:"session_expire" validate:"omitempty"`
	TargetGroupID      string                        `json:"target_group_id" validate:"omitempty"`
	CloudTargetGroupID string                        `json:"cloud_target_group_id" validate:"omitempty"`
	Domain             string                        `json:"domain" validate:"omitempty"`
	Url                string                        `json:"url" validate:"omitempty"`
	SniSwitch          enumor.SniType                `json:"sni_switch" validate:"omitempty"`
	Certificate        *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
}

// Validate 验证监听器跟规则创建的参数
func (req *ListenerWithRuleCreateReq) Validate() error {
	if req.SessionExpire > 0 && (req.SessionExpire < 30 || req.SessionExpire > 3600) {
		return errors.New("session_expire must be '0' or between `30` and `3600` ")
	}
	return validator.Validate.Struct(req)
}

// -------------------------- Update Listener --------------------------

// ListenerBatchUpdateReq listener batch update req.
type ListenerBatchUpdateReq[Extension corelb.ListenerExtension] struct {
	Listeners []*ListenerUpdateReq[Extension] `json:"listeners" validate:"required,min=1"`
}

// TCloudListenerUpdateReq ...
type TCloudListenerUpdateReq = ListenerBatchUpdateReq[corelb.TCloudListenerExtension]

// Validate 验证监听器更新参数
func (req *ListenerBatchUpdateReq[T]) Validate() error {
	for _, item := range req.Listeners {
		if err := item.Validate(); err != nil {
			return errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	return validator.Validate.Struct(req)
}

// ListenerUpdateReq listener update req.
type ListenerUpdateReq[Extension corelb.ListenerExtension] struct {
	ID            string         `json:"id" validate:"required"`
	Name          string         `json:"name" validate:"omitempty"`
	BkBizID       int64          `json:"bk_biz_id" validate:"omitempty"`
	SniSwitch     enumor.SniType `json:"sni_switch" validate:"omitempty"`
	DefaultDomain string         `json:"default_domain" validate:"omitempty"`
	Extension     *Extension     `json:"extension"`
}

// Validate 验证监听器更新的参数
func (req *ListenerUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudListenerUpdate ...
type TCloudListenerUpdate = ListenerUpdateReq[corelb.TCloudListenerExtension]

// -------------------------- Create Target --------------------------

// TargetBatchCreateReq batch create target req
type TargetBatchCreateReq struct {
	Targets []*TargetBaseReq `json:"targets" validate:"required,min=1"`
}

// Validate 验证目标组创建参数
func (req *TargetBatchCreateReq) Validate() error {
	if req.Targets != nil {
		for _, item := range req.Targets {
			if len(item.AccountID) == 0 {
				return errf.Newf(errf.InvalidParameter, "account_id is required")
			}
			if len(item.TargetGroupID) == 0 {
				return errf.Newf(errf.InvalidParameter, "target_group_id is required")
			}
			if err := item.Validate(); err != nil {
				return err
			}
		}
	}
	return validator.Validate.Struct(req)
}

// -------------------------- update target --------------------------

// TargetBatchUpdateReq 批量更新RS
type TargetBatchUpdateReq struct {
	Targets []*TargetUpdate `json:"targets" validate:"required,min=1,dive"`
}

// Validate ...
func (r *TargetBatchUpdateReq) Validate() error {
	if len(r.Targets) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("targets length count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(r)
}

// TargetUpdate target update.
type TargetUpdate struct {
	ID string `json:"id" validate:"required,lte=255"`

	InstName         string            `json:"inst_name"`
	Port             int64             `json:"port"`
	Weight           *int64            `json:"weight"`
	PrivateIPAddress types.StringArray `json:"private_ip_address" validate:"omitempty,dive,ip"`
	PublicIPAddress  types.StringArray `json:"public_ip_address" validate:"omitempty,dive,ip"`
	Memo             *string           `json:"memo" validate:"omitempty,lte=255"`
}

// Validate ...
func (req *TargetUpdate) Validate() error {
	return validator.Validate.Struct(req)
}
