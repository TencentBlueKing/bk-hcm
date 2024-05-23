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

package cloud

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (ad Audit) listenerDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}

	idMap, err := ad.listListener(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range deletes {
		resData, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: resData.CloudID,
			ResName:    resData.Name,
			ResType:    enumor.ListenerAuditResType,
			Action:     enumor.Delete,
			BkBizID:    resData.BkBizID,
			Vendor:     resData.Vendor,
			AccountID:  resData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: resData,
			},
		})
	}

	return audits, nil
}

func (ad Audit) listListener(kt *kit.Kit, ids []string) (map[string]*tablelb.LoadBalancerListenerTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.LoadBalancerListener().List(kt, opt)
	if err != nil {
		logs.Errorf("list listener db failed, ids: %v, err: %v, rid: %s", ids, err, kt.Rid)
		return nil, err
	}

	result := make(map[string]*tablelb.LoadBalancerListenerTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = &one
	}

	return result, nil
}
