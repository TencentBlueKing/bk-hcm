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

package networkinterface

import (
	"fmt"

	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	protocvm "hcm/pkg/api/hc-service/cvm"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// GetHcNetworkInterfaceDatas get networkinterface datas from hc
func GetHcNetworkInterfaceDatas(kt *kit.Kit, req *protocvm.OperateSyncReq,
	dataCli *dataservice.Client) (map[string]coreni.BaseNetworkInterface, error) {

	dsMap := make(map[string]coreni.BaseNetworkInterface)

	page := uint32(0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: req.AccountID,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dbQueryReq.Filter.Rules = append(dbQueryReq.Filter.Rules, filter)
		}

		dbList, err := dataCli.Global.NetworkInterface.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("networkinterface batch get networkinterfacelist db error, err: %v", err)
			return nil, err
		}

		for _, detail := range dbList.Details {
			dsMap[detail.CloudID] = detail
		}

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}

	if len(dsMap) == 0 {
		return nil, fmt.Errorf("network interface not found")
	}

	return dsMap, nil
}
