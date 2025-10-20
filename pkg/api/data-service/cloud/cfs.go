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

// Package cloud 如下:
// data-service cfs request
package cloud

import (
	"hcm/pkg/api/core"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CfsCreateReq Cfs create req.
type CfsCreateReq[Extension corecfs.Extension] struct {
	// BkBizID 业务id
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
	// AccountID 账号id
	AccountID string `json:"account_id" validate:"required"`

	// Name 资源名称
	Name string `json:"name"`
	// CloudID 云上资源id
	CloudID string `json:"cloud_id" validate:"required"`
	// Region 所属地域
	Region string `json:"region" validate:"required"`
	// Zone 所属可用区
	Zone string `json:"zone" validate:"omitempty"`
	// SizeLimit 文件系统最大空间限制(单位:GiB, 示例值：50)
	SizeLimit uint64 `json:"size_limit" validate:"omitempty"`
	// SizeByte 文件系统已使用容量.单位：Byte; 示例值：10
	SizeByte uint64 `json:"size_byte" validate:"omitempty"`
	// AvailCapacity 文件系统剩余容量. 单位：Byte. 示例值：10
	AvailCapacity uint64 `json:"avail_capacity" validate:"omitempty"`
	// BandwidthLimit 文件系统吞吐上限，吞吐上限是根据文件系统当前已使用存储量、绑定的存储资源包以及吞吐资源包一同确定. 单位MiB/s
	BandwidthLimit float64 `json:"bandwidth_limit" validate:"omitempty"`
	// Protocol 文件系统协议类型, 支持 NFS,CIFS,TURBO; 示例值：NFS
	Protocol string `json:"protocol" validate:"omitempty"`
	// StorageType 文件系统存储类型. HP：通用性能型；SD：通用标准型；TP:turbo性能型；TB：turbo标准型；THP：吞吐型; 示例值：HP
	StorageType string `json:"storage_type" validate:"omitempty"`
	// Encrypted 文件系统是否加密
	Encrypted bool `json:"encrypted" validate:"omitempty"`
	// CryptKeyId 加密所使用的密钥，可以为密钥的 ID 或者 ARN
	CryptKeyId string `json:"crypt_key_id" validate:"omitempty"`

	// CloudVpcIDs 云上vpc
	CloudVpcIDs []string `json:"cloud_vpc_ids" validate:"omitempty"`
	// CloudSubnetIDs 云上子网
	CloudSubnetIDs []string `json:"cloud_subnet_ids" validate:"omitempty"`
	// VpcIDs vpc
	VpcIDs []string `json:"vpc_ids" validate:"omitempty"`
	// SubnetIDs 子网
	SubnetIDs []string `json:"subnet_ids" validate:"omitempty"`

	// Memo 备注字段
	Memo *string `json:"memo" validate:"omitempty"`
	// Status 资源状态
	// 取值范围: creating:创建中; mounting:挂载中;create_failed:创建失败;available:可使用;unserviced:停服中;upgrading:升级中;
	Status string `json:"status" validate:"required"`
	// CloudCreatedTime 云上资源创建时间
	CloudCreatedTime string `json:"cloud_created_time" validate:"omitempty"`
	// Extension 差异字段
	Extension *Extension `json:"extension" validate:"required"`
}

// Validate Cfs create request.
func (req *CfsCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Delete --------------------------

// CfsBatchDeleteReq Cfs delete request.
type CfsBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate Cfs delete request.
func (req *CfsBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

//
//// -------------------------- Update --------------------------
//
//// CfsBatchUpdateReq Cfs batch update req.
//type CfsBatchUpdateReq[Extension corecfs.Extension] struct {
//	Cfss []CfsBatchUpdateWithExtension[Extension] `json:"Cfss" validate:"required"`
//}
//
//// CfsBatchUpdate Cfs batch update.
//type CfsBatchUpdate struct {
//	ID                   string   `json:"id" validate:"required"`
//	Name                 string   `json:"name"`
//	BkBizID              int64    `json:"bk_biz_id" validate:"required"`
//	BkHostID             int64    `json:"bk_host_id" validate:"required"`
//	BkCloudID            *int64   `json:"bk_cloud_id"`
//	CloudVpcIDs          []string `json:"cloud_vpc_ids"`
//	VpcIDs               []string `json:"vpc_ids"`
//	CloudSubnetIDs       []string `json:"cloud_subnet_ids"`
//	SubnetIDs            []string `json:"subnet_ids"`
//	CloudImageID         string   `json:"cloud_image_id"`
//	ImageID              string   `json:"image_id"`
//	Memo                 *string  `json:"memo"`
//	Status               string   `json:"status" validate:"required"`
//	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
//	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
//	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
//	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
//	CloudLaunchedTime    string   `json:"cloud_launched_time"`
//	CloudExpiredTime     string   `json:"cloud_expired_time"`
//	OsName               string   `json:"os_name"`
//	MachineType          string   `json:"machine_type"`
//}

//// CfsBatchUpdateWithExtension Cfs batch update with extension.
//type CfsBatchUpdateWithExtension[Extension corecfs.Extension] struct {
//	CfsBatchUpdate `json:",inline"`
//	Extension      *Extension `json:"extension,omitempty"`
//}
//
//// Validate Cfs update request.
//func (req *CfsBatchUpdateReq[T]) Validate() error {
//	if len(req.Cfss) > constant.BatchOperationMaxLimit {
//		return fmt.Errorf("Cfss count should <= %d", constant.BatchOperationMaxLimit)
//	}
//
//	return validator.Validate.Struct(req)
//}
//
//// CfsCommonInfoBatchUpdateReq define Cfs common info batch update req.
//type CfsCommonInfoBatchUpdateReq struct {
//	Cfss []CfsCommonInfoBatchUpdateData `json:"Cfss" validate:"required"`
//}
//
//// Validate Cfs common info batch update req.
//func (req *CfsCommonInfoBatchUpdateReq) Validate() error {
//	if err := validator.Validate.Struct(req); err != nil {
//		return err
//	}
//
//	if len(req.Cfss) > constant.BatchOperationMaxLimit {
//		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
//	}
//
//	return nil
//}
//
//// CfsCommonInfoBatchUpdateData define Cfs common info batch update data.
//type CfsCommonInfoBatchUpdateData struct {
//	ID        string  `json:"id" validate:"required"`
//	BkBizID   *int64  `json:"bk_biz_id"`
//	BkCloudID *int64  `json:"bk_cloud_id"`
//	BkHostID  *int64  `json:"bk_host_id"`
//	Name      *string `json:"name"`
//	// PrivateIPv4Addresses 内网IP
//	PrivateIPv4Addresses *[]string `json:"private_ipv4_addresses"`
//	PrivateIPv6Addresses *[]string `json:"private_ipv6_addresses"`
//	// PublicIPv6Addresses 公网IP
//	PublicIPv4Addresses *[]string `json:"public_ipv4_addresses"`
//	PublicIPv6Addresses *[]string `json:"public_ipv6_addresses"`
//}

// -------------------------- Get --------------------------

// CfsGetResp define Cfs get resp.
type CfsGetResp[T corecfs.Extension] struct {
	rest.BaseResp `json:",inline"`

	Data *corecfs.Cfs[T] `json:"data"`
}

// -------------------------- List --------------------------

// CfsListReq list req.
type CfsListReq struct {
	Field []string `json:"field" validate:"omitempty"`

	Filter *filter.Expression `json:"filter" validate:"required"`

	Page *core.BasePage `json:"page" validate:"required"`
}

// Validate list request.
func (req *CfsListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CfsListResult define Cfs list result.
type CfsListResult struct {
	Count uint64 `json:"count"`

	Details []*corecfs.BaseCfs `json:"details"`
}

// CfsListResp define list resp.
type CfsListResp struct {
	rest.BaseResp `json:",inline"`

	Data *CfsListResult `json:"data"`
}

// CfsExtListReq list req.
type CfsExtListReq struct {
	Field []string `json:"field" validate:"omitempty"`

	Filter *filter.Expression `json:"filter" validate:"required"`

	Page *core.BasePage `json:"page" validate:"required"`
}

// Validate list request.
func (req *CfsExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CfsExtListResult define Cfs with extension list result.
type CfsExtListResult[T corecfs.Extension] struct {
	Count uint64 `json:"count,omitempty"`

	Details []corecfs.Cfs[T] `json:"details,omitempty"`
}

// CfsExtListResp define list resp. (client resp专用)
type CfsExtListResp[T corecfs.Extension] struct {
	rest.BaseResp `json:",inline"`

	Data *CfsExtListResult[T] `json:"data"`
}
