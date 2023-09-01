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

package huawei

import (
	"context"
	"net/http"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountClient is hc service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// SyncSubAccount sync sub account
func (a *AccountClient) SyncSubAccount(kt *kit.Kit, req *sync.HuaWeiGlobalSyncReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_accounts/sync").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// Check 联通性和云上字段匹配校验
func (a *AccountClient) Check(ctx context.Context, h http.Header, request *hsaccount.HuaWeiAccountCheckReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/check").
		WithHeaders(h).
		Do().
		Into(resp)

	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// GetRegionQuota get account region quota.
func (a *AccountClient) GetRegionQuota(ctx context.Context, h http.Header,
	request *hsaccount.GetHuaWeiAccountRegionQuotaReq) (*typeaccount.HuaWeiAccountQuota, error) {

	resp := new(hsaccount.GetHuaWeiAccountQuotaResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/regions/quotas").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetBySecret get account info by secret
func (a *AccountClient) GetBySecret(ctx context.Context, h http.Header,
	request *cloud.HuaWeiSecret) (*cloud.HuaWeiInfoBySecret, error) {

	resp := new(account.BySecretResp[cloud.HuaWeiInfoBySecret])
	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/secret").
		WithHeaders(h).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil

}

// GetResCountBySecret get account res count by secret
func (a *AccountClient) GetResCountBySecret(ctx context.Context, h http.Header,
	request *cloud.HuaWeiSecret) (*hsaccount.ResCount, error) {

	resp := new(hsaccount.ResCountResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/res_counts/by_secrets").
		WithHeaders(h).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}
