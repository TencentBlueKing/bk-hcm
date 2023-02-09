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

package huawei

import (
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

// ListRegion 查看地域
// reference: https://support.huaweicloud.com/api-iam/iam_05_0001.html
func (h *HuaWei) ListRegion(kt *kit.Kit) ([]iammodel.Region, error) {

	client, err := h.clientSet.iamRegionClient(region.AP_SOUTHEAST_1.Id)
	if err != nil {
		return nil, err
	}

	req := new(iammodel.KeystoneListRegionsRequest)

	resp, err := client.KeystoneListRegions(req)
	if err != nil {
		logs.Errorf("list huawei region failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return *resp.Regions, nil
}
