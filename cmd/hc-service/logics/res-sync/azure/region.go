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

package azure

import (
	"errors"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesregion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	coreregion "hcm/pkg/api/core/cloud/region"
	dataregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
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

	addSlice, updateMap, delCloudIDs := common.Diff[typesregion.AzureRegion, coreregion.AzureRegion](
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
	addSlice []typesregion.AzureRegion) error {

	if len(addSlice) <= 0 {
		return errors.New("region addSlice is <= 0, not create")
	}

	list := make([]dataregion.AzureRegionBatchCreate, 0, len(addSlice))
	for _, one := range addSlice {
		regionOne := dataregion.AzureRegionBatchCreate{
			CloudID:           converter.PtrToVal(one.ID),
			Name:              converter.PtrToVal(one.Name),
			Type:              one.Type,
			DisplayName:       converter.PtrToVal(one.DisplayName),
			RegionDisplayName: converter.PtrToVal(one.RegionalDisplayName),
		}
		list = append(list, regionOne)
	}

	createReq := &dataregion.AzureRegionBatchCreateReq{
		Regions: list,
	}
	_, err := cli.dbCli.Azure.Region.BatchCreateRegion(kt.Ctx, kt.Header(),
		createReq)
	if err != nil {
		logs.Errorf("[%s] create region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to create region success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateRegion(kt *kit.Kit, opt *SyncRegionOption,
	updateMap map[string]typesregion.AzureRegion) error {

	if len(updateMap) <= 0 {
		return errors.New("region updateMap is <= 0, not update")
	}

	list := make([]dataregion.AzureRegionBatchUpdate, 0, len(updateMap))

	for id, one := range updateMap {
		regionOne := dataregion.AzureRegionBatchUpdate{
			ID:   id,
			Type: one.Type,
		}
		list = append(list, regionOne)
	}

	updateReq := &dataregion.AzureRegionBatchUpdateReq{
		Regions: list,
	}
	if err := cli.dbCli.Azure.Region.BatchUpdateRegion(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] update region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync region to update region success, accountID: %s, count: %d, rid: %s", enumor.Azure,
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
		if _, exsit := delCloudMap[converter.PtrToVal(one.ID)]; exsit {
			logs.Errorf("[%s] validate region not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.Azure, opt, len(delRegionFromCloud), kt.Rid)
			return errors.New("validate region not exist failed, before delete")
		}
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &dataregion.AzureRegionBatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", parts),
		}
		if err := cli.dbCli.Azure.Region.BatchDeleteRegion(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("[%s] delete region failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to delete region success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRegionFromCloud(kt *kit.Kit, opt *SyncRegionOption) ([]typesregion.AzureRegion, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regions, err := cli.cloudCli.ListRegion(kt)
	if err != nil {
		logs.Errorf("[%s] list region from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesregion.AzureRegion, 0, len(regions))
	for _, one := range regions {
		results = append(results, converter.PtrToVal(one))
	}

	return results, nil
}

func (cli *client) listRegionFromDB(kt *kit.Kit, opt *SyncRegionOption) (
	[]coreregion.AzureRegion, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]coreregion.AzureRegion, 0)
	for {
		req.Page.Start = start
		regions, err := cli.dbCli.Azure.Region.ListRegion(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list region from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
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

func isRegionChange(cloud typesregion.AzureRegion, db coreregion.AzureRegion) bool {

	if cloud.Type != db.Type {
		return true
	}

	return false
}
