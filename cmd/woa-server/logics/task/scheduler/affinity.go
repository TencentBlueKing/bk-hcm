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

// Package scheduler Package task affinity check logic
package scheduler

import (
	"fmt"
	"time"

	"hcm/cmd/woa-server/logics/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// AffinityService 亲和性检查服务
type AffinityService struct {
	configLogics config.Logics
	crpClient    cvmapi.CVMClientInterface
}

// NewAffinityService 创建亲和性检查服务
func NewAffinityService(configLogics config.Logics, crpClient cvmapi.CVMClientInterface) (*AffinityService, error) {
	return &AffinityService{
		configLogics: configLogics,
		crpClient:    crpClient,
	}, nil
}

// GetAffinityMatchDetail 获取亲和性匹配详情
func (s *AffinityService) GetAffinityMatchDetail(kt *kit.Kit, req *types.AffinityMatchReq) (
	*types.AffinityMatchResp, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("invalid affinity match request, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	var details []types.AffinityMatchDetail

	for _, spec := range req.Specs {
		specDetails, err := s.processSpec(kt, req.BkBizID, &spec)
		if err != nil {
			logs.Errorf("failed to process spec, err: %v, spec: %+v, rid: %s", err, spec, kt.Rid)
			return nil, err
		}
		details = append(details, specDetails...)
	}

	return &types.AffinityMatchResp{
		Details: details,
	}, nil
}

// processSpec 处理单个规格的亲和性匹配
func (s *AffinityService) processSpec(kt *kit.Kit, bizID int64, spec *types.AffinityMatchSpec) (
	[]types.AffinityMatchDetail, error) {
	var details []types.AffinityMatchDetail

	// 检查是否有Campus分区
	hasCampus := spec.IsCVMSeparateCampus()

	if hasCampus {
		campusList, err := s.getCampusList(kt, spec)
		if err != nil {
			return nil, err
		}

		for _, campus := range campusList {
			detail, err := s.matchAffinityForCampus(kt, bizID, spec, campus)
			if err != nil {
				logs.Errorf("failed to match affinity for campus, err: %v, campus: %+v, rid: %s", err, campus, kt.Rid)
				return nil, err
			}
			details = append(details, detail)
		}
		return details, nil
	}

	for _, zone := range spec.Zones {
		detail, err := s.matchAffinityForZone(kt, bizID, spec, zone)
		if err != nil {
			logs.Errorf("failed to match affinity for zone, err: %v, zone: %s, rid: %s", err, zone, kt.Rid)
			return nil, err
		}
		details = append(details, detail)
	}

	return details, nil
}

// getCampusList 获取园区列表
func (s *AffinityService) getCampusList(kt *kit.Kit, spec *types.AffinityMatchSpec) ([]types.CampusInfo, error) {
	cond := mapstr.MapStr{}

	if spec.IsCVMSeparateCampus() {
		if spec.Region == "" {
			return []types.CampusInfo{}, nil
		}
		cond["region"] = mapstr.MapStr{pkg.BKDBIN: []string{spec.Region}}

		var campusZones []string
		for _, zone := range spec.Zones {
			if zone == cvmapi.CvmSeparateCampus || zone == cvmapi.CvmZoneAll {
				continue
			}
			campusZones = append(campusZones, zone)
		}
		if len(campusZones) > 0 {
			cond["zone"] = mapstr.MapStr{pkg.BKDBIN: campusZones}
		}
	} else {
		if len(spec.Zones) > 0 {
			cond["zone"] = mapstr.MapStr{pkg.BKDBIN: spec.Zones}
		}
	}

	zoneResult, err := s.configLogics.Zone().GetZone(kt, &cond)
	if err != nil {
		logs.Errorf("failed to get zone list from config, cond: %+v, err: %v, rid: %s", cond, err, kt.Rid)
		return nil, err
	}

	var campusList []types.CampusInfo
	for _, zone := range zoneResult.Info {
		campusList = append(campusList, types.CampusInfo{
			CampusID:   zone.Zone,
			CampusName: zone.ZoneCn,
			Zone:       zone.Zone,
		})
	}
	return campusList, nil
}

// matchAffinityForCampus 为园区匹配亲和性
func (s *AffinityService) matchAffinityForCampus(kt *kit.Kit, bizID int64, spec *types.AffinityMatchSpec,
	campus types.CampusInfo) (types.AffinityMatchDetail, error) {
	crpReq := &cvmapi.MatchSwapGroupReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      fmt.Sprintf("hcm-affinity-%d", time.Now().UnixNano()),
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmMatchSwapGroupMethod,
		},
		Params: cvmapi.MatchSwapGroupParams{
			DeptID:       bizID,
			ApplyNum:     spec.Replicas,
			Zone:         campus.Zone,
			InstanceType: spec.DeviceType,
		},
	}

	crpResp, err := s.crpClient.MatchSwapGroup(kt.Ctx, kt.Header(), crpReq)
	if err != nil {
		traceID := getCRPMatchTraceID(crpResp)
		logs.Errorf("failed to call CRP matchSwapGroup, bizID: %d, zone: %s, deviceType: %s, err: %v, "+
			"crpTraceID: %s, rid: %s", bizID, campus.Zone, spec.DeviceType, err, traceID, kt.Rid)
		return types.AffinityMatchDetail{}, err
	}
	if crpResp == nil {
		logs.Errorf("CRP matchSwapGroup returns nil resp, bizID: %d, zone: %s, deviceType: %s, rid: %s",
			bizID, campus.Zone, spec.DeviceType, kt.Rid)
		return types.AffinityMatchDetail{}, fmt.Errorf("matchSwapGroup returns nil resp")
	}

	if !crpResp.Result.Matched {
		return types.AffinityMatchDetail{
			Region:     spec.Region,
			Zone:       campus.Zone,
			DeviceType: spec.DeviceType,
			Replicas:   spec.Replicas,
			Status:     enumor.AffinityStatusNoData,
			MaxCutNum:  0,
			IPs:        []string{},
		}, nil
	}

	queryReq := &cvmapi.QueryMatchTaskReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      fmt.Sprintf("hcm-query-%d", time.Now().UnixNano()),
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmQueryMatchTaskMethod,
		},
		Params: cvmapi.QueryMatchTaskParams{
			OrderID: crpResp.Result.MatchID,
		},
	}

	queryResp, err := s.crpClient.QueryMatchTask(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		traceID := getCRPQueryTraceID(queryResp)
		logs.Errorf("failed to call CRP queryMatchTask, bizID: %d, zone: %s, deviceType: %s, err: %v, "+
			"crpTraceID: %s, rid: %s", bizID, campus.Zone, spec.DeviceType, err, traceID, kt.Rid)
		return types.AffinityMatchDetail{}, err
	}
	if queryResp == nil {
		logs.Errorf("CRP queryMatchTask returns nil resp, bizID: %d, zone: %s, deviceType: %s, matchID: %s, rid: %s",
			bizID, campus.Zone, spec.DeviceType, crpResp.Result.MatchID, kt.Rid)
		return types.AffinityMatchDetail{}, fmt.Errorf("queryMatchTask returns nil resp")
	}

	status := enumor.AffinityStatusNoData
	if queryResp.Result.Status {
		status = enumor.AffinityStatusHasData
	}

	return types.AffinityMatchDetail{
		Region:     spec.Region,
		Zone:       campus.Zone,
		DeviceType: spec.DeviceType,
		Replicas:   spec.Replicas,
		Status:     status,
		MaxCutNum:  queryResp.Result.MaxCutNum,
		IPs:        queryResp.Result.IPs,
	}, nil
}

