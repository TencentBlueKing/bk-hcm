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
				Name:    lb.Name,
				BkBizID: lb.BkBizID,
				Domain:  lb.Domain,
				Status:  lb.Status,

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

// BatchUpdateClbBizInfo 批量更新业务信息
func (svc *lbSvc) BatchUpdateClbBizInfo(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ClbBizBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateFiled := &tablelb.LoadBalancerTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}
	return nil, svc.dao.LoadBalancer().Update(cts.Kit, updateFilter, updateFiled)
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

	updateData := &tablelb.LoadBalancerTargetGroupTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}

	if len(req.Name) > 0 {
		updateData.Name = req.Name
	}
	if len(req.TargetGroupType) > 0 {
		updateData.TargetGroupType = req.TargetGroupType
	}
	if len(req.CloudVpcID) > 0 {
		// 根据cloudVpcID查询VPC信息，如查不到vpcInfo则报错
		vpcReq := []dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{{CloudVpcID: req.CloudVpcID}}
		vpcInfoMap, err := getVpcMapByIDs(cts.Kit, vpcReq)
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
	if len(req.Region) > 0 {
		updateData.Region = req.Region
	}
	if len(req.Protocol) > 0 {
		updateData.Protocol = req.Protocol
	}
	if req.Port >= 0 {
		updateData.Port = req.Port
	}
	if len(req.HealthCheck) > 0 {
		updateData.HealthCheck = req.HealthCheck
	}

	if err := svc.dao.LoadBalancerTargetGroup().Update(
		cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
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
				SessionExpire:      rule.SessionExpire,
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
			model := &tablelb.LoadBalancerFlowRelTable{
				ResID:   item.ResID,
				Status:  item.Status,
				Reviser: cts.Kit.User,
			}
			filter := tools.ExpressionAnd(
				tools.RuleEqual("id", item.ID),
				tools.RuleEqual("res_id", item.ResID),
				tools.RuleEqual("flow_id", item.FlowID),
			)
			if err := svc.dao.LoadBalancerFlowRel().Update(cts.Kit, filter, model); err != nil {
				logs.Errorf("update res flow rel failed, err: %v, id: %s, rid: %s", err, item.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update res flow rel failed, id: %s, serr: %v", item.ID, err)
			}
		}
		return nil, nil
	})
}
