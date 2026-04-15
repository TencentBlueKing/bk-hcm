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
	"fmt"

	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
)

func (ad Audit) permissionTemplateUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}

	permTmplMap, accountMap, err := ad.listPermissionTemplateAndAccount(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		resData, exist := permTmplMap[one.ResID]
		if !exist {
			continue
		}
		account, exist := accountMap[resData.AccountID]
		if !exist {
			return nil, fmt.Errorf("account %s not found", resData.AccountID)
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ResID,
			ResName:  resData.Name,
			ResType:  enumor.PermissionTemplateAuditResType,
			Action:   enumor.Update,
			BkBizID:  account.BkBizID,
			Vendor:   resData.Vendor,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    resData,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

func (ad Audit) permissionTemplateDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}

	permTmplMap, accountMap, err := ad.listPermissionTemplateAndAccount(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		resData, exist := permTmplMap[one.ResID]
		if !exist {
			continue
		}
		account, exist := accountMap[resData.AccountID]
		if !exist {
			return nil, fmt.Errorf("account %s not found", resData.AccountID)
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ResID,
			ResName:  resData.Name,
			ResType:  enumor.PermissionTemplateAuditResType,
			Action:   enumor.Delete,
			BkBizID:  account.BkBizID,
			Vendor:   resData.Vendor,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: resData,
			},
		})
	}

	return audits, nil
}

func (ad Audit) listPermissionTemplateAndAccount(kt *kit.Kit, ids []string) (
	map[string]*tablecloud.PermissionTemplateTable, map[string]tablecloud.AccountTable, error) {

	idMap, err := ad.listPermissionTemplate(kt, ids)
	if err != nil {
		return nil, nil, err
	}

	accountIDMap := make(map[string]struct{})
	for _, resData := range idMap {
		accountIDMap[resData.AccountID] = struct{}{}
	}
	accounts, err := ad.listAccount(kt, maps.Keys(accountIDMap))
	if err != nil {
		return nil, nil, err
	}
	accountMap := make(map[string]tablecloud.AccountTable, len(accounts))
	for _, one := range accounts {
		accountMap[one.ID] = one
	}

	return idMap, accountMap, nil
}

func (ad Audit) listPermissionTemplate(kt *kit.Kit, ids []string) (
	map[string]*tablecloud.PermissionTemplateTable, error) {

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.PermissionTemplate().List(kt, opt)
	if err != nil {
		logs.Errorf("list permission_template failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]*tablecloud.PermissionTemplateTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = converter.ValToPtr(one)
	}

	return result, nil
}
