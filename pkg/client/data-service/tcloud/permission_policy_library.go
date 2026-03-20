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
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewPermissionPolicyLibraryClient create a new permission policy library api client.
func NewPermissionPolicyLibraryClient(client rest.ClientInterface) *PermissionPolicyLibraryClient {
	return &PermissionPolicyLibraryClient{client: client}
}

// PermissionPolicyLibraryClient is data service permission policy library api client.
type PermissionPolicyLibraryClient struct {
	client rest.ClientInterface
}

// BatchCreate batch create permission policy libraries.
func (c *PermissionPolicyLibraryClient) BatchCreate(kt *kit.Kit,
	req *protocloud.PermissionPolicyLibraryBatchCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_policy_libraries/create").
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

// BatchUpdate batch update permission policy libraries.
func (c *PermissionPolicyLibraryClient) BatchUpdate(kt *kit.Kit,
	req *protocloud.PermissionPolicyLibraryBatchUpdateReq) error {

	resp := new(rest.BaseResp)
	err := c.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_policy_libraries/batch/update").
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
