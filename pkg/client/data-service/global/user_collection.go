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
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	dsuser "hcm/pkg/api/data-service/user"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// UserCollectionClient is data service user collection api client.
type UserCollectionClient struct {
	client rest.ClientInterface
}

// NewUserCollectionClient create a new user collection api client.
func NewUserCollectionClient(client rest.ClientInterface) *UserCollectionClient {
	return &UserCollectionClient{
		client: client,
	}
}

// List ...
func (a *UserCollectionClient) List(kt *kit.Kit, request *core.ListReq) (*dsuser.UserCollectionListResult, error) {
	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          *dsuser.UserCollectionListResult `json:"data"`
	}{}

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/users/collections/list").
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

// BatchDelete ...
func (a *UserCollectionClient) BatchDelete(kt *kit.Kit, request *dataservice.BatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := a.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/users/collections/batch").
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

// Create ...
func (a *UserCollectionClient) Create(kt *kit.Kit, request *dsuser.UserCollectionCreateReq) (
	*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/users/collections/create").
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
