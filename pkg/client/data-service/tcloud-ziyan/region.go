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

package ziyan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/region"
	dataservice "hcm/pkg/api/data-service"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/client/common"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// RegionClient is data service region api client.
type RegionClient struct {
	client rest.ClientInterface
}

// NewRegionClient create a new region api client.
func NewRegionClient(client rest.ClientInterface) *RegionClient {
	return &RegionClient{
		client: client,
	}
}

// BatchCreate batch create tcloud ziyan region.
func (v *RegionClient) BatchCreate(kt *kit.Kit,
	req *protoregion.TCloudRegionCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[protoregion.TCloudRegionCreateReq, core.BatchCreateResult](
		v.client, rest.POST, kt, req, "/regions/batch/create")
}

// BatchUpdate batch update tcloud ziyan region.
func (v *RegionClient) BatchUpdate(kt *kit.Kit, req *protoregion.TCloudRegionBatchUpdateReq) error {

	return common.RequestNoResp[protoregion.TCloudRegionBatchUpdateReq](v.client, rest.PATCH, kt, req, "/regions/batch")

}

// BatchForbiddenRegionState batch forbidden tcloud ziyan region state.
func (v *RegionClient) BatchForbiddenRegionState(kt *kit.Kit, req *protoregion.TCloudRegionBatchUpdateReq) error {

	return common.RequestNoResp[protoregion.TCloudRegionBatchUpdateReq](v.client, rest.PATCH, kt, req,
		"/regions/batch/state")
}

// BatchDelete batch delete tcloud ziyan region.
func (v *RegionClient) BatchDelete(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {

	return common.RequestNoResp[dataservice.BatchDeleteReq](v.client, rest.DELETE, kt, req, "/regions/batch")
}

// ListRegion get tcloud ziyan region list.
func (v *RegionClient) ListRegion(kt *kit.Kit, req *core.ListReq) (*types.ListResult[region.TCloudZiyanRegion], error) {

	return common.Request[core.ListReq, types.ListResult[region.TCloudZiyanRegion]](
		v.client, rest.POST, kt, req, "/regions/list")
}
