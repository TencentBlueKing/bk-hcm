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

var _ ImportExecutor = (*Layer7ListenerBindRSExecutor)(nil)

func newLayer7ListenerBindRSExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor,
	bkBizID int64, accountID string, regionIDs []string) *Layer7ListenerBindRSExecutor {

	return &Layer7ListenerBindRSExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// Layer7ListenerBindRSExecutor excel导入——创建四层监听器执行器
type Layer7ListenerBindRSExecutor struct {
	*basePreviewExecutor

	taskCli *taskserver.Client
	details []*Layer7ListenerBindRSDetail
}

// Execute ...
func (c *Layer7ListenerBindRSExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (string, error) {
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

func (c *Layer7ListenerBindRSExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := unmarshalData(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *Layer7ListenerBindRSExecutor) validate(kt *kit.Kit) error {
	validator := &Layer7ListenerBindRSPreviewExecutor{
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
func (c *Layer7ListenerBindRSExecutor) filter() {
	c.details = slice.Filter[*Layer7ListenerBindRSDetail](c.details, func(detail *Layer7ListenerBindRSDetail) bool {
		if detail.Status == Executable {
			return true
		}
		return false
	})
}

func (c *Layer7ListenerBindRSExecutor) buildFlows(kt *kit.Kit) (string, error) {
	// group by clb
	//TODO implement me

	panic("")
}

func (c *Layer7ListenerBindRSExecutor) buildTask(kt *kit.Kit) (string, error) {
	//TODO implement me
	panic("implement me")
}
