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

package eip

import (
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	protoeip "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// TCloudEipSync
type TCloudEipSync struct {
	IsUpdate bool
	Eip      *eip.TCloudEip
}

// TCloudDSEipSync ...
type TCloudDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.TCloudEipExtensionResult]
}

// AwsEipSync
type AwsEipSync struct {
	IsUpdate bool
	Eip      *eip.AwsEip
}

// AwsDSEipSync
type AwsDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.AwsEipExtensionResult]
}

// HuaWeiEipSync
type HuaWeiEipSync struct {
	IsUpdate bool
	Eip      *eip.HuaWeiEip
}

// HuaWeiDSEipSync
type HuaWeiDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.HuaWeiEipExtensionResult]
}

// GcpEipSync
type GcpEipSync struct {
	IsUpdate bool
	Eip      *eip.GcpEip
}

// GcpDSEipSync
type GcpDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.GcpEipExtensionResult]
}

// AzureEipSync
type AzureEipSync struct {
	IsUpdate bool
	Eip      *eip.AzureEip
}

// AzureDSEipSync
type AzureDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.AzureEipExtensionResult]
}

// GetHcEipDatas get eip datas from hc
func GetHcEipDatas(kt *kit.Kit, req *protoeip.EipSyncReq,
	dataCli *dataservice.Client) (map[string]*dataproto.EipResult, error) {

	dsMap := make(map[string]*dataproto.EipResult)

	start := 0
	for {
		dataReq := &dataproto.EipListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsMap[detail.CloudID] = detail
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsMap, nil
}

func syncEipDelete(kt *kit.Kit, deleteCloudIDs []string,
	dataCli *dataservice.Client) error {

	batchDeleteReq := &dataproto.EipDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}

	if _, err := dataCli.Global.DeleteEip(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
