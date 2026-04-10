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

package permissiontemplate

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// ListPermissionTemplate list permission template without extension.
func (svc *service) ListPermissionTemplate(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.PermissionTemplateListReq)
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
	result, err := svc.dao.PermissionTemplate().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list permission_template failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.PermissionTemplateListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.BasePermissionTemplate, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, convTableToBasePermissionTemplate(one))
	}

	return &protocloud.PermissionTemplateListResult{Details: details}, nil
}

// ListPermissionTemplateExt list permission template with extension.
func (svc *service) ListPermissionTemplateExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.PermissionTemplateExtListReq)
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
	result, err := svc.dao.PermissionTemplate().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list permission_template ext failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.PermissionTemplateExtListResult[corecloud.TCloudPermissionTemplateExtension]{
			Count: result.Count,
		}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convPermissionTemplateExtListResult[corecloud.TCloudPermissionTemplateExtension](result.Details)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convPermissionTemplateExtListResult[T corecloud.PermissionTemplateExtension](
	tables []tablecloud.PermissionTemplateTable) (*protocloud.PermissionTemplateExtListResult[T], error) {

	details := make([]corecloud.PermissionTemplate[T], 0, len(tables))
	for _, one := range tables {
		ext := new(T)
		if len(one.Extension) != 0 {
			if err := json.UnmarshalFromString(string(one.Extension), ext); err != nil {
				return nil, fmt.Errorf("unmarshal permission_template extension failed, err: %v", err)
			}
		}

		details = append(details, corecloud.PermissionTemplate[T]{
			BasePermissionTemplate: convTableToBasePermissionTemplate(one),
			Extension:              ext,
		})
	}

	return &protocloud.PermissionTemplateExtListResult[T]{Details: details}, nil
}

func convTableToBasePermissionTemplate(one tablecloud.PermissionTemplateTable) corecloud.BasePermissionTemplate {
	return corecloud.BasePermissionTemplate{
		ID:                    one.ID,
		CloudID:               one.CloudID,
		Name:                  one.Name,
		AccountID:             one.AccountID,
		PolicyLibraryID:       one.PolicyLibraryID,
		PolicyLibraryVersion:  one.PolicyLibraryVersion,
		PolicyLibrarySyncTime: one.PolicyLibrarySyncTime,
		PolicyDocument:        one.PolicyDocument,
		PolicyHash:            one.PolicyHash,
		Memo:                  one.Memo,
		Vendor:                one.Vendor,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}
