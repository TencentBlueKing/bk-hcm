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

// Package task ...
package task

import (
	"time"

	configlogic "hcm/cmd/woa-server/logics/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	devicecapacity "hcm/pkg/api/data-service/device-capacity"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	croncore "hcm/pkg/cron/core"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"

	"golang.org/x/sync/errgroup"
)

// deviceCapacityRelInfo 主机库存关联的其他资源信息
type deviceCapacityRelInfo struct {
	requireTypeMap map[enumor.RequireType]struct{}
	regionZonesMap map[string]map[string]struct{}
	deviceTypeMap  map[string]struct{}
}

// DeviceCapacityTask is the task for device capacity.
type DeviceCapacityTask struct {
	clientSet    *client.ClientSet
	configLogics configlogic.Logics
}

// NewDeviceCapacityTask create a new device capacity task.
func NewDeviceCapacityTask(clientSet *client.ClientSet, configLogics configlogic.Logics) (croncore.Task, error) {
	return &DeviceCapacityTask{
		clientSet:    clientSet,
		configLogics: configLogics,
	}, nil
}

// Name return the name of the task.
func (d *DeviceCapacityTask) Name() string {
	return string(enumor.CronTaskSyncDeviceCapacity)
}

// Next return the next time to run the task.
func (d *DeviceCapacityTask) Next() (time.Time, error) {
	return time.Now().Add(time.Duration(cc.WoaServer().ResourceSync.SyncCapacity.Interval) * time.Minute), nil
}

