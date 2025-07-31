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
	"reflect"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchCreateTCloudUrlRule 批量创建腾讯云url规则 纯规则条目创建，不校验监听器， 有目标组则一起创建关联关系
func (svc *lbSvc) BatchCreateTCloudUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TCloudUrlRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("[ds] BatchCreateTCloudUrlRule request validate failed, err:%v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleModels := make([]*tablelb.TCloudLbUrlRuleTable, 0, len(req.UrlRules))
	for _, rule := range req.UrlRules {
		ruleModel, err := svc.convRule(cts.Kit, rule)
		if err != nil {
			return nil, err
		}
		ruleModels = append(ruleModels, ruleModel)
	}

	// 创建规则和关联关系
	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		ids, err := svc.dao.LoadBalancerTCloudUrlRule().BatchCreateWithTx(cts.Kit, txn, ruleModels)
		if err != nil {
			logs.Errorf("fail to batch create lb rule, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create lb rule failed, err: %v", err)
		}
		// 根据id 创建关联关系
		relModels := make([]*tablelb.TargetGroupListenerRuleRelTable, 0, len(req.UrlRules))
		for i, rule := range req.UrlRules {
			// 跳过没有设置目标组id的规则
			if len(rule.TargetGroupID) == 0 {
				continue
			}
			// 默认设置为绑定中状态，防止同步时本地目标组rs被清掉
			relModels = append(relModels, svc.convRuleRel(cts.Kit, ids[i], rule, enumor.BindingBindingStatus))
		}
		if len(relModels) == 0 {
			return ids, nil
		}
		_, err = svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(cts.Kit, txn, relModels)
		if err != nil {
			logs.Errorf("fail to create rule rel, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create tcloud url rule but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *lbSvc) convRuleRel(kt *kit.Kit, listenerRuleID string, rule dataproto.TCloudUrlRuleCreate,
	bindingStatus enumor.BindingStatus) *tablelb.TargetGroupListenerRuleRelTable {

	return &tablelb.TargetGroupListenerRuleRelTable{
		Vendor:              rule.Vendor,
		ListenerRuleID:      listenerRuleID,
		CloudListenerRuleID: rule.CloudID,
		ListenerRuleType:    enumor.Layer7RuleType,
		TargetGroupID:       rule.TargetGroupID,
		CloudTargetGroupID:  rule.CloudTargetGroupID,
		LbID:                rule.LbID,
		CloudLbID:           rule.CloudLbID,
		LblID:               rule.LblID,
		CloudLblID:          rule.CloudLBLID,
		BindingStatus:       bindingStatus,
		Detail:              "{}",
		Creator:             kt.User,
		Reviser:             kt.User,
	}
}

func (svc *lbSvc) convRule(kt *kit.Kit, rule dataproto.TCloudUrlRuleCreate) (
	*tablelb.TCloudLbUrlRuleTable, error) {

	ruleModel := &tablelb.TCloudLbUrlRuleTable{
		CloudID:            rule.CloudID,
		Name:               rule.Name,
		RuleType:           rule.RuleType,
		LbID:               rule.LbID,
		CloudLbID:          rule.CloudLbID,
		LblID:              rule.LblID,
		CloudLBLID:         rule.CloudLBLID,
		TargetGroupID:      rule.TargetGroupID,
		CloudTargetGroupID: rule.CloudTargetGroupID,
		Region:             rule.Region,
		Domain:             rule.Domain,
		URL:                rule.URL,
		Scheduler:          rule.Scheduler,
		SessionType:        rule.SessionType,
		SessionExpire:      rule.SessionExpire,
		Memo:               rule.Memo,

		Creator: kt.User,
		Reviser: kt.User,
	}
	healthCheckJson, err := json.MarshalToString(rule.HealthCheck)
	if err != nil {
		logs.Errorf("fail to marshal health check into json, err: %v, healthcheck: %+v, rid: %s",
			err, rule.HealthCheck, kt.Rid)
		return nil, err
	}
	ruleModel.HealthCheck = types.JsonField(healthCheckJson)
	certJson, err := json.MarshalToString(rule.Certificate)
	if err != nil {
		logs.Errorf("fail to marshal certificate into json, err: %v, certificate: %+v, rid: %s",
			err, rule.Certificate, kt.Rid)
		return nil, err
	}
	ruleModel.Certificate = types.JsonField(certJson)
	return ruleModel, nil
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
				Region:             rule.Region,
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
				update.HealthCheck = types.JsonField(mergedHealth)

			}
			if rule.Certificate != nil {
				hc := healthCertMap[rule.ID]
				mergedCert, err := json.UpdateMerge(rule.Certificate, string(hc.Cert))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge rule cert failed, err: %v", err)
				}
				update.Certificate = types.JsonField(mergedCert)
			}

			if err = svc.dao.LoadBalancerTCloudUrlRule().UpdateByIDWithTx(cts.Kit, txn, rule.ID, update); err != nil {
				logs.Errorf("update tcloud rule by id failed, err: %v, id: %s, rid: %s", err, rule.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
}

func (svc *lbSvc) listRuleHealthAndCert(kt *kit.Kit, ruleIds []string) (map[string]tcloudHealthCert, error) {
	opt := &daotypes.ListOption{
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
