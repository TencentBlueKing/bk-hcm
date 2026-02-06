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

package devicetype

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	devicetype "hcm/pkg/dal/table/cloud/device-type"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateDeviceType batch create device type
func (svc *service) BatchCreateDeviceType(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloudZiyan:
		return batchCreateDeviceType(cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchCreateDeviceType(cts *rest.Contexts, svc *service, vendor enumor.Vendor) (interface{}, error) {
	req := new(protocloud.DeviceTypeBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	models := make([]devicetype.DeviceTypeTable, 0, len(req.DeviceTypes))
	for _, createReq := range req.DeviceTypes {
		model := devicetype.DeviceTypeTable{
			Vendor:          vendor,
			DeviceType:      createReq.DeviceType,
			DeviceClass:     createReq.DeviceClass,
			DeviceFamily:    createReq.DeviceFamily,
			CoreType:        createReq.CoreType,
			CpuCore:         createReq.CpuCore,
			Memory:          createReq.Memory,
			DeviceTypeClass: createReq.DeviceTypeClass,
			TechnicalClass:  createReq.TechnicalClass,
			Region:          createReq.Region,
			Zone:            createReq.Zone,
			Disable:         cvt.ValToPtr(createReq.Disable),
			Source:          createReq.Source,
			Creator:         cts.Kit.User,
			Reviser:         cts.Kit.User,
		}
		models = append(models, model)
	}

	dtIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.dao.DeviceType().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create device type failed, err: %v", err)
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("batch create device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, err := util.GetStrSliceByInterface(dtIDs)
	if err != nil {
		logs.Errorf("batch create device type but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("batch create device type but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
