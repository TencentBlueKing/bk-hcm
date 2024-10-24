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

// OsType define os type
type OsType string

const (
	LinuxOsType   OsType = "Linux"
	WindowsOsType OsType = "Windows"
	OtherOsType   OsType = "Other"
)

// TCloudImageTypeImageType 镜像
type TCloudImageType string

const (
	// TCloudPrivateImage 私有镜像 (本账户创建的镜像)
	TCloudPrivateImage TCloudImageType = "PRIVATE_IMAGE"
	// TCloudPublicImage 公共镜像 (腾讯云官方镜像)
	TCloudPublicImage TCloudImageType = "PUBLIC_IMAGE"
	// TCloudSharedImage 共享镜像(其他账户共享给本账户的镜像)
	TCloudSharedImage TCloudImageType = "SHARED_IMAGE"
)
