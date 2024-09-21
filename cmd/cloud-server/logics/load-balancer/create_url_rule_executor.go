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

var _ ImportExecutor = (*CreateUrlRuleExecutor)(nil)

func newCreateUrlRuleExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *CreateUrlRuleExecutor {

	return &CreateUrlRuleExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateUrlRuleExecutor excel导入——创建四层监听器执行器
type CreateUrlRuleExecutor struct {
	*basePreviewExecutor

	taskCli *taskserver.Client
	details []*CreateUrlRuleDetail
}

// Execute ...
func (c *CreateUrlRuleExecutor) Execute(kt *kit.Kit, source string, rawDetails json.RawMessage) (string, error) {
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

func (c *CreateUrlRuleExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := unmarshalData(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *CreateUrlRuleExecutor) validate(kt *kit.Kit) error {
	validator := &CreateUrlRulePreviewExecutor{
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
func (c *CreateUrlRuleExecutor) filter() {
	c.details = slice.Filter[*CreateUrlRuleDetail](c.details, func(detail *CreateUrlRuleDetail) bool {
		if detail.Status == Executable {
			return true
		}
		return false
	})
}

func (c *CreateUrlRuleExecutor) buildFlows(kt *kit.Kit) (string, error) {
	// group by clb
	//TODO implement me

	panic("")
}

func (c *CreateUrlRuleExecutor) buildTask(kt *kit.Kit) (string, error) {
	//TODO implement me
	panic("implement me")
}
