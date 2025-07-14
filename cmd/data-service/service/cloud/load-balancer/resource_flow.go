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
	"hcm/pkg/api/core/audit"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	typesdao "hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateResFlowLock 创建资源跟Flow的锁定关系
func (svc *lbSvc) CreateResFlowLock(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowLockCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		model := &tablelb.ResourceFlowLockTable{
			ResID:   req.ResID,
			ResType: req.ResType,
			Owner:   req.Owner,
			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		}
		err := svc.dao.ResourceFlowLock().CreateWithTx(cts.Kit, txn, model)
		if err != nil {
			logs.Errorf("[%s]fail to create load balancer flow lock, req: %+v, err: %v, rid:%s", req, err, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow lock failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchCreateResFlowRel 批量创建资源跟Flow的关系记录
func (svc *lbSvc) BatchCreateResFlowRel(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.ResourceFlowRelTable, 0, len(req.ResFlowRels))
		for _, item := range req.ResFlowRels {
			models = append(models, &tablelb.ResourceFlowRelTable{
				ResID:    item.ResID,
				ResType:  item.ResType,
				FlowID:   item.FlowID,
				TaskType: item.TaskType,
				Status:   item.Status,
				Creator:  cts.Kit.User,
				Reviser:  cts.Kit.User,
			})
		}
		ids, err := svc.dao.ResourceFlowRel().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create load balancer flow rel, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create load balancer flow rel failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create load balancer flow rel but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// ResFlowLock 锁定资源跟Flow 1. 锁表 2. 创建关联记录
func (svc *lbSvc) ResFlowLock(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowLockReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		lockModel := &tablelb.ResourceFlowLockTable{
			ResID:   req.ResID,
			ResType: req.ResType,
			Owner:   req.FlowID,
			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		}
		err := svc.dao.ResourceFlowLock().CreateWithTx(cts.Kit, txn, lockModel)
		if err != nil {
			logs.Errorf("fail to create load balancer flow lock, req: %+v, err: %v, rid:%s", req, err, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow lock failed, err: %v", err)
		}

		relModels := []*tablelb.ResourceFlowRelTable{{
			ResID:    req.ResID,
			ResType:  req.ResType,
			FlowID:   req.FlowID,
			TaskType: req.TaskType,
			Status:   req.Status,
			Creator:  cts.Kit.User,
			Reviser:  cts.Kit.User,
		}}
		_, err = svc.dao.ResourceFlowRel().BatchCreateWithTx(cts.Kit, txn, relModels)
		if err != nil {
			logs.Errorf("fail to create load balancer flow rel, err: %v, req: %+v, rid:%s", err, req, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow rel failed, err: %v", err)
		}

		// 创建目标组的操作记录
		if req.ResType == enumor.LoadBalancerCloudResType {
			err = svc.createTargetGroupOfResFlowAudit(cts.Kit, req, txn)
			if err != nil {
				logs.Errorf("fail to create res flow audits, err: %v, req: %+v, rid:%s", err, req, cts.Kit.Rid)
				return nil, fmt.Errorf("create res flow audits failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// createTargetGroupOfResFlowAudit 创建目标组的操作记录
func (svc *lbSvc) createTargetGroupOfResFlowAudit(kt *kit.Kit, req *dataproto.ResFlowLockReq, txn *sqlx.Tx) error {
	resReq := &typesdao.ListOption{
		Filter: tools.EqualExpression("id", req.ResID),
		Page:   core.NewDefaultBasePage(),
	}
	resList, err := svc.dao.LoadBalancer().List(kt, resReq)
	if err != nil {
		return err
	}
	if len(resList.Details) == 0 {
		return errf.Newf(errf.RecordNotFound, "resID: %s, resType: %s, is not found", req.ResID, req.ResType)
	}

	resInfo := resList.Details[0]

	var auditData = audit.TargetGroupAsyncAuditDetail{
		LoadBalancer: resInfo,
		ResFlow:      req,
	}

	audits := make([]*tableaudit.AuditTable, 0)
	audits = append(audits, &tableaudit.AuditTable{
		ResID:      resInfo.ID,
		CloudResID: resInfo.CloudID,
		ResName:    resInfo.Name,
		ResType:    enumor.AuditResourceType(req.ResType),
		Action:     enumor.Update,
		BkBizID:    resInfo.BkBizID,
		Vendor:     resInfo.Vendor,
		AccountID:  resInfo.AccountID,
		Operator:   kt.User,
		Source:     kt.GetRequestSource(),
		Rid:        kt.Rid,
		AppCode:    kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: auditData,
		},
	})
	if err = svc.dao.Audit().BatchCreateWithTx(kt, txn, audits); err != nil {
		logs.Errorf("batch create %s audit failed, err: %v, req: %+v, rid: %s", req.ResType, err, req, kt.Rid)
		return err
	}

	return nil
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

// ListResFlowLock list res flow lock.
func (svc *lbSvc) ListResFlowLock(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typesdao.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.ResourceFlowLock().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list res flow lock failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list res flow lock failed, err: %v", err)
	}

	if req.Page.Count {
		return &dataproto.ResFlowLockListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseResFlowLock, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corelb.BaseResFlowLock{
			ResID:   one.ResID,
			ResType: one.ResType,
			Owner:   one.Owner,
			Revision: &core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &dataproto.ResFlowLockListResult{Details: details}, nil
}

// ListResFlowRel list res flow rel.
func (svc *lbSvc) ListResFlowRel(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typesdao.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.ResourceFlowRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list res flow rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list res flow rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &dataproto.ResFlowRelListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseResFlowRel, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corelb.BaseResFlowRel{
			ID:       one.ID,
			ResID:    one.ResID,
			FlowID:   one.FlowID,
			TaskType: one.TaskType,
			Status:   one.Status,
			Revision: &core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &dataproto.ResFlowRelListResult{Details: details}, nil
}
