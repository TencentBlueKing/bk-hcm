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

package cmdb

import (
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// AddCloudHostToBizReq defines add cmdb cloud host to biz http request.
type AddCloudHostToBizReq[T cvm.Extension] struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	BizID  int64         `json:"bk_biz_id" validate:"min=1"`
	Hosts  []cvm.Cvm[T]  `json:"hosts" validate:"min=1,max=100"`
}

// Validate AddCloudHostToBizReq.
func (c *AddCloudHostToBizReq[T]) Validate() error {
	if err := c.Vendor.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(c)
}

// AddBaseCloudHostToBizReq defines add cmdb cloud host basic info to biz http request.
type AddBaseCloudHostToBizReq struct {
	BizID int64         `json:"bk_biz_id" validate:"required"`
	Hosts []cvm.BaseCvm `json:"hosts" validate:"min=1,max=100"`
}

// Validate AddBaseCloudHostToBizReq.
func (c *AddBaseCloudHostToBizReq) Validate() error {
	return validator.Validate.Struct(c)
}

// DeleteCloudHostFromBizReq is cmdb delete cloud host from biz http request.
type DeleteCloudHostFromBizReq struct {
	BizID          int64                      `json:"bk_biz_id" validate:"min=1"`
	VendorCloudIDs map[enumor.Vendor][]string `json:"vendor_cloud_ids" validate:"required"`
}

// Validate DeleteCloudHostFromBizReq.
func (c *DeleteCloudHostFromBizReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	idLen := 0
	for _, ids := range c.VendorCloudIDs {
		idLen += len(ids)
	}
	if idLen <= 0 || idLen >= 500 {
		return errf.New(errf.InvalidParameter, "delete cmdb cloud ids length is invalid")
	}

	return nil
}
