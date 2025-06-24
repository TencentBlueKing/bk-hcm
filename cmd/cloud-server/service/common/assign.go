package common

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// CheckAssignableToBiz 检查指定资源是否可以分配到目标业务。如果资源所属账号的使用业务范围不包含目标业务，则返回错误。
func CheckAssignableToBiz(cli *dataservice.Client, cts *rest.Contexts, resType enumor.CloudResourceType,
	resIDs []string, bkBizID int64) error {
	req := cloud.ListResourceBasicInfoReq{
		ResourceType: resType,
		IDs:          resIDs,
		Fields:       []string{"id", "account_id"},
	}
	resMap, err := cli.Global.Cloud.ListResBasicInfo(cts.Kit, req)
	if err != nil {
		return err
	}

	accountIDs := make([]string, len(resMap))
	for _, info := range resMap {
		accountIDs = append(accountIDs, info.AccountID)
	}

	accountReq := &protocloud.AccountListReq{
		Filter: tools.ContainersExpression("id", accountIDs),
		Page:   core.NewDefaultBasePage(),
	}
	accountResp, err := cli.Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), accountReq)
	if err != nil {
		return err
	}

	accountMap := make(map[string]*corecloud.BaseAccount, len(accountResp.Details))
	for _, account := range accountResp.Details {
		accountMap[account.ID] = account
	}

	for _, resID := range resIDs {
		// 拿到当前resID对应的accountID
		accountID := resMap[resID].AccountID
		// 拿到当前accountID的所有使用业务
		usageBizIDs := accountMap[accountID].UsageBizIDs

		if slice.IsItemInSlice(usageBizIDs, bkBizID) ||
			(len(usageBizIDs) == 1 && usageBizIDs[0] == constant.AttachedAllBiz) {
			continue
		}

		// 要分配到的业务不在账号的使用业务范围内则报错
		resTableName, err := resType.ConvTableName()
		if err != nil {
			return err
		}
		return fmt.Errorf("biz %d to be assigned for %s %s is not in account's usageBizIDs", bkBizID,
			resTableName, resID)
	}
	return nil
}
