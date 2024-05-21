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

package cert

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchDeleteCert batch delete cert.
func (svc *certSvc) BatchDeleteCert(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CertBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete cert decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete cert validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor", "cloud_id", "bk_biz_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.Cert().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cert db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cert failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if err = svc.dao.Cert().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete cert failed, delIDs: %v, err: %v, rid: %s", delIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
