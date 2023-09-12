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

// Package classifier ...
package classifier

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
)

// ClassifyBasicInfoByVendor classify basic info map by vendor.
func ClassifyBasicInfoByVendor(infoMap map[string]types.CloudResourceBasicInfo) map[enumor.Vendor][]types.
	CloudResourceBasicInfo {

	cvmVendorMap := make(map[enumor.Vendor][]types.CloudResourceBasicInfo, 0)
	for _, info := range infoMap {
		if _, exist := cvmVendorMap[info.Vendor]; !exist {
			cvmVendorMap[info.Vendor] = []types.CloudResourceBasicInfo{info}
			continue
		}

		cvmVendorMap[info.Vendor] = append(cvmVendorMap[info.Vendor], info)
	}

	return cvmVendorMap
}

// ClassifyBasicInfoByAccount classify basic infos by account and region, returns region to ids map group by account.
func ClassifyBasicInfoByAccount(infos []types.CloudResourceBasicInfo) map[string]map[string][]string {
	infoMap := make(map[string]map[string][]string, 0)
	for _, one := range infos {
		if _, exist := infoMap[one.AccountID]; !exist {
			infoMap[one.AccountID] = map[string][]string{
				one.Region: {one.ID},
			}

			continue
		}

		if _, exist := infoMap[one.AccountID][one.Region]; !exist {
			infoMap[one.AccountID][one.Region] = []string{one.ID}
			continue
		}

		infoMap[one.AccountID][one.Region] = append(infoMap[one.AccountID][one.Region], one.ID)
	}

	return infoMap
}
