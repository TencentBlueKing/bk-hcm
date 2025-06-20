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

// Package constant 存放CVM的常量
package constant

// CVM的常量
const (
	// IdleMachine CVM模块-空闲机
	IdleMachine = "空闲机"
	// CCIdleMachine CVM模块-CC_空闲机
	CCIdleMachine = "CC_空闲机"
	// IdleMachineModuleName CVM模块-空闲机模块
	IdleMachineModuleName = "空闲机模块"
	// ResetingSrvStatus CC运营状态-重装中
	ResetingSrvStatus = "重装中"
	// CvmBatchTaskRetryDelayMinMS CVM-批量任务默认重试最小延迟时间
	CvmBatchTaskRetryDelayMinMS = 1000
	// CvmBatchTaskRetryDelayMaxMS CVM-批量任务默认重试最大延迟时间
	CvmBatchTaskRetryDelayMaxMS = 5000
	// UnBindBkHostID defines default value for unbind cvm's host id.
	UnBindBkHostID int64 = -1
)
