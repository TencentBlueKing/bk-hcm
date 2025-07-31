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
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
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

// BatchDeleteLoadBalancer delete load balancer
func (svc *lbSvc) BatchDeleteLoadBalancer(cts *rest.Contexts) (any, error) {
	req := new(dataproto.LoadBalancerBatchDeleteReq)
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
		logs.Errorf("list lb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list lb failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	lbIds := slice.Map(listResp.Details, func(one tablelb.LoadBalancerTable) string { return one.ID })

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		// 本层直接级联删除，有数据不报错
		// 删除对应监听器
		for _, lbId := range lbIds {
			err := svc.deleteListenerByLb(cts.Kit, txn, lbId)
			if err != nil {
				logs.Errorf("fail to delete listener of load balancer(%s), err: %v, rid: %s", lbId, err, cts.Kit.Rid)
				return nil, err
			}
		}
		// 删除安全组关联关系
		sgRelFilter := tools.ExpressionAnd(
			tools.RuleIn("res_id", lbIds),
			tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
		)
		err := svc.dao.SGCommonRel().DeleteWithTx(cts.Kit, txn, sgRelFilter)
		if err != nil {
			logs.Errorf("delete lb sg rel failed , err: %v, lb_ids: %v, rid: %s", err, lbIds, cts.Kit.Rid)
			return nil, err
		}

		// 删除负载均衡
		delFilter := tools.ContainersExpression("id", lbIds)
		return nil, svc.dao.LoadBalancer().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete lb(ids=%v) failed, err: %v, rid: %s", lbIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// 删除负载均衡关联规则
func (svc *lbSvc) deleteListenerByLb(kt *kit.Kit, txn *sqlx.Tx, lbId string) error {
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
		err := svc.deleteRuleByListener(kt, txn, listener.ID)
		if err != nil {
			logs.Errorf("fail to delete load balancer rule of listener(%s), err: %v, rid: %s",
				listener.ID, err, kt.Rid)
			return err
		}
	}
	// 删除监听器本身
	listenerIds := slice.Map(listenerResp.Details, func(r tablelb.LoadBalancerListenerTable) string { return r.ID })
	listenerIdFilter := tools.ContainersExpression("id", listenerIds)
	return svc.dao.LoadBalancerListener().DeleteWithTx(kt, txn, listenerIdFilter)
}

// 删除监听器关联规则
func (svc *lbSvc) deleteRuleByListener(kt *kit.Kit, txn *sqlx.Tx, listenerID string) error {
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
	ruleIds := slice.Map(ruleResp.Details, func(r tablelb.TCloudLbUrlRuleTable) string { return r.ID })

	// 删除跟目标组的绑定关系
	err = svc.deleteTGListenerRuleRelByListener(kt, txn, []string{listenerID})
	if err != nil {
		logs.Errorf("fail to delete target rule rel of listener(%s), err: %v, rid: %s", listenerID, err, kt.Rid)
		return err
	}
	ruleIDFilter := tools.ContainersExpression("id", ruleIds)
	return svc.dao.LoadBalancerTCloudUrlRule().DeleteWithTx(kt, txn, ruleIDFilter)
}

// BatchDeleteTargetGroup batch delete target group.
func (svc *lbSvc) BatchDeleteTargetGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.TargetGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete target group decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete target group validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor", "cloud_id", "bk_biz_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.LoadBalancerTargetGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list target group db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list target group failed, err: %v", err)
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
		if err = svc.dao.LoadBalancerTargetGroup().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			logs.Errorf("fail to delete target group, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		rsDelFilter := tools.ContainersExpression("target_group_id", delIDs)
		// 删除关联RS
		if err = svc.dao.LoadBalancerTarget().DeleteWithTx(cts.Kit, txn, rsDelFilter); err != nil {
			logs.Errorf("fail to delete target, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete target group failed, delIDs: %v, err: %v, rid: %s", delIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteTCloudUrlRule 批量删除腾讯云规则
func (svc *lbSvc) BatchDeleteTCloudUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.LoadBalancerBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "cloud_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.LoadBalancerTCloudUrlRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud lb rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud lb rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	ruleIds := slice.Map(listResp.Details, func(one tablelb.TCloudLbUrlRuleTable) string { return one.ID })

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		// 删除关联关系
		ruleFilter := tools.ContainersExpression("listener_rule_id", ruleIds)
		err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().DeleteWithTx(cts.Kit, txn, ruleFilter)
		if err != nil {
			logs.Errorf("fail to delete rule target group relations, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		// 删除对应的规则
		delFilter := tools.ContainersExpression("id", ruleIds)
		return nil, svc.dao.LoadBalancerTCloudUrlRule().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete rules(ids=%v) failed, err: %v, rid: %s", ruleIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteListener delete listener
func (svc *lbSvc) BatchDeleteListener(cts *rest.Contexts) (any, error) {
	req := new(dataproto.LoadBalancerBatchDeleteReq)
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
	listResp, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list listener failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	lblIds := slice.Map(listResp.Details, func(one tablelb.LoadBalancerListenerTable) string { return one.ID })
	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		// 本层直接级联删除，有数据不报错
		// 删除对应监听器规则
		for _, lblId := range lblIds {
			err = svc.deleteRuleByListener(cts.Kit, txn, lblId)
			if err != nil {
				logs.Errorf("fail to delete rule of listener(%s), err: %v, rid: %s", lblId, err, cts.Kit.Rid)
				return nil, err
			}
		}

		// 删除跟目标组的绑定关系
		err = svc.deleteTGListenerRuleRelByListener(cts.Kit, txn, lblIds)
		if err != nil {
			logs.Errorf("fail to delete target rule rel of listener(%v), err: %v, rid: %s", lblIds, err, cts.Kit.Rid)
			return nil, err
		}

		// 删除监听器
		delFilter := tools.ContainersExpression("id", lblIds)
		return nil, svc.dao.LoadBalancerListener().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete listener(ids=%v) failed, err: %v, rid: %s", lblIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// 删除监听器关联目标组关系数据
func (svc *lbSvc) deleteTGListenerRuleRelByListener(kt *kit.Kit, txn *sqlx.Tx, listenerIDs []string) error {
	ruleRelResp, err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().List(kt, &types.ListOption{
		Filter: tools.ContainersExpression("lbl_id", listenerIDs),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list listener rule and target group relation(ids=%v), err: %v, rid: %s",
			listenerIDs, err, kt.Rid)
		return err
	}
	if len(ruleRelResp.Details) == 0 {
		return nil
	}

	relIDs := slice.Map(ruleRelResp.Details, func(r tablelb.TargetGroupListenerRuleRelTable) string { return r.ID })
	return svc.dao.LoadBalancerTargetGroupListenerRuleRel().DeleteWithTx(
		kt, txn, tools.ContainersExpression("id", relIDs))
}

// BatchDeleteResFlowRel batch delete res flow rel.
func (svc *lbSvc) BatchDeleteResFlowRel(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete res flow rel decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete res flow rel validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "res_id", "flow_id", "task_type"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.ResourceFlowRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list res flow rel db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list res flow rel failed, err: %v", err)
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
		if err = svc.dao.ResourceFlowRel().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete res flow rel failed, delIDs: %v, err: %v, rid: %s", delIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteResFlowLock batch delete res flow lock.
func (svc *lbSvc) DeleteResFlowLock(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ResFlowLockDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete res flow lock decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete res flow lock validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ExpressionAnd(
			tools.RuleEqual("res_id", req.ResID),
			tools.RuleEqual("res_type", req.ResType),
			tools.RuleEqual("owner", req.Owner),
		)
		if err := svc.dao.ResourceFlowLock().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete res flow lock failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
