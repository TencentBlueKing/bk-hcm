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

package securitygroup

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

func (svc *securityGroupSvc) listSecurityGroupByIDs(kt *kit.Kit, ids []string) ([]cloud.BaseSecurityGroup, error) {
	resultMap := make(map[string]cloud.BaseSecurityGroup, len(ids))
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.SecurityGroupListReq{
			Filter: tools.ContainersExpression("id", sgIDs),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("ListSecurityGroup failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			resultMap[detail.ID] = detail
		}
	}
	result := make([]cloud.BaseSecurityGroup, 0, len(ids))
	for _, id := range ids {
		item, ok := resultMap[id]
		if !ok {
			logs.Errorf("security group %s not found, rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("security group %s not found", id)
		}
		result = append(result, item)
	}
	return result, nil
}
