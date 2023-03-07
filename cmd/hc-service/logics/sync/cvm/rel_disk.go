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

// SyncCvmDiskRel sync cvm disk rel
func SyncCvmDiskRel(kt *kit.Kit, cloudDiskMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client) error {

	// change map key use hc id
	hcDiskMap := changCloudMapToHcMap(cloudDiskMap)

	hcInstanceIDs := make([]string, 0)
	for _, id := range hcDiskMap {
		hcInstanceIDs = append(hcInstanceIDs, id.HCInstanceID)
	}

	results, err := listCvmDiskRelSync(kt, hcInstanceIDs, dataCli)
	if err != nil {
		logs.Errorf("sync list disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	dsMap := make(map[string]uint64)
	if len(results.Details) > 0 {
		for _, detail := range results.Details {
			id := getCVMRelID(detail.DiskID, detail.CvmID)
			dsMap[id] = detail.ID
		}
	}

	for k := range dsMap {
		if _, ok := hcDiskMap[k]; ok {
			delete(dsMap, k)
			delete(hcDiskMap, k)
		}
	}

	if len(hcDiskMap) > 0 {
		err := addCvmDiskRelSync(kt, hcDiskMap, dataCli)
		if err != nil {
			logs.Errorf("sync add disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(dsMap) > 0 {
		err := deleteCvmDiskRelSync(kt, dsMap, dataCli)
		if err != nil {
			logs.Errorf("sync delete disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func listCvmDiskRelSync(kt *kit.Kit, instanceIDs []string,
	dataCli *dataservice.Client) (*dataproto.DiskCvmRelListResult, error) {

	listReq := &dataproto.DiskCvmRelListReq{
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

	results, err := dataCli.Global.ListDiskCvmRel(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("from data-service list disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results, nil
}

func addCvmDiskRelSync(kt *kit.Kit, addMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client) error {

	lists := make([]dataproto.DiskCvmRelCreateReq, 0)

	for _, id := range addMap {
		rel := dataproto.DiskCvmRelCreateReq{
			DiskID: id.HCRelID,
			CvmID:  id.HCInstanceID,
		}
		lists = append(lists, rel)
	}

	createReq := &dataproto.DiskCvmRelBatchCreateReq{
		Rels: lists,
	}

	err := dataCli.Global.BatchCreateDiskCvmRel(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("from data-service add disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func deleteCvmDiskRelSync(kt *kit.Kit, deleteMap map[string]uint64,
	dataCli *dataservice.Client) error {

	ids := make([]uint64, 0)
	for _, id := range deleteMap {
		ids = append(ids, id)
	}

	batchDeleteReq := &dataproto.DiskCvmRelDeleteReq{
		Filter: tools.ContainersExpression("id", ids),
	}

	err := dataCli.Global.DeleteDiskCvmRel(kt.Ctx, kt.Header(), batchDeleteReq)
	if err != nil {
		logs.Errorf("from data-service delete disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
