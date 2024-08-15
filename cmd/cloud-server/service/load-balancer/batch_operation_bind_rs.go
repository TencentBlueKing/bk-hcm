/*
 *
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

package loadbalancer

import (
	"encoding/json"
	"fmt"
	"strings"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BindRS 绑定RS接口
func (svc *lbSvc) BindRS(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.BatchOperationReq[*lblogic.BindRSRecord])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch operation modify weight auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.bindRS(cts, req)
}

func (svc *lbSvc) bindRS(cts *rest.Contexts, req *cloud.BatchOperationReq[*lblogic.BindRSRecord]) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	// validate request
	for _, tmp := range req.Data {
		for _, record := range tmp.Listeners {
			errList := record.CheckWithDataService(cts.Kit, svc.client.DataService(), bizID)
			if len(errList) > 0 {
				logs.Errorf("validate request failed, err: %v, rid: %s", errList, cts.Kit.Rid)
				errStr := &strings.Builder{}
				for _, validateError := range errList {
					errStr.WriteString(validateError.Reason)
					errStr.WriteString(";")
				}
				return nil, fmt.Errorf("validate request failed, err: %s", errStr.String())
			}
		}
	}

	flows := make([]string, 0)
	flowAuditMap := make(map[string]uint64)
	for _, tmp := range req.Data {
		// 一个CLB一个异步任务
		lb, err := svc.getLoadBalancersByID(cts.Kit, bizID, tmp.ClbID)
		if err != nil {
			logs.Errorf("get load balancer failed, lbID: %s, err: %v, rid: %s", tmp.ClbID, err, cts.Kit.Rid)
			return nil, err
		}

		flowID, err := buildAsyncFlow(cts.Kit, svc, tmp.Listeners, lb, svc.initBatchBindRSTask)
		if err != nil {
			logs.Errorf("build async flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		auditRecord, err := svc.getAuditByLoadBalanceID(cts.Kit, lb.ID)
		if err != nil {
			logs.Errorf("get audit failed, lbID: %s, err: %v, rid: %s", lb.ID, err, cts.Kit.Rid)
			return nil, err
		}
		flows = append(flows, flowID)
		flowAuditMap[flowID] = auditRecord.ID
	}

	// save preview data to db
	detail, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	batchOperationID, err := svc.saveBatchOperationRecord(cts, string(detail), flowAuditMap, req.AccountID)
	if err != nil {
		logs.Errorf("save batch operation record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return batchOperationID, nil
}

func (svc *lbSvc) initBatchBindRSTask(kt *kit.Kit, listeners []*lblogic.BindRSRecord,
	lb *corelb.BaseLoadBalancer) (string, error) {

	createTasks := make([]ts.CustomFlowTask, 0)
	bindRSTasks := make([]ts.CustomFlowTask, 0)
	tgIDs := make([]string, 0)
	for _, listener := range listeners {
		tgID, createListenerTasks, bindTasks, err := svc.buildBatchBindRSTask(kt, listener, lb)
		if err != nil {
			logs.Errorf("build batch bind rs task failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}
		createTasks = append(createTasks, createListenerTasks...)
		bindRSTasks = append(bindRSTasks, bindTasks...)
		tgIDs = append(tgIDs, tgID)
	}

	// 对所有task进行排序并指定actionID
	getActionID := counter.NewNumStringCounter(0, 10)
	var lastActionID action.ActIDType
	for i := 0; i < len(createTasks); i++ {
		actionID := action.ActIDType(getActionID())
		createTasks[i].ActionID = actionID
		if len(lastActionID) > 0 {
			createTasks[i].DependOn = []action.ActIDType{lastActionID}
		}
		lastActionID = actionID
	}
	for i := 0; i < len(bindRSTasks); i++ {
		actionID := action.ActIDType(getActionID())
		bindRSTasks[i].ActionID = actionID
		if len(lastActionID) > 0 {
			bindRSTasks[i].DependOn = []action.ActIDType{lastActionID}
		}
		lastActionID = actionID
	}
	tasks := append(createTasks, bindRSTasks...)

	flowID, err := svc.buildBatchOperationFlow(kt, lb.ID, enumor.FlowLoadBalancerBatchOperation, tasks, tgIDs)
	if err != nil {
		logs.Errorf("build batch operation flow failed, err: %v, rid: %s, tasks: %v", err, kt.Rid, tasks)
		return "", err
	}
	return flowID, nil
}

func (svc *lbSvc) buildBatchBindRSTask(kt *kit.Kit, listener *lblogic.BindRSRecord,
	lb *corelb.BaseLoadBalancer) (targetGroupID string, createTasks []ts.CustomFlowTask, bindTasks []ts.CustomFlowTask, err error) {

	if listener.Action != enumor.AppendRS {
		switch lb.Vendor {
		case enumor.TCloud:
			targetGroupID, err = svc.createTCloudTargetGroup(kt, listener, lb)
			if err != nil {
				logs.Errorf("create tcloud target group failed, err: %v, rid: %s", err, kt.Rid)
				return "", nil, nil, err
			}
		default:
			return "", nil, nil, fmt.Errorf("unsupported vendor %s", lb.Vendor)
		}
	} else {
		err := listener.LoadDataFromDB(kt, svc.client.DataService(), lb)
		if err != nil {
			logs.Errorf("load data from db failed, err: %v, rid: %s", err, kt.Rid)
			return "", nil, nil, err
		}
		targetGroupID = listener.ListenerID
	}

	switch listener.Action {
	case enumor.CreateURLAndAppendRS:
		task, err := svc.buildCreateURLTask(kt, listener, lb.ID, targetGroupID, lb.Vendor)
		if err != nil {
			logs.Errorf("build create url task failed, err: %v, rid: %s", err, kt.Rid)
			return "", nil, nil, err
		}
		createTasks = append(createTasks, *task)
	case enumor.CreateListenerWithURLAndAppendRS, enumor.CreateListenerAndAppendRS:
		task := svc.buildCreateListenerTask(listener, lb.ID, targetGroupID, lb.BkBizID, lb.Vendor)
		createTasks = append(createTasks, task)
	}

	tasks, err := svc.buildBindRSTasks(kt, lb, listener, targetGroupID)
	if err != nil {
		logs.Errorf("build bind rs tasks failed, err: %v, rid: %s", err, kt.Rid)
		return "", nil, nil, err
	}
	bindTasks = append(bindTasks, tasks...)
	return
}

func (svc *lbSvc) buildBindRSTasks(kt *kit.Kit, lb *corelb.BaseLoadBalancer,
	listener *lblogic.BindRSRecord, tgID string) ([]ts.CustomFlowTask, error) {

	elems := slice.Split(listener.RSInfos, constant.BatchAddRSCloudMaxLimit)
	result := make([]ts.CustomFlowTask, 0, len(listener.RSInfos))
	for _, parts := range elems {
		addRsParams := &hcproto.TCloudBatchOperateTargetReq{
			TargetGroupID: tgID,
			LbID:          lb.ID,
		}
		for _, rs := range parts {
			targetReq, err := rs.GetTargetReq(kt, lb.Vendor, lb.BkBizID, tgID, lb.AccountID,
				svc.client.DataService(), svc.cvmLgc)
			if err != nil {
				logs.Errorf("get target req failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			addRsParams.RsList = append(addRsParams.RsList, targetReq)
		}

		tmpTask := ts.CustomFlowTask{
			ActionName: enumor.ActionTargetGroupAddRS,
			Params: &actionlb.OperateRsOption{
				Vendor:                      lb.Vendor,
				TCloudBatchOperateTargetReq: *addRsParams,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		result = append(result, tmpTask)
	}

	return result, nil
}

func (svc *lbSvc) createTCloudTargetGroup(kt *kit.Kit, listener *lblogic.BindRSRecord,
	lb *corelb.BaseLoadBalancer) (string, error) {

	targets, err := svc.client.DataService().TCloud.LoadBalancer.BatchCreateTCloudTargetGroup(kt,
		&cloud.TCloudTargetGroupCreateReq{
			TargetGroups: []cloud.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
				{
					Name:            fmt.Sprintf("auto-%s", listener.ListenerName),
					Vendor:          lb.Vendor,
					AccountID:       lb.AccountID,
					BkBizID:         lb.BkBizID,
					Region:          lb.Region,
					Protocol:        listener.Protocol,
					Port:            int64(listener.VPorts[0]),
					VpcID:           lb.VpcID,
					CloudVpcID:      lb.CloudVpcID,
					TargetGroupType: enumor.LocalTargetGroupType,
					Weight:          0,
					Memo:            cvt.ValToPtr("auto created for listener " + listener.ListenerName),
				},
			},
		},
	)
	if err != nil {
		logs.Errorf("create tcloud target group failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	if len(targets.IDs) == 0 {
		return "", fmt.Errorf("create target group failed")
	}

	return targets.IDs[0], nil
}

func (svc *lbSvc) buildCreateURLTask(kt *kit.Kit, listener *lblogic.BindRSRecord,
	lbID, targetGroupID string, vendor enumor.Vendor) (*ts.CustomFlowTask, error) {

	listenerID, err := listener.GetListenerID(kt, svc.client.DataService().Global.LoadBalancer, lbID)
	if err != nil {
		logs.Errorf("get listener id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// TODO 支持多vendor实现
	req := hcproto.TCloudRuleBatchCreateReq{
		Rules: []hcproto.TCloudRuleCreate{
			{
				Url:                listener.URLPath,
				TargetGroupID:      targetGroupID,
				CloudTargetGroupID: targetGroupID,
				Domains:            []string{listener.Domain},
				SessionExpireTime:  &listener.SessionExpired,
				Scheduler:          &listener.Scheduler,
				HealthCheck:        getHealthCheck(listener.HealthCheck),
				Certificates:       getCertInfo(listener),
			},
		},
	}
	task := &ts.CustomFlowTask{
		ActionName: enumor.ActionURLRuleCreate,
		Params: &actionlb.CreateURLRuleOption{
			Vendor:                   vendor,
			ListenerID:               listenerID,
			TCloudRuleBatchCreateReq: req,
		},
		Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
	}
	return task, nil
}

func (svc *lbSvc) buildCreateListenerTask(listener *lblogic.BindRSRecord, lbID, tgID string, bkBizID int64,
	vendor enumor.Vendor) ts.CustomFlowTask {

	var endPort uint64
	if len(listener.VPorts) == 2 {
		endPort = uint64(listener.VPorts[1])
	}
	req := hcproto.ListenerWithRuleCreateReq{
		Name:          listener.ListenerName,
		BkBizID:       bkBizID,
		LbID:          lbID,
		Protocol:      listener.Protocol,
		Port:          int64(listener.VPorts[0]),
		Scheduler:     listener.Scheduler,
		SessionExpire: listener.SessionExpired,
		TargetGroupID: tgID,
		Domain:        listener.Domain,
		Url:           listener.URLPath,
		SessionType:   "NORMAL",
		//SniSwitch:     0,
		HealthCheck: getHealthCheck(listener.HealthCheck),
		Certificate: getCertInfo(listener),
		EndPort:     endPort,
	}

	task := ts.CustomFlowTask{
		ActionName: enumor.ActionListenerCreate,
		Params: &actionlb.CreateListenerOption{
			Vendor:                    vendor,
			ListenerWithRuleCreateReq: req,
		},
		Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
	}
	return task
}
