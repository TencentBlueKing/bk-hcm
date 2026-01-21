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

package zone

import (
	"fmt"

	cloudproto "hcm/pkg/api/cloud-server/zone"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	dataproto "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/slice"
)

// importTCloudZiyanZone 导入TCloudZiyan zone
func (dSvc *ZoneSvc) importTCloudZiyanZone(cts *rest.Contexts, req *cloudproto.ZoneImportReq) (interface{}, error) {
	extension := &zone.TCloudZoneExtension{}
	if req.Extension != nil {
		if extMap, ok := req.Extension.(map[string]interface{}); ok {
			if cityName, ok := extMap["city_name"].(string); ok {
				extension.CityName = cityName
			}
			if logicCampusName, ok := extMap["logic_campus_name"].(string); ok {
				extension.LogicCampusName = logicCampusName
			}
		}
	}

	// 如果没有提供logic_campus_name，尝试从cmdb查询
	var err error
	if extension.LogicCampusName == "" && req.NameCn != "" {
		extension.LogicCampusName, err = getLogicCampusNameFromCmdb(cts.Kit, req.NameCn)
		if err != nil {
			logs.Errorf("get logic_campus_name from cmdb failed, err: %v, zone: %s, rid: %s", err, req.NameCn,
				cts.Kit.Rid)
			return nil, err
		}
	}

	createReq := &dataproto.ZoneBatchCreateReq[zone.TCloudZoneExtension]{
		Zones: []dataproto.ZoneBatchCreate[zone.TCloudZoneExtension]{
			{
				CloudID:   req.CloudID,
				Name:      req.Name,
				State:     req.State,
				Region:    req.Region,
				NameCn:    req.NameCn,
				Source:    enumor.RegionSourceManually,
				Extension: extension,
			},
		},
	}
	return dSvc.client.DataService().TCloudZiyan.Zone.BatchCreateZone(cts.Kit, createReq)
}

// getLogicCampusNameFromCmdb 从cmdb查询可用区对应的logic_campus_name
// 根据可用区名称查询，直接使用第一个结果作为logic_campus_name
func getLogicCampusNameFromCmdb(kt *kit.Kit, zoneName string) (string, error) {
	if zoneName == "" {
		return "", nil
	}

	// 构建查询条件：根据可用区名称查询
	params := &cmdb.FindManyCmdbModuleParams{
		Filter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.In("availabilityZoneName", []string{zoneName}),
				},
			},
		},
		ScrollID: "0",
	}

	// 无需遍历，直接使用第一个结果作为logic_campus_name
	result, err := cmdb.CmdbClient().FindManyCmdbModule(kt, params)
	if err != nil {
		logs.Errorf("query cmdb module failed, err: %v, zone: %s, rid: %s", err, zoneName, kt.Rid)
		return "", fmt.Errorf("query cmdb module failed, err: %v", err)
	}

	// 处理返回的结果，直接使用第一个匹配的logic_campus_name
	for _, module := range result.List {
		// 遍历可用区信息，查找匹配的可用区
		for _, azInfo := range module.AvailabilityZoneInfos {
			if azInfo.AvailabilityZoneName == zoneName {
				return module.LogicCampusName, nil
			}
		}
	}

	return "", fmt.Errorf("no logic_campus_name found for zone: %s", zoneName)
}

// batchUpdateTCloudZiyanZoneDisableCvm batch update tcloud-ziyan zone disable_cvm field.
func (dSvc *ZoneSvc) batchUpdateTCloudZiyanZoneDisableCvm(cts *rest.Contexts,
	req *cloudproto.ZoneBatchUpdateDisableCvmReq) (interface{}, error) {

	var zonesFromDB []zone.Zone[zone.TCloudZoneExtension]
	// 先查询现有的 zone 数据，获取当前的 Extension 值，通过 name 查询
	for _, batch := range slice.Split(req.Zones, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.ZoneListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "name", Op: filter.In.Factory(), Value: batch},
				},
			},
			Page: core.NewDefaultBasePage(),
		}

		// 查询 zone 列表
		listResp, err := dSvc.client.DataService().TCloudZiyan.Zone.ListZoneExt(cts.Kit, listReq)
		if err != nil {
			return nil, fmt.Errorf("query zones failed, err: %v", err)
		}

		if len(listResp.Details) == 0 {
			return nil, errf.New(errf.RecordNotFound, "no zones found with provided zones name")
		}

		zonesFromDB = append(zonesFromDB, listResp.Details...)
	}

	// 构建批量更新请求
	updates := make([]dataproto.ZoneBatchUpdate[zone.TCloudZoneExtension], 0, len(zonesFromDB))
	for _, existingZone := range zonesFromDB {
		extension := &zone.TCloudZoneExtension{
			DisableCvm: req.DisableCvm,
		}
		// 保留其他字段的原有值
		if existingZone.Extension != nil {
			extension.CityName = existingZone.Extension.CityName
			extension.LogicCampusName = existingZone.Extension.LogicCampusName
		}

		updates = append(updates, dataproto.ZoneBatchUpdate[zone.TCloudZoneExtension]{
			ID:        existingZone.ID,
			Extension: extension,
		})
	}

	// 分批更新 zone 的 disable_cvm 字段
	for _, batch := range slice.Split(updates, int(core.DefaultMaxPageLimit)) {
		updateReq := &dataproto.ZoneBatchUpdateReq[zone.TCloudZoneExtension]{
			Zones: batch,
		}

		err := dSvc.client.DataService().TCloudZiyan.Zone.BatchUpdateZone(cts.Kit, updateReq)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
