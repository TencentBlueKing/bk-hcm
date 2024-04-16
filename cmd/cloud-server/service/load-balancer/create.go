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
	"encoding/json"
	"errors"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchCreateLB 批量创建负载均衡
func (svc *lbSvc) BatchCreateLB(cts *rest.Contexts) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create clb request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type:       meta.LoadBalancer,
		Action:     meta.Create,
		ResourceID: req.AccountID,
	}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create load balancer auth failed, err: %v, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.
		GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudLB(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudLB(kt *kit.Kit, rawReq json.RawMessage) (any, error) {
	req := new(hcproto.TCloudBatchCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.client.HCService().TCloud.Clb.BatchCreate(kt, req)
}

// CreateBizTargetGroup create biz target group.
func (svc *lbSvc) CreateBizTargetGroup(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createBizTargetGroup(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) createBizTargetGroup(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler,
	bkBizID int64) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create target group request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Create, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("create target group auth failed, err: %v, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudTargetGroup(cts.Kit, req.Data, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudTargetGroup(kt *kit.Kit, rawReq json.RawMessage, bkBizID int64) (any, error) {
	req := new(dataproto.TargetGroupCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &dataproto.TCloudTargetGroupCreateReq{
		TargetGroups: []dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			{
				Name:            req.Name,
				Vendor:          enumor.TCloud,
				AccountID:       req.AccountID,
				BkBizID:         bkBizID,
				Region:          req.Region,
				Protocol:        req.Protocol,
				Port:            req.Port,
				CloudVpcID:      req.CloudVpcID,
				TargetGroupType: enumor.LocalTargetGroupType,
				RsList:          req.RsList,
			},
		},
	}
	return svc.client.DataService().TCloud.LoadBalancer.BatchCreateTCloudTargetGroup(kt, opt)
}

// CreateBizListener create biz listener.
func (svc *lbSvc) CreateBizListener(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createListener(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) createListener(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler,
	bkBizID int64) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create listener request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Create, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("create listener auth failed, err: %v, account id: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudListener(cts.Kit, req.Data, bkBizID, lbID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudListener(kt *kit.Kit, rawReq json.RawMessage, bkBizID int64,
	lbID string) (any, error) {

	req := new(hcproto.ListenerWithRuleCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 预检测-是否有执行中的负载均衡
	err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	req.BkBizID = bkBizID
	req.LbID = lbID
	createResp, err := svc.client.HCService().TCloud.Clb.CreateListener(kt, req)
	if err != nil {
		logs.Errorf("fail to create tcloud url rule, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(createResp.CloudLblID) == 0 {
		logs.Errorf("no listener have been created, lbID: %s, req: %+v, rid: %s", lbID, req, kt.Rid)
		return nil, errors.New("create listener failed")
	}

	// 构建异步任务将目标组中的RS绑定到对应规则上
	lblInfo := &corelb.BaseListener{CloudID: createResp.CloudLblID, Protocol: req.Protocol, LbID: req.LbID}
	err = svc.applyTargetToRule(kt, req.TargetGroupID, createResp.CloudRuleID, lblInfo)
	if err != nil {
		logs.Errorf("fail to bind listener and target group register flow, err: %v, req: %+v, createResp: %+v, rid: %s",
			err, req, createResp, kt.Rid)
		return nil, err
	}
	return &core.BatchCreateResult{IDs: []string{createResp.CloudLblID}}, nil
}
