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

package disk

import (
	"hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// decodeDiskSyncReq get par from body
func (da *diskAdaptor) decodeDiskSyncReq(cts *rest.Contexts) (*protodisk.DiskSyncReq, error) {

	req := new(protodisk.DiskSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return req, nil
}

// getDatasFromDSForDiskSync get datas from data-service
func (da *diskAdaptor) getDatasFromDSForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq) (map[string]*protodisk.DiskSyncDS, error) {

	start := 0
	resultsHcm := make([]*datadisk.DiskResult, 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		results, err := da.dataCli.Global.ListDisk(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*protodisk.DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(protodisk.DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// diffSecurityGroupSyncDelete for delete
func (da *diskAdaptor) diffDiskSyncDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	batchDeleteReq := &datadisk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if _, err := da.dataCli.Global.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

// getAddCloudIDs
func getAddCloudIDs[T any](cloudMap map[string]T, dsMap map[string]*protodisk.DiskSyncDS) []string {

	addCloudIDs := []string{}
	for id := range cloudMap {
		if _, ok := dsMap[id]; !ok {
			addCloudIDs = append(addCloudIDs, id)
		} else {
			dsMap[id].IsUpdated = true
		}
	}

	return addCloudIDs
}

// getDeleteAndUpdateCloudIDs
func getDeleteAndUpdateCloudIDs(dsMap map[string]*protodisk.DiskSyncDS) ([]string, []string) {

	deleteCloudIDs := []string{}
	updateCloudIDs := []string{}
	for id, one := range dsMap {
		if !one.IsUpdated {
			deleteCloudIDs = append(deleteCloudIDs, id)
		} else {
			updateCloudIDs = append(updateCloudIDs, id)
		}
	}

	return deleteCloudIDs, updateCloudIDs
}
