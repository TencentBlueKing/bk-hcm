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
	coredevicecapacity "hcm/pkg/api/core/device-capacity"
	"hcm/pkg/api/data-service/device-capacity"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// DeviceCapacityClient is data service device capacity api client.
type DeviceCapacityClient struct {
	client rest.ClientInterface
}

// NewDeviceCapacityClient create a new device capacity api client.
func NewDeviceCapacityClient(client rest.ClientInterface) *DeviceCapacityClient {
	return &DeviceCapacityClient{
		client: client,
	}
}

// List device capacity.
func (d *DeviceCapacityClient) List(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[coredevicecapacity.DeviceCapacity], error) {

	return common.Request[core.ListReq, core.ListResultT[coredevicecapacity.DeviceCapacity]](
		d.client, rest.POST, kt, req, "/device_capacities/list")
}

// Create device capacity.
func (d *DeviceCapacityClient) Create(kt *kit.Kit, req *devicecapacity.CreateDeviceCapacityReq) (
	*core.BatchCreateResult, error) {

	return common.Request[devicecapacity.CreateDeviceCapacityReq, core.BatchCreateResult](
		d.client, rest.POST, kt, req, "/device_capacities/create")
}

// Update update device capacity.
func (d *DeviceCapacityClient) Update(kt *kit.Kit, req *devicecapacity.UpdateDeviceCapacityReq) error {
	return common.RequestNoResp[devicecapacity.UpdateDeviceCapacityReq](d.client, rest.PATCH, kt, req,
		"/device_capacities/update")
}

// Delete device capacity.
func (d *DeviceCapacityClient) Delete(kt *kit.Kit, req *devicecapacity.DeleteDeviceCapacityReq) error {
	return common.RequestNoResp[devicecapacity.DeleteDeviceCapacityReq](d.client, rest.DELETE, kt, req,
		"/device_capacities/delete")
}

// ListCapacityWithDeviceInfo list device capacity with device info.
func (d *DeviceCapacityClient) ListCapacityWithDeviceInfo(kt *kit.Kit, req *devicecapacity.ListCapacityWithDeviceInfoReq) (
	*devicecapacity.ListCapacityWithDeviceInfoResult, error) {

	return common.Request[devicecapacity.ListCapacityWithDeviceInfoReq, devicecapacity.ListCapacityWithDeviceInfoResult](
		d.client, rest.POST, kt, req, "/device_capacities/list_with_device_info")
}
