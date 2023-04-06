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
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncCvmEipRel sync cvm eip rel
func SyncCvmEipRel(kt *kit.Kit, cloudEipMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client, account_id string, hcCloudIDs []string) error {

	// change map key use hc id
	hcEipMap := changCloudMapToHcMap(cloudEipMap)

	hcIDs, err := getCvmHCIDs(kt, account_id, hcCloudIDs, dataCli)
	if err != nil {
		logs.Errorf("request getCvmHCIDs to get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	results, err := listCvmEipRelSync(kt, hcIDs, dataCli)
	if err != nil {
		logs.Errorf("sync list eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	dsMap := make(map[string]uint64)
	if len(results.Details) > 0 {
		for _, detail := range results.Details {
			id := getCVMRelID(detail.EipID, detail.CvmID)
			dsMap[id] = detail.ID
		}
	}

	for k := range dsMap {
		if _, ok := hcEipMap[k]; ok {
			delete(dsMap, k)
			delete(hcEipMap, k)
		}
	}

	if len(hcEipMap) > 0 {
		err := addCvmEipRelSync(kt, hcEipMap, dataCli)
		if err != nil {
			logs.Errorf("sync add eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(dsMap) > 0 {
		err := deleteCvmEipRelSync(kt, dsMap, dataCli)
		if err != nil {
			logs.Errorf("sync delete eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func listCvmEipRelSync(kt *kit.Kit, instanceIDs []string,
	dataCli *dataservice.Client) (*dataproto.EipCvmRelListResult, error) {

	listReq := &dataproto.EipCvmRelListReq{
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

	results, err := dataCli.Global.ListEipCvmRel(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("from data-service list eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results, nil
}

func addCvmEipRelSync(kt *kit.Kit, addMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client) error {

	lists := make([]dataproto.EipCvmRelCreateReq, 0)

	for _, id := range addMap {
		rel := dataproto.EipCvmRelCreateReq{
			EipID: id.HCRelID,
			CvmID: id.HCInstanceID,
		}
		lists = append(lists, rel)
	}

	createReq := &dataproto.EipCvmRelBatchCreateReq{
		Rels: lists,
	}

	err := dataCli.Global.BatchCreateEipCvmRel(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("from data-service add eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func deleteCvmEipRelSync(kt *kit.Kit, deleteMap map[string]uint64,
	dataCli *dataservice.Client) error {

	ids := make([]uint64, 0)
	for _, id := range deleteMap {
		ids = append(ids, id)
	}

	batchDeleteReq := &dataproto.EipCvmRelDeleteReq{
		Filter: tools.ContainersExpression("id", ids),
	}

	err := dataCli.Global.DeleteEipCvmRel(kt.Ctx, kt.Header(), batchDeleteReq)
	if err != nil {
		logs.Errorf("from data-service delete eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
