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

// GlobalConfigType global config type
type GlobalConfigType string

const (
	// GlobalConfigTypeCloudSync 云资源同步相关配置
	GlobalConfigTypeCloudSync GlobalConfigType = "cloud_sync"
	// GlobalConfigTypeGPUMachineType GPU 机型清单配置，config_key 为云厂商，config_value 为机型字符串数组
	GlobalConfigTypeGPUMachineType GlobalConfigType = "gpu_machine_type"
	// GlobalConfigTypeITSM ITSM 流程相关配置
	GlobalConfigTypeITSM GlobalConfigType = "itsm"
)

// GlobalConfigKeyCloudSync cloud sync global config key
type GlobalConfigKeyCloudSync string

const (
	// GlobalConfigKeyCloudSyncBizIDs 云资源同步业务白名单，config_value 为 JSON 对象 {"tenantID": [bizID1, bizID2], ...}
	GlobalConfigKeyCloudSyncBizIDs GlobalConfigKeyCloudSync = "sync_biz_ids"
)

// GlobalConfigKeyGPUMachineType GPU 机型清单配置，config_key 为云厂商
type GlobalConfigKeyGPUMachineType string

const (
	// GlobalConfigKeyHuaweiGPUPrefix 华为云GPU机型前缀
	GlobalConfigKeyHuaweiGPUPrefix GlobalConfigKeyGPUMachineType = "huawei_gpu_machine_prefix"
	// GlobalConfigKeyTcloudGPUPrefix 腾讯云 GPU 机型清单配置，config_key 为云厂商
	GlobalConfigKeyTcloudGPUPrefix GlobalConfigKeyGPUMachineType = "tcloud_gpu_machine_prefix"
	// GlobalConfigKeyGcpGPUPrefix 谷歌云 GPU 机型清单配置，config_key 为云厂商
	GlobalConfigKeyGcpGPUPrefix GlobalConfigKeyGPUMachineType = "gcp_gpu_machine_prefix"
	// GlobalConfigKeyAws 亚马逊云 GPU 机型清单配置，config_key 为云厂商
	GlobalConfigKeyAws GlobalConfigKeyGPUMachineType = "aws"
	// GlobalConfigKeyAzureGPUPrefix 微软 Azure 云 GPU 机型清单配置，config_key 为云厂商
	GlobalConfigKeyAzureGPUPrefix GlobalConfigKeyGPUMachineType = "azure_gpu_machine_prefix"
)

// GlobalConfigKeyITSM ITSM global config key
type GlobalConfigKeyITSM string

const (
	// GlobalConfigKeyItsmMigrateVersionPrefix ITSM 流程注册进度，config_key 格式为 "itsm_migrate_version_{tenantID}"
	GlobalConfigKeyItsmMigrateVersionPrefix GlobalConfigKeyITSM = "itsm_migrate_version"
)
