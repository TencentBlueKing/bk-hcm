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

package aws

import (
	"errors"
	"fmt"

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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncZoneOption ...
type SyncZoneOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate ...
func (opt SyncZoneOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Zone ...
func (cli *client) Zone(kt *kit.Kit, opt *SyncZoneOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	zoneFromCloud, err := cli.listZoneFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	zoneFromDB, err := cli.listZoneFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(zoneFromCloud) == 0 && len(zoneFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeszone.AwsZone, corezone.BaseZone](
		zoneFromCloud, zoneFromDB, isZoneChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteZone(kt, opt, delCloudIDs); err != nil {
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

func (cli *client) createZone(kt *kit.Kit, opt *SyncZoneOption, addSlice []typeszone.AwsZone) error {
	if len(addSlice) <= 0 {
		return errors.New("zone addSlice is <= 0, not create")
	}

	list := make([]datazone.ZoneBatchCreate[zone.AwsZoneExtension], 0, len(addSlice))

	for _, one := range addSlice {
		zoneOne := datazone.ZoneBatchCreate[zone.AwsZoneExtension]{
			CloudID:   converter.PtrToVal(one.ZoneId),
			Name:      converter.PtrToVal(one.ZoneName),
			State:     converter.PtrToVal(one.State),
			Region:    converter.PtrToVal(one.RegionName),
			Extension: &zone.AwsZoneExtension{},
		}
		list = append(list, zoneOne)
	}

	createReq := &datazone.ZoneBatchCreateReq[zone.AwsZoneExtension]{
		Zones: list,
	}
	_, err := cli.dbCli.Aws.Zone.BatchCreateZone(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] create zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync zone to create zone success, accountID: %s, region: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, opt.Region, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateZone(kt *kit.Kit, opt *SyncZoneOption,
	updateMap map[string]typeszone.AwsZone) error {

	if len(updateMap) <= 0 {
		return errors.New("zone updateMap is <= 0, not update")
	}

	list := make([]datazone.ZoneBatchUpdate[zone.AwsZoneExtension], 0, len(updateMap))

	for id, one := range updateMap {
		one := datazone.ZoneBatchUpdate[zone.AwsZoneExtension]{
			ID:    id,
			State: converter.PtrToVal(one.State),
		}
		list = append(list, one)
	}

	updateReq := &datazone.ZoneBatchUpdateReq[zone.AwsZoneExtension]{
		Zones: list,
	}
	if err := cli.dbCli.Aws.Zone.BatchUpdateZone(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] update zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync zone to update zone success, accountID: %s, region: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, opt.Region, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteZone(kt *kit.Kit, opt *SyncZoneOption, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return errors.New("zone delCloudIDs is <= 0, not delete")
	}

	delZoneFromCloud, err := cli.listZoneFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delZoneFromCloud {
		if _, exsit := delCloudMap[converter.PtrToVal(one.ZoneId)]; exsit {
			logs.Errorf("[%s] validate zone not exist failed, before delete, opt: %v, exist zone id: %s, "+
				"del cloud ids: %v, rid: %s", enumor.Aws, opt, converter.PtrToVal(one.ZoneId), delCloudIDs, kt.Rid)
			return fmt.Errorf("validate zone not exist failed, before delete")
		}
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &datazone.ZoneBatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", parts),
		}

		err := cli.dbCli.Global.Zone.BatchDeleteZone(kt.Ctx, kt.Header(), deleteReq)
		if err != nil {
			logs.Errorf("[%s] delete zone failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync zone to delete zone success, accountID: %s, region: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, opt.Region, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listZoneFromCloud(kt *kit.Kit, opt *SyncZoneOption) ([]typeszone.AwsZone, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	zoneOpt := &typeszone.AwsZoneListOption{
		Region: opt.Region,
	}
	results, err := cli.cloudCli.ListZone(kt, zoneOpt)
	if err != nil {
		logs.Errorf("[%s] list zone from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results, nil
}

func (cli *client) listZoneFromDB(kt *kit.Kit, opt *SyncZoneOption) (
	[]corezone.BaseZone, error) {

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
					Value: enumor.Aws,
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
	results := make([]corezone.BaseZone, 0)
	for {
		req.Page.Start = start
		zones, err := cli.dbCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list zone from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws, err,
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

func isZoneChange(cloud typeszone.AwsZone, db corezone.BaseZone) bool {

	if converter.PtrToVal(cloud.State) != db.State {
		return true
	}

	return false
}
