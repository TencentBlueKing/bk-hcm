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

package cmdb

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
	"hcm/pkg/tools/uuid"
)

// Client is an esb client to request cmdb.
type Client interface {
	SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error)
	SearchCloudArea(ctx context.Context, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error)
}

// NewClient initialize a new cmdb client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &cmdb{
		client: client,
		config: config,
	}
}

// cmdb is an esb client to request cmdb.
type cmdb struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

func (c *cmdb) SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error) {
	resp := new(SearchBizResp)
	req := &esbSearchBizParams{
		CommParams:      types.GetCommParams(c.config),
		SearchBizParams: params,
	}
	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())
	err := c.client.Post().
		SubResourcef("/cc/search_business/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("search business failed, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid)
	}
	return resp, nil
}

// SearchCloudArea search cmdb cloud area
func (c *cmdb) SearchCloudArea(ctx context.Context, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error) {
	resp := new(SearchCloudAreaResp)

	req := &esbSearchCloudAreaParams{
		CommParams:            types.GetCommParams(c.config),
		SearchCloudAreaParams: params,
	}

	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())

	err := c.client.Post().
		SubResourcef("/cc/search_cloud_area/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("find cloud area failed, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid)
	}
	return resp.Data, nil
}
