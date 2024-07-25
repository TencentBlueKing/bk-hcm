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

package billitem

import (
	rawjson "encoding/json"
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// CreateBillItemRaw create bill item with options
func (svc *service) CreateBillItemRaw(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	return createBillItem[rawjson.RawMessage](cts, svc, vendor)
}

// CreateBillItem create bill item with options
func (svc *service) CreateBillItem(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Aws:
		return createBillItem[bill.AwsBillItemExtension](cts, svc, vendor)
	case enumor.HuaWei:
		return createBillItem[bill.HuaweiBillItemExtension](cts, svc, vendor)
	case enumor.Azure:
		return createBillItem[bill.AzureBillItemExtension](cts, svc, vendor)
	case enumor.Gcp:
		return createBillItem[bill.GcpBillItemExtension](cts, svc, vendor)
	case enumor.Kaopu:
		return createBillItem[bill.KaopuBillItemExtension](cts, svc, vendor)
	case enumor.Zenlayer:
		return createBillItem[bill.ZenlayerBillItemExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport vendor %s ", vendor)
	}
}

func createBillItem[E bill.BillItemExtension](cts *rest.Contexts, svc *service, vendor enumor.Vendor) (any, error) {
	var req dsbill.BatchBillItemCreateReq[E]
	if err := cts.DecodeInto(&req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	idList, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		var billItemTables []*tablebill.AccountBillItem
		for _, item := range req.Items {
			extJson, err := json.MarshalToString(item.Extension)
			if err != nil {
				logs.Errorf("fail marashal %s bill item into json, err: %v, rid: %v", vendor, err, cts.Kit.Rid)
				return nil, err
			}
			billItem := tablebill.AccountBillItem{
				RootAccountID: item.RootAccountID,
				MainAccountID: item.MainAccountID,
				Vendor:        vendor,
				ProductID:     item.ProductID,
				BkBizID:       item.BkBizID,
				BillYear:      item.BillYear,
				BillMonth:     item.BillMonth,
				BillDay:       item.BillDay,
				VersionID:     item.VersionID,
				Currency:      item.Currency,
				Cost:          &types.Decimal{Decimal: item.Cost},
				HcProductCode: item.HcProductCode,
				HcProductName: item.HcProductName,
				ResAmount:     &types.Decimal{Decimal: item.ResAmount},
				ResAmountUnit: item.ResAmountUnit,
				Extension:     types.JsonField(extJson),
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			}
			billItemTables = append(billItemTables, &billItem)
		}

		ids, err := svc.dao.AccountBillItem().CreateWithTx(cts.Kit, txn, req.ItemCommonOpt, billItemTables)
		if err != nil {
			logs.Errorf("fail to create %s bill item, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("create account bill item list failed, err: %v", err)
		}
		return ids, nil
	})

	if err != nil {
		return nil, err
	}
	retList, ok := idList.([]string)
	if !ok {
		return nil, fmt.Errorf("create account bill item but return ids type not []string, ids type: %v",
			reflect.TypeOf(idList).String())
	}

	return &core.BatchCreateResult{IDs: retList}, nil
}
