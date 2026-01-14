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
	"errors"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	corezone "hcm/pkg/api/core/cloud/zone"
	datazone "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

// SyncZoneOption ...
type SyncZoneOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	CityName  string
}

// Validate ...
func (opt SyncZoneOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// extractCityName 从 region_name 中提取 city_name
// 例如：从 "华南地区(广州)" 提取 "广州"
// 注意：此函数内部调用 extractAreaAndCityName，仅返回 city_name
func extractCityName(regionName string) string {
	_, cityName := extractAreaAndCityName(regionName)
	return cityName
}

// batchGetLogicCampusNameFromCmdb 批量从cmdb查询可用区对应的logic_campus_name
// 返回map[zoneName]logicCampusName，如果一个可用区对应多个logic_campus_name，只返回第一个
func (cli *client) batchGetLogicCampusNameFromCmdb(kt *kit.Kit, zoneNames []string) (map[string]string, error) {
	zoneToCampusMap := make(map[string]string) // 可用区名称 -> logic_campus_name

	// 过滤空字符串
	validZoneNames := make([]string, 0, len(zoneNames))
	for _, name := range zoneNames {
		if name != "" {
			validZoneNames = append(validZoneNames, name)
		}
	}

	if len(validZoneNames) == 0 {
		return zoneToCampusMap, nil
	}

	// 构建查询条件：根据可用区名称列表查询
	params := &cmdb.FindManyCmdbModuleParams{
		Filter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.In("availabilityZoneName", validZoneNames),
				},
			},
		},
		ScrollID: "0",
	}

	// 使用 scroll_id 遍历所有结果
	for {
		// 失败重试3次
		retryPolicy := retry.NewRetryPolicy(3, [2]uint{1000, 2000})
		var result *cmdb.FindManyCmdbModuleResult
		err := retryPolicy.BaseExec(kt, func() error {
			var retryErr error
			result, retryErr = cli.cmdbCli.FindManyCmdbModule(kt, params)
			if retryErr != nil {
				logs.Errorf("[%s] batch query cmdb module failed, err: %v, zones: %v, retry: %d, rid: %s",
					enumor.TCloudZiyan, retryErr, validZoneNames, retryPolicy.RetryCount(), kt.Rid)
			}
			return retryErr
		})
		if err != nil {
			logs.Errorf("[%s] batch query cmdb module failed after retries, err: %v, zones: %v, rid: %s",
				enumor.TCloudZiyan, err, validZoneNames, kt.Rid)
			return zoneToCampusMap, err
		}

		// 处理返回的结果
		for _, module := range result.List {
			// 遍历可用区信息，建立映射关系
			for _, azInfo := range module.AvailabilityZoneInfos {
				zoneName := azInfo.AvailabilityZoneName
				// 如果该可用区还没有对应的logic_campus_name，则设置
				// 如果已有，则跳过（只取第一个）
				if _, exists := zoneToCampusMap[zoneName]; !exists {
					zoneToCampusMap[zoneName] = module.LogicCampusName
				}
			}
		}

		// 如果结果列表为空，说明已经遍历完所有数据
		if len(result.List) == 0 {
			break
		}

		// 更新scroll_id继续查询
		params.ScrollID = result.ScrollID
		if params.ScrollID == "" || params.ScrollID == "0" {
			break
		}
	}

	return zoneToCampusMap, nil
}

// getCityNameFromRegion 根据 region_id 查询 region 信息并提取 city_name
func (cli *client) getCityNameFromRegion(kt *kit.Kit, regionID string) string {
	regionReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				tools.RuleEqual("region_id", regionID),
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	regions, err := cli.dbCli.TCloudZiyan.Region.ListRegion(kt, regionReq)
	if err != nil {
		logs.Warnf("[%s] query region failed, err: %v, region: %s, rid: %s", enumor.TCloudZiyan,
			err, regionID, kt.Rid)
		return ""
	}

	if len(regions.Details) == 0 {
		return ""
	}

	// 从 region_name 中提取 city_name
	return extractCityName(regions.Details[0].RegionName)
}

