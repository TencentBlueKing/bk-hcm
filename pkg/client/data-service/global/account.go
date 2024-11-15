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

package global

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountClient is data service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// List ...
func (a *AccountClient) List(ctx context.Context, h http.Header, request *protocloud.AccountListReq) (
	*protocloud.AccountListResult, error,
) {
	resp := new(protocloud.AccountListResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/list").
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

// Delete ...
func (a *AccountClient) Delete(ctx context.Context, h http.Header, request *protocloud.AccountDeleteReq) (
	interface{}, error,
) {
	resp := new(core.DeleteResp)

	err := a.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts").
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

// DeleteValidate ...
func (a *AccountClient) DeleteValidate(ctx context.Context, h http.Header, accountID string) (
	map[string]uint64, error) {

	resp := new(protocloud.AccountDeleteValidateResp)

	err := a.client.Post().
		WithContext(ctx).
		SubResourcef("/accounts/%s/delete/validate", accountID).
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return resp.Data, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// UpdateBizRel update account and biz rel api, only support put all biz for account_id.
func (a *AccountClient) UpdateBizRel(ctx context.Context, h http.Header, accountID string,
	request *protocloud.AccountBizRelUpdateReq) (interface{}, error) {

	resp := new(core.UpdateResp)

	err := a.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/account_biz_rels/accounts/%s", accountID).
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

// ListWithExtension ...
func (a *AccountClient) ListWithExtension(ctx context.Context, h http.Header, request *protocloud.AccountListReq) (
	*protocloud.AccountWithExtensionListResult, error,
) {
	resp := new(protocloud.AccountWithExtensionListResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/extensions/list").
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

// ListAccountBizRelWithAccount ...
func (a *AccountClient) ListAccountBizRelWithAccount(kt *kit.Kit, request *protocloud.AccountBizRelWithAccountListReq) (
	[]*protocloud.AccountBizRelWithAccount, error) {

	resp := new(protocloud.AccountBizRelWithAccountListResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/account_biz_rels/with/accounts/list").
		WithHeaders(kt.Header()).
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

// ListAccountBizRel list account biz relation.
func (a *AccountClient) ListAccountBizRel(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.AccountBizRelListResult, error) {

	resp := new(protocloud.AccountBizRelListResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/account_biz_rels/list").
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
