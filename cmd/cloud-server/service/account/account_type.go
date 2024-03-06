package account

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

func (a *accountSvc) GetTCloudNetworkAccountType(cts *rest.Contexts) (any, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "accountID is required")
	}

	// 校验用户有该账号的查看权限
	if err := a.checkPermission(cts, meta.Find, accountID); err != nil {
		return nil, err
	}
	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if baseInfo.Vendor != enumor.TCloud {
		return nil, errf.New(errf.InvalidParameter, "only TCloud account is support now")
	}
	return a.client.HCService().TCloud.Account.GetNetworkAccountType(cts.Kit, accountID)
}
