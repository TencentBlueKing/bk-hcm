/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"errors"

	"hcm/cmd/woa-server/types/meta"
	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	protozone "hcm/pkg/api/data-service/cloud/zone"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	dtypes "hcm/pkg/dal/dao/types/meta"
	imeta "hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/maps"
)

// ListDiskType lists disk type.
func (s *service) ListDiskType(_ *rest.Contexts) (interface{}, error) {
	// get disk type members.
	diskTypes := enumor.GetDiskTypeMembers()
	// convert to meta.DiskTypeItem slice.
	details := make([]meta.DiskTypeItem, 0, len(diskTypes))
	for _, diskType := range diskTypes {
		details = append(details, meta.DiskTypeItem{
			DiskType:     diskType,
			DiskTypeName: diskType.Name(),
		})
	}
	return &core.ListResultT[meta.DiskTypeItem]{Details: details}, nil
}

// ListObsProject lists obs project.
func (s *service) ListObsProject(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.ObsProject]{Details: enumor.GetObsProjectMembersForResPlan()}, nil
}

// ListRegion lists region.
func (s *service) ListRegion(cts *rest.Contexts) (interface{}, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
		Page:   core.NewDefaultBasePage(),
	}

	// 从 region 表获取 region 列表
	details := make([]dtypes.RegionElem, 0)
	for {
		regions, err := s.client.DataService().TCloudZiyan.Region.ListRegion(cts.Kit, req)
		if err != nil {
			logs.Errorf("failed to get region list, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}

		for _, region := range regions.Details {
			details = append(details, dtypes.RegionElem{
				RegionID:   region.RegionID,
				RegionName: region.CityName,
			})
		}

		if len(regions.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return &core.ListResultT[dtypes.RegionElem]{Details: details}, nil
}

// ListZone lists zone.
func (s *service) ListZone(cts *rest.Contexts) (interface{}, error) {
	req := new(mtypes.ListZoneReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list zone parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 构建查询条件
	rules := []filter.RuleFactory{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
	}
	if len(req.RegionIDs) > 0 {
		rules = append(rules, tools.RuleIn("region", req.RegionIDs))
	}

	zoneReq := &protozone.ZoneListReq{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: core.NewDefaultBasePage(),
	}

	// 遍历获取zones
	details := make([]dtypes.ZoneElem, 0)
	for {
		zones, err := s.client.DataService().TCloudZiyan.Zone.ListZoneExt(cts.Kit, zoneReq)
		if err != nil {
			logs.Errorf("failed to get zone list, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}

		// 转换为 ZoneElem 格式
		for _, zone := range zones.Details {
			details = append(details, dtypes.ZoneElem{
				ZoneID:   zone.Name,
				ZoneName: zone.NameCn,
			})
		}

		if uint(len(zones.Details)) < zoneReq.Page.Limit {
			break
		}
		zoneReq.Page.Start += uint32(zoneReq.Page.Limit)
	}

	return &core.ListResultT[dtypes.ZoneElem]{Details: details}, nil
}

// ListDeviceClass lists region.
func (s *service) ListDeviceClass(cts *rest.Contexts) (interface{}, error) {
	// 分页获取所有设备类型数据，然后提取去重的 device_class
	deviceClassSet := make(map[string]struct{})
	page := core.NewDefaultBasePage()
	for {
		req := &protocloud.DistinctDeviceTypeListReq{
			ListReq: core.ListReq{
				Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
				Page:   page,
			},
		}
		result, err := s.client.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(cts.Kit, req)
		if err != nil {
			logs.Errorf("failed to list device type, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}
		// 提取 device_class 并去重
		for _, detail := range result.Details {
			if detail.DeviceClass != "" {
				deviceClassSet[detail.DeviceClass] = struct{}{}
			}
		}
		if len(result.Details) < int(page.Limit) {
			break
		}
		page.Start += uint32(page.Limit)
	}

	return &core.ListResultT[string]{Details: maps.Keys(deviceClassSet)}, nil
}

// ListDeviceType lists device type.
func (s *service) ListDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(mtypes.ListDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := []*filter.AtomRule{tools.RuleEqual("vendor", enumor.TCloudZiyan)}
	if len(req.DeviceClasses) > 0 {
		opt = append(opt, tools.RuleIn("device_class", req.DeviceClasses))
	}
	if len(req.DeviceTypes) > 0 {
		opt = append(opt, tools.RuleIn("device_type", req.DeviceTypes))
	}
	deviceDetails := make([]mtypes.ListDeviceTypeRst, 0)
	listReq := &protocloud.DistinctDeviceTypeListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(opt...),
			Page:   core.NewDefaultBasePage(),
		},
	}
	for {
		result, err := s.client.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("failed to list device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}
		for _, detail := range result.Details {
			deviceDetails = append(deviceDetails, mtypes.ListDeviceTypeRst{
				DeviceType:   detail.DeviceType,
				CoreType:     string(detail.CoreType),
				CpuCore:      detail.CpuCore,
				Memory:       detail.Memory,
				DeviceClass:  detail.DeviceClass,
				DeviceFamily: detail.DeviceFamily,
			})
		}
		if len(result.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return &core.ListResultT[mtypes.ListDeviceTypeRst]{Details: deviceDetails}, nil
}

// ListPlanType lists plan type.
func (s *service) ListPlanType(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.PlanType]{Details: enumor.GetPlanTypeHcmMembers()}, nil
}

// ListTicketType lists ticket type.
func (s *service) ListTicketType(_ *rest.Contexts) (interface{}, error) {
	// get ticket type members.
	ticketTypes := enumor.GetRPTicketTypeMembers()
	// convert to meta.DiskTypeItem slice.
	details := make([]meta.TicketTypeItem, len(ticketTypes))
	for idx, ticketType := range ticketTypes {
		details[idx] = meta.TicketTypeItem{
			TicketType:     ticketType,
			TicketTypeName: ticketType.Name(),
		}
	}
	return &core.ListResultT[meta.TicketTypeItem]{Details: details}, nil
}

// ListBizsByOpProduct lists bizs by op product.
func (s *service) ListBizsByOpProduct(cts *rest.Contexts) (interface{}, error) {
	req := new(mtypes.ListBizsByOpProdReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list bizs by op product, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list bizs by op product parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizs, err := s.logics.GetBizsByOpProd(cts.Kit, req.OpProductID)
	if err != nil {
		logs.Errorf("failed to get bizs by op product, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[mtypes.Biz]{Details: bizs}, nil
}

// ListOpProducts lists op products.
func (s *service) ListOpProducts(cts *rest.Contexts) (interface{}, error) {
	opProds, err := s.logics.GetOpProducts(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get op products, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[mtypes.OpProduct]{Details: opProds}, nil
}

// ListPlanProducts lists plan products.
func (s *service) ListPlanProducts(cts *rest.Contexts) (interface{}, error) {
	planProds, err := s.logics.GetPlanProducts(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get plan products, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[mtypes.PlanProduct]{Details: planProds}, nil
}

// ListResourcePoolBiz list all resource pool biz. no authorized
func (s *service) ListResourcePoolBiz(cts *rest.Contexts) (any, error) {
	listRst := new(rsproto.ResourcePoolBusinessListResult)
	listReq := &rsproto.ResourcePoolBusinessListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		res, err := s.client.DataService().Global.RollingServer.ListResPoolBiz(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("list resource pool business failed, err: %v, req: %+v, rid: %s", err, *listReq, cts.Kit.Rid)
			return nil, err
		}

		listRst.Details = append(listRst.Details, res.Details...)

		if len(res.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return listRst, nil
}

// DeleteResourcePoolBiz delete resource pool biz. need authorized
func (s *service) DeleteResourcePoolBiz(cts *rest.Contexts) (any, error) {
	delID := cts.PathParameter("id").String()
	if delID == "" {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("id can't be empty"))
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, imeta.ResourceAttribute{
		Basic: &imeta.Basic{Type: imeta.RollingServerManage, Action: imeta.Find}})
	if err != nil {
		logs.Errorf("delete resource pool business failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.BatchDeleteReq{
		Filter: tools.EqualExpression("id", delID),
	}
	return nil, s.client.DataService().Global.RollingServer.DeleteResPoolBiz(cts.Kit, delReq)
}

// CreateResourcePoolBiz create resource pool biz. need authorized
func (s *service) CreateResourcePoolBiz(cts *rest.Contexts) (any, error) {
	req := new(rsproto.ResourcePoolBusinessCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, imeta.ResourceAttribute{
		Basic: &imeta.Basic{Type: imeta.RollingServerManage, Action: imeta.Find}})
	if err != nil {
		logs.Errorf("create resource pool business failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.client.DataService().Global.RollingServer.BatchCreateResPoolBiz(cts.Kit, req)
}

// ListOrgTopos list org topos.
func (s *service) ListOrgTopos(cts *rest.Contexts) (any, error) {
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, imeta.ResourceAttribute{Basic: &imeta.Basic{
		Type: imeta.ServiceResDissolve, Action: imeta.Find}}); err != nil {
		logs.Errorf("no permission to get org topo resource, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(mtypes.OrgTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("get org topo decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("get org topo request validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ret, err := s.logics.GetOrgTopo(cts.Kit, req.View)
	if err != nil {
		logs.Errorf("failed to get org topo, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return ret, nil
}

// ListRequireObsProject lists require obs project.
func (s *service) ListRequireObsProject(_ *rest.Contexts) (any, error) {
	return enumor.RequireTypeObsProjectMap, nil
}
