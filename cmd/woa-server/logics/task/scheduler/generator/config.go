/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package generator generate task
package generator

import (
	"errors"
	"math"

	cfgtype "hcm/cmd/woa-server/types/config"
	"hcm/cmd/woa-server/types/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// getAvailableZoneInfo get available cvm zone info
func (g *Generator) getAvailableZoneInfo(kt *kit.Kit, requireType enumor.RequireType, deviceType, region string) (
	map[string]*cfgtype.Zone, error) {

	allZones, err := g.getZoneList(kt, region)
	if err != nil {
		return nil, err
	}

	availZoneIds, err := g.getAvailableZoneIds(kt, deviceType, region)
	if err != nil {
		return nil, err
	}

	availZonesMap := make(map[string]*cfgtype.Zone, 0)
	for _, zone := range allZones {
		for _, zoneId := range availZoneIds {
			if zone.Zone == zoneId {
				availZonesMap[zone.Zone] = zone
				break
			}
		}
	}

	return availZonesMap, nil
}

// getAvailableZoneIds get available cvm zone id
func (g *Generator) getAvailableZoneIds(kt *kit.Kit, deviceType, region string) ([]string, error) {
	zoneMap := make(map[string]struct{})
	req := &protocloud.DeviceTypeListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", enumor.TCloudZiyan), tools.RuleEqual("region", region),
				tools.RuleEqual("device_type", deviceType), tools.RuleEqual("disable", false),
			),
			Page: core.NewDefaultBasePage(),
		},
	}
	for {
		resp, err := g.configLogics.Device().ListDeviceType(kt, req)
		if err != nil {
			logs.Errorf("failed to list device type, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, deviceType := range resp.Details {
			zoneMap[deviceType.Zone] = struct{}{}
		}
		if len(resp.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return maps.Keys(zoneMap), nil
}

// getZoneList get zone info in certain region
func (g *Generator) getZoneList(kt *kit.Kit, region string) ([]*cfgtype.Zone, error) {
	req := &cfgtype.GetZoneParam{}
	// if input region is empty list, return all zone info
	if len(region) > 0 {
		req.Region = []string{region}
	}
	zoneResp, err := g.configLogics.Zone().GetZone(kt, req)
	if err != nil {
		return nil, err
	}

	return zoneResp.Info, nil
}

// getRegionList get region list by zone list
func (g *Generator) getRegionList(kt *kit.Kit, zoneList []string) ([]*cfgtype.Zone, error) {
	req := &cfgtype.GetZoneParam{}
	// if input is empty list, return all zone info
	if len(zoneList) > 0 {
		req.Zone = zoneList
	}
	zoneResp, err := g.configLogics.Zone().GetZone(kt, req)
	if err != nil {
		return nil, err
	}

	return zoneResp.Info, nil
}

// getCapacity get resource apply capacity info
func (g *Generator) getCapacity(kt *kit.Kit, order *task.ApplyOrder, zone, vpc, subnet string, orderZones []string) (
	map[string]int64, error) {

	// 小额绿通不需要查询库存
	if order.RequireType.NotNeedVerifyCapacity() {
		// 不是分Campus的话，直接返回
		if len(zone) > 0 && zone != cvmapi.CvmSeparateCampus {
			return map[string]int64{zone: math.MaxInt}, nil
		}
		// 是分Campus的话，返回所有zone
		if zone == cvmapi.CvmSeparateCampus && len(orderZones) > 0 {
			return slice.FuncToMap(orderZones, func(zone string) (string, int64) { return zone, math.MaxInt }), nil
		}
		return map[string]int64{}, nil
	}

	param := &cfgtype.GetCapacityParam{
		RequireType:      order.RequireType,
		DeviceType:       order.Spec.DeviceType,
		Region:           order.Spec.Region,
		Zone:             zone,
		Vpc:              vpc,
		Subnet:           subnet,
		IgnorePrediction: !order.RequireType.NeedVerifyResPlan(),
		BizID:            order.BkBizId,
	}
	// 计费模式,默认包年包月
	if len(order.Spec.ChargeType) > 0 {
		param.ChargeType = order.Spec.ChargeType
	}

	rst, err := g.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return nil, err
	}

	zoneCapacity := make(map[string]int64)
	for _, capInfo := range rst.Info {
		zoneCapacity[capInfo.Zone] = capInfo.MaxNum
	}

	return zoneCapacity, nil
}

// getCapacityDetail get resource apply capacity detail info
func (g *Generator) getCapacityDetail(kt *kit.Kit, order *types.ApplyOrder, zone, vpc, subnet string) (
	[]*cfgtype.CapacityInfo, error) {

	param := &cfgtype.GetCapacityParam{
		RequireType: order.RequireType,
		DeviceType:  order.Spec.DeviceType,
		Region:      order.Spec.Region,
		Zone:        zone,
		Vpc:         vpc,
		Subnet:      subnet,
		BizID:       order.BkBizId,
	}
	// 计费模式,默认包年包月
	if len(order.Spec.ChargeType) > 0 {
		param.ChargeType = order.Spec.ChargeType
	}

	rst, err := g.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return nil, err
	}

	if len(rst.Info) == 0 {
		return nil, errors.New("get no capacity info")
	}

	logs.Infof("get order zone capacity detail, subOrderID: %s, zone: %s, vpc: %s, subnet: %s, param: %+v, "+
		"rstList: %+v, rid: %s", order.SubOrderId, zone, vpc, subnet, cvt.PtrToVal(param),
		cvt.PtrToSlice(rst.Info), kt.Rid)

	return rst.Info, nil
}
