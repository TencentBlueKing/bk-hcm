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
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	etcd3 "go.etcd.io/etcd/client/v3"
)

func (a *accountSvc) awsCondSyncRes(cts *rest.Contexts, accountID string, resType enumor.CloudResourceType) (
	any, error) {

	req, syncFunc, err := a.decodeAwsCondSyncRequest(cts, accountID, resType)
	if err != nil {
		return nil, err
	}

	leaseID, err := lock.Manager.TryLock(lock.Key(accountID))
	if err != nil {
		if err == lock.ErrLockFailed {
			return nil, errors.New("synchronization is in progress")
		}
		return nil, err
	}
	logs.Infof("lock account sync key: %s, rid: %s", lock.Key(accountID), cts.Kit.Rid)
	syncParams := &aws.CondSyncParams{
		AccountID: accountID,
		Regions:   req.Regions,
		CloudIDs:  req.CloudIDs,
	}
	startAt := time.Now()
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
			logs.Infof("unlock account sync key: %s, rid: %s", lock.Key(accountID), cts.Kit.Rid)

		}()

		err = syncFunc(cts.Kit, a.client, syncParams)
		if err != nil {
			logs.Errorf("[%s] conditional sync failed on resource(%s), err: %v, account: %s, req: %+v, "+
				"cost: %s, rid: %s", err, enumor.Aws, resType, accountID, req, time.Since(startAt), cts.Kit.Rid)
			return
		}
		logs.Infof("[%s] conditional sync succeed on resource(%s), account: %s, req: %+v, cost: %s, rid: %s",
			enumor.Aws, resType, accountID, req, time.Since(startAt), cts.Kit.Rid)
	}(leaseID)

	return "started", nil
}

func (a *accountSvc) decodeAwsCondSyncRequest(cts *rest.Contexts, accountID string,
	resType enumor.CloudResourceType) (*cloudaccount.ResCondSyncReq, aws.CondSyncFunc, error) {

	req := new(cloudaccount.ResCondSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	syncFunc, ok := aws.GetCondSyncFunc(resType)
	if !ok {
		return nil, nil, fmt.Errorf("aws conditional sync resource does not support %s", resType)
	}

	var rules []*filter.AtomRule
	rules = append(rules, tools.RuleEqual("account_id", accountID))
	if len(req.Regions) > 0 {
		rules = append(rules, tools.RuleIn("region_id", req.Regions))
	}

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
	if len(req.Regions) > 0 && len(regionList) != len(req.Regions) {
		return nil, nil, errors.New("request regions mismatch regions on db")
	}
	req.Regions = slice.Unique(req.Regions)
	return req, syncFunc, nil
}
