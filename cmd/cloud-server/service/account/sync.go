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
	"errors"
	"fmt"
	"time"

	"hcm/cmd/cloud-server/logics/account"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// AccountSyncDefaultTimeout 账号同步的默认超时时间
const AccountSyncDefaultTimeout = time.Minute * 10

// SyncCloudResource ...
func (a *accountSvc) SyncCloudResource(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, meta.Update, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err = account.Sync(cts.Kit, a.client, baseInfo.Vendor, accountID); err != nil {
		return nil, err
	}

	return nil, nil
}

// SyncCloudResourceByCond sync cloud resource by given condition
func (a *accountSvc) SyncCloudResourceByCond(cts *rest.Contexts) (any, error) {
	accountID := cts.PathParameter("account_id").String()
	resName := enumor.CloudResourceType(cts.PathParameter("res").String())
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

	// 校验用户有该账号的访问权限
	if err := a.checkPermission(cts, meta.Find, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if baseInfo.Vendor != vendor {
		return nil, errf.Newf(errf.InvalidParameter, "account not found by vendor: %s", vendor)
	}

	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return a.tcloudCondSyncRes(cts, accountID, resName)
	case enumor.HuaWei:
		return a.huaweiCondSyncRes(cts, accountID, resName)
	case enumor.Aws:
		return a.awsCondSyncRes(cts, accountID, resName)
	case enumor.Azure:
		return a.azureCondSyncRes(cts, accountID, resName)
	default:
		return nil, fmt.Errorf("conditional sync doesnot support vendor: %s", vendor)
	}
}

// SyncBizCloudResourceByCond sync cloud resource of biz by given condition
func (a *accountSvc) SyncBizCloudResourceByCond(cts *rest.Contexts) (any, error) {
	bkBizId, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	accountID := cts.PathParameter("account_id").String()
	resName := enumor.CloudResourceType(cts.PathParameter("res").String())
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

	// 校验用户有业务访问权限
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizId,
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	// 查询该账号对应的Vendor
	usageBizIDs := []int64{bkBizId}
	// 关联了所有使用业务的账号也应该被查出来
	if bkBizId != constant.AttachedAllBiz {
		usageBizIDs = append(usageBizIDs, constant.AttachedAllBiz)
	}
	bizRelReq := &cloud.AccountBizRelWithAccountListReq{UsageBizIDs: usageBizIDs}
	accountBizList, err := a.client.DataService().Global.Account.ListAccountBizRelWithAccount(cts.Kit, bizRelReq)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	found := false
	for i := range accountBizList {
		accountWithBiz := accountBizList[i]
		if accountWithBiz.ID != accountID || accountWithBiz.Vendor != vendor {
			continue
		}
		found = true
		break
	}
	if !found {
		return nil, errors.New("account not found by biz or vendor")
	}

	switch vendor {
	case enumor.TCloud:
		return a.tcloudCondSyncRes(cts, accountID, resName)
	case enumor.HuaWei:
		return a.huaweiCondSyncRes(cts, accountID, resName)
	case enumor.Aws:
		return a.awsCondSyncRes(cts, accountID, resName)
	case enumor.Azure:
		return a.azureCondSyncRes(cts, accountID, resName)
	default:
		return nil, fmt.Errorf("conditional sync not supports vendor: %s", vendor)
	}
}
