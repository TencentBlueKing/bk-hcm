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

// Package cvm ...
package cvm

import (
	"encoding/json"
	"fmt"
	"strings"

	typecore "hcm/pkg/adaptor/types/core"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

type validateTCloudOperateStatusFunc func(moduleName, cloudCvmStatus string) enumor.CvmOperateStatus

// chooseTCloudValidateFunc 根据operateType选择校验函数
func chooseTCloudValidateFunc(operateType enumor.CvmOperateType) (validateTCloudOperateStatusFunc, error) {
	switch operateType {
	case enumor.CvmOperateTypeReset:
		return validateTCloudOperateStatusForReset, nil
	case enumor.CvmOperateTypeStart, enumor.CvmOperateTypeStop, enumor.CvmOperateTypeReboot:
		return func(moduleName, cloudCvmStatus string) enumor.CvmOperateStatus {
			return validateTCloudOperateStatusForOperate(moduleName, cloudCvmStatus, operateType)
		}, nil
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter,
			fmt.Errorf("chooseTCloudValidateFunc invalid operate type: %s", operateType))
	}
}

// validateTCloudOperateStatusForReset 主机重装的校验
func validateTCloudOperateStatusForReset(moduleName, _ string) enumor.CvmOperateStatus {

	if moduleName != constant.IdleMachine && moduleName != constant.CCIdleMachine &&
		moduleName != constant.IdleMachineModuleName {
		return enumor.CvmOperateStatusNoIdle
	}

	return enumor.CvmOperateStatusNormal
}

// validateTCloudOperateStatusForOperate 主机开关机、重启的校验
func validateTCloudOperateStatusForOperate(_, cloudCvmStatus string,
	operationType enumor.CvmOperateType) enumor.CvmOperateStatus {

	switch operationType {
	case enumor.CvmOperateTypeStart:
		if cloudCvmStatus != enumor.TCloudCvmStatusStopped {
			return enumor.CvmOperateStatusNoStop
		}
	case enumor.CvmOperateTypeStop, enumor.CvmOperateTypeReboot:
		if cloudCvmStatus != enumor.TCloudCvmStatusRunning {
			return enumor.CvmOperateStatusNoRunning
		}
	default:
		logs.Errorf("validateTCloudOperateStatusForOperate invalid operate type: %s", operationType)
	}
	return enumor.CvmOperateStatusNormal
}

