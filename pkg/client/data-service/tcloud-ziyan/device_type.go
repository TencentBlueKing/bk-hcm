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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// DeviceTypeClient is data service device type api client.
type DeviceTypeClient struct {
	client rest.ClientInterface
}

// NewDeviceTypeClient create a new device type api client.
func NewDeviceTypeClient(client rest.ClientInterface) *DeviceTypeClient {
	return &DeviceTypeClient{
		client: client,
	}
}

// ListDeviceType list device type
func (d *DeviceTypeClient) ListDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeListReq) (
	*protocloud.DeviceTypeListResult, error) {

	return common.Request[protocloud.DeviceTypeListReq, protocloud.DeviceTypeListResult](
		d.client, rest.POST, kt, req, "/device_types/list")
}

// ListDistinctDeviceType list distinct device type
func (d *DeviceTypeClient) ListDistinctDeviceType(kt *kit.Kit, req *protocloud.DistinctDeviceTypeListReq) (
	*protocloud.DistinctDeviceTypeListResult, error) {

	return common.Request[protocloud.DistinctDeviceTypeListReq, protocloud.DistinctDeviceTypeListResult](
		d.client, rest.POST, kt, req, "/device_types/distinct/list")
}

// BatchCreateDeviceType batch create device type
func (d *DeviceTypeClient) BatchCreateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[protocloud.DeviceTypeBatchCreateReq, core.BatchCreateResult](
		d.client, rest.POST, kt, req, "/device_types/batch/create")
}

// BatchUpdateDeviceType update device type
func (d *DeviceTypeClient) BatchUpdateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchUpdateReq) error {
	return common.RequestNoResp[protocloud.DeviceTypeBatchUpdateReq](
		d.client, rest.PATCH, kt, req, "/device_types/batch")
}
