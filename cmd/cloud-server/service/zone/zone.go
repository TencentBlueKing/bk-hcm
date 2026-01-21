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

// Package zone ...
package zone

import (
	"fmt"
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	cloudproto "hcm/pkg/api/cloud-server/zone"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	dataproto "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

const (
	Azure = "azure"
	Zone1 = "1"
	Zone2 = "2"
	Zone3 = "3"
)

var AzureNoZonesMap map[string]struct{} = map[string]struct{}{
	"australiacentral":   {},
	"australiasoutheast": {},
	"japanwest":          {},
	"koreasouth":         {},
	"southindia":         {},
	"westindia":          {},
	"canadaeast":         {},
	"ukwest":             {},
	"northcentralus":     {},
	"westcentralus":      {},
	"westus":             {},
}

// InitZoneService initialize the zone service.
func InitZoneService(c *capability.Capability) {
	svc := &ZoneSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListZone", http.MethodPost, "/vendors/{vendor}/regions/{region}/zones/list", svc.ListZone)
	h.Add("ImportZone", http.MethodPost, "/vendors/{vendor}/zones/import", svc.ImportZone)
	h.Add("DeleteZone", http.MethodDelete, "/vendors/{vendor}/zones/{id}", svc.DeleteZone)
	h.Add("BatchUpdateZoneDisableCvm", http.MethodPatch, "/vendors/{vendor}/zones/disable_cvm/batch",
		svc.BatchUpdateZoneDisableCvm)

	h.Load(c.WebService)
}

// ZoneSvc zone svc
type ZoneSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListZone ...
func (dSvc *ZoneSvc) ListZone(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	region := cts.PathParameter("region").String()
	if len(region) == 0 {
		return nil, errf.New(errf.InvalidParameter, "region is required")
	}

	if vendor == enumor.Azure {
		return makeAzureZones(region)
	}

	req := new(cloudproto.ZoneListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	zoneReq := &dataproto.ZoneListReq{
		Page: req.Page,
	}

	if req.Filter != nil {
		zoneReq.Filter = req.Filter
	} else {
		zoneReq.Filter = &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		}
	}

	vendorFilter := filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor}
	regionFilter := filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region}
	zoneReq.Filter.Rules = append(zoneReq.Filter.Rules, vendorFilter)
	zoneReq.Filter.Rules = append(zoneReq.Filter.Rules, regionFilter)

	switch vendor {
	case enumor.TCloud, enumor.Aws, enumor.HuaWei, enumor.Azure, enumor.Gcp:
		// TODO 这些云厂商暂时统一用 listZone 查询明细，屏蔽extension差异，后续应支持单独的ListZoneExt接口
	case enumor.TCloudZiyan:
		return dSvc.client.DataService().TCloudZiyan.Zone.ListZoneExt(cts.Kit, zoneReq)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	return dSvc.client.DataService().Global.Zone.ListZone(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		zoneReq,
	)
}

func makeAzureZones(region string) (*dataproto.ZoneListResult, error) {
	resp := new(dataproto.ZoneListResult)
	resp.Details = []zone.BaseZone{}

	if _, ok := AzureNoZonesMap[region]; ok {
		resp.Count = 0
	} else {
		resp.Count = 3
		resp.Details = append(resp.Details, zone.BaseZone{
			ID:      "",
			Vendor:  Azure,
			CloudID: "",
			Name:    Zone1,
			Region:  region,
		})
		resp.Details = append(resp.Details, zone.BaseZone{
			ID:      "",
			Vendor:  Azure,
			CloudID: "",
			Name:    Zone2,
			Region:  region,
		})
		resp.Details = append(resp.Details, zone.BaseZone{
			ID:      "",
			Vendor:  Azure,
			CloudID: "",
			Name:    Zone3,
			Region:  region,
		})
	}

	return resp, nil
}

// ImportZone import zone from cloud to local database.
func (dSvc *ZoneSvc) ImportZone(cts *rest.Contexts) (interface{}, error) {
	vendorStr := cts.PathParameter("vendor").String()
	if len(vendorStr) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	vendor := enumor.Vendor(vendorStr)
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudproto.ZoneImportReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查是否已存在
	if err := dSvc.checkZoneExists(cts, vendor, req.Name); err != nil {
		return nil, err
	}

	// 根据 vendor 创建 zone
	switch vendor {
	case enumor.TCloud:
		return dSvc.importTCloudZone(cts, req)
	case enumor.Aws:
		return dSvc.importAwsZone(cts, req)
	case enumor.HuaWei:
		return dSvc.importHuaWeiZone(cts, req)
	case enumor.Gcp:
		return dSvc.importGcpZone(cts, req)
	case enumor.TCloudZiyan:
		return dSvc.importTCloudZiyanZone(cts, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

// checkZoneExists 检查zone是否已存在
func (dSvc *ZoneSvc) checkZoneExists(cts *rest.Contexts, vendor enumor.Vendor, name string) error {
	checkFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "name", Op: filter.Equal.Factory(), Value: name},
		},
	}
	checkReq := &dataproto.ZoneListReq{
		Filter: checkFilter,
		Page:   core.NewCountPage(),
	}
	checkResp, err := dSvc.client.DataService().Global.Zone.ListZone(cts.Kit.Ctx, cts.Kit.Header(), checkReq)
	if err != nil {
		return fmt.Errorf("check zone existence failed, err: %v", err)
	}
	if checkResp != nil && checkResp.Count > 0 {
		return errf.Newf(errf.InvalidParameter, "zone already exists with vendor: %s, name: %s", vendor, name)
	}
	return nil
}

