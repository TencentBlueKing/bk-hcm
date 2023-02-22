/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package iam defines esb client to request iam.
package iam

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
	"hcm/pkg/tools/uuid"
)

// Client is an esb client to request iam.
type Client interface {
	RegisterResourceCreatorAction(ctx context.Context, inst *client.InstanceWithCreator) ([]client.CreatorActionPolicy,
		error)
	GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error)
}

// NewClient initialize a new iam client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &iam{
		client: client,
		config: config,
	}
}

// iam is an esb client to request iam.
type iam struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// RegisterResourceCreatorAction register iam resource instance with creator
// returns related actions with policy id that the creator gained
func (i *iam) RegisterResourceCreatorAction(ctx context.Context, inst *client.InstanceWithCreator) (
	[]client.CreatorActionPolicy, error) {

	resp := new(esbIamCreatorActionResp)
	req := &esbInstanceWithCreatorParams{
		CommParams:          types.GetCommParams(i.config),
		InstanceWithCreator: inst,
	}
	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())

	err := i.client.Post().
		SubResourcef("/iam/authorization/resource_creator_action/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("register iam resource creator instance failed, code: %d, msg: %s, rid: %s", resp.Code,
			resp.Message, resp.Rid)
	}

	return resp.Data, nil
}

// GetApplyPermUrl get iam apply permission url.
func (i *iam) GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error) {
	resp := new(esbIamGetApplyPermUrlResp)
	req := &esbIamGetApplyPermUrlParams{
		CommParams:    types.GetCommParams(i.config),
		IamPermission: opts,
	}

	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())

	err := i.client.Post().
		SubResourcef("/iam/application/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return "", err
	}

	if !resp.Result || resp.Code != 0 {
		return "", fmt.Errorf("get iam apply permission url failed, code: %d, msg: %s, rid: %s", resp.Code,
			resp.Message, resp.Rid)
	}

	return resp.Data.Url, nil
}
