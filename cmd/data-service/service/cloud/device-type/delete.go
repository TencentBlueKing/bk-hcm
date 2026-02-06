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
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchDeleteDeviceType batch delete device type
func (svc *service) BatchDeleteDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &types.ListOption{
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}

	delIDs := make([]string, 0)
	for {
		listResp, err := svc.dao.DeviceType().List(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("batch delete list device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("delete list device type failed, err: %v", err)
		}

		for _, one := range listResp.Details {
			delIDs = append(delIDs, one.ID)
		}

		if len(listResp.Details) < int(listOpt.Page.Limit) {
			break
		}
		listOpt.Page.Start += uint32(listOpt.Page.Limit)
	}

	if len(delIDs) == 0 {
		return nil, nil
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, batchIDs := range slice.Split(delIDs, constant.BatchOperationMaxLimit) {
			delFilter := tools.ContainersExpression("id", batchIDs)
			if err := svc.dao.DeviceType().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch delete device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
