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
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	typesdao "hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateListener 批量创建监听器
func (svc *lbSvc) BatchCreateListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateListener[corelb.TCloudListenerExtension](cts, svc)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateListener[T corelb.ListenerExtension](cts *rest.Contexts, svc *lbSvc) (any, error) {
	req := new(dataproto.ListenerBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.LoadBalancerListenerTable, 0, len(req.Listeners))
		for _, item := range req.Listeners {
			ext, err := json.MarshalToString(item.Extension)
			if err != nil {
				logs.Errorf("fail to marshal listener extension to json, err: %v, extension: %+v, rid: %s",
					err, ext, cts.Kit.Rid)
				return nil, err
			}
			models = append(models, &tablelb.LoadBalancerListenerTable{
				CloudID:       item.CloudID,
				Name:          item.Name,
				Vendor:        item.Vendor,
				AccountID:     item.AccountID,
				BkBizID:       item.BkBizID,
				LBID:          item.LbID,
				CloudLBID:     item.CloudLbID,
				Protocol:      item.Protocol,
				Port:          item.Port,
				DefaultDomain: item.DefaultDomain,
				Region:        item.Region,
				Extension:     types.JsonField(ext),
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			})
		}
		ids, err := svc.dao.LoadBalancerListener().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("fail to batch create listener, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create listener failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create listener but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchCreateListenerWithRule 批量创建监听器及规则
func (svc *lbSvc) BatchCreateListenerWithRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudListenerWithRule(cts)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func (svc *lbSvc) batchCreateTCloudListenerWithRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ListenerWithRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids, err := svc.insertListenerWithRule(cts.Kit, enumor.TCloud, req)
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *lbSvc) insertListenerWithRule(kt *kit.Kit, vendor enumor.Vendor,
	req *dataproto.ListenerWithRuleBatchCreateReq) ([]string, error) {

	result, err := svc.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		lblIDs := make([]string, 0, len(req.ListenerWithRules))
		for _, item := range req.ListenerWithRules {
			lblID, ruleID, err := svc.createListenerWithRule(kt, txn, item)
			if err != nil {
				logs.Errorf("fail to create listener with rule, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			lblIDs = append(lblIDs, lblID)
			if len(item.TargetGroupID) == 0 {
				continue
			}
			// 目标组如果没有RS，则直接绑定成功
			rsReq := &typesdao.ListOption{
				Filter: tools.EqualExpression("target_group_id", item.TargetGroupID),
				Page:   core.NewDefaultBasePage(),
			}
			targetResp, err := svc.dao.LoadBalancerTarget().List(kt, rsReq)
			if err != nil {
				logs.Errorf("fail to list target by target group id, err: %v, tgID: %s, rid: %s",
					err, item.TargetGroupID, kt.Rid)
				return nil, err
			}
			bindStatus := enumor.BindingBindingStatus
			if len(targetResp.Details) == 0 {
				bindStatus = enumor.SuccessBindingStatus
			}

			ruleRelModels := []*tablelb.TargetGroupListenerRuleRelTable{{
				Vendor:              vendor,
				ListenerRuleID:      ruleID,
				CloudListenerRuleID: item.CloudRuleID,
				ListenerRuleType:    item.RuleType,
				TargetGroupID:       item.TargetGroupID,
				CloudTargetGroupID:  item.CloudTargetGroupID,
				LbID:                item.LbID,
				CloudLbID:           item.CloudLbID,
				LblID:               lblID,
				CloudLblID:          item.CloudID,
				BindingStatus:       bindStatus,
				Creator:             kt.User,
				Reviser:             kt.User,
			}}
			_, err = svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(kt, txn, ruleRelModels)
			if err != nil {
				logs.Errorf("fail to batch create listener rule rel, err: %v, rid:%s", err, kt.Rid)
				return nil, fmt.Errorf("batch create listener rule rel failed, err: %v", err)
			}
		}
		return lblIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// createListenerWithRule 创建监听器和规则
func (svc *lbSvc) createListenerWithRule(kt *kit.Kit, txn *sqlx.Tx, item dataproto.ListenerWithRuleCreateReq) (
	lblID string, ruleID string, err error) {

	ext := corelb.TCloudListenerExtension{
		EndPort:     item.EndPort,
		Certificate: nil,
	}
	extRaw, err := json.MarshalToString(ext)
	if err != nil {
		logs.Errorf("json marshal json  extension failed, err: %v, ext: %+v, rid: %s", err, ext, kt.Rid)
		return "", "", err
	}

	models := []*tablelb.LoadBalancerListenerTable{{
		CloudID:       item.CloudID,
		Name:          item.Name,
		Vendor:        item.Vendor,
		AccountID:     item.AccountID,
		BkBizID:       item.BkBizID,
		LBID:          item.LbID,
		CloudLBID:     item.CloudLbID,
		Protocol:      item.Protocol,
		Port:          item.Port,
		DefaultDomain: item.Domain,
		Region:        item.Region,
		SniSwitch:     item.SniSwitch,
		Extension:     types.JsonField(extRaw),
		Creator:       kt.User,
		Reviser:       kt.User,
	}}
	lblIDs, err := svc.dao.LoadBalancerListener().BatchCreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("fail to batch create listener, err: %v, rid:%s", err, kt.Rid)
		return "", "", fmt.Errorf("batch create listener failed, err: %v", err)
	}
	certJSON, err := json.MarshalToString(item.Certificate)
	if err != nil {
		logs.Errorf("json marshal Certificate failed, err: %v", err)
		return "", "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleModels := []*tablelb.TCloudLbUrlRuleTable{{
		CloudID:            item.CloudRuleID,
		RuleType:           item.RuleType,
		LbID:               item.LbID,
		CloudLbID:          item.CloudLbID,
		LblID:              lblIDs[0],
		CloudLBLID:         item.CloudID,
		TargetGroupID:      item.TargetGroupID,
		CloudTargetGroupID: item.CloudTargetGroupID,
		Region:             item.Region,
		Domain:             item.Domain,
		URL:                item.Url,
		Scheduler:          item.Scheduler,
		SessionType:        item.SessionType,
		SessionExpire:      item.SessionExpire,
		Certificate:        types.JsonField(certJSON),
		BkBizID:            item.BkBizID,
		AccountID:          item.AccountID,
		Creator:            kt.User,
		Reviser:            kt.User,
	}}
	ruleIDs, err := svc.dao.LoadBalancerTCloudUrlRule().BatchCreateWithTx(kt, txn, ruleModels)
	if err != nil {
		logs.Errorf("fail to batch create listener url rule, err: %v, rid:%s", err, kt.Rid)
		return "", "", fmt.Errorf("batch create listener url rule failed, err: %v", err)
	}
	return lblIDs[0], ruleIDs[0], nil
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
			extensionJSON, err := types.NewJsonField(item.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			// 更新监听器
			lblInfo := &tablelb.LoadBalancerListenerTable{
				Name:          item.Name,
				BkBizID:       item.BkBizID,
				SniSwitch:     item.SniSwitch,
				DefaultDomain: item.DefaultDomain,
				Region:        item.Region,
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
