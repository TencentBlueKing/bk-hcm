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

// Package cloudserver defines cloud-server api client.
package cloudserver

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// Client is cloud-server api client.
type Client struct {
	rest.ClientInterface

	Account           *AccountClient
	Vpc               *VpcClient
	Subnet            *SubnetClient
	Cvm               *CvmClient
	RouteTable        *RouteTableClient
	ApprovalProcess   *ApprovalProcessClient
	ApplicationClient *ApplicationClient
}

// NewClient create a new cloud-server api client.
func NewClient(c *client.Capability, version string) *Client {
	restCli := rest.NewClient(c, fmt.Sprintf("/api/%s/cloud", version))
	return &Client{
		ClientInterface:   restCli,
		Account:           NewAccountClient(restCli),
		Vpc:               NewVpcClient(restCli),
		Subnet:            NewSubnetClient(restCli),
		Cvm:               NewCvmClient(restCli),
		ApprovalProcess:   NewApprovalProcessClient(restCli),
		RouteTable:        NewRouteTable(restCli),
		ApplicationClient: NewApplicationClient(restCli),
	}
}

// Request is a general-purpose helper method to reduce redundant code.
// Type parameter `T` is the type of request type, and `R` is the type of response.
func Request[T any, R any](cli rest.ClientInterface, method rest.VerbType, kt *kit.Kit, req *T,
	url string, urlArgs ...any) (*R, error) {

	resp := new(core.BaseResp[*R])

	err := cli.Verb(method).
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(url, urlArgs...).
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
