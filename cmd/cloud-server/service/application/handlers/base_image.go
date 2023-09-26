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

	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/runtime/filter"
)

// GetImage 查询镜像
func (a *BaseApplicationHandler) GetImage(vendor enumor.Vendor, cloudImageID string) (
	*coreimage.BaseImage, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudImageID},
		},
	}
	// 查询
	listReq := &core.ListReq{
		Filter: reqFilter,
		Page:   a.getPageOfOneLimit(),
	}
	resp, err := a.Client.DataService().Global.ListImage(a.Cts.Kit, listReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s image by cloud_id(%s)", vendor, cloudImageID)
	}

	return resp.Details[0], nil
}
