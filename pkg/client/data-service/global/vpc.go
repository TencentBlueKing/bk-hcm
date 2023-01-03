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
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// VpcClient is data service vpc api client.
type VpcClient struct {
	client rest.ClientInterface
}

// NewVpcClient create a new vpc api client.
func NewVpcClient(client rest.ClientInterface) *VpcClient {
	return &VpcClient{
		client: client,
	}
}

// List vpcs.
func (v *VpcClient) List(ctx context.Context, h http.Header, req *core.ListReq) (*protocloud.VpcListResult, error) {
	resp := new(protocloud.VpcListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/list").
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

// BatchUpdateAttachment batch update vpc attachment.
func (v *VpcClient) BatchUpdateAttachment(ctx context.Context, h http.Header,
	req *protocloud.VpcAttachmentBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/attachments/batch").
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

// BatchDelete vpc.
func (v *VpcClient) BatchDelete(ctx context.Context, h http.Header, req *dataservice.BatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Delete().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/batch").
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
