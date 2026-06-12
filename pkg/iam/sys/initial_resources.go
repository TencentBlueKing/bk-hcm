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

package sys

import (
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// ResourceTypeIDMap resource type map.
var ResourceTypeIDMap = map[iam.TypeID]string{
	Account:              "账号",
	Biz:                  "业务",
	CloudSelectionScheme: "方案",
	MainAccount:          "二级账号",
	BillCloudVendor:      "账单云厂商",
}

// GenerateStaticResourceTypes generate all the static resource types to register to IAM.
func GenerateStaticResourceTypes() []iam.ResourceType {
	resourceTypeList := make([]iam.ResourceType, 0)

	// add account resources
	resourceTypeList = append(resourceTypeList, genAccountResources()...)
	return resourceTypeList
}

func genAccountResources() []iam.ResourceType {
	return []iam.ResourceType{
		{
			ID:            Account,
			Name:          ResourceTypeIDMap[Account],
			NameEn:        "Account",
			Description:   "账号",
			DescriptionEn: "account",
			Parents: []iam.Parent{{
				SystemID:   SystemIDHCM,
				ResourceID: Account,
			}},
			ProviderConfig: iam.ResourceConfig{
				Path: "/api/v1/auth/iam/find/resource",
			},
			Version: 1,
		},
		{
			ID:            CloudSelectionScheme,
			Name:          ResourceTypeIDMap[CloudSelectionScheme],
			NameEn:        "Scheme",
			Description:   "方案",
			DescriptionEn: "scheme",
			Parents: []iam.Parent{{
				SystemID:   SystemIDHCM,
				ResourceID: CloudSelectionScheme,
			}},
			ProviderConfig: iam.ResourceConfig{
				Path: "/api/v1/auth/iam/find/resource",
			},
			Version: 1,
		},
		{
			ID:            MainAccount,
			Name:          ResourceTypeIDMap[MainAccount],
			NameEn:        "MainAccount",
			Description:   "二级账号",
			DescriptionEn: "main account",
			Parents: []iam.Parent{{
				SystemID:   SystemIDHCM,
				ResourceID: MainAccount,
			}},
			ProviderConfig: iam.ResourceConfig{
				Path: "/api/v1/auth/iam/find/resource",
			},
			Version: 1,
		},
		{
			ID:            BillCloudVendor,
			Name:          ResourceTypeIDMap[BillCloudVendor],
			NameEn:        "BillCloudVendor",
			Description:   "账单云厂商",
			DescriptionEn: "bill cloud vendor",
			Parents: []iam.Parent{{
				SystemID:   SystemIDHCM,
				ResourceID: BillCloudVendor,
			}},
			ProviderConfig: iam.ResourceConfig{
				Path: "/api/v1/auth/iam/find/resource",
			},
			Version: 1,
		},
	}
}
