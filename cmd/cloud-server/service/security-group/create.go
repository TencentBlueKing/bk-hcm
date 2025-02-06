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

package securitygroup

import (
	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/hooks/handler"
)

// CreateSecurityGroup create security group.
func (svc *securityGroupSvc) CreateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	bizID := int64(constant.UnassignedBiz)
	return svc.createSecurityGroup(cts, bizID, handler.ResOperateAuth)
}

// CreateBizSecurityGroup create biz security group.
func (svc *securityGroupSvc) CreateBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createSecurityGroup(cts, bizID, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) createSecurityGroup(cts *rest.Contexts, bizID int64,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	req := new(proto.SecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if bizID > 0 {
		// ensure usage biz id is empty or contains only current biz id
		if len(req.UsageBizIds) == 0 {
			// prepend current biz id to usage biz ids list
			req.UsageBizIds = []int64{bizID}
		}
		if bizID != req.UsageBizIds[0] || len(req.UsageBizIds) > 1 {
			return nil, errf.New(errf.InvalidParameter, "usage biz id can only be current biz")
		}
	}
	// check is biz out of account biz scope
	if err := svc.checkAccountBizScope(cts.Kit, req.AccountID, req.UsageBizIds); err != nil {
		return nil, err
	}

	// validate  authorize
	err := validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Create, BasicInfo: common.GetCloudResourceBasicInfo(req.AccountID, bizID)})
	if err != nil {
		return nil, err
	}

	switch req.Vendor {
	case enumor.TCloud:
		return svc.createTCloudSecurityGroup(cts, bizID, req)
	case enumor.Aws:
		return svc.createAwsSecurityGroup(cts, bizID, req)
	case enumor.HuaWei:
		return svc.createHuaWeiSecurityGroup(cts, bizID, req)
	case enumor.Azure:
		return svc.createAzureSecurityGroup(cts, bizID, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", req.Vendor)
	}
}

func (svc *securityGroupSvc) createTCloudSecurityGroup(cts *rest.Contexts, bizID int64,
	req *proto.SecurityGroupCreateReq) (interface{}, error) {

	createReq := &hcproto.TCloudSecurityGroupCreateReq{
		Region:      req.Region,
		Name:        req.Name,
		Memo:        req.Memo,
		AccountID:   req.AccountID,
		BkBizID:     bizID,
		Tags:        req.Tags,
		Manager:     req.Manager,
		BakManager:  req.BakManager,
		UsageBizIds: req.UsageBizIds,
	}
	result, err := svc.client.HCService().TCloud.SecurityGroup.CreateSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("create tcloud security group failed, err: %v, req: %v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *securityGroupSvc) createAwsSecurityGroup(cts *rest.Contexts, bizID int64,
	req *proto.SecurityGroupCreateReq) (interface{}, error) {

	extension := new(proto.AwsSecurityGroupExtensionCreate)
	if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
		return nil, err
	}

	createReq := &hcproto.AwsSecurityGroupCreateReq{
		Region:     req.Region,
		Name:       req.Name,
		Memo:       req.Memo,
		AccountID:  req.AccountID,
		BkBizID:    bizID,
		CloudVpcID: extension.CloudVpcID,
		// Tags:        req.Tags,
		Manager:     req.Manager,
		BakManager:  req.BakManager,
		UsageBizIds: req.UsageBizIds,
	}
	result, err := svc.client.HCService().Aws.SecurityGroup.CreateSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("create aws security group failed, err: %v, req: %v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *securityGroupSvc) createHuaWeiSecurityGroup(cts *rest.Contexts, bizID int64,
	req *proto.SecurityGroupCreateReq) (interface{}, error) {

	createReq := &hcproto.HuaWeiSecurityGroupCreateReq{
		Region:    req.Region,
		Name:      req.Name,
		Memo:      req.Memo,
		AccountID: req.AccountID,
		BkBizID:   bizID,
		// Tags:        req.Tags,
		Manager:     req.Manager,
		BakManager:  req.BakManager,
		UsageBizIds: req.UsageBizIds,
	}
	result, err := svc.client.HCService().HuaWei.SecurityGroup.CreateSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("create huawei security group failed, err: %v, req: %v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *securityGroupSvc) createAzureSecurityGroup(cts *rest.Contexts, bizID int64,
	req *proto.SecurityGroupCreateReq) (interface{}, error) {

	extension := new(proto.AzureSecurityGroupExtensionCreate)
	if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
		return nil, err
	}

	// Check Azure's SecurityGroup Params
	if err := svc.checkAzureSGParams(req, extension.ResourceGroupName); err != nil {
		return nil, err
	}

	createReq := &hcproto.AzureSecurityGroupCreateReq{
		Region:            req.Region,
		Name:              req.Name,
		Memo:              req.Memo,
		AccountID:         req.AccountID,
		BkBizID:           bizID,
		ResourceGroupName: extension.ResourceGroupName,
		// Tags:        req.Tags,
		Manager:     req.Manager,
		BakManager:  req.BakManager,
		UsageBizIds: req.UsageBizIds,
	}
	result, err := svc.client.HCService().Azure.SecurityGroup.CreateSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("create azure security group failed, err: %v, req: %v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// checkAzureSGParams check azure security group params
func (svc *securityGroupSvc) checkAzureSGParams(req *proto.SecurityGroupCreateReq, resGroupName string) error {
	if !assert.IsSameCaseNoSpaceString(req.Region) {
		return errf.New(errf.InvalidParameter, "region can only be lowercase")
	}

	if !assert.IsSameCaseString(req.Name) {
		return errf.New(errf.InvalidParameter, "name can only be lowercase")
	}

	if !assert.IsSameCaseString(resGroupName) {
		return errf.New(errf.InvalidParameter, "resource_group_name can only be lowercase")
	}

	return nil
}

// check given bizIDs in given account biz scope, return error if any given bizIDs not in account biz scope
func (svc *securityGroupSvc) checkAccountBizScope(kt *kit.Kit, accountID string, bizIDs []int64) error {

	reqBizMap := make(map[int64]struct{}, len(bizIDs))
	for _, bizID := range bizIDs {
		reqBizMap[bizID] = struct{}{}
	}

	bizRelReq := &core.ListReq{
		Filter: tools.EqualExpression("account_id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		relResp, err := svc.client.DataService().Global.Account.ListAccountBizRel(kt.Ctx, kt.Header(), bizRelReq)
		if err != nil {
			return err
		}
		for i := range relResp.Details {
			rel := relResp.Details[i]
			if rel.BkBizID == constant.UnassignedBiz {
				// for -1 means no biz restriction, should match any given usage biz
				return nil
			}
			if _, ok := reqBizMap[rel.BkBizID]; ok {
				delete(reqBizMap, rel.BkBizID)
			}
			if len(reqBizMap) == 0 {
				// all requested biz in account's biz scope, return no error
				return nil
			}
		}

		if uint(len(relResp.Details)) < bizRelReq.Page.Limit {
			break
		}
	}
	if len(reqBizMap) > 0 {
		return errf.Newf(errf.InvalidParameter, "some biz not in account %s's biz scope", accountID)
	}
	return nil
}
