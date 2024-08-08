package actionlb

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
)

// --------------------------[创建Listener&Rule]-----------------------------

// CreateListenerOption define operate rs option.
type CreateListenerOption struct {
	Vendor                         enumor.Vendor `json:"vendor" validate:"required"`
	hclb.ListenerWithRuleCreateReq `json:",inline"`
}

var _ action.Action = new(CreateListenerAction)
var _ action.ParameterAction = new(CreateListenerAction)

// CreateListenerAction define modify target weight action.
type CreateListenerAction struct{}

// ParameterNew return request params.
func (act CreateListenerAction) ParameterNew() (params interface{}) {
	return new(CreateListenerOption)
}

// Name return action name
func (act CreateListenerAction) Name() enumor.ActionName {
	return enumor.ActionListenerCreate
}

// Run modify target port.
func (act CreateListenerAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*CreateListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	var result *hclb.ListenerWithRuleCreateResult
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Clb.CreateListener(
			kt.Kit(), &opt.ListenerWithRuleCreateReq)
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("[%s] create listener failed, err: %v, rid: %s", opt.Vendor, err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量修改RS权重失败时的回滚Action，此处不需要回滚处理
func (act CreateListenerAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- CreateListenerAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}

// --------------------------[创建URLRule]-----------------------------

// CreateURLRuleOption define operate rs option.
type CreateURLRuleOption struct {
	Vendor                        enumor.Vendor `json:"vendor" validate:"required"`
	ListenerID                    string        `json:"listener_id" validate:"required"`
	hclb.TCloudRuleBatchCreateReq `json:",inline"`
}

var _ action.Action = new(CreateURLRuleAction)
var _ action.ParameterAction = new(CreateURLRuleAction)

// CreateURLRuleAction define modify target weight action.
type CreateURLRuleAction struct{}

// ParameterNew return request params.
func (act CreateURLRuleAction) ParameterNew() (params interface{}) {
	return new(CreateURLRuleOption)
}

// Name return action name
func (act CreateURLRuleAction) Name() enumor.ActionName {
	return enumor.ActionURLRuleCreate
}

// Run modify target port.
func (act CreateURLRuleAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*CreateURLRuleOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	var result *hclb.BatchCreateResult
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Clb.BatchCreateUrlRule(
			kt.Kit(), opt.ListenerID, &opt.TCloudRuleBatchCreateReq)
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("[%s] create url rule failed, err: %v, rid: %s", opt.Vendor, err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量修改RS权重失败时的回滚Action，此处不需要回滚处理
func (act CreateURLRuleAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- CreateURLRuleAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}
