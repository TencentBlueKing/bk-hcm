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
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protoargstpl "hcm/pkg/api/hc-service/argument-template"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewArgsTplClient create a new argument template api client.
func NewArgsTplClient(client rest.ClientInterface) *ArgsTplClient {
	return &ArgsTplClient{
		client: client,
	}
}

// ArgsTplClient is hc service argument template api client.
type ArgsTplClient struct {
	client rest.ClientInterface
}

// CreateArgsTpl ....
func (cli *ArgsTplClient) CreateArgsTpl(kt *kit.Kit, request *protoargstpl.TCloudCreateReq) (
	*coreargstpl.ArgsTplCreateResult, error) {

	return common.Request[protoargstpl.TCloudCreateReq, coreargstpl.ArgsTplCreateResult](
		cli.client, rest.POST, kt, request, "/argument_templates/create")
}

// UpdateArgsTpl ....
func (cli *ArgsTplClient) UpdateArgsTpl(kt *kit.Kit, id string, request *protoargstpl.TCloudUpdateReq) error {
	return common.RequestNoResp[protoargstpl.TCloudUpdateReq](cli.client, rest.PUT, kt, request,
		"/argument_templates/%s", id)
}

// DeleteArgsTpl ....
func (cli *ArgsTplClient) DeleteArgsTpl(kt *kit.Kit, request *protoargstpl.TCloudDeleteReq) error {
	return common.RequestNoResp[protoargstpl.TCloudDeleteReq](
		cli.client, rest.DELETE, kt, request, "/argument_templates")
}

// ListArgsTpl ....
func (cli *ArgsTplClient) ListArgsTpl(kt *kit.Kit, request *protoargstpl.ArgsTplListReq) (
	[]typeargstpl.TCloudArgsTplAddress, error) {

	resp := &struct {
		*rest.BaseResp `json:",inline"`
		Data           []typeargstpl.TCloudArgsTplAddress `json:"data"`
	}{}

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/argument_templates/list").
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

// SyncArgsTpl sync argument template.
func (cli *ArgsTplClient) SyncArgsTpl(kt *kit.Kit, request *sync.TCloudSyncReq) error {
	return common.RequestNoResp[sync.TCloudSyncReq](cli.client, rest.POST, kt, request, "/argument_templates/sync")
}
