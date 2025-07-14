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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchCreateTargetGroupWithRel 批量创建目标组，并绑定监听器/规则
func (svc *lbSvc) BatchCreateTargetGroupWithRel(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateTargetGroupWithRel[T corelb.TargetGroupExtension](cts *rest.Contexts,
	svc *lbSvc, vendor enumor.Vendor) (any, error) {

	req := new(dataproto.BatchCreateTgWithRelReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcCloudIDs := slice.Map(req.TargetGroups,
		func(g dataproto.CreateTargetGroupWithRel[T]) string { return g.TargetGroup.CloudVpcID })

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		vpcInfoMap, err := getVpcMapByIDs(cts.Kit, vpcCloudIDs)
		if err != nil {
			return nil, err
		}

		tgIDs := make([]string, 0, len(req.TargetGroups))
		for _, tgReq := range req.TargetGroups {
			// 创建目标组
			tgTable, err := convTargetGroupCreateReqToTable(cts.Kit, vendor, tgReq.TargetGroup, vpcInfoMap)
			if err != nil {
				return nil, err
			}

			models := []*tablelb.LoadBalancerTargetGroupTable{tgTable}
			createTGIDs, err := svc.dao.LoadBalancerTargetGroup().BatchCreateWithTx(cts.Kit, txn, models)
			if err != nil {
				logs.Errorf("[%s]fail to batch create target group for create tg, err: %v, rid:%s",
					vendor, err, cts.Kit.Rid)
				return nil, fmt.Errorf("batch create target group failed, err: %v", err)
			}
			tgIDs = append(tgIDs, createTGIDs...)
			tgID := createTGIDs[0]
			// 添加RS
			if len(tgReq.TargetGroup.RsList) != 0 {
				_, err = svc.batchCreateTargetWithGroupID(cts.Kit, txn, "", tgID, tgReq.TargetGroup.RsList)
				if err != nil {
					logs.Errorf("fail to batch create target for create tg, err: %v, rid:%s", err, cts.Kit.Rid)
					return nil, fmt.Errorf("batch create target failed, err: %v", err)
				}
			}
			// 创建绑定关系
			if _, err := createRel(cts.Kit, svc, txn, tgReq, tgID, vendor); err != nil {
				logs.Errorf("fail to batch create listener rule rel for create tg, err: %v, rid:%s",
					err, cts.Kit.Rid)
				return nil, fmt.Errorf("batch create listener rule rel failed, err: %v", err)
			}
		}

		return tgIDs, nil
	})
	if err != nil {
		return nil, err
	}
	return &core.BatchCreateResult{IDs: result.([]string)}, nil
}

func createRel[T corelb.TargetGroupExtension](kt *kit.Kit, svc *lbSvc, txn *sqlx.Tx,
	tgReq dataproto.CreateTargetGroupWithRel[T], tgID string, vendor enumor.Vendor) ([]string, error) {

	// 创建关系
	ruleRelModels := []*tablelb.TargetGroupListenerRuleRelTable{{
		Vendor:              vendor,
		ListenerRuleID:      tgReq.ListenerRuleID,
		CloudListenerRuleID: tgReq.CloudListenerRuleID,
		ListenerRuleType:    tgReq.ListenerRuleType,
		TargetGroupID:       tgID,
		CloudTargetGroupID:  tgID,
		LbID:                tgReq.LbID,
		CloudLbID:           tgReq.CloudLbID,
		LblID:               tgReq.LblID,
		CloudLblID:          tgReq.CloudLblID,
		BindingStatus:       tgReq.BindingStatus,
		Detail:              tgReq.Detail,
		Creator:             kt.User,
		Reviser:             kt.User,
	}}
	switch vendor {
	case enumor.TCloud:
		// 更新规则表
		rule := &tablelb.TCloudLbUrlRuleTable{
			TargetGroupID:      tgID,
			CloudTargetGroupID: tgID,
			Reviser:            kt.User,
		}
		err := svc.dao.LoadBalancerTCloudUrlRule().UpdateByIDWithTx(kt, txn, tgReq.ListenerRuleID, rule)
		if err != nil {
			logs.Errorf("fail to update rule while creating target group with rel, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	tx, err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(kt, txn, ruleRelModels)
	if err != nil {
		logs.Errorf("fail to create tg rel while creating target group with rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return tx, err
}

// BatchDeleteTarget 批量删除本地RS
func (svc *lbSvc) BatchDeleteTarget(cts *rest.Contexts) (any, error) {

	req := new(dataproto.LoadBalancerBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor", "cloud_inst_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.LoadBalancerTarget().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list target for deletion failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list target failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	targetIds := slice.Map(listResp.Details, func(one tablelb.LoadBalancerTargetTable) string { return one.ID })
	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		delFilter := tools.ContainersExpression("id", targetIds)
		return nil, svc.dao.LoadBalancerTarget().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete target(ids=%v) failed, err: %v, rid: %s", targetIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchUpdateTarget 批量更新RS
func (svc *lbSvc) BatchUpdateTarget(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TargetBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, target := range req.Targets {
			update := &tablelb.LoadBalancerTargetTable{
				InstName:          target.InstName,
				TargetGroupRegion: target.TargetGroupRegion,
				Port:              target.Port,
				Weight:            target.Weight,
				PrivateIPAddress:  target.PrivateIPAddress,
				PublicIPAddress:   target.PublicIPAddress,
				Memo:              target.Memo,
				Reviser:           cts.Kit.User,
			}

			if err := svc.dao.LoadBalancerTarget().UpdateByIDWithTx(cts.Kit, txn, target.ID, update); err != nil {
				logs.Errorf("update tcloud target by id failed, err: %v, id: %s, rid: %s", err, target.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update target failed, err: %v", err)
			}
		}

		return nil, nil
	})
}
