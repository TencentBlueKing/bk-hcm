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

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchDeleteLoadBalancer delete clb
func (svc *lbSvc) BatchDeleteLoadBalancer(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ClbBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor", "cloud_id", "bk_biz_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list clb failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	lbIds := slice.Map(listResp.Details, func(one tablelb.LoadBalancerTable) string { return one.ID })

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		// 本层直接级联删除，有数据不报错
		// 删除对应监听器
		for _, lbId := range lbIds {
			err := svc.deleteListener(cts.Kit, txn, lbId)
			if err != nil {
				logs.Errorf("fail to delete listener of load balancer(%s), err: %v, rid: %s", lbId, err, cts.Kit.Rid)
				return nil, err
			}
		}
		// 删除负载均衡
		delFilter := tools.ContainersExpression("id", lbIds)
		return nil, svc.dao.LoadBalancer().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete clb(ids=%v) failed, err: %v, rid: %s", lbIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// 删除负载均衡关联规则
func (svc *lbSvc) deleteListener(kt *kit.Kit, txn *sqlx.Tx, lbId string) error {
	listenerResp, err := svc.dao.LoadBalancerListener().List(kt, &types.ListOption{
		Filter: tools.EqualExpression("lb_id", lbId),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list listener of load balancer(%s), err: %v, rid: %s", lbId, err, kt.Rid)
		return err
	}
	if len(listenerResp.Details) == 0 {
		return nil
	}
	// 删除对应的规则
	for _, listener := range listenerResp.Details {
		err := svc.deleteRule(kt, txn, listener.ID)
		if err != nil {
			logs.Errorf("fail to delete load balancer rule of listener(%s), err: %v, rid: %s",
				listener.ID, err, kt.Rid)
			return err
		}
	}
	// 删除监听器本身
	listenerIds := slice.Map(listenerResp.Details, func(r tablelb.LoadBalancerListenerTable) string { return r.ID })
	listenerIdFilter := tools.ContainersExpression("lbl_id", listenerIds)
	return svc.dao.LoadBalancerTCloudUrlRule().DeleteWithTx(kt, txn, listenerIdFilter)
}

// 删除监听器关联规则
func (svc *lbSvc) deleteRule(kt *kit.Kit, txn *sqlx.Tx, listenerID string) error {
	ruleResp, err := svc.dao.LoadBalancerTCloudUrlRule().List(kt, &types.ListOption{
		Filter: tools.EqualExpression("lbl_id", listenerID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list load balancer rule of listener(%s), err: %v, rid: %s", listenerID, err, kt.Rid)
		return err
	}
	if len(ruleResp.Details) == 0 {
		return nil
	}
	ruleIds := slice.Map(ruleResp.Details, func(r tablelb.TCloudClbUrlRuleTable) string { return r.ID })

	ruleIDFilter := tools.ContainersExpression("lbl_id", ruleIds)
	return svc.dao.LoadBalancerTCloudUrlRule().DeleteWithTx(kt, txn, ruleIDFilter)
}
