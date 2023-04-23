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
	hcproto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// GetAccountZoneQuota 获取腾讯云账号配额.
func (a *accountSvc) GetAccountZoneQuota(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.GetAccountZoneQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Quota, Action: meta.Find}, BizID: bizID}
	if err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
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

// GetHuaWeiRegionQuota 获取华为云账号配额.
func (a *accountSvc) GetHuaWeiRegionQuota(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.GetAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Quota, Action: meta.Find}, BizID: bizID}
	if err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	getReq := &hcproto.GetHuaWeiAccountRegionQuotaReq{
		AccountID: accountID,
		Region:    req.Region,
	}
	return a.client.HCService().HuaWei.Account.GetRegionQuota(cts.Kit.Ctx, cts.Kit.Header(), getReq)
}

// GetGcpRegionQuota 获取Gcp账号配额.
func (a *accountSvc) GetGcpRegionQuota(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.GetAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Quota, Action: meta.Find}, BizID: bizID}
	if err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
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
