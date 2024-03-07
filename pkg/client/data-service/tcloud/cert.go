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

	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
)

// ListCert 查询证书列表(带 extension 字段)
func (rc *restClient) ListCert(ctx context.Context, h http.Header, request *core.ListReq) (
	*protocloud.CertListExtResult[corecert.TCloudCertExtension], error) {

	resp := new(protocloud.CertListExtResp[corecert.TCloudCertExtension])
	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/certs/list").
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

// BatchCreateCert batch create cert.
func (rc *restClient) BatchCreateCert(ctx context.Context, h http.Header,
	request *protocloud.CertBatchCreateReq[corecert.TCloudCertExtension]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/certs/create").
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

// BatchUpdateCert batch update cert.
func (rc *restClient) BatchUpdateCert(ctx context.Context, h http.Header,
	request *protocloud.CertExtBatchUpdateReq[corecert.TCloudCertExtension]) (interface{}, error) {

	resp := new(core.UpdateResp)
	err := rc.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/certs").
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
