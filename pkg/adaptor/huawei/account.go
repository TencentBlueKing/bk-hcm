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

package huawei

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

var _ types.AccountInterface = new(huawei)

func validateAccountCheckOption(opt *types.AccountCheckOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "account check option is required")
	}

	if opt.HuaWei == nil {
		return errf.New(errf.InvalidParameter, "huawei account info is required")
	}

	if len(opt.HuaWei.MainAccountName) == 0 {
		return errf.New(errf.InvalidParameter, "main account name is required")
	}

	if len(opt.HuaWei.SubAccountName) == 0 {
		return errf.New(errf.InvalidParameter, "sub account name is required")
	}

	if len(opt.HuaWei.SubAccountCID) == 0 {
		return errf.New(errf.InvalidParameter, "sub account cid is required")
	}

	if len(opt.HuaWei.IamUserCID) == 0 {
		return errf.New(errf.InvalidParameter, "iam user cid is required")
	}

	if len(opt.HuaWei.IamUserName) == 0 {
		return errf.New(errf.InvalidParameter, "iam user name is required")
	}

	return nil
}

// AccountCheck check account authentication information and permissions.
// KeystoneListAuthDomains: https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_07_0001.html
func (h *huawei) AccountCheck(kt *kit.Kit, secret *types.Secret, opt *types.AccountCheckOption) error {
	if err := validateSecret(secret); err != nil {
		return err
	}

	if err := validateAccountCheckOption(opt); err != nil {
		return err
	}

	client, err := h.iamClient(secret.HuaWei, region.AP_SOUTHEAST_1)
	if err != nil {
		return fmt.Errorf("init huawei client failed, err: %v", err)
	}

	_, err = client.KeystoneListAuthDomains(nil)
	if err != nil {
		logs.Errorf("describe regions failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
