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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudArgumentTemplateClient create a new argument template api client.
func NewCloudArgumentTemplateClient(client rest.ClientInterface) *ArgsTplClient {
	return &ArgsTplClient{
		client: client,
	}
}

// ArgsTplClient is data service argument template api client.
type ArgsTplClient struct {
	client rest.ClientInterface
}

// ListArgsTpl list argument template.
func (cli *ArgsTplClient) ListArgsTpl(kt *kit.Kit, request *core.ListReq) (*protocloud.ArgsTplListResult, error) {
	return common.Request[core.ListReq, protocloud.ArgsTplListResult](cli.client, rest.POST, kt, request,
		"/argument_templates/list")
}

// BatchUpdateArgsTpl batch update argument template.
func (cli *ArgsTplClient) BatchUpdateArgsTpl(kt *kit.Kit, request *protocloud.ArgsTplBatchUpdateExprReq) (
	interface{}, error) {

	return common.Request[protocloud.ArgsTplBatchUpdateExprReq, core.UpdateResp](cli.client, rest.PUT, kt, request,
		"/argument_templates")
}

// BatchDeleteArgsTpl batch delete argument template.
func (cli *ArgsTplClient) BatchDeleteArgsTpl(kt *kit.Kit, request *protocloud.ArgsTplBatchDeleteReq) error {
	return common.RequestNoResp[protocloud.ArgsTplBatchDeleteReq](cli.client, rest.DELETE, kt, request,
		"/argument_templates/batch")
}
