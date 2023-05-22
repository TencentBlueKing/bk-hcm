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

package bill

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dsbill "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/cloud/bill"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// BatchUpdateAccountBillConfig batch update account bill config.
func (svc *billConfigSvc) BatchUpdateAccountBillConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Aws:
		return batchUpdateAccountBillConfig[cloud.AwsBillConfigExtension](cts, svc)
	case enumor.Gcp:
		return batchUpdateAccountBillConfig[cloud.GcpBillConfigExtension](cts, svc)
	}

	return nil, nil
}

// batchUpdateAccountBillConfig batch update account bill config.
func batchUpdateAccountBillConfig[T dsbill.AccountBillConfigExtension](cts *rest.Contexts,
	svc *billConfigSvc) (interface{}, error) {

	req := new(dsbill.AccountBillConfigBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Bills))
	for _, niItem := range req.Bills {
		ids = append(ids, niItem.ID)
	}

	// check if all account bill config exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.AccountBillConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch update list account bill config failed, err: %+v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account bill config failed, err: %+v", err)
	}

	if listRes.Count != uint64(len(req.Bills)) {
		return nil, fmt.Errorf("list account bill config failed, some bill(ids=%+v) doesn't exist", ids)
	}

	// update account bill config
	billInfo := &tablebill.AccountBillConfigTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Bills {
		billInfo.CloudDatabaseName = updateReq.CloudDatabaseName
		billInfo.CloudTableName = updateReq.CloudTableName
		billInfo.Status = updateReq.Status
		tmpErrMsg, err := convertArrToTableJSON(updateReq.ErrMsg)
		if err != nil {
			return nil, err
		}
		billInfo.ErrMsg = tmpErrMsg

		// update extension
		if updateReq.Extension != nil {
			// get account bill config before for expression
			dbBill, err := getAccountBillConfigFromTable(cts.Kit, svc.dao, updateReq.ID)
			if err != nil {
				return nil, err
			}

			updatedExtension, err := json.UpdateMerge(updateReq.Extension, string(dbBill.Extension))
			if err != nil {
				return nil, fmt.Errorf("extension update account bill config merge failed, err: %v", err)
			}

			billInfo.Extension = tabletype.JsonField(updatedExtension)
		}

		err = svc.dao.AccountBillConfig().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), billInfo)
		if err != nil {
			logs.Errorf("batch update account bill config failed, err: %+v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update account bill config failed, err: %+v", err)
		}
	}

	return nil, nil
}

func convertArrToTableJSON(arr []string) (tabletype.JsonField, error) {
	if len(arr) == 0 {
		return "[]", nil
	}

	str, err := json.Marshal(arr)
	if err != nil {
		logs.Errorf("convert arr to table json, json.Marshal failed, arr: %v, err: %v", arr, err)
		return "[]", err
	}

	return tabletype.JsonField(str), nil
}
