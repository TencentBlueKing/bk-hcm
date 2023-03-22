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

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
)

// CloudClient is data service cloud api client.
type CloudClient struct {
	client rest.ClientInterface
}

// NewCloudClient create a new cloud api client.
func NewCloudClient(client rest.ClientInterface) *CloudClient {
	return &CloudClient{
		client: client,
	}
}

// GetResourceBasicInfo get cloud resource basic info.
func (cli *CloudClient) GetResourceBasicInfo(ctx context.Context, h http.Header, resType enumor.CloudResourceType,
	resID string, fields ...string) (*types.CloudResourceBasicInfo, error) {

	req := &protocloud.GetResourceBasicInfoReq{Fields: fields}
	resp := new(protocloud.GetResourceBasicInfoResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/cloud/resources/bases/%s/id/%s", resType, resID).
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

// ListResourceBasicInfo list cloud resource basic info.
func (cli *CloudClient) ListResourceBasicInfo(ctx context.Context, h http.Header,
	req protocloud.ListResourceBasicInfoReq,
) (map[string]types.CloudResourceBasicInfo, error) {
	resp := new(protocloud.ListResourceBasicInfoResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/cloud/resources/bases/list").
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

// AssignResourceToBiz assign an account's cloud resource to biz, **only for ui**.
func (cli *CloudClient) AssignResourceToBiz(ctx context.Context, h http.Header,
	req *protocloud.AssignResourceToBizReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/cloud/resources/assign/bizs").
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
