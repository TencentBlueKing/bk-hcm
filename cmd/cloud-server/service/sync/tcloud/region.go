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
	"errors"
	"time"

	"hcm/pkg/api/core"
	protohcregion "hcm/pkg/api/hc-service/region"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncRegion sync region
func SyncRegion(kt *kit.Kit, hcCli *hcservice.Client, accountID string) error {

	// 重新设置rid方便定位
	kt = kt.NewSubKit()

	start := time.Now()
	logs.V(3).Infof("tcloud account[%s] sync region start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("tcloud account[%s] sync region end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	req := &protohcregion.TCloudRegionSyncReq{
		AccountID: accountID,
	}
	if err := hcCli.TCloud.Region.Sync(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("sync tcloud region failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

// ListRegion ...
func ListRegion(kt *kit.Kit, dataCli *dataservice.Client) ([]string, error) {
	listReq := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := dataCli.TCloud.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list tcloud region failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errors.New("tcloud region is empty")
	}

	regions := make([]string, 0, len(result.Details))
	for _, one := range result.Details {
		regions = append(regions, one.RegionID)
	}

	return regions, nil
}
