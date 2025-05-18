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

package handlers

import (
	"fmt"

	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// ListBizNames 查询业务名称列表
func (a *BaseApplicationHandler) ListBizNames(bkBizIDs []int64) ([]string, error) {
	// 查询CC业务
	searchResp, err := a.CmdbClient.SearchBusiness(a.Cts.Kit, &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
	})
	if err != nil {
		return []string{}, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}
	// 业务ID和Name映射关系
	bizNameMap := map[int64]string{}
	for _, biz := range searchResp.Info {
		bizNameMap[biz.BizID] = biz.BizName
	}
	// 匹配出业务名称列表
	bizNames := make([]string, 0, len(bkBizIDs))
	for _, bizID := range bkBizIDs {
		bizNames = append(bizNames, bizNameMap[bizID])
	}

	return bizNames, nil
}

// GetBizName 查询业务名称
func (a *BaseApplicationHandler) GetBizName(bkBizID int64) (string, error) {
	bizNames, err := a.ListBizNames([]int64{bkBizID})
	if err != nil || len(bizNames) != 1 {
		return "", err
	}

	return bizNames[0], nil
}

// GetCloudAreaName 查询云区域名称
func (a *BaseApplicationHandler) GetCloudAreaName(bkCloudAreaID int64) (string, error) {
	res, err := a.CmdbClient.SearchCloudArea(
		a.Cts.Kit,
		&cmdb.SearchCloudAreaParams{
			Fields: []string{"bk_cloud_id", "bk_cloud_name"},
			Page: cmdb.BasePage{
				Limit: 1,
				Start: 0,
				Sort:  "bk_cloud_id",
			},
			Condition: map[string]interface{}{"bk_cloud_id": bkCloudAreaID},
		},
	)
	if err != nil {
		return "", fmt.Errorf("call cmdb search cloud area api failed, err: %v", err)
	}

	for _, cloudArea := range res.Info {
		if cloudArea.CloudID == bkCloudAreaID {
			return cloudArea.CloudName, nil
		}
	}

	return "", fmt.Errorf("not found bk cloud area by bk_cloud_area_id(%d)", bkCloudAreaID)
}
