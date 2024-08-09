package billadjustment

import (
	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"

	"github.com/shopspring/decimal"
)

// SumBillAdjustmentItem Summarize the adjustment items bill
func (b *billAdjustmentSvc) SumBillAdjustmentItem(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.AdjustmentItemSumReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(cts.Kit,
		&dsbillapi.BillAdjustmentItemListReq{
			Filter: req.Filter,
			Page: &core.BasePage{
				Count: true,
			},
		})
	if err != nil {
		return nil, err
	}
	var adjustmentItemList []*billcore.AdjustmentItem
	for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		tmpResult, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(cts.Kit,
			&dsbillapi.BillAdjustmentItemListReq{
				Filter: req.Filter,
				Page: &core.BasePage{
					Start: uint32(offset),
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, err
		}
		adjustmentItemList = append(adjustmentItemList, tmpResult.Details...)
	}
	return b.doCalculate(adjustmentItemList, result.Count)
}

func (b *billAdjustmentSvc) doCalculate(adjustmentItems []*billcore.AdjustmentItem, count uint64) (interface{}, error) {
	retMap := make(map[enumor.BillAdjustmentType]map[enumor.CurrencyCode]*billcore.CostWithCurrency)
	retMap[enumor.BillAdjustmentIncrease] = make(map[enumor.CurrencyCode]*billcore.CostWithCurrency)
	retMap[enumor.BillAdjustmentDecrease] = make(map[enumor.CurrencyCode]*billcore.CostWithCurrency)

	for _, item := range adjustmentItems {
		tmpMap := retMap[item.Type]
		currencyCode := enumor.CurrencyCode(item.Currency)
		if _, ok := tmpMap[currencyCode]; !ok {
			tmpMap[currencyCode] = &billcore.CostWithCurrency{
				Cost:     decimal.NewFromFloat(0),
				RMBCost:  decimal.NewFromFloat(0),
				Currency: currencyCode,
			}
		}
		tmpMap[currencyCode].Cost = tmpMap[currencyCode].Cost.Add(item.Cost)
		tmpMap[currencyCode].RMBCost = tmpMap[currencyCode].RMBCost.Add(item.RMBCost)
	}
	return &asbillapi.AdjustmentItemSumResult{
		Count:   count,
		CostMap: retMap,
	}, nil
}
