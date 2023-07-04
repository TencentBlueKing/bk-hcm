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

import (
	"fmt"
)

// ApplicationType 申请单类型
type ApplicationType string

// Validate the ApplicationType is valid or not
func (a ApplicationType) Validate() error {
	switch a {
	case AddAccount:
	case CreateCvm:
	case CreateVpc:
	case CreateDisk:
	default:
		return fmt.Errorf("unsupported application type: %s", a)
	}

	return nil
}

const (
	// AddAccount 新增账号
	AddAccount ApplicationType = "add_account"
	// CreateCvm 创建虚拟机
	CreateCvm ApplicationType = "create_cvm"
	// CreateVpc 创建VPC
	CreateVpc ApplicationType = "create_vpc"
	// CreateDisk 创建云盘
	CreateDisk ApplicationType = "create_disk"
)

type ApplicationStatus string

const (
	// Pending 审批中
	Pending ApplicationStatus = "pending"
	// Pass 审批通过
	Pass ApplicationStatus = "pass"
	// Rejected 审批驳回
	Rejected ApplicationStatus = "rejected"
	// Cancelled 单据撤销
	Cancelled ApplicationStatus = "cancelled"
	// Delivering 单据交付中
	Delivering ApplicationStatus = "delivering"
	// Completed 单据完成
	Completed ApplicationStatus = "completed"
	// DeliverPartial 部分交付
	DeliverPartial = "deliver_partial"
	// DeliverError 单据交付异常
	DeliverError ApplicationStatus = "deliver_error"
)
