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

package securitygroup

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/async"
	actionsg "hcm/cmd/task-server/logics/action/security-group"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchDeleteSecurityGroup batch delete security group.
func (svc *securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSecurityGroup(cts, handler.ResOperateAuth)
}

// BatchDeleteBizSecurityGroup batch delete biz security group.
func (svc *securityGroupSvc) BatchDeleteBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) batchDeleteSecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.SecurityGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgIDs := slice.Unique(req.IDs)
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		Fields:       types.CommonFieldsWithRegion,
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          sgIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("fail to list security group to delete, err: %v, ids: %v, rid: %s", err, sgIDs, cts.Kit.Rid)
		return nil, err
	}

	if len(basicInfoMap) != len(sgIDs) {
		return nil, errf.New(errf.RecordNotFound, "some security groups can not be found")
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// 检查是否还有绑定的资源
	if err = svc.checkSGBinding(cts.Kit, basicInfoMap); err != nil {
		return nil, err
	}
	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.SecurityGroupAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tasks := make([]ts.CustomFlowTask, 0, len(req.IDs))

	nextID := counter.NewNumStringCounter(1, 10)
	for _, info := range basicInfoMap {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(nextID()),
			ActionName: enumor.ActionDeleteSecurityGroup,
			Params: actionsg.DeleteSGOption{
				Vendor: info.Vendor,
				ID:     info.ID,
			},
			DependOn: nil,
		})
	}
	flowReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowDeleteSecurityGroup,
		Tasks: tasks,
	}

	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID); err != nil {
		return nil, err
	}
	return result, nil
}

type statSGFunc func(kt *kit.Kit, req *hcservice.ListSecurityGroupStatisticReq) (
	*hcservice.ListSecurityGroupStatisticResp, error)

// check if sg is binding to any resource, if yes, return error.
func (svc *securityGroupSvc) checkSGBinding(kt *kit.Kit, sgInfos map[string]types.CloudResourceBasicInfo) error {
	sgVendorMap := classifier.ClassifyBasicInfoByVendor(sgInfos)

	for vendor := range sgVendorMap {
		var sgList = sgVendorMap[vendor]
		statSGBindingRes, err := svc.getSGBingResFunc(vendor)
		if err != nil {
			return err
		}
		accountRegionMap := classifier.ClassifyBasicInfoByAccount(sgList)
		for accountID := range accountRegionMap {
			regionMap := accountRegionMap[accountID]
			for region := range regionMap {
				statReq := &hcservice.ListSecurityGroupStatisticReq{
					SecurityGroupIDs: regionMap[region],
					Region:           region,
					AccountID:        accountID,
				}
				stats, err := statSGBindingRes(kt, statReq)
				if err != nil {
					logs.Errorf("stat %s security group binding failed, err: %v, req: %+v, rid: %s",
						vendor, err, statReq, kt.Rid)
					return err
				}
				if stats == nil {
					return fmt.Errorf("stat %s security group binding failed", vendor)
				}
				if err := checkStat(stats.Details); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (svc *securityGroupSvc) getSGBingResFunc(vendor enumor.Vendor) (statSGFunc, error) {
	var statSGBindingRes statSGFunc
	switch vendor {
	case enumor.TCloud:
		statSGBindingRes = svc.client.HCService().TCloud.SecurityGroup.ListSecurityGroupStatistic
	case enumor.Aws:
		statSGBindingRes = svc.client.HCService().Aws.SecurityGroup.ListSecurityGroupStatistic
	case enumor.HuaWei:
		statSGBindingRes = svc.client.HCService().HuaWei.SecurityGroup.ListSecurityGroupStatistic
	case enumor.Azure:
		statSGBindingRes = svc.client.HCService().Azure.SecurityGroup.ListSecurityGroupStatistic
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor for sg binding check: %s", vendor)
	}
	return statSGBindingRes, nil
}

func checkStat(stats []*hcservice.SecurityGroupStatisticItem) error {
	for i := range stats {
		stat := stats[i]
		for j := range stat.Resources {
			if stat.Resources[j].Count <= 0 {
				continue
			}
			return fmt.Errorf("security group %s is binding to resource: %s, count: %d",
				stat.ID, stat.Resources[j].ResName, stat.Resources[j].Count)
		}
	}
	return nil
}
