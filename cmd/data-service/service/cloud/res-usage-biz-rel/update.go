/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * r Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain r copy of the License at http://opensource.org/licenses/MIT
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

package resusagebizrel

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SetResUsageBizRel update resUsage biz rel. 全量覆盖更新
func (r *service) SetResUsageBizRel(cts *rest.Contexts) (interface{}, error) {

	resType := cts.PathParameter("res_type").String()
	resID := cts.PathParameter("res_id").String()

	req := new(protocloud.ResUsageBizRelUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := r.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ftr := tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType))

		if err := r.dao.ResUsageBizRel().DeleteWithTx(cts.Kit, txn, ftr); err != nil {
			return nil, fmt.Errorf("delete res usage biz rels failed, err: %v", err)
		}

		rels := make([]*tablecloud.ResUsageBizRelTable, len(req.UsageBizIDs))
		for index, bizID := range req.UsageBizIDs {
			rels[index] = &tablecloud.ResUsageBizRelTable{
				UsageBizID: bizID,
				ResID:      resID,
				ResType:    enumor.CloudResourceType(resType),
				ResVendor:  req.ResVendor,
				ResCloudID: req.ResCloudID,
				RelCreator: cts.Kit.User,
			}
		}
		if err := r.dao.ResUsageBizRel().BatchCreateWithTx(cts.Kit, txn, rels); err != nil {
			return nil, fmt.Errorf("batch create res usage biz rels failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res usage biz rel failed, err: %v, res_type: %s, res_id: %s, usage bizs: %v, rid: %s",
			err, resType, resID, req.UsageBizIDs, cts.Kit.Rid)
		return nil, err
	}

	return nil, err
}
