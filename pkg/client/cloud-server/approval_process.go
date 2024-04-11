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

package cloudserver

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ApprovalProcessClient is approval process client.
type ApprovalProcessClient struct {
	client rest.ClientInterface
}

// NewApprovalProcessClient create a new approval process client.
func NewApprovalProcessClient(client rest.ClientInterface) *ApprovalProcessClient {
	return &ApprovalProcessClient{
		client: client,
	}
}

// GetApprovalProcessServiceID ...
func (cli *ApprovalProcessClient) GetApprovalProcessServiceID(kt *kit.Kit) ([]int64, error) {

	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          []int64 `json:"data"`
	}{}

	err := cli.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/approval_processes/service_id").
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
