package lblogic

import (
	"encoding/json"
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var _ ImportExecutor = (*Layer4ListenerBindRSExecutor)(nil)

func newLayer4ListenerBindRSExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *Layer4ListenerBindRSExecutor {

	return &Layer4ListenerBindRSExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// Layer4ListenerBindRSExecutor excel导入——创建四层监听器执行器
type Layer4ListenerBindRSExecutor struct {
	*basePreviewExecutor

	validator Layer4ListenerBindRSPreviewExecutor

	taskCli     *taskserver.Client
	details     []*Layer4ListenerBindRSDetail
	taskDetails []*layer4ListenerBindRSTaskDetail
}

type layer4ListenerBindRSTaskDetail struct {
	flowID   string
	actionID string
	detail   *Layer4ListenerBindRSDetail
}

// Execute ...
func (c *Layer4ListenerBindRSExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (string, error) {
	err := c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		return "", err
	}
	c.filter()

	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		logs.Errorf("build flows failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	taskID, err := c.buildTask(kt, flowIDs)
	if err != nil {
		logs.Errorf("build taskManagement task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return taskID, nil
}

func (c *Layer4ListenerBindRSExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := unmarshalData(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *Layer4ListenerBindRSExecutor) validate(kt *kit.Kit) error {
	validator := &Layer4ListenerBindRSPreviewExecutor{
		basePreviewExecutor: c.basePreviewExecutor,
		details:             c.details,
	}
	err := validator.validate(kt)
	if err != nil {
		return err
	}

	for _, detail := range c.details {
		if detail.Status == NotExecutable {
			return fmt.Errorf("record(%v) is not executable", detail)
		}
	}
	return nil
}

func (c *Layer4ListenerBindRSExecutor) filter() {
	c.details = slice.Filter[*Layer4ListenerBindRSDetail](c.details, func(detail *Layer4ListenerBindRSDetail) bool {
		if detail.Status == Executable {
			return true
		}
		return false
	})
}

func (c *Layer4ListenerBindRSExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// group by clb
	clbToDetails := make(map[string][]*Layer4ListenerBindRSDetail)
	for _, detail := range c.details {
		clbToDetails[detail.CloudClbID] = append(clbToDetails[detail.CloudClbID], detail)
	}
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.accountID, c.bkBizID, converter.MapKeyToSlice(clbToDetails))
	if err != nil {
		return nil, err
	}

	flowIDs := make([]string, 0, len(clbToDetails))
	for clbCloudID, details := range clbToDetails {
		lb := lbMap[clbCloudID]
		flowID, err := c.buildFlow(kt, lb, details)
		if err != nil {
			logs.Errorf("build flow for clb(%s) failed, err: %v, rid: %s", clbCloudID, err, kt.Rid)
			return nil, err
		}
		flowIDs = append(flowIDs, flowID)
	}

	return flowIDs, nil
}

func (c *Layer4ListenerBindRSExecutor) buildFlow(kt *kit.Kit, lb corelb.BaseLoadBalancer, details []*Layer4ListenerBindRSDetail) (string, error) {

	// 将details根据targetGroupID进行分组，以targetGroupID的纬度创建flowTask
	tgToDetails, err := c.createTaskDetailsGroupByTargetGroup(kt, lb.ID, details)
	if err != nil {
		logs.Errorf("create task details group by target group failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	actionIDGenerator := newActionIDGenerator(1, 10)
	flowTasks := make([]ts.CustomFlowTask, 0, len(tgToDetails))
	for targetGroupID, detailList := range tgToDetails {
		tmpTask, err := c.buildFlowTask(kt, lb, targetGroupID, detailList, actionIDGenerator)
		if err != nil {
			return "", err
		}
		flowTasks = append(flowTasks, tmpTask...)
	}

	_, err = checkResFlowRel(kt, c.dataServiceCli, lb.ID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	flowID, err := c.createFlowTask(kt, lb.ID, converter.MapKeyToSlice(tgToDetails), flowTasks)
	if err != nil {
		return "", err
	}
	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lb.ID, enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	for _, taskDetails := range tgToDetails {
		for _, detail := range taskDetails {
			detail.flowID = flowID
			c.taskDetails = append(c.taskDetails, detail)
		}
	}
	return flowID, nil
}

func (c *Layer4ListenerBindRSExecutor) createTaskDetailsGroupByTargetGroup(kt *kit.Kit, lbCloudID string,
	details []*Layer4ListenerBindRSDetail) (map[string][]*layer4ListenerBindRSTaskDetail, error) {

	tgToDetails := make(map[string][]*layer4ListenerBindRSTaskDetail)
	for _, detail := range details {
		listener, err := getListener(kt, c.dataServiceCli, c.accountID, lbCloudID, detail.ListenerPort[0], c.bkBizID)
		if err != nil {
			return nil, err
		}
		if listener == nil {
			return nil, fmt.Errorf("loadbalancer(%s) listener(%v) not found", detail.CloudClbID, detail.ListenerPort)
		}

		targetGroupID, err := getTargetGroupID(kt, c.dataServiceCli, listener.CloudID)
		if err != nil {
			return nil, err
		}
		tgToDetails[targetGroupID] = append(tgToDetails[targetGroupID], &layer4ListenerBindRSTaskDetail{
			detail: detail,
		})
	}
	return tgToDetails, nil
}

func (c *Layer4ListenerBindRSExecutor) createFlowTask(kt *kit.Kit, lbID string, tgIDs []string, flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowTargetGroupAddRS,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.taskCli.CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch add rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flowID := result.ID
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  tgIDs,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.AddRSTaskType,
			},
		}},
	}
	_, err = c.taskCli.CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (c *Layer4ListenerBindRSExecutor) buildFlowTask(kt *kit.Kit, lb corelb.BaseLoadBalancer, targetGroupID string,
	details []*layer4ListenerBindRSTaskDetail, generator func() (cur string, prev string)) ([]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchAddRSCloudMaxLimit) {
		cur, prev := generator()

		rsList := make([]*dataproto.TargetBaseReq, 0, len(taskDetails))
		for _, detail := range taskDetails {
			// TODO cmd/cloud-server/service/load-balancer/async_target_group_add_rs.go:320
			// 如果传入ENI的信息，需要怎么处理，已有的实现没有看到ENI相关的逻辑
			cvm, err := getCvm(kt, c.dataServiceCli, detail.detail.RsIp, c.vendor, c.bkBizID, c.accountID)
			if err != nil {
				return nil, err
			}
			if cvm == nil {
				return nil, fmt.Errorf("rs(%s) not found", detail.detail.RsIp)
			}

			rs := &dataproto.TargetBaseReq{
				IP:            detail.detail.RsIp,
				InstType:      detail.detail.InstType,
				Port:          int64(detail.detail.RsPort[0]),
				Weight:        converter.ValToPtr(int64(detail.detail.Weight)),
				AccountID:     c.accountID,
				TargetGroupID: targetGroupID,
				CloudInstID:   cvm.CloudID,
				InstName:      cvm.Name,
				// TODO 需要确认是否加上ipv6的数据 ref: cmd/cloud-server/service/load-balancer/async_target_group_add_rs.go:331
				PrivateIPAddress: cvm.PrivateIPv4Addresses,
				PublicIPAddress:  cvm.PublicIPv4Addresses,
				CloudVpcIDs:      cvm.CloudVpcIDs,
				Zone:             cvm.Zone,
			}
			rsList = append(rsList, rs)
		}

		req := hclb.TCloudBatchOperateTargetReq{
			TargetGroupID: targetGroupID,
			LbID:          lb.ID,
			RsList:        rsList,
		}
		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionTargetGroupAddRS,
			Params: &actionlb.OperateRsOption{
				Vendor:                      c.vendor,
				TCloudBatchOperateTargetReq: req,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)
		// update taskDetail.actionID
		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}

func (c *Layer4ListenerBindRSExecutor) buildTask(kt *kit.Kit, strings []string) (string, error) {
	//TODO implement me
	panic("implement me")
}
