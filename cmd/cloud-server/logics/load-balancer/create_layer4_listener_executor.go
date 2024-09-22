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

var _ ImportExecutor = (*CreateLayer4ListenerExecutor)(nil)

func newCreateLayer4ListenerExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *CreateLayer4ListenerExecutor {

	return &CreateLayer4ListenerExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer4ListenerExecutor excel导入——创建四层监听器执行器
type CreateLayer4ListenerExecutor struct {
	*basePreviewExecutor

	taskCli *taskserver.Client
	details []*CreateLayer4ListenerDetail
}

// Execute ...
func (c *CreateLayer4ListenerExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (string, error) {
	err := c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		return "", err
	}
	c.filter()

	//flow, err := c.buildFlows(kt)

	return "", nil
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
	//TODO implement me

	panic("")
}

func (c *CreateLayer4ListenerExecutor) buildTask(kt *kit.Kit, strings []string) (string, error) {
	//TODO implement me
	panic("implement me")
}
