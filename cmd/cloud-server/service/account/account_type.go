/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

func (a *accountSvc) GetTCloudNetworkAccountType(cts *rest.Contexts) (any, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "accountID is required")
	}

	// 校验用户有该账号的查看权限
	if err := a.checkPermission(cts, meta.Find, accountID); err != nil {
		return nil, err
	}
	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if baseInfo.Vendor != enumor.TCloud {
		return nil, errf.New(errf.InvalidParameter, "only TCloud account is support now")
	}
	return a.client.HCService().TCloud.Account.GetNetworkAccountType(cts.Kit, accountID)
}
