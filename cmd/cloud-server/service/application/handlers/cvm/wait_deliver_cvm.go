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

package appcvm

import (
	"fmt"
	"time"

	"hcm/cmd/cloud-server/logics/tenant"
	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	"hcm/pkg/api/core"
	coreasync "hcm/pkg/api/core/async"
	ds "hcm/pkg/api/data-service"
	hccvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
)

// TimingHandleDeliverApplication 定时处理处于回收状态的单据。
func TimingHandleDeliverApplication(cliSet *client.ClientSet, interval time.Duration) {
	for {
		time.Sleep(2 * time.Second)

		kt := core.NewBackendKit()
		if err := WaitAndHandleDeliverCvmByTenant(kt, cliSet.DataService(), cliSet.TaskServer()); err != nil {
			logs.Errorf("WaitAndHandleDeliverCvm err: %v, rid: %s", err, kt.Rid)
		}
	}
}

// WaitAndHandleDeliverCvmByTenant wait deliver cvm.
func WaitAndHandleDeliverCvmByTenant(kt *kit.Kit, dsCli *dataservice.Client, tsCli *taskserver.Client) error {
	tenantIDs, err := tenant.ListAllTenantID(kt, dsCli)
	if err != nil {
		logs.Errorf("failed to list all tenant ids, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	eg, _ := errgroup.WithContext(kt.Ctx)
	for _, id := range tenantIDs {
		tenantID := id
		eg.Go(func() error {
			tenantKt := kt.NewSubKitWithTenant(tenantID)
			subErr := WaitAndHandleDeliverCvm(tenantKt, dsCli, tsCli)
			if subErr != nil {
				logs.Errorf("failed to wait and handle deliver cvm, err: %v, tenant: %s, rid: %s", subErr,
					tenantID, tenantKt.Rid)
				return subErr
			}
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return err
	}
	return nil
}

// WaitAndHandleDeliverCvm wait deliver cvm.
func WaitAndHandleDeliverCvm(kt *kit.Kit, dsCli *dataservice.Client, tsCli *taskserver.Client) error {

	// 查询交付状态中的单据
	apps, err := queryDeliveringApplication(kt, dsCli)
	if err != nil {
		return err
	}

	// 如果没有需要处理的单据跳过即可
	if len(apps) == 0 {
		logs.V(5).Infof("delivering application not found, skip handle, rid: %s", kt.Rid)
		return nil
	}

	flowIDs := make([]string, 0, len(apps))
	flowAppMap := make(map[string]*ds.ApplicationResp, len(apps))
	for _, app := range apps {
		flowID := gjson.Get(app.DeliveryDetail, "flow_id").String()
		flowIDs = append(flowIDs, flowID)
		flowAppMap[flowID] = app
	}

	// 查询处于结束状态的Flow
	flowResultMap, err := queryAndParseEndStateFlowByFlowID(kt, tsCli, flowIDs)
	if err != nil {
		return err
	}

	if len(flowResultMap) == 0 {
		logs.V(5).Infof("flow not found, that in end state, skip handle, rid: %s", kt.Rid)
		return nil
	}

	// 将生产出来的机器及其关联资源分配到业务下，并将结果保存到单据中
	for flowID, result := range flowResultMap {
		if err = handleDeliverCvm(kt, dsCli, tsCli, flowAppMap, flowID, result); err != nil {
			return err
		}
	}

	return nil
}

func handleDeliverCvm(kt *kit.Kit, dsCli *dataservice.Client, tsCli *taskserver.Client,
	flowAppMap map[string]*ds.ApplicationResp, flowID string, result *hccvm.BatchCreateResult) error {

	app := flowAppMap[flowID]

	state := enumor.Completed
	detail := map[string]interface{}{
		"result": result,
	}
	if len(result.SuccessCloudIDs) != 0 {
		req := &core.ListReq{
			Filter: tools.EqualWithOpExpression(filter.And, map[string]interface{}{
				"flow_id":     flowID,
				"action_name": enumor.ActionAssignCvm,
			}),
			Page: core.NewDefaultBasePage(),
		}
		listResult, err := tsCli.ListTask(kt, req)
		if err != nil {
			logs.Errorf("list task failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if len(listResult.Details) == 0 {
			return fmt.Errorf("flow: %s not found assign cvm action", flowID)
		}

		task := listResult.Details[0]
		assignResult := new(actioncvm.AssignCvmResult)
		if err = json.UnmarshalFromString(string(task.Result), &assignResult); err != nil {
			logs.Errorf("unmarshal task result failed, err: %v, result: %s, rid: %s", err, assignResult, kt.Rid)
			return err
		}

		detail["cvm_ids"] = assignResult.IDs
		requiredCount := gjson.Get(app.Content, "required_count").Int()
		if len(result.SuccessCloudIDs) != int(requiredCount) {
			state = enumor.DeliverPartial
		}
	} else {
		state = enumor.DeliverError
	}

	marshal, err := json.Marshal(detail)
	if err != nil {
		logs.Errorf("marshal delver result failed, err: %v, detail: %+v, rid: %s", err, detail, kt.Rid)
		return err
	}

	req := &ds.ApplicationUpdateReq{
		Status:         state,
		DeliveryDetail: converter.ValToPtr(string(marshal)),
	}
	if _, err := dsCli.Global.Application.UpdateApplication(kt, app.ID, req); err != nil {
		logs.Errorf("update application failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// queryAndParseEndStateFlowByFlowID 查询并解析结束状态Flow
func queryAndParseEndStateFlowByFlowID(kt *kit.Kit, cli *taskserver.Client, flowIDs []string) (
	map[string]*hccvm.BatchCreateResult, error) {

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.In.Factory(),
					Value: []enumor.FlowState{enumor.FlowSuccess, enumor.FlowFailed},
				},
				&filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: flowIDs,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	flowResult, err := cli.ListFlow(kt, req)
	if err != nil {
		logs.Errorf("list flow failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(flowResult.Details) == 0 {
		return nil, nil
	}

	endFlowIDs := make([]string, 0, len(flowResult.Details))
	for _, one := range flowResult.Details {
		endFlowIDs = append(endFlowIDs, one.ID)
	}

	tasks, err := queryTaskByFlowIDs(kt, cli, endFlowIDs)
	if err != nil {
		logs.Errorf("query task by flowIDs failed, err: %v, flowIDs: %+v, rid: %s", err, endFlowIDs, kt.Rid)
		return nil, err
	}

	flowResultMap := make(map[string]*hccvm.BatchCreateResult)
	for _, task := range tasks {
		tmp := new(hccvm.BatchCreateResult)
		if err := json.UnmarshalFromString(string(task.Result), tmp); err != nil {
			logs.Errorf("unmarshal tasks result failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if _, exist := flowResultMap[task.FlowID]; !exist {
			flowResultMap[task.FlowID] = new(hccvm.BatchCreateResult)
		}

		result := flowResultMap[task.FlowID]
		result.SuccessCloudIDs = append(result.SuccessCloudIDs, tmp.SuccessCloudIDs...)
		result.FailedCloudIDs = append(result.FailedCloudIDs, tmp.FailedCloudIDs...)
		result.UnknownCloudIDs = append(result.UnknownCloudIDs, tmp.UnknownCloudIDs...)
	}

	return flowResultMap, nil
}

func queryTaskByFlowIDs(kt *kit.Kit, cli *taskserver.Client, endFlowIDs []string) (
	[]coreasync.AsyncFlowTask, error) {

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "flow_id",
					Op:    filter.In.Factory(),
					Value: endFlowIDs,
				},
				&filter.AtomRule{
					Field: "action_name",
					Op:    filter.Equal.Factory(),
					Value: enumor.ActionCreateCvm,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	tasks := make([]coreasync.AsyncFlowTask, 0)
	for {
		taskResult, err := cli.ListTask(kt, req)
		if err != nil {
			logs.Errorf("list task failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		tasks = append(tasks, taskResult.Details...)

		if len(taskResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return tasks, nil
}

// queryDeliveringApplication 查询处于交付状态的单据
func queryDeliveringApplication(kt *kit.Kit, cli *dataservice.Client) ([]*ds.ApplicationResp, error) {
	req := &ds.ApplicationListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "status",
					Op:    filter.Equal.Factory(),
					Value: enumor.Delivering,
				},
				&filter.AtomRule{
					Field: "type",
					Op:    filter.Equal.Factory(),
					Value: enumor.CreateCvm,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.Application.ListApplication(kt, req)
	if err != nil {
		logs.Errorf("list application failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}
