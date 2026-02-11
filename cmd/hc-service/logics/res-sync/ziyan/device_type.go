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
	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/api/core"
	devicetype "hcm/pkg/api/core/cloud/device-type"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// DeviceType 同步机型
func (cli *client) DeviceType(kt *kit.Kit, params *SyncDeviceTypeParams) (*SyncResult, error) {
	deviceTypesFromDB := make([]devicetype.DeviceType, 0)
	deviceTypesByZone := groupDeviceTypesByZone(params.DeviceTypes)

	for zone, deviceTypes := range deviceTypesByZone {
		deviceTypesNames := make([]string, 0, len(deviceTypes))
		for _, dt := range deviceTypes {
			deviceTypesNames = append(deviceTypesNames, dt.DeviceType)
		}

		for _, batch := range slice.Split(deviceTypesNames, int(core.DefaultMaxPageLimit)) {
			expression := tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleEqual("region", params.Region), tools.RuleEqual("zone", zone),
				tools.RuleIn("device_type", batch))
			curDeviceTypes, err := cli.listDeviceTypeFromDB(kt, expression)
			if err != nil {
				logs.Errorf("list device type from db failed, err: %v, region: %s, zone: %s, rid: %s",
					err, params.Region, zone, kt.Rid)
				return nil, err
			}
			deviceTypesFromDB = append(deviceTypesFromDB, curDeviceTypes...)
		}
	}

	addDeviceTypes, updateMap, delCloudIDs := common.Diff(params.DeviceTypes, deviceTypesFromDB, isDeviceTypeChanged)

	if err := cli.deleteDeviceType(kt, delCloudIDs); err != nil {
		return nil, err
	}

	if err := cli.updateDeviceType(kt, updateMap); err != nil {
		return nil, err
	}

	if err := cli.createDeviceType(kt, addDeviceTypes); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

// groupDeviceTypesByZone 按可用区对 DeviceTypes 进行分组
func groupDeviceTypesByZone(deviceTypes []devicetype.DeviceType) map[string][]devicetype.DeviceType {
	result := make(map[string][]devicetype.DeviceType)
	for _, dt := range deviceTypes {
		if _, ok := result[dt.Zone]; !ok {
			result[dt.Zone] = make([]devicetype.DeviceType, 0)
		}
		result[dt.Zone] = append(result[dt.Zone], dt)
	}
	return result
}

