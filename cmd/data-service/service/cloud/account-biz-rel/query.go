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

package accountbizrel

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListAccountBizRel list account biz relation.
func (a *service) ListAccountBizRel(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	data, err := a.dao.AccountBizRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account biz relations failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.AccountBizRelListResult{Count: data.Count}, nil
	}

	details := make([]corecloud.AccountBizRel, len(data.Details))
	for idx, table := range data.Details {
		details[idx] = corecloud.AccountBizRel{
			BkBizID:   table.BkBizID,
			AccountID: table.AccountID,
			Creator:   table.Creator,
			CreatedAt: table.CreatedAt.String(),
		}
	}

	return &protocloud.AccountBizRelListResult{Details: details}, nil
}

// ListWithAccount ...
func (a *service) ListWithAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountBizRelWithAccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	details, err := a.dao.AccountBizRel().ListJoinAccount(cts.Kit, req.BkBizIDs)
	if err != nil {
		logs.Errorf("list account biz rels join account failed, err: %v, bkBizIds: %v, rid: %s", err,
			req.BkBizIDs, cts.Kit.Rid)
		return nil, err
	}

	accounts := make([]*protocloud.AccountBizRelWithAccount, 0, len(details.Details))
	for _, one := range details.Details {
		// 过滤账号类型
		if req.AccountType != "" && req.AccountType != one.Type {
			continue
		}

		accounts = append(accounts, &protocloud.AccountBizRelWithAccount{
			BaseAccount: corecloud.BaseAccount{
				ID:        one.ID,
				Vendor:    enumor.Vendor(one.Vendor),
				Name:      one.Name,
				Managers:  one.Managers,
				Type:      enumor.AccountType(one.Type),
				Site:      enumor.AccountSiteType(one.Site),
				Price:     one.Price,
				PriceUnit: one.PriceUnit,
				Memo:      one.Memo,
				Revision: core.Revision{
					Creator:   one.Creator,
					Reviser:   one.Reviser,
					CreatedAt: one.CreatedAt.String(),
					UpdatedAt: one.UpdatedAt.String(),
				},
			},
			BkBizID:      one.BkBizID,
			RelCreator:   one.RelCreator,
			RelCreatedAt: one.RelCreatedAt.String(),
		})
	}

	return accounts, nil
}
