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
	"crypto/sha256"
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// CreatePermissionTemplate batch create permission templates.
func (svc *service) CreatePermissionTemplate(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreatePermissionTemplate[corecloud.TCloudPermissionTemplateExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func batchCreatePermissionTemplate[T corecloud.PermissionTemplateExtension](
	cts *rest.Contexts, svc *service, vendor enumor.Vendor) (interface{}, error) {

	req := new(protocloud.PermissionTemplateBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablecloud.PermissionTemplateTable, 0, len(req.PermissionTemplates))
		for _, one := range req.PermissionTemplates {
			extJSON, mErr := json.MarshalToString(one.Extension)
			if mErr != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, mErr)
			}

			models = append(models, tablecloud.PermissionTemplateTable{
				CloudID:               one.CloudID,
				Name:                  one.Name,
				AccountID:             one.AccountID,
				PolicyLibraryID:       one.PolicyLibraryID,
				PolicyLibraryVersion:  one.PolicyLibraryVersion,
				PolicyLibrarySyncTime: one.PolicyLibrarySyncTime,
				PolicyDocument:        one.PolicyDocument,
				PolicyHash:            computePolicyHash(one.PolicyDocument),
				Memo:                  one.Memo,
				Extension:             types.JsonField(extJSON),
				Vendor:                vendor,
				TenantID:              cts.Kit.TenantID,
				Creator:               cts.Kit.User,
				Reviser:               cts.Kit.User,
			})
		}

		ids, err := svc.dao.PermissionTemplate().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create permission_template failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return &core.BatchCreateResult{IDs: ids}, nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// computePolicyHash computes SHA256 hash of policy document.
func computePolicyHash(policyDocument string) string {
	h := sha256.New()
	h.Write([]byte(policyDocument))
	return fmt.Sprintf("%x", h.Sum(nil))
}
