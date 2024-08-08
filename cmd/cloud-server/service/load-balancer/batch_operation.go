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

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	loadbalancer "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/audit"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

type initBatchTaskFunc[T any] func(kt *kit.Kit, listenerList []*T,
	lb *corelb.BaseLoadBalancer) (string, error)

func buildAsyncFlow[T any](kt *kit.Kit, svc *lbSvc, listenerList []*T,
	lb *corelb.BaseLoadBalancer, opFunc initBatchTaskFunc[T]) (string, error) {
	// 预检测
	_, err := svc.checkResFlowRel(kt, lb.ID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check res flow rel failed, lb: %s, err: %v, rid: %s", lb.ID, err, kt.Rid)
		return "", err
	}
	flowID, err := opFunc(kt, listenerList, lb)
	if err != nil {
		logs.Errorf("init batch modify weight task failed, lb: %s, err: %v, rid: %s", lb.ID, err, kt.Rid)
		return "", err
	}
	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lb.ID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyWeightTaskType)
	if err != nil {
		logs.Errorf("lock res flow status failed, lb: %s, err: %v, rid: %s", lb.ID, err, kt.Rid)
		return "", err
	}
	return flowID, nil
}

func (svc *lbSvc) getLoadBalancersByID(kt *kit.Kit, bizID int64, lbID string) (*corelb.BaseLoadBalancer, error) {
	expr := tools.ExpressionAnd(
		tools.RuleEqual("id", lbID),
		tools.RuleEqual("bk_biz_id", bizID),
	)
	lbReq := &core.ListReq{
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
	}
	loadBalancers, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("")
		return nil, err
	}
	if len(loadBalancers.Details) == 0 {
		logs.Errorf("biz[%d] load balancer (%s) not found, rid: %s", bizID, lbID, kt.Rid)
		return nil, fmt.Errorf("biz[%d] load balancer (%s) not found", bizID, lbID)
	}
	return &loadBalancers.Details[0], nil
}

func (svc *lbSvc) getAuditByLoadBalanceID(kt *kit.Kit, lbID string) (*audit.Audit, error) {
	filter := tools.ExpressionAnd(
		tools.RuleEqual("res_id", lbID),
		tools.RuleEqual("res_type", enumor.LoadBalancerAuditResType),
		tools.RuleEqual("rid", kt.Rid),
	)
	listReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	audits, err := svc.client.DataService().Global.Audit.ListAudit(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(audits.Details) == 0 {
		return nil, fmt.Errorf("audit not found for load balancer(%s)", lbID)
	}
	return &audits.Details[0], nil
}

func (svc *lbSvc) saveBatchOperationRecord(cts *rest.Contexts, detail string,
	flowAuditMap map[string]uint64, accountID string) (string, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return "", err
	}
	batchOperationIDs, err := svc.client.DataService().Global.LoadBalancer.BatchCreateBatchOperation(cts.Kit,
		&cloud.BatchOperationBatchCreateReq{
			AccountID: accountID,
			Tasks: []*cloud.BatchOperationCreateReq{
				{
					BkBizID: bizID,
					Detail:  detail,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	batchOperationID := batchOperationIDs.IDs[0]
	createReq := &cloud.BatchOperationAsyncFlowRelBatchCreateReq{
		Rels: make([]*cloud.BatchOperationAsyncFlowRelCreateReq, 0, len(flowAuditMap)),
	}
	for flowID, auditID := range flowAuditMap {
		createReq.Rels = append(createReq.Rels, &cloud.BatchOperationAsyncFlowRelCreateReq{
			BatchOperationID: batchOperationID,
			FlowID:           flowID,
			AuditID:          &auditID,
		})
	}

	_, err = svc.client.DataService().Global.LoadBalancer.BatchCreateBatchOperationAsyncFlowRel(
		cts.Kit, createReq)
	if err != nil {
		return "", err
	}
	return batchOperationID, nil
}

func (svc *lbSvc) buildBatchOperationFlow(kt *kit.Kit, lbID string, flowName enumor.FlowName,
	tasks []ts.CustomFlowTask, targetGroupIDs []string) (string, error) {

	rsWeightReq := &ts.AddCustomFlowReq{
		Name: flowName,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, rsWeightReq)
	if err != nil {
		logs.Errorf("call taskserver to create %s custom flow failed, err: %v, rid: %s", flowName, err, kt.Rid)
		return "", err
	}

	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     result.ID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  targetGroupIDs,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ModifyWeightTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, result.ID, kt.Rid)
		return "", err
	}
	return result.ID, nil
}

func getHealthCheck(isHealthCheck bool) *corelb.TCloudHealthCheckInfo {
	var flag int64 = 0
	if isHealthCheck {
		flag = 1
	}
	return &corelb.TCloudHealthCheckInfo{HealthSwitch: &flag}
}

func getCertInfo(listener *lblogic.BindRSRecord) *corelb.TCloudCertificateInfo {
	if listener.Protocol != enumor.HttpsProtocol {
		return nil
	}
	certObj := &corelb.TCloudCertificateInfo{
		CertCloudIDs: listener.ServerCerts,
	}
	if len(listener.ClientCert) > 0 {
		certObj.CaCloudID = cvt.ValToPtr(listener.ClientCert)
		certObj.SSLMode = (*string)(cvt.ValToPtr(loadbalancer.TCloudSslMutual))
	} else {
		certObj.SSLMode = (*string)(cvt.ValToPtr(loadbalancer.TCloudSslUniDirect))
	}
	return certObj
}

// GetBatchOperation 获取批量操作详情
func (svc *lbSvc) GetBatchOperation(cts *rest.Contexts) (interface{}, error) {
	batchOperationID := cts.PathParameter("batch_operation_id").String()
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	expr := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", batchOperationID),
			tools.RuleEqual("bk_biz_id", bizID),
		),
		Page: core.NewDefaultBasePage(),
	}

	operations, err := svc.client.DataService().Global.LoadBalancer.ListBatchOperation(cts.Kit, expr)
	if err != nil {
		return nil, err
	}
	if len(operations.Details) == 0 {
		return nil, fmt.Errorf("batch operation[%s] not found", batchOperationID)
	}

	op := operations.Details[0]
	preview := make([]interface{}, 0)
	if err = json.Unmarshal([]byte(op.Detail), &preview); err != nil {
		return nil, err
	}

	result := &cloud.BatchOperationResult{
		AuditId:          op.AuditID,
		BatchOperationID: op.ID,
		Preview:          preview,
	}

	// get async flows
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("batch_operation_id", op.ID),
		),
		Page: core.NewDefaultBasePage(),
	}
	flows, err := svc.client.DataService().Global.LoadBalancer.ListBatchOperationAsyncFlowRel(
		cts.Kit,
		req,
	)
	if err != nil {
		return nil, err
	}
	if len(flows.Details) == 0 {
		return result, nil
	}
	result.Flows = make([]cloud.AsyncFlow, 0, len(flows.Details))
	for _, detail := range flows.Details {
		result.Flows = append(result.Flows, cloud.AsyncFlow{
			FlowId:  detail.FlowID,
			AuditId: detail.AuditID,
		})
	}

	return result, nil
}
