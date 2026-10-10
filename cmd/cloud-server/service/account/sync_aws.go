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
	"time"

	"hcm/cmd/cloud-server/service/sync/aws"
	"hcm/cmd/cloud-server/service/sync/lock"
	cloudaccount "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

func (a *accountSvc) awsCondSyncRes(cts *rest.Contexts, accountID string, resType enumor.CloudResourceType) (
	any, error) {

	req, syncFunc, err := a.decodeAwsCondSyncRequest(cts, accountID, resType)
	if err != nil {
		return nil, err
	}

	resLockKey := lock.Key(fmt.Sprintf("%s", accountID))
	leaseID, err := lock.Manager.TryLock(resLockKey)
	if err != nil {
		if errors.Is(err, lock.ErrLockFailed) {
			return nil, errf.New(errf.SyncRepeatLockError, "synchronization is in progress")
		}
		return nil, err
	}

	defer func() {
		if err = lock.Manager.UnLock(leaseID); err != nil {
			// 锁已经超时释放了
			if strings.Contains(err.Error(), "requested lease not found") {
				return
			}

			logs.Errorf("[%s]: unlock account sync lock for cond sync failed, err: %v, account: %s, leaseID: %d, "+
				"resType: %s, rid: %s", constant.AccountSyncFailed, err, accountID, leaseID, resType, cts.Kit.Rid)
		}
		logs.Infof("unlock account sync key: %s, resType: %s, rid: %s", resLockKey, resType, cts.Kit.Rid)
	}()

	logs.Infof("lock account sync key: %s, resType: %s, rid: %s", resLockKey, resType, cts.Kit.Rid)
	syncParams := &aws.CondSyncParams{
		AccountID: accountID,
		Regions:   req.Regions,
		CloudIDs:  req.CloudIDs,
	}

	startAt := time.Now()
	kt := cts.Kit.NewSubKit()
	// 设置超时控制
	cancel := kt.CtxWithTimeoutMS(int(AccountSyncDefaultTimeout / time.Millisecond))
	defer cancel()
	err = syncFunc(kt, a.client, syncParams)
	if err != nil {
		logs.Errorf("[%s] conditional sync failed on resource(%s), err: %v, account: %s, req: %+v, "+
			"cost: %s, rid: %s", enumor.Aws, resType, err, accountID, req, time.Since(startAt), cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("[%s] conditional sync succeed on resource(%s), account: %s, req: %+v, resType: %s, "+
		"cost: %s, rid: %s", enumor.Aws, resType, accountID, req, resType, time.Since(startAt), cts.Kit.Rid)

	return nil, nil
}

func (a *accountSvc) decodeAwsCondSyncRequest(cts *rest.Contexts, accountID string,
	resType enumor.CloudResourceType) (*cloudaccount.ResCondSyncReq, aws.CondSyncFunc, error) {

	req := new(cloudaccount.ResCondSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, err
	}

	// 部分资源允许全量同步（不要求必须提供regions参数）
	allowedSyncAllRes := resType.IsAllowedSyncAll()
	// 不允许全量同步的资源类型需要 regions 参数
	// allowedSyncAllRes 为 false 时，needRegion 为 true，表示需要 regions 参数
	if err := req.Validate(!allowedSyncAllRes); err != nil {
		return nil, nil, err
	}

	syncFunc, ok := aws.GetCondSyncFunc(resType)
	if !ok {
		return nil, nil, fmt.Errorf("aws conditional sync resource does not support %s", resType)
	}

	// IN 查询是集合语义，数量校验必须用去重后的地域，否则重复入参会被误判为不存在。
	req.Regions = slice.Unique(req.Regions)

	rules := buildAwsCondSyncRegionRules(accountID, req.Regions)

	// check region
	regionListReq := &core.ListReq{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewDefaultBasePage(),
	}
	var regionList = make([]region.AwsRegion, 0, len(req.Regions))
	for {
		regionResult, err := a.client.DataService().Aws.Region.ListRegion(
			cts.Kit.Ctx, cts.Kit.Header(), regionListReq)
		if err != nil {
			return nil, nil, err
		}
		regionList = append(regionList, regionResult.Details...)
		if uint(len(regionResult.Details)) < regionListReq.Page.Limit {
			break
		}
		regionListReq.Page.Start += uint32(regionListReq.Page.Limit)
	}
	if len(req.Regions) > 0 {
		if err := checkAwsRequestRegions(req.Regions, regionList); err != nil {
			return nil, nil, err
		}
	}
	return req, syncFunc, nil
}

// buildAwsCondSyncRegionRules 构造条件同步的地域查询条件。
// 指定地域时不按 sync_enable 过滤，便于区分不存在和已禁用；未指定地域时只查启用同步的地域。
func buildAwsCondSyncRegionRules(accountID string, regions []string) []*filter.AtomRule {
	rules := []*filter.AtomRule{tools.RuleEqual("account_id", accountID)}
	if len(regions) > 0 {
		rules = append(rules, tools.RuleIn("region_id", regions))
		return rules
	}
	rules = append(rules, tools.RuleEqual("sync_enable", true))
	return rules
}

// checkAwsRequestRegions 区分请求地域不存在和同步已禁用。
func checkAwsRequestRegions(requestRegions []string, regionList []region.AwsRegion) error {
	found := make(map[string]bool, len(regionList))
	for _, one := range regionList {
		found[one.RegionID] = one.SyncEnable
	}

	missing := make([]string, 0)
	disabled := make([]string, 0)
	for _, regionID := range requestRegions {
		syncEnable, ok := found[regionID]
		if !ok {
			missing = append(missing, regionID)
			continue
		}
		if !syncEnable {
			disabled = append(disabled, regionID)
		}
	}

	switch {
	case len(missing) > 0 && len(disabled) > 0:
		return errf.Newf(errf.InvalidParameter,
			"some request regions don't exist: %v, sync is disabled: %v", missing, disabled)
	case len(missing) > 0:
		return errf.Newf(errf.InvalidParameter, "some request regions don't exist: %v", missing)
	case len(disabled) > 0:
		return errf.Newf(errf.InvalidParameter, "some request regions sync is disabled: %v", disabled)
	default:
		return nil
	}
}
