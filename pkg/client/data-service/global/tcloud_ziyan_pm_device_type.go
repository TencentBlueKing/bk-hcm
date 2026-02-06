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
	pmdevicetype "hcm/pkg/api/data-service/tcloud-ziyan-pm-device-type"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// TCloudZiyanPmDeviceTypeClient is data service tcloud ziyan pm device type api client.
type TCloudZiyanPmDeviceTypeClient struct {
	client rest.ClientInterface
}

// NewTCloudZiyanPmDeviceTypeClient create a new tcloud ziyan pm device type api client.
func NewTCloudZiyanPmDeviceTypeClient(client rest.ClientInterface) *TCloudZiyanPmDeviceTypeClient {
	return &TCloudZiyanPmDeviceTypeClient{
		client: client,
	}
}

// Create ...
func (t *TCloudZiyanPmDeviceTypeClient) Create(kt *kit.Kit, req *pmdevicetype.CreateTCloudZiyanPmDeviceTypeReq) (
	*core.BatchCreateResult, error) {

	return common.Request[pmdevicetype.CreateTCloudZiyanPmDeviceTypeReq, core.BatchCreateResult](
		t.client, rest.POST, kt, req, "/tcloud_ziyan_pm_device_types/create")
}

// Update ...
func (t *TCloudZiyanPmDeviceTypeClient) Update(kt *kit.Kit, req *pmdevicetype.UpdateTCloudZiyanPmDeviceTypeReq) error {
	return common.RequestNoResp[pmdevicetype.UpdateTCloudZiyanPmDeviceTypeReq](
		t.client, rest.PATCH, kt, req, "/tcloud_ziyan_pm_device_types/update")
}

// List ...
func (t *TCloudZiyanPmDeviceTypeClient) List(kt *kit.Kit,
	req *core.ListReq) (*pmdevicetype.ListTCloudZiyanPmDeviceTypeResult, error) {

	return common.Request[core.ListReq, pmdevicetype.ListTCloudZiyanPmDeviceTypeResult](
		t.client, rest.POST, kt, req, "/tcloud_ziyan_pm_device_types/list")
}

// Delete ...
func (t *TCloudZiyanPmDeviceTypeClient) Delete(kt *kit.Kit, req *pmdevicetype.DeleteTCloudZiyanPmDeviceTypeReq) error {
	return common.RequestNoResp[pmdevicetype.DeleteTCloudZiyanPmDeviceTypeReq](
		t.client, rest.DELETE, kt, req, "/tcloud_ziyan_pm_device_types/delete")
}
