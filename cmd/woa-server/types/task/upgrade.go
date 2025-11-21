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

package task

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// CreateUpgradeCrpOrderReq is request for upgrade crp order
type CreateUpgradeCrpOrderReq struct {
	BkBizID        int64
	User           string
	RequireType    enumor.RequireType
	Remark         string           `json:"remark" bson:"remark"`
	UpgradeCvmList []UpgradeCvmItem `json:"upgrade_cvm_list" bson:"upgrade_cvm_list"`
}

// UpgradeCvmItem is upgrade cvm item
type UpgradeCvmItem struct {
	BkHostID           int64  `json:"bk_host_id" bson:"bk_host_id"`
	InstanceID         string `json:"instance_id" bson:"instance_id"`
	TargetInstanceType string `json:"target_instance_type" bson:"target_instance_type"`
}

// Validate whether CreateUpgradeCrpOrderReq is valid
func (req CreateUpgradeCrpOrderReq) Validate() error {
	if len(req.UpgradeCvmList) == 0 {
		return fmt.Errorf("upgrade cvm list is required")
	}

	if len(req.UpgradeCvmList) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("upgrade cvm list length must be less than %d", core.DefaultMaxPageLimit)
	}

	for _, item := range req.UpgradeCvmList {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate whether UpgradeCvmItem is valid
func (i UpgradeCvmItem) Validate() error {
	if i.TargetInstanceType == "" {
		return fmt.Errorf("target_instance_type is required")
	}

	// bk_biz_id 和 instance_id 必须且只能提供其中一个
	if i.BkHostID == 0 && i.InstanceID == "" {
		return fmt.Errorf("bk_host_id or instance_id must be provided")
	}

	if i.BkHostID != 0 && i.InstanceID != "" {
		return fmt.Errorf("bk_host_id and instance_id cannot be provided at the same time")
	}

	return nil
}

// CreateUpgradeCrpOrderResult result of create upgrade crp order
type CreateUpgradeCrpOrderResult struct {
	CRPOrderID string `json:"crp_order_id"`
}
