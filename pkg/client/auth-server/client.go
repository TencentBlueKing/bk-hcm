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

// Package authserver defines auth-server api client.
package authserver

import (
	"context"
	"fmt"
	"net/http"

	"hcm/cmd/auth-server/types"
	authserver "hcm/pkg/api/auth-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// Client is auth-server api client.
type Client struct {
	client rest.ClientInterface
}

// NewClient create a new auth-server api client.
func NewClient(c *client.Capability, version string) *Client {
	base := fmt.Sprintf("/api/%s/auth", version)
	return &Client{
		client: rest.NewClient(c, base),
	}
}

// InitAuthCenter init auth center's auth model.
func (c *Client) InitAuthCenter(ctx context.Context, h http.Header, request *authserver.InitAuthCenterReq) error {
	resp := new(rest.BaseResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/init/authcenter").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return err
}

// PullResource iam pull resource callback.
func (c *Client) PullResource(ctx context.Context, h http.Header, request *types.PullResourceReq) (
	interface{}, error) {

	resp := new(authserver.PullResourceResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/iam/find/resource").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// AuthorizeBatch batch authorize resource.
func (c *Client) AuthorizeBatch(ctx context.Context, h http.Header, request *authserver.AuthorizeBatchReq) (
	[]meta.Decision, error) {

	resp := new(authserver.AuthorizeBatchResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/auth/authorize/batch").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// AuthorizeAnyBatch batch authorize if resource has any permission.
func (c *Client) AuthorizeAnyBatch(ctx context.Context, h http.Header, request *authserver.AuthorizeBatchReq) (
	[]meta.Decision, error) {

	resp := new(authserver.AuthorizeBatchResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/auth/authorize/any/batch").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// GetPermissionToApply iam pull resource callback.
func (c *Client) GetPermissionToApply(ctx context.Context, h http.Header, request *authserver.GetPermissionToApplyReq) (
	*meta.IamPermission, error) {

	resp := new(authserver.GetPermissionToApplyResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/auth/find/permission_to_apply").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// ListAuthorizedInstances list authorized instances.
func (c *Client) ListAuthorizedInstances(ctx context.Context, h http.Header,
	req *authserver.ListAuthorizedInstancesReq) (*meta.AuthorizedInstances, error) {

	resp := new(authserver.ListAuthorizedInstancesResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/auth/list/authorized_resource").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// RegisterResourceCreatorAction register resource creator action instances.
func (c *Client) RegisterResourceCreatorAction(ctx context.Context, h http.Header,
	req *authserver.RegisterResourceCreatorActionReq) ([]iam.CreatorActionPolicy, error) {

	resp := new(authserver.RegisterResourceCreatorActionResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/auth/register/resource_create_action").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// GetApplyPermUrl get iam apply permission url.
func (c *Client) GetApplyPermUrl(ctx context.Context, h http.Header, req *meta.IamPermission) (string, error) {
	resp := new(authserver.GetNoAuthSkipUrlResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/auth/find/apply_perm_url").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return "", errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}
