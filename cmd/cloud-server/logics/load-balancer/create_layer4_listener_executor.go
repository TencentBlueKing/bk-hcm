package lblogic

import (
	"encoding/json"
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/api/data-service/task"
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

var _ ImportExecutor = (*CreateLayer4ListenerExecutor)(nil)

func newCreateLayer4ListenerExecutor(cli *dataservice.Client, taskCli *taskserver.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *CreateLayer4ListenerExecutor {

	return &CreateLayer4ListenerExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer4ListenerExecutor excel导入——创建四层监听器执行器
type CreateLayer4ListenerExecutor struct {
	*basePreviewExecutor

	taskCli     *taskserver.Client
	details     []*CreateLayer4ListenerDetail
	taskDetails []*createLayer4ListenerTaskDetail
}

// 用于记录 detail - 异步任务flow&task - 任务管理 之间的关系
type createLayer4ListenerTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*CreateLayer4ListenerDetail
}

// Execute ...
func (c *CreateLayer4ListenerExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (
	string, error) {

	err := c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		return "", err
	}
	c.filter()

	taskID, err := c.buildTaskManagementAndDetails(kt, source)
	if err != nil {
		return "", err
	}
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		return "", err
	}
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID)
	if err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return taskID, nil
}

func (c *CreateLayer4ListenerExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := unmarshalData(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *CreateLayer4ListenerExecutor) validate(kt *kit.Kit) error {
	validator := &CreateLayer4ListenerPreviewExecutor{
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
func (c *CreateLayer4ListenerExecutor) filter() {
	c.details = slice.Filter[*CreateLayer4ListenerDetail](c.details, func(detail *CreateLayer4ListenerDetail) bool {
		if detail.Status == Executable {
			return true
		}
		return false
	})
}

func (c *CreateLayer4ListenerExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// group by clb
	clbToDetails := make(map[string][]*createLayer4ListenerTaskDetail)
	for _, detail := range c.taskDetails {
		clbToDetails[detail.CloudClbID] = append(clbToDetails[detail.CloudClbID], detail)
	}
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli,
		c.accountID, c.bkBizID, converter.MapKeyToSlice(clbToDetails))
	if err != nil {
		return nil, err
	}

	flowIDs := make([]string, 0, len(clbToDetails))
	for clbCloudID, details := range clbToDetails {
		lb := lbMap[clbCloudID]
		flowID, err := c.buildFlow(kt, lb.ID, details)
		if err != nil {
			logs.Errorf("build flow for clb(%s) failed, err: %v, rid: %s", clbCloudID, err, kt.Rid)
			return nil, err
		}
		flowIDs = append(flowIDs, flowID)
	}

	return flowIDs, nil
}

func (c *CreateLayer4ListenerExecutor) buildFlow(kt *kit.Kit, lbID string,
	details []*createLayer4ListenerTaskDetail) (string, error) {

	flowTasks, err := c.buildFlowTask(lbID, details)
	if err != nil {
		logs.Errorf("build flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	_, err = checkResFlowRel(kt, c.dataServiceCli, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	flowID, err := c.createFlowTask(kt, lbID, flowTasks)
	if err != nil {
		return "", err
	}
	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lbID,
		enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}
	return flowID, nil
}

func (c *CreateLayer4ListenerExecutor) createFlowTask(kt *kit.Kit, lbID string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowLoadBalancerCreateListener,
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
				FlowID:   flowID,
				ResID:    lbID,
				ResType:  enumor.LoadBalancerCloudResType,
				TaskType: enumor.CreateListenerTaskType,
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

func (c *CreateLayer4ListenerExecutor) buildFlowTask(lbID string,
	details []*createLayer4ListenerTaskDetail) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(lbID, details)
	default:
		return nil, fmt.Errorf("vendor %s not supported", c.vendor)
	}
}

func (c *CreateLayer4ListenerExecutor) buildTCloudFlowTask(lbID string,
	details []*createLayer4ListenerTaskDetail) ([]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	actionIDGenerator := newActionIDGenerator(1, 10)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := actionIDGenerator()

		managementDetailIDs := make([]string, 0, len(taskDetails))
		listeners := make([]*hclb.TCloudListenerCreateReq, 0, len(taskDetails))
		for _, detail := range taskDetails {
			req := &hclb.TCloudListenerCreateReq{
				Name:          fmt.Sprintf("%s-%d", detail.Protocol, detail.ListenerPorts[0]),
				BkBizID:       c.bkBizID,
				LbID:          lbID,
				Protocol:      detail.Protocol,
				Port:          int64(detail.ListenerPorts[0]),
				Scheduler:     string(detail.Scheduler),
				SessionExpire: int64(detail.Session),
			}

			if len(detail.ListenerPorts) > 1 {
				req.EndPort = converter.ValToPtr(int64(detail.ListenerPorts[1]))
			}
			listeners = append(listeners, req)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudCreateListener,
			Params: &actionlb.BatchTaskTCloudCreateListenerOption{
				ManagementDetailIDs: managementDetailIDs,
				Listeners:           listeners,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)

		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}
	return result, nil
}

func (c *CreateLayer4ListenerExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskCreateLayer4Listener,
			Param:            detail,
		})
	}

	result, err := c.dataServiceCli.Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		return err
	}
	if len(result.IDs) != len(c.details) {
		return fmt.Errorf("create task details failed, expect created %d task details, but got %d",
			len(c.details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &createLayer4ListenerTaskDetail{
			taskDetailID:               result.IDs[i],
			CreateLayer4ListenerDetail: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}
	return nil
}

func (c *CreateLayer4ListenerExecutor) buildTaskManagementAndDetails(kt *kit.Kit, source string) (string, error) {
	taskID, err := createTaskManagement(kt, c.dataServiceCli, c.bkBizID, c.vendor, c.accountID,
		source, enumor.TaskCreateLayer4Listener)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	err = c.createTaskDetails(kt, taskID)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return taskID, nil
}

func (c *CreateLayer4ListenerExecutor) updateTaskManagementAndDetails(kt *kit.Kit,
	flowIDs []string, taskID string) error {

	if err := updateTaskManagement(kt, c.dataServiceCli, taskID, flowIDs); err != nil {
		logs.Errorf("update task management failed, taskID(%s), err: %v, rid: %s", taskID, err, kt.Rid)
		return err
	}
	if err := c.updateTaskDetails(kt); err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (c *CreateLayer4ListenerExecutor) updateTaskDetails(kt *kit.Kit) error {
	updateItems := make([]task.UpdateTaskDetailField, 0, len(c.taskDetails))
	for _, detail := range c.taskDetails {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:            detail.taskDetailID,
			FlowID:        detail.flowID,
			TaskActionIDs: []string{detail.actionID},
			State:         enumor.TaskDetailInit,
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.dataServiceCli.Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
