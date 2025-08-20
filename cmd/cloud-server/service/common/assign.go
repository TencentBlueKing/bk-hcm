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

package common

import (
	"fmt"
	
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// ValidateTargetBizID 检查指定资源是否可以分配到目标业务。如果资源所属账号的使用业务范围不包含目标业务，则返回错误。
func ValidateTargetBizID(kt *kit.Kit, cli *dataservice.Client, resType enumor.CloudResourceType,
	resIDs []string, bkBizID int64) error {

	// 资源ID映射到账号ID
	resMap, accountIDs, err := getAccountInfoForRes(kt, cli, resType, resIDs)
	if err != nil {
		logs.Errorf("get account info for resource failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	accountMap, err := getAccountUsageBizIDs(kt, cli, accountIDs)
	if err != nil {
		logs.Errorf("get account usage biz ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, resID := range resIDs {
		// 检查映射是否存在
		accountID, exists := resMap[resID]
		if !exists {
			return fmt.Errorf("resource %s not found, rid: %s", resID, kt.Rid)
		}

		// 检查账号是否存在
		usageBizIDs, exists := accountMap[accountID]
		if !exists {
			return fmt.Errorf("account %s not found for resource %s, rid: %s", accountID, resID, kt.Rid)
		}

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

// getAccountInfoForRes 根据资源ID得到资源ID与账号ID的映射map并拿到这些资源涉及到的所有账号ID去重后的切片
func getAccountInfoForRes(kt *kit.Kit, cli *dataservice.Client, resType enumor.CloudResourceType, resIDs []string) (
	map[string]string, []string, error) {
	// 资源ID映射到账号ID
	resMap := make(map[string]string)
	accountIDs := make([]string, 0)
	for _, parts := range slice.Split(resIDs, constant.BatchOperationMaxLimit) {
		req := cloud.ListResourceBasicInfoReq{
			ResourceType: resType,
			IDs:          parts,
			Fields:       []string{"id", "account_id"},
		}
		tempResMap, err := cli.Global.Cloud.ListResBasicInfo(kt, req)
		if err != nil {
			logs.Errorf("list account info failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		for _, resInfo := range tempResMap {
			resMap[resInfo.ID] = resInfo.AccountID
			accountIDs = append(accountIDs, resInfo.AccountID)
		}
	}
	accountIDs = slice.Unique(accountIDs)
	return resMap, accountIDs, nil
}

func getAccountUsageBizIDs(kt *kit.Kit, cli *dataservice.Client, accountIDs []string) (map[string][]int64, error) {
	if len(accountIDs) == 0 {
		return nil, nil
	}

	accountMap := make(map[string][]int64, len(accountIDs))
	for _, parts := range slice.Split(accountIDs, constant.BatchOperationMaxLimit) {
		accountReq := &protocloud.AccountListReq{
			Filter: tools.ContainersExpression("id", parts),
			Page:   core.NewDefaultBasePage(),
		}
		accountResp, err := cli.Global.Account.List(kt.Ctx, kt.Header(), accountReq)
		if err != nil {
			logs.Errorf("list account info failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, account := range accountResp.Details {
			accountMap[account.ID] = account.UsageBizIDs
		}
	}

	return accountMap, nil
}
