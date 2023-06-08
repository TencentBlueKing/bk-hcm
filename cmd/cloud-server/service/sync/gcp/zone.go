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

package gcp

import (
	"errors"
	"time"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/api/hc-service/zone"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncZone sync zone
func SyncZone(kt *kit.Kit, hcCli *hcservice.Client, accountID string) error {

	start := time.Now()
	logs.V(3).Infof("gcp account[%s] sync zone start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("gcp account[%s] sync zone end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	syncReq := &zone.GcpZoneSyncReq{
		AccountID: accountID,
	}
	if err := hcCli.Gcp.Zone.SyncZone(kt.Ctx, kt.Header(), syncReq); err != nil {
		logs.Errorf("sync gcp zone failed, err: %v, req: %v, rid: %s", err, syncReq, kt.Rid)
		return err
	}

	return nil
}

// GetRegionZoneMap ...
func GetRegionZoneMap(kt *kit.Kit, dataCli *dataservice.Client) (map[string][]string, error) {
	listReq := &protocloud.ZoneListReq{
		Filter: tools.EqualExpression("vendor", enumor.Gcp),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := dataCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list gcp zone failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errors.New("gcp zone is empty")
	}

	regionZoneMap := make(map[string][]string)
	for _, one := range result.Details {
		if _, exist := regionZoneMap[one.Region]; !exist {
			regionZoneMap[one.Region] = make([]string, 0)
		}

		regionZoneMap[one.Region] = append(regionZoneMap[one.Region], one.Name)
	}

	return regionZoneMap, nil
}
