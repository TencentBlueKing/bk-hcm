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

package cvmrelmgr

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	datacloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

func (mgr *CvmRelManger) syncCvmNetworkInterfaceRel(kt *kit.Kit, cvmMap map[string]string, opt *SyncRelOption) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	niMap, err := mgr.getNetworkInterfaceMap(kt)
	if err != nil {
		logs.Errorf("get network interface map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmIDs, cvmRelMapFromCloud, err := mgr.getCvmIDWithAssResIDMap(enumor.NetworkInterfaceCloudResType, cvmMap, niMap)
	if err != nil {
		logs.Errorf("get cvm id with ass res id map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmRelMapFromDB, err := mgr.getCvmNetworkInterfaceRelMapFromDB(kt, cvmIDs)
	if err != nil {
		logs.Errorf("get cvm_ni_rel map from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	addRels, delIDs := diffCvmWithAssResRel(cvmRelMapFromCloud, cvmRelMapFromDB)

	if len(addRels) > 0 {
		if err = mgr.createCvmNetworkInterfaceRel(kt, addRels); err != nil {
			return err
		}
	}

	if len(delIDs) > 0 {
		if err = mgr.deleteCvmNetworkInterfaceRel(kt, delIDs); err != nil {
			return err
		}
	}

	return nil
}

func (mgr *CvmRelManger) deleteCvmNetworkInterfaceRel(kt *kit.Kit, ids []uint64) error {

	split := slice.Split(ids, constant.BatchOperationMaxLimit)
	for _, partIDs := range split {
		batchDeleteReq := &dataproto.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", partIDs),
		}

		if err := mgr.dataCli.Global.NetworkInterfaceCvmRel.BatchDelete(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
			logs.Errorf("batch delete ni_cvm_rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("delete cvm network interface rel success, count: %d, rid: %s", len(ids), kt.Rid)

	return nil
}

func (mgr *CvmRelManger) createCvmNetworkInterfaceRel(kt *kit.Kit, addRels []cvmRelInfo) error {
	split := slice.Split(addRels, constant.BatchOperationMaxLimit)

	for _, part := range split {
		lists := make([]datacloud.NetworkInterfaceCvmRelCreateReq, 0)
		for _, one := range part {
			rel := datacloud.NetworkInterfaceCvmRelCreateReq{
				NetworkInterfaceID: one.AssResID,
				CvmID:              one.CvmID,
			}
			lists = append(lists, rel)
		}

		createReq := &datacloud.NetworkInterfaceCvmRelBatchCreateReq{
			Rels: lists,
		}

		if err := mgr.dataCli.Global.NetworkInterfaceCvmRel.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create ni_cvm_rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("create cvm network interface rel success, count: %d, rid: %s", len(addRels), kt.Rid)

	return nil
}

func (mgr *CvmRelManger) getCvmNetworkInterfaceRelMapFromDB(kt *kit.Kit, cvmIDs []string) (
	map[string]map[string]cvmRelInfo, error) {

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIDs),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	result := make(map[string]map[string]cvmRelInfo)
	for {
		respResult, err := mgr.dataCli.Global.NetworkInterfaceCvmRel.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list network interface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, rel := range respResult.Details {
			if _, exist := result[rel.CvmID]; !exist {
				result[rel.CvmID] = make(map[string]cvmRelInfo)
			}

			result[rel.CvmID][rel.NetworkInterfaceID] = cvmRelInfo{
				RelID:    rel.ID,
				AssResID: rel.NetworkInterfaceID,
			}
		}

		if len(respResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}

func (mgr *CvmRelManger) getNetworkInterfaceMap(kt *kit.Kit) (map[string]string, error) {
	cloudIDs := mgr.getAllCvmAssResCloudIDs(enumor.NetworkInterfaceCloudResType)

	niMap := make(map[string]string)
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, partCloudIDs := range split {
		req := &core.ListReq{
			Fields: []string{"id", "cloud_id"},
			Filter: tools.ContainersExpression("cloud_id", partCloudIDs),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := mgr.dataCli.Global.NetworkInterface.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list network interface failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			niMap[one.CloudID] = one.ID
		}
	}

	return niMap, nil
}
