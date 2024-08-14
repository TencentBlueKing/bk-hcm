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

	case CreateSecurityGroupRule:
	case UpdateSecurityGroupRule:
	case DeleteSecurityGroupRule:

	case CreateSecurityGroup:
	case UpdateSecurityGroup:
	case DeleteSecurityGroup:

	case AssociateSecurityGroup:
	case DisassociateSecurityGroup:

	case CreateMainAccount:
	case UpdateMainAccount:

	case CreateLoadBalancer:
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
	// CreateMainAccount 创建主账号/二级账号
	CreateMainAccount ApplicationType = "create_main_account"
	// UpdateMainAccount 修改主账号/二级账号
	UpdateMainAccount ApplicationType = "update_main_account"
	// CreateLoadBalancer 创建负载均衡
	CreateLoadBalancer ApplicationType = "create_load_balancer"

	// CreateSecurityGroup 创建安全组
	CreateSecurityGroup ApplicationType = "create_security_group"
	// UpdateSecurityGroup 更新安全组
	UpdateSecurityGroup ApplicationType = "update_security_group"
	// DeleteSecurityGroup 删除安全组
	DeleteSecurityGroup ApplicationType = "delete_security_group"
	// AssociateSecurityGroup 安全组关联资源
	AssociateSecurityGroup ApplicationType = "associate_security_group"
	// DisassociateSecurityGroup 安全组资源解关联
	DisassociateSecurityGroup ApplicationType = "disassociate_security_group"

	// CreateSecurityGroupRule 创建安全组规则
	CreateSecurityGroupRule ApplicationType = "create_security_group_rule"
	// UpdateSecurityGroupRule 更新安全组规则
	UpdateSecurityGroupRule ApplicationType = "update_security_group_rule"
	// DeleteSecurityGroupRule 删除安全组规则
	DeleteSecurityGroupRule ApplicationType = "delete_security_group_rule"
)

// ApplicationStatus 单据状态
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

// ApplicationSource 单据来源
type ApplicationSource string

const (
	// ApplicationSourceITSM itsm 单据
	ApplicationSourceITSM ApplicationSource = "itsm"
)
