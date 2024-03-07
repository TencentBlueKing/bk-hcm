package account

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

func (svc *service) GetTCloudNetworkAccountType(cts *rest.Contexts) (any, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "accountID is required")
	}

	client, err := svc.ad.TCloud(cts.Kit, accountID)
	if err != nil {
		return nil, err
	}

	return client.DescribeNetworkAccountType(cts.Kit)
}
