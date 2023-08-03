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

package gcp

import (
	"context"
	"net/http"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/hc-service"
	protoaccount "hcm/pkg/api/hc-service/account"
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
func (a *AccountClient) SyncSubAccount(kt *kit.Kit, req *sync.GcpGlobalSyncReq) error {

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

// Check account
func (a *AccountClient) Check(ctx context.Context, h http.Header, request *hcservice.GcpAccountCheckReq) error {

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
	request *protoaccount.GetGcpAccountRegionQuotaReq) (*typeaccount.GcpProjectQuota, error) {

	resp := new(protoaccount.GetGcpAccountQuotaResp)

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
