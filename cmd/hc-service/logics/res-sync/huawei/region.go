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

package huawei

import (
	"errors"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesregion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	coreregion "hcm/pkg/api/core/cloud/region"
	dataregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncRegionOption ...
type SyncRegionOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate ...
func (opt SyncRegionOption) Validate() error {
	return validator.Validate.Struct(opt)
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

	regionFromDB, err := cli.listRegionFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(regionFromCloud) == 0 && len(regionFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesregion.HuaWeiRegionModel, coreregion.HuaWeiRegion](
		regionFromCloud, regionFromDB, isRegionChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteRegion(kt, opt, delCloudIDs); err != nil {
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
	addSlice []typesregion.HuaWeiRegionModel) error {

	if len(addSlice) <= 0 {
		return errors.New("region addSlice is <= 0, not create")
	}

	list := make([]dataregion.HuaWeiRegionBatchCreate, 0, len(addSlice))
	for _, one := range addSlice {
		regionOne := dataregion.HuaWeiRegionBatchCreate{
			RegionID:    one.RegionID,
			Type:        one.Type,
			Service:     one.Service,
			LocalesZhCn: one.ChinaName,
		}
		list = append(list, regionOne)
	}

	createReq := &dataregion.HuaWeiRegionBatchCreateReq{
		Regions: list,
	}
	_, err := cli.dbCli.HuaWei.Region.BatchCreateRegion(kt.Ctx, kt.Header(),
		createReq)
	if err != nil {
		logs.Errorf("[%s] create region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to create region success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateRegion(kt *kit.Kit, opt *SyncRegionOption,
	updateMap map[string]typesregion.HuaWeiRegionModel) error {

	if len(updateMap) <= 0 {
		return errors.New("region updateMap is <= 0, not update")
	}

	list := make([]dataregion.HuaWeiRegionBatchUpdate, 0, len(updateMap))

	for id, one := range updateMap {
		regionOne := dataregion.HuaWeiRegionBatchUpdate{
			ID:          id,
			Type:        one.Type,
			LocalesZhCn: one.ChinaName,
		}
		list = append(list, regionOne)
	}

	updateReq := &dataregion.HuaWeiRegionBatchUpdateReq{
		Regions: list,
	}
	if err := cli.dbCli.HuaWei.Region.BatchUpdateRegion(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] update region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to update region success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteRegion(kt *kit.Kit, opt *SyncRegionOption, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return errors.New("region delCloudIDs is <= 0, not delete")
	}

	delRegionFromCloud, err := cli.listRegionFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delRegionFromCloud {
		if _, exsit := delCloudMap[one.RegionID+"|"+one.Service]; exsit {
			logs.Errorf("[%s] validate region not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.HuaWei, opt, len(delRegionFromCloud), kt.Rid)
			return errors.New("validate region not exist failed, before delete")
		}
	}

	for _, one := range delCloudIDs {
		idArray := strings.Split(one, "|")
		if len(idArray) < 2 {
			continue
		}
		deleteReq := &dataregion.HuaWeiRegionBatchDeleteReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "service",
						Op:    filter.Equal.Factory(),
						Value: idArray[1],
					},
					&filter.AtomRule{
						Field: "region_id",
						Op:    filter.Equal.Factory(),
						Value: idArray[0],
					},
				},
			},
		}
		if err := cli.dbCli.HuaWei.Region.BatchDeleteRegion(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] delete region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to delete region success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRegionFromCloud(kt *kit.Kit, opt *SyncRegionOption) ([]typesregion.HuaWeiRegionModel, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regions, err := cli.cloudCli.ListRegion(kt)
	if err != nil {
		logs.Errorf("[%s] list region from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesregion.HuaWeiRegionModel, 0, len(regions))
	for _, one := range regions {
		results = append(results, converter.PtrToVal(one))
	}

	return results, nil
}

func (cli *client) listRegionFromDB(kt *kit.Kit, opt *SyncRegionOption) (
	[]coreregion.HuaWeiRegion, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]coreregion.HuaWeiRegion, 0)
	for {
		req.Page.Start = start
		regions, err := cli.dbCli.HuaWei.Region.ListRegion(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list region from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei, err,
				opt.AccountID, req, kt.Rid)
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

func isRegionChange(cloud typesregion.HuaWeiRegionModel, db coreregion.HuaWeiRegion) bool {

	if cloud.Type != db.Type {
		return true
	}

	if cloud.ChinaName != db.LocalesZhCn {
		return true
	}

	return false
}
