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

package enumor

import "fmt"

// ResPlanDemandGpuOrderStatus GPU需求提报主单状态
type ResPlanDemandGpuOrderStatus string

const (
	// ResPlanDemandGpuOrderStatusInit 待评审
	ResPlanDemandGpuOrderStatusInit ResPlanDemandGpuOrderStatus = "INIT"
	// ResPlanDemandGpuOrderStatusPending 评审中
	ResPlanDemandGpuOrderStatusPending ResPlanDemandGpuOrderStatus = "PENDING"
	// ResPlanDemandGpuOrderStatusDone 已评审
	ResPlanDemandGpuOrderStatusDone ResPlanDemandGpuOrderStatus = "DONE"
	// ResPlanDemandGpuOrderStatusReject 部分已驳回
	ResPlanDemandGpuOrderStatusReject ResPlanDemandGpuOrderStatus = "REJECT"
	// ResPlanDemandGpuOrderStatusRejectAll 全部已驳回
	ResPlanDemandGpuOrderStatusRejectAll ResPlanDemandGpuOrderStatus = "REJECT_ALL"
	// ResPlanDemandGpuOrderStatusTerminate 已终止
	ResPlanDemandGpuOrderStatusTerminate ResPlanDemandGpuOrderStatus = "TERMINATE"
)

// Validate ResPlanDemandGpuOrderStatus.
func (s ResPlanDemandGpuOrderStatus) Validate() error {
	switch s {
	case ResPlanDemandGpuOrderStatusInit,
		ResPlanDemandGpuOrderStatusPending,
		ResPlanDemandGpuOrderStatusDone,
		ResPlanDemandGpuOrderStatusReject,
		ResPlanDemandGpuOrderStatusRejectAll,
		ResPlanDemandGpuOrderStatusTerminate:
	default:
		return fmt.Errorf("unsupported res plan demand gpu order status: %s", s)
	}

	return nil
}
