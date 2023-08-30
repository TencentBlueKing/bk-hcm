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
	networkinterface "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/cloud/eip"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
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
				req := &hcprotocvm.TCloudBatchDeleteReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := c.client.HCService().TCloud.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchDeleteReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := c.client.HCService().Aws.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchDeleteReq{
					AccountID:      accountID,
					Region:         region,
					IDs:            ids,
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

// DeleteRecycledCvm batch delete recycled cvm.
func (c *cvm) DeleteRecycledCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

	if len(basicInfoMap) == 0 {
		return nil, nil
	}

	if len(basicInfoMap) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "cvm length should <= %d", constant.BatchOperationMaxLimit)
	}

	ids := make([]string, 0, len(basicInfoMap))

	bizHostsIds := make(map[int64][]string)
	for id, hostInfo := range basicInfoMap {
		if hostInfo.BkBizID > 0 {
			// 业务下机器加入bizHostsIds中
			bizHostsIds[hostInfo.BkBizID] = append(bizHostsIds[hostInfo.BkBizID], id)
		}
		ids = append(ids, id)
	}
	notRecyclableIds, err := c.GetNotRecyclableHosts(kt, bizHostsIds)
	if err != nil {
		return nil, err
	}
	// 存在非待回收模块主机直接报错
	if len(notRecyclableIds) > 0 {
		return nil, fmt.Errorf("host(%v) not belongs to recycle module in cc", notRecyclableIds)
	}

	// disassociate eip
	eipCvmMap, eipMap, err := c.getEipByCvm(kt, ids)
	if err != nil {
		return nil, err
	}

	for id, cvmID := range eipCvmMap {
		eip, exists := eipMap[id]
		if !exists {
			return nil, errf.Newf(errf.InvalidParameter, "eip %s not exists", id)
		}
		vendor := enumor.Vendor(eip.Vendor)

		// TODO get nic id by eip InstanceType
		var nicID string
		switch vendor {
		case enumor.Azure, enumor.Gcp, enumor.HuaWei:
			nicID = converter.PtrToVal(eip.InstanceID)
		}

		err = c.eip.DisassociateEip(kt, vendor, id, cvmID, nicID, eip.AccountID)
		if err != nil {
			logs.Errorf("disassociate eip %s failed, err: %v, cvm: %s, nic: %s, rid: %s", id, err, cvmID, nicID, kt.Rid)
			return nil, err
		}
	}

	// delete cvm
	delRes, err := c.BatchDeleteCvm(kt, basicInfoMap)
	if err != nil {
		// associate eip again if cvm deletion failed.
		for id, cvmID := range eipCvmMap {
			eip := eipMap[id]
			vendor := enumor.Vendor(eip.Vendor)

			// TODO get nic id by eip InstanceType
			var nicID string
			switch vendor {
			case enumor.Azure, enumor.Gcp, enumor.HuaWei:
				nicID = converter.PtrToVal(eip.InstanceID)
			}

			err = c.eip.AssociateEip(kt, vendor, id, cvmID, nicID, eip.AccountID)
			if err != nil {
				logs.Errorf("asst eip %s failed, err: %v, cvm: %s, nic: %s, rid: %s", id, err, cvmID, nicID, kt.Rid)
			}
		}

		return delRes, err
	}

	return nil, nil
}

