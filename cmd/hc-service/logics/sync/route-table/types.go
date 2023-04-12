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

// Package routetable defines route table service.
package routetable

import (
	"hcm/pkg/api/core"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

func getVpcBkBizIDFromDB(kt *kit.Kit, dataCli *dataclient.Client, accountID string,
	cloudVpcID string) (int64, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: accountID,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.Equal.Factory(),
				Value: cloudVpcID,
			},
		},
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return constant.UnassignedBiz, err
	}

	if len(dbList.Details) <= 0 {
		return constant.UnassignedBiz, nil
	}

	return dbList.Details[0].BkBizID, nil
}
