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

package account

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// DeleteValidate ...
func (a *accountSvc) DeleteValidate(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// auth
	if err := a.checkPermission(cts, meta.Delete, accountID); err != nil {
		return nil, err
	}

	return a.client.DataService().Global.Account.DeleteValidate(cts.Kit.Ctx, cts.Kit.Header(), accountID)
}

// DeleteAccount ...
func (a *accountSvc) DeleteAccount(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// TODO 添加审计，最终是在dao去记录，因为账号删除有校验，存在删除不了的情况，且账号不需要在云上删除。

	// auth
	if err := a.checkPermission(cts, meta.Delete, accountID); err != nil {
		return nil, err
	}

	// 查询账号基本信息
	resp, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), &protocloud.AccountListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("accountID: %s is not found", accountID)
	}

	req := &protocloud.AccountDeleteReq{
		Filter: tools.EqualExpression("id", accountID),
	}
	accountResp, err := a.client.DataService().Global.Account.Delete(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return accountResp, err
	}

	retryNum := 3
	vendor := resp.Details[0].Vendor
	switch vendor {
	case enumor.Aws:
		for retryNum > 0 {
			// 删除云账单配置信息
			billErr := a.client.HCService().Aws.Bill.Delete(cts.Kit.Ctx, cts.Kit.Header(), accountID)
			if billErr != nil {
				logs.Errorf("aws account db delete success and bill config delete failed, accountID: %s, err: %+v",
					accountID, billErr)
				retryNum--
			}
			break
		}
	}

	return accountResp, nil
}
