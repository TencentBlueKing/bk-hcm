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

package clb

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tableclb "hcm/pkg/dal/table/cloud/clb"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateCLB 批量创建clb
func (svc *clbSvc) BatchCreateCLB(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateClb[coreclb.TCloudClbExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}

}
func batchCreateClb[T coreclb.Extension](cts *rest.Contexts, svc *clbSvc, vendor enumor.Vendor) (any, error) {
	req := new(dataproto.ClbBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tableclb.LoadBalancerTable, 0, len(req.Clbs))
		for _, lb := range req.Clbs {
			lbTable, err := convClbReqToTable(cts.Kit, vendor, lb)
			if err != nil {
				return nil, err
			}
			models = append(models, lbTable)
		}

		ids, err := svc.dao.LoadBalancer().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create clb, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create clb failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create clb but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func convClbReqToTable[T coreclb.Extension](kt *kit.Kit, vendor enumor.Vendor, lb dataproto.ClbBatchCreate[T]) (
	*tableclb.LoadBalancerTable, error) {
	extension, err := json.MarshalToString(lb.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return &tableclb.LoadBalancerTable{
		CloudID:              lb.CloudID,
		Name:                 lb.Name,
		Vendor:               vendor,
		AccountID:            lb.AccountID,
		BkBizID:              lb.BkBizID,
		Region:               lb.Region,
		Zones:                lb.Zones,
		BackupZones:          lb.BackupZones,
		LBType:               lb.LoadBalancerType,
		VpcID:                lb.VpcID,
		CloudVpcID:           lb.CloudVpcID,
		SubnetID:             lb.SubnetID,
		CloudSubnetID:        lb.CloudSubnetID,
		PrivateIPv4Addresses: lb.PrivateIPv4Addresses,
		PrivateIPv6Addresses: lb.PrivateIPv6Addresses,
		PublicIPv4Addresses:  lb.PublicIPv4Addresses,
		PublicIPv6Addresses:  lb.PublicIPv6Addresses,
		Domain:               lb.Domain,
		Status:               lb.Status,
		Memo:                 converter.PtrToVal(lb.Memo),
		CloudCreatedTime:     lb.CloudCreatedTime,
		CloudStatusTime:      lb.CloudStatusTime,
		CloudExpiredTime:     lb.CloudExpiredTime,
		Extension:            types.JsonField(extension),
		Creator:              kt.User,
		Reviser:              kt.User,
	}, nil
}
