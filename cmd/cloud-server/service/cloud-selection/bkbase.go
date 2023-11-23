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
	"fmt"
	"strings"

	cssel "hcm/pkg/api/cloud-server/cloud-selection"
	"hcm/pkg/api/core"
	coresel "hcm/pkg/api/core/cloud-selection"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/bkbase"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// ListAvailableCountry 列出支持的国家
func (svc *service) ListAvailableCountry(cts *rest.Contexts) (any, error) {
	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionDistributionData,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("list available country failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	countries, err := svc.listAvailableCountry(cts.Kit)
	if err != nil {
		return nil, err
	}
	return core.ListResultT[string]{Details: countries}, nil
}

// QueryUserDistribution 查询用户分布
func (svc *service) QueryUserDistribution(cts *rest.Contexts) (any, error) {

	req := new(cssel.QueryDistReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionDistributionData,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("list available country failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	result := make([]coresel.AreaValue[float64], 0, len(req.AreaTopo))

	notBefore := times.DateAfterNow(-30)
	total := float64(0)

	for _, areaInfo := range req.AreaTopo {
		country, err := svc.listUserDistInCountry(cts.Kit, areaInfo.Name, notBefore)
		if err != nil {
			logs.Errorf("fail to query user distribution of %s, err: %v, rid: %s", areaInfo.Name, err, cts.Kit.Rid)
			return nil, err
		}
		result = append(result, coresel.AreaValue[float64]{
			Name: areaInfo.Name,
			Children: slice.Map(country, func(one coresel.UserDistribution) coresel.AreaValue[float64] {
				total += one.Count
				return coresel.AreaValue[float64]{Name: one.Province, Value: one.Count}
			}),
		})
	}

	// calculate ratio of all selected
	for i, areaValue := range result {
		for j := range areaValue.Children {
			result[i].Children[j].Value = result[i].Children[j].Value * 100 / total
		}
	}

	return result, nil
}

// QueryPingLatency 查询ping延迟数据
func (svc *service) QueryPingLatency(cts *rest.Contexts) (any, error) {

	areaTopo, idcList, err := svc.decodeAreaTopoIDCReq(cts)
	if err != nil {
		return nil, err
	}

	// auth
	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionDistributionData,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("query ping latency auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bizID, idcNames, err := getBizIDAndIDCNames(idcList)
	if err != nil {
		logs.Errorf("fail to getBizIDAndIDCNames, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.queryLatency(cts.Kit, areaTopo, svc.cfg.TableNames.LatencyPingProvinceIdc, bizID, idcNames)
}

// QueryBizLatency 查询业务延迟数据
func (svc *service) QueryBizLatency(cts *rest.Contexts) (any, error) {
	areaTopo, idcList, err := svc.decodeAreaTopoIDCReq(cts)
	if err != nil {
		return nil, err
	}

	// auth
	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionDistributionData,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("query biz latency auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bizID, idcNames, err := getBizIDAndIDCNames(idcList)
	if err != nil {
		logs.Errorf("fail to getBizIDAndIDCNames, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return svc.queryLatency(cts.Kit, areaTopo, svc.cfg.TableNames.LatencyBizProvinceIdc, bizID, idcNames)
}

// QueryServiceArea 查询机房服务区域接口
func (svc *service) QueryServiceArea(cts *rest.Contexts) (any, error) {

	topoList, idcList, err := svc.decodeAreaTopoIDCReq(cts)
	if err != nil {
		return nil, err
	}

	// auth
	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionDistributionData,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("query idc service area auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	idcNameToID := converter.SliceToMap(idcList, func(idc coresel.Idc) (string, string) {
		return idc.Name, idc.ID
	})
	notBefore := times.DateAfterNow(-30)
	bizID, idcNames, err := getBizIDAndIDCNames(idcList)
	if err != nil {
		logs.Errorf("fail to getBizIDAndIDCNames, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	idcIdToServiceArea := make(map[string][]coresel.FlatAreaInfo, len(idcList))

	// find ping latency data by top layer area
	for _, topArea := range topoList {

		provincePingIDCList, err := svc.getAvgProvincePingData(cts.Kit,
			svc.cfg.TableNames.LatencyPingProvinceIdc, topArea.Name, notBefore, bizID, idcNames)
		if err != nil {
			logs.Errorf("fail to query %s with area %s data, err: %v, rid: %s",
				svc.cfg.TableNames.LatencyPingProvinceIdc, topArea.Name, err, cts.Kit.Rid)
			return nil, err
		}

		// find the idc with the lowest latency for each province
		provinceIdcMap := map[string]coresel.ProvinceToIDCLatency{}
		for _, ppi := range provincePingIDCList {
			curIDCInfo, ok := provinceIdcMap[ppi.Province]
			if !ok || ppi.Latency <= curIDCInfo.Latency {
				provinceIdcMap[ppi.Province] = ppi
			}
		}
		// add to each idc
		for province, idcLatency := range provinceIdcMap {

			idcID := idcNameToID[idcLatency.IDCName]
			idcIdToServiceArea[idcID] = append(idcIdToServiceArea[idcID],
				coresel.FlatAreaInfo{
					CountryName:    topArea.Name,
					ProvinceName:   province,
					NetworkLatency: idcLatency.Latency,
				})
		}
	}

	// convert map to slice type
	resp := converter.MapToSlice(idcIdToServiceArea,
		func(idcID string, areas []coresel.FlatAreaInfo) coresel.IdcServiceAreaRel {
			return coresel.IdcServiceAreaRel{
				IdcID:        idcID,
				ServiceAreas: areas,
			}
		})
	return resp, nil
}

func (svc *service) decodeAreaTopoIDCReq(cts *rest.Contexts) (areaTopo []coresel.AreaInfo,
	idcList []coresel.Idc, err error) {

	req := new(cssel.AreaTopoIDCQueryReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// query idc list first
	idcResult, err := svc.client.DataService().Global.CloudSelection.ListIdc(cts.Kit, &core.ListReq{
		Filter: tools.ContainersExpression("id", req.IDCIds),
		Page:   core.NewDefaultBasePage(),
	})

	if err != nil {
		logs.Errorf("fail to query idc info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, nil, err
	}
	return req.AreaTopo, idcResult.Details, nil

}

func getBizIDAndIDCNames(idcList []coresel.Idc) (bizId int64, idcNames []string, err error) {
	bizId = -1
	idcNames = make([]string, 0, len(idcList))
	for _, idc := range idcList {
		if bizId != -1 && bizId != idc.BkBizID {

			return 0, nil, errf.Newf(errf.InvalidParameter,
				"idc ids belong to more than one biz: %d,%d", bizId, idc.BkBizID)
		}
		bizId = idc.BkBizID
		idcNames = append(idcNames, idc.Name)
	}
	return bizId, idcNames, nil
}

func (svc *service) queryLatency(kt *kit.Kit, areaTopo []coresel.AreaInfo, table string, bizId int64,

	idcNames []string) ([]cssel.MultiIdcTopo, error) {
	notBefore := times.DateAfterNow(-30)
	userDist := make([]cssel.MultiIdcTopo, 0, len(areaTopo))

	// 根据国家查询不同层级结果
	for _, topArea := range areaTopo {
		pingDataList, err := svc.getAvgProvincePingData(kt, table,
			topArea.Name, notBefore, bizId, idcNames)
		if err != nil {
			logs.Errorf("fail to query %s with area %s data, err: %v, rid: %s", table, topArea.Name, err, kt.Rid)
			return nil, err
		}

		// group by province
		provinceIdcMap := map[string][]coresel.ProvinceToIDCLatency{}
		for _, one := range pingDataList {
			provinceIdcMap[one.Province] = append(provinceIdcMap[one.Province], one)
		}

		byProvince := make([]cssel.MultiIdcTopo, 0, len(provinceIdcMap))
		for provinceName, latencies := range provinceIdcMap {
			toIdc := converter.SliceToMap(latencies,
				func(la coresel.ProvinceToIDCLatency) (k string, v float64) {
					return la.IDCName, la.Latency
				})
			byProvince = append(byProvince, cssel.MultiIdcTopo{
				Name:  provinceName,
				Value: toIdc,
			})
		}

		userDist = append(userDist, cssel.MultiIdcTopo{
			Name:     topArea.Name,
			Children: byProvince,
		})
	}
	return userDist, nil
}

func (svc *service) listAvailableCountry(kt *kit.Kit) ([]string, error) {

	yesterday := times.DateAfterNow(-3)

	sql := fmt.Sprintf("SELECT DISTINCT country FROM %s WHERE thedate='%s' LIMIT 1000",
		svc.cfg.TableNames.UserCountryDistribution, yesterday)

	countries, err := bkbase.QuerySql[coresel.CountryInfo](svc.bkBase, kt, sql)
	if err != nil {
		logs.Errorf("fail to query supported country, err: %v, date: %s, rid: %s", err, yesterday, kt.Rid)
		return nil, err
	}
	return slice.Map(countries, func(c coresel.CountryInfo) string { return c.Country }), nil
}

func (svc *service) listUserDistInCountry(kt *kit.Kit, country string, notBeforeDate string) (
	[]coresel.UserDistribution, error) {

	sql := fmt.Sprintf(`SELECT province, avg(cnt) AS count
				FROM %s
				WHERE thedate >= '%s' AND country= '%s' 
				GROUP BY province
				LIMIT 30000`,
		svc.cfg.TableNames.UserProvinceDistribution, notBeforeDate, country)
	return bkbase.QuerySql[coresel.UserDistribution](svc.bkBase, kt, sql)

}

func (svc *service) getAvgProvincePingData(kt *kit.Kit, table, country string, notBeforeDate string, idcBizId int64,
	idcNames []string) ([]coresel.ProvinceToIDCLatency, error) {

	sql := fmt.Sprintf(`SELECT province, idc_name, avg(avg_ping) AS latency 
		FROM %s
        WHERE thedate >= '%s' AND country = '%s' AND bk_biz_id = %d AND idc_name IN ('%s') 
        GROUP BY province,idc_name 
        LIMIT 30000`,
		table, notBeforeDate, country, idcBizId, strings.Join(idcNames, "','"))

	latencyList, err := bkbase.QuerySql[coresel.ProvinceToIDCLatency](svc.bkBase, kt, sql)
	if err != nil {
		logs.Errorf("fail to query province idc average ping data, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return latencyList, nil
}
