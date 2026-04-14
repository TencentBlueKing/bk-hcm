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
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewSubAccountSecretClient create a new sub account secret api client.
func NewSubAccountSecretClient(client rest.ClientInterface) *SubAccountSecretClient {
	return &SubAccountSecretClient{
		client: client,
	}
}

// SubAccountSecretClient is data service sub account secret api client.
type SubAccountSecretClient struct {
	client rest.ClientInterface
}

// BatchCreateSubAccountSecret batch create sub account secret.
func (s *SubAccountSecretClient) BatchCreateSubAccountSecret(kt *kit.Kit,
	req *protocloud.SubAccountSecretBatchCreateReq[coresass.TCloudSubAccountSecretExtension]) (
	*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := s.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_account_secrets/batch/create").
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

// BatchUpdateSubAccountSecret batch update sub account secret.
func (s *SubAccountSecretClient) BatchUpdateSubAccountSecret(kt *kit.Kit,
	req *protocloud.SubAccountSecretBatchUpdateReq[coresass.TCloudSubAccountSecretExtension]) error {

	resp := new(rest.BaseResp)

	err := s.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_account_secrets/batch/update").
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

// ListSubAccountSecretWithExtension list sub account secret with extension.
func (s *SubAccountSecretClient) ListSubAccountSecretWithExtension(kt *kit.Kit,
	req *protocloud.SubAccountSecretExtListReq) (
	*protocloud.SubAccountSecretExtListResult[coresass.TCloudSubAccountSecretExtension], error) {

	resp := new(protocloud.SubAccountSecretExtListResp[coresass.TCloudSubAccountSecretExtension])

	err := s.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_account_secrets/extensions/list").
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
