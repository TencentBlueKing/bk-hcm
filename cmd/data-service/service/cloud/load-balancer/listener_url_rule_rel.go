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

package loadbalancer

import (
	"fmt"

	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateListenerRuleRelStatusByTGID 根据目标组id 批量修改目标组和规则、监听器关系的状态
func (svc *lbSvc) BatchUpdateListenerRuleRelStatusByTGID(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("tg_id").String()

	req := new(dataproto.TGListenerRelStatusUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	model := &tablelb.TargetGroupListenerRuleRelTable{
		BindingStatus: req.BindingStatus,
		Detail:        req.Detail,
		Reviser:       cts.Kit.User,
	}
	tgFilter := tools.EqualExpression("target_group_id", tgID)
	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().Update(cts.Kit, tgFilter, model)
		if err != nil {
			logs.Errorf("fail to update listener rule rel status by target group(%s), err: %v, rid:%s",
				tgID, err, cts.Kit.Rid)
			return nil, fmt.Errorf("update target group listener rel by target group(%s) failed, err: %v", tgID, err)
		}
		return nil, nil
	})
}
