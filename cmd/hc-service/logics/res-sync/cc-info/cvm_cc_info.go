/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package ccinfo ...
package ccinfo

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/api/data-service/cloud/network-interface"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// CvmCCInfoRelManager ...
type CvmCCInfoRelManager struct {
	dbCli *dataservice.Client
}

// NewCvmCCInfoRelManager ...
func NewCvmCCInfoRelManager(dbCli *dataservice.Client) *CvmCCInfoRelManager {
	return &CvmCCInfoRelManager{
		dbCli: dbCli,
	}
}

// SyncCvmCCInfo 从cc同步主机所属业务，并将eip、disk、NI同步转移到该业务, 其他信息保持不变
func (mgr *CvmCCInfoRelManager) SyncCvmCCInfo(kt *kit.Kit, cvms []cvm.BaseCvm) error {
	hostIDs := make([]int64, 0)
	for _, cvm := range cvms {
		hostIDs = append(hostIDs, cvm.BkHostID)
	}

	hostBizIDMap := make(map[int64]int64)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		req := &cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := cmdb.CmdbClient().FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("fail to find cmdb topo relation, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return err
		}

		for _, relation := range converter.PtrToVal(relationRes) {
			hostBizIDMap[relation.HostID] = relation.BizID
		}
	}

	if err := mgr.updateCvmAndRelResBiz(kt, cvms, hostBizIDMap); err != nil {
		logs.Errorf("update cvm and rel res biz failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (mgr *CvmCCInfoRelManager) updateCvmAndRelResBiz(kt *kit.Kit, cvms []cvm.BaseCvm,
	hostBizIDMap map[int64]int64) error {
	updates := make([]cloud.CvmCommonInfoBatchUpdateData, 0)
	cvmIDBizMap := make(map[string]int64)
	for _, cvm := range cvms {
		bizID, ok := hostBizIDMap[cvm.BkHostID]
		if !ok {
			logs.Errorf("host not found biz in cmdb, hostID: %d, rid: %s", cvm.BkHostID, kt.Rid)
			continue
		}

		if cvm.BkBizID == bizID {
			continue
		}

		cvmIDBizMap[cvm.ID] = bizID
		update := cloud.CvmCommonInfoBatchUpdateData{ID: cvm.ID, BkBizID: converter.ValToPtr(bizID)}
		updates = append(updates, update)
	}

	for _, batch := range slice.Split(updates, constant.BatchOperationMaxLimit) {
		update := &cloud.CvmCommonInfoBatchUpdateReq{Cvms: batch}
		if err := mgr.dbCli.Global.Cvm.BatchUpdateCvmCommonInfo(kt, update); err != nil {
			logs.Errorf("update host common info failed, err: %v, req: %+v, rid: %s", err, update, kt.Rid)
			return err
		}
	}

	if err := mgr.updateCvmRelResBiz(kt, cvmIDBizMap); err != nil {
		logs.Errorf("update cvm rel res biz failed, err: %v, cvmIDBizMap: %v, rid: %s", err, cvmIDBizMap, kt.Rid)
		return err
	}

	return nil
}

func (mgr *CvmCCInfoRelManager) updateCvmRelResBiz(kt *kit.Kit, cvmIDBizMap map[string]int64) error {
	bizIDEipIDsMap, bizIDDiskIDsMap, bizIDNiIDsMap, err := mgr.getCvmRelResBiz(kt, cvmIDBizMap)
	if err != nil {
		logs.Errorf("build cvm rel res biz failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err = mgr.updateEipBizID(kt, bizIDEipIDsMap); err != nil {
		logs.Errorf("update eip biz failed, err: %v, bizIDEipIDsMap: %v, rid: %s", err, bizIDEipIDsMap, kt.Rid)
		return err
	}

	if err = mgr.updateDiskBizID(kt, bizIDDiskIDsMap); err != nil {
		logs.Errorf("update disk biz failed, err: %v, bizIDDiskIDsMap: %v, rid: %s", err, bizIDDiskIDsMap, kt.Rid)
		return err
	}

	if err = mgr.updateNiBizID(kt, bizIDNiIDsMap); err != nil {
		logs.Errorf("update ni biz failed, err: %v, bizIDNiIDsMap: %v, rid: %s", err, bizIDNiIDsMap, kt.Rid)
		return err
	}

	return nil
}

func (mgr *CvmCCInfoRelManager) getCvmRelResBiz(kt *kit.Kit,
	cvmIDBizMap map[string]int64) (bizIDEipIDsMap map[int64][]string,
	bizIDDiskIDsMap map[int64][]string, bizIDNiIDsMap map[int64][]string, err error) {

	cvmIDs := maps.Keys(cvmIDBizMap)

	bizIDEipIDsMap = make(map[int64][]string)
	for _, batch := range slice.Split(cvmIDs, constant.BatchOperationMaxLimit) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		eipResp, err := mgr.dbCli.Global.ListEipCvmRel(kt, req)
		if err != nil {
			logs.Errorf("list eip cvm rel failed, err: %v, cvmIDs: %v, rid: %s", err, batch, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range eipResp.Details {
			bizID := cvmIDBizMap[detail.CvmID]
			bizIDEipIDsMap[bizID] = append(bizIDEipIDsMap[bizID], detail.EipID)
		}
	}

	bizIDDiskIDsMap = make(map[int64][]string)
	for _, batch := range slice.Split(cvmIDs, constant.BatchOperationMaxLimit) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		diskResp, err := mgr.dbCli.Global.ListDiskCvmRel(kt, req)
		if err != nil {
			logs.Errorf("list disk cvm rel failed, err: %v, cvmIDs: %v, rid: %s", err, batch, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range diskResp.Details {
			bizID := cvmIDBizMap[detail.CvmID]
			bizIDDiskIDsMap[bizID] = append(bizIDDiskIDsMap[bizID], detail.DiskID)
		}
	}

	bizIDNiIDsMap = make(map[int64][]string)
	for _, batch := range slice.Split(cvmIDs, constant.BatchOperationMaxLimit) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		niResp, err := mgr.dbCli.Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(kt, req)
		if err != nil {
			logs.Errorf("list network interface cvm rel failed, err: %v, cvmIDs: %v, rid: %s", err, batch, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range niResp.Details {
			bizID := cvmIDBizMap[detail.CvmID]
			bizIDNiIDsMap[bizID] = append(bizIDNiIDsMap[bizID], detail.NetworkInterfaceID)
		}
	}

	return
}

func (mgr *CvmCCInfoRelManager) updateEipBizID(kt *kit.Kit, bizIDEipIDsMap map[int64][]string) error {
	for bizID, eipIDs := range bizIDEipIDsMap {
		for _, batch := range slice.Split(eipIDs, constant.BatchOperationMaxLimit) {
			req := &eip.EipBatchUpdateReq{
				IDs:     batch,
				BkBizID: uint64(bizID),
			}
			if _, err := mgr.dbCli.Global.BatchUpdateEip(kt, req); err != nil {
				logs.Errorf("batch update eip biz id failed, err: %v, req: %+v, rid: %s", err, converter.ValToPtr(req),
					kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (mgr *CvmCCInfoRelManager) updateDiskBizID(kt *kit.Kit, bizIDDiskIDsMap map[int64][]string) error {
	for bizID, diskIDs := range bizIDDiskIDsMap {
		for _, batch := range slice.Split(diskIDs, constant.BatchOperationMaxLimit) {
			req := &disk.DiskBatchUpdateReq{
				IDs:     batch,
				BkBizID: uint64(bizID),
			}
			if _, err := mgr.dbCli.Global.BatchUpdateDisk(kt, req); err != nil {
				logs.Errorf("batch update disk biz id failed, err: %v, req: %+v, rid: %s", err, converter.ValToPtr(req),
					kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (mgr *CvmCCInfoRelManager) updateNiBizID(kt *kit.Kit, bizIDNiIDsMap map[int64][]string) error {
	for bizID, niIDs := range bizIDNiIDsMap {
		for _, batch := range slice.Split(niIDs, constant.BatchOperationMaxLimit) {
			req := &networkinterface.NetworkInterfaceCommonInfoBatchUpdateReq{
				IDs:     batch,
				BkBizID: bizID,
			}
			if err := mgr.dbCli.Global.NetworkInterface.BatchUpdateNICommonInfo(kt, req); err != nil {
				logs.Errorf("batch update network interface biz id failed, err: %v, req: %+v, rid: %s", err,
					converter.ValToPtr(req), kt.Rid)
				return err
			}
		}
	}

	return nil
}
