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

package tcloud

import (
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cloudtask "hcm/pkg/api/cloud-server/task"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// taskManagementAdapter adapts dataservice.Client to lblogic.taskManagementLister.
type taskManagementAdapter struct {
	cli *dataservice.Client
}

// List 查询 task_management 列表。
func (a taskManagementAdapter) List(kt *kit.Kit, req *core.ListReq) (*task.ListManagementResult, error) {
	return a.cli.Global.TaskManagement.List(kt, req)
}

// flowAdapter adapts taskserver.Client to lblogic.flowLister.
type flowAdapter struct {
	cli *taskserver.Client
}

// ListFlow 查询 flow 列表。
func (a flowAdapter) ListFlow(kt *kit.Kit, req *core.ListReq) (*ts.ListFlowResult, error) {
	return a.cli.ListFlow(kt, req)
}

// AsyncCondSyncLoadBalancer creates an asynchronous CLB conditional sync task.
func AsyncCondSyncLoadBalancer(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) (any, error) {
	opt := &lblogic.CondSyncLoadBalancerOption{
		Vendor:     enumor.TCloud,
		AccountID:  params.AccountID,
		BkBizID:    params.BkBizID,
		Regions:    params.Regions,
		CloudIDs:   params.CloudIDs,
		TagFilters: params.TagFilters,
	}

	mgmtCli := taskManagementAdapter{cli: cliSet.DataService()}
	flowCli := flowAdapter{cli: cliSet.TaskServer()}
	if err := lblogic.CheckRunningCondSyncTask(kt, mgmtCli, flowCli, opt); err != nil {
		return nil, err
	}

	regionDiffs := make([]lblogic.LoadBalancerRegionDiff, 0, len(opt.Regions))
	total := 0
	for _, region := range opt.Regions {
		diff, err := lblogic.ListAndDiffLoadBalancerByRegion(kt, cliSet, opt, region)
		if err != nil {
			logs.Errorf("[%s] diff conditional sync load balancer failed, err: %v, account: %s, region: %s, rid: %s",
				enumor.TCloud, err, opt.AccountID, region, kt.Rid)
			return nil, err
		}
		regionDiffs = append(regionDiffs, *diff)
		total += len(diff.Create) + len(diff.Update) + len(diff.Delete)
	}

	// total=0 表示请求条件下没有任何可处理的 CLB。
	// 两侧已有同一批 CLB 时全部进入 update，仍会创建任务。
	if total == 0 {
		logs.Infof("[%s] skip create conditional sync load balancer task because no processable load balancer, "+
			"account: %s, rid: %s",
			enumor.TCloud, opt.AccountID, kt.Rid)
		return cloudtask.CreateTaskManagementResp{}, nil
	}

	taskManagementID, err := lblogic.CreateCondSyncTask(kt, cliSet, opt, regionDiffs)
	if err != nil {
		logs.Errorf("[%s] create conditional sync load balancer task failed, err: %v, account: %s, rid: %s",
			enumor.TCloud, err, opt.AccountID, kt.Rid)
		return nil, err
	}

	return cloudtask.CreateTaskManagementResp{TaskManagementID: taskManagementID}, nil
}
