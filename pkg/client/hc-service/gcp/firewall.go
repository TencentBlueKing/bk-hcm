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

package gcp

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewFirewallClient create a new firewall api client.
func NewFirewallClient(client rest.ClientInterface) *FirewallClient {
	return &FirewallClient{
		client: client,
	}
}

// FirewallClient is data service firewall api client.
type FirewallClient struct {
	client rest.ClientInterface
}

// SyncFirewall sync fire wall.
func (cli *FirewallClient) SyncFirewall(ctx context.Context, h http.Header,
	request *sync.GcpGlobalRegionResSyncReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/firewalls/rules/sync").
		WithHeaders(h).
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

// CreateFirewallRule create firewall rule.
func (cli *FirewallClient) CreateFirewallRule(ctx context.Context, h http.Header,
	request *proto.GcpFirewallRuleCreateReq) (*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/firewalls/rules/create").
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

// UpdateFirewallRule update firewall rule.
func (cli *FirewallClient) UpdateFirewallRule(ctx context.Context, h http.Header, id string,
	request *proto.GcpFirewallRuleUpdateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/firewalls/rules/%s", id).
		WithHeaders(h).
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

// DeleteFirewallRule delete gcp firewall rule rule.
func (cli *FirewallClient) DeleteFirewallRule(ctx context.Context, h http.Header, id string) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		SubResourcef("/firewalls/rules/%s", id).
		WithHeaders(h).
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
