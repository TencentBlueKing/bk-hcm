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

package tcloud

import (
	typecvm "hcm/pkg/adaptor/types/cvm"
)

var (
	SystemDiskTypeNameMap = map[typecvm.TCloudSystemDiskType]string{
		typecvm.LocalBasic:   "本地硬盘",
		typecvm.LocalSsd:     "本地SSD硬盘",
		typecvm.CloudBasic:   "普通云硬盘",
		typecvm.CloudSsd:     "SSD云硬盘",
		typecvm.CloudPremium: "高性能云硬盘",
		typecvm.CloudBssd:    "通用性SSD云硬盘",
	}
	DataDiskTypeNameMap = map[typecvm.TCloudDataDiskType]string{
		typecvm.LocalBasicDataDiskType:   "本地硬盘",
		typecvm.LocalSsdDataDiskType:     "本地SSD硬盘",
		typecvm.LocalNvmeDataDiskType:    "本地NVME硬盘",
		typecvm.LocalProDataDiskType:     "本地HDD硬盘",
		typecvm.CloudBasicDataDiskType:   "普通云硬盘",
		typecvm.CloudPremiumDataDiskType: "高性能云硬盘",
		typecvm.CloudSsdDataDiskType:     "SSD云硬盘",
		typecvm.CloudHssdDataDiskType:    "增强型SSD云硬盘",
		typecvm.CloudTssdDataDiskType:    "极速型SSD云硬盘",
		typecvm.CloudBssdDataDiskType:    "通用型SSD云硬盘",
	}
	InstanceChargeTypeNameMap = map[typecvm.TCloudInstanceChargeType]string{
		typecvm.Prepaid:        "包年包月",
		typecvm.PostpaidByHour: "按量计费",
		typecvm.Cdhpaid:        "独享子机",
		typecvm.Spotpaid:       "竞价付费",
		typecvm.Cdcpaid:        "专用集群付费",
	}
	// InternetChargeTypeNameMap TCloud cvm internet charge type
	InternetChargeTypeNameMap = map[typecvm.TCloudInternetChargeType]string{
		typecvm.TCloudInternetBandwidthPrepaid:        "预付费按带宽结算",
		typecvm.TCloudInternetBandwidthPackage:        "带宽包用户",
		typecvm.TCloudInternetTrafficPostpaidByHour:   "流量按小时后付费",
		typecvm.TCloudInternetBandwidthPostpaidByHour: "带宽按小时后付费",
	}
)
