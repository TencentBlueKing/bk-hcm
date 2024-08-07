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

package rootaccount

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// Get get root account with options
func (s *service) Get(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有一级账号管理权限
	if err := s.checkPermission(cts, meta.RootAccount, meta.Find); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := s.client.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.Aws:
		account, err := s.client.DataService().Aws.RootAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudSecretKey = ""
		}
		return account, err
	case enumor.Gcp:
		account, err := s.client.DataService().Gcp.RootAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudServiceSecretKey = ""
		}
		return account, err
	case enumor.Azure:
		account, err := s.client.DataService().Azure.RootAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudClientSecretKey = ""
		}
		return account, err
	case enumor.HuaWei:
		account, err := s.client.DataService().HuaWei.RootAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudSecretKey = ""
		}
		return account, err
	case enumor.Zenlayer:
		account, err := s.client.DataService().Zenlayer.RootAccount.Get(cts.Kit, accountID)
		// zenlayer not support store secret info
		// if account != nil {
		// }
		return account, err
	case enumor.Kaopu:
		account, err := s.client.DataService().Kaopu.RootAccount.Get(cts.Kit, accountID)
		// kaopu not support store secret info
		// if account != nil {
		// }
		return account, err
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}
