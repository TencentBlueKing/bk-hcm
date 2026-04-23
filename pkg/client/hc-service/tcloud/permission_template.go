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
	proto "hcm/pkg/api/hc-service/permission-template"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// PermissionTemplateClient is hc service permission template api client.
type PermissionTemplateClient struct {
	client rest.ClientInterface
}

// NewPermissionTemplateClient create a new permission template api client.
func NewPermissionTemplateClient(client rest.ClientInterface) *PermissionTemplateClient {
	return &PermissionTemplateClient{
		client: client,
	}
}

// CreateCAMPolicy creates a CAM policy via hc-service.
func (c *PermissionTemplateClient) CreateCAMPolicy(kt *kit.Kit, req *proto.CreateCAMPolicyReq) (
	*proto.CreateCAMPolicyResult, error) {

	resp := new(proto.CreateCAMPolicyResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/cam/create_policy").
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

// DeleteCAMPolicy deletes a CAM policy via hc-service.
func (c *PermissionTemplateClient) DeleteCAMPolicy(kt *kit.Kit, req *proto.DeleteCAMPolicyReq) error {
	resp := new(rest.BaseResp)

	err := c.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/cam/delete_policy").
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

// UpdateCAMPolicy updates a CAM policy via hc-service.
func (c *PermissionTemplateClient) UpdateCAMPolicy(kt *kit.Kit, req *proto.UpdateCAMPolicyReq) error {
	resp := new(rest.BaseResp)

	err := c.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/cam/update_policy").
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
