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
	"reflect"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateCvm cvm.
func (svc *cvmSvc) BatchCreateCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateCvm[corecvm.TCloudCvmExtension](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateCvm[corecvm.AwsCvmExtension](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateCvm[corecvm.HuaWeiCvmExtension](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateCvm[corecvm.AzureCvmExtension](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateCvm[corecvm.GcpCvmExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchCreateCvm[T corecvm.Extension](cts *rest.Contexts, svc *cvmSvc, vendor enumor.Vendor) (interface{}, error) {
	req := new(protocloud.CvmBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]*tablecvm.Table, 0, len(req.Cvms))
		for _, one := range req.Cvms {
			extension, err := json.MarshalToString(one.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			models = append(models, &tablecvm.Table{
				CloudID:              one.CloudID,
				Name:                 one.Name,
				Vendor:               vendor,
				BkBizID:              one.BkBizID,
				BkCloudID:            one.BkCloudID,
				AccountID:            one.AccountID,
				Region:               one.Region,
				Zone:                 one.Zone,
				CloudVpcIDs:          one.CloudVpcIDs,
				VpcIDs:               one.VpcIDs,
				CloudSubnetIDs:       one.CloudSubnetIDs,
				SubnetIDs:            one.SubnetIDs,
				CloudImageID:         one.CloudImageID,
				ImageID:              one.ImageID,
				OsName:               one.OsName,
				Memo:                 one.Memo,
				Status:               one.Status,
				PrivateIPv4Addresses: one.PrivateIPv4Addresses,
				PrivateIPv6Addresses: one.PrivateIPv6Addresses,
				PublicIPv4Addresses:  one.PublicIPv4Addresses,
				PublicIPv6Addresses:  one.PublicIPv6Addresses,
				MachineType:          one.MachineType,
				Extension:            tabletype.JsonField(extension),
				CloudCreatedTime:     one.CloudCreatedTime,
				CloudLaunchedTime:    one.CloudLaunchedTime,
				CloudExpiredTime:     one.CloudExpiredTime,
				Creator:              cts.Kit.User,
				Reviser:              cts.Kit.User,
			})
		}

		ids, err := svc.dao.Cvm().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create cvm failed, err: %v", err)
		}

		// create cmdb cloud hosts
		// 如果主机同步Cmdb失败，但写入HCM成功，忽略该错误。
		err = upsertCmdbHosts[T](svc, cts.Kit, vendor, models)
		if err != nil {
			logs.Errorf("[%s] upsert cmdb hosts failed, err: %v, rid: %s", constant.CmdbSyncFailed, err, cts.Kit.Rid)
			return nil, nil
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create cvm but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// InitCvm init cvm 仅验证建表使用，方案一不需要建表
func (svc *cvmSvc) InitCvm(cts *rest.Contexts) (any, error) {
	if err := svc.dao.Cvm().InitCreateTable(cts.Kit); err != nil {
		logs.Errorf("fail to init create cvm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("init create cvm failed, err: %v", err)
	}
	return nil, nil
}
