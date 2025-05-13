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

package other

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewCloudCvmClient create a new cvm api client.
func NewCloudCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// CvmClient is data service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// BatchCreateCvm batch create cvm rule.
func (cli *CvmClient) BatchCreateCvm(ctx context.Context, h http.Header,
	request *protocloud.CvmBatchCreateReq[corecvm.OtherCvmExtension]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cvms/batch/create").
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

// BatchUpdateCvm batch update cvm.
func (cli *CvmClient) BatchUpdateCvm(ctx context.Context, h http.Header,
	request *protocloud.CvmBatchUpdateReq[corecvm.OtherCvmExtension]) error {

	resp := new(rest.BaseResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cvms/batch/update").
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

// GetCvm get cvm.
func (cli *CvmClient) GetCvm(ctx context.Context, h http.Header, id string) (
	*corecvm.Cvm[corecvm.OtherCvmExtension], error) {

	resp := new(protocloud.CvmGetResp[corecvm.OtherCvmExtension])

	err := cli.client.Get().
		WithContext(ctx).
		SubResourcef("/cvms/%s", id).
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

// ListCvmExt list cvm with extension.
func (cli *CvmClient) ListCvmExt(ctx context.Context, h http.Header, request *protocloud.CvmListReq) (
	*protocloud.CvmExtListResult[corecvm.OtherCvmExtension], error) {

	resp := new(protocloud.CvmExtListResp[corecvm.OtherCvmExtension])

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cvms/list").
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
