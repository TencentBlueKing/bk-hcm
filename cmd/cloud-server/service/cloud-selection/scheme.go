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

package csselection

import (
	"errors"
	"fmt"

	"hcm/cmd/cloud-server/plugin/recommend"
	csselection "hcm/pkg/api/cloud-server/cloud-selection"
	"hcm/pkg/api/core"
	coresel "hcm/pkg/api/core/cloud-selection"
	dsselection "hcm/pkg/api/data-service/cloud-selection"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/plugin"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/bkbase"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// BatchDeleteScheme ..
func (svc *service) BatchDeleteScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reses := make([]meta.ResourceAttribute, 0, len(req.IDs))
	for _, one := range req.IDs {
		reses = append(reses, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.CloudSelectionScheme,
				Action:     meta.Delete,
				ResourceID: one,
			},
		})
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, reses...); err != nil {
		logs.Errorf("batch delete scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := svc.client.DataService().Global.CloudSelection.BatchDeleteScheme(cts.Kit, req); err != nil {
		logs.Errorf("call dataservice to batch delete scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateScheme ...
func (svc *service) CreateScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(csselection.SchemeCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionScheme,
			Action: meta.Create,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("create scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	vendors, err := svc.getIdcVendorByIDs(cts.Kit, req.ResultIdcIDs)
	if err != nil {
		logs.Errorf("get idc vendor by ids failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &dsselection.SchemeCreateReq{
		BkBizID:                req.BkBizID,
		Name:                   req.Name,
		BizType:                req.BizType,
		Vendors:                vendors,
		DeploymentArchitecture: req.DeploymentArchitecture,
		CoverPing:              req.CoverPing,
		CompositeScore:         req.CompositeScore,
		NetScore:               req.NetScore,
		CostScore:              req.CostScore,
		CoverRate:              req.CoverRate,
		UserDistribution:       req.UserDistribution,
		ResultIdcIDs:           req.ResultIdcIDs,
	}
	result, err := svc.client.DataService().Global.CloudSelection.CreateScheme(cts.Kit, createReq)
	if err != nil {
		logs.Errorf("call dataservice to create scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 尝试注册创建者权限
	regCreatorReq := &meta.RegisterResCreatorActionInst{
		Type: meta.CloudSelectionScheme.String(),
		ID:   result.ID,
		Name: req.Name,
	}

	if err := svc.authorizer.RegisterResourceCreatorAction(cts.Kit, regCreatorReq); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateScheme ...
func (svc *service) UpdateScheme(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errors.New("id is required")
	}

	req := new(csselection.SchemeUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.CloudSelectionScheme,
			Action:     meta.Update,
			ResourceID: id,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("update scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &dsselection.SchemeUpdateReq{
		Name:    req.Name,
		BkBizID: req.BkBizID,
	}
	if err := svc.client.DataService().Global.CloudSelection.UpdateScheme(cts.Kit, id, updateReq); err != nil {
		logs.Errorf("call dataservice to update scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetScheme ...
func (svc *service) GetScheme(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errors.New("id is required")
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.CloudSelectionScheme,
			Action:     meta.Find,
			ResourceID: id,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("get scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.CloudSelection.ListScheme(cts.Kit, req)
	if err != nil {
		logs.Errorf("call dataservice to list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, fmt.Errorf("scheme: %s not found", id)
	}

	return result.Details[0], nil
}

// ListScheme ...
func (svc *service) ListScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := &meta.ListAuthResInput{
		Type:   meta.CloudSelectionScheme,
		Action: meta.Find,
	}
	expr, noPermFlag, err := svc.authorizer.ListAuthInstWithFilter(cts.Kit, res, req.Filter, "id")
	if err != nil {
		logs.Errorf("list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	result, err := svc.client.DataService().Global.CloudSelection.ListScheme(cts.Kit, req)
	if err != nil {
		logs.Errorf("call dataservice to list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// GenerateRecommendScheme 推荐选型方案
func (svc *service) GenerateRecommendScheme(cts *rest.Contexts) (any, error) {
	req := new(csselection.GenSchemeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{Basic: &meta.Basic{
		Type:   meta.CloudSelectionScheme,
		Action: meta.Create,
	}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("generate scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// idc 列表
	idcResp, err := svc.client.DataService().Global.CloudSelection.ListIdc(cts.Kit,
		&core.ListReq{Page: core.NewDefaultBasePage(), Filter: tools.AllExpression()})
	if err != nil {
		logs.Errorf("fail to list IDC, err: %v", err)
		return nil, err
	}
	idcByID := converter.SliceToMap(idcResp.Details, func(idc coresel.Idc) (string, coresel.Idc) { return idc.ID, idc })

	// 前端输入转算法输入
	algIn, err := svc.buildALgIn(cts, req, idcByID)
	if err != nil {
		return nil, err
	}
	algPlugin, err := plugin.NewPlugin[recommend.AlgorithmInput, recommend.AlgorithmOutput](
		svc.cfg.AlgorithmPlugin.BinaryPath, svc.cfg.AlgorithmPlugin.Args...)
	if err != nil {
		logs.Errorf("init algorithm plugin fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("init algorithm plugin fail")
	}

	algOut, err := algPlugin.Execute(cts.Kit, algIn)
	if err != nil {
		logs.Errorf("fail to execute algorithm plugin, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// converter result
	return slice.Map(algOut.ParetoList, func(s recommend.Solution) csselection.GeneratedSchemeResult {
		return csselection.GeneratedSchemeResult{
			CoverRate:      s.CoverRate,
			CompositeScore: (s.F1Score + s.F2Score) / 2,
			AvgPing:        s.F1,
			NetScore:       s.F1Score,
			CostScore:      s.F2Score,
			ResultIdcIds:   s.Idc,
			Vendors:        slice.Unique(slice.Map(s.Idc, func(id string) enumor.Vendor { return idcByID[id].Vendor })),
			// only distributed is supported
			DeploymentArchitecture: enumor.Distributed,
		}
	}), nil

}

func (svc *service) buildALgIn(cts *rest.Contexts, req *csselection.GenSchemeReq,
	idcByID map[string]coresel.Idc) (*recommend.AlgorithmInput, error) {

	idcPriceMap := make(map[string]float64, len(idcByID))
	usedIdcIds := make([]string, 0, len(idcByID))
	idcByName := make(map[string]coresel.Idc, len(idcByID))
	var idcBizID int64 = -1
	for _, idc := range idcByID {
		price, ok := svc.cfg.DefaultIdcPrice[idc.Vendor]
		if !ok {
			continue
		}
		usedIdcIds = append(usedIdcIds, idc.ID)
		// TODO: use user input
		idcBizID = idc.BkBizID
		idcPriceMap[idc.ID] = price
		idcByName[idc.Name] = idc
	}

	// 人口分布和ping数据
	startDate := bkbase.DateBefore(svc.cfg.AvgLatencySampleDays)
	userDistribution := map[string]float64{}
	pingInfo := make(map[string]map[string]float64, len(req.UserDistributions))

	allProvinceData, err := svc.listAllAvgProvincePingData(cts.Kit, svc.getRecommendDataSource(), startDate, idcBizID,
		converter.MapKeyToStringSlice(idcByName))
	if err != nil {
		logs.Errorf("fail to get avg ping data, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	for _, area := range req.UserDistributions {
		// ping数据
		pingData, exists := allProvinceData[area.Name]
		if !exists {
			return nil, errf.Newf(errf.InvalidParameter, "wrong country: %s", area.Name)
		}
		for _, pd := range pingData {
			name := getCombinedKey(area.Name, pd.Province)
			if _, exists := pingInfo[name]; !exists {
				pingInfo[name] = make(map[string]float64, len(idcByID))
			}
			pingInfo[name][idcByName[pd.IDCName].ID] = pd.Latency
		}
		// 人口数据
		for _, p := range area.Children {
			name := getCombinedKey(area.Name, p.Name)
			userDistribution[name] = p.Value
		}
	}

	algIn := &recommend.AlgorithmInput{
		CountryRate:     userDistribution,
		CoverRate:       svc.cfg.CoverRate,
		CoverPing:       req.CoverPing,
		PingInfo:        pingInfo,
		IdcPrice:        idcPriceMap,
		IdcList:         usedIdcIds,
		CoverPingRanges: svc.cfg.CoverPingRanges,
		IDCPriceRanges:  svc.cfg.IDCPriceRanges,
		BanIdcList:      []string{},
		PickIdcList:     []string{},
	}
	return algIn, nil
}

// 拼接唯一key main-sub
func getCombinedKey(main, sub string) string {
	return main + "-" + sub
}

func (svc *service) getRecommendDataSource() string {
	tableName := svc.cfg.TableNames.RecommendDataSource
	if tableName == "" {
		tableName = svc.cfg.TableNames.LatencyPingProvinceIdc
	}
	return tableName
}
