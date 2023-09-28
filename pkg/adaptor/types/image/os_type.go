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

package image

import "hcm/pkg/criteria/enumor"

// GetOsTypeByPlatform 通过镜像平台获取所属操作系统类型
func GetOsTypeByPlatform(vendor enumor.Vendor, platform string) enumor.OsType {
	switch vendor {
	case enumor.TCloud:
		return getTCloudOsTypeByPlatform(platform)
	case enumor.Aws:
		return getAwsOsTypeByPlatform(platform)
	case enumor.HuaWei:
		return getHuaWeiOsTypeByPlatform(platform)
	case enumor.Gcp:
		return getGcpOsTypeByPlatform(platform)
	case enumor.Azure:
		return getAzureOsTypeByPlatform(platform)
	default:
		return enumor.OtherOsType
	}
}

func getAzureOsTypeByPlatform(platform string) enumor.OsType {
	switch platform {
	case "CentOS":
		return enumor.LinuxOsType
	case "WindowsServer":
		return enumor.WindowsOsType
	default:
		return enumor.OtherOsType
	}
}

func getGcpOsTypeByPlatform(platform string) enumor.OsType {
	switch platform {
	case "Centos":
		return enumor.LinuxOsType
	case "Windows":
		return enumor.WindowsOsType
	default:
		return enumor.OtherOsType
	}
}

func getHuaWeiOsTypeByPlatform(platform string) enumor.OsType {
	switch platform {
	case "CentOS":
		return enumor.LinuxOsType
	case "Windows":
		return enumor.WindowsOsType
	default:
		return enumor.OtherOsType
	}
}

func getAwsOsTypeByPlatform(platform string) enumor.OsType {
	switch platform {
	case "Linux/UNIX", "SUSE Linux", "Ubuntu Pro Linux", "Red Hat Enterprise Linux with High Availability",
		"Red Hat Enterprise Linux", "Red Hat Enterprise Linux with SQL Server Standard":
		return enumor.LinuxOsType
	case "Windows", "Windows with SQL Server Enterprise":
		return enumor.WindowsOsType
	case "SQL Server Enterprise", "SQL Server Standard":
		return enumor.OtherOsType
	default:
		return enumor.OtherOsType
	}
}

func getTCloudOsTypeByPlatform(platform string) enumor.OsType {
	switch platform {
	case "TencentOS":
		return enumor.LinuxOsType
	case "Windows":
		return enumor.WindowsOsType
	default:
		return enumor.OtherOsType
	}
}
