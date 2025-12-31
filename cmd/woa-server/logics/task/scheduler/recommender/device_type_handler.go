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

// Package recommender define the device type modification recommend handler
package recommender

import (
	"errors"
	"sort"

	configLogics "hcm/cmd/woa-server/logics/config"
	cfgtype "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/util"
)

// DeviceTypeHandler apply order device type modification recommend handler
type DeviceTypeHandler struct {
	handler      Handler
	cvm          cvmapi.CVMClientInterface
	configLogics configLogics.Logics
}

// DeviceInfoList device list
type DeviceInfoList []*cfgtype.DeviceInfo

// Len returns list length
func (l DeviceInfoList) Len() int {
	return len(l)
}

// Swap swaps two items in the list
func (l DeviceInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Less compares two items
func (l DeviceInfoList) Less(i, j int) bool {
	return l[i].Score < l[j].Score
}

// Handle make modification recommendation of apply order
func (dh *DeviceTypeHandler) Handle(order *types.ApplyOrder) (*Recommendation, bool) {
	kt := kit.New()
	// get candidate device
	candidates, err := dh.getCandidateDevice(kt, order)
	if err != nil {
		logs.Errorf("failed to ge candidate device")
		return nil, false
	}

	// sort candidate device
	sortCandidates := DeviceInfoList{}
	sortCandidates = append(sortCandidates, candidates...)
	sort.Sort(sort.Reverse(sortCandidates))

	// cvm_separate_campus pick candidate with the highest score
	if order.Spec.IsCVMSeparateCampus() {
		rst := &Recommendation{
			Zone:       order.Spec.Zone,
			DeviceType: sortCandidates[0].DeviceType,
		}

		return rst, true
	}

	// query capacity and pick satisfied candidate device
	for _, candidate := range sortCandidates {
		capNum, err := dh.getCapacity(kt, candidate.RequireType, candidate.DeviceType, candidate.Region,
			candidate.Zone, order.Spec.ChargeType)
		if err != nil {
			logs.Warnf("failed to get device capacity, err: %v", err)
			continue
		}

		if capNum < int64(order.PendingNum) {
			logs.Warnf("device capacity %d less than need %d", capNum, order.PendingNum)
			continue
		}

		rst := &Recommendation{
			Zone:       order.Spec.Zone,
			DeviceType: candidate.DeviceType,
		}

		return rst, true
	}

	return nil, false
}

func (dh *DeviceTypeHandler) getCandidateDevice(kt *kit.Kit, order *types.ApplyOrder) ([]*cfgtype.DeviceInfo, error) {
	curDevice, err := dh.getCurDevice(kt, order)
	if err != nil {
		logs.Errorf("failed to get device info, err: %v", err)
		return nil, err
	}

	deviceGroup, ok := curDevice.Label["device_group"]
	if !ok {
		logs.Errorf("failed to get current device group")
		return nil, errors.New("failed to get current device group")
	}

	rules := []querybuilder.Rule{
		querybuilder.AtomRule{
			Field:    "require_type",
			Operator: querybuilder.OperatorEqual,
			Value:    order.RequireType,
		},
		querybuilder.AtomRule{
			Field:    "region",
			Operator: querybuilder.OperatorEqual,
			Value:    order.Spec.Region,
		},
		// with the same device group
		querybuilder.AtomRule{
			Field:    "label.device_group",
			Operator: querybuilder.OperatorEqual,
			Value:    deviceGroup,
		},
		// with the same hardware configuration
		querybuilder.AtomRule{
			Field:    "cpu",
			Operator: querybuilder.OperatorEqual,
			Value:    curDevice.Cpu,
		},
		querybuilder.AtomRule{
			Field:    "mem",
			Operator: querybuilder.OperatorEqual,
			Value:    curDevice.Mem,
		},
	}

	// with the same zone
	if order.Spec.Zone != "" && order.Spec.Zone != cvmapi.CvmSeparateCampus {
		rules = append(rules, querybuilder.AtomRule{
			Field:    "zone",
			Operator: querybuilder.OperatorEqual,
			Value:    order.Spec.Zone,
		})
	}

	param := &cfgtype.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules:     rules,
			},
		},
		Page: metadata.BasePage{
			Limit: pkg.BKNoLimit,
			Start: 0,
		},
	}

	rst, err := dh.configLogics.Device().GetDevice(kt, param)
	if err != nil {
		logs.Errorf("failed to get device info, err: %v", err)
		return nil, err
	}

	if len(rst.Info) == 0 {
		logs.Errorf("get no device info")
		return nil, errors.New("get no device info")
	}

	return rst.Info, nil
}

