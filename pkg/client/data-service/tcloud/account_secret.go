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

package tcloud

import (
	"hcm/pkg/api/core"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewAccountSecretClient create a new account secret api client.
func NewAccountSecretClient(client rest.ClientInterface) *AccountSecretClient {
	return &AccountSecretClient{
		client: client,
	}
}

// AccountSecretClient is data service account secret api client.
type AccountSecretClient struct {
	client rest.ClientInterface
}

// BatchCreateAccountSecret batch create account secret.
func (a *AccountSecretClient) BatchCreateAccountSecret(kt *kit.Kit,
	req *protocloud.AccountSecretBatchCreateReq[coreas.TCloudAccountSecretExtension]) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/account_secrets/batch/create").
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

// BatchUpdateAccountSecret batch update account secret.
func (a *AccountSecretClient) BatchUpdateAccountSecret(kt *kit.Kit,
	req *protocloud.AccountSecretBatchUpdateReq[coreas.TCloudAccountSecretExtension]) error {

	resp := new(rest.BaseResp)

	err := a.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/account_secrets/batch/update").
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

// ListAccountSecretWithExtension list account secret with extension.
func (a *AccountSecretClient) ListAccountSecretWithExtension(kt *kit.Kit, req *protocloud.AccountSecretExtListReq) (
	*protocloud.AccountSecretExtListResult[coreas.TCloudAccountSecretExtension], error) {

	resp := new(protocloud.AccountSecretExtListResp[coreas.TCloudAccountSecretExtension])

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/account_secrets/extensions/list").
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
