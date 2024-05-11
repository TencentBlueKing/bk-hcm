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

	"hcm/cmd/cloud-server/service/common"
	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// BatchCreateLB 批量创建负载均衡
func (svc *lbSvc) BatchCreateLB(cts *rest.Contexts) (any, error) {
	return svc.batchCreateLB(cts, handler.ResOperateAuth, constant.UnassignedBiz)
}

// BizBatchCreateLB 业务下直接创建 负载均衡，TODO: 用申请流程替换
func (svc *lbSvc) BizBatchCreateLB(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchCreateLB(cts, handler.BizOperateAuth, bizID)
}
func (svc *lbSvc) batchCreateLB(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler, bkBizID int64) (
	any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create clb request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 权限校验
	err := validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Create,
		BasicInfo:  common.GetCloudResourceBasicInfo(req.AccountID, bkBizID),
	})
	if err != nil {
		logs.Errorf("create load balancer auth failed, err: %v, account id: %s, bk_biz_id: %d, rid: %s",
			err, req.AccountID, bkBizID, cts.Kit.Rid)
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
		return svc.batchCreateTCloudLB(cts.Kit, req.Data, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudLB(kt *kit.Kit, rawReq json.RawMessage, bkBizID int64) (any, error) {
	req := new(cslb.TCloudBatchCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	hcReq := &hcproto.TCloudBatchCreateReq{
		BkBizID:                 bkBizID,
		AccountID:               req.AccountID,
		Region:                  req.Region,
		Name:                    req.Name,
		LoadBalancerType:        req.LoadBalancerType,
		AddressIPVersion:        req.AddressIPVersion,
		Zones:                   req.Zones,
		BackupZones:             req.BackupZones,
		CloudVpcID:              req.CloudVpcID,
		CloudSubnetID:           req.CloudSubnetID,
		Vip:                     req.Vip,
		CloudEipID:              req.CloudEipID,
		VipIsp:                  req.VipIsp,
		InternetChargeType:      req.InternetChargeType,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,
		BandwidthPackageID:      req.BandwidthPackageID,
		SlaType:                 req.SlaType,
		AutoRenew:               req.AutoRenew,
		RequireCount:            req.RequireCount,
		Memo:                    req.Memo,
	}
	return svc.client.HCService().TCloud.Clb.BatchCreate(kt, hcReq)
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
	req := new(cslb.TargetGroupCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if cvt.PtrToVal(req.HealthCheck.HealthSwitch) == 0 {
		req.HealthCheck.HealthSwitch = cvt.ValToPtr(int64(0))
	}
	healthJson, err := json.Marshal(req.HealthCheck)
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
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
				VpcID:           "",
				CloudVpcID:      req.CloudVpcID,
				TargetGroupType: enumor.LocalTargetGroupType,
				Weight:          0,
				HealthCheck:     tabletype.JsonField(healthJson),
				Memo:            nil,
				Extension:       nil,
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
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	// 预检测-检查四层监听器，绑定的目标组里面的RS，是否已绑定其他监听器
	err = svc.checkLayerFourGlobalUniqueTarget(kt, req)
	if err != nil {
		return nil, err
	}

	req.BkBizID = bkBizID
	req.LbID = lbID
	createResp, err := svc.client.HCService().TCloud.Clb.CreateListener(kt, req)
	if err != nil {
		logs.Errorf("fail to create tcloud url rule, err: %v, req: %+v, cert: %+v, rid: %s",
			err, req, cvt.PtrToVal(req.Certificate), kt.Rid)
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

// checkLayerFourGlobalUniqueTarget 检查四层监听器，绑定的目标组里面的RS，是否已绑定其他监听器
func (svc *lbSvc) checkLayerFourGlobalUniqueTarget(kt *kit.Kit, req *hcproto.ListenerWithRuleCreateReq) error {
	if req.Protocol.IsLayer7Protocol() {
		return nil
	}

	// 检查要绑定的目标组中是否有rs
	listRsReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", req.TargetGroupID),
		Page:   core.NewDefaultBasePage(),
	}
	rsResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listRsReq)
	if err != nil {
		logs.Errorf("fail to list target by target group id, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}
	if len(rsResp.Details) == 0 {
		return nil
	}

	// 查找该负载均衡下的4层监听器，绑定的所有目标组ID
	targetGroupIDs, err := svc.getBindTargetGroupIDsByLbID(kt, req)
	if err != nil {
		return err
	}
	if len(targetGroupIDs) == 0 {
		return nil
	}

	// 查找关联表中所有目标组的rs
	listRelRsReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", targetGroupIDs),
		Page:   core.NewDefaultBasePage(),
	}
	relRsResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listRelRsReq)
	if err != nil {
		logs.Errorf("fail to list target by target group ids, err: %v, tgIDs: %v, rid: %s", err, targetGroupIDs, kt.Rid)
		return err
	}

	existRsMap := make(map[string]struct{})
	for _, tgItem := range relRsResp.Details {
		uniqueKey := fmt.Sprintf("%s:%d", tgItem.CloudInstID, tgItem.Port)
		if _, exist := existRsMap[uniqueKey]; exist {
			return errf.Newf(errf.RecordDuplicated, "(vip+protocol+rsip+rsport) should be globally unique for fourth "+
				"layer listeners, targetGroupID: %s, CloudInstID: %s, PrivateIPAddress: %v, Port: %d has bind listener",
				req.TargetGroupID, tgItem.CloudInstID, tgItem.PrivateIPAddress, tgItem.Port)
		}
		existRsMap[uniqueKey] = struct{}{}
	}

	return nil
}

// getBindTargetGroupIDsByLbID 查找该负载均衡下的4层监听器，绑定的所有目标组ID
func (svc *lbSvc) getBindTargetGroupIDsByLbID(kt *kit.Kit, req *hcproto.ListenerWithRuleCreateReq) ([]string, error) {
	listTGReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", req.LbID),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus),
			tools.RuleEqual("listener_rule_type", enumor.Layer4RuleType),
		),
		Page: core.NewDefaultBasePage(),
	}
	tgResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, listTGReq)
	if err != nil {
		logs.Errorf("fail to list listener rule rel by lbid, err: %v, lbID: %s, rid: %s", err, req.LbID, kt.Rid)
		return nil, err
	}
	if len(tgResp.Details) == 0 {
		return nil, nil
	}

	lblIDs := make([]string, len(tgResp.Details))
	lblTGMap := make(map[string][]string)
	for _, item := range tgResp.Details {
		lblIDs = append(lblIDs, item.LblID)
		lblTGMap[item.LblID] = append(lblTGMap[item.LblID], item.TargetGroupID)
	}
	// 查找对应Protocol的监听器列表
	lblReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", lblIDs), tools.RuleEqual("protocol", req.Protocol)),
		Page:   core.NewDefaultBasePage(),
	}
	lblResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("fail to list listener by lblids, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
		return nil, err
	}
	if len(lblResp.Details) == 0 {
		return nil, nil
	}
	targetGroupIDs := make([]string, 0)
	for _, item := range lblResp.Details {
		tmpTGIDs, ok := lblTGMap[item.ID]
		if !ok {
			continue
		}
		targetGroupIDs = append(targetGroupIDs, tmpTGIDs...)
	}
	// 加入即将关联的目标组
	targetGroupIDs = append(targetGroupIDs, req.TargetGroupID)

	return targetGroupIDs, nil
}
