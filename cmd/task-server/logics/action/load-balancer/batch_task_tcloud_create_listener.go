/*
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

package actionlb

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
)

// --------------------------[创建TCloud监听器]-----------------------------

var _ action.Action = new(BatchTaskTCloudCreateListenerAction)
var _ action.ParameterAction = new(BatchTaskTCloudCreateListenerAction)

// BatchTaskTCloudCreateListenerAction 创建TCloud监听器
type BatchTaskTCloudCreateListenerAction struct{}

// BatchTaskTCloudCreateListenerOption ...
type BatchTaskTCloudCreateListenerOption struct {
	ManagementDetailIDs []string                        `json:"management_detail_ids" validate:"required,min=1,max=20"`
	Listeners           []*hclb.TCloudListenerCreateReq `json:"Listeners,required,min=1,max=20,dive,required"`
}

// Validate validate option.
func (opt BatchTaskTCloudCreateListenerOption) Validate() error {
	if len(opt.ManagementDetailIDs) != len(opt.Listeners) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and listeners length not match: %d! = %d",
			len(opt.ManagementDetailIDs), len(opt.Listeners))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskTCloudCreateListenerAction) ParameterNew() (params any) {
	return new(BatchTaskTCloudCreateListenerOption)
}

// Name return action name
func (act BatchTaskTCloudCreateListenerAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudCreateListener
}

// Run 创建监听器
func (act BatchTaskTCloudCreateListenerAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskTCloudCreateListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	results := make([]*hclb.ListenerCreateResult, 0, len(opt.Listeners))
	for i := range opt.Listeners {
		detailID := opt.ManagementDetailIDs[i]
		ret, createErr := act.createSingleListener(kt.Kit(), detailID, opt.Listeners[i]) // 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if createErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailState(kt.Kit(), []string{detailID}, targetState)
		if err != nil {
			logs.Errorf("fail to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
			return nil, err
		}
		if createErr != nil {
			// abort
			return nil, err
		}
		results = append(results, ret)
	}
	// all success
	return results, nil
}

func (act BatchTaskTCloudCreateListenerAction) createSingleListener(kt *kit.Kit, detailId string,
	req *hclb.TCloudListenerCreateReq) (*hclb.ListenerCreateResult, error) {
	detailList, err := listTaskDetail(kt, []string{detailId})
	if err != nil {
		logs.Errorf("fail to query task detail, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	detail := detailList[0]
	if detail.State == enumor.TaskDetailCancel {
		// 任务被取消，跳过该任务, 直接成功即可
		return nil, nil
	}
	if detail.State != enumor.TaskDetailInit {
		return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
			detail.ID, detail.State)
	}
	exists, err := act.checkListenerExists(kt, req)
	if err != nil {
		return nil, err
	}
	if exists {
		// 已存在跳过
		return nil, nil
	}

	// 更新任务状态为 running
	if err := batchUpdateTaskDetailState(kt, []string{detailId}, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("fail to update detail to running, err: %v", err)
	}
	lblResp, err := actcli.GetHCService().TCloud.Clb.CreateListener(kt, req)
	if err != nil {
		logs.Errorf("fail to call hc to create listener, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return lblResp, err
}

func (act BatchTaskTCloudCreateListenerAction) checkListenerExists(kt *kit.Kit,
	req *hclb.TCloudListenerCreateReq) (exists bool, err error) {
	// 查询是否已经存在对应监听器
	lbReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("lb_id", req.LbID),
			tools.RuleEqual("protocol", req.Protocol),
			tools.RuleEqual("port", req.Port),
		),
		Page: core.NewDefaultBasePage(),
	}
	lblResp, err := actcli.GetDataService().TCloud.LoadBalancer.ListListener(kt, lbReq)
	if err != nil {
		return false, fmt.Errorf("fail to query listener, err: %v", err)
	}
	if len(lblResp.Details) == 0 {
		return false, nil
	}
	// 存在则判断是否和入参一致
	lbl := lblResp.Details[0]

	if req.Name != lbl.Name {
		return true, fmt.Errorf("listener(%s) already exist, name mismatch, want: %s, db: %s",
			lbl.CloudID, req.Name, lbl.Name)
	}
	if req.BkBizID != lbl.BkBizID {
		return true, fmt.Errorf("listener(%s) already exist, biz id mismatch, want: %d, db: %d",
			lbl.CloudID, req.BkBizID, lbl.BkBizID)
	}
	if req.SniSwitch != lbl.SniSwitch {
		return true, fmt.Errorf("listener(%s) already exist, sni switch mismatch, want: %t, db: %t",
			lbl.CloudID, req.SniSwitch, lbl.SniSwitch)
	}
	if req.EndPort != nil {
		if lbl.Extension == nil {
			return true, fmt.Errorf("listener(%s) already exist, session expire mismatch, want: %d, db no ext",
				lbl.CloudID, req.SessionExpire)
		}
		if !assert.IsPtrInt64Equal(lbl.Extension.EndPort, req.EndPort) {
			return true, fmt.Errorf("listener(%s) already exist, session expire mismatch, want: %+v, db: %+v",
				lbl.CloudID, req.SessionExpire, lbl.Extension.EndPort)
		}
	}
	if req.Certificate != nil {
		if lbl.Extension == nil {
			return true, fmt.Errorf("listener(%s) already exist, cert mismatch, want: %+v, db no ext",
				lbl.CloudID, req.Certificate)
		}
		if isListenerCertChange(req.Certificate, lbl.Extension.Certificate) {
			return true, fmt.Errorf("listener(%s) already exist, cert mismatch, want: %+v, got: %+v", lbl.CloudID,
				req.Certificate, lbl.Extension.Certificate)
		}

	}

	if req.Protocol.IsLayer7Protocol() {
		return true, nil
	}
	// 对于四层需要继续查询规则

	return true, nil

}

// Rollback 支持重入，无需回滚
func (act BatchTaskTCloudCreateListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskTCloudCreateListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
