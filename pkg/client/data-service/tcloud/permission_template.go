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
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewPermissionTemplateClient create a new permission template api client.
func NewPermissionTemplateClient(client rest.ClientInterface) *PermissionTemplateClient {
	return &PermissionTemplateClient{client: client}
}

// PermissionTemplateClient is data service permission template api client.
type PermissionTemplateClient struct {
	client rest.ClientInterface
}

// BatchCreate batch create permission templates.
func (c *PermissionTemplateClient) BatchCreate(kt *kit.Kit,
	req *protocloud.PermissionTemplateBatchCreateReq[corecloud.TCloudPermissionTemplateExtension]) (
	*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/create").
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

// BatchUpdate batch update permission templates.
func (c *PermissionTemplateClient) BatchUpdate(kt *kit.Kit,
	req *protocloud.PermissionTemplateBatchUpdateReq[corecloud.TCloudPermissionTemplateExtension]) error {

	resp := new(rest.BaseResp)
	err := c.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/batch/update").
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

// ListPermissionTemplateExt list permission templates with extension.
func (c *PermissionTemplateClient) ListPermissionTemplateExt(kt *kit.Kit, req *protocloud.PermissionTemplateExtListReq,
) (*protocloud.PermissionTemplateExtListResult[corecloud.TCloudPermissionTemplateExtension], error) {

	resp := new(protocloud.PermissionTemplateExtListResp[corecloud.TCloudPermissionTemplateExtension])
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/list").
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
