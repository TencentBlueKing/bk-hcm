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

package sgcvmrel

import (
	"fmt"

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateSgCvmRels rels.
func (svc *sgCvmRelSvc) BatchCreateSgCvmRels(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCvmRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablecloud.SecurityGroupCvmRelTable, 0, len(req.Rels))
		for _, one := range req.Rels {
			models = append(models, tablecloud.SecurityGroupCvmRelTable{
				CvmID:           one.CvmID,
				SecurityGroupID: one.SecurityGroupID,
				Creator:         cts.Kit.User,
			})
		}

		if err := svc.dao.SGCvmRel().BatchCreateWithTx(cts.Kit, txn, models); err != nil {
			return nil, fmt.Errorf("batch create sg cvm rels failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch create sg cvm rels failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
