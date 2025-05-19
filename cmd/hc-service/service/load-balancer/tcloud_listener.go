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

	loadbalancer "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// QueryListenerTargetsByCloudIDs 直接从云上查询监听器RS列表
func (svc *clbSvc) QueryListenerTargetsByCloudIDs(cts *rest.Contexts) (any, error) {

	req := new(protolb.QueryTCloudListenerTargets)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	listOpt := &loadbalancer.TCloudListTargetsOption{
		Region:         req.Region,
		LoadBalancerId: req.LoadBalancerCloudId,
		ListenerIds:    req.ListenerCloudIDs,
		Protocol:       req.Protocol,
		Port:           req.Port,
	}
	return tcloud.ListTargets(cts.Kit, listOpt)
}

// CreateTCloudListener 仅创建监听器自身
func (svc *clbSvc) CreateTCloudListener(cts *rest.Contexts) (interface{}, error) {
	req := new(protolb.TCloudListenerCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据lbID，查询负载均衡信息
	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.LbID),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, lbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, id: %s, err: %v, rid: %s", req.LbID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(lbList.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "load balancer: %s not found", req.LbID)
	}
	lbInfo := lbList.Details[0]
	if req.BkBizID == 0 {
		// 默认使用负载均衡所在业务
		req.BkBizID = lbInfo.BkBizID
	}
	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	lblOpt := &loadbalancer.TCloudCreateListenerOption{
		Region:            lbInfo.Region,
		LoadBalancerId:    lbInfo.CloudID,
		ListenerName:      req.Name,
		Protocol:          req.Protocol,
		Port:              req.Port,
		SessionExpireTime: req.SessionExpire,
		Scheduler:         req.Scheduler,
		SniSwitch:         req.SniSwitch,
		SessionType:       req.SessionType,
		Certificate:       req.Certificate,
		HealthCheck:       req.HealthCheck,
		EndPort:           uint64(cvt.PtrToVal(req.EndPort)),
	}

	result, err := tcloudAdpt.CreateListener(cts.Kit, lblOpt)
	if err != nil {
		logs.Errorf("create tcloud listener api failed, err: %v, lblOpt: %+v, cert: %+v, rid: %s",
			err, lblOpt, cvt.PtrToVal(lblOpt.Certificate), cts.Kit.Rid)
		return nil, err
	}
	if len(result.SuccessCloudIDs) == 0 {
		return nil, errors.New("create tcloud listener failed, no any listener being created")
	}
	cloudLblID := result.SuccessCloudIDs[0]

	// 插入新的监听器、规则信息到DB
	id, err := svc.createListenerDB(cts.Kit, req, lbInfo, cloudLblID)
	if err != nil {
		logs.Errorf("failed to create tcloud listener db, req: %+v, lbInfo: %+v, cloudID: %s, err: %v, rid: %s",
			req, lbInfo, cloudLblID, err, cts.Kit.Rid)
		return nil, err
	}

	return &protolb.ListenerCreateResult{CloudID: cloudLblID, ID: id}, nil
}

func (svc *clbSvc) createListenerDB(kt *kit.Kit, req *protolb.TCloudListenerCreateReq, lbInfo corelb.BaseLoadBalancer,
	cloudID string) (string, error) {

	if req.Protocol.IsLayer7Protocol() {
		// for layer 7 only create listeners itself
		lblCreateReq := &dataproto.TCloudListenerBatchCreateReq{
			Listeners: []dataproto.ListenersCreateReq[corelb.TCloudListenerExtension]{{
				CloudID:   cloudID,
				Name:      req.Name,
				Vendor:    lbInfo.Vendor,
				AccountID: lbInfo.AccountID,
				BkBizID:   lbInfo.BkBizID,
				LbID:      lbInfo.ID,
				CloudLbID: lbInfo.CloudID,
				Protocol:  req.Protocol,
				Port:      req.Port,
				Region:    lbInfo.Region,
				Extension: &corelb.TCloudListenerExtension{
					Certificate: req.Certificate,
					EndPort:     req.EndPort,
				},
			}}}
		created, err := svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudListener(kt, lblCreateReq)
		if err != nil {
			logs.Errorf("fail to create l7 listener for create listener only, err: %v, req: %+v, rid: %s",
				err, req, kt.Rid)
			return "", err
		}
		if len(created.IDs) == 0 {
			return "", errors.New("create tcloud listener db failed, no any listener being created")
		}
		return created.IDs[0], nil
	}
	// L4 create with rule
	ruleCreateReq := &dataproto.ListenerWithRuleBatchCreateReq{
		ListenerWithRules: []dataproto.ListenerWithRuleCreateReq{{
			CloudID:       cloudID,
			Name:          req.Name,
			Vendor:        lbInfo.Vendor,
			AccountID:     lbInfo.AccountID,
			BkBizID:       req.BkBizID,
			LbID:          lbInfo.ID,
			CloudLbID:     lbInfo.CloudID,
			Protocol:      req.Protocol,
			Port:          req.Port,
			CloudRuleID:   cloudID,
			Scheduler:     req.Scheduler,
			RuleType:      enumor.Layer4RuleType,
			SessionType:   cvt.PtrToVal(req.SessionType),
			SessionExpire: req.SessionExpire,
			SniSwitch:     req.SniSwitch,
			Certificate:   req.Certificate,
			Region:        lbInfo.Region,
			EndPort:       req.EndPort,
		}}}
	created, err := svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudListenerWithRule(kt, ruleCreateReq)
	if err != nil {
		logs.Errorf("fail to create l4 listener for create listener only, err: %v, req: %+v, rid: %s",
			err, req, kt.Rid)
		return "", err
	}
	if len(created.IDs) == 0 {
		return "", errors.New("create tcloud listener db failed, no any listener being created")
	}
	return created.IDs[0], nil
}