// listDeviceTypeFromDB 从数据库获取机型数据
func (cli *client) listDeviceTypeFromDB(kt *kit.Kit, filter *filter.Expression) ([]devicetype.DeviceType, error) {
	req := &protocloud.DeviceTypeListReq{
		ListReq: core.ListReq{
			Filter: filter,
			Page:   core.NewDefaultBasePage(),
		},
	}

	deviceTypes := make([]devicetype.DeviceType, 0)
	for {
		result, err := cli.dbCli.TCloudZiyan.DeviceType.ListDeviceType(kt, req)
		if err != nil {
			logs.Errorf("list device type from db failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		deviceTypes = append(deviceTypes, result.Details...)

		if len(result.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return deviceTypes, nil
}

func isDeviceTypeChanged(cloud devicetype.DeviceType, db devicetype.DeviceType) bool {
	if cloud.DeviceTypeClass != db.DeviceTypeClass {
		return true
	}

	if cloud.DeviceClass != db.DeviceClass {
		return true
	}

	if cloud.DeviceFamily != db.DeviceFamily {
		return true
	}

	if cloud.CoreType != db.CoreType {
		return true
	}

	if cloud.CpuCore != db.CpuCore {
		return true
	}

	if cloud.Memory != db.Memory {
		return true
	}

	if cloud.TechnicalClass != db.TechnicalClass {
		return true
	}

	if cloud.Source != db.Source {
		return true
	}

	return false
}

func (cli *client) createDeviceType(kt *kit.Kit, deviceTypes []devicetype.DeviceType) error {
	createItems := make([]protocloud.DeviceTypeCreate, 0, len(deviceTypes))
	for _, dt := range deviceTypes {
		createItems = append(createItems, protocloud.DeviceTypeCreate{
			DeviceType:      dt.DeviceType,
			DeviceClass:     dt.DeviceClass,
			DeviceFamily:    dt.DeviceFamily,
			CoreType:        dt.CoreType,
			CpuCore:         dt.CpuCore,
			Memory:          dt.Memory,
			DeviceTypeClass: dt.DeviceTypeClass,
			TechnicalClass:  dt.TechnicalClass,
			Region:          dt.Region,
			Zone:            dt.Zone,
			Disable:         false,
			Source:          enumor.DeviceTypeSourceSync,
		})
	}

	for _, batch := range slice.Split(createItems, constant.BatchOperationMaxLimit) {
		_, err := cli.dbCli.TCloudZiyan.DeviceType.BatchCreateDeviceType(kt, &protocloud.DeviceTypeBatchCreateReq{
			DeviceTypes: batch,
		})
		if err != nil {
			logs.Errorf("batch create device type failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func (cli *client) updateDeviceType(kt *kit.Kit, deviceTypes map[string]devicetype.DeviceType) error {
	updateItems := make([]protocloud.DeviceTypeUpdate, 0, len(deviceTypes))
	for id, dt := range deviceTypes {
		curDt := dt
		updateItems = append(updateItems, protocloud.DeviceTypeUpdate{
			ID:              id,
			DeviceType:      &curDt.DeviceType,
			DeviceClass:     &curDt.DeviceClass,
			DeviceFamily:    &curDt.DeviceFamily,
			CoreType:        &curDt.CoreType,
			CpuCore:         &curDt.CpuCore,
			Memory:          &curDt.Memory,
			DeviceTypeClass: &curDt.DeviceTypeClass,
			TechnicalClass:  &curDt.TechnicalClass,
			Source:          &curDt.Source,
		})
	}

	for _, batch := range slice.Split(updateItems, constant.BatchOperationMaxLimit) {
		err := cli.dbCli.TCloudZiyan.DeviceType.BatchUpdateDeviceType(kt, &protocloud.DeviceTypeBatchUpdateReq{
			DeviceTypes: batch,
		})
		if err != nil {
			logs.Errorf("batch update device type failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func (cli *client) deleteDeviceType(kt *kit.Kit, delCloudIDs []string) error {
	regionZoneDeviceTypeMap := make(map[string]map[string]map[string]struct{})
	for _, cloudID := range delCloudIDs {
		info, err := devicetype.GetInfoFromCloudID(cloudID)
		if err != nil {
			logs.Errorf("get cloud id info failed, err: %v, cloudID: %s, rid: %s", err, cloudID, kt.Rid)
			return err
		}
		if _, ok := regionZoneDeviceTypeMap[info.Region]; !ok {
			regionZoneDeviceTypeMap[info.Region] = make(map[string]map[string]struct{})
		}
		if _, ok := regionZoneDeviceTypeMap[info.Region][info.Zone]; !ok {
			regionZoneDeviceTypeMap[info.Region][info.Zone] = make(map[string]struct{})
		}
		regionZoneDeviceTypeMap[info.Region][info.Zone][info.DeviceType] = struct{}{}
	}

	for region, zoneDeviceTypeMap := range regionZoneDeviceTypeMap {
		for zone, deviceTypeMap := range zoneDeviceTypeMap {
			for _, batch := range slice.Split(maps.Keys(deviceTypeMap), constant.BatchOperationMaxLimit) {
				ft := tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
					tools.RuleEqual("region", region), tools.RuleEqual("zone", zone),
					tools.RuleIn("device_type", batch))
				err := cli.dbCli.Global.DeviceType.BatchDeleteDeviceType(kt, &dataproto.BatchDeleteReq{Filter: ft})
				if err != nil {
					logs.Errorf("batch delete device type failed, err: %v, region: %s, zone: %s, device types: %v, "+
						"rid: %s", err, region, zone, batch, kt.Rid)
					return err
				}
			}
		}
	}

	return nil
}

// RemoveDeviceTypeDeleteFromCloud 删除云上已不存在的机型
func (cli *client) RemoveDeviceTypeDeleteFromCloud(kt *kit.Kit, params *SyncRemovedParams,
	allCloudIDMap map[string]struct{}) error {

	deleteIDs := make([]string, 0)
	listReq := core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleEqual("region", params.Region)),
		Page: core.NewDefaultBasePage(),
	}
	for {
		result, err := cli.dbCli.TCloudZiyan.DeviceType.ListDeviceType(kt, &protocloud.DeviceTypeListReq{
			ListReq: listReq,
		})
		if err != nil {
			logs.Errorf("list device type from db failed, err: %v, req: %v, rid: %s", err, listReq, kt.Rid)
			return err
		}

		for _, dt := range result.Details {
			if _, ok := allCloudIDMap[dt.GetCloudID()]; !ok && dt.Source == enumor.DeviceTypeSourceSync {
				deleteIDs = append(deleteIDs, dt.ID)
			}
		}
		if len(result.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	for _, batch := range slice.Split(deleteIDs, constant.BatchOperationMaxLimit) {
		err := cli.dbCli.Global.DeviceType.BatchDeleteDeviceType(kt, &dataproto.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", batch),
		})
		if err != nil {
			logs.Errorf("batch delete device type failed, err: %v, batch: %v, rid: %s", err, batch, kt.Rid)
			return err
		}
	}

	return nil
}
