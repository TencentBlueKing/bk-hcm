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
	"hcm/pkg/api/core"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewSubAccountClient create a new sub_account api client.
func NewSubAccountClient(client rest.ClientInterface) *SubAccountClient {
	return &SubAccountClient{
		client: client,
	}
}

// SubAccountClient is data service sub_account api client.
type SubAccountClient struct {
	client rest.ClientInterface
}

// Get get sub_account.
func (cli *SubAccountClient) Get(kt *kit.Kit, id string) (
	*coresubaccount.SubAccount[coresubaccount.HuaWeiExtension], error) {

	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          *coresubaccount.SubAccount[coresubaccount.HuaWeiExtension] `json:"data"`
	}{}

	err := cli.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/sub_accounts/%s", id).
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

// ListExt list sub_account with extension.
func (cli *SubAccountClient) ListExt(kt *kit.Kit, request *core.ListReq) (
	*dssubaccount.ListExtResult[coresubaccount.HuaWeiExtension], error) {

	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          *dssubaccount.ListExtResult[coresubaccount.HuaWeiExtension] `json:"data"`
	}{}

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/sub_accounts/list").
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
