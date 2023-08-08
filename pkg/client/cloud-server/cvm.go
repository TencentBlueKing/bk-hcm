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

package cloudserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// CvmClient is data service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// NewCvmClient create a new cvm api client.
func NewCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// List cvms.
func (v *CvmClient) List(kt *kit.Kit, bizID int64, req *core.ListReq) (*cloud.CvmListResult, error) {

	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          *cloud.CvmListResult `json:"data"`
	}{}

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/bizs/%d/cvms/list", bizID).
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
