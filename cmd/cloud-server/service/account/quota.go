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

package account

import (
	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// GetBizTCloudZoneQuota 获取腾讯云账号配额.
func (a *accountSvc) GetBizTCloudZoneQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getTCloudZoneQuota(cts, handler.BizValidWithAuth)
}

// GetResTCloudZoneQuota 获取腾讯云账号配额.
func (a *accountSvc) GetResTCloudZoneQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getTCloudZoneQuota(cts, handler.ResValidWithAuth)
}

func (a *accountSvc) getTCloudZoneQuota(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	req := new(proto.GetAccountZoneQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: accountID,
	}
	// validate biz and authorize
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: a.authorizer, ResType: meta.Quota,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	getReq := &hcproto.GetTCloudAccountZoneQuotaReq{
		AccountID: accountID,
		Region:    req.Region,
		Zone:      req.Zone,
	}
	return a.client.HCService().TCloud.Account.GetZoneQuota(cts.Kit.Ctx, cts.Kit.Header(), getReq)
}

// GetBizHuaWeiRegionQuota 获取华为云账号配额.
func (a *accountSvc) GetBizHuaWeiRegionQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getHuaWeiRegionQuota(cts, handler.BizValidWithAuth)
}

// GetResHuaWeiRegionQuota 获取华为云账号配额.
func (a *accountSvc) GetResHuaWeiRegionQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getHuaWeiRegionQuota(cts, handler.ResValidWithAuth)
}

// getHuaWeiRegionQuota 获取华为云账号配额.
func (a *accountSvc) getHuaWeiRegionQuota(cts *rest.Contexts,
	authHandler handler.ValidWithAuthHandler) (interface{}, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	req := new(proto.GetAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: accountID,
	}
	// validate biz and authorize
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: a.authorizer, ResType: meta.Quota,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	getReq := &hcproto.GetHuaWeiAccountRegionQuotaReq{
		AccountID: accountID,
		Region:    req.Region,
	}
	getResult, err := a.client.HCService().HuaWei.Account.GetRegionQuota(cts.Kit.Ctx, cts.Kit.Header(), getReq)
	if err != nil {
		logs.Errorf("call hcservice get huawei region quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ruleMap := map[string]interface{}{
		"account_id": accountID,
		"region":     req.Region,
	}
	listReq := &protocloud.CvmListReq{
		Filter: tools.EqualWithOpExpression(filter.And, ruleMap),
		Page:   core.NewCountPage(),
	}
	resp, err := a.client.DataService().Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("call dataservice to list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
	}

	result := &proto.HuaWeiGetAccountRegionQuotaResult{
		HuaWeiAccountQuota: *getResult,
		UsedInstances:      int32(resp.Count),
	}

	return result, nil
}

// GetBizGcpRegionQuota 获取Gcp账号配额.
func (a *accountSvc) GetBizGcpRegionQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getGcpRegionQuota(cts, handler.BizValidWithAuth)
}

// GetResGcpRegionQuota 获取Gcp账号配额.
func (a *accountSvc) GetResGcpRegionQuota(cts *rest.Contexts) (interface{}, error) {
	return a.getGcpRegionQuota(cts, handler.ResValidWithAuth)
}

// getGcpRegionQuota 获取Gcp账号配额.
func (a *accountSvc) getGcpRegionQuota(cts *rest.Contexts,
	authHandler handler.ValidWithAuthHandler) (interface{}, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	req := new(proto.GetAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: accountID,
	}
	// validate biz and authorize
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: a.authorizer, ResType: meta.Quota,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	getReq := &hcproto.GetGcpAccountRegionQuotaReq{
		AccountID: accountID,
		Region:    req.Region,
	}
	return a.client.HCService().Gcp.Account.GetRegionQuota(cts.Kit.Ctx, cts.Kit.Header(), getReq)
}
