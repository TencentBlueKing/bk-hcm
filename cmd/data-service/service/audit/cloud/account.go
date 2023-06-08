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
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

func (ad Audit) accountUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idAccountMap, err := ad.listAccount(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		account, exist := idAccountMap[one.ResID]
		if !exist {
			continue
		}

		// remove secret key from account info & update fields
		extension := tools.AccountExtensionRemoveSecretKey(string(account.Extension))
		account.Extension = tabletype.JsonField(extension)

		updateExt, exists := one.UpdateFields["extension"]
		if exists {
			updateExtJson, err := json.Marshal(updateExt)
			if err != nil {
				logs.Errorf("marshal update account extension failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			updateExtension := tools.AccountExtensionRemoveSecretKey(string(updateExtJson))
			one.UpdateFields["extension"] = updateExtension
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:     one.ResID,
			ResName:   account.Name,
			ResType:   enumor.AccountAuditResType,
			Action:    enumor.Update,
			Vendor:    enumor.Vendor(account.Vendor),
			AccountID: account.ID,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    account,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

func (ad Audit) listAccount(kt *kit.Kit, ids []string) (map[string]tablecloud.AccountTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.Account().List(kt, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.AccountTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = *one
	}

	return result, nil
}
