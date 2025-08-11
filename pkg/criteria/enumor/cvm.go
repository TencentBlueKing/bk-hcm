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

// CvmResetStatus define cvm reset status
type CvmResetStatus int

// ResetStatus cvm重装状态
const (
	// NormalCvmResetStatus 状态-正常
	NormalCvmResetStatus CvmResetStatus = 0
	// NoOperatorCvmResetStatus 状态-不是主备负责人
	NoOperatorCvmResetStatus CvmResetStatus = 1
	// NoIdleCvmResetStatus 状态-不在空闲机模块
	NoIdleCvmResetStatus CvmResetStatus = 2
)

// 任务类型
const (
	// ResetCvmTaskType 任务类型-CVM重装
	ResetCvmTaskType = TaskType(FlowResetCvm)
	// StartCvmTaskType 任务类型-启动云服务器
	StartCvmTaskType = TaskType(FlowStartCvm)
	// StopCvmTaskType 任务类型-停止云服务器
	StopCvmTaskType = TaskType(FlowStopCvm)
	// RebootCvmTaskType 任务类型-重启云服务器
	RebootCvmTaskType = TaskType(FlowRebootCvm)
)

// CvmOperateStatus define cvm operate status
type CvmOperateStatus int

// OperateStatus cvm 电源操作状态
const (
	// CvmOperateStatusNormal 状态-正常
	CvmOperateStatusNormal CvmOperateStatus = 0
	// CvmOperateStatusNoOperator 状态-不是主备负责人
	CvmOperateStatusNoOperator CvmOperateStatus = 1
	// CvmOperateStatusNoIdle 状态-不在空闲机模块
	CvmOperateStatusNoIdle CvmOperateStatus = 2
	// CvmOperateStatusNoStop 状态-云服务器未处于关机状态
	CvmOperateStatusNoStop CvmOperateStatus = 3
	// CvmOperateStatusNoRunning 状态-云服务器未处于开机状态
	CvmOperateStatusNoRunning CvmOperateStatus = 4
	// CvmOperateStatusPmNoOperate 状态-物理机不支持操作
	CvmOperateStatusPmNoOperate CvmOperateStatus = 5
)

// CvmOperateType define cvm operate type
type CvmOperateType string

const (
	// CvmOperateTypeStart 启动云服务器
	CvmOperateTypeStart = "start"
	// CvmOperateTypeStop 停止云服务器
	CvmOperateTypeStop = "stop"
	// CvmOperateTypeReboot 重启云服务器
	CvmOperateTypeReboot = "reboot"
	// CvmOperateTypeReset 重装云服务器
	CvmOperateTypeReset = "reset"
)

const (

	// TCloudCvmStatusPending 状态-创建中
	TCloudCvmStatusPending = "PENDING"
	// TCloudCvmStatusLaunchFailed 状态-创建失败
	TCloudCvmStatusLaunchFailed = "LAUNCH_FAILED"
	// TCloudCvmStatusRunning 状态-运行中
	TCloudCvmStatusRunning = "RUNNING"
	// TCloudCvmStatusStopped 状态-关机
	TCloudCvmStatusStopped = "STOPPED"
	// TCloudCvmStatusStarting 状态-开机中
	TCloudCvmStatusStarting = "STARTING"
	// TCloudCvmStatusStopping 状态-关机中
	TCloudCvmStatusStopping = "STOPPING"
	// TCloudCvmStatusRebooting 状态-重启中
	TCloudCvmStatusRebooting = "REBOOTING"
	// TCloudCvmStatusShutdown 状态-停止待销毁
	TCloudCvmStatusShutdown = "SHUTDOWN"
	// TCloudCvmStatusTerminating 状态-销毁中
	TCloudCvmStatusTerminating = "TERMINATING"
)

// CvmMatchType cvm匹配类型
type CvmMatchType string

const (
	// AutoMatchCvm 自动匹配cvm
	AutoMatchCvm CvmMatchType = "auto"
	// ManualMatchCvm 手动匹配cvm
	ManualMatchCvm CvmMatchType = "manual"
	// NoMatchCvm 待关联cvm
	NoMatchCvm = "no_match"
)
