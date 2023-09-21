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
	"context"
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
)

// BatchDeleteCvm batch delete cvm.
func (c *cvm) BatchDeleteCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

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
			if err := c.client.HCService().Gcp.Cvm.DeleteCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			req := &hcprotocvm.AzureDeleteReq{
				Force: true,
			}
			if err := c.client.HCService().Azure.Cvm.DeleteCvm(kt.Ctx, kt.Header(), one.ID, req); err != nil {
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
				if err := c.client.HCService().TCloud.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchDeleteReq{AccountID: accountID, Region: region, IDs: ids}
				if err := c.client.HCService().Aws.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchDeleteReq{AccountID: accountID, Region: region, IDs: ids,
					DeletePublicIP: true,
					DeleteDisk:     true,
				}
				if err := c.client.HCService().HuaWei.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
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
//
// TODO: 检查回收记录创建的时候的eip、disk、network interface 快照，要求完全一致，否则报错
func (c *cvm) DestroyRecycledCvm(kt *kit.Kit, cvmBasicInfo map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

	if len(cvmBasicInfo) == 0 {
		return nil, nil
	}
	if len(cvmBasicInfo) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "cvm length should <= %d", constant.BatchOperationMaxLimit)
	}
	leftCvmInfo := maps.Clone(cvmBasicInfo)
	destroyResult := new(core.BatchOperateResult)
	failedIds := make([]string, 0)
	markDestroyFailed := func(id string, reason error) {
		delete(leftCvmInfo, id)
		destroyResult.Failed = &core.FailedInfo{ID: id, Error: reason}
	}

	// 1. 预检，有一个失败则直接失败
	checkResult := c.RecyclePreCheck(kt, cvmBasicInfo)
	if len(checkResult.Failed) > 0 {
		return nil, checkResult.Failed[0].Error
	}

	// 到销毁机器的时候应该处理剩下的全部关联设备
	// 2. 解绑数据盘
	diskResult, diskRollBack, err := c.disk.BatchDetachWithRollback(kt, leftCvmInfo)
	if err != nil {
		for cvmId, err := range diskResult.FailedCvm {
			markDestroyFailed(cvmId, err)
		}
	}
	// 3. 解绑eip
	eipFailedCvmIds := make([]string, 0, len(leftCvmInfo))
	eipResult, eipRollback, err := c.eip.BatchDisassociateWithRollback(kt, converter.MapKeyToStringSlice(leftCvmInfo))
	if err != nil {
		for cvmId, err := range eipResult.FailedCvm {
			markDestroyFailed(cvmId, err)
			eipFailedCvmIds = append(eipFailedCvmIds, cvmId)
		}
	}
	defer func() {
		eipRollback(kt, failedIds)
		diskRollBack(kt, append(eipFailedCvmIds, failedIds...))
	}()
	if len(leftCvmInfo) == 0 {
		return destroyResult, destroyResult.Failed.Error
	}

	// 4. 销毁主机
	delRes, err := c.BatchDeleteCvm(kt, leftCvmInfo)
	if err != nil {
		logs.Errorf("Fail to delete cvm, err: %v, cvmIds: %v, rid: %s", err, delRes.Failed, kt.Rid)
		for _, cvmId := range delRes.Succeeded {
			delete(leftCvmInfo, cvmId)
		}
		// 回滚剩下的
		failedIds = append(failedIds, converter.MapKeyToStringSlice(leftCvmInfo)...)
	}
	// 5. 销毁数据盘
	for diskId, cvmId := range diskResult.SucceedResCvm {
		err := c.disk.DeleteDisk(kt, cvmBasicInfo[cvmId].Vendor, diskId)
		if err != nil {
			logs.Errorf("fail to delete %s disk, err: %v, diskId: %s, rid: %s", cvmBasicInfo[cvmId].Vendor, err, kt.Rid)
		}
	}
	// 6. 销毁eip
	for eipId, cvmId := range eipResult.SucceedResCvm {
		err := c.eip.DeleteEip(kt, cvmBasicInfo[cvmId].Vendor, eipId)
		if err != nil {
			logs.Errorf("fail to delete %s disk, err: %v, diskId: %s, rid: %s", cvmBasicInfo[cvmId].Vendor, err, kt.Rid)
		}
	}
	return destroyResult, nil
}