// Zone ...
func (cli *client) Zone(kt *kit.Kit, opt *SyncZoneOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 查询 region 获取 city_name
	opt.CityName = cli.getCityNameFromRegion(kt, opt.Region)

	zoneFromCloud, err := cli.listZoneFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	// 批量查询所有zone的logic_campus_name
	zoneNames := make([]string, 0, len(zoneFromCloud))
	for _, one := range zoneFromCloud {
		if one.ZoneName != "" {
			zoneNames = append(zoneNames, one.ZoneName)
		}
	}
	zoneToCampusMap, err := cli.batchGetLogicCampusNameFromCmdb(kt, zoneNames)
	if err != nil {
		return nil, err
	}
	// 将查询结果设置到云上zone数据中
	for i, one := range zoneFromCloud {
		if campusName, exists := zoneToCampusMap[one.ZoneName]; exists {
			zoneFromCloud[i].LogicCampusName = campusName
		}
	}

	// 从本地查询所有 zone
	allZoneFromDB, err := cli.listZoneFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	// 仅对 source 为 sync 的数据进行云上对比
	zoneFromDB := slice.Filter(allZoneFromDB, func(zone corezone.Zone[corezone.TCloudZoneExtension]) bool {
		return zone.Source == enumor.RegionSourceSync
	})

	if len(zoneFromCloud) == 0 && len(zoneFromDB) == 0 {
		return new(SyncResult), nil
	}

	// 只对 sync 的数据进行 diff
	addSlice, updateMap, delCloudIDs := common.Diff[typeszone.TCloudZone, corezone.Zone[corezone.TCloudZoneExtension]](
		zoneFromCloud, zoneFromDB, isZoneChange)

	// 对于需要 add 的 zone，检查是否在 allZoneFromDB 中存在（可能是过去临时手动添加的）
	// 如果存在，需要先删除
	var toDeleteForAdd []string
	if len(addSlice) > 0 {
		toDeleteForAdd = cli.findExistingZonesToDelete(addSlice, allZoneFromDB)
	}

	// 删除 diff 出来的 zone（这些 zone 在云上不存在，需要验证）
	if len(delCloudIDs) > 0 || len(toDeleteForAdd) > 0 {
		if err := cli.deleteZone(kt, opt, delCloudIDs, toDeleteForAdd); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createZone(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateZone(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createZone(kt *kit.Kit, opt *SyncZoneOption, addSlice []typeszone.TCloudZone) error {
	if len(addSlice) <= 0 {
		return errors.New("zone addSlice is <= 0, not create")
	}

	list := make([]datazone.ZoneBatchCreate[zone.TCloudZoneExtension], 0, len(addSlice))

	for _, one := range addSlice {
		zoneOne := datazone.ZoneBatchCreate[zone.TCloudZoneExtension]{
			CloudID: one.CloudID,
			Name:    one.ZoneID,
			NameCn:  one.ZoneName,
			State:   one.State,
			Region:  opt.Region,
			Source:  enumor.RegionSourceSync,
			Extension: &zone.TCloudZoneExtension{
				CityName:        opt.CityName,
				LogicCampusName: one.LogicCampusName,
			},
		}
		list = append(list, zoneOne)
	}

	createReq := &datazone.ZoneBatchCreateReq[zone.TCloudZoneExtension]{
		Zones: list,
	}
	_, err := cli.dbCli.TCloudZiyan.Zone.BatchCreateZone(kt, createReq)
	if err != nil {
		logs.Errorf("[%s] create zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync zone to create zone success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateZone(kt *kit.Kit, opt *SyncZoneOption, updateMap map[string]typeszone.TCloudZone) error {
	if len(updateMap) <= 0 {
		return errors.New("zone updateMap is <= 0, not update")
	}

	updates := make([]datazone.ZoneBatchUpdate[zone.TCloudZoneExtension], 0, len(updateMap))

	for id, one := range updateMap {
		update := datazone.ZoneBatchUpdate[zone.TCloudZoneExtension]{
			ID:    id,
			State: one.State,
			Extension: &zone.TCloudZoneExtension{
				CityName:        opt.CityName,
				LogicCampusName: one.LogicCampusName,
			},
		}
		updates = append(updates, update)
	}

	updateReq := &datazone.ZoneBatchUpdateReq[zone.TCloudZoneExtension]{
		Zones: updates,
	}
	if err := cli.dbCli.TCloudZiyan.Zone.BatchUpdateZone(kt, updateReq); err != nil {
		logs.Errorf("[%s] update zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync zone to update zone success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

// findExistingZonesToDelete 查找需要删除的已存在 zone
// 返回这些 zone 的 cloud_id 列表
func (cli *client) findExistingZonesToDelete(addZones []typeszone.TCloudZone,
	allDBZones []corezone.Zone[corezone.TCloudZoneExtension]) []string {

	toDeleteCloudIDs := make([]string, 0)
	addZoneMap := converter.SliceToMap(addZones,
		func(t typeszone.TCloudZone) (string, interface{}) {
			return t.GetCloudID(), nil
		})

	for _, dbZone := range allDBZones {
		if _, exists := addZoneMap[dbZone.GetCloudID()]; exists {
			toDeleteCloudIDs = append(toDeleteCloudIDs, dbZone.GetCloudID())
		}
	}

	return toDeleteCloudIDs
}

// deleteZone 删除 zone
// delCloudIDs: 要删除的 zone cloud_id 列表
// toDeleteForAdd: 因新增而需要删除的 zone cloud_id 列表（可能是之前手动添加的）
func (cli *client) deleteZone(kt *kit.Kit, opt *SyncZoneOption, delCloudIDs, toDeleteForAdd []string) error {
	if len(delCloudIDs) <= 0 && len(toDeleteForAdd) <= 0 {
		return errors.New("zone delCloudIDs is <= 0, not delete")
	}

	delZoneFromCloud, err := cli.listZoneFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delZoneFromCloud {
		if _, exsit := delCloudMap[one.GetCloudID()]; exsit {
			logs.Errorf("[%s] validate zone not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.TCloudZiyan, opt, len(delZoneFromCloud), kt.Rid)
			return errors.New("validate zone not exist failed, before delete")
		}
	}

	// 因新增而删除的本地资源，不需要和云上对比
	if len(toDeleteForAdd) > 0 {
		delCloudIDs = append(delCloudIDs, toDeleteForAdd...)
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &datazone.ZoneBatchDeleteReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					tools.RuleEqual("vendor", enumor.TCloudZiyan),
					tools.ContainersExpression("cloud_id", parts),
				},
			},
		}

		err := cli.dbCli.Global.Zone.BatchDeleteZone(kt.Ctx, kt.Header(), deleteReq)
		if err != nil {
			logs.Errorf("[%s] delete zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync zone to delete zone success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listZoneFromCloud(kt *kit.Kit, opt *SyncZoneOption) ([]typeszone.TCloudZone, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	zoneOpt := &typeszone.TCloudZoneListOption{
		Region: opt.Region,
	}
	results, err := cli.cloudCli.ListZone(kt, zoneOpt)
	if err != nil {
		logs.Errorf("[%s] list zone from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results, nil
}

func (cli *client) listZoneFromDB(kt *kit.Kit, opt *SyncZoneOption) (
	[]corezone.Zone[corezone.TCloudZoneExtension], error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &datazone.ZoneListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.TCloudZiyan,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: opt.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]corezone.Zone[corezone.TCloudZoneExtension], 0)
	for {
		req.Page.Start = start
		zones, err := cli.dbCli.TCloudZiyan.Zone.ListZoneExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list zone from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloudZiyan,
				err,
				opt.AccountID, req, kt.Rid)
			return nil, err
		}
		results = append(results, zones.Details...)

		if len(zones.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}

func isZoneChange(cloud typeszone.TCloudZone, db corezone.Zone[corezone.TCloudZoneExtension]) bool {

	if cloud.ZoneID != db.Name {
		return true
	}

	if cloud.ZoneName != db.NameCn {
		return true
	}

	if cloud.State != db.State {
		return true
	}

	// cityName实际不是云上字段，因此仅在该字段为空时需关注该对比
	if db.Extension.CityName == "" {
		return true
	}

	if cloud.LogicCampusName != db.Extension.LogicCampusName {
		return true
	}

	return false
}
