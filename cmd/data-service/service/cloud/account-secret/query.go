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

package accountsecret

import (
	"fmt"

	"hcm/pkg/api/core"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tableas "hcm/pkg/dal/table/cloud/account-secret"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// ListAccountSecret list account secret
func (svc *accountSecretSvc) ListAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountSecretListReq)
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

	result, err := svc.dao.AccountSecret().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.AccountSecretListResult{Count: result.Count}, nil
	}

	details := make([]coreas.BaseAccountSecret, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, convTableToBaseAccountSecret(one))
	}

	return &protocloud.AccountSecretListResult{Details: details}, nil
}

func convTableToBaseAccountSecret(one tableas.Table) coreas.BaseAccountSecret {
	return coreas.BaseAccountSecret{
		ID:        one.ID,
		AccountID: one.AccountID,
		Vendor:    one.Vendor,
		Type:      one.Type,
		Status:    one.Status,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListAccountSecretWithExtension list account secret with extension
func (svc *accountSecretSvc) ListAccountSecretWithExtension(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.AccountSecretExtListReq)
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

	result, err := svc.dao.AccountSecret().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.AccountSecretExtListResult[coreas.TCloudAccountSecretExtension]{Count: result.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convAccountSecretListResult[coreas.TCloudAccountSecretExtension](result.Details)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func convAccountSecretListResult[T coreas.Extension](tables []tableas.Table) (
	*protocloud.AccountSecretExtListResult[T], error) {

	details := make([]coreas.AccountSecret[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if len(one.Extension) != 0 {
			if err := json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("unmarshal extension failed, err: %v", err)
			}
		}
		details = append(details, coreas.AccountSecret[T]{
			BaseAccountSecret: convTableToBaseAccountSecret(one),
			Extension:         extension,
		})
	}

	return &protocloud.AccountSecretExtListResult[T]{
		Details: details,
	}, nil
}
