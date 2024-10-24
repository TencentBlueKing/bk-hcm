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
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// ApprovalProcessClient is data service approval process api client.
type ApprovalProcessClient struct {
	client rest.ClientInterface
}

// NewApprovalProcessClient create a new approval process api client.
func NewApprovalProcessClient(client rest.ClientInterface) *ApprovalProcessClient {
	return &ApprovalProcessClient{
		client: client,
	}
}

// CreateApprovalProcesses ...
func (a *ApprovalProcessClient) CreateApprovalProcesses(ctx context.Context, h http.Header, request *proto.ApplicationCreateReq) (
	*core.CreateResult, error,
) {
	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/approval_processes/create").
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

// UpdateApprovalProcesses ...
func (a *ApprovalProcessClient) UpdateApprovalProcesses(ctx context.Context, h http.Header,
	approvalProcessID string, request *proto.ApprovalProcessUpdateReq) (
	interface{}, error,
) {
	resp := new(core.UpdateResp)

	err := a.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/approval_processes/%s", approvalProcessID).
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

// ListApprovalProcesses ...
func (a *ApprovalProcessClient) ListApprovalProcesses(ctx context.Context, h http.Header, request *proto.ApprovalProcessListReq) (
	*proto.ApprovalProcessListResult, error,
) {
	resp := new(proto.ApprovalProcessListResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/approval_processes/list").
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