// GetNotRecyclableHosts 获取根据业务id分类的主机id列表中不在cc待回收模块的主机id
func (c *cvm) GetNotRecyclableHosts(kt *kit.Kit, bizHostsIds map[int64][]string) ([]string, error) {
	// cloud id -> host id
	cloudToHostMap := make(map[string]string)
	notRecyclableIds := make([]string, 0)

	for bizID, hostIDs := range bizHostsIds {
		// 获取cloud id
		req := &cloud.CvmListReq{
			Field:  []string{"cloud_id", "vendor", "bk_biz_id", "id"},
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

func (c *cvm) getEipByCvm(kt *kit.Kit, ids []string) (map[string]string, map[string]*eip.EipResult, error) {
	// list eip and cvm relation
	relReq := &cloud.EipCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	relRes, err := c.client.DataService().Global.ListEipCvmRel(kt.Ctx, kt.Header(), relReq)
	if err != nil {
		return nil, nil, err
	}

	if len(relRes.Details) == 0 {
		return make(map[string]string), make(map[string]*eip.EipResult), nil
	}

	eipCvmMap := make(map[string]string)
	eipIDs := make([]string, 0, len(relRes.Details))
	for _, detail := range relRes.Details {
		eipCvmMap[detail.EipID] = detail.CvmID
		eipIDs = append(eipIDs, detail.EipID)
	}

	// list eip
	eipReq := &eip.EipListReq{
		Filter: tools.ContainersExpression("id", eipIDs),
		Page:   core.NewDefaultBasePage(),
	}
	eipRes, err := c.client.DataService().Global.ListEip(kt.Ctx, kt.Header(), eipReq)
	if err != nil {
		return nil, nil, err
	}

	// list network interface attached with eip
	eipMap := make(map[string]*eip.EipResult)
	publicIPs := make([]string, 0, len(eipRes.Details))
	for _, detail := range eipRes.Details {
		eipMap[detail.ID] = detail
		publicIPs = append(publicIPs, detail.PublicIp)
	}

	nicMap, err := c.listNicByCvmAndEip(kt, ids, publicIPs)
	if err != nil {
		return nil, nil, err
	}

	for eipID, cvmID := range eipCvmMap {
		ip := eipMap[eipID].PublicIp
		for _, nic := range nicMap[cvmID] {
			if slice.IsItemInSlice(nic.PublicIPv4, ip) || slice.IsItemInSlice(nic.PublicIPv6, ip) {
				eipMap[eipID].InstanceID = converter.ValToPtr(nic.ID)
			}
		}
	}

	return eipCvmMap, eipMap, nil
}

// TODO save eip and nic relation in db
func (c *cvm) listNicByCvmAndEip(kt *kit.Kit, ids []string, publicIPs []string) (
	map[string][]networkinterface.BaseNetworkInterface, error) {

	nicRelRes, err := c.client.DataService().Global.NetworkInterfaceCvmRel.List(kt.Ctx, kt.Header(),
		&core.ListReq{Filter: tools.ContainersExpression("cvm_id", ids), Page: core.NewDefaultBasePage()})
	if err != nil {
		return nil, err
	}

	if len(nicRelRes.Details) == 0 {
		return make(map[string][]networkinterface.BaseNetworkInterface), nil
	}

	nicIDs := make([]string, len(nicRelRes.Details))
	nicCvmMap := make(map[string]string)
	for idx, rel := range nicRelRes.Details {
		nicIDs[idx] = rel.NetworkInterfaceID
		nicCvmMap[rel.NetworkInterfaceID] = rel.CvmID
	}

	nicRes, err := c.client.DataService().Global.NetworkInterface.List(kt.Ctx, kt.Header(),
		&core.ListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{tools.ContainersExpression("id", nicIDs),
				&filter.Expression{Op: filter.Or, Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "public_ipv4", Op: filter.JSONOverlaps.Factory(), Value: publicIPs},
					filter.AtomRule{Field: "public_ipv6", Op: filter.JSONOverlaps.Factory(), Value: publicIPs},
				}},
			},
		}, Page: core.NewDefaultBasePage()})
	if err != nil {
		return nil, err
	}

	nicMap := make(map[string][]networkinterface.BaseNetworkInterface)
	for _, nic := range nicRes.Details {
		nicMap[nicCvmMap[nic.ID]] = append(nicMap[nicCvmMap[nic.ID]], nic)
	}
	return nicMap, nil
}

// CheckBizHostInRecycleModule 查询业务下主机是否在CC待回收模块中，CC只有业务下主机，完全没有业务下主机会报错
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
		BizID:  bizID,
		Fields: []string{"bk_host_id", "bk_cloud_inst_id"},
		Page:   cmdb.BasePage{Limit: 500},
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: "OR",
				Rules:     rules,
			},
		},
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
