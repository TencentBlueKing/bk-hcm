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

package coreimage

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseImage define base image.
type BaseImage struct {
	ID            string        `json:"id"`
	Vendor        string        `json:"vendor"`
	CloudID       string        `json:"cloud_id"`
	Name          string        `json:"name"`
	Architecture  string        `json:"architecture"`
	Platform      string        `json:"platform"`
	State         string        `json:"state"`
	Type          string        `json:"type"`
	OsType        enumor.OsType `json:"os_type"`
	core.Revision `json:",inline"`
}

// Extension ...
type Extension interface {
	TCloudExtension | AwsExtension | GcpExtension | HuaWeiExtension | AzureExtension
}

// Image ...
type Image[Ext Extension] struct {
	BaseImage `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (image Image[T]) GetID() string {
	return image.ID
}

// GetCloudID ...
func (image Image[T]) GetCloudID() string {
	return image.CloudID
}

// TCloudExtension ...
type TCloudExtension struct {
	Region      string `json:"region"`
	ImageSource string `json:"image_source"`
	ImageSize   uint64 `json:"image_size"`
}

// AwsExtension ...
type AwsExtension struct {
	Region string `json:"region"`
}

// GcpExtension ...
type GcpExtension struct {
	Region    string `json:"region" validate:"required"`
	SelfLink  string `json:"self_link" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}

// HuaWeiExtension ...
type HuaWeiExtension struct {
	Region string `json:"region"`
}

// AzureExtension ...
type AzureExtension struct {
	Region    string `json:"region" validate:"required"`
	Publisher string `json:"publisher" validate:"required"`
	Offer     string `json:"offer" validate:"required"`
	Sku       string `json:"sku" validate:"required"`
}
