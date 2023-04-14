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
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// NetworkInterfaceCvmRelBatchCreateReq ...
type NetworkInterfaceCvmRelBatchCreateReq struct {
	Rels []NetworkInterfaceCvmRelCreateReq `json:"rels" validate:"required"`
}

// Validate ...
func (req *NetworkInterfaceCvmRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// NetworkInterfaceCvmRelCreateReq ...
type NetworkInterfaceCvmRelCreateReq struct {
	NetworkInterfaceID string `json:"network_interface_id" validate:"required"`
	CvmID              string `json:"cvm_id" validate:"required"`
}

// NetworkInterfaceCvmRelListReq ...
type NetworkInterfaceCvmRelListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *NetworkInterfaceCvmRelListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// NetworkInterfaceCvmRelDeleteReq ...
type NetworkInterfaceCvmRelDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *NetworkInterfaceCvmRelDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// NetworkInterfaceCvmRelListResult ...
type NetworkInterfaceCvmRelListResult struct {
	Count   *uint64                         `json:"count,omitempty"`
	Details []*NetworkInterfaceCvmRelResult `json:"details"`
}

// NetworkInterfaceCvmRelResult ...
type NetworkInterfaceCvmRelResult struct {
	ID                 uint64 `json:"id,omitempty"`
	NetworkInterfaceID string `json:"network_interface_id"`
	CvmID              string `json:"cvm_id,omitempty"`
	Creator            string `json:"creator,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
}

// NetworkInterfaceCvmRelListResp ...
type NetworkInterfaceCvmRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *NetworkInterfaceCvmRelListResult `json:"data"`
}

// NetworkInterfaceCvmRelWithListReq ...
type NetworkInterfaceCvmRelWithListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *NetworkInterfaceCvmRelWithListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// NetworkInterfaceCvmRelWithListResp ...
type NetworkInterfaceCvmRelWithListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*NetworkInterfaceWithCvmID `json:"data"`
}

// NetworkInterfaceWithCvmID ...
type NetworkInterfaceWithCvmID struct {
	coreni.BaseNetworkInterface `json:",inline"`
	CvmID                       string `json:"cvm_id"`
	RelCreator                  string `json:"rel_creator"`
	RelCreatedAt                string `json:"rel_created_at"`
}

// NetworkInterfaceCvmRelWithExtListReq ...
type NetworkInterfaceCvmRelWithExtListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *NetworkInterfaceCvmRelWithExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// NetworkInterfaceCvmRelWithExtListResp ...
type NetworkInterfaceCvmRelWithExtListResp[T coreni.NetworkInterfaceExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          []*NetworkInterfaceExtWithCvmID[T] `json:"data"`
}

// NetworkInterfaceExtWithCvmID ...
type NetworkInterfaceExtWithCvmID[T coreni.NetworkInterfaceExtension] struct {
	coreni.NetworkInterface[T] `json:",inline"`
	CvmID                      string `json:"cvm_id"`
	RelCreator                 string `json:"rel_creator"`
	RelCreatedAt               string `json:"rel_created_at"`
}
