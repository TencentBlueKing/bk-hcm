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

package cloudserver

import (
	"errors"
	"fmt"

	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// NetworkInterfaceListResult defines list network interface result.
type NetworkInterfaceListResult struct {
	Count   uint64                        `json:"count"`
	Details []coreni.BaseNetworkInterface `json:"details"`
}

// NetworkInterfaceAssociateListResult defines list network interface associate result.
type NetworkInterfaceAssociateListResult struct {
	Count   uint64                             `json:"count"`
	Details []coreni.NetworkInterfaceAssociate `json:"details"`
}

// AssignNetworkInterfaceToBizReq define assign network interface to biz req.
type AssignNetworkInterfaceToBizReq struct {
	BkBizID             int64    `json:"bk_biz_id" validate:"required"`
	NetworkInterfaceIDs []string `json:"network_interface_ids" validate:"required"`
}

// Validate assign network interface to biz request.
func (req *AssignNetworkInterfaceToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.NetworkInterfaceIDs) == 0 {
		return errors.New("network_interface ids is required")
	}

	if len(req.NetworkInterfaceIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("network_interface ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Get --------------------------

// NetworkInterfaceDetail define network interface detail
type NetworkInterfaceDetail[Extension coreni.NetworkInterfaceExtension] struct {
	coreni.BaseNetworkInterface `json:",inline"`
	CvmID                       string     `json:"cvm_id"`
	Extension                   *Extension `json:"extension"`
}
