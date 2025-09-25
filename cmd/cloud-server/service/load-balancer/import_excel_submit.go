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
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/cloud-server/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ImportSubmit CLB数据导入接口
func (svc *lbSvc) ImportSubmit(cts *rest.Contexts) (interface{}, error) {
	operationType := cts.PathParameter("operation_type").String()
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

	req := new(cslb.ImportExcelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	handlerOpt := &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  &types.CloudResourceBasicInfo{AccountID: req.AccountID},
	}
	if err = handler.BizOperateAuth(cts, handlerOpt); err != nil {
		return nil, err
	}

	executor, err := lblogic.NewImportExecutor(lblogic.OperationType(operationType), svc.client.DataService(),
		svc.client.TaskServer(), vendor, bizID, req.AccountID, req.RegionIDs)
	if err != nil {
		logs.Errorf("new ImportExecutor failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	taskID, err := executor.Execute(cts.Kit, req.Source, req.Details)
	if err != nil {
		logs.Errorf("execute ImportExecutor failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return task.CreateTaskManagementResp{TaskManagementID: taskID}, nil
}
