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

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	devicetype "hcm/pkg/dal/table/cloud/device-type"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateDeviceType batch update device type
func (svc *service) BatchUpdateDeviceType(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloudZiyan:
		return batchUpdateDeviceType(cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchUpdateDeviceType(cts *rest.Contexts, svc *service, vendor enumor.Vendor) (interface{}, error) {
	req := new(protocloud.DeviceTypeBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, updateReq := range req.DeviceTypes {
			record := &devicetype.DeviceTypeTable{
				ID:      updateReq.ID,
				Vendor:  vendor,
				Reviser: cts.Kit.User,
			}
			if updateReq.DeviceType != nil {
				record.DeviceType = cvt.PtrToVal(updateReq.DeviceType)
			}
			if updateReq.DeviceClass != nil {
				record.DeviceClass = cvt.PtrToVal(updateReq.DeviceClass)
			}
			if updateReq.DeviceFamily != nil {
				record.DeviceFamily = cvt.PtrToVal(updateReq.DeviceFamily)
			}
			if updateReq.CoreType != nil {
				record.CoreType = cvt.PtrToVal(updateReq.CoreType)
			}
			if updateReq.CpuCore != nil {
				record.CpuCore = cvt.PtrToVal(updateReq.CpuCore)
			}
			if updateReq.Memory != nil {
				record.Memory = cvt.PtrToVal(updateReq.Memory)
			}
			if updateReq.DeviceTypeClass != nil {
				record.DeviceTypeClass = cvt.PtrToVal(updateReq.DeviceTypeClass)
			}
			if updateReq.TechnicalClass != nil {
				record.TechnicalClass = cvt.PtrToVal(updateReq.TechnicalClass)
			}
			if updateReq.Region != nil {
				record.Region = cvt.PtrToVal(updateReq.Region)
			}
			if updateReq.Zone != nil {
				record.Zone = cvt.PtrToVal(updateReq.Zone)
			}
			if updateReq.Disable != nil {
				record.Disable = updateReq.Disable
			}
			if updateReq.Source != nil {
				record.Source = cvt.PtrToVal(updateReq.Source)
			}

			flt := tools.EqualExpression("id", updateReq.ID)
			if err := svc.dao.DeviceType().UpdateWithTx(cts.Kit, txn, flt, record); err != nil {
				logs.Errorf("update device type loop failed, id: %s, err: %v, rid: %s", updateReq.ID, err, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update device type failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