func (dh *DeviceTypeHandler) getCurDevice(kt *kit.Kit, order *types.ApplyOrder) (*cfgtype.DeviceInfo, error) {
	rules := []querybuilder.Rule{
		querybuilder.AtomRule{
			Field:    "device_type",
			Operator: querybuilder.OperatorEqual,
			Value:    order.Spec.DeviceType,
		},
		querybuilder.AtomRule{
			Field:    "require_type",
			Operator: querybuilder.OperatorEqual,
			Value:    order.RequireType,
		},
		querybuilder.AtomRule{
			Field:    "region",
			Operator: querybuilder.OperatorEqual,
			Value:    order.Spec.Region,
		},
	}

	if order.Spec.Zone != "" && order.Spec.Zone != cvmapi.CvmSeparateCampus {
		rules = append(rules, querybuilder.AtomRule{
			Field:    "zone",
			Operator: querybuilder.OperatorEqual,
			Value:    order.Spec.Zone,
		})
	}

	param := &cfgtype.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules:     rules,
			},
		},
		Page: metadata.BasePage{
			Limit: 1,
			Start: 0,
		},
	}

	rst, err := dh.configLogics.Device().GetDevice(kt, param)
	if err != nil {
		logs.Errorf("failed to get device info, err: %v", err)
		return nil, err
	}

	if len(rst.Info) == 0 {
		logs.Errorf("get no device info")
		return nil, errors.New("get no device info")
	}

	return rst.Info[0], nil
}

// getAvailableZoneInfo get available cvm zone info
func (dh *DeviceTypeHandler) getAvailableZoneInfo(kt *kit.Kit, requireType int64, deviceType, region string) (
	[]*cfgtype.Zone, error) {

	allZones, err := dh.getZoneList(kt, region)
	if err != nil {
		return nil, err
	}

	availZoneIds, err := dh.getAvailableZoneIds(kt, requireType, deviceType, region)
	if err != nil {
		return nil, err
	}

	availZones := make([]*cfgtype.Zone, 0)
	for _, zone := range allZones {
		for _, zoneId := range availZoneIds {
			if zone.Zone == zoneId {
				availZones = append(availZones, zone)
				break
			}
		}
	}

	return availZones, nil
}

// getAvailableZoneIds get available cvm zone id
func (dh *DeviceTypeHandler) getAvailableZoneIds(kt *kit.Kit, requireType int64, deviceType, region string) (
	[]string, error) {

	param := &cfgtype.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					&querybuilder.AtomRule{
						Field:    "require_type",
						Operator: querybuilder.OperatorEqual,
						Value:    requireType,
					},
					&querybuilder.AtomRule{
						Field:    "device_type",
						Operator: querybuilder.OperatorEqual,
						Value:    deviceType,
					},
					&querybuilder.AtomRule{
						Field:    "region",
						Operator: querybuilder.OperatorEqual,
						Value:    region,
					},
				},
			},
		},
	}
	zoneResp, err := dh.configLogics.Device().GetDevice(kt, param)
	if err != nil {
		return nil, err
	}

	zoneIds := make([]string, 0)
	for _, device := range zoneResp.Info {
		zoneIds = append(zoneIds, device.Zone)
	}

	zoneIds = util.StrArrayUnique(zoneIds)
	return zoneIds, nil
}

// getZoneList get zone info in certain region
func (dh *DeviceTypeHandler) getZoneList(kt *kit.Kit, region string) ([]*cfgtype.Zone, error) {
	req := &cfgtype.GetZoneParam{}
	// if input region is empty list, return all zone info
	if len(region) > 0 {
		req.Region = []string{region}
	}
	zoneResp, err := dh.configLogics.Zone().GetZone(kt, req)
	if err != nil {
		return nil, err
	}

	return zoneResp.Info, nil
}

// getCapacity get resource apply capacity info
func (dh *DeviceTypeHandler) getCapacity(kt *kit.Kit, requireType enumor.RequireType, deviceType, region, zone string,
	chargeType cvmapi.ChargeType) (int64, error) {

	param := &cfgtype.GetCapacityParam{
		RequireType: requireType,
		DeviceType:  deviceType,
		Region:      region,
		Zone:        zone,
	}
	// 计费模式,默认包年包月
	if len(chargeType) > 0 {
		param.ChargeType = chargeType
	}

	rst, err := dh.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return 0, err
	}

	num := len(rst.Info)
	if num == 0 {
		logs.Errorf("failed to get capacity, for return empty info")
		return 0, errors.New("failed to get capacity, for return empty info")
	}

	return rst.Info[0].MaxNum, nil
}
