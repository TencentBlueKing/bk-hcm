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
	"strings"

	"hcm/cmd/cloud-server/logics/account"
	"hcm/cmd/cloud-server/service/sync/lock"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	cloudaccount "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// SyncCloudResource ...
func (a *accountSvc) SyncCloudResource(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, meta.Update, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, accountID)
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

	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return a.tcloudCondSyncRes(cts, accountID, resName)

	default:
		return nil, fmt.Errorf("conditional sync not supports vendor: %s", vendor)
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
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if baseInfo.Vendor != vendor {
		return nil, errf.Newf(errf.InvalidParameter, "account not found by vendor: %s", vendor)
	}

	if baseInfo.BkBizID != bkBizId {
		return nil, errors.New("account not found by biz")
	}

	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return a.tcloudCondSyncRes(cts, accountID, resName)

	default:
		return nil, fmt.Errorf("conditional sync not supports vendor: %s", vendor)
	}
}

func (a *accountSvc) tcloudCondSyncRes(cts *rest.Contexts, accountID string, resName enumor.CloudResourceType) (
	any, error) {

	req := &cloudaccount.TCloudResCondSyncReq{}
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	syncFunc, ok := tcloud.GetCondSyncFunc(resName)
	if !ok {
		return nil, fmt.Errorf("tcloud conditional sync resource not supported: %s", resName)
	}

	var rules []*filter.AtomRule
	if len(req.Regions) > 0 {
		rules = append(rules, tools.RuleIn("region_id", req.Regions))
	}

	// check region
	regionListReq := &core.ListReq{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	regionResult, err := a.client.DataService().TCloud.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), regionListReq)
	if err != nil {
		return nil, err
	}
	if len(req.Regions) > 0 && len(regionResult.Details) != len(req.Regions) {
		return nil, errors.New("request regions not match regions on db")
	}
	req.Regions = slice.Map(regionResult.Details, region.TCloudRegion.GetCloudID)

	leaseID, err := lock.Manager.TryLock(lock.Key(accountID))
	if err != nil {
		if err == lock.ErrLockFailed {
			return nil, errors.New("synchronization is in progress")
		}
		return nil, err
	}
	syncParams := &tcloud.CondSyncParams{
		AccountID:  accountID,
		Regions:    req.Regions,
		CloudIDs:   req.CloudIDs,
		TagFilters: req.TagFilters,
	}

	go func(leaseID etcd3.LeaseID) {
		defer func() {
			if err := lock.Manager.UnLock(leaseID); err != nil {
				// 锁已经超时释放了
				if strings.Contains(err.Error(), "requested lease not found") {
					return
				}

				logs.Errorf("[%s]: unlock account sync lock for cond sync failed, "+
					"err: %v, account: %s, leaseID: %d, rid: %s",
					constant.AccountSyncFailed, err, accountID, leaseID, cts.Kit.Rid)
			}
		}()

		err = syncFunc(cts.Kit, a.client, syncParams)
		if err != nil {
			logs.Errorf("[%s] failed to perform conditional syncing on resource(%s), account: %s, req: %+v, rid: %s",
				enumor.TCloud, resName, accountID, req, cts.Kit.Rid)
		}

	}(leaseID)

	return nil, err
}
