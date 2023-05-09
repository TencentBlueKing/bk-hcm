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

// SyncCvmNetworkInterfaceRel sync cvm networkinterface rel
func SyncCvmNetworkInterfaceRel(kt *kit.Kit, cloudSGMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client, accountID string, hcCloudIDs []string) error {

	// change map key use hc id
	hcSGMap := changCloudMapToHcMap(cloudSGMap)

	hcIDs, err := getCvmHCIDs(kt, accountID, hcCloudIDs, dataCli)
	if err != nil {
		logs.Errorf("request getCvmHCIDs to get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	results, err := listCvmNetworkInterfaceRelSync(kt, hcIDs, dataCli)
	if err != nil {
		logs.Errorf("sync list networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	dsMap := make(map[string]uint64)
	if len(results.Details) > 0 {
		for _, detail := range results.Details {
			id := getCVMRelID(detail.NetworkInterfaceID, detail.CvmID)
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
		err := addCvmNetworkInterfaceRelSync(kt, hcSGMap, dataCli)
		if err != nil {
			logs.Errorf("sync add networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(dsMap) > 0 {
		err := deleteCvmNetworkInterfaceRelSync(kt, dsMap, dataCli)
		if err != nil {
			logs.Errorf("sync delete networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func listCvmNetworkInterfaceRelSync(kt *kit.Kit, instanceIDs []string,
	dataCli *dataservice.Client) (*dataproto.NetworkInterfaceCvmRelListResult, error) {

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

	results, err := dataCli.Global.NetworkInterfaceCvmRel.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("from data-service list networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results, nil
}

func addCvmNetworkInterfaceRelSync(kt *kit.Kit, addMap map[string]*CVMOperateSync,
	dataCli *dataservice.Client) error {

	lists := make([]dataproto.NetworkInterfaceCvmRelCreateReq, len(addMap))

	count := 0
	for _, id := range addMap {
		lists[count] = dataproto.NetworkInterfaceCvmRelCreateReq{
			NetworkInterfaceID: id.HCRelID,
			CvmID:              id.HCInstanceID,
		}
		count++
	}

	createReq := &dataproto.NetworkInterfaceCvmRelBatchCreateReq{
		Rels: lists,
	}

	err := dataCli.Global.NetworkInterfaceCvmRel.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("from data-service add networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func deleteCvmNetworkInterfaceRelSync(kt *kit.Kit, deleteMap map[string]uint64,
	dataCli *dataservice.Client) error {

	ids := make([]uint64, 0)
	for _, id := range deleteMap {
		ids = append(ids, id)
	}

	batchDeleteReq := &protods.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", ids),
	}

	err := dataCli.Global.NetworkInterfaceCvmRel.BatchDelete(kt.Ctx, kt.Header(), batchDeleteReq)
	if err != nil {
		logs.Errorf("from data-service delete networkinterface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
