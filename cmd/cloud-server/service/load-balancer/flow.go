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

package loadbalancer

import (
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/api/hc-service/sync"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (svc *lbSvc) buildFlow(kt *kit.Kit, flowName enumor.FlowName, shareData *tableasync.ShareData,
	tasks []ts.CustomFlowTask) (flowID string, err error) {

	addReq := &ts.AddCustomFlowReq{
		Name:        flowName,
		ShareData:   shareData,
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch add rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return result.ID, nil
}

func (svc *lbSvc) buildSubFlow(kt *kit.Kit, flowID, lbID string,
	subResIDs []string, subResType enumor.CloudResourceType, taskType enumor.TaskType) error {

	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  subResIDs,
				SubResType: subResType,
				TaskType:   taskType,
			},
		}},
	}
	_, err := svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return err
	}
	return nil
}

func buildSyncClbFlowTask(vendor enumor.Vendor, lbCloudID, accountID, region string,
	generator func() (cur string, prev string)) ts.CustomFlowTask {

	cur, prev := generator()
	tmpTask := ts.CustomFlowTask{
		ActionID:   action.ActIDType(cur),
		ActionName: enumor.ActionSyncTCloudLoadBalancer,
		Params: &actionlb.SyncTCloudLoadBalancerOption{
			Vendor: vendor,
			TCloudSyncReq: &sync.TCloudSyncReq{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  []string{lbCloudID},
			},
		},
		Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
	}
	if prev != "" {
		tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
	}
	return tmpTask
}
