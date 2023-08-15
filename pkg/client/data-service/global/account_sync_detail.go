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
	dssync "hcm/pkg/api/data-service/cloud/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountSyncDetailClient is data service account_sync_detail api client.
type AccountSyncDetailClient struct {
	client rest.ClientInterface
}

// NewAccountSyncDetailClient create a new account_sync_detail api client.
func NewAccountSyncDetailClient(client rest.ClientInterface) *AccountSyncDetailClient {
	return &AccountSyncDetailClient{
		client: client,
	}
}

// List ...
func (a *AccountSyncDetailClient) List(kt *kit.Kit, request *core.ListReq) (*dssync.ListResult, error) {
	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          *dssync.ListResult `json:"data"`
	}{}

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/account_sync_details/list").
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
func (a *AccountSyncDetailClient) BatchDelete(kt *kit.Kit, request *dataservice.BatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := a.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/account_sync_details/batch").
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

// BatchCreate ...
func (a *AccountSyncDetailClient) BatchCreate(kt *kit.Kit, request *dssync.CreateReq) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/account_sync_details/batch/create").
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

// BatchUpdate ...
func (a *AccountSyncDetailClient) BatchUpdate(kt *kit.Kit, request *dssync.UpdateReq) error {
	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/account_sync_details/batch/update").
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
