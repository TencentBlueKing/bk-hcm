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

package cloud

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// AccountClient is data service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewCloudAccountClient create a new account api client.
func NewCloudAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// CreateTCloud account.
func (a *AccountClient) CreateTCloud(ctx context.Context, h http.Header,
	request *protocloud.AccountCreateReq[protocloud.TCloudAccountExtensionCreateReq]) (
	*core.CreateResult, error,
) {
	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/vendors/tcloud/accounts/create").
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

// UpdateAws ...
func (a *AccountClient) UpdateAws(ctx context.Context, h http.Header, accountID uint64,
	request *protocloud.AccountUpdateReq[protocloud.AwsAccountExtensionUpdateReq]) (
	interface{}, error,
) {
	resp := new(core.UpdateResp)

	err := a.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/vendors/aws/accounts/%d", accountID).
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

// -------------------------- Get --------------------------

// GetTCloudAccount get tcloud account detail.
func (a *AccountClient) GetTCloudAccount(ctx context.Context, h http.Header, id string) (
	*protocloud.AccountGetResult[protocore.TCloudAccountExtension], error) {

	return getAccount[protocore.TCloudAccountExtension](ctx, h, a, id, enumor.TCloud)
}

// GetAwsAccount get aws account detail.
func (a *AccountClient) GetAwsAccount(ctx context.Context, h http.Header, id string) (
	*protocloud.AccountGetResult[protocore.AwsAccountExtension], error) {

	return getAccount[protocore.AwsAccountExtension](ctx, h, a, id, enumor.Aws)
}

// GetHuaWeiAccount get huawei account detail.
func (a *AccountClient) GetHuaWeiAccount(ctx context.Context, h http.Header, id string) (
	*protocloud.AccountGetResult[protocore.HuaWeiAccountExtension], error) {

	return getAccount[protocore.HuaWeiAccountExtension](ctx, h, a, id, enumor.HuaWei)
}

// GetAzureAccount get azure account detail.
func (a *AccountClient) GetAzureAccount(ctx context.Context, h http.Header, id string) (
	*protocloud.AccountGetResult[protocore.AzureAccountExtension], error) {

	return getAccount[protocore.AzureAccountExtension](ctx, h, a, id, enumor.Azure)
}

// GetGcpAccount get gcp account detail.
func (a *AccountClient) GetGcpAccount(ctx context.Context, h http.Header, id string) (
	*protocloud.AccountGetResult[protocore.GcpAccountExtension], error) {

	return getAccount[protocore.GcpAccountExtension](ctx, h, a, id, enumor.Gcp)
}

func getAccount[T protocloud.AccountExtensionGetResp](ctx context.Context, h http.Header, cli *AccountClient,
	id string, vendor enumor.Vendor) (*protocloud.AccountGetResult[T], error) {

	resp := new(protocloud.AccountGetResp[T])

	err := cli.client.Get().
		WithContext(ctx).
		SubResourcef("vendors/%s/accounts/%s", vendor, id).
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
		SubResourcef("/accounts/delete").
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

// UpdateAccountBizRel update account and biz rel api, only support put all biz for account_id.
func (a *AccountClient) UpdateAccountBizRel(ctx context.Context, h http.Header, accountID uint64,
	request *protocloud.AccountBizRelUpdateReq) (interface{}, error) {

	resp := new(core.UpdateResp)

	err := a.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/account_biz_rels/accounts/%d", accountID).
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
