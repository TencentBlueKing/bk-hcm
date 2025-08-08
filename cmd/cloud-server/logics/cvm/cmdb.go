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

package cvm

import (
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	cvt "hcm/pkg/tools/converter"
)

// GetHostTopoInfo get host topo info in cc 3.0
func (c *cvm) GetHostTopoInfo(kt *kit.Kit, hostIds []int64) ([]cmdb.HostTopoRelation, error) {
	req := &cmdb.HostModuleRelationParams{
		HostID: hostIds,
	}

	resp, err := c.cmdbClient.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return cvt.PtrToVal(resp), nil
}

// GetModuleInfo get module info in cc 3.0
func (c *cvm) GetModuleInfo(kit *kit.Kit, bkBizID int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleParams{
		BizID: bkBizID,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				pkg.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}
	resp, err := c.cmdbClient.SearchModule(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return resp.Info, nil
}
