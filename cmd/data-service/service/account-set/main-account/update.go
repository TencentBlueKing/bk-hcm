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

// Package mainaccount ...
package mainaccount

import (
	"fmt"

	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	table "hcm/pkg/dal/table/account-set"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateAccount account with filter.
func (svc *service) UpdateMainAccount(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	return updateMainAccount(accountID, svc, cts)
}

func updateMainAccount(accountID string,
	svc *service, cts *rest.Contexts) (interface{}, error) {

	req := new(dataproto.MainAccountUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account := &table.MainAccountTable{
		Managers:    req.Managers,
		BakManagers: req.BakManagers,
		Status:      string(req.Status),
		DeptID:      req.DeptID,
		BkBizID:     req.BkBizID,
		OpProductID: req.OpProductID,
		Reviser:     cts.Kit.User,
	}

	err := svc.dao.MainAccount().Update(cts.Kit, tools.EqualExpression("id", accountID), account)
	if err != nil {
		logs.Errorf("update main account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("update main account failed, err: %v", err)
	}

	return nil, nil
}
