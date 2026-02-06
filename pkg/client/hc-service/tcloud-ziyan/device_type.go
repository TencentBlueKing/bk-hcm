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

package hcziyancli

import (
	"net/http"

	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// DeviceTypeClient is hc service tcloud ziyan device type api client.
type DeviceTypeClient struct {
	client rest.ClientInterface
}

// NewDeviceTypeClient create a new device type api client.
func NewDeviceTypeClient(client rest.ClientInterface) *DeviceTypeClient {
	return &DeviceTypeClient{
		client: client,
	}
}

// SyncDeviceType sync tcloud ziyan device type.
func (d *DeviceTypeClient) SyncDeviceType(kt *kit.Kit, req *sync.TCloudSyncReq) error {
	return common.RequestNoResp[sync.TCloudSyncReq](d.client, http.MethodPost, kt, req, "/device_types/sync")
}
