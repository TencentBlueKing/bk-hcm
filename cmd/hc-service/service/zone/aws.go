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
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	protozone "hcm/pkg/api/data-service/cloud/zone"
	apizone "hcm/pkg/api/hc-service/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// SyncAwsZone sync all zone
func (z *zoneHC) SyncAwsZone(cts *rest.Contexts) (interface{}, error) {

	req := new(apizone.AwsZoneSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := z.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeszone.AwsZoneListOption{
		Region: req.Region,
	}

	zones, err := client.ListZone(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list aws zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	cloudMap := make(map[string]*AwsZoneSync)
	cloudIDs := make([]string, 0, len(zones))
	for _, zone := range zones {
		cloudMap[*zone.ZoneId] = &AwsZoneSync{IsUpdate: false, Zone: zone}
		cloudIDs = append(cloudIDs, *zone.ZoneId)
		cloudAllIDs[*zone.ZoneId] = true
	}

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*DSZoneSync)

	start := 0
	step := int(filter.DefaultMaxInLimit)
	for {
		var tmpCloudIDs []string
		if start+step > len(cloudIDs) {
			tmpCloudIDs = make([]string, len(cloudIDs)-start)
			copy(tmpCloudIDs, cloudIDs[start:])
		} else {
			tmpCloudIDs = make([]string, step)
			copy(tmpCloudIDs, cloudIDs[start:start+step])
		}

		if len(tmpCloudIDs) > 0 {
			tmpIDs, tmpMap, err := z.getAwsZoneDSSync(cts, cloudIDs[start:], req)
			if err != nil {
				logs.Errorf("request getAwsZoneDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}

			updateIDs = append(updateIDs, tmpIDs...)
			for k, v := range tmpMap {
				dsMap[k] = v
			}
		}

		start = start + step
		if start > len(cloudIDs) {
			break
		}
	}

	if len(updateIDs) > 0 {
		err := z.syncAwsZoneUpdate(cts, updateIDs, cloudMap, dsMap)
		if err != nil {
			logs.Errorf("request syncAwsZoneUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	addIDs := make([]string, 0)
	for _, id := range updateIDs {
		if _, ok := cloudMap[id]; ok {
			cloudMap[id].IsUpdate = true
		}
	}
	for k, v := range cloudMap {
		if !v.IsUpdate {
			addIDs = append(addIDs, k)
		}
	}

	if len(addIDs) > 0 {
		err := z.syncAwsZoneAdd(addIDs, cts, req, cloudMap)
		if err != nil {
			logs.Errorf("request syncAwsZoneAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := z.getAwsZoneAllDS(cts, req)
	if err != nil {
		logs.Errorf("request getAwsZoneAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		zones, err := client.ListZone(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		for _, id := range deleteIDs {
			realDeleteFlag := true
			for _, zone := range zones {
				if *zone.ZoneId == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		err = z.syncZoneDelete(cts, realDeleteIDs)
		if err != nil {
			logs.Errorf("request syncZoneDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (z *zoneHC) getAwsZoneAllDS(cts *rest.Contexts, req *apizone.AwsZoneSyncReq) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)

	for {
		dataReq := &protozone.ZoneListReq{
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
						Value: req.Region,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		results, err := z.dataCli.Global.Zone.ListZone(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}

func (z *zoneHC) syncAwsZoneAdd(addIDs []string, cts *rest.Contexts, req *apizone.AwsZoneSyncReq,
	cloudMap map[string]*AwsZoneSync) error {

	list := make([]protozone.ZoneBatchCreate[zone.AwsZoneExtension], 0, len(addIDs))
	for _, id := range addIDs {
		one := protozone.ZoneBatchCreate[zone.AwsZoneExtension]{
			CloudID:   id,
			Name:      *cloudMap[id].Zone.ZoneName,
			State:     *cloudMap[id].Zone.State,
			Region:    *cloudMap[id].Zone.RegionName,
			Extension: &zone.AwsZoneExtension{},
		}
		list = append(list, one)
	}

	createReq := &protozone.ZoneBatchCreateReq[zone.AwsZoneExtension]{
		Zones: list,
	}
	_, err := z.dataCli.Aws.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return err
	}

	return nil
}

func (z *zoneHC) syncAwsZoneUpdate(cts *rest.Contexts, updateIDs []string, cloudMap map[string]*AwsZoneSync,
	dsMap map[string]*DSZoneSync) error {

	list := make([]protozone.ZoneBatchUpdate[zone.AwsZoneExtension], 0, len(updateIDs))
	for _, id := range updateIDs {
		if *cloudMap[id].Zone.State == dsMap[id].Zone.State {
			continue
		}

		one := protozone.ZoneBatchUpdate[zone.AwsZoneExtension]{
			ID:    dsMap[id].Zone.ID,
			State: *cloudMap[id].Zone.State,
		}
		list = append(list, one)
	}

	updateReq := &protozone.ZoneBatchUpdateReq[zone.AwsZoneExtension]{
		Zones: list,
	}

	if len(updateReq.Zones) > 0 {
		if err := z.dataCli.Aws.Zone.BatchUpdateZone(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateZone failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	return nil
}

func (z *zoneHC) getAwsZoneDSSync(cts *rest.Contexts, cloudIDs []string,
	req *apizone.AwsZoneSyncReq) ([]string, map[string]*DSZoneSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*DSZoneSync)

	start := 0
	for {
		dataReq := &protozone.ZoneListReq{
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
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "cloud_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		results, err := z.dataCli.Global.Zone.ListZone(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsZoneSync := new(DSZoneSync)
				dsZoneSync.Zone = detail
				dsMap[detail.CloudID] = dsZoneSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, dsMap, nil
}
