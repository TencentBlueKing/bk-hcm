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
	"fmt"

	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// ClbBatchCreateReq clb create req.
type ClbBatchCreateReq[Extension coreclb.Extension] struct {
	Clbs []ClbBatchCreate[Extension] `json:"clbs" validate:"required,min=1"`
}

type TCloudCLBCreateReq = ClbBatchCreateReq[coreclb.TCloudClbExtension]

// ClbBatchCreate define clb batch create.
type ClbBatchCreate[Extension coreclb.Extension] struct {
	CloudID   string `json:"cloud_id" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Vendor    string `json:"vendor" validate:"required"`
	AccountID string `json:"account_id" validate:"required"`
	BkBizID   int64  `json:"bk_biz_id" validate:"omitempty"`

	LoadBalancerType     string   `json:"load_balancer_type" validate:"required"`
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

// Validate clb create request.
func (req *ClbBatchCreateReq[T]) Validate() error {
	if len(req.Clbs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("clbs count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// ClbExtUpdateReq ...
type ClbExtUpdateReq[T coreclb.Extension] struct {
	ID        string `json:"id" validate:"required"`
	Name      string `json:"name"`
	Vendor    string `json:"vendor"`
	AccountID string `json:"account_id"`
	BkBizID   uint64 `json:"bk_biz_id"`

	Region               string   `json:"region" validate:"omitempty"`
	Zones                []string `json:"zones"`
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
	Memo                 *string  `json:"memo"`

	*core.Revision `json:",inline"`
	Extension      *T `json:"extension"`
}

// Validate ...
func (req *ClbExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ClbExtBatchUpdateReq ...
type ClbExtBatchUpdateReq[T coreclb.Extension] []*ClbExtUpdateReq[T]

// Validate ...
func (req *ClbExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- List --------------------------

// ClbListResult define clb list result.
type ClbListResult struct {
	Count   uint64            `json:"count"`
	Details []coreclb.BaseClb `json:"details"`
}

// ClbListResp define list resp.
type ClbListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ClbListResult `json:"data"`
}

// ClbExtListReq list req.
type ClbExtListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *ClbExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ClbExtListResult define clb with extension list result.
type ClbExtListResult[T coreclb.Extension] struct {
	Count   uint64           `json:"count,omitempty"`
	Details []coreclb.Clb[T] `json:"details,omitempty"`
}

// ClbExtListResp define clb list resp.
type ClbExtListResp[T coreclb.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *ClbExtListResult[T] `json:"data"`
}

// ClbListExtResp ...
type ClbListExtResp[T coreclb.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *ClbListExtResult[T] `json:"data"`
}

// ClbListExtResult ...
type ClbListExtResult[T coreclb.Extension] struct {
	Count   uint64            `json:"count,omitempty"`
	Details []*coreclb.Clb[T] `json:"details"`
}

// -------------------------- Delete --------------------------

// ClbBatchDeleteReq delete request.
type ClbBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *ClbBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