// matchAffinityForZone 为可用区匹配亲和性
func (s *AffinityService) matchAffinityForZone(kt *kit.Kit, bizID int64, spec *types.AffinityMatchSpec,
	zone string) (types.AffinityMatchDetail, error) {
	crpReq := &cvmapi.MatchSwapGroupReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      fmt.Sprintf("hcm-affinity-%d", time.Now().UnixNano()),
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmMatchSwapGroupMethod,
		},
		Params: cvmapi.MatchSwapGroupParams{
			DeptID:       bizID,
			ApplyNum:     spec.Replicas,
			Zone:         zone,
			InstanceType: spec.DeviceType,
		},
	}

	crpResp, err := s.crpClient.MatchSwapGroup(kt.Ctx, kt.Header(), crpReq)
	if err != nil {
		traceID := getCRPMatchTraceID(crpResp)
		logs.Errorf("failed to call CRP matchSwapGroup, bizID: %d, zone: %s, deviceType: %s, err: %v, "+
			"crpTraceID: %s, rid: %s", bizID, zone, spec.DeviceType, err, traceID, kt.Rid)
		return types.AffinityMatchDetail{}, err
	}
	if crpResp == nil {
		return types.AffinityMatchDetail{}, fmt.Errorf("matchSwapGroup returns nil resp")
	}

	if !crpResp.Result.Matched {
		return types.AffinityMatchDetail{
			Region:     spec.Region,
			Zone:       zone,
			DeviceType: spec.DeviceType,
			Replicas:   spec.Replicas,
			Status:     enumor.AffinityStatusNoData,
			MaxCutNum:  0,
			IPs:        []string{},
		}, nil
	}

	queryReq := &cvmapi.QueryMatchTaskReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      fmt.Sprintf("hcm-query-%d", time.Now().UnixNano()),
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmQueryMatchTaskMethod,
		},
		Params: cvmapi.QueryMatchTaskParams{
			OrderID: crpResp.Result.MatchID,
		},
	}

	queryResp, err := s.crpClient.QueryMatchTask(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		traceID := getCRPQueryTraceID(queryResp)
		logs.Errorf("failed to call CRP queryMatchTask, bizID: %d, zone: %s, deviceType: %s, err: %v, "+
			"crpTraceID: %s, rid: %s", bizID, zone, spec.DeviceType, err, traceID, kt.Rid)
		return types.AffinityMatchDetail{}, err
	}
	if queryResp == nil {
		return types.AffinityMatchDetail{}, fmt.Errorf("queryMatchTask returns nil resp")
	}

	status := enumor.AffinityStatusNoData
	if queryResp.Result.Status {
		status = enumor.AffinityStatusHasData
	}

	return types.AffinityMatchDetail{
		Region:     spec.Region,
		Zone:       zone,
		DeviceType: spec.DeviceType,
		Replicas:   spec.Replicas,
		Status:     status,
		MaxCutNum:  queryResp.Result.MaxCutNum,
		IPs:        queryResp.Result.IPs,
	}, nil
}

func getCRPMatchTraceID(resp *cvmapi.MatchSwapGroupResp) string {
	if resp == nil {
		return ""
	}
	return resp.TraceId
}

func getCRPQueryTraceID(resp *cvmapi.QueryMatchTaskResp) string {
	if resp == nil {
		return ""
	}
	return resp.TraceId
}
