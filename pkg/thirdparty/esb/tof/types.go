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

package tof

// GetStaffInfoResp TOF接口响应结构
type GetStaffInfoResp struct {
	Result  bool       `json:"result"`
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Data    *StaffInfo `json:"data"`
}

// StaffInfo 员工信息
type StaffInfo struct {
	StatusName   string `json:"StatusName"`   // 状态名称：在职
	StatusId     string `json:"StatusId"`     // 状态ID
	TypeId       string `json:"TypeId"`       // 类型ID
	PostName     string `json:"PostName"`     // 岗位名称
	WorkDeptId   string `json:"WorkDeptId"`   // 工作部门ID
	OfficialId   string `json:"OfficialId"`   // 职级ID
	TypeName     string `json:"TypeName"`     // 类型名称：正式
	WorkDeptName string `json:"WorkDeptName"` // 工作部门名称
	OfficialName string `json:"OfficialName"` // 职级名称：普通员工
	RTX          string `json:"RTX"`          // RTX账号
	LoginName    string `json:"LoginName"`    // 员工英文名（login_name）
	Gender       string `json:"Gender"`       // 性别
	Enabled      string `json:"Enabled"`      // 是否启用
	GroupName    string `json:"GroupName"`    // 组名称
	DepartmentId string `json:"DepartmentId"` // 部门ID
	EnglishName  string `json:"EnglishName"`  // 英文名
	ID           string `json:"ID"`           // 员工ID
	ChineseName  string `json:"ChineseName"`  // 中文名
	FullName     string `json:"FullName"`     // 全名
	PostId       string `json:"PostId"`       // 岗位ID
	GroupId      string `json:"GroupId"`      // 组ID
}
