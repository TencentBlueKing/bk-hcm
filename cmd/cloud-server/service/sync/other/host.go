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

package other

import (
	"fmt"
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// SyncHost sync host
func SyncHost(kt *kit.Kit, cliSet *client.ClientSet, accountID string, sd *detail.SyncDetail) error {
	// 重新设置rid方便定位
	kt = kt.NewSubKit()

	start := time.Now()
	logs.V(3).Infof("other account[%s] sync host start, time: %v, rid: %s", accountID, start, kt.Rid)

	// 同步详情同步中
	if err := sd.ResSyncStatusSyncing(enumor.CvmCloudResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("other account[%s] sync host end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	bizIDs, err := listBizIDsContainsHostPool(kt)
	if err != nil {
		logs.Errorf("list biz ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, bizID := range bizIDs {
		subStart := time.Now()
		logs.Infof("start sync biz(%d) host, time: %v, rid: %s", bizID, subStart, kt.Rid)

		req := &sync.OtherSyncHostReq{
			AccountID: accountID,
			BizID:     bizID,
		}
		if err = cliSet.HCService().Other.Host.SyncHostWithRelResource(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("sync other host failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			return err
		}

		subEnd := time.Now()
		logs.Infof("sync biz(%d) host success, time: %v, cost: %v, rid: %s", bizID, subEnd, subEnd.Sub(subStart),
			kt.Rid)
	}

	// 同步详情同步成功
	if err := sd.ResSyncStatusSuccess(enumor.CvmCloudResType); err != nil {
		return err
	}

	return nil
}

func listBizIDsContainsHostPool(kt *kit.Kit) ([]int64, error) {
	params := &cmdb.SearchBizParams{Fields: []string{"bk_biz_id"}}
	resp, err := cmdb.CmdbClient().SearchBusiness(kt, params)
	if err != nil {
		logs.Errorf("search business from cc failed, err: %v, param: %+v, rid: %s", err, params, kt.Rid)
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	bizIDs := make([]int64, 0)
	for _, biz := range resp.Info {
		bizIDs = append(bizIDs, biz.BizID)
	}

	bizIDs = append(bizIDs, constant.HostPoolBiz)

	return bizIDs, nil
}
