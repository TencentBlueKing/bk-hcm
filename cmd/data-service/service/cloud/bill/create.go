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
	"reflect"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dsbill "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/cloud/bill"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateAccountBillConfig batch create account bill config.
func (svc *billConfigSvc) BatchCreateAccountBillConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Aws:
		return batchCreateAccountBillConfig[cloud.AwsBillConfigExtension](cts, vendor, svc)
	case enumor.Gcp:
		return batchCreateAccountBillConfig[cloud.GcpBillConfigExtension](cts, vendor, svc)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// batchCreateAccountBillConfig create account bill config.
func batchCreateAccountBillConfig[T dsbill.AccountBillConfigExtension](cts *rest.Contexts,
	vendor enumor.Vendor, svc *billConfigSvc) (interface{}, error) {

	req := new(dsbill.AccountBillConfigBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		bills := make([]tablebill.AccountBillConfigTable, 0, len(req.Bills))
		for _, createReq := range req.Bills {
			ext, err := tabletype.NewJsonField(createReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			tmpErrMsg, err := convertArrToTableJSON(createReq.ErrMsg)
			if err != nil {
				return nil, err
			}

			tmpBill := tablebill.AccountBillConfigTable{
				Vendor:            vendor,
				AccountID:         createReq.AccountID,
				CloudDatabaseName: createReq.CloudDatabaseName,
				CloudTableName:    createReq.CloudTableName,
				Status:            createReq.Status,
				ErrMsg:            tmpErrMsg,
				Extension:         ext,
				Creator:           cts.Kit.User,
				Reviser:           cts.Kit.User,
			}

			bills = append(bills, tmpBill)
		}

		billID, err := svc.dao.AccountBillConfig().CreateWithTx(cts.Kit, txn, bills)
		if err != nil {
			return nil, fmt.Errorf("create account bill config failed, err: %+v", err)
		}

		return billID, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := billIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create account bill config but return id type is not string, "+
			"id type: %v", reflect.TypeOf(billIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
