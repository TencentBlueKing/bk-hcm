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
	"strconv"
	"strings"

	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// GetAccount get account with options
func (a *accountSvc) GetAccount(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的查看权限
	if err := a.checkGetAccountPermission(cts, accountID); err != nil {
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
		return a.getTCloudAccount(cts, accountID)
	case enumor.Aws:
		return a.getAwsAccount(cts, accountID)
	case enumor.HuaWei:
		return a.getHuaWeiAccount(cts, accountID)
	case enumor.Gcp:
		return a.getGcpAccount(cts, accountID)
	case enumor.Azure:
		return a.getAzureAccount(cts, accountID)
	case enumor.Other:
		return a.getOtherAccount(cts, accountID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

func (a *accountSvc) getTCloudAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		logs.Errorf("get tcloud acc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 敏感信息不显示，置空
	if acc != nil {
		acc.Extension.CloudSecretKey = ""
	}
	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) getAwsAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		logs.Errorf("get aws acc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 敏感信息不显示，置空
	if acc != nil {
		acc.Extension.CloudSecretKey = ""
	}
	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) getHuaWeiAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().HuaWei.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		logs.Errorf("get huawei acc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 敏感信息不显示，置空
	if acc != nil {
		acc.Extension.CloudSecretKey = ""
	}
	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) getGcpAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		logs.Errorf("get gcp acc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 敏感信息不显示，置空
	if acc != nil {
		acc.Extension.CloudServiceSecretKey = ""
	}
	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) getAzureAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		logs.Errorf("get azure acc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 敏感信息不显示，置空
	if acc != nil {
		acc.Extension.CloudClientSecretKey = ""
	}
	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) getOtherAccount(cts *rest.Contexts, accountID string) (interface{}, error) {
	acc, err := a.client.DataService().Other.Account.Get(cts.Kit, accountID)
	if err != nil {
		logs.Errorf("get account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err = accountDetailFullFill(a, cts, acc); err != nil {
		logs.Errorf("acc detail full fill failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return acc, nil
}

func (a *accountSvc) checkGetAccountPermission(cts *rest.Contexts, accountID string) error {
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	accounts, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, req: %v, rid: %s", err, listReq, cts.Kit.Rid)
		return err
	}
	if len(accounts.Details) == 0 {
		return fmt.Errorf("account not found: %s", accountID)
	}
	if slice.IsItemInSlice(accounts.Details[0].Managers, cts.Kit.User) {
		return nil
	}

	// 账号查看权限校验 & 业务访问权限
	resources := []meta.ResourceAttribute{
		{
			Basic: &meta.Basic{
				Type:       meta.Account,
				Action:     meta.Find,
				ResourceID: accountID,
			},
		},
	}
	for _, bkBizID := range accounts.Details[0].UsageBizIDs {
		resources = append(resources, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.Biz,
				Action:     meta.Access,
				ResourceID: strconv.FormatInt(bkBizID, 10),
			},
			BizID: bkBizID,
		})
	}
	decisions, authorized, err := a.authorizer.Authorize(cts.Kit, resources...)
	if authorized {
		return nil
	}
	for i := range decisions {
		if decisions[i].Authorized {
			return nil
		}
	}
	return fmt.Errorf("no permission for access account %s", accountID)
}

// 补充回收详情，转换回收时间
func accountDetailFullFill[T protocloud.AccountExtensionGetResp](svc *accountSvc, cts *rest.Contexts,
	acc *protocloud.AccountGetResult[T]) error {

	acc.RecycleReserveTime = convertRecycleReverseTime(acc.RecycleReserveTime)
	syncDetails, err := svc.getAccountsSyncDetail(cts, acc.ID)
	if err != nil {
		logs.Errorf("fail to get account sync detail, accountID: %s, rid: %s", acc.ID, cts.Kit.Rid)
		return err
	}
	for _, detail := range syncDetails[acc.ID] {
		acc.SyncStatus = detail.ResStatus
		if detail.ResStatus == string(enumor.SyncFailed) {
			acc.SyncFailedReason = string(detail.ResFailedReason)
			break
		}
	}
	return nil
}

func convertRecycleReverseTime(val int) int {
	if val == constant.UnsetRecycleTime {
		return int(cc.CloudServer().Recycle.AutoDeleteTime)
	}
	return val
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
	req := new(account.TCloudAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().TCloud.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.TCloudSecret)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if req.DisableCheck {
		return info, nil
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.TCloud, enumor.ResourceAccount,
		info.CloudMainAccountID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckAwsAccountInfo(cts *rest.Contexts) (*cloud.AwsInfoBySecret, error) {
	req := new(account.AwsAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Aws.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.AwsSecret)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if req.DisableCheck {
		return info, nil
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Aws, enumor.ResourceAccount,
		info.CloudAccountID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return info, nil
}

func (a *accountSvc) getAndCheckAzureAccountInfo(cts *rest.Contexts) (*account.AzureAccountInfoBySecretResp, error) {
	req := new(account.AzureAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Azure.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.AzureSecret)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 校验订阅数量，要求订阅数量刚好一个
	if len(info.SubscriptionInfos) > 1 {
		subs := make([]string, len(info.SubscriptionInfos))
		for i, sub := range info.SubscriptionInfos {
			subs[i] = "(" + sub.CloudSubscriptionID + ")" + sub.CloudSubscriptionName
		}
		logs.Errorf("more than one subscription found: %s, rid: %s", strings.Join(subs, ","), cts.Kit.Rid)
		return nil, fmt.Errorf("more than one subscription found: %s", strings.Join(subs, ","))
	}
	subscription := info.SubscriptionInfos[0]
	result := &account.AzureAccountInfoBySecretResp{
		CloudSubscriptionID:   subscription.CloudSubscriptionID,
		CloudSubscriptionName: subscription.CloudSubscriptionName,
	}
	// 补全ApplicationName
	for _, one := range info.ApplicationInfos {
		if one.CloudApplicationID == req.CloudApplicationID {
			result.CloudApplicationName = one.CloudApplicationName
			break
		}
	}
	// 没有拿到应用id的情况
	if len(result.CloudApplicationName) == 0 {
		logs.Errorf("failed to get application name, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("failed to get application name")
	}

	if req.DisableCheck {
		return result, nil
	}
	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Azure, enumor.ResourceAccount,
		subscription.CloudSubscriptionID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func (a *accountSvc) getAndCheckGcpAccountInfo(cts *rest.Contexts) (*cloud.GcpProjectInfo, error) {
	req := new(account.GcpAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().Gcp.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.GcpSecret)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 只能有一个对应的project，多project报错。
	if len(info.CloudProjectInfos) > 1 {
		projects := make([]string, len(info.CloudProjectInfos))
		for i, project := range info.CloudProjectInfos {
			projects[i] = "(" + project.CloudProjectID + ")" + project.CloudProjectName
		}
		logs.Errorf("more than one project found: %s, rid: %s", strings.Join(projects, ","), cts.Kit.Rid)
		return nil, fmt.Errorf("more than one project found: %s", strings.Join(projects, ","))
	}
	result := info.CloudProjectInfos[0]

	if req.DisableCheck {
		return &result, nil
	}

	if err = CheckDuplicateMainAccount(cts, a.client, enumor.Gcp, enumor.ResourceAccount,
		result.CloudProjectID); err != nil {
		logs.Errorf("check whether main account duplicate fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return &result, nil
}

func (a *accountSvc) getAndCheckHuaWeiAccountInfo(cts *rest.Contexts) (*cloud.HuaWeiInfoBySecret, error) {
	req := new(account.HuaWeiAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := a.client.HCService().HuaWei.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.HuaWeiSecret)
	if err != nil {
		logs.Errorf("fail to get account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if req.DisableCheck {
		return info, nil
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

}
