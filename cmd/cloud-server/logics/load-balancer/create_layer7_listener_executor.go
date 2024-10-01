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

var _ ImportExecutor = (*CreateLayer7ListenerExecutor)(nil)

func newCreateLayer7ListenerExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *CreateLayer7ListenerExecutor {

	return &CreateLayer7ListenerExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer7ListenerExecutor excel导入——创建四层监听器执行器
type CreateLayer7ListenerExecutor struct {
	*basePreviewExecutor

	taskCli *taskserver.Client
	details []*CreateLayer7ListenerDetail
}

// Execute ...
func (c *CreateLayer7ListenerExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (string, error) {
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

func (c *CreateLayer7ListenerExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := unmarshalData(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *CreateLayer7ListenerExecutor) validate(kt *kit.Kit) error {
	validator := &CreateLayer7ListenerPreviewExecutor{
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
func (c *CreateLayer7ListenerExecutor) filter() {
	c.details = slice.Filter[*CreateLayer7ListenerDetail](c.details, func(detail *CreateLayer7ListenerDetail) bool {
		if detail.Status == Executable {
			return true
		}
		return false
	})
}

func (c *CreateLayer7ListenerExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// group by clb
	//TODO implement me

	panic("")
}

func (c *CreateLayer7ListenerExecutor) buildTask(kt *kit.Kit, strings []string, s string) (string, error) {
	//TODO implement me
	panic("implement me")
}
