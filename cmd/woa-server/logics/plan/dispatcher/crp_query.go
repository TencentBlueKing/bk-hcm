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

package dispatcher

import (
	"errors"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// queryTransferCRPDemands 查询CRP中转产品中可转移的预测，按照项目类型和技术大类
func (c *CrpTicketCreator) queryTransferCRPDemands(kt *kit.Kit, obsProjects []enumor.ObsProject,
	technicalClasses []string) error {

	crpDemands, err := c.resFetcher.QueryCRPTransferPoolDemands(kt, obsProjects, technicalClasses)
	if err != nil {
		logs.Errorf("failed to query transfer pool demands, err: %v, obs_project: %v, technical_classes: %v, rid: %s",
			err, obsProjects, technicalClasses, kt.Rid)
		return err
	}

	for _, demand := range crpDemands {
		c.transferAbleDemands[demand.Year] = append(c.transferAbleDemands[demand.Year], demand)
	}
	return nil
}

// queryAdjustAbleDemands query demands that can be adjusted.
// TODO 在splitter 和 dispatcher 中都调用，可以抽象为一个公共函数
func (c *CrpTicketCreator) queryAdjustAbleDemands(kt *kit.Kit, req *ptypes.AdjustAbleDemandsReq) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// init request parameter.
	queryReq := &cvmapi.CvmCbsAdjustAblePlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsAdjustAblePlanQueryMethod,
		},
		Params: convAdjustAbleQueryParam(req),
	}

	rst, err := c.crpCli.QueryAdjustAbleDemand(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, *queryReq.Params,
			kt.Rid)
		return nil, err
	}

	if rst.Error.Code != 0 {
		logs.Errorf("failed to query adjust able demands, err: %v, crp_trace: %s, rid: %s", rst.Error.Message,
			rst.TraceId, kt.Rid)
		return nil, errors.New(rst.Error.Message)
	}

	if len(rst.Result.Data) == 0 {
		logs.Errorf("failed to query adjust able demands, return is empty, crp_trace: %s, rid: %s",
			rst.TraceId, kt.Rid)
		return nil, errors.New("no demands can be adjusted in CRP")
	}

	return rst.Result.Data, nil
}

func convAdjustAbleQueryParam(req *ptypes.AdjustAbleDemandsReq) *cvmapi.CvmCbsAdjustAblePlanQueryParam {
	reqParams := new(cvmapi.CvmCbsAdjustAblePlanQueryParam)

	if len(req.RegionName) > 0 {
		reqParams.CityName = req.RegionName
	}

	if len(req.DeviceFamily) > 0 {
		reqParams.InstanceFamily = req.DeviceFamily
	}

	if len(req.DeviceType) > 0 {
		reqParams.InstanceModel = req.DeviceType
	}

	if len(req.ExpectTime) > 0 {
		reqParams.UseTime = req.ExpectTime
	}

	if len(req.PlanProductName) > 0 {
		reqParams.PlanProductName = req.PlanProductName
	}

	if len(req.OpProductName) > 0 {
		reqParams.ProductName = req.OpProductName
	}

	if len(req.ObsProject) > 0 {
		reqParams.ProjectName = string(req.ObsProject)
	}

	if len(req.DiskType) > 0 {
		// TODO 因CRP会在拆单时改变云盘的类型，且云盘不会参与模糊匹配，因此查询时暂时不考虑diskType参数.
		// reqParams.DiskTypeName = req.DiskType.Name()
	}

	if len(req.ResMode) > 0 {
		reqParams.ResourceMode = string(req.ResMode)
	}

	return reqParams
}
