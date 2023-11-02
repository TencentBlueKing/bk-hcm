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

	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Get create account with options
func (a *accountSvc) Get(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

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

	switch baseInfo.Vendor {
	case enumor.TCloud:
		account, err := a.client.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		if account != nil {
			account.Extension.CloudSecretKey = ""
		}
		return account, err
	case enumor.Aws:
		account, err := a.client.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		if account != nil {
			account.Extension.CloudSecretKey = ""
		}
		return account, err
	case enumor.HuaWei:
		account, err := a.client.DataService().HuaWei.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		if account != nil {
			account.Extension.CloudSecretKey = ""
		}
		return account, err
	case enumor.Gcp:
		account, err := a.client.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		if account != nil {
			account.Extension.CloudServiceSecretKey = ""
		}
		return account, err
	case enumor.Azure:
		account, err := a.client.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
		// 敏感信息不显示，置空
		if account != nil {
			account.Extension.CloudClientSecretKey = ""
		}
		return account, err
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

// GetAccountBySecret 根据秘钥获取账号信息
func (a *accountSvc) GetAccountBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 获取vendor
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 鉴权 要求录入账号权限
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Import}}
	err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}
	// 3. 根据vendor处理具体内容
	switch vendor {
	case enumor.TCloud:
		return a.getAndCheckTCloudAccountInfo(cts)
	case enumor.Aws:
		return a.getAndCheckAwsAccountInfo(cts)
	case enumor.Azure:
		return a.getAndCheckAzureAccountInfo(cts)
	case enumor.Gcp:
		return a.getAndCheckGcpAccountInfo(cts)
	case enumor.HuaWei:
		return a.getAndCheckHuaWeiAccountInfo(cts)
	}

	return nil, nil
}

func (a *accountSvc) getAndCheckTCloudAccountInfo(cts *rest.Contexts) (*cloud.TCloudInfoBySecret, error) {
	req := new(cloud.TCloudSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().TCloud.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.TCloud, enumor.ResourceAccount,
		info.CloudMainAccountID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckAwsAccountInfo(cts *rest.Contexts) (*cloud.AwsInfoBySecret, error) {
	req := new(cloud.AwsSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Aws.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Aws, enumor.ResourceAccount,
		info.CloudAccountID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckAzureAccountInfo(cts *rest.Contexts) (*cloud.AzureInfoBySecret, error) {
	req := new(cloud.AzureSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Azure.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Azure, enumor.ResourceAccount,
		info.CloudSubscriptionID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckGcpAccountInfo(cts *rest.Contexts) (*cloud.GcpInfoBySecret, error) {
	req := new(cloud.GcpSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Gcp.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Gcp, enumor.ResourceAccount,
		info.CloudProjectID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckHuaWeiAccountInfo(cts *rest.Contexts) (*cloud.HuaWeiInfoBySecret, error) {
	req := new(cloud.HuaWeiSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().HuaWei.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.HuaWei, enumor.ResourceAccount,
		info.CloudSubAccountID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

// GetResCountBySecret 根据秘钥获取账号对应资源数量
func (a *accountSvc) GetResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 获取vendor
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 鉴权 要求录入账号权限
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Import}}
	err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// 3. 根据vendor处理具体内容
	switch vendor {
	case enumor.TCloud:
		req := new(cloud.TCloudSecret)
		if err := cts.DecodeInto(req); err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		if err := req.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		return a.client.HCService().TCloud.Account.GetResCountBySecret(cts.Kit, req)
	case enumor.Aws:
		req := new(cloud.AwsSecret)
		if err := cts.DecodeInto(req); err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		if err := req.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		return a.client.HCService().Aws.Account.GetResCountBySecret(cts.Kit, req)
	case enumor.Azure:
		req := new(cloud.AzureAuthSecret)
		if err := cts.DecodeInto(req); err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		if err := req.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		return a.client.HCService().Azure.Account.GetResCountBySecret(cts.Kit, req)
	case enumor.Gcp:
		req := new(cloud.GcpCredential)
		if err := cts.DecodeInto(req); err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		if err := req.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		return a.client.HCService().Gcp.Account.GetResCountBySecret(cts.Kit, req)
	case enumor.HuaWei:
		req := new(cloud.HuaWeiSecret)
		if err := cts.DecodeInto(req); err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		if err := req.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		return a.client.HCService().HuaWei.Account.GetResCountBySecret(cts.Kit, req)
	default:
		return nil, fmt.Errorf("not support vendor %s", vendor)
	}

	return nil, nil
}
