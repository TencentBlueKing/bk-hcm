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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
)

// ListCert list cert.
func (cli *restClient) ListCert(kt *kit.Kit, request *core.ListReq) (*protocloud.CertListResult, error) {
	resp := new(protocloud.CertListResp)

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

// BatchUpdateCert batch update cert.
func (cli *restClient) BatchUpdateCert(kt *kit.Kit, request *protocloud.CertBatchUpdateExprReq) (interface{}, error) {
	resp := new(core.UpdateResp)
	err := cli.client.Patch().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/certs").
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

// BatchDeleteCert batch delete cert.
func (cli *restClient) BatchDeleteCert(ctx context.Context, h http.Header,
	request *protocloud.CertBatchDeleteReq) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/certs/batch").
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
