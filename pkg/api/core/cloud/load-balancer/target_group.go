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

package loadbalancer

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

// BaseTargetGroup define base target group.
type BaseTargetGroup struct {
	ID              string                 `json:"id"`
	CloudID         string                 `json:"cloud_id"`
	Name            string                 `json:"name"`
	Vendor          enumor.Vendor          `json:"vendor"`
	AccountID       string                 `json:"account_id"`
	BkBizID         int64                  `json:"bk_biz_id"`
	TargetGroupType enumor.TargetGroupType `json:"target_group_type"`
	VpcID           string                 `json:"vpc_id"`
	CloudVpcID      string                 `json:"cloud_vpc_id"`
	Region          string                 `json:"region"`
	Protocol        string                 `json:"protocol"`
	Port            int64                  `json:"port"`
	Weight          int64                  `json:"weight"`
	HealthCheck     *HealthCheckInfo       `json:"health_check"`
	Memo            *string                `json:"memo"`
	*core.Revision  `json:",inline"`
}

// TargetGroup define target group.
type TargetGroup[Ext TargetGroupExtension] struct {
	BaseTargetGroup `json:",inline"`
	Extension       *Ext `json:"extension"`
}

// GetID ...
func (cert TargetGroup[T]) GetID() string {
	return cert.BaseTargetGroup.ID
}

// GetCloudID ...
func (cert TargetGroup[T]) GetCloudID() string {
	return cert.BaseTargetGroup.CloudID
}

// TargetGroupExtension extension.
type TargetGroupExtension interface {
	TCloudTargetGroupExtension
}

// BaseTarget define base target.
type BaseTarget struct {
	ID                 string            `json:"id"`
	AccountID          string            `json:"account_id"`
	InstType           string            `json:"inst_type"`
	InstID             string            `json:"inst_id"`
	CloudInstID        string            `json:"cloud_inst_id"`
	InstName           string            `json:"inst_name"`
	TargetGroupID      string            `json:"target_group_id"`
	CloudTargetGroupID string            `json:"cloud_target_group_id"`
	Port               int64             `json:"port"`
	Weight             int64             `json:"weight"`
	PrivateIPAddress   types.StringArray `json:"private_ip_address"`
	PublicIPAddress    types.StringArray `json:"public_ip_address"`
	Zone               string            `json:"zone"`
	Memo               *string           `json:"memo"`
	*core.Revision     `json:",inline"`
}

// BaseTargetListenerRuleRel define base target listener rule rel.
type BaseTargetListenerRuleRel struct {
	ID               string          `json:"id"`
	ListenerRuleID   string          `json:"listener_rule_id"`
	ListenerRuleType string          `json:"listener_rule_type"`
	TargetGroupID    string          `json:"target_group_id"`
	LbID             string          `json:"lb_id"`
	LblID            string          `json:"lbl_id"`
	BindingStatus    string          `json:"binding_status"`
	Detail           types.JsonField `json:"detail"`
	*core.Revision   `json:",inline"`
}
