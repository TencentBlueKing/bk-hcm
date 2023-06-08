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
	dataproto "hcm/pkg/api/data-service/cloud"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

func (mgr *CvmRelManger) syncCvmDiskRel(kt *kit.Kit, cvmMap map[string]string, opt *SyncRelOption) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	diskMap, err := mgr.getDiskMap(kt)
	if err != nil {
		logs.Errorf("get disk map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmIDs, cvmRelMapFromCloud, err := mgr.getCvmIDWithAssResIDMap(enumor.DiskCloudResType, cvmMap, diskMap)
	if err != nil {
		logs.Errorf("get cvm id with ass res id map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmRelMapFromDB, err := mgr.getCvmDiskRelMapFromDB(kt, cvmIDs)
	if err != nil {
		logs.Errorf("get cvm_disk_rel map from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	addRels, delIDs := diffCvmWithAssResRel(cvmRelMapFromCloud, cvmRelMapFromDB)

	if len(addRels) > 0 {
		if err = mgr.createCvmDiskRel(kt, addRels); err != nil {
			return err
		}
	}

	if len(delIDs) > 0 {
		if err = mgr.deleteCvmDiskRel(kt, delIDs); err != nil {
			return err
		}
	}

	return nil
}

func (mgr *CvmRelManger) deleteCvmDiskRel(kt *kit.Kit, ids []uint64) error {

	split := slice.Split(ids, constant.BatchOperationMaxLimit)
	for _, partIDs := range split {
		batchDeleteReq := &dataproto.DiskCvmRelDeleteReq{
			Filter: tools.ContainersExpression("id", partIDs),
		}

		if err := mgr.dataCli.Global.DeleteDiskCvmRel(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
			logs.Errorf("batch delete disk_cvm_rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("delete cvm disk rel success, count: %d, rid: %s", len(ids), kt.Rid)

	return nil
}

func (mgr *CvmRelManger) createCvmDiskRel(kt *kit.Kit, addRels []cvmRelInfo) error {
	split := slice.Split(addRels, constant.BatchOperationMaxLimit)

	for _, part := range split {
		lists := make([]dataproto.DiskCvmRelCreateReq, 0)
		for _, one := range part {
			rel := dataproto.DiskCvmRelCreateReq{
				DiskID: one.AssResID,
				CvmID:  one.CvmID,
			}
			lists = append(lists, rel)
		}

		createReq := &dataproto.DiskCvmRelBatchCreateReq{
			Rels: lists,
		}

		if err := mgr.dataCli.Global.BatchCreateDiskCvmRel(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create disk_cvm_rel failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("create cvm disk rel success, count: %d, rid: %s", len(addRels), kt.Rid)

	return nil
}

func (mgr *CvmRelManger) getCvmDiskRelMapFromDB(kt *kit.Kit, cvmIDs []string) (
	map[string]map[string]cvmRelInfo, error) {

	listReq := &dataproto.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIDs),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
		Fields: nil,
	}
	result := make(map[string]map[string]cvmRelInfo)
	for {
		respResult, err := mgr.dataCli.Global.ListDiskCvmRel(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, rel := range respResult.Details {
			if _, exist := result[rel.CvmID]; !exist {
				result[rel.CvmID] = make(map[string]cvmRelInfo)
			}

			result[rel.CvmID][rel.DiskID] = cvmRelInfo{
				RelID:    rel.ID,
				AssResID: rel.DiskID,
			}
		}

		if len(respResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}

func (mgr *CvmRelManger) getDiskMap(kt *kit.Kit) (map[string]string, error) {
	cloudIDs := mgr.getAllCvmAssResCloudIDs(enumor.DiskCloudResType)

	diskMap := make(map[string]string)
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, partCloudIDs := range split {
		req := &datadisk.DiskListReq{
			Fields: []string{"id", "cloud_id"},
			Filter: tools.ContainersExpression("cloud_id", partCloudIDs),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := mgr.dataCli.Global.ListDisk(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			diskMap[one.CloudID] = one.ID
		}
	}

	return diskMap, nil
}
