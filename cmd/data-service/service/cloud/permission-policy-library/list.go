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
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// ListPermissionPolicyLibrary list permission policy library.
func (svc *service) ListPermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.PermissionPolicyLibraryListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.PermissionPolicyLibrary().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list permission_policy_library failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.PermissionPolicyLibraryListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.BasePermissionPolicyLibrary, 0, len(result.Details))
	for _, one := range result.Details {
		base, err := convTableToBasePermissionPolicyLibrary(one)
		if err != nil {
			return nil, err
		}
		details = append(details, base)
	}

	return &protocloud.PermissionPolicyLibraryListResult{Details: details}, nil
}

func convTableToBasePermissionPolicyLibrary(one tablecloud.PermissionPolicyLibraryTable) (
	corecloud.BasePermissionPolicyLibrary, error) {

	var bkBizIDs []int64
	if len(one.BkBizIDs) > 0 {
		if err := json.UnmarshalFromString(string(one.BkBizIDs), &bkBizIDs); err != nil {
			return corecloud.BasePermissionPolicyLibrary{},
				fmt.Errorf("unmarshal bk_biz_ids failed, err: %v", err)
		}
	}

	return corecloud.BasePermissionPolicyLibrary{
		ID:             one.ID,
		Name:           one.Name,
		PolicyDocument: one.PolicyDocument,
		PolicyHash:     one.PolicyHash,
		Version:        one.Version,
		BkBizIDs:       bkBizIDs,
		Memo:           one.Memo,
		Vendor:         one.Vendor,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}, nil
}
