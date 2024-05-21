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

import "errors"

// RuleType 负载均衡类型
type RuleType string

// 负载均衡类型
const (
	// Layer4RuleType 4层负载均衡
	Layer4RuleType RuleType = "layer_4"
	// Layer7RuleType 7层负载均衡
	Layer7RuleType RuleType = "layer_7"
)

// TargetGroupType 目标组类型
type TargetGroupType string

// 目标组类型
const (
	// LocalTargetGroupType 本地目标组类型
	LocalTargetGroupType TargetGroupType = "local"
	// CloudTargetGroupType 云端目标组类型
	CloudTargetGroupType TargetGroupType = "cloud"
)

// BindingStatus 绑定状态
type BindingStatus string

// 目标组类型
const (
	// SuccessBindingStatus 绑定状态-成功
	SuccessBindingStatus BindingStatus = "success"
	// BindingBindingStatus 绑定中
	BindingBindingStatus BindingStatus = "binding"
	// FailedBindingStatus 绑定失败
	FailedBindingStatus BindingStatus = "failed"
	// PartialFailedBindingStatus 部分失败
	PartialFailedBindingStatus BindingStatus = "partial_failed"
)

// ProtocolType 协议类型
type ProtocolType string

// 目标组类型
const (
	// HttpProtocol 协议类型-HTTP
	HttpProtocol ProtocolType = "HTTP"
	// HttpsProtocol 协议类型-HTTPS
	HttpsProtocol  ProtocolType = "HTTPS"
	TcpProtocol    ProtocolType = "TCP"
	UdpProtocol    ProtocolType = "UDP"
	TcpSslProtocol ProtocolType = "TCP_SSL"
	QuicProtocol   ProtocolType = "QUIC"
)

// IsLayer7Protocol 是否7层规则类型
func (p ProtocolType) IsLayer7Protocol() bool {
	return p == HttpProtocol || p == HttpsProtocol
}

// SniType SNI类型
type SniType int64

// 目标组类型
const (
	// SniTypeClose SNI类型-关闭
	SniTypeClose SniType = 0
	// SniTypeOpen SNI类型-开启
	SniTypeOpen SniType = 1
)

// Validate SNI类型是否合法
func (s SniType) Validate() error {
	if s != SniTypeClose && s != SniTypeOpen {
		return errors.New("sni_switch is illegal")
	}
	return nil
}

// TaskType 任务类型
type TaskType string

// 任务类型
const (
	// AddRSTaskType 任务类型-添加RS
	AddRSTaskType = TaskType(FlowTargetGroupAddRS)
	// RemoveRSTaskType 任务类型-移除RS
	RemoveRSTaskType = TaskType(FlowTargetGroupRemoveRS)
	// ModifyPortTaskType 任务类型-修改端口
	ModifyPortTaskType = TaskType(FlowTargetGroupModifyPort)
	// ModifyWeightTaskType 任务类型-修改权重
	ModifyWeightTaskType = TaskType(FlowTargetGroupModifyWeight)
	// ApplyTargetGroupType 任务类型应用目标组到监听器/规则上
	ApplyTargetGroupType = TaskType(FlowApplyTargetGroupToListenerRule)
)

// InstType 实例类型
type InstType string

// 任务类型
const (
	// CvmInstType 实例类型-云服务器
	CvmInstType InstType = "CVM"
	// EniInstType 实例类型-弹性网卡
	EniInstType InstType = "ENI"
)

// ResFlowStatus 资源跟Flow的状态类型
type ResFlowStatus string

// 资源跟Flow的状态类型
const (
	// ExecutingResFlowStatus 资源跟Flow的状态类型-执行中
	ExecutingResFlowStatus ResFlowStatus = "executing"
	// TimeoutResFlowStatus 资源跟Flow的状态类型-超时
	TimeoutResFlowStatus ResFlowStatus = "timeout"
	// SuccessResFlowStatus 资源跟Flow的状态类型-成功
	SuccessResFlowStatus = ResFlowStatus(FlowSuccess)
	// CancelResFlowStatus 资源跟Flow的状态类型-取消
	CancelResFlowStatus = ResFlowStatus(FlowCancel)
)

// IsEnd 是否为终态
func (r ResFlowStatus) IsEnd() bool {
	return r == CancelResFlowStatus || r == TimeoutResFlowStatus || r == SuccessResFlowStatus
}
