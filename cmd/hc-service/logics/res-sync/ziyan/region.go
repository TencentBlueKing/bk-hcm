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
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesregion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud/region"
	dataservice "hcm/pkg/api/data-service"
	dataregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncRegionOption ...
type SyncRegionOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate ...
func (opt SyncRegionOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// extractAreaAndCityName 从 region_name 中提取 area_name 和 city_name
// 例如：从 "华南地区(广州)" 提取 area_name="华南地区", city_name="广州"
func extractAreaAndCityName(regionName string) (areaName, cityName string) {
	if len(regionName) == 0 {
		return "", ""
	}

	leftIdx := strings.Index(regionName, "(")
	if leftIdx < 0 {
		// 没有英文括号，整个字符串作为 city_name
		return "", strings.TrimSpace(regionName)
	}

	// 查找第一个右括号
	content := regionName[leftIdx+1:]
	rightIdx := strings.Index(content, ")")
	if rightIdx < 0 {
		// 没有找到右括号，整个字符串作为 city_name
		return "", strings.TrimSpace(regionName)
	}

	// 提取括号内容，如果包含嵌套括号，只取第一个括号对的内容
	cityName = strings.TrimSpace(content[:rightIdx])
	if nestedIdx := strings.Index(cityName, "("); nestedIdx >= 0 {
		cityName = strings.TrimSpace(cityName[:nestedIdx])
	}

	// 如果左括号在开头，area_name 为空
	if leftIdx == 0 {
		return "", cityName
	}

	return strings.TrimSpace(regionName[:leftIdx]), cityName
}

// Region ...
func (cli *client) Region(kt *kit.Kit, opt *SyncRegionOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regionFromCloud, err := cli.listRegionFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	// 从本地查询region
	allRegionFromDB, err := cli.listRegionFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	// 仅对 source 为 sync 的数据进行云上对比
	regionFromDB := slice.Filter(allRegionFromDB, func(region cloudcore.TCloudZiyanRegion) bool {
		return region.Source == enumor.RegionSourceSync
	})

	if len(regionFromCloud) == 0 && len(regionFromDB) == 0 {
		return new(SyncResult), nil
	}

	// 只对 sync 的数据进行 diff
	addSlice, updateMap, delCloudIDs := common.Diff[typesregion.TCloudRegion, cloudcore.TCloudZiyanRegion](
		regionFromCloud, regionFromDB, isRegionChange)

	// 对于需要 add 的 region，检查是否在 allRegionFromDB 中存在（可能是过去临时手动添加的）
	// 如果存在，需要先删除
	var toDeleteForAdd []string
	if len(addSlice) > 0 {
		toDeleteForAdd = cli.findExistingRegionsToDelete(addSlice, allRegionFromDB)
	}

	// 删除 diff 出来的 region（这些 region 在云上不存在，需要验证）
	if len(delCloudIDs) > 0 || len(toDeleteForAdd) > 0 {
		if err := cli.deleteRegion(kt, opt, delCloudIDs, toDeleteForAdd); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createRegion(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateRegion(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createRegion(kt *kit.Kit, opt *SyncRegionOption,
	addSlice []typesregion.TCloudRegion) error {

	if len(addSlice) <= 0 {
		return errors.New("region addSlice is <= 0, not create")
	}

	createResources := make([]dataregion.TCloudRegionBatchCreate, 0, len(addSlice))

	for _, one := range addSlice {
		areaName, cityName := extractAreaAndCityName(one.RegionName)
		tmpRes := dataregion.TCloudRegionBatchCreate{
			Vendor:     enumor.TCloudZiyan,
			RegionID:   one.RegionID,
			RegionName: one.RegionName,
			AreaName:   areaName,
			CityName:   cityName,
			Status:     one.RegionState,
			Source:     enumor.RegionSourceSync,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dataregion.TCloudRegionCreateReq{
		Regions: createResources,
	}
	if _, err := cli.dbCli.TCloudZiyan.Region.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("[%s] create region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to create region success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateRegion(kt *kit.Kit, opt *SyncRegionOption,
	updateMap map[string]typesregion.TCloudRegion) error {

	if len(updateMap) <= 0 {
		return errors.New("region updateMap is <= 0, not update")
	}

	updateResources := make([]dataregion.TCloudRegionBatchUpdate, 0, len(updateMap))

	for id, one := range updateMap {
		areaName, cityName := extractAreaAndCityName(one.RegionName)
		tmpRes := dataregion.TCloudRegionBatchUpdate{
			ID:         id,
			RegionID:   one.RegionID,
			RegionName: one.RegionName,
			AreaName:   areaName,
			CityName:   cityName,
			Status:     one.RegionState,
		}
		updateResources = append(updateResources, tmpRes)
	}

	updateReq := &dataregion.TCloudRegionBatchUpdateReq{
		Regions: updateResources,
	}
	if err := cli.dbCli.TCloudZiyan.Region.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("[%s] update region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to update region success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

// findExistingRegionsToDelete 查找需要删除的已存在 region
// 返回这些 region 的 region_id 列表
func (cli *client) findExistingRegionsToDelete(addRegions []typesregion.TCloudRegion,
	allDBRegions []cloudcore.TCloudZiyanRegion) []string {

	toDeleteRegionIDs := make([]string, 0)
	addRegionMap := converter.SliceToMap(addRegions,
		func(t typesregion.TCloudRegion) (string, interface{}) {
			return t.RegionID, nil
		})

	for _, dbRegion := range allDBRegions {
		if _, exists := addRegionMap[dbRegion.RegionID]; exists {
			toDeleteRegionIDs = append(toDeleteRegionIDs, dbRegion.RegionID)
		}
	}

	return toDeleteRegionIDs
}

// deleteRegion 删除 region
// delCloudIDs: 要删除的 region_id 列表
func (cli *client) deleteRegion(kt *kit.Kit, opt *SyncRegionOption, delCloudIDs, toDeleteForAdd []string) error {
	if len(delCloudIDs) <= 0 && len(toDeleteForAdd) <= 0 {
		return errors.New("region delCloudIDs is <= 0, not delete")
	}

	delRegionFromCloud, err := cli.listRegionFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delRegionFromCloud {
		if _, exist := delCloudMap[one.RegionID]; exist {
			logs.Errorf("[%s] validate region not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.TCloudZiyan, opt, len(delRegionFromCloud), kt.Rid)
			return errors.New("validate region not exist failed, before delete")
		}
	}

	// 因新增而删除的本地资源，不需要和云上对比
	if len(toDeleteForAdd) > 0 {
		delCloudIDs = append(delCloudIDs, toDeleteForAdd...)
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("region_id", parts),
		}
		if err := cli.dbCli.TCloudZiyan.Region.BatchDelete(kt, deleteReq); err != nil {
			logs.Errorf("[%s] delete region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to delete region success, accountID: %s, count: %d, rid: %s", enumor.TCloudZiyan,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRegionFromCloud(kt *kit.Kit, opt *SyncRegionOption) ([]typesregion.TCloudRegion, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results, err := cli.cloudCli.ListRegion(kt)
	if err != nil {
		logs.Errorf("[%s] list region from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloudZiyan,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results.Details, nil
}

func (cli *client) listRegionFromDB(kt *kit.Kit, opt *SyncRegionOption) (
	[]cloudcore.TCloudZiyanRegion, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.TCloudZiyan,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]cloudcore.TCloudZiyanRegion, 0)
	for {
		req.Page.Start = start
		regions, err := cli.dbCli.TCloudZiyan.Region.ListRegion(kt, req)
		if err != nil {
			logs.Errorf("[%s] list region from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloudZiyan,
				err, opt.AccountID, req, kt.Rid)
			return nil, err
		}
		results = append(results, regions.Details...)

		if len(regions.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}

func isRegionChange(cloud typesregion.TCloudRegion, db cloudcore.TCloudZiyanRegion) bool {

	if cloud.RegionID != db.RegionID {
		return true
	}

	if cloud.RegionName != db.RegionName {
		return true
	}

	// area_name 和 city_name 也需相同，仅在第一次同步补充这些字段时需关注该对比
	areaName, cityName := extractAreaAndCityName(cloud.RegionName)
	if areaName != db.AreaName {
		return true
	}
	if cityName != db.CityName {
		return true
	}

	if cloud.RegionState != db.Status {
		return true
	}

	return false
}
