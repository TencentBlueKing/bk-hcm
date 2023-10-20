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
	"hcm/pkg/api/hc-service/eip"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewEipClient create a new eip api client.
func NewEipClient(client rest.ClientInterface) *EipClient {
	return &EipClient{
		client: client,
	}
}

// EipClient is hc service eip api client.
type EipClient struct {
	client rest.ClientInterface
}

// SyncEip eip.
func (cli *EipClient) SyncEip(ctx context.Context, h http.Header, request *sync.GcpSyncReq) error {
	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eips/sync").
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

// DeleteEip ...
func (cli *EipClient) DeleteEip(kt *kit.Kit, req *eip.EipDeleteReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/eips").
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

// AssociateEip ...
func (cli *EipClient) AssociateEip(ctx context.Context, h http.Header, req *eip.GcpEipAssociateReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/eips/associate").
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

// DisassociateEip ...
func (cli *EipClient) DisassociateEip(ctx context.Context, h http.Header, req *eip.GcpEipDisassociateReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/eips/disassociate").
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

// CreateEip ...
func (cli *EipClient) CreateEip(
	ctx context.Context,
	h http.Header,
	req *eip.GcpEipCreateReq,
) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/eips/create").
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
