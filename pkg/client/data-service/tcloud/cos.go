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

// Package tcloud cos.
package tcloud

import (
	"hcm/pkg/api/core"
	corecos "hcm/pkg/api/core/cloud/cos"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// CosClient cos client.
type CosClient struct {
	client rest.ClientInterface
}

// NewCosClient new cos client.
func NewCosClient(client rest.ClientInterface) *CosClient {
	return &CosClient{client: client}
}

// BatchCreateTCloudCos 批量创建腾讯云Cos
func (cli *CosClient) BatchCreateTCloudCos(kt *kit.Kit, req *protocloud.TCloudCosBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[protocloud.TCloudCosBatchCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/cos/batch/create")
}

// Get 获取Cos 详情
func (cli *CosClient) Get(kt *kit.Kit, id string) (*corecos.Cos[corecos.TCloudCosExtension],
	error) {

	return common.Request[common.Empty, corecos.Cos[corecos.TCloudCosExtension]](
		cli.client, rest.GET, kt, nil, "/cos/%s", id)
}

// BatchUpdate 批量更新Cos
func (cli *CosClient) BatchUpdate(kt *kit.Kit, req *protocloud.TCloudCosBatchUpdateReq) error {
	return common.RequestNoResp[protocloud.TCloudCosBatchUpdateReq](cli.client,
		rest.PATCH, kt, req, "/cos/batch/update")
}

// List list tcloud load balancer
func (cli *CosClient) List(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[corecos.TCloudCos], error) {

	return common.Request[core.ListReq, core.ListResultT[corecos.TCloudCos]](
		cli.client, rest.POST, kt, req, "/cos/list")

}
