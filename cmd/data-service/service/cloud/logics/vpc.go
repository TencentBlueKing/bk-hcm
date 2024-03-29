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

package logics

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// GetVpcIDByCloudID get vpc cloud id to id map from cloud ids, used for related resources.
func GetVpcIDByCloudID(kt *kit.Kit, dao dao.Set, vendor enumor.Vendor, cloudIDs []string) (map[string]string, error) {
	if len(cloudIDs) == 0 {
		return make(map[string]string), nil
	}

	opt := &types.ListOption{
		Page: &core.BasePage{Count: false, Start: 0, Limit: uint(len(cloudIDs))},
	}

	switch vendor {
	// gcp vpc cloud id is self link for related resources.
	case enumor.Gcp:
		opt.Filter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(), Value: cloudIDs},
				filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			},
		}
	default:
		opt.Filter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: cloudIDs},
				filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			},
		}
	}

	res, err := dao.Vpc().List(kt, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}

	idMap := make(map[string]string, len(res.Details))
	for _, detail := range res.Details {
		idMap[detail.CloudID] = detail.ID
	}

	return idMap, nil
}
