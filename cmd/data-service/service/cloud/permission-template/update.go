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

	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// BatchUpdatePermissionTemplate batch update permission templates.
func (svc *service) BatchUpdatePermissionTemplate(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdatePermissionTemplate[corecloud.TCloudPermissionTemplateExtension](cts, svc)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func batchUpdatePermissionTemplate[T corecloud.PermissionTemplateExtension](
	cts *rest.Contexts, svc *service) (interface{}, error) {

	req := new(protocloud.PermissionTemplateBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	models := make([]tablecloud.PermissionTemplateTable, 0, len(req.PermissionTemplates))
	for _, one := range req.PermissionTemplates {
		model := tablecloud.PermissionTemplateTable{
			ID:                    one.ID,
			Name:                  one.Name,
			PolicyLibraryID:       one.PolicyLibraryID,
			PolicyLibraryVersion:  one.PolicyLibraryVersion,
			PolicyLibrarySyncTime: one.PolicyLibrarySyncTime,
			Memo:                  one.Memo,
			Reviser:               cts.Kit.User,
		}

		if one.Extension != nil {
			extJSON, mErr := json.MarshalToString(one.Extension)
			if mErr != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, mErr)
			}
			model.Extension = tabletypes.JsonField(extJSON)
		}

		if len(one.PolicyDocument) > 0 {
			model.PolicyDocument = one.PolicyDocument
			model.PolicyHash = computePolicyHash(one.PolicyDocument)
		}

		models = append(models, model)
	}

	if err := svc.dao.PermissionTemplate().BatchUpdate(cts.Kit, models); err != nil {
		logs.Errorf("batch update permission_template failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
