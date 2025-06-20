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

// Package loadbalancer ...
package loadbalancer

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/async"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
)

// DeleteBizTargetGroup delete biz target group.
func (svc *lbSvc) DeleteBizTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteTargetGroup(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) deleteTargetGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.TargetGroupCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list target group basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.TargetGroupAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit target group failed, ids: %v, err: %v, rid: %s",
			basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	// delete tcloud cloud target group
	err = svc.client.DataService().Global.LoadBalancer.DeleteTargetGroup(cts.Kit, &core.ListReq{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("[%s] request dataservice to delete target group failed, ids: %s, err: %v, rid: %s",
			enumor.TCloud, req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteLoadBalancer 批量删除负载均衡
func (svc *lbSvc) BatchDeleteLoadBalancer(cts *rest.Contexts) (any, error) {
	return svc.batchDeleteLoadBalancer(cts, handler.ResOperateAuth)
}

// BatchDeleteBizLoadBalancer 业务下批量删除负载均衡
func (svc *lbSvc) BatchDeleteBizLoadBalancer(cts *rest.Contexts) (any, error) {
	return svc.batchDeleteLoadBalancer(cts, handler.BizOperateAuth)
}

// batchDeleteLoadBalancer 批量删除负载均衡
func (svc *lbSvc) batchDeleteLoadBalancer(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (any, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	infoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.LoadBalancerCloudResType,
		IDs:          req.IDs,
		Fields:       append(types.CommonBasicInfoFields, "region"),
	}
	lbInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, infoReq)
	if err != nil {
		return nil, err
	}
	for _, lbID := range req.IDs {
		_, exist := lbInfoMap[lbID]
		if !exist {
			return nil, fmt.Errorf("load balancer(%s) not found", lbID)
		}
	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Delete,
		BasicInfos: lbInfoMap,
	})
	if err != nil {
		return nil, err
	}

	if err = svc.loadBalancerDeleteCheck(cts.Kit, req.IDs); err != nil {
		return nil, err
	}
	// 按规则删除审计
	err = svc.audit.ResDeleteAudit(cts.Kit, enumor.LoadBalancerAuditResType, req.IDs)
	if err != nil {
		logs.Errorf("create load balancer delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 按账号+地域分列表
	tasks := buildLBDeletionTasks(lbInfoMap)
	flowReq := &ts.AddCustomFlowReq{
		Name:        enumor.FlowDeleteLoadBalancer,
		ShareData:   nil,
		Tasks:       tasks,
		IsInitState: false,
	}
	flowResp, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		return nil, err
	}
	return nil, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), flowResp.ID)
}

// 负载均衡删除检查
func (svc *lbSvc) loadBalancerDeleteCheck(kt *kit.Kit, lbIDs []string) error {
	// 检查是否启用删除检查
	lbReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lbIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lbResp, err := svc.client.DataService().TCloud.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("fail to query load balancer for delete load balancers, err: %v, lb ids: %v, rid: %s",
			err, lbIDs, kt.Rid)
		return nil
	}
	for _, lb := range lbResp.Details {
		if cvt.PtrToVal(lb.Extension.DeleteProtect) {
			return fmt.Errorf("%s(%s) is protected for delection", lb.Name, lb.CloudID)
		}
	}

	// 检查是否存在监听器
	lblListReq := &core.ListReq{
		Filter: tools.ContainersExpression("lb_id", lbIDs),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblListReq)
	if err != nil {
		logs.Errorf("fail to query listener for delete load balancers, err: %v, lb ids: %v, rid: %s",
			err, lbIDs, kt.Rid)
		return nil
	}
	if len(listenerResp.Details) != 0 {
		lbl := listenerResp.Details[0]
		return fmt.Errorf("load balancer(%s) with listener(%s:%s) can not be deleted",
			lbl.CloudLbID, lbl.CloudID, lbl.Name)
	}
	return nil
}

func buildLBDeletionTasks(infoMap map[string]types.CloudResourceBasicInfo) (tasks []ts.CustomFlowTask) {
	reqMap := make(map[string]*actionlb.DeleteLoadBalancerOption, len(infoMap))
	for id, info := range infoMap {
		key := genAccountRegionKey(info)
		if reqMap[key] == nil {
			reqMap[key] = &actionlb.DeleteLoadBalancerOption{
				BatchDeleteLoadBalancerReq: hcproto.BatchDeleteLoadBalancerReq{
					AccountID: info.AccountID,
					Region:    info.Region,
					IDs:       []string{},
				},
				Vendor: info.Vendor,
			}

		}
		req := reqMap[key]
		req.IDs = append(req.IDs, id)
	}
	getNextID := counter.NewNumStringCounter(1, 10)
	for _, req := range reqMap {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(getNextID()),
			ActionName: enumor.ActionDeleteLoadBalancer,
			Params:     cvt.PtrToVal(req),
			Retry:      tableasync.NewRetryWithPolicy(3, 1000, 5000),
		})

	}
	return tasks
}

func genAccountRegionKey(info types.CloudResourceBasicInfo) string {
	return info.AccountID + "_" + string(info.Vendor) + "_" + info.Region
}
