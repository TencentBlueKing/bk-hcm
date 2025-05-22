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

	"hcm/cmd/cloud-server/service/sync/azure"
	"hcm/cmd/cloud-server/service/sync/lock"
	cloudaccount "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	resourcegroup "hcm/pkg/api/core/cloud/resource-group"
	protorg "hcm/pkg/api/data-service/cloud/resource-group"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	etcd3 "go.etcd.io/etcd/client/v3"
)

func (a *accountSvc) azureCondSyncRes(cts *rest.Contexts, accountID string, resType enumor.CloudResourceType) (
	any, error) {

	req, syncFunc, err := a.decodeAzureCondSyncRequest(cts, resType)
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
	syncParams := &azure.CondSyncParams{
		AccountID:          accountID,
		ResourceGroupNames: req.ResourceGroupNames,
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
				"cost: %s, rid: %s", err, enumor.Azure, resType, accountID, req, time.Since(startAt), cts.Kit.Rid)
			return
		}
		logs.Infof("[%s] conditional sync succeed on resource(%s), account: %s, req: %+v, cost: %s, rid: %s",
			enumor.Azure, resType, accountID, req, time.Since(startAt), cts.Kit.Rid)
	}(leaseID)

	return "started", nil
}

func (a *accountSvc) decodeAzureCondSyncRequest(cts *rest.Contexts, resType enumor.CloudResourceType) (
	*cloudaccount.AzureResCondSyncReq, azure.CondSyncFunc, error) {

	req := new(cloudaccount.AzureResCondSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	syncFunc, ok := azure.GetCondSyncFunc(resType)
	if !ok {
		return nil, nil, fmt.Errorf("azure conditional sync resource does not support %s", resType)
	}

	var rules []*filter.AtomRule
	if len(req.ResourceGroupNames) > 0 {
		rules = append(rules, tools.RuleIn("name", req.ResourceGroupNames))
	}

	// check resourceGroup
	regionListReq := &protorg.AzureRGListReq{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewDefaultBasePage(),
	}
	var regionList = make([]resourcegroup.AzureRG, 0, len(req.ResourceGroupNames))
	for {
		regionResult, err := a.client.DataService().Azure.ResourceGroup.ListResourceGroup(
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
	if len(req.ResourceGroupNames) > 0 && len(regionList) != len(req.ResourceGroupNames) {
		return nil, nil, errors.New("request regions mismatch regions on db")
	}

	return req, syncFunc, nil
}
