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

// Vendor defines the cloud type where the hybrid cloud service is supported.
type Vendor string

// Validate the vendor is valid or not
func (v Vendor) Validate() error {
	switch v {
	case TCloud:
	case Aws:
	case Gcp:
	case Azure:
	case HuaWei:
	default:
		return fmt.Errorf("unsupported cloud vendor: %s", v)
	}

	return nil
}

const (
	// TCloud is tencent cloud
	TCloud Vendor = "tcloud"
	// Aws is amazon cloud
	Aws Vendor = "aws"
	// Gcp is the Google Cloud Platform
	Gcp Vendor = "gcp"
	// Azure is microsoft azure cloud.
	Azure Vendor = "azure"
	// HuaWei is hua wei cloud.
	HuaWei Vendor = "huawei"
)
