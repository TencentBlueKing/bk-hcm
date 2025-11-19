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

package constant

import "time"

// 自研账号需要使用内部域名
const (
	// InternalVpcEndpoint vpc 内部域名
	InternalVpcEndpoint = "vpc.internal.tencentcloudapi.com"
	// InternalCvmEndpoint cvm 内部域名
	InternalCvmEndpoint = "cvm.internal.tencentcloudapi.com"
	// InternalCbsEndpoint cbs 内部域名
	InternalCbsEndpoint = "cbs.internal.tencentcloudapi.com"
	// InternalCamEndpoint cam 内部域名
	InternalCamEndpoint = "cam.internal.tencentcloudapi.com"
	// InternalBillingEndpoint 账单内部域名
	InternalBillingEndpoint = "billing.internal.tencentcloudapi.com"
	// InternalCertEndpoint 证书内部域名
	InternalCertEndpoint = "ssl.internal.tencentcloudapi.com"
	// InternalClbEndpoint 负载均衡内部域名
	InternalClbEndpoint = "clb.internal.tencentcloudapi.com"
	// InternalTagEndpoint 标签内部域名
	InternalTagEndpoint = "tag.internal.tencentcloudapi.com"

	// InternalBPaasEndpoint bpaas 内部域名
	InternalBPaasEndpoint = "bpaas.internal.tencentcloudapi.com"
	// InternalCosEndpoint cos内部域名
	InternalCosEndpoint = "https://service.cos-internal.tencentcos.cn"
	// InternalCosEndpointWithNameAndRegion cos带存储桶和地域的内部域名
	InternalCosEndpointWithNameAndRegion = "https://%s.cos-internal.%s.tencentcos.cn"

	// IEGDeptName 互动娱乐事业部
	IEGDeptName = "互动娱乐事业部"
)

// 等待时间
const (
	// IntervalWaitTaskStart 任务启动前的等待时间
	IntervalWaitTaskStart = time.Second * 5

	// IntervalWaitResourceSync 资源变更后，考虑云上存在的延迟，允许同步等待一段时间
	IntervalWaitResourceSync = time.Second * 3
)

const (
	// DefaultTechnicalClass 预测转移额度-默认的技术类型
	DefaultTechnicalClass = "DEFAULT"
)

const (
	// GpuInstanceClass 实例族-GPU型
	GpuInstanceClass = "GPU"
	// GPUDeviceTypeChargeMonth GPU机型-计费模式-72个月
	GPUDeviceTypeChargeMonth uint = 72
)

// TgwGroupNameZiyan Tgw独占集群标签
// TGW：Tencent GateWay， 是负载均衡 （Load Balancer，LB）， 在公司内部简称TGW了；
// 原理TGW包含了TGW，伪7层TGW（定制了tgw_forward），伪7层已经2年多不维护2020年全面下线，可以使用STGW替代接入（TGW+nginx+加解密）；
// CLB：负载均衡（Cloud Load Balancer，CLB），是腾讯云产品化的名称，CLB 4层就对应自研的TGW， CLB 7层对应自研的STGW；
const TgwGroupNameZiyan = "ziyan"

const (
	// DataDiskTotalNum 数据盘总数量
	DataDiskTotalNum = 20
	// DataDiskMinSize 最小数据盘大小，单位GB
	DataDiskMinSize = 10
	// DataDiskMaxSize 最大数据盘大小，单位GB
	DataDiskMaxSize = 32000
	// DataDiskMultiple 数据盘大小倍数
	DataDiskMultiple = 10
	// SystemDiskMinSize 最小数据盘大小，单位GB
	SystemDiskMinSize = 50
	// SystemDiskMaxSize 最大数据盘大小，单位GB
	SystemDiskMaxSize = 1000
	// SystemDiskMultiple 数据盘大小倍数
	SystemDiskMultiple = 50
)

const (
	// ResourcePlanTransferKey this is the config type for resource plan transfer.
	ResourcePlanTransferKey = "resource_plan_transfer"
	// TransferQuotaKey this is the config key for transfer quota.
	TransferQuotaKey = "quota"
	// TransferAuditQuotaKey this is the config key for transfer audit quota.
	TransferAuditQuotaKey = "audit_quota"
)

// AdminAuditStepName 管理员审批阶段名称
const AdminAuditStepName = "管理员审批"

// AdminHandler 管理员用户列表，兜底
const AdminHandler = "dommyzhang;forestchen"

// ResPlanItsmAuditSkip 资源预测itsm免审标志
const ResPlanItsmAuditSkip string = "skip"
