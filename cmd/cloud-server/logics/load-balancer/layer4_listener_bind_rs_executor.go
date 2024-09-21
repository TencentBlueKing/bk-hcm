package lblogic

import (
	"encoding/json"
	"fmt"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
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

	taskCli *taskserver.Client
	details []*Layer4ListenerBindRSDetail
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

	// TODO
	//flow, err := c.buildFlows(kt)

	return "", nil
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

func (c *Layer4ListenerBindRSExecutor) buildFlows(kt *kit.Kit) (string, error) {
	// group by clb
	//TODO implement me

	panic("")
}

func (c *Layer4ListenerBindRSExecutor) buildTask(kt *kit.Kit) (string, error) {
	//TODO implement me
	panic("implement me")
}
