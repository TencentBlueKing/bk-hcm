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

package permissionpolicylibrary

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// BatchUpdatePermissionPolicyLibrary batch update permission policy libraries.
func (svc *service) BatchUpdatePermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.PermissionPolicyLibraryBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	existingMap, err := svc.listExistingForUpdate(cts, req.PermissionPolicyLibraries)
	if err != nil {
		return nil, err
	}

	models := make([]tablecloud.PermissionPolicyLibraryTable, 0, len(req.PermissionPolicyLibraries))
	for _, one := range req.PermissionPolicyLibraries {
		model := tablecloud.PermissionPolicyLibraryTable{
			ID:      one.ID,
			Name:    one.Name,
			Memo:    one.Memo,
			Reviser: cts.Kit.User,
		}

		if len(one.BkBizIDs) > 0 {
			bkBizIDsJSON, mErr := json.MarshalToString(one.BkBizIDs)
			if mErr != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, mErr)
			}
			model.BkBizIDs = tabletypes.JsonField(bkBizIDsJSON)
		}

		if len(one.PolicyDocument) > 0 {
			existing, ok := existingMap[one.ID]
			if !ok {
				return nil, errf.New(errf.RecordNotFound,
					fmt.Sprintf("permission_policy_library id %s not found", one.ID))
			}
			newHash := computePolicyHash(one.PolicyDocument)
			model.PolicyDocument = one.PolicyDocument
			model.PolicyHash = newHash
			if newHash != existing.PolicyHash {
				model.Version = existing.Version + 1
			}
		}

		models = append(models, model)
	}

	if err = svc.dao.PermissionPolicyLibrary().BatchUpdate(cts.Kit, models); err != nil {
		logs.Errorf("batch update permission_policy_library failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// listExistingForUpdate queries existing records for items that need policy_document update.
func (svc *service) listExistingForUpdate(cts *rest.Contexts, items []protocloud.PermissionPolicyLibraryUpdate) (
	map[string]tablecloud.PermissionPolicyLibraryTable, error) {

	ids := make([]string, 0)
	for _, one := range items {
		if len(one.PolicyDocument) > 0 {
			ids = append(ids, one.ID)
		}
	}

	result := make(map[string]tablecloud.PermissionPolicyLibraryTable, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	opt := &types.ListOption{
		Fields: []string{"id", "policy_hash", "version"},
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.PermissionPolicyLibrary().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list permission_policy_library for update failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list permission_policy_library failed, err: %v", err)
	}

	for _, one := range listResp.Details {
		result[one.ID] = one
	}

	return result, nil
}
