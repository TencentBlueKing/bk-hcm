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

	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// BatchCreateEipCvmRel ...
func (rc *restClient) BatchCreateEipCvmRel(ctx context.Context,
	h http.Header,
	request *dataproto.EipCvmRelBatchCreateReq,
) error {
	resp := new(rest.BaseResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eip_cvm_rels/batch/create").
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

// ListEipCvmRel ...
func (rc *restClient) ListEipCvmRel(ctx context.Context,
	h http.Header, request *dataproto.EipCvmRelListReq,
) (*dataproto.EipCvmRelListResult, error) {
	resp := new(dataproto.EipCvmRelListResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eip_cvm_rels/list").
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

// ListEipCvmRelWithEip ...
func (rc *restClient) ListEipCvmRelWithEip(ctx context.Context,
	h http.Header,
	request *dataproto.EipCvmRelWithEipListReq,
) ([]*dataproto.EipWithCvmID, error) {
	resp := new(dataproto.EipCvmRelWithEipListResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eip_cvm_rels/with/eips/list").
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

// DeleteEipCvmRel ...
func (rc *restClient) DeleteEipCvmRel(
	ctx context.Context,
	h http.Header,
	request *dataproto.EipCvmRelDeleteReq,
) error {
	resp := new(rest.BaseResp)

	err := rc.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eip_cvm_rels/batch").
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

// ListEipWithoutCvm ...
func (rc *restClient) ListEipWithoutCvm(ctx context.Context, h http.Header, request *dataproto.ListEipWithoutCvmReq,
) (*dataproto.ListEipWithoutCvmResult, error) {

	resp := new(dataproto.ListEipWithoutCvmResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/eip_cvm_rels/with/eips/without/cvm/list").
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
