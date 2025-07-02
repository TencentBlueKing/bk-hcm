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
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetAccountInfo ...
func GetAccountInfo(kt *kit.Kit, cli *dataservice.Client, accountID string) (*cloud.BaseAccount, error) {
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account id is required")
	}

	listReq := &protocloud.AccountListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.Account.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		logs.Errorf("account: %s not found, rid: %s", accountID, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "account: %s not found", accountID)
	}

	return result.Details[0], nil
}
