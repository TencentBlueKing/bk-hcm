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

package cvm

import (
	"fmt"
	"time"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// SyncCvmToCmdb sync cvm to cmdb.
func SyncCvmToCmdb(kt *kit.Kit, accountID string, bkBizID int64) error {
	listAccountOpt := &types.ListOption{
		Fields: []string{"vendor"},
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.dao.Account().List(kt, listAccountOpt)
	if err != nil {
		logs.Errorf("sync cvm to cmdb query account failed, err: %v, account: %d, rid: %s", err, accountID, kt.Rid)
		return fmt.Errorf("query account: %s failed, err: %v", accountID, listAccountOpt)
	}

	if len(list.Details) != 1 {
		return fmt.Errorf("account: %s not found", accountID)
	}
	vendor := enumor.Vendor(list.Details[0].Vendor)

	listCvmOpt := &types.ListOption{
		Filter: tools.EqualExpression("account_id", accountID),
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	totalCount := 0
	start := time.Now()
	for {
		result, err := svc.dao.Cvm().List(kt, listCvmOpt)
		if err != nil {
			logs.Errorf("sync cvm to list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return fmt.Errorf("list cvm failed, err: %v", err)
		}

		if len(result.Details) == 0 {
			break
		}

		for index := range result.Details {
			result.Details[index].BkBizID = bkBizID
		}

		switch vendor {
		case enumor.TCloud:
			err = upsertCmdbHosts[corecvm.TCloudCvmExtension](svc, kt, enumor.TCloud,
				converter.SliceToPtr(result.Details))
		case enumor.Aws:
			err = upsertCmdbHosts[corecvm.AwsCvmExtension](svc, kt, enumor.Aws,
				converter.SliceToPtr(result.Details))
		case enumor.HuaWei:
			err = upsertCmdbHosts[corecvm.HuaWeiCvmExtension](svc, kt, enumor.HuaWei,
				converter.SliceToPtr(result.Details))
		case enumor.Gcp:
			err = upsertCmdbHosts[corecvm.GcpCvmExtension](svc, kt, enumor.Gcp,
				converter.SliceToPtr(result.Details))
		case enumor.Azure:
			err = upsertCmdbHosts[corecvm.AzureCvmExtension](svc, kt, enumor.Azure,
				converter.SliceToPtr(result.Details))
		}
		if err != nil {
			logs.Errorf("upsertCmdbHosts failed, err: %v, rid; %s", err, kt.Rid)
			return err
		}
		totalCount += len(result.Details)

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	logs.Infof("sync cmdb to cmdb success, account: %s, bkBizID: %d, cvmCount: %d, cost: %v", accountID,
		bkBizID, totalCount, time.Since(start))

	return nil
}
