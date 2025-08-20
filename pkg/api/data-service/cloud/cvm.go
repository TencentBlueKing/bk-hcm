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
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CvmBatchCreateReq cvm create req.
type CvmBatchCreateReq[Extension corecvm.Extension] struct {
	Cvms []CvmBatchCreate[Extension] `json:"cvms" validate:"required"`
}

// CvmBatchCreate define cvm batch create.
type CvmBatchCreate[Extension corecvm.Extension] struct {
	CloudID              string     `json:"cloud_id" validate:"required"`
	Name                 string     `json:"name"`
	BkBizID              int64      `json:"bk_biz_id" validate:"required"`
	BkHostID             int64      `json:"bk_host_id" validate:"required"`
	BkCloudID            int64      `json:"bk_cloud_id" validate:"required"`
	AccountID            string     `json:"account_id" validate:"required"`
	Region               string     `json:"region" validate:"required"`
	Zone                 string     `json:"zone"`
	CloudVpcIDs          []string   `json:"cloud_vpc_ids"`
	VpcIDs               []string   `json:"vpc_ids"`
	CloudSubnetIDs       []string   `json:"cloud_subnet_ids"`
	SubnetIDs            []string   `json:"subnet_ids"`
	CloudImageID         string     `json:"cloud_image_id" validate:"required"`
	ImageID              string     `json:"image_id"`
	OsName               string     `json:"os_name"`
	Memo                 *string    `json:"memo"`
	Status               string     `json:"status" validate:"required"`
	PrivateIPv4Addresses []string   `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string   `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string   `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string   `json:"public_ipv6_addresses"`
	MachineType          string     `json:"machine_type" validate:"required"`
	CloudCreatedTime     string     `json:"cloud_created_time"`
	CloudLaunchedTime    string     `json:"cloud_launched_time"`
	CloudExpiredTime     string     `json:"cloud_expired_time"`
	Extension            *Extension `json:"extension" validate:"required"`
}

// Validate cvm create request.
func (req *CvmBatchCreateReq[T]) Validate() error {
	if len(req.Cvms) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvms count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// CvmBatchUpdateReq cvm batch update req.
type CvmBatchUpdateReq[Extension corecvm.Extension] struct {
	Cvms []CvmBatchUpdateWithExtension[Extension] `json:"cvms" validate:"required"`
}

// CvmBatchUpdate cvm batch update.
type CvmBatchUpdate struct {
	ID                   string   `json:"id" validate:"required"`
	Name                 string   `json:"name"`
	BkBizID              int64    `json:"bk_biz_id" validate:"required"`
	BkHostID             int64    `json:"bk_host_id" validate:"required"`
	BkCloudID            *int64   `json:"bk_cloud_id"`
	CloudVpcIDs          []string `json:"cloud_vpc_ids"`
	VpcIDs               []string `json:"vpc_ids"`
	CloudSubnetIDs       []string `json:"cloud_subnet_ids"`
	SubnetIDs            []string `json:"subnet_ids"`
	CloudImageID         string   `json:"cloud_image_id"`
	ImageID              string   `json:"image_id"`
	Memo                 *string  `json:"memo"`
	Status               string   `json:"status" validate:"required"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
	CloudLaunchedTime    string   `json:"cloud_launched_time"`
	CloudExpiredTime     string   `json:"cloud_expired_time"`
	OsName               string   `json:"os_name"`
	MachineType          string   `json:"machine_type"`
}

// CvmBatchUpdateWithExtension cvm batch update with extension.
type CvmBatchUpdateWithExtension[Extension corecvm.Extension] struct {
	CvmBatchUpdate `json:",inline"`
	Extension      *Extension `json:"extension,omitempty"`
}

// Validate cvm update request.
func (req *CvmBatchUpdateReq[T]) Validate() error {
	if len(req.Cvms) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvms count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// CvmCommonInfoBatchUpdateReq define cvm common info batch update req.
type CvmCommonInfoBatchUpdateReq struct {
	Cvms []CvmCommonInfoBatchUpdateData `json:"cvms" validate:"required"`
}

// Validate cvm common info batch update req.
func (req *CvmCommonInfoBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Cvms) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// CvmCommonInfoBatchUpdateData define cvm common info batch update data.
type CvmCommonInfoBatchUpdateData struct {
	ID        string  `json:"id" validate:"required"`
	BkBizID   *int64  `json:"bk_biz_id"`
	BkCloudID *int64  `json:"bk_cloud_id"`
	BkHostID  *int64  `json:"bk_host_id"`
	Name      *string `json:"name"`
	// PrivateIPv4Addresses 内网IP
	PrivateIPv4Addresses *[]string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses *[]string `json:"private_ipv6_addresses"`
	// PublicIPv6Addresses 公网IP
	PublicIPv4Addresses *[]string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses *[]string `json:"public_ipv6_addresses"`
}

// -------------------------- List --------------------------

// CvmListReq list req.
type CvmListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *CvmListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CvmListResult define cvm list result.
type CvmListResult struct {
	Count   uint64            `json:"count"`
	Details []corecvm.BaseCvm `json:"details"`
}

// CvmListResp define list resp.
type CvmListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CvmListResult `json:"data"`
}

// CvmExtListReq list req.
type CvmExtListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *CvmExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CvmExtListResult define cvm with extension list result.
type CvmExtListResult[T corecvm.Extension] struct {
	Count   uint64           `json:"count,omitempty"`
	Details []corecvm.Cvm[T] `json:"details,omitempty"`
}

// CvmExtListResp define list resp.
type CvmExtListResp[T corecvm.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *CvmExtListResult[T] `json:"data"`
}

// -------------------------- Delete --------------------------

// CvmBatchDeleteReq cvm delete request.
type CvmBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate cvm delete request.
func (req *CvmBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Get --------------------------

// CvmGetResp define cvm get resp.
type CvmGetResp[T corecvm.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *corecvm.Cvm[T] `json:"data"`
}