// GetNotCmdbRecyclableHosts 获取根据业务id分类的主机id列表中不在cmdb待回收模块的主机id
func (c *cvm) GetNotCmdbRecyclableHosts(kt *kit.Kit, bizHostsIds map[int64][]string) ([]string, error) {
	// cloud id -> host id
	cloudToHostMap := make(map[string]string)
	notRecyclableIds := make([]string, 0)

	for bizID, hostIDs := range bizHostsIds {
		// 获取cloud id
		req := &cloud.CvmListReq{
			Field:  []string{"cloud_id", "vendor", "bk_biz_id", "id", "status"},
			Filter: tools.ContainersExpression("id", hostIDs),
			Page:   core.NewDefaultBasePage(),
		}
		relResp, err := c.client.DataService().Global.Cvm.ListCvm(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf(" fail to get host Info:%v", err)
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
		inRecycle, err := c.CheckBizHostInRecycleModule(kt.Ctx, bizID, cloudIds)
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
func (c *cvm) CheckBizHostInRecycleModule(ctx context.Context, bizID int64,
	cloudIDs map[enumor.Vendor][]string) (map[string]bool, error) {

	// 1. 获取cmdb主机id
	cloudToHost, err := c.getCmdbHostId(ctx, bizID, cloudIDs)
	if err != nil {
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
	relation, err := c.esbClient.Cmdb().FindHostTopoRelation(ctx,
		&cmdb.FindHostTopoRelationParams{
			HostIDs: hostIDs, BizID: bizID,
			Page: cmdb.BasePage{Limit: 200, Start: 0},
		},
	)
	if err != nil {
		return nil, err
	}

	modRecyclable := make(map[int64]bool, len(relation.Data))
	hostRecyclable := make(map[string]bool, len(cloudIDs))

	// 3. 逐个查询主机模块信息
	for _, rel := range relation.Data {
		if _, ok := modRecyclable[rel.BkModuleID]; !ok {
			module, err := c.esbClient.Cmdb().SearchModule(ctx, &cmdb.SearchModuleParams{
				BizID:  bizID,
				Fields: []string{"default", "bk_module_id"},
				Condition: map[string]interface{}{
					"bk_module_id": rel.BkModuleID,
				},
			})
			if err != nil {
				return nil, err
			}
			if len(module.Info) != 1 {
				logs.Errorf("module info count mismatch:%v", module)
				return nil, errors.New("module info count mismatch")
			}
			// default 值为3 的是可回收模块
			modRecyclable[rel.BkModuleID] = module.Info[0].Default == 3
		}

		hostRecyclable[hostToCloud[rel.HostID]] = modRecyclable[rel.BkModuleID]
	}
	return hostRecyclable, nil
}

func (c *cvm) getCmdbHostId(ctx context.Context, bizID int64,
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
	hosts, err := c.esbClient.Cmdb().ListBizHost(ctx, listParams)
	if err != nil {
		return nil, err
	}

	if len(hosts.Info) == 0 {
		logs.Errorf("no host in business(%d):%v", bizID, cloudIDs)
		return nil, nil
	}

	hostIDs := make(map[string]int64, len(hosts.Info))
	for _, host := range hosts.Info {
		hostIDs[host.BkCloudInstID] = host.BkHostID
	}
	return hostIDs, nil
}

// RecyclePreCheck  回收预校验、包含主机状态和CC待回收模块检查，see Interface.RecyclePreCheck
func (c *cvm) RecyclePreCheck(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (

	result *core.BatchOperateAllResult) {

	leftInfo := maps.Clone(basicInfoMap)
	result = new(core.BatchOperateAllResult)
	markFail := func(err error, ids ...string) {
		for _, id := range ids {
			delete(leftInfo, id)
			result.Failed = append(result.Failed, core.FailedInfo{Error: err, ID: id})
		}
	}

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
			// 	获取信息失败，全部标记失败
			for _, ids := range bizHostsMap {
				markFail(err, ids...)
			}
			return result
		}
		e := fmt.Errorf("host not belongs to recycle module in cmdb, host id: %v", notRecyclableIds)
		markFail(e, notRecyclableIds...)
	}

	// 2. CVM尝试关机检查
	checkResult := c.checkAndStopCvm(kt, leftInfo)
	result.Failed = append(result.Failed, checkResult.Failed...)
	result.Succeeded = converter.MapKeyToStringSlice(leftInfo)

	return result
}