// importTCloudZone 导入TCloud zone
func (dSvc *ZoneSvc) importTCloudZone(cts *rest.Contexts, req *cloudproto.ZoneImportReq) (interface{}, error) {
	extension := &zone.TCloudZoneExtension{}
	if req.Extension != nil {
		if extMap, ok := req.Extension.(map[string]interface{}); ok {
			if cityName, ok := extMap["city_name"].(string); ok {
				extension.CityName = cityName
			}
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
	return dSvc.client.DataService().TCloud.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
}

// importAwsZone 导入Aws zone
func (dSvc *ZoneSvc) importAwsZone(cts *rest.Contexts, req *cloudproto.ZoneImportReq) (interface{}, error) {
	extension := &zone.AwsZoneExtension{}
	createReq := &dataproto.ZoneBatchCreateReq[zone.AwsZoneExtension]{
		Zones: []dataproto.ZoneBatchCreate[zone.AwsZoneExtension]{
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
	return dSvc.client.DataService().Aws.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
}

// importHuaWeiZone 导入HuaWei zone
func (dSvc *ZoneSvc) importHuaWeiZone(cts *rest.Contexts, req *cloudproto.ZoneImportReq) (interface{}, error) {
	extension := &zone.HuaWeiZoneExtension{}
	if req.Extension != nil {
		if extMap, ok := req.Extension.(map[string]interface{}); ok {
			if port, ok := extMap["port"].(string); ok {
				extension.Port = port
			}
		}
	}
	createReq := &dataproto.ZoneBatchCreateReq[zone.HuaWeiZoneExtension]{
		Zones: []dataproto.ZoneBatchCreate[zone.HuaWeiZoneExtension]{
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
	return dSvc.client.DataService().HuaWei.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
}

// importGcpZone 导入Gcp zone
func (dSvc *ZoneSvc) importGcpZone(cts *rest.Contexts, req *cloudproto.ZoneImportReq) (interface{}, error) {
	extension := &zone.GcpZoneExtension{}
	if req.Extension != nil {
		if extMap, ok := req.Extension.(map[string]interface{}); ok {
			if selfLink, ok := extMap["self_link"].(string); ok {
				extension.SelfLink = selfLink
			}
		}
	}
	createReq := &dataproto.ZoneBatchCreateReq[zone.GcpZoneExtension]{
		Zones: []dataproto.ZoneBatchCreate[zone.GcpZoneExtension]{
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
	return dSvc.client.DataService().Gcp.Zone.BatchCreateZone(cts.Kit.Ctx, cts.Kit.Header(), createReq)
}

// DeleteZone delete zone by id.
func (dSvc *ZoneSvc) DeleteZone(cts *rest.Contexts) (interface{}, error) {
	vendorStr := cts.PathParameter("vendor").String()
	if len(vendorStr) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	// 通过 id 删除
	deleteFilter := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: id}},
	}
	deleteReq := &dataproto.ZoneBatchDeleteReq{
		Filter: deleteFilter,
	}

	err := dSvc.client.DataService().Global.Zone.BatchDeleteZone(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchUpdateZoneDisableCvm batch update zone disable_cvm field.
func (dSvc *ZoneSvc) BatchUpdateZoneDisableCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudproto.ZoneBatchUpdateDisableCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据 vendor 调用不同的更新方法
	switch vendor {
	case enumor.TCloud:
		return dSvc.batchUpdateTCloudZoneDisableCvm(cts, req)
	case enumor.TCloudZiyan:
		return dSvc.batchUpdateTCloudZiyanZoneDisableCvm(cts, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support batch update disable_cvm", vendor)
	}
}

// batchUpdateTCloudZoneDisableCvm batch update tcloud zone disable_cvm field.
func (dSvc *ZoneSvc) batchUpdateTCloudZoneDisableCvm(cts *rest.Contexts,
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
		listResp, err := dSvc.client.DataService().TCloud.Zone.ListZoneExt(cts.Kit, listReq)
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

		err := dSvc.client.DataService().TCloud.Zone.BatchUpdateZone(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
