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
	corecloud "hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudSGCommonRelClient create a new security group common rel api client.
func NewCloudSGCommonRelClient(client rest.ClientInterface) *SGCommonRelClient {
	return &SGCommonRelClient{
		client: client,
	}
}

// SGCommonRelClient is data service security group common rel api client.
type SGCommonRelClient struct {
	client rest.ClientInterface
}

// BatchCreateSgCommonRels security group common rels.
func (cli *SGCommonRelClient) BatchCreateSgCommonRels(kt *kit.Kit, request *protocloud.SGCommonRelBatchCreateReq) error {
	return common.RequestNoResp[protocloud.SGCommonRelBatchCreateReq](cli.client, rest.POST, kt, request,
		"/security_group_common_rels/batch/create")
}

// BatchUpsertSgCommonRels security group common rels.
func (cli *SGCommonRelClient) BatchUpsertSgCommonRels(kt *kit.Kit, request *protocloud.SGCommonRelBatchUpsertReq) error {
	return common.RequestNoResp[protocloud.SGCommonRelBatchUpsertReq](cli.client, rest.POST, kt, request,
		"/security_group_common_rels/batch/upsert")
}

// BatchDeleteSgCommonRels security group common rels.
func (cli *SGCommonRelClient) BatchDeleteSgCommonRels(kt *kit.Kit, request *proto.BatchDeleteReq) error {
	return common.RequestNoResp[proto.BatchDeleteReq](cli.client, rest.DELETE, kt, request,
		"/security_group_common_rels/batch")
}

// ListSgCommonRels security group common rels.
func (cli *SGCommonRelClient) ListSgCommonRels(kt *kit.Kit, request *core.ListReq) (*protocloud.SGCommonRelListResult, error) {
	return common.Request[core.ListReq, protocloud.SGCommonRelListResult](cli.client, rest.POST, kt, request,
		"/security_group_common_rels/list")
}

// ListWithSecurityGroup security group common rels with security group.
func (cli *SGCommonRelClient) ListWithSecurityGroup(kt *kit.Kit,
	request *protocloud.SGCommonRelWithSecurityGroupListReq) (*[]corecloud.SGCommonRelWithBaseSecurityGroup, error) {

	return common.Request[protocloud.SGCommonRelWithSecurityGroupListReq, []corecloud.SGCommonRelWithBaseSecurityGroup](
		cli.client, rest.POST, kt, request, "/security_group_common_rels/with/security_group/list")
}
