/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package account

import (
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// BasicInfo contains account fields used by biz sub-account ext response.
type BasicInfo struct {
	ID      string
	Name    string
	BkBizID int64
}

// BatchListBasicInfoByAccountIDs lists account basic info by account IDs.
func BatchListBasicInfoByAccountIDs(kt *kit.Kit, cli *dataservice.Client, accountIDs []string) (
	map[string]BasicInfo, error) {

	accountIDMap := make(map[string]struct{}, len(accountIDs))
	for _, accountID := range accountIDs {
		if accountID == "" {
			continue
		}
		accountIDMap[accountID] = struct{}{}
	}

	uniqueIDs := make([]string, 0, len(accountIDMap))
	for accountID := range accountIDMap {
		uniqueIDs = append(uniqueIDs, accountID)
	}

	result := make(map[string]BasicInfo, len(uniqueIDs))
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	for _, ids := range slice.Split(uniqueIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.AccountListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", ids), tools.RuleEqual("type", enumor.ResourceAccount)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "name", "bk_biz_id"},
		}

		accounts, err := cli.Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account basic info err: %v rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, accountItem := range accounts.Details {
			result[accountItem.ID] = BasicInfo{
				ID:      accountItem.ID,
				Name:    accountItem.Name,
				BkBizID: accountItem.BkBizID,
			}
		}
	}

	return result, nil
}

// BuildOperableMapByAccountMap builds account operable map with current biz ID.
func BuildOperableMapByAccountMap(bkBizID int64, accountIDs []string, accountMap map[string]BasicInfo) map[string]bool {
	result := make(map[string]bool, len(accountIDs))
	for _, accountID := range accountIDs {
		if accountID == "" {
			continue
		}

		accountInfo, exists := accountMap[accountID]
		if !exists {
			result[accountID] = false
			continue
		}

		result[accountID] = accountInfo.BkBizID == bkBizID
	}

	return result
}

// BuildAccountNameMapByAccountMap builds account name map by account ID.
func BuildAccountNameMapByAccountMap(accountIDs []string, accountMap map[string]BasicInfo) map[string]string {
	result := make(map[string]string, len(accountIDs))
	for _, accountID := range accountIDs {
		if accountID == "" {
			continue
		}

		accountInfo, exists := accountMap[accountID]
		if !exists {
			result[accountID] = ""
			continue
		}

		result[accountID] = accountInfo.Name
	}

	return result
}

// BatchBuildOperableAndNameMap returns account-info map and account-id keyed operable map.
func BatchBuildOperableAndNameMap(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, accountIDs []string) (
	map[string]BasicInfo, map[string]bool, error) {

	accountMap, err := BatchListBasicInfoByAccountIDs(kt, cli, accountIDs)
	if err != nil {
		return nil, nil, err
	}

	return accountMap, BuildOperableMapByAccountMap(bkBizID, accountIDs, accountMap), nil
}

// ListAccountIDsByBizID lists all resource type account IDs under specified biz.
func ListAccountIDsByBizID(kt *kit.Kit, cli *dataservice.Client, bkBizID int64) ([]string, error) {
	listReq := &protocloud.AccountListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("type", enumor.ResourceAccount),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id"},
	}

	resultMap := make(map[string]struct{})
	for {
		accounts, err := cli.Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return nil, err
		}

		for _, account := range accounts.Details {
			if account.ID == "" {
				continue
			}
			resultMap[account.ID] = struct{}{}
		}

		if len(accounts.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	result := make([]string, 0, len(resultMap))
	for accountID := range resultMap {
		result = append(result, accountID)
	}

	return result, nil
}