// Do execute the task.
func (d *DeviceCapacityTask) Do(kt *kit.Kit) error {
	// 1. 根据其他资源的情况获取需求类型，地域、可用区、机型
	relInfoFromRes, err := d.getRelInfoFromRes(kt)
	if err != nil {
		logs.Errorf("failed to get rel res from res, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 根据已有的库存数据获取关联的需求类型，地域、可用区、机型
	relInfoFromCapacity, err := d.getRelInfoFromCapacity(kt)
	if err != nil {
		logs.Errorf("failed to get rel res from capacity, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 3. 对比需求类型、地域、可用区、机型差异, 删除无效的库存信息
	if err := d.deleteInvalidCapacity(kt, relInfoFromRes, relInfoFromCapacity); err != nil {
		logs.Errorf("failed to delete invalid capacity, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 4. 同步库存
	if err := d.syncCapacity(kt, relInfoFromRes); err != nil {
		logs.Errorf("failed to sync capacity, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (d *DeviceCapacityTask) getRelInfoFromRes(kt *kit.Kit) (*deviceCapacityRelInfo, error) {
	result := &deviceCapacityRelInfo{
		requireTypeMap: make(map[enumor.RequireType]struct{}),
		regionZonesMap: make(map[string]map[string]struct{}),
		deviceTypeMap:  make(map[string]struct{}),
	}

	// 获取需求类型
	for _, requireType := range enumor.GetRequireTypesForDeviceCapacity() {
		result.requireTypeMap[requireType] = struct{}{}
	}

	// 获取机型、地域、可用区
	deviceReq := &protocloud.DeviceTypeListReq{
		ListReq: core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: 0,
				Limit: core.DefaultMaxPageLimit,
				Sort:  "id",
			},
		},
	}
	for {
		deviceResp, err := d.clientSet.DataService().TCloudZiyan.DeviceType.ListDeviceType(kt, deviceReq)
		if err != nil {
			logs.Errorf("failed to list woa device type, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, one := range deviceResp.Details {
			result.deviceTypeMap[one.DeviceType] = struct{}{}
			if _, ok := result.regionZonesMap[one.Region]; !ok {
				result.regionZonesMap[one.Region] = make(map[string]struct{})
			}
			result.regionZonesMap[one.Region][one.Zone] = struct{}{}
		}
		if len(deviceResp.Details) < int(deviceReq.Page.Limit) {
			break
		}
		deviceReq.Page.Start += uint32(deviceReq.Page.Limit)
	}

	return result, nil
}

func (d *DeviceCapacityTask) getRelInfoFromCapacity(kt *kit.Kit) (*deviceCapacityRelInfo, error) {
	result := &deviceCapacityRelInfo{
		requireTypeMap: make(map[enumor.RequireType]struct{}),
		regionZonesMap: make(map[string]map[string]struct{}),
		deviceTypeMap:  make(map[string]struct{}),
	}
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}
	for {
		resp, err := d.clientSet.DataService().Global.DeviceCapacity.List(kt, req)
		if err != nil {
			logs.Errorf("failed to list device capacity, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range resp.Details {
			result.requireTypeMap[one.RequireType] = struct{}{}
			if _, ok := result.regionZonesMap[one.Region]; !ok {
				result.regionZonesMap[one.Region] = make(map[string]struct{})
			}
			result.regionZonesMap[one.Region][one.Zone] = struct{}{}
			result.deviceTypeMap[one.DeviceType] = struct{}{}
		}

		if len(resp.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return result, nil
}

func (d *DeviceCapacityTask) deleteInvalidCapacity(kt *kit.Kit, relInfoFromRes *deviceCapacityRelInfo,
	relInfoFromCapacity *deviceCapacityRelInfo) error {

	requireTypes := make([]enumor.RequireType, 0)
	regions := make([]string, 0)
	zones := make([]string, 0)
	deviceTypes := make([]string, 0)
	for requireType := range relInfoFromCapacity.requireTypeMap {
		if _, ok := relInfoFromRes.requireTypeMap[requireType]; !ok {
			requireTypes = append(requireTypes, requireType)
		}
	}
	for region, zoneMap := range relInfoFromCapacity.regionZonesMap {
		if _, ok := relInfoFromRes.regionZonesMap[region]; !ok {
			regions = append(regions, region)
			continue
		}
		for zone := range zoneMap {
			if _, ok := relInfoFromRes.regionZonesMap[region][zone]; !ok {
				zones = append(zones, zone)
			}
		}
	}
	for deviceType := range relInfoFromCapacity.deviceTypeMap {
		if _, ok := relInfoFromRes.deviceTypeMap[deviceType]; !ok {
			deviceTypes = append(deviceTypes, deviceType)
		}
	}

	if len(requireTypes) > 0 {
		for _, batch := range slice.Split(requireTypes, int(core.DefaultMaxPageLimit)) {
			req := &devicecapacity.DeleteDeviceCapacityReq{Filter: tools.ContainersExpression("require_type", batch)}
			if err := d.clientSet.DataService().Global.DeviceCapacity.Delete(kt, req); err != nil {
				logs.Errorf("failed to delete device capacity by require_type, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}
	if len(regions) > 0 {
		for _, batch := range slice.Split(regions, int(core.DefaultMaxPageLimit)) {
			req := &devicecapacity.DeleteDeviceCapacityReq{Filter: tools.ContainersExpression("region", batch)}
			if err := d.clientSet.DataService().Global.DeviceCapacity.Delete(kt, req); err != nil {
				logs.Errorf("failed to delete device capacity by region, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}
	if len(zones) > 0 {
		for _, batch := range slice.Split(zones, int(core.DefaultMaxPageLimit)) {
			req := &devicecapacity.DeleteDeviceCapacityReq{Filter: tools.ContainersExpression("zone", batch)}
			if err := d.clientSet.DataService().Global.DeviceCapacity.Delete(kt, req); err != nil {
				logs.Errorf("failed to delete device capacity by zone, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}
	if len(deviceTypes) > 0 {
		for _, batch := range slice.Split(deviceTypes, int(core.DefaultMaxPageLimit)) {
			req := &devicecapacity.DeleteDeviceCapacityReq{Filter: tools.ContainersExpression("device_type", batch)}
			if err := d.clientSet.DataService().Global.DeviceCapacity.Delete(kt, req); err != nil {
				logs.Errorf("failed to delete device capacity by device_type, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (d *DeviceCapacityTask) syncCapacity(kt *kit.Kit, relInfo *deviceCapacityRelInfo) error {

	var eg errgroup.Group
	eg.SetLimit(cc.WoaServer().ResourceSync.SyncCapacity.Concurrent)
	for requireType := range relInfo.requireTypeMap {
		for region, zoneMap := range relInfo.regionZonesMap {
			for zone := range zoneMap {
				for deviceType := range relInfo.deviceTypeMap {
					eg.Go(func() error {
						input := types.UpdateCapacityParam{
							RequireType: requireType,
							Region:      region,
							Zone:        zone,
							DeviceType:  deviceType,
						}
						if err := d.configLogics.Capacity().UpsertCapacity(kt, &input); err != nil {
							logs.Errorf("failed to upsert capacity, err: %v, input: %+v, rid: %s", err, input, kt.Rid)
							// 只打印日志，不影响其他同步
							return nil
						}
						return nil
					})
				}
			}
		}
	}

	return eg.Wait()
}

// GetURL get the url of the task, require every task to have external api in service.
func (d *DeviceCapacityTask) GetURL() string {
	return "/capacities/sync"
}