// listTCloudCvmOperateHost 获取主机列表&可操作状态
func (svc *cvmSvc) listTCloudCvmOperateHost(kt *kit.Kit, cvmIDs []string,
	validateStatusFunc validateTCloudOperateStatusFunc) ([]cscvm.CvmBatchOperateHostInfo, error) {

	// 根据主机ID获取主机列表
	hostIDs, hostCvmMap, err := svc.listTCloudCvmExtMapByIDs(kt, cvmIDs)
	if err != nil {
		return nil, err
	}

	// 查询cc的Topo关系
	mapHostToRel, mapModuleIdToModule, err := svc.listCmdbHostRelModule(kt, hostIDs)
	if err != nil {
		return nil, err
	}

	mapCloudIDToCvm, err := svc.mapTCloudCloudCvms(kt, hostCvmMap)
	if err != nil {
		logs.Errorf("fail to map tcloud cloud cvms, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cvmHosts := make([]cscvm.CvmBatchOperateHostInfo, 0)
	for _, host := range hostCvmMap {
		hostID := host.BkHostID
		moduleName := ""
		if rel, ok := mapHostToRel[hostID]; ok {
			if module, exist := mapModuleIdToModule[rel.BkModuleID]; exist {
				moduleName = module.BkModuleName
			}
		}

		// 只有虚拟机才需要判断云端状态
		cvmCloudInfo, ok := mapCloudIDToCvm[host.CloudID]
		if !ok {
			logs.Warnf("cloud cvm not found, cloud_id: %v, host_id: %v, rid: %s", host.CloudID, hostID, kt.Rid)
			return nil, fmt.Errorf("cloud cvm not found, cloud_id: %v, host_id: %v", host.CloudID, hostID)
		}

		cvmHosts = append(cvmHosts, cscvm.CvmBatchOperateHostInfo{
			ID:                   host.ID,
			Vendor:               host.Vendor,
			AccountID:            host.AccountID,
			Name:                 host.Name,
			BkHostID:             hostID,
			CloudID:              hostCvmMap[hostID].CloudID,
			PrivateIPv4Addresses: hostCvmMap[hostID].PrivateIPv4Addresses,
			PrivateIPv6Addresses: hostCvmMap[hostID].PrivateIPv6Addresses,
			PublicIPv4Addresses:  hostCvmMap[hostID].PublicIPv4Addresses,
			PublicIPv6Addresses:  hostCvmMap[hostID].PublicIPv6Addresses,
			Region:               hostCvmMap[hostID].Region,
			Zone:                 hostCvmMap[hostID].Zone,
			TopoModule:           moduleName,
			Status:               hostCvmMap[hostID].Status,
			OperateStatus:        validateStatusFunc(moduleName, cvmCloudInfo.Status),
		})
	}

	return cvmHosts, nil
}

// mapTCloudCloudCvms 查询云上的主机信息
func (svc *cvmSvc) mapTCloudCloudCvms(kt *kit.Kit, cvms map[int64]corecvm.Cvm[corecvm.TCloudCvmExtension]) (
	map[string]corecvm.Cvm[corecvm.TCloudCvmExtension], error) {

	// group by account, region
	mapAccountRegionToCvmCloudID := make(map[string][]string)
	for _, host := range cvms {
		key := getCombinedKey(host.AccountID, host.Region, "+")
		mapAccountRegionToCvmCloudID[key] = append(mapAccountRegionToCvmCloudID[key], host.CloudID)
	}

	result := make(map[string]corecvm.Cvm[corecvm.TCloudCvmExtension])
	for key, cloudIDs := range mapAccountRegionToCvmCloudID {
		split := strings.Split(key, "+")
		accountID, region := split[0], split[1]
		if region == "" {
			logs.Errorf("region is empty, account_id: %s, cvm_ids: %v, rid: %s", accountID, cloudIDs, kt.Rid)
			return nil, fmt.Errorf("region is empty, account_id: %s, cvm_ids: %v", accountID, cloudIDs)
		}

		for _, ids := range slice.Split(cloudIDs, typecore.TCloudQueryLimit) {
			req := &corecvm.QueryCloudCvmReq{
				Vendor:    enumor.TCloud,
				AccountID: accountID,
				Region:    region,
				CvmIDs:    ids,
				Page:      &core.BasePage{Start: 0, Limit: typecore.TCloudQueryLimit},
			}
			resp, err := svc.client.HCService().TCloud.Cvm.QueryTCloudCVM(kt, req)
			if err != nil {
				logs.Errorf("fail to query tcloud cvm, err: %v, cloud_ids: %v, rid:%s", err, cloudIDs, kt.Rid)
				return nil, err
			}
			for _, detail := range resp.Details {
				result[detail.CloudID] = detail
			}
		}
	}
	return result, nil
}

// 拼接唯一key main-sub
func getCombinedKey(main, sub, sep string) string {
	return main + sep + sub
}

// listCmdbHostRelModule 查询cc的主机列表及Topo关系
func (svc *cvmSvc) listCmdbHostRelModule(kt *kit.Kit, hostIDs []int64) (map[int64]cmdb.HostTopoRelation,
	map[int64]*cmdb.ModuleInfo, error) {

	// get host topo info
	relations, err := svc.cvmLgc.GetHostTopoInfo(kt, hostIDs)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}

	bizIds := make([]int64, 0)
	mapBizToModule := make(map[int64][]int64)
	mapHostToRel := make(map[int64]cmdb.HostTopoRelation)
	for _, rel := range relations {
		mapHostToRel[rel.HostID] = rel
		if _, ok := mapBizToModule[rel.BizID]; !ok {
			mapBizToModule[rel.BizID] = []int64{rel.BkModuleID}
			bizIds = append(bizIds, rel.BizID)
		} else {
			mapBizToModule[rel.BizID] = append(mapBizToModule[rel.BizID], rel.BkModuleID)
		}
	}

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := util.IntArrayUnique(moduleIds)
		moduleList, err := svc.cvmLgc.GetModuleInfo(kt, bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to cvm reset check, for get module info, err: %v, bizId: %d, "+
				"moduleIdUniq: %v, rid: %s", err, bizId, moduleIdUniq, kt.Rid)
			return nil, nil, err
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}
	// 记录日志
	hostRelJson, err := json.Marshal(mapHostToRel)
	if err != nil {
		logs.Errorf("failed to marshal mapHostToRel, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}
	moduleIdToModuleJson, err := json.Marshal(mapModuleIdToModule)
	if err != nil {
		logs.Errorf("failed to marshal mapModuleIdToModule, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}
	logs.Infof("list cmdb host rel module success, hostIDs: %v, mapHostToRel: %s, mapModuleIdToModule: %s, rid: %s",
		hostIDs, hostRelJson, moduleIdToModuleJson, kt.Rid)

	return mapHostToRel, mapModuleIdToModule, nil
}

// listTCloudCvmExtMapByIDs 根据主机ID获取主机列表（含扩展信息）
func (svc *cvmSvc) listTCloudCvmExtMapByIDs(kt *kit.Kit, cvmIDs []string) (
	[]int64, map[int64]corecvm.Cvm[corecvm.TCloudCvmExtension], error) {

	// 查询云主机的扩展信息
	extReq := &dataproto.CvmListReq{
		Filter: tools.ContainersExpression("id", cvmIDs),
		Page:   core.NewDefaultBasePage(),
	}
	cvmExtList := make([]corecvm.Cvm[corecvm.TCloudCvmExtension], 0)
	for {
		extResp, err := svc.client.DataService().TCloud.Cvm.ListCvmExt(kt.Ctx, kt.Header(), extReq)
		if err != nil {
			logs.Errorf("fail to list tcloud cvm ext map, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, kt.Rid)
			return nil, nil, err
		}

		cvmExtList = append(cvmExtList, extResp.Details...)
		if len(extResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		extReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	hostIDs := make([]int64, 0)
	hostCvmMap := make(map[int64]corecvm.Cvm[corecvm.TCloudCvmExtension], 0)
	for _, item := range cvmExtList {
		if item.BkHostID == 0 {
			continue
		}
		hostCvmMap[item.BkHostID] = item
		hostIDs = append(hostIDs, item.BkHostID)
	}
	hostIDs = slice.Unique(hostIDs)
	return hostIDs, hostCvmMap, nil
}
