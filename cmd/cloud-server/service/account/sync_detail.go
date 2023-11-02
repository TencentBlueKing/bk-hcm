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
	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// GetSyncDetail get account sync detail
func (a *accountSvc) GetSyncDetail(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

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

	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: accountID,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: baseInfo.Vendor,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	accountSyncDetail, err := a.client.DataService().Global.AccountSyncDetail.List(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}

	iassRes := make([]account.IassResItem, 0, len(accountSyncDetail.Details))
	for _, one := range accountSyncDetail.Details {
		item := account.IassResItem{
			ResName:         one.ResName,
			ResStatus:       one.ResStatus,
			ResEndTime:      one.ResEndTime,
			ResFailedReason: string(one.ResFailedReason),
		}
		iassRes = append(iassRes, item)
	}

	return account.SyncDetailRsp{
		IassRes: iassRes,
	}, nil
}
