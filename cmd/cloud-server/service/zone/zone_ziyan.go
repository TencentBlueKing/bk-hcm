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
	"hcm/pkg/api/core/cloud/zone"
	dataproto "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
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
