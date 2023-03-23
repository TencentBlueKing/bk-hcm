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
	typescore "hcm/pkg/adaptor/types/core"
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

// SyncTCloudZone sync all zone
func (z *zoneHC) SyncTCloudZone(cts *rest.Contexts) (interface{}, error) {

	req := new(apizone.TCloudZoneSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := z.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeszone.TCloudZoneListOption{
		Region: req.Region,
	}

	zones, err := client.ListZone(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	cloudMap := make(map[string]*TCloudZoneSync)
	cloudIDs := make([]string, 0, len(zones))
	for _, zone := range zones {
		cloudMap[*zone.ZoneId] = &TCloudZoneSync{IsUpdate: false, Zone: zone}
		cloudIDs = append(cloudIDs, *zone.ZoneId)
		cloudAllIDs[*zone.ZoneId] = true
	}

	updateIDs, dsMap, err := z.getTCloudZoneDSSync(cts, cloudIDs, req)
	if err != nil {
		logs.Errorf("request getTCloudZoneDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := z.syncTCloudZoneUpdate(cts, updateIDs, cloudMap, dsMap)
		if err != nil {
			logs.Errorf("request syncTCloudZoneUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		err := z.syncTCloudZoneAdd(addIDs, cts, req, cloudMap)
		if err != nil {
			logs.Errorf("request syncTCloudZoneAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := z.getTCloudZoneAllDS(cts, req)
	if err != nil {
		logs.Errorf("request getTCloudZoneAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
			logs.Errorf("request adaptor to list tcloud zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

func (z *zoneHC) syncTCloudZoneUpdate(cts *rest.Contexts, updateIDs []string, cloudMap map[string]*TCloudZoneSync,
	dsMap map[string]*DSZoneSync) error {

	list := make([]protozone.ZoneBatchUpdate[zone.TCloudZoneExtension], 0, len(updateIDs))
	for _, id := range updateIDs {
		if *cloudMap[id].Zone.ZoneState == dsMap[id].Zone.State {
			continue
		}

		one := protozone.ZoneBatchUpdate[zone.TCloudZoneExtension]{
			ID:    dsMap[id].Zone.ID,
			State: *cloudMap[id].Zone.ZoneState,
		}
		list = append(list, one)
	}

	updateReq := &protozone.ZoneBatchUpdateReq[zone.TCloudZoneExtension]{
		Zones: list,
	}

	if len(updateReq.Zones) > 0 {
		if err := z.dataCli.TCloud.Zone.BatchUpdateZone(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateZone failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	return nil
}

func (z *zoneHC) getTCloudZoneAllDS(cts *rest.Contexts, req *apizone.TCloudZoneSyncReq) ([]string, error) {

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
						Value: enumor.TCloud,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: typescore.TCloudQueryLimit},
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

func (z *zoneHC) syncTCloudZoneAdd(addIDs []string, cts *rest.Contexts, req *apizone.TCloudZoneSyncReq,
	cloudMap map[string]*TCloudZoneSync) error {

	list := make([]protozone.ZoneBatchCreate[zone.TCloudZoneExtension], 0, len(addIDs))
	for _, id := range addIDs {
		one := protozone.ZoneBatchCreate[zone.TCloudZoneExtension]{
			CloudID:   *cloudMap[id].Zone.ZoneId,
			Name:      *cloudMap[id].Zone.Zone,
			State:     *cloudMap[id].Zone.ZoneState,
			Region:    req.Region,
			NameCn:    *cloudMap[id].Zone.ZoneName,
			Extension: &zone.TCloudZoneExtension{},
		}
		list = append(list, one)
	}

	createReq := &protozone.ZoneBatchCreateReq[zone.TCloudZoneExtension]{
		Zones: list,
	}
	_, err := z.dataCli.TCloud.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return err
	}

	return nil
}

func (z *zoneHC) getTCloudZoneDSSync(cts *rest.Contexts, cloudIDs []string,
	req *apizone.TCloudZoneSyncReq) ([]string, map[string]*DSZoneSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*DSZoneSync)

	if len(cloudIDs) <= 0 {
		return updateIDs, dsMap, nil
	}

	start := 0
	for {
		dataReq := &protozone.ZoneListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.TCloud,
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
			Page: &core.BasePage{Start: uint32(start), Limit: typescore.TCloudQueryLimit},
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
