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

package eip

import (
	"hcm/pkg/criteria/validator"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
)

// HuaWeiEipListOption ...
type HuaWeiEipListOption struct {
	Region   string   `json:"region" validate:"required"`
	Limit    *int32   `json:"limit" validate:"omitempty"`
	Marker   *string  `json:"marker" validate:"omitempty"`
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	Ips      []string `json:"ips" validate:"omitempty"`
}

// Validate ...
func (o *HuaWeiEipListOption) Validate() error {
	return validator.Validate.Struct(o)
}

// HuaWeiEipListResult ...
type HuaWeiEipListResult struct {
	Details []*HuaWeiEip
}

// HuaWeiEip ...
type HuaWeiEip struct {
	CloudID       string
	Name          *string
	Region        string
	InstanceId    *string
	Status        *string
	PublicIp      *string
	PrivateIp     *string
	PortID        *string
	BandwidthId   *string
	BandwidthName *string
	BandwidthSize *int32
}

// HuaWeiEipDeleteOption ...
type HuaWeiEipDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDeletePublicipRequest ...
func (opt *HuaWeiEipDeleteOption) ToDeletePublicipRequest() (*model.DeletePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &model.DeletePublicipRequest{PublicipId: opt.CloudID}, nil
}

// HuaWeiEipAssociateOption ...
type HuaWeiEipAssociateOption struct {
	CloudEipID              string `json:"cloud_eip_id" validate:"required"`
	CloudNetworkInterfaceID string `json:"cloud_network_interface_id" validate:"required"`
	Region                  string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToUpdatePublicipRequest ...
func (opt *HuaWeiEipAssociateOption) ToUpdatePublicipRequest() (*model.UpdatePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.UpdatePublicipRequest{PublicipId: opt.CloudEipID}
	req.Body = &model.UpdatePublicipsRequestBody{
		Publicip: &model.UpdatePublicipOption{PortId: &opt.CloudNetworkInterfaceID},
	}
	return req, nil
}

// HuaWeiEipDisassociateOption ...
type HuaWeiEipDisassociateOption struct {
	CloudEipID string `json:"cloud_eip_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToUpdatePublicipRequest ...
func (opt *HuaWeiEipDisassociateOption) ToUpdatePublicipRequest() (*model.UpdatePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.UpdatePublicipRequest{PublicipId: opt.CloudEipID, Body: &model.UpdatePublicipsRequestBody{}}
	return req, nil
}
