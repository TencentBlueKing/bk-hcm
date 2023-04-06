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

package cvm

import (
	"hcm/pkg/api/core"
	protods "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncCvmSGRel sync cvm sg rel
func SyncCvmSGRel(kt *kit.Kit, cloudSGMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client, account_id string, hcCloudIDs []string) error {

	// change map key use hc id
	hcSGMap := changCloudMapToHcMap(cloudSGMap)

	hcIDs, err := getCvmHCIDs(kt, account_id, hcCloudIDs, dataCli)
	if err != nil {
		logs.Errorf("request getCvmHCIDs to get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	results, err := listCvmSGRelSync(kt, hcIDs, dataCli)
	if err != nil {
		logs.Errorf("sync list sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	dsMap := make(map[string]uint64)
	if len(results.Details) > 0 {
		for _, detail := range results.Details {
			id := getCVMRelID(detail.SecurityGroupID, detail.CvmID)
			dsMap[id] = detail.ID
		}
	}

	for k := range dsMap {
		if _, ok := hcSGMap[k]; ok {
			delete(dsMap, k)
			delete(hcSGMap, k)
		}
	}

	if len(hcSGMap) > 0 {
		err := addCvmSGRelSync(kt, hcSGMap, dataCli)
		if err != nil {
			logs.Errorf("sync add sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(dsMap) > 0 {
		err := deleteCvmSGRelSync(kt, dsMap, dataCli)
		if err != nil {
			logs.Errorf("sync delete sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func listCvmSGRelSync(kt *kit.Kit, instanceIDs []string,
	dataCli *dataservice.Client) (*dataproto.SGCvmRelListResult, error) {

	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cvm_id",
					Op:    filter.In.Factory(),
					Value: instanceIDs,
				},
			},
		},
		Page: core.DefaultBasePage,
	}

	results, err := dataCli.Global.SGCvmRel.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("from data-service list sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results, nil
}

func addCvmSGRelSync(kt *kit.Kit, addMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client) error {

	lists := make([]dataproto.SGCvmRelCreate, len(addMap))

	count := 0
	for _, id := range addMap {
		lists[count] = dataproto.SGCvmRelCreate{
			SecurityGroupID: id.HCRelID,
			CvmID:           id.HCInstanceID,
		}
		count++
	}

	createReq := &dataproto.SGCvmRelBatchCreateReq{
		Rels: lists,
	}

	err := dataCli.Global.SGCvmRel.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("from data-service add sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func deleteCvmSGRelSync(kt *kit.Kit, deleteMap map[string]uint64,
	dataCli *dataservice.Client) error {

	ids := make([]uint64, 0)
	for _, id := range deleteMap {
		ids = append(ids, id)
	}

	batchDeleteReq := &protods.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", ids),
	}

	err := dataCli.Global.SGCvmRel.BatchDelete(kt.Ctx, kt.Header(), batchDeleteReq)
	if err != nil {
		logs.Errorf("from data-service delete sg cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
