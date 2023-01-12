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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// Get create account with options
func (a *accountSvc) Get(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// TODO: 校验用户有该账号的权限

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		account, err := a.client.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		account.Extension.CloudSecretKey = ""
		return account, err
	case enumor.Aws:
		account, err := a.client.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		account.Extension.CloudSecretKey = ""
		return account, err
	case enumor.HuaWei:
		account, err := a.client.DataService().HuaWei.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		account.Extension.CloudSecretKey = ""
		return account, err
	case enumor.Gcp:
		account, err := a.client.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		account.Extension.CloudServiceSecretKey = ""
		return account, err
	case enumor.Azure:
		account, err := a.client.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		account.Extension.CloudClientSecretKey = ""
		return account, err
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}
