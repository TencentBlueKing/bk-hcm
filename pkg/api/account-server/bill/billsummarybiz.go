package bill

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
)

// BizSummaryListReq list request for Biz summary
type BizSummaryListReq struct {
	BillYear  int            `json:"bill_year" validate:"required"`
	BillMonth int            `json:"bill_month" validate:"required"`
	BKBizIDs  []int64        `json:"bk_biz_ids" validate:"required"`
	Page      *core.BasePage `json:"page" validate:"omitempty"`
}

// Validate ...
func (req *BizSummaryListReq) Validate() error {
	return validator.Validate.Struct(req)
}
