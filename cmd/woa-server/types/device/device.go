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

// Package device ...
package device

import (
	"sync"
	"time"

	"hcm/pkg/api/core"
	dt "hcm/pkg/api/core/cloud/device-type"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// DeviceTypesMap cache of device_type, reducing the pressure of MySQL.
type DeviceTypesMap struct {
	lock        sync.RWMutex
	client      *client.ClientSet
	DeviceTypes map[string]dt.DistinctDeviceType
	TTL         time.Time
}

// NewDeviceTypesMap ...
func NewDeviceTypesMap(client *client.ClientSet) *DeviceTypesMap {
	return &DeviceTypesMap{
		client:      client,
		DeviceTypes: make(map[string]dt.DistinctDeviceType),
		TTL:         time.Now(),
	}
}

func (d *DeviceTypesMap) updateDeviceTypesMap(kt *kit.Kit) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	deviceTypeMap := make(map[string]dt.DistinctDeviceType)
	page := core.NewDefaultBasePage()
	for {
		req := &protocloud.DistinctDeviceTypeListReq{
			ListReq: core.ListReq{
				Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
				Page:   page,
			},
		}
		result, err := d.client.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(kt, req)
		if err != nil {
			logs.Errorf("failed to list device type, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, detail := range result.Details {
			deviceTypeMap[detail.DeviceType] = detail
		}
		if len(result.Details) < int(page.Limit) {
			break
		}
		page.Start += uint32(page.Limit)
	}
	d.DeviceTypes = deviceTypeMap
	d.TTL = time.Now().Add(1 * time.Minute)
	return nil
}

// GetDeviceTypes get device type map from cache.
func (d *DeviceTypesMap) GetDeviceTypes(kt *kit.Kit) (map[string]dt.DistinctDeviceType, error) {
	d.lock.RLock()
	res := make(map[string]dt.DistinctDeviceType)
	if time.Now().After(d.TTL) {
		d.lock.RUnlock()
		err := d.updateDeviceTypesMap(kt)
		if err != nil {
			return nil, err
		}
		d.lock.RLock()
	}

	defer d.lock.RUnlock()
	for k := range d.DeviceTypes {
		res[k] = d.DeviceTypes[k]
	}
	return res, nil
}
