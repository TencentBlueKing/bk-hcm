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
	"errors"
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	corerecord "hcm/pkg/api/core/recycle-record"
	dsrecord "hcm/pkg/api/data-service/recycle-record"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// BatchDeleteCvm batch delete cvm.
func (c *cvm) BatchDeleteCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

	if len(basicInfoMap) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(basicInfoMap))
	for id := range basicInfoMap {
		ids = append(ids, id)
	}

	if err := c.audit.ResDeleteAudit(kt, enumor.CvmAuditResType, ids); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cvmVendorMap := classifier.ClassifyBasicInfoByVendor(basicInfoMap)
	successIDs := make([]string, 0)
	for vendor, infos := range cvmVendorMap {
		switch vendor {
		case enumor.TCloud, enumor.Aws, enumor.HuaWei:
			ids, err := c.batchDeleteCvm(kt, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return &core.BatchOperateResult{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						Error: err,
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		case enumor.Azure, enumor.Gcp:
			ids, failedID, err := c.deleteCvm(kt, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return &core.BatchOperateResult{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						ID:    failedID,
						Error: err,
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		default:
			return &core.BatchOperateResult{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    infos[0].ID,
					Error: errf.Newf(errf.Unknown, "vendor: %s not support", vendor),
				},
			}, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}

	}

	return nil, nil
}

func (c *cvm) deleteCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, string, error) {

	successIDs := make([]string, 0)
	for _, one := range infoMap {
		switch vendor {
		case enumor.Gcp:
			if err := c.client.HCService().Gcp.Cvm.DeleteCvm(kt, one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			req := &hcprotocvm.AzureDeleteReq{
				Force: true,
			}
			if err := c.client.HCService().Azure.Cvm.DeleteCvm(kt, one.ID, req); err != nil {
				return successIDs, one.ID, err
			}

		default:
			return successIDs, one.ID, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}
	}

	return successIDs, "", nil
}

// batchDeleteCvm delete cvm.
func (c *cvm) batchDeleteCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, error) {

	cvmMap := classifier.ClassifyBasicInfoByAccount(infoMap)
	successIDs := make([]string, 0)
	for accountID, reginMap := range cvmMap {
		for region, ids := range reginMap {
			switch vendor {
			case enumor.TCloud:
				req := &hcprotocvm.TCloudBatchDeleteReq{AccountID: accountID, Region: region, IDs: ids}
				if err := c.client.HCService().TCloud.Cvm.BatchDeleteCvm(kt, req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchDeleteReq{AccountID: accountID, Region: region, IDs: ids}
				if err := c.client.HCService().Aws.Cvm.BatchDeleteCvm(kt, req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchDeleteReq{AccountID: accountID, Region: region, IDs: ids,
					DeletePublicIP: true,
					DeleteDisk:     true,
				}
				if err := c.client.HCService().HuaWei.Cvm.BatchDeleteCvm(kt, req); err != nil {
					return successIDs, err
				}

			default:
				return successIDs, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
			}
			successIDs = append(successIDs, ids...)
		}
	}

	return successIDs, nil
}

// DestroyRecycledCvm 销毁已经处于回收状态的Cvm，并连带当前主机上绑定的eip、disk、nic一同销毁。
// 该动作由：
//  1. 用户手动发起
//  2. 由定时回收任务触发
func (c *cvm) DestroyRecycledCvm(kt *kit.Kit, cvmBasicInfo map[string]types.CloudResourceBasicInfo,
	records []corerecord.CvmRecycleRecord) (*core.BatchOperateResult, error) {

	if len(cvmBasicInfo) == 0 {
		return nil, nil
	}
	if len(cvmBasicInfo) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "cvm length should <= %d", constant.BatchOperationMaxLimit)
	}
	// 获取回收时的挂载信息
	cvmRecycleDetails := make(map[string]corerecord.CvmRecycleDetail, len(records))
	for _, record := range records {
		cvmRecycleDetails[record.ResID] = record.Detail
	}

	leftCvmInfo := maps.Clone(cvmBasicInfo)
	destroyResult := new(core.BatchOperateResult)

	// 1. 检查cmdb模块和机器状态
	if err := c.RecyclePreCheck(kt, cvmBasicInfo); err != nil {
		logs.Errorf("destroy precheck fail, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 到销毁机器的时候应该处理剩下的全部关联资源
	cvmStatus, err := c.checkAndUnbindCvmRelated(kt, cvmBasicInfo, cvmRecycleDetails)
	if err != nil {
		logs.Errorf("fail to unbind related res of cvm before destroy, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 全部失败
	if len(cvmStatus) == 0 {
		return destroyResult, destroyResult.Failed.Error
	}
	defer func(c *cvm, kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) {
		err := c.destroyCleanUp(kt, cvmStatus)
		if err != nil {
			logs.Errorf("failed to cleanup, err: %v, rid: %s", err, kt.Rid)
		}
	}(c, kt, cvmStatus)
	// 4. 销毁主机
	delRes, err := c.BatchDeleteCvm(kt, maps.FilterByValue(leftCvmInfo, func(info types.CloudResourceBasicInfo) bool {
		return cvmStatus[info.ID] != nil && cvmStatus[info.ID].FailedAt == ""
	}))
	if err != nil {
		logs.Errorf("Fail to delete cvm, err: %v, cvmIds: %v, rid: %s", err, delRes, kt.Rid)
		for _, cvmId := range delRes.Succeeded {
			delete(cvmStatus, cvmId)
		}
	}
	// 销毁关联资源
	c.destroyRelatedRes(kt, cvmStatus)
	return destroyResult, nil
}

func (c *cvm) destroyRelatedRes(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) {
	for _, recycleDetail := range cvmStatus {

		for _, disk := range recycleDetail.DiskList {
			err := c.disk.DeleteDisk(kt, recycleDetail.Vendor, disk.DiskID)
			if err != nil {
				logs.Errorf("fail to delete %s disk, err: %v, diskId: %s, rid: %s",
					recycleDetail.Vendor, err, disk.DiskID, kt.Rid)
			}

			// 处理关联磁盘回收任务
			// 根据磁盘id 和回收任务类型获取对应的回收任务id
			queryReq := &core.ListReq{Page: core.NewDefaultBasePage(),
				Filter: tools.EqualWithOpExpression(filter.And, map[string]interface{}{
					"recycle_type": enumor.RecycleTypeRelated,
					"status":       enumor.WaitingRecycleRecordStatus,
					"res_id":       disk.DiskID,
				})}
			resp, err := c.client.DataService().Global.RecycleRecord.ListRecycleRecord(kt, queryReq)
			if err != nil {
				logs.Errorf("fail to query related disk recycle record, err: %v, req: %v, rid: %s",
					err, queryReq, kt.Rid)
				return
			}
			if len(resp.Details) != 1 {
				logs.Errorf("query related disk recycle record length mismatch, want:1 ,got %v, req: %v, rid: %s",
					len(resp.Details), queryReq, kt.Rid)
				return
			}
			recordID := resp.Details[0].ID
			updateReq := &dsrecord.BatchUpdateReq{
				Data: []dsrecord.UpdateReq{{ID: recordID, Status: enumor.RecycledRecycleRecordStatus}},
			}
			err = c.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(kt, updateReq)
			if err != nil {
				logs.Errorf("fail to update related disk recycle record status, err: %v, recordID: %s, rid: %s",
					err, recordID, kt.Rid)
			}

		}
		for _, eip := range recycleDetail.EipList {
			err := c.eip.DeleteEip(kt, recycleDetail.Vendor, eip.EipID)
			if err != nil {
				logs.Errorf("fail to delete %s eip, err: %v, eipID: %s, rid: %s",
					recycleDetail.Vendor, err, eip.EipID, kt.Rid)
			}
		}
	}
}

// 获取磁盘绑定信息，和回收时的对比，如果一致则解绑
func (c *cvm) checkAndUnbindCvmRelated(kt *kit.Kit, cvmBasicInfo map[string]types.CloudResourceBasicInfo,
	originDetails map[string]corerecord.CvmRecycleDetail) (cvmStatus map[string]*recycle.CvmDetail, err error) {

	cvmStatus = make(map[string]*recycle.CvmDetail, len(cvmBasicInfo))
	for cvmId, basicInfo := range cvmBasicInfo {
		cvmStatus[cvmId] = &recycle.CvmDetail{
			Vendor:    basicInfo.Vendor,
			AccountID: basicInfo.AccountID,
			CvmID:     cvmId,
		}
	}
	// 获取信息
	if err := c.disk.BatchGetDiskInfo(kt, cvmStatus); err != nil {
		logs.Errorf("failed to get disk info of cvm, err: %v, rid: %s", err, kt.Rid)
		return cvmStatus, err
	}
	if err := c.eip.BatchGetEipInfo(kt, cvmStatus); err != nil {
		logs.Errorf("failed to get eip info of cvm, err: %v, rid: %s", err, kt.Rid)
	}
	// 对比前后是否发生变化
	for cvmId, now := range cvmStatus {
		origin := originDetails[cvmId]
		if origin.WithEip {
			newData, changed, deleted := common.Diff(now.EipList, origin.EipList,
				func(now corerecord.EipBindInfo, origin corerecord.EipBindInfo) bool {
					return origin.NicID != now.NicID
				})
			if len(newData) > 0 || len(changed) > 0 || len(deleted) > 0 {
				return nil, fmt.Errorf("eip bind status changed, added: %+v, modified: %+v, deleted: %+v",
					newData, changed, deleted)
			}
		}
		if origin.WithDisk {
			newData, changed, deleted := common.Diff(now.DiskList, origin.DiskList,
				func(now corerecord.DiskAttachInfo, origin corerecord.DiskAttachInfo) bool {
					return origin.DeviceName != now.DeviceName || origin.CachingType != now.CachingType
				})
			if len(newData) > 0 || len(changed) > 0 || len(deleted) > 0 {
				return nil, fmt.Errorf("disk mount status changed, added: %+v, modified: %+v, deleted: %+v",
					newData, changed, deleted)
			}
		}
	}

	// 解绑全部磁盘
	failed, err := c.disk.BatchDetach(kt, cvmStatus)
	if err != nil {
		logs.Errorf("failed to detach some disks of cvm(%v), err: %v, rid: %s", failed, err, kt.Rid)
		return cvmStatus, err
	}

	// 解绑全部eip
	failed, err = c.eip.BatchUnbind(kt, cvmStatus)
	if err != nil {
		logs.Errorf("failed to unbind eip of cvm(%v), err: %v, rid: %s", failed, err, kt.Rid)
	}
	return cvmStatus, err

}

// destroyCleanUp 处理回收失败需要尝试重新绑定的eip、disk
func (c *cvm) destroyCleanUp(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) error {

	eipRebind := make(map[string]*recycle.CvmDetail, len(cvmStatus))
	diskRebind := make(map[string]*recycle.CvmDetail, len(cvmStatus))

	for cvmId, detail := range cvmStatus {
		switch detail.FailedAt {
		case "":
			continue
		case enumor.DiskCloudResType:
			continue
		case enumor.EipCloudResType:
			diskRebind[cvmId] = detail
		case enumor.CvmCloudResType:
			// 	重新挂载磁盘和绑定eip
			eipRebind[cvmId] = detail
			diskRebind[cvmId] = detail
		default:
			return fmt.Errorf("unknown failed type: %v", detail.FailedAt)
		}
	}
	// 	尝试重新挂载磁盘
	err := c.eip.BatchRebind(kt, eipRebind)
	if err != nil {
		return err
	}
	err = c.disk.BatchReattachDisk(kt, diskRebind)
	if err != nil {
		return err
	}
	// 标记关联磁盘任务为失败
	var diskIds []string
	for _, cvmDetail := range diskRebind {
		if cvmDetail.WithDisk && len(cvmDetail.DiskList) > 0 {
			for _, disk := range cvmDetail.DiskList {
				diskIds = append(diskIds, disk.DiskID)
			}
		}
	}
	err = c.BatchFinalizeRelRecord(kt, enumor.DiskCloudResType, enumor.FailedRecycleRecordStatus, diskIds)
	if err != nil {
		logs.Errorf("fail to mark related disk recycle record failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetNotCmdbRecyclableHosts 获取根据业务id分类的主机id列表中不在cmdb待回收模块的主机id
func (c *cvm) GetNotCmdbRecyclableHosts(kt *kit.Kit, bizHostsIds map[int64][]string) ([]string, error) {
	// cloud id -> host id
	cloudToHostMap := make(map[string]string)
	notRecyclableIds := make([]string, 0)

	for bizID, hostIDs := range bizHostsIds {
		// 获取cloud id
		req := &core.ListReq{
			Fields: []string{"cloud_id", "vendor", "bk_biz_id", "id", "status"},
			Filter: tools.ContainersExpression("id", hostIDs),
			Page:   core.NewDefaultBasePage(),
		}
		relResp, err := c.client.DataService().Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.Errorf("fail to get host Info, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		// 按vendor 归类
		cloudIds := make(map[enumor.Vendor][]string)
		for _, detail := range relResp.Details {
			cloudToHostMap[detail.CloudID] = detail.ID

			vendors, ok := cloudIds[detail.Vendor]
			if !ok {
				vendors = make([]string, 0, len(relResp.Details))
			}
			vendors = append(vendors, detail.CloudID)
			cloudIds[detail.Vendor] = vendors
		}
		// 	去cmdb处检查，是否在待回收模块
		inRecycle, err := c.CheckBizHostInRecycleModule(kt, bizID, cloudIds)
		if err != nil {
			return nil, err
		}
		for cloudID, recyclable := range inRecycle {
			hostID := cloudToHostMap[cloudID]
			if !recyclable {
				notRecyclableIds = append(notRecyclableIds, hostID)
			}
		}
	}
	return notRecyclableIds, nil
}

// CheckBizHostInRecycleModule 查询业务下主机是否在cmdb待回收模块中，cmdb只有业务下主机，完全没有业务下主机会报错
func (c *cvm) CheckBizHostInRecycleModule(kt *kit.Kit, bizID int64,
	cloudIDs map[enumor.Vendor][]string) (map[string]bool, error) {

	// 1. 获取cmdb主机id
	cloudToHost, err := c.getCmdbHostId(kt, bizID, cloudIDs)
	if err != nil {
		logs.Errorf("fail to get cmdb host id, err: %v, bizID: %v, cloudIDs: %v, rid: %s",
			err, bizID, cloudIDs, kt.Rid)
		return nil, err
	}
	if cloudToHost == nil {
		return nil, errf.Newf(errf.InvalidParameter, "no host in business(%d)", bizID)
	}
	hostToCloud := make(map[int64]string)
	hostIDs := make([]int64, 0, len(cloudToHost))
	for cloudID, hostID := range cloudToHost {
		hostIDs = append(hostIDs, hostID)
		hostToCloud[hostID] = cloudID
	}
	//  2. 查找主机关系，获取模块信息
	relation, err := c.esbClient.Cmdb().FindHostTopoRelation(kt,
		&cmdb.FindHostTopoRelationParams{
			HostIDs: hostIDs, BizID: bizID,
			Page: cmdb.BasePage{Limit: 200, Start: 0},
		},
	)
	if err != nil {
		logs.Errorf("fail to find cmdb topo rel, err: %v, hostIDs: %v, bizID:%v, rid: %s", err, hostIDs, bizID, kt.Rid)
		return nil, err
	}

	modRecyclable := make(map[int64]bool, len(relation.Data))
	hostRecyclable := make(map[string]bool, len(cloudIDs))

	// 3. 逐个查询主机模块信息
	for _, rel := range relation.Data {
		if _, ok := modRecyclable[rel.BkModuleID]; !ok {
			module, err := c.esbClient.Cmdb().SearchModule(kt, &cmdb.SearchModuleParams{
				BizID:  bizID,
				Fields: []string{"default", "bk_module_id"},
				Condition: map[string]interface{}{
					"bk_module_id": rel.BkModuleID,
				},
			})
			if err != nil {
				logs.Errorf("fail to search module in cmdb, err: %v, bk_module_id: %s, rid: %s", err, rel.BkModuleID)
				return nil, err
			}
			if len(module.Info) != 1 {
				logs.Errorf("module info count mismatch, got: %v, length should be 1", module)
				return nil, errors.New("module info count mismatch")
			}
			// default 值为3 的是可回收模块
			modRecyclable[rel.BkModuleID] = module.Info[0].Default == 3
		}
		hostRecyclable[hostToCloud[rel.HostID]] = modRecyclable[rel.BkModuleID]
	}
	return hostRecyclable, nil
}

func (c *cvm) getCmdbHostId(kt *kit.Kit, bizID int64,
	cloudIDs map[enumor.Vendor][]string) (map[string]int64,
	error) {
	// get cmdb host ids
	rules := make([]cmdb.Rule, 0)
	for vendor, ids := range cloudIDs {
		rule := &cmdb.CombinedRule{
			Condition: "AND",
			Rules: []cmdb.Rule{
				&cmdb.AtomRule{
					Field:    "bk_cloud_vendor",
					Operator: cmdb.OperatorEqual,
					Value:    cmdb.HcmCmdbVendorMap[vendor],
				},
				&cmdb.AtomRule{
					Field:    "bk_cloud_inst_id",
					Operator: cmdb.OperatorIn,
					Value:    ids,
				},
			},
		}
		rules = append(rules, rule)
	}

	listParams := &cmdb.ListBizHostParams{
		BizID:              bizID,
		Fields:             []string{"bk_host_id", "bk_cloud_inst_id"},
		Page:               cmdb.BasePage{Limit: 500},
		HostPropertyFilter: &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "OR", Rules: rules}},
	}
	hosts, err := c.esbClient.Cmdb().ListBizHost(kt, listParams)
	if err != nil {
		logs.Errorf("fail to list cmdb biz host, err: %v, bizID:%v, rid: %s", err, bizID, kt.Rid)
		return nil, err
	}

	if len(hosts.Info) == 0 {
		logs.Infof("no host in business(%d), cloudIDs: %v", bizID, cloudIDs)
		return nil, nil
	}

	hostIDs := make(map[string]int64, len(hosts.Info))
	for _, host := range hosts.Info {
		hostIDs[host.BkCloudInstID] = host.BkHostID
	}
	return hostIDs, nil
}

// RecyclePreCheck  回收预校验，包含主机状态和CC待回收模块检查
func (c *cvm) RecyclePreCheck(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) error {

	leftInfo := maps.Clone(basicInfoMap)
	bizHostsMap := make(map[int64][]string)
	for id, hostInfo := range leftInfo {
		if hostInfo.BkBizID > 0 {
			// 业务下机器加入bizHostsIds中
			bizHostsMap[hostInfo.BkBizID] = append(bizHostsMap[hostInfo.BkBizID], id)
		}
	}

	// 1. cmdb 待回收检查有业务下的主机，检查是否在cmdb待回收模块
	if len(bizHostsMap) > 0 {
		notRecyclableIds, err := c.GetNotCmdbRecyclableHosts(kt, bizHostsMap)
		if err != nil {
			logs.Errorf("fail to check cvm in cmdb recyclable module, err: %v, bizHostMap: %v, rid: %s",
				err, bizHostsMap, kt.Rid)
			return err
		}
		if len(notRecyclableIds) > 0 {
			return fmt.Errorf("host not belongs to recycle module in cmdb, host id: %v", notRecyclableIds)
		}
	}

	// 2. CVM尝试关机检查
	err := c.checkAndStopCvm(kt, leftInfo)
	if err != nil {
		logs.Errorf("fail to check or stop cvm, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// BatchFinalizeRelRecord 批量更改关联资源的状态
func (c *cvm) BatchFinalizeRelRecord(kt *kit.Kit, resType enumor.CloudResourceType,
	status enumor.RecycleRecordStatus, resIds []string) error {
	// 查询关联的磁盘回收任务
	relListReq := &core.ListReq{Page: core.NewDefaultBasePage(), Filter: &filter.Expression{Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.EqualExpression("res_type", resType),
			tools.EqualExpression("status", enumor.WaitingRecycleRecordStatus),
			tools.EqualExpression("recycle_type", enumor.RecycleTypeRelated),
			tools.ContainersExpression("res_id", resIds),
		},
	}}
	relRecords, err := c.client.DataService().Global.RecycleRecord.ListRecycleRecord(kt, relListReq)
	if err != nil {
		logs.Errorf("fail to list related disk recycle record, err: %s, rid: %s", err, kt.Rid)
		return err
	}
	// 更新关联资源回收任务状态
	updateRecordOpt := dsrecord.BatchUpdateReq{Data: slice.Map(relRecords.Details,
		func(rel corerecord.RecycleRecord) dsrecord.UpdateReq {
			return dsrecord.UpdateReq{ID: rel.ID, Status: status}
		}),
	}
	err = c.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(kt, &updateRecordOpt)
	if err != nil {
		logs.Errorf("fail to update related %s recycle record status to '%s', err: %s, rid: %s",
			resType, status, err, kt.Rid)
		return err
	}
	return nil
}
