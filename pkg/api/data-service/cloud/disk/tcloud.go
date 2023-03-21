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

package disk

// TCloudDiskExtensionCreateReq ...
type TCloudDiskExtensionCreateReq struct {
	DiskChargeType     string                   `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid  *TCloudDiskChargePrepaid `json:"disk_charge_prepaid,omitempty"`
	Encrypted          *bool                    `json:"encrypted,omitempty"`
	Attached           *bool                    `json:"attached,omitempty"`
	DiskUsage          *string                  `json:"disk_usage,omitempty"`
	InstanceId         *string                  `json:"instance_id,omitempty"`
	InstanceType       *string                  `json:"instance_type,omitempty"`
	DeleteWithInstance *bool                    `json:"delete_with_instance,omitempty"`
	DeadlineTime       *string                  `json:"deadline_time,omitempty"`
}

// TCloudDiskExtensionResult ...
type TCloudDiskExtensionResult struct {
	DiskChargeType     string                   `json:"disk_charge_type"`
	DiskChargePrepaid  *TCloudDiskChargePrepaid `json:"disk_charge_prepaid,omitempty"`
	Encrypted          *bool                    `json:"encrypted,omitempty"`
	Attached           *bool                    `json:"attached,omitempty"`
	DiskUsage          *string                  `json:"disk_usage,omitempty"`
	InstanceId         *string                  `json:"instance_id,omitempty"`
	InstanceType       *string                  `json:"instance_type,omitempty"`
	DeleteWithInstance *bool                    `json:"delete_with_instance,omitempty"`
	DeadlineTime       *string                  `json:"deadline_time,omitempty"`
}

// TCloudDiskChargePrepaid ...
type TCloudDiskChargePrepaid struct {
	Period    *int64  `json:"period"`
	RenewFlag *string `json:"renew_flag"`
}

// TCloudDiskExtensionUpdateReq ...
// 根据情况增加 omitempty tag, 因为会调用 json.UpdateMerge 完成字段合并
type TCloudDiskExtensionUpdateReq struct {
	DiskChargeType     string                   `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid  *TCloudDiskChargePrepaid `json:"disk_charge_prepaid,omitempty"`
	Encrypted          *bool                    `json:"encrypted,omitempty"`
	Attached           *bool                    `json:"attached,omitempty"`
	DiskUsage          *string                  `json:"disk_usage,omitempty"`
	InstanceId         *string                  `json:"InstanceId,omitempty"`
	InstanceType       *string                  `json:"InstanceType,omitempty"`
	DeleteWithInstance *bool                    `json:"delete_with_instance,omitempty"`
	DeadlineTime       *string                  `json:"deadline_time,omitempty"`
}
