/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateLoadBalancer 批量跟新clb信息
func (svc *lbSvc) BatchUpdateLoadBalancer(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateLoadBalancer[corelb.TCloudClbExtension](cts, svc)

	default:
		return nil, fmt.Errorf("unsupport  vendor %s", vendor)
	}

}

func batchUpdateLoadBalancer[T corelb.Extension](cts *rest.Contexts, svc *lbSvc) (any, error) {

	req := new(dataproto.LbExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lbIds := slice.Map(req.Lbs, func(one *dataproto.LoadBalancerExtUpdateReq[T]) string { return one.ID })

	extensionMap, err := svc.listClbExt(cts.Kit, lbIds)
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, lb := range req.Lbs {
			update := &tablelb.LoadBalancerTable{
				Name:                 lb.Name,
				BkBizID:              lb.BkBizID,
				Domain:               lb.Domain,
				Status:               lb.Status,
				VpcID:                lb.VpcID,
				CloudVpcID:           lb.CloudVpcID,
				SubnetID:             lb.SubnetID,
				CloudSubnetID:        lb.CloudSubnetID,
				IPVersion:            string(lb.IPVersion),
				PrivateIPv4Addresses: lb.PrivateIPv4Addresses,
				PrivateIPv6Addresses: lb.PrivateIPv6Addresses,
				PublicIPv4Addresses:  lb.PublicIPv4Addresses,
				PublicIPv6Addresses:  lb.PublicIPv6Addresses,

				CloudCreatedTime: lb.CloudCreatedTime,
				CloudStatusTime:  lb.CloudStatusTime,
				CloudExpiredTime: lb.CloudExpiredTime,
				Memo:             lb.Memo,
				Reviser:          cts.Kit.User,
			}

			if lb.Extension != nil {
				extension, exist := extensionMap[lb.ID]
				if !exist {
					continue
				}

				merge, err := json.UpdateMerge(lb.Extension, string(extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				update.Extension = tabletype.JsonField(merge)
			}

			if err := svc.dao.LoadBalancer().UpdateByIDWithTx(cts.Kit, txn, lb.ID, update); err != nil {
				logs.Errorf("update load balancer by id failed, err: %v, id: %s, rid: %s", err, lb.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update load balancer failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *lbSvc) listClbExt(kt *kit.Kit, ids []string) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
	}

	resp, err := svc.dao.LoadBalancer().List(kt, opt)
	if err != nil {
		return nil, err
	}

	return converter.SliceToMap(resp.Details, func(t tablelb.LoadBalancerTable) (string, tabletype.JsonField) {
		return t.ID, t.Extension
	}), nil

}

// BatchUpdateLbBizInfo 批量更新业务信息
func (svc *lbSvc) BatchUpdateLbBizInfo(cts *rest.Contexts) (any, error) {
	req := new(dataproto.BizBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateField := &tablelb.LoadBalancerTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}
	return nil, svc.dao.LoadBalancer().Update(cts.Kit, updateFilter, updateField)
}

// BatchUpdateTargetGroupBizInfo 批量更新目标组业务信息
func (svc *lbSvc) BatchUpdateTargetGroupBizInfo(cts *rest.Contexts) (any, error) {
	req := new(dataproto.BizBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateField := &tablelb.LoadBalancerTargetGroupTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}
	return nil, svc.dao.LoadBalancerTargetGroup().Update(cts.Kit, updateFilter, updateField)
}

// BatchUpdateListenerBizInfo 批量更新监听器业务信息
func (svc *lbSvc) BatchUpdateListenerBizInfo(cts *rest.Contexts) (any, error) {
	req := new(dataproto.BizBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateField := &tablelb.LoadBalancerListenerTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}
	return nil, svc.dao.LoadBalancerListener().Update(cts.Kit, updateFilter, updateField)
}

// UpdateTargetGroup batch update argument template
func (svc *lbSvc) UpdateTargetGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.TargetGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(req.IDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "ids is empty")
	}

	tgReq := &types.ListOption{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   &core.BasePage{Limit: 1},
	}
	tgList, err := svc.dao.LoadBalancerTargetGroup().List(cts.Kit, tgReq)
	if err != nil {
		return nil, err
	}
	if len(tgList.Details) != len(req.IDs) {
		return nil, errors.New("not all target groups can be found")
	}
	updateDataList := make([]*tablelb.LoadBalancerTargetGroupTable, 0, len(req.IDs))
	for _, oldTg := range tgList.Details {

		updateData := &tablelb.LoadBalancerTargetGroupTable{
			ID:              oldTg.ID,
			Name:            req.Name,
			BkBizID:         req.BkBizID,
			TargetGroupType: req.TargetGroupType,
			Region:          req.Region,
			Protocol:        req.Protocol,
			Port:            req.Port,
			Weight:          req.Weight,
			Reviser:         cts.Kit.User,
		}

		if len(req.CloudVpcID) > 0 {
			// 根据cloudVpcID查询VPC信息，如查不到vpcInfo则报错
			vpcInfoMap, err := getVpcMapByIDs(cts.Kit, []string{req.CloudVpcID})
			if err != nil {
				return nil, err
			}
			vpcInfo, ok := vpcInfoMap[req.CloudVpcID]
			if !ok {
				return nil, errf.Newf(errf.RecordNotFound, "vpcID[%s] not found", req.VpcID)
			}
			updateData.VpcID = vpcInfo.ID
			updateData.CloudVpcID = vpcInfo.CloudID
		}
		if req.HealthCheck != nil {
			mergedHealth, err := json.UpdateMerge(req.HealthCheck, string(oldTg.HealthCheck))
			if err != nil {
				return nil, fmt.Errorf("json UpdateMerge rule health check failed, err: %v", err)
			}
			updateData.HealthCheck = tabletype.JsonField(mergedHealth)
		}
		updateDataList = append(updateDataList, updateData)
	}
	if err := svc.dao.LoadBalancerTargetGroup().UpdateBatch(cts.Kit, updateDataList); err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchUpdateTCloudUrlRule ..
func (svc *lbSvc) BatchUpdateTCloudUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TCloudUrlRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleIds := slice.Map(req.UrlRules, func(one *dataproto.TCloudUrlRuleUpdate) string { return one.ID })

	healthCertMap, err := svc.listRuleHealthAndCert(cts.Kit, ruleIds)
	if err != nil {
		logs.Errorf("fail to list health and cert of tcloud url rule, err: %s, ruleIds: %v, rid: %s",
			err, ruleIds, cts.Kit.Rid)
		return nil, err
	}

	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, rule := range req.UrlRules {
			update := &tablelb.TCloudLbUrlRuleTable{
				Name:               rule.Name,
				Domain:             rule.Domain,
				URL:                rule.URL,
				TargetGroupID:      rule.TargetGroupID,
				CloudTargetGroupID: rule.CloudTargetGroupID,
				Scheduler:          rule.Scheduler,
				SessionExpire:      converter.PtrToVal(rule.SessionExpire),
				SessionType:        rule.SessionType,
				Memo:               rule.Memo,
				Reviser:            cts.Kit.User,
			}

			if rule.HealthCheck != nil {
				hc := healthCertMap[rule.ID]
				mergedHealth, err := json.UpdateMerge(rule.HealthCheck, string(hc.Health))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge rule health check failed, err: %v", err)
				}
				update.HealthCheck = tabletype.JsonField(mergedHealth)

			}
			if rule.Certificate != nil {
				hc := healthCertMap[rule.ID]
				mergedCert, err := json.UpdateMerge(rule.Certificate, string(hc.Cert))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge rule cert failed, err: %v", err)
				}
				update.Certificate = tabletype.JsonField(mergedCert)
			}

			if err = svc.dao.LoadBalancerTCloudUrlRule().UpdateByIDWithTx(cts.Kit, txn, rule.ID, update); err != nil {
				logs.Errorf("update tcloud rule by id failed, err: %v, id: %s, rid: %s", err, rule.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
}

// 更新目标组健康检查
func (svc *lbSvc) updateTGHealth(kt *kit.Kit, txn *sqlx.Tx, tgID string, health tabletype.JsonField) error {
	if len(tgID) == 0 {
		return nil
	}
	tgUpdate := &tablelb.LoadBalancerTargetGroupTable{
		HealthCheck: health,
		Reviser:     kt.User,
	}
	return svc.dao.LoadBalancerTargetGroup().UpdateByIDWithTx(kt, txn, tgID, tgUpdate)
}

// tcloudHealthCert 腾讯云监听器、规则健康检查和证书信息
type tcloudHealthCert struct {
	Health tabletype.JsonField
	Cert   tabletype.JsonField
}

func (svc *lbSvc) listRuleHealthAndCert(kt *kit.Kit, ruleIds []string) (map[string]tcloudHealthCert, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ruleIds),
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
	}

	resp, err := svc.dao.LoadBalancerTCloudUrlRule().List(kt, opt)
	if err != nil {
		return nil, err
	}

	return converter.SliceToMap(resp.Details, func(t tablelb.TCloudLbUrlRuleTable) (string, tcloudHealthCert) {
		return t.ID, tcloudHealthCert{Health: t.HealthCheck, Cert: t.Certificate}
	}), nil
}

// BatchUpdateListener 批量更新监听器基本信息
func (svc *lbSvc) BatchUpdateListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateListener[corelb.TCloudListenerExtension](cts)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchUpdateListener[T corelb.ListenerExtension](cts *rest.Contexts) (any, error) {
	req := new(dataproto.ListenerBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, item := range req.Listeners {
			extensionJSON, err := tabletype.NewJsonField(item.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			// 更新监听器
			lblInfo := &tablelb.LoadBalancerListenerTable{
				Name:          item.Name,
				BkBizID:       item.BkBizID,
				SniSwitch:     item.SniSwitch,
				DefaultDomain: item.DefaultDomain,
				Extension:     extensionJSON,
				Reviser:       cts.Kit.User,
			}
			if err = svc.dao.LoadBalancerListener().Update(
				cts.Kit, tools.EqualExpression("id", item.ID), lblInfo); err != nil {
				logs.Errorf("update listener by id failed, err: %v, id: %s, rid: %s", err, item.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update listener by id failed, lblID: %s, serr: %v", item.ID, err)
			}
		}
		return nil, nil
	})
}

// BatchUpdateResFlowRel 批量更新资源跟Flow关联关系的记录
func (svc *lbSvc) BatchUpdateResFlowRel(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowRelBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, item := range req.ResFlowRels {
			model := &tablelb.ResourceFlowRelTable{
				TaskType: item.TaskType,
				Status:   item.Status,
				Reviser:  cts.Kit.User,
			}
			filter := tools.ExpressionAnd(
				tools.RuleEqual("id", item.ID),
				tools.RuleEqual("res_id", item.ResID),
				tools.RuleEqual("res_type", item.ResType),
				tools.RuleEqual("flow_id", item.FlowID),
			)
			if err := svc.dao.ResourceFlowRel().Update(cts.Kit, filter, model); err != nil {
				logs.Errorf("update res flow rel failed, err: %v, id: %s, rid: %s", err, item.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update res flow rel failed, id: %s, serr: %v", item.ID, err)
			}
		}
		return nil, nil
	})
}

// ResFlowUnLock res flow unlock.
func (svc *lbSvc) ResFlowUnLock(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ResFlowLockReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("res flow unlock decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("res flow unlock validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		lockDelFilter := tools.ExpressionAnd(
			tools.RuleEqual("res_id", req.ResID),
			tools.RuleEqual("res_type", req.ResType),
			tools.RuleEqual("owner", req.FlowID),
		)
		if err := svc.dao.ResourceFlowLock().DeleteWithTx(cts.Kit, txn, lockDelFilter); err != nil {
			logs.Errorf("delete res flow lock failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}

		relModel := &tablelb.ResourceFlowRelTable{
			Status:  req.Status,
			Reviser: cts.Kit.User,
		}
		filter := tools.ExpressionAnd(
			tools.RuleEqual("res_id", req.ResID),
			tools.RuleEqual("res_type", req.ResType),
			tools.RuleEqual("flow_id", req.FlowID),
		)
		if err := svc.dao.ResourceFlowRel().Update(cts.Kit, filter, relModel); err != nil {
			logs.Errorf("update res flow rel failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, fmt.Errorf("update res flow rel failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("res flow unlock failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

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
