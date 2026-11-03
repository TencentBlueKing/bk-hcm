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

import "fmt"

// OsType define os type
type OsType string

const (
	LinuxOsType   OsType = "Linux"
	WindowsOsType OsType = "Windows"
	OtherOsType   OsType = "Other"
)

// TCloudImageType is the Tencent Cloud API image type (cloud-side only).
type TCloudImageType string

const (
	// TCloudPrivateImage 私有镜像 (本账户创建的镜像)
	TCloudPrivateImage TCloudImageType = "PRIVATE_IMAGE"
	// TCloudPublicImage 公共镜像 (腾讯云官方镜像)
	TCloudPublicImage TCloudImageType = "PUBLIC_IMAGE"
	// TCloudSharedImage 共享镜像(其他账户共享给本账户的镜像)
	TCloudSharedImage TCloudImageType = "SHARED_IMAGE"
)

// ImageTypeNormalized is the unified image type stored locally and exposed by APIs.
type ImageTypeNormalized string

const (
	// ImageTypePublic is the normalized public image type.
	ImageTypePublic ImageTypeNormalized = "public"
	// ImageTypePrivate is the normalized private image type.
	ImageTypePrivate ImageTypeNormalized = "private"
	// ImageTypeShared is the normalized shared image type.
	ImageTypeShared ImageTypeNormalized = "shared"
)

// NormalizeImageType maps cloud-side type values to unified local format for storage.
// Unknown values are returned as-is (fallback).
func NormalizeImageType(rawType string) string {
	switch rawType {
	case string(TCloudPublicImage), string(ImageTypePublic):
		return string(ImageTypePublic)
	case string(TCloudPrivateImage), string(ImageTypePrivate):
		return string(ImageTypePrivate)
	case string(TCloudSharedImage), string(ImageTypeShared):
		return string(ImageTypeShared)
	default:
		return rawType
	}
}

// Validate validates unified type used by local storage / APIs.
func (n ImageTypeNormalized) Validate() error {
	switch n {
	case ImageTypePublic, ImageTypePrivate, ImageTypeShared:
		return nil
	default:
		return fmt.Errorf("unsupported image type: %s", n)
	}
}
