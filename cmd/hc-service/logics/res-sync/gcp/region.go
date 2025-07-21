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

package gcp

import (
	"errors"
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	typescore "hcm/pkg/adaptor/types/core"
	adaptorregion "hcm/pkg/adaptor/types/region"
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
}

// Validate ...
func (opt SyncRegionOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Region ...
func (cli *client) Region(kt *kit.Kit, params *SyncBaseParams, opt *SyncRegionOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regionFromCloud, err := cli.listRegionFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	regionFromDB, err := cli.listRegionFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(regionFromCloud) == 0 && len(regionFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesregion.GcpRegion, cloudcore.GcpRegion](
		regionFromCloud, regionFromDB, isRegionChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteRegion(kt, params.AccountID, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createRegion(kt, params.AccountID, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateRegion(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) listRegionFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]adaptorregion.GcpRegion, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescore.GcpListOption{
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
		CloudIDs: params.CloudIDs,
	}
	result, err := cli.cloudCli.ListRegion(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list region from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listRegionFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.GcpRegion, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Gcp,
				},
				&filter.AtomRule{
					Field: "region_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list region from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) createRegion(kt *kit.Kit, accountID string, addSlice []typesregion.GcpRegion) error {
	if len(addSlice) <= 0 {
		return errors.New("region addSlice is <= 0, not create")
	}

	createResources := make([]dataregion.GcpRegionBatchCreate, 0, len(addSlice))

	for _, one := range addSlice {
		tmpRes := dataregion.GcpRegionBatchCreate{
			Vendor:     enumor.Gcp,
			RegionID:   one.RegionID,
			RegionName: one.RegionName,
			Status:     one.RegionState,
			SelfLink:   one.SelfLink,
		}
		createResources = append(createResources, tmpRes)
	}

	// 底层单次操作，最大支持100个地域
	elems := slice.Split(createResources, constant.BatchOperationMaxLimit)
	for _, parts := range elems {
		createReq := &dataregion.GcpRegionCreateReq{
			Regions: parts,
		}
		if _, err := cli.dbCli.Gcp.Region.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("[%s] create region failed, err: %v, account: %s, rid: %s", enumor.Gcp,
				err, accountID, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to create region success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateRegion(kt *kit.Kit, accountID string,
	updateMap map[string]typesregion.GcpRegion) error {

	if len(updateMap) <= 0 {
		return errors.New("region updateMap is <= 0, not update")
	}

	updateResources := make([]dataregion.GcpRegionBatchUpdate, 0, len(updateMap))

	for id, one := range updateMap {
		tmpRes := dataregion.GcpRegionBatchUpdate{
			ID:         id,
			RegionID:   one.RegionID,
			RegionName: one.RegionName,
			Status:     one.RegionState,
		}
		updateResources = append(updateResources, tmpRes)
	}

	// 底层单次操作，最大支持100个地域
	elems := slice.Split(updateResources, constant.BatchOperationMaxLimit)
	for _, parts := range elems {
		updateReq := &dataregion.GcpRegionBatchUpdateReq{
			Regions: parts,
		}
		if err := cli.dbCli.Gcp.Region.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("[%s] update region failed, err: %v, account: %s, rid: %s", enumor.Gcp,
				err, accountID, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to update region success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteRegion(kt *kit.Kit, accountID string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return errors.New("region delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delRegionFromCloud, err := cli.listRegionFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delRegionFromCloud) > 0 {
		logs.Errorf("[%s] validate region not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delRegionFromCloud), kt.Rid)
		return fmt.Errorf("validate region not exist failed, before delete")
	}

	// 底层单次操作，最大支持100个地域
	elems := slice.Split(delCloudIDs, constant.BatchOperationMaxLimit)
	for _, parts := range elems {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("region_id", parts),
		}
		if err := cli.dbCli.Gcp.Region.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("[%s] delete region failed, err: %v, account: %s, rid: %s", enumor.Gcp,
				err, accountID, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync region to delete region success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) RemoveRegionDeleteFromCloud(kt *kit.Kit, accountID string) error {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: enumor.Gcp},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Gcp.Region.ListRegion(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list region failed, err: %v, req: %v, rid: %s", enumor.Gcp,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.RegionID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listRegionFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.RegionID)
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteRegion(kt, accountID, cloudIDs); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func isRegionChange(cloud typesregion.GcpRegion, db cloudcore.GcpRegion) bool {

	if cloud.RegionID != db.RegionID {
		return true
	}

	if cloud.RegionName != db.RegionName {
		return true
	}

	if cloud.RegionState != db.Status {
		return true
	}

	return false
}
