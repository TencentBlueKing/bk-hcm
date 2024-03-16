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
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListLoadBalancer list load balancer.
func (svc *lbSvc) ListLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancer(cts, handler.ListResourceAuthRes)
}

// ListBizLoadBalancer list biz load balancer.
func (svc *lbSvc) ListBizLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancer(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listLoadBalancer(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Find,
		Filter:     req.Filter,
	})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(cts.Kit, listReq)
}

// GetLoadBalancer getLoadBalancer clb.
func (svc *lbSvc) GetLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancer(cts, handler.ListResourceAuthRes)
}

// GetBizLoadBalancer getLoadBalancer biz clb.
func (svc *lbSvc) GetBizLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancer(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) getLoadBalancer(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.LoadBalancerCloudResType,
		id)
	if err != nil {
		logs.Errorf("fail to get clb basic info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get clb")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.LoadBalancer.Get(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

// ListTargetsByTGID ...
func (svc *lbSvc) ListTargetsByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsByTGID(cts, handler.ResOperateAuth)
}

// ListBizTargetsByTGID ...
func (svc *lbSvc) ListBizTargetsByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsByTGID(cts, handler.BizOperateAuth)
}

// listTargetsByTGID 目标组下RS列表
func (svc *lbSvc) listTargetsByTGID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TargetGroup,
		Action:     meta.Find,
		BasicInfo:  basicInfo,
	})
	if err != nil {
		return nil, err
	}
	filter, err := tools.And(req.Filter, tools.RuleEqual("target_group_id", tgID))
	if err != nil {
		logs.Errorf("merge filter failed, err: %v, target_group_id: %s, rid: %s", err, tgID, cts.Kit.Rid)
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.LoadBalancer.ListTarget(cts.Kit, listReq)
}
