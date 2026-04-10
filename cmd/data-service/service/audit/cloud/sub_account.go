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
	tablesubaccount "hcm/pkg/dal/table/cloud/sub-account"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (ad Audit) subAccountUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}

	idMap, err := ad.listSubAccount(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		subAccount, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:     one.ResID,
			ResName:   subAccount.Name,
			ResType:   enumor.SubAccountAuditResType,
			Action:    enumor.Update,
			Vendor:    subAccount.Vendor,
			AccountID: subAccount.AccountID,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    subAccount,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

func (ad Audit) subAccountDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}

	idMap, err := ad.listSubAccount(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		subAccount, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:     one.ResID,
			ResName:   subAccount.Name,
			ResType:   enumor.SubAccountAuditResType,
			Action:    enumor.Delete,
			Vendor:    subAccount.Vendor,
			AccountID: subAccount.AccountID,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: subAccount,
			},
		})
	}

	return audits, nil
}

func (ad Audit) listSubAccount(kt *kit.Kit, ids []string) (map[string]*tablesubaccount.Table, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}

	list, err := ad.dao.SubAccount().List(kt, opt)
	if err != nil {
		logs.Errorf("list sub account failed, ids: %v, err: %v, rid: %s", ids, err, kt.Rid)
		return nil, err
	}

	result := make(map[string]*tablesubaccount.Table, len(list.Details))
	for idx := range list.Details {
		result[list.Details[idx].ID] = &list.Details[idx]
	}

	return result, nil
}
