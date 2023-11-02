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

package subaccount

import (
	cssubaccount "hcm/pkg/api/cloud-server/sub-account"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateSubAccount update sub account.
func (svc *service) UpdateSubAccount(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(cssubaccount.UpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SubAccountCloudResType, id)
	if err != nil {
		logs.Errorf("get sub account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// list authorized instances
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SubAccount, Action: meta.Update,
		ResourceID: basicInfo.AccountID}}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("update sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &dssubaccount.UpdateReq{
		Items: []dssubaccount.UpdateField{
			{
				ID:       id,
				Managers: req.Managers,
				BkBizIDs: req.BkBizIDs,
				Memo:     req.Memo,
			},
		},
	}

	if err := svc.client.DataService().Global.SubAccount.BatchUpdate(cts.Kit, updateReq); err != nil {
		logs.Errorf("request ds to update sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
