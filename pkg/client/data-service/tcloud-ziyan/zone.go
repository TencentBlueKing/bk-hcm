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
	corecloud "hcm/pkg/api/core/cloud/zone"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewZoneClient create a new zone api client.
func NewZoneClient(client rest.ClientInterface) *ZoneClient {
	return &ZoneClient{
		client: client,
	}
}

// ZoneClient is data service zone api client.
type ZoneClient struct {
	client rest.ClientInterface
}

// BatchCreateZone batch create zone.
func (cli *ZoneClient) BatchCreateZone(kt *kit.Kit,
	req *protocloud.ZoneBatchCreateReq[corecloud.TCloudZiyanZoneExtension]) (*core.BatchCreateResult, error) {

	return common.Request[protocloud.ZoneBatchCreateReq[corecloud.TCloudZiyanZoneExtension], core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/zones/batch/create")
}

// BatchUpdateZone batch update zone.
func (cli *ZoneClient) BatchUpdateZone(kt *kit.Kit,
	req *protocloud.ZoneBatchUpdateReq[corecloud.TCloudZiyanZoneExtension]) error {

	return common.RequestNoResp[protocloud.ZoneBatchUpdateReq[corecloud.TCloudZiyanZoneExtension]](
		cli.client, rest.PATCH, kt, req, "/zones/batch/update")
}

// ListZoneExt list zone with extension.
func (cli *ZoneClient) ListZoneExt(kt *kit.Kit, req *protocloud.ZoneListReq) (
	*protocloud.ZoneExtListResult[corecloud.TCloudZiyanZoneExtension], error) {

	return common.Request[protocloud.ZoneListReq, protocloud.ZoneExtListResult[corecloud.TCloudZiyanZoneExtension]](
		cli.client, rest.POST, kt, req, "/zones/list")
}
