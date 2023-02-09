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

package region

import (
	apicloudregion "hcm/pkg/api/core/cloud/region"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
)

// HuaWeiRegionSync ...
type HuaWeiRegionSync struct {
	IsUpdate bool
	Region   iammodel.Region
}

// HuaWeiDSRegionSync ...
type HuaWeiDSRegionSync struct {
	Region apicloudregion.HuaWeiRegion
}

// AzureRegionSync ...
type AzureRegionSync struct {
	IsUpdate bool
	Region   *armsubscriptions.Location
}

// AzureDSRegionSync ...
type AzureDSRegionSync struct {
	Region apicloudregion.AzureRegion
}

// AzureRGSync ...
type AzureRGSync struct {
	IsUpdate bool
	Region   *armresources.ResourceGroup
}

// AzureDSRGSync ...
type AzureDSRGSync struct {
	Region apicloudregion.AzureRG
}
