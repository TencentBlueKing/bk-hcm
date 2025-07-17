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

package azure

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCvmClient create a new cvm api client.
func NewCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// CvmClient is hc service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// SyncCvmWithRelResource sync cvm with rel resource.
func (cli *CvmClient) SyncCvmWithRelResource(ctx context.Context, h http.Header, request *sync.AzureSyncReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cvms/with/relation_resources/sync").
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

// StartCvm ....
func (cli *CvmClient) StartCvm(kt *kit.Kit, id string) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		SubResourcef("/cvms/%s/start", id).
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

// StopCvm ....
func (cli *CvmClient) StopCvm(kt *kit.Kit, id string, req *protocvm.AzureStopReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/cvms/%s/stop", id).
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

// RebootCvm ....
func (cli *CvmClient) RebootCvm(kt *kit.Kit, id string) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		SubResourcef("/cvms/%s/reboot", id).
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

// DeleteCvm ....
func (cli *CvmClient) DeleteCvm(kt *kit.Kit, id string, req *protocvm.AzureDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/cvms/%s", id).
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

// CreateCvm ....
func (cli *CvmClient) CreateCvm(kt *kit.Kit, request *protocvm.AzureCreateReq) (
	*protocvm.AzureCreateResp, error) {

	resp := new(core.BaseResp[*protocvm.AzureCreateResp])

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/create").
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

// SyncCCInfo ...
func (cli *CvmClient) SyncCCInfo(kt *kit.Kit, req *sync.AzureSyncReq) error {
	return common.RequestNoResp[sync.AzureSyncReq](cli.client, rest.POST, kt, req, "/cvms/cc_info/sync")
}

// SyncCCInfoByCond ...
func (cli *CvmClient) SyncCCInfoByCond(kt *kit.Kit, req *sync.SyncCvmByCondReq) error {
	return common.RequestNoResp[sync.SyncCvmByCondReq](cli.client, rest.POST, kt, req,
		"/cvms/cc_info/by_condition/sync")
}
