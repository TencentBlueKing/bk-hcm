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

const (
	// UnBindBkHostID defines default value for unbind cvm's host id.
	UnBindBkHostID int64 = -1
)

// Monitor namespace constants
const (
	// TCloudCvmNamespace 腾讯云CVM监控命名空间
	TCloudCvmNamespace = "QCE/CVM"
	// TCloudCvmInstanceIDKey 腾讯云CVM实例ID key
	TCloudCvmInstanceIDKey = "InstanceId"
	// HuaWeiCvmNamespace 华为云CVM监控命名空间
	HuaWeiCvmNamespace = "SYS.ECS"
	// HuaWeiVpcNamespace 华为云VPC监控命名空间
	HuaWeiVpcNamespace = "SYS.VPC"
	// HuaWeiCvmInstanceIDKey 华为云CVM实例ID key
	HuaWeiCvmInstanceIDKey = "instance_id"
	// HuaWeiPublicIPIDKey 华为云弹性公网IP ID key
	HuaWeiPublicIPIDKey = "publicip_id"
	// HuaWeiMonitorDefaultFilter 华为云监控默认聚合方式
	HuaWeiMonitorDefaultFilter = "average"
	// AwsCvmNamespace AWS CVM 监控命名空间
	AwsCvmNamespace = "AWS/EC2"
	// AwsCvmInstanceIDKey AWS CVM 实例ID key
	AwsCvmInstanceIDKey = "InstanceId"
	// AwsMetricNetworkIn AWS 入流量指标
	AwsMetricNetworkIn = "NetworkIn"
	// AwsMetricNetworkOut AWS 出流量指标
	AwsMetricNetworkOut = "NetworkOut"
	// MetricLanIntraffic 内网入带宽
	MetricLanIntraffic = "LanIntraffic"
	// MetricWanIntraffic 外网入带宽
	MetricWanIntraffic = "WanIntraffic"
	// MetricLanOuttraffic 内网出带宽
	MetricLanOuttraffic = "LanOuttraffic"
	// MetricWanOuttraffic 外网出带宽
	MetricWanOuttraffic = "WanOuttraffic"
	// AzureCvmInstanceIDKey Azure CVM 实例维度 key（对应 Azure metadata name: Microsoft.ResourceId）
	AzureCvmInstanceIDKey = "Microsoft.ResourceId"
	// AzureMetricNamespaceDefault Azure 默认监控命名空间
	AzureMetricNamespaceDefault = "Microsoft.Compute/virtualMachines"
	// AzureMetricNetworkInTotal Azure 入流量总计指标
	AzureMetricNetworkInTotal = "Network In Total"
	// AzureMetricNetworkOutTotal Azure 出流量总计指标
	AzureMetricNetworkOutTotal = "Network Out Total"
	// AzureMonitorAggregationTotal Azure 监控 Total 聚合
	AzureMonitorAggregationTotal = "Total"
	// AzureMonitorAggregationAverage Azure 监控 Average 聚合
	AzureMonitorAggregationAverage = "Average"
)
