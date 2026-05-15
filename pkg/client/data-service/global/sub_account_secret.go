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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// SubAccountSecretClient is data service sub account secret api client.
type SubAccountSecretClient struct {
	client rest.ClientInterface
}

// NewSubAccountSecretClient create a new sub account secret api client.
func NewSubAccountSecretClient(client rest.ClientInterface) *SubAccountSecretClient {
	return &SubAccountSecretClient{
		client: client,
	}
}

// BatchDelete batch delete sub account secret.
func (a *SubAccountSecretClient) BatchDelete(kt *kit.Kit, req *protocloud.SubAccountSecretBatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := a.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_account_secrets/batch").
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

// ListSubAccountSecret list sub account secret.
func (a *SubAccountSecretClient) ListSubAccountSecret(kt *kit.Kit, req *protocloud.SubAccountSecretListReq) (
	*protocloud.SubAccountSecretListResult, error) {

	resp := new(protocloud.SubAccountSecretListResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_account_secrets/list").
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
