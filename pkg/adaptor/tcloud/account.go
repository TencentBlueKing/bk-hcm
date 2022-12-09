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

package tcloud

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
)

var _ types.AccountInterface = new(tcloud)

// AccountCheck check account authentication information and permissions.
// reference: https://cloud.tencent.com/document/api/598/70416
func (t *tcloud) AccountCheck(kt *kit.Kit, secret *types.Secret, opt *types.AccountCheckOption) error {
	if err := validateSecret(secret); err != nil {
		return err
	}

	if err := validateAccountCheckOption(opt); err != nil {
		return err
	}

	camClient, err := t.camServiceClient(secret.TCloud, "")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewGetUserAppIdRequest()
	resp, err := camClient.GetUserAppIdWithContext(kt.Ctx, req)
	if err != nil {
		return fmt.Errorf("get user app id failed, err: %v", err)
	}

	if resp.Response.Uin == nil {
		return fmt.Errorf("user uin is empty, err: %v", err)
	}

	if resp.Response.OwnerUin == nil {
		return fmt.Errorf("user owner uin is empty, err: %v", err)
	}

	// check if cloud account info matches the hcm account detail.
	if *resp.Response.Uin != opt.Tcloud.AccountCid {
		return fmt.Errorf("account id does not match the account to which the secret belongs")
	}

	if *resp.Response.OwnerUin != opt.Tcloud.MainAccountCid {
		return fmt.Errorf("main account id does not match the account to which the secret belongs")
	}

	return nil
}

func validateAccountCheckOption(opt *types.AccountCheckOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "account check option is required")
	}

	if opt.Tcloud == nil {
		return errf.New(errf.InvalidParameter, "tencent cloud account info is required")
	}

	if len(opt.Tcloud.MainAccountCid) == 0 {
		return errf.New(errf.InvalidParameter, "main account cid is required")
	}

	if len(opt.Tcloud.AccountCid) == 0 {
		return errf.New(errf.InvalidParameter, "account cid is required")
	}

	return nil
}
