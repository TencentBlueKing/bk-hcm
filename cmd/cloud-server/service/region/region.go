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

// Package region ...
package region

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"hcm/cmd/cloud-server/service/capability"
	protoregion "hcm/pkg/api/cloud-server/region"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	dataprotoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitRegionService initialize the region service.
func InitRegionService(c *capability.Capability) {
	svc := &RegionSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListRegion", http.MethodPost, "/vendors/{vendor}/regions/list", svc.ListRegion)
	h.Add("ImportRegions", http.MethodPost, "/vendors/{vendor}/regions/import", svc.ImportRegions)
	h.Add("DeleteRegions", http.MethodDelete, "/vendors/{vendor}/regions/batch", svc.DeleteRegions)

	h.Load(c.WebService)
}

// RegionSvc region svc
type RegionSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListRegion ...
func (svc *RegionSvc) ListRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(protoregion.RegionListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reqPage := &core.BasePage{
		Count: req.Page.Count,
		Start: req.Page.Start,
		Limit: req.Page.Limit,
	}
	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   reqPage,
	}
	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	case enumor.Aws:
		return svc.client.DataService().Aws.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	case enumor.Azure:
		return svc.client.DataService().Azure.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	case enumor.TCloudZiyan:
		return svc.client.DataService().TCloudZiyan.Region.ListRegion(cts.Kit, listReq)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

// ImportRegions batch import regions to local database.
func (svc *RegionSvc) ImportRegions(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(protoregion.RegionImportReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloudZiyan:
		return svc.TCloudZiyanImportRegions(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

// TCloudZiyanImportRegions batch import regions to local database.
func (svc *RegionSvc) TCloudZiyanImportRegions(kt *kit.Kit, req *protoregion.RegionImportReq) (interface{}, error) {
	// 收集所有 region_id 用于批量检查
	regionIDs := make([]string, 0, len(req.Regions))
	for _, r := range req.Regions {
		regionIDs = append(regionIDs, r.RegionID)
	}

	// 批量检查是否已存在（通过 vendor + region_id）
	// 批量导入有上限，不需要分页查询
	checkReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleIn("region_id", regionIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	checkResp, err := svc.client.DataService().TCloudZiyan.Region.ListRegion(kt, checkReq)
	if err != nil {
		return nil, fmt.Errorf("check region existence failed, err: %v", err)
	}
	if checkResp == nil {
		return nil, errors.New("check region existence failed, resp is nil")
	}

	// 如果存在重复的 region，返回错误
	if len(checkResp.Details) > 0 {
		existRegionIDs := make([]string, 0, len(checkResp.Details))
		for _, r := range checkResp.Details {
			existRegionIDs = append(existRegionIDs, r.RegionID)
		}
		return nil, errf.Newf(errf.InvalidParameter, "regions already exist with region_ids: %v", existRegionIDs)
	}

	// 构建批量创建请求，设置 source 为 manually
	regions := make([]dataprotoregion.TCloudRegionBatchCreate, 0, len(req.Regions))
	for _, r := range req.Regions {
		regions = append(regions, dataprotoregion.TCloudRegionBatchCreate{
			Vendor:     enumor.TCloudZiyan,
			RegionID:   r.RegionID,
			RegionName: r.RegionName,
			AreaName:   extractAreaName(r.RegionName),
			Status:     r.Status,
			Source:     enumor.RegionSourceManually,
		})
	}

	createReq := &dataprotoregion.TCloudRegionCreateReq{
		Regions: regions,
	}

	return svc.client.DataService().TCloudZiyan.Region.BatchCreate(kt, createReq)
}

// extractAreaName 从 region_name 中提取 area_name
// 例如：从 "华南地区(广州)" 提取 "华南地区"
func extractAreaName(regionName string) string {
	if len(regionName) == 0 {
		return ""
	}

	// 查找左括号的位置
	idx := strings.Index(regionName, "(")
	if idx > 0 {
		// 如果找到左括号，返回括号前面的部分
		return strings.TrimSpace(regionName[:idx])
	}

	// 如果没有找到括号，返回原始字符串
	return regionName
}

// DeleteRegions batch delete regions by ids.
func (svc *RegionSvc) DeleteRegions(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(protoregion.RegionDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloudZiyan:
		return svc.TCloudZiyanDeleteRegions(cts.Kit, req.IDs)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

// TCloudZiyanDeleteRegions batch delete regions by ids.
func (svc *RegionSvc) TCloudZiyanDeleteRegions(kt *kit.Kit, ids []string) (interface{}, error) {
	// 通过 ids 批量删除
	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", ids),
	}

	err := svc.client.DataService().TCloudZiyan.Region.BatchDelete(kt, deleteReq)
	if err != nil {
		logs.Errorf("batch delete tcloud-ziyan region failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}
