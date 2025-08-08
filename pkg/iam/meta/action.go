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

package meta

// Action 表示 hcm 这一侧的资源类型， 对应的有 client.ActionID 表示 iam 一侧的资源类型
// 两者之间有映射关系，详情见 AdaptAuthOptions
type Action string

// String convert Action to string.
func (a Action) String() string {
	return string(a)
}

const (
	// Create operation's hcm auth action type
	Create Action = "create"
	// Update operation's hcm auth action type
	Update Action = "update"
	// UpdateRRT RecycleReserveTime operation's hcm auth action type
	UpdateRRT Action = "update_rrt"
	// Delete operation's hcm auth action type
	Delete Action = "delete"
	// Find operation's hcm auth action type
	Find Action = "find"
	// KeyAccess access secret key operation's hcm auth action type
	KeyAccess Action = "key_access"
	// Assign cloud resource to biz operation's hcm auth action type
	Assign Action = "assign"
	// Recycle cloud resource from biz operation's hcm auth action type
	Recycle Action = "recycle"
	// Destroy 删除回收站中的资源
	Destroy Action = "destroy"
	// Recover cloud resource from recycle bin operation's hcm auth action type
	Recover Action = "recover"
	// SkipAction means the operation do not need to do authentication, skip auth
	SkipAction Action = "skip"
	// Start operation's hcm auth action type
	Start Action = "start"
	// Stop operation's hcm auth action type
	Stop Action = "stop"
	// Reboot operation's hcm auth action type
	Reboot Action = "reboot"
	// ResetPwd operation's hcm auth action type
	ResetPwd Action = "reset_pwd"
	// ResetSystem operation's hcm auth action type
	ResetSystem Action = "reset_system"
	// Import operation's hcm auth action type
	Import Action = "import"
	// Associate operation's hcm auth action type
	Associate Action = "associate"
	// Disassociate operation's hcm auth action type
	Disassociate Action = "disassociate"
	// Access operation's hcm auth action type
	Access Action = "access"
	// Apply 资源申请权限
	Apply Action = "apply"
)
