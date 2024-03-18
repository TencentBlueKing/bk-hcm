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
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// LoadBalancerBatchCreateReq load balancer create req.
type LoadBalancerBatchCreateReq[Extension corelb.Extension] struct {
	Lbs []ClbBatchCreate[Extension] `json:"lbs" validate:"required,min=1"`
}

type TCloudCLBCreateReq = LoadBalancerBatchCreateReq[corelb.TCloudClbExtension]

// ClbBatchCreate define load balancer batch create.
type ClbBatchCreate[Extension corelb.Extension] struct {
	CloudID   string        `json:"cloud_id" validate:"required"`
	Name      string        `json:"name" validate:"required"`
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	BkBizID   int64         `json:"bk_biz_id" validate:"omitempty"`

	LoadBalancerType     string               `json:"load_balancer_type" validate:"required"`
	IPVersion            enumor.IPAddressType `json:"ip_version" validate:"required"`
	Region               string               `json:"region" validate:"omitempty"`
	Zones                []string             `json:"zones" `
	BackupZones          []string             `json:"backup_zones"`
	VpcID                string               `json:"vpc_id" validate:"omitempty"`
	CloudVpcID           string               `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID             string               `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID        string               `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4Addresses []string             `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string             `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string             `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string             `json:"public_ipv6_addresses"`
	Domain               string               `json:"domain"`
	Status               string               `json:"status"`
	CloudCreatedTime     string               `json:"cloud_created_time"`
	CloudStatusTime      string               `json:"cloud_status_time"`
	CloudExpiredTime     string               `json:"cloud_expired_time"`

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
	Name    string `json:"name"`
	BkBizID int64  `json:"bk_biz_id"`

	IPVersion            enumor.IPAddressType `json:"ip_version"`
	PrivateIPv4Addresses []string             `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string             `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string             `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string             `json:"public_ipv6_addresses"`
	Domain               string               `json:"domain"`
	Status               string               `json:"status"`
	CloudCreatedTime     string               `json:"cloud_created_time"`
	CloudStatusTime      string               `json:"cloud_status_time"`
	CloudExpiredTime     string               `json:"cloud_expired_time"`
	Memo                 *string              `json:"memo"`

	*core.Revision `json:",inline"`
	Extension      *T `json:"extension"`
}

// Validate ...
func (req *LoadBalancerExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ClbExtBatchUpdateReq ...
type ClbExtBatchUpdateReq[T corelb.Extension] struct {
	Lbs []*LoadBalancerExtUpdateReq[T] `json:"lbs" validate:"min=1"`
}

// Validate ...
func (req *ClbExtBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

type TCloudClbBatchUpdateReq = ClbExtBatchUpdateReq[corelb.TCloudClbExtension]

// ClbBizBatchUpdateReq 批量更新业务id
type ClbBizBatchUpdateReq struct {
	IDs     []string `json:"ids" validate:"required"`
	BkBizID int64    `json:"bk_biz_id" validate:"required"`
}

// Validate ...
func (req *ClbBizBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.IDs) == 0 {
		return errors.New("ids required")
	}

	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// LbListResult define lb list result.
type LbListResult = core.ListResultT[corelb.BaseLoadBalancer]

// LbExtListResult define lb with extension list result.
type LbExtListResult[T corelb.Extension] struct {
	Count   uint64                   `json:"count,omitempty"`
	Details []corelb.LoadBalancer[T] `json:"details,omitempty"`
}

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

// ListListenerReq ...
type ListListenerReq struct {
	LbID         string `json:"lb_id" validate:"omitempty"`
	core.ListReq `json:",inline"`
}

// Validate list request.
func (req *ListListenerReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListenerListResult define listener list result.
type ListenerListResult = core.ListResultT[corelb.BaseListener]

// -------------------------- List Target --------------------------

// TargetListResult define target list result.
type TargetListResult = core.ListResultT[corelb.BaseTarget]

// -------------------------- List TCloud Url Rule --------------------------

// TCloudURLRuleListResult define tcloud url rule list result.
type TCloudURLRuleListResult = core.ListResultT[corelb.BaseTCloudLbUrlRule]
