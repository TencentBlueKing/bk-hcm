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

package mainaccount

import (
	"fmt"
	"reflect"

	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
)

// Deliver 审批通过后资源的交付
func (a *ApplicationOfUpdateMainAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	req := a.req

	var (
		err error
	)
	switch req.Vendor {
	case enumor.Aws, enumor.Gcp, enumor.HuaWei, enumor.Azure, enumor.Zenlayer, enumor.Kaopu:
		err = a.update()
	default:
		err = errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", req.Vendor))
	}
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err}, err
	}

	account, err := a.Client.DataService().Global.MainAccount.GetBasicInfo(a.Cts.Kit, req.ID)
	if err != nil {
		err := fmt.Errorf("update account success, accountId: %s, "+
			"but get main account basic info failed, err: %v, rid: %s", req.ID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	detail := map[string]any{
		"account_id":         account.ID,
		"cloud_account_name": account.Name,
		"cloud_account_id":   account.CloudID,
	}
	return enumor.Completed, detail, nil
}

func (a *ApplicationOfUpdateMainAccount) update() error {
	req := a.req

	oldAccount, err := a.Client.DataService().Global.MainAccount.GetBasicInfo(a.Cts.Kit, a.req.ID)
	if err != nil {
		return err
	}

	updateReq := &dataproto.MainAccountUpdateReq{
		Managers:    oldAccount.Managers,
		BakManagers: oldAccount.BakManagers,
		Status:      oldAccount.Status,
		DeptID:      oldAccount.DeptID,
		OpProductID: oldAccount.OpProductID,
		BkBizID:     oldAccount.BkBizID,
	}

	// 更新负责人，顺序变更时同样需要更新
	if !reflect.DeepEqual(req.Managers, oldAccount.Managers) {
		updateReq.Managers = req.Managers
	}

	// 更新备份负责人，顺序变更时同样需要更新
	if !reflect.DeepEqual(req.BakManagers, oldAccount.BakManagers) {
		updateReq.BakManagers = req.BakManagers
	}

	// 更新业务
	if req.BkBizID != 0 && req.BkBizID != oldAccount.BkBizID {
		updateReq.BkBizID = req.BkBizID
	}

	// 更新组织架
	if req.DeptID != 0 && req.DeptID != oldAccount.DeptID {
		updateReq.DeptID = req.DeptID
	}

	// 更新运营产品
	if req.OpProductID != 0 && req.OpProductID != oldAccount.OpProductID {
		updateReq.OpProductID = req.OpProductID
	}

	_, err = a.Client.DataService().Global.MainAccount.Update(
		a.Cts.Kit,
		a.req.ID,
		updateReq,
	)

	return err
}
