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
	"context"
	"net/http"

	"hcm/pkg/adaptor/types/cert"
	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	protocert "hcm/pkg/api/hc-service/cert"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCertClient create a new cert api client.
func NewCertClient(client rest.ClientInterface) *CertClient {
	return &CertClient{
		client: client,
	}
}

// CertClient is hc service cert api client.
type CertClient struct {
	client rest.ClientInterface
}

// CreateCert ....
func (cli *CertClient) CreateCert(kt *kit.Kit, request *protocert.TCloudCreateReq) (
	*corecert.CertCreateResult, error) {

	resp := new(protocert.BatchCreateResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/certs/create").
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

// DeleteCert ....
func (cli *CertClient) DeleteCert(kt *kit.Kit, request *protocert.TCloudDeleteReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/certs").
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

// ListCert ....
func (cli *CertClient) ListCert(kt *kit.Kit, request *protocert.TCloudListOption) (
	[]cert.TCloudCert, error) {

	resp := &struct {
		*rest.BaseResp `json:",inline"`
		Data           []cert.TCloudCert `json:"data"`
	}{}

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/certs/list").
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

// SyncCert sync cert.
func (cli *CertClient) SyncCert(ctx context.Context, h http.Header, request *sync.TCloudSyncReq) error {
	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/certs/sync").
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
