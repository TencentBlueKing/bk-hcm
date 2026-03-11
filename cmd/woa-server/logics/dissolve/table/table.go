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

package table

import (
	"errors"
	"fmt"
	"sync"

	"hcm/cmd/woa-server/logics/config"
	dissolveconfig "hcm/cmd/woa-server/logics/dissolve/config"
	logicshost "hcm/cmd/woa-server/logics/dissolve/host"
	logicsmodule "hcm/cmd/woa-server/logics/dissolve/module"
	model "hcm/cmd/woa-server/model/task"
	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/condition"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/es"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// Table provides interface for operations of dissolve table.
type Table interface {
	FindOriginHost(kt *kit.Kit, req *dissolve.HostListReq, source ReqSourceI) (*dissolve.ListHostDetails, error)
	FindCurHost(kt *kit.Kit, req *dissolve.HostListReq, source ReqSourceI) (*dissolve.ListHostDetails, error)
	ListResDissolveTable(kt *kit.Kit, req *dissolve.ResDissolveReq) ([]dissolve.BizDetail, error)
	ListBizCpuCoreSummary(kt *kit.Kit, bizIDs []int64) (map[int64]dissolve.CpuCoreSummary, error)
}

type logics struct {
	recycledHost   logicshost.RecycledHost
	recycledModule logicsmodule.RecycledModule
	dissolveConfig dissolveconfig.Config
	configLogics   config.Logics
	cmdbCli        cmdb.Client
	esCli          *es.EsCli
	originDate     string
	blacklist      string
}

// New create resource dissolve table logics.
func New(recycledModule logicsmodule.RecycledModule, recycledHost logicshost.RecycledHost,
	dissolveConfig dissolveconfig.Config, configLogics config.Logics, cmdbCli cmdb.Client, esCli *es.EsCli,
	originDate string, blacklist string) Table {

	return &logics{
		recycledHost:   recycledHost,
		recycledModule: recycledModule,
		dissolveConfig: dissolveConfig,
		configLogics:   configLogics,
		cmdbCli:        cmdbCli,
		esCli:          esCli,
		originDate:     originDate,
		blacklist:      blacklist,
	}
}

// FindOriginHost find origin host
func (l *logics) FindOriginHost(kt *kit.Kit, req *dissolve.HostListReq, source ReqSourceI) (
	*dissolve.ListHostDetails, error) {

	assetIDs, err := l.getAssetIDByModule(kt, req.ModuleNames, false)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if len(assetIDs) == 0 {
		return &dissolve.ListHostDetails{}, nil
	}

	bizIDName, err := l.getBizIDNameByName(kt, req.BizNames, make([]int64, 0))
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if len(bizIDName) == 0 {
		return &dissolve.ListHostDetails{}, nil
	}

	blackBizIDName, err := l.getBlackBizIDName(kt)
	if err != nil {
		logs.Errorf("get black biz ids failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	cond, err := req.GetESCond(assetIDs, bizIDName, blackBizIDName)
	if err != nil {
		logs.Errorf("get es cond failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	res, err := l.findHostFromES(kt, cond, l.getOriginHostIndex(), req.Page, source.GetEsHostFields())
	if err != nil {
		logs.Errorf("find host from es failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return res, nil
	}

	originHosts := res.Details
	res.Details, err = l.getCurBizName(kt, originHosts)
	if err != nil {
		logs.Errorf("get host current biz name failed, err: %v, host: %+v, rid: %s", err, originHosts, kt.Rid)
		return nil, err
	}

	return res, nil
}

func (l *logics) getOriginHostIndex() string {
	return es.GetIndex(l.originDate)
}

func (l *logics) getCurBizName(kt *kit.Kit, hosts []dissolve.Host) ([]dissolve.Host, error) {
	if len(hosts) == 0 {
		return make([]dissolve.Host, 0), nil
	}

	hostBizIDMap := make(map[string]int64)
	bizIDs := make([]int64, 0)
	for _, host := range hosts {
		hostBizIDMap[host.ServerAssetID] = host.BizID
		bizIDs = append(bizIDs, host.BizID)
	}

	bizIDName, err := l.getBizIDNameByID(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, ids: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	for i, host := range hosts {
		bizID, ok := hostBizIDMap[host.ServerAssetID]
		if !ok {
			logs.Errorf("can not find biz id, host asset id: %s, map: %+v, rid: %s", host.ServerAssetID, hostBizIDMap,
				kt.Rid)
			return nil, errors.New("host is invalid")
		}

		bizName, ok := bizIDName[bizID]
		if !ok {
			logs.Errorf("can not find biz name, biz id: %d, map: %+v,rid: %s", bizID, bizIDName, kt.Rid)
			return nil, errors.New("biz is invalid")
		}

		hosts[i].AppName = bizName
	}

	return hosts, nil
}

// FindCurHost find current host
func (l *logics) FindCurHost(kt *kit.Kit, req *dissolve.HostListReq, source ReqSourceI) (
	*dissolve.ListHostDetails, error) {

	assetIDs, err := l.getAssetIDByModule(kt, req.ModuleNames, true)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(assetIDs) == 0 {
		return &dissolve.ListHostDetails{Details: []dissolve.Host{}}, nil
	}

	// 1.由于有些条件值不在cc的主机字段上，所以先根据主机上有的字段，查出host id
	conds := req.GetCCHostCond(assetIDs)
	originHostIDs, err := l.getAllHostIDFromCC(kt, conds)
	if err != nil {
		logs.Errorf("get host id from cc failed, err: %v, cond: %+v, rid: %s", err, conds, kt.Rid)
		return nil, err
	}

	if len(originHostIDs) == 0 {
		return &dissolve.ListHostDetails{Details: []dissolve.Host{}}, nil
	}

	// 2.根据业务条件筛选host id
	bizIDName, err := l.getBizIDNameByName(kt, req.BizNames, req.GroupIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if len(bizIDName) == 0 {
		return &dissolve.ListHostDetails{}, nil
	}

	blackBizIDName, err := l.getBlackBizIDName(kt)
	if err != nil {
		logs.Errorf("get black biz ids failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	hostIDs, hostBizIDMap, err := l.getHostIDByBizCond(kt, originHostIDs, bizIDName, blackBizIDName)
	if err != nil {
		logs.Errorf("get host id by biz cond failed, err: %v, hostIDs: %v, bizIDName: %v, blackBizIDName: %v, rid: %s",
			err, originHostIDs, bizIDName, blackBizIDName, kt.Rid)
		return nil, err
	}

	var firstErr error
	wg := sync.WaitGroup{}
	wg.Add(2)

	// 3.根据过滤出来的hostIDs查询cc中的主机
	var ccHosts []cmdb.Host
	var count int64
	go func() {
		defer func() {
			wg.Done()
		}()

		ccHosts, count, err = l.getHostByIDFromCC(kt, hostIDs, req.Page, source)
		if err != nil {
			logs.Errorf("get host from cc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			firstErr = err
		}
	}()

	// 4.根据条件查询es主机数据
	esHostMap := make(map[string]dissolve.Host)
	go func() {
		defer func() {
			wg.Done()
		}()

		if req.Page.Count {
			return
		}

		cond, err := req.GetESCond(assetIDs, bizIDName, blackBizIDName)
		if err != nil {
			logs.Errorf("get es cond failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			firstErr = err
			return
		}
		res, err := l.findHostFromES(kt, cond, l.getOriginHostIndex(), &core.BasePage{Limit: noLimit},
			source.GetEsHostFields())
		if err != nil {
			logs.Errorf("find host from es failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			firstErr = err
			return
		}

		for _, host := range res.Details {
			esHostMap[host.ServerAssetID] = host
		}
	}()

	wg.Wait()
	if firstErr != nil {
		logs.Errorf("find current host data failed, err: %v, req: %v, rid: %s", firstErr, req, kt.Rid)
		return nil, firstErr
	}

	if req.Page.Count {
		return &dissolve.ListHostDetails{Count: count}, nil
	}

	// 5.从es中填充主机缺少的字段
	data, err := l.fillHostDataByES(kt, ccHosts, esHostMap, hostBizIDMap)
	if err != nil {
		logs.Errorf("fill host data by es failed, err: %v, host: %+v, rid: %s", err, ccHosts, kt.Rid)
		return nil, err
	}

	// 6.填充主机项目类型
	data, err = l.fillProjectName(kt, data)
	if err != nil {
		logs.Errorf("fill project name failed, err: %v, host: %+v, rid: %s", err, data, kt.Rid)
		return nil, err
	}

	return &dissolve.ListHostDetails{Details: data}, nil
}

// fillProjectName 填充主机项目类型
func (l *logics) fillProjectName(kt *kit.Kit, hosts []dissolve.Host) ([]dissolve.Host, error) {
	assetIDs := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		assetIDs[host.ServerAssetID] = struct{}{}
	}
	uniqueAssetIDs := maps.Keys(assetIDs)
	if len(uniqueAssetIDs) == 0 {
		return hosts, nil
	}

	assetIDProjectNameMap := make(map[string]string, len(uniqueAssetIDs))

	for _, batch := range slice.Split(uniqueAssetIDs, int(core.DefaultMaxPageLimit)) {
		opt := &types.ListOption{
			Fields: []string{"asset_id", "project_name"},
			Filter: tools.ExpressionAnd(
				tools.RuleIn("asset_id", batch),
				// 查询当前阶段为未裁撤的主机
				tools.RuleIn("abolish_phase", []enumor.AbolishPhase{enumor.Incomplete, enumor.BsiComplete,
					enumor.Retain}),
			),
			Page: core.NewDefaultBasePage(),
		}
		list, err := l.recycledHost.List(kt, opt)
		if err != nil {
			logs.Errorf("list recycled hosts failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		for _, recycledHost := range list.Details {
			assetID := cvt.PtrToVal(recycledHost.AssetID)
			if projectName := cvt.PtrToVal(recycledHost.ProjectName); projectName != "" {
				assetIDProjectNameMap[assetID] = projectName
			}
		}
	}

	for i := range hosts {
		if projectName, ok := assetIDProjectNameMap[hosts[i].ServerAssetID]; ok {
			hosts[i].ProjectName = projectName
		}
	}

	return hosts, nil
}

// ListResDissolveTable list resource dissolve table
func (l *logics) ListResDissolveTable(kt *kit.Kit, req *dissolve.ResDissolveReq) ([]dissolve.BizDetail, error) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	var firstErr error

	// 1.获取原始业务和主机数据
	var bizMap map[int64]dissolve.BizDetail
	go func() {
		defer func() {
			wg.Done()
		}()

		var err error
		bizMap, err = l.getOriginBizData(kt, req)
		if err != nil {
			logs.Errorf("get origin business data failed, err: %v, cond: %v, rid: %s", err, req, kt.Rid)
			firstErr = err
		}
	}()

	// 2.获取当前主机数据
	var res *dissolve.ListHostDetails
	go func() {
		defer func() {
			wg.Done()
		}()

		cond := &dissolve.HostListReq{ResDissolveReq: *req, Page: &core.BasePage{Limit: noLimit}}
		var err error
		res, err = l.FindCurHost(kt, cond, ReqForGetDissolveTable)
		if err != nil {
			logs.Errorf("find current host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
			firstErr = err
		}
	}()

	wg.Wait()
	if firstErr != nil {
		logs.Errorf("find host data failed, err: %v, req: %v, rid: %s", firstErr, req, kt.Rid)
		return nil, firstErr
	}

	// 3.补充当前的主机相关数据到原始数据中
	bizMap, err := l.fillCurHostData(kt, res, bizMap)
	if err != nil {
		logs.Errorf("fill current host data failed, err: %v, cond: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if len(bizMap) == 0 {
		return make([]dissolve.BizDetail, 0), nil
	}

	// 4.计算总数以及裁撤进度
	result, err := calculateBizData(bizMap)
	if err != nil {
		logs.Errorf("calculate business data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (l *logics) getOriginBizData(kt *kit.Kit, cond *dissolve.ResDissolveReq) (map[int64]dissolve.BizDetail, error) {
	req := &dissolve.HostListReq{ResDissolveReq: *cond, Page: &core.BasePage{Limit: noLimit}}
	res, err := l.FindOriginHost(kt, req, ReqForGetDissolveTable)
	if err != nil {
		logs.Errorf("find origin host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	bizMap := make(map[int64]dissolve.BizDetail, 0)
	for _, host := range res.Details {
		bizDetail, ok := bizMap[host.BizID]
		if !ok {
			bizDetail = dissolve.BizDetail{
				BizID: host.BizID, BizName: host.AppName, ModuleHostCount: make(map[string]int),
			}
		}

		var hostCount int64 = 1
		if bizDetail.Total.Origin.HostCount != nil {
			hostCount += bizDetail.Total.Origin.HostCount.(int64)
		}
		bizDetail.Total.Origin.HostCount = hostCount
		bizDetail.Total.Origin.CpuCount += host.MaxCPUCoreAmount

		bizMap[host.BizID] = bizDetail
	}

	return bizMap, nil
}

func (l *logics) fillCurHostData(kt *kit.Kit, res *dissolve.ListHostDetails, bizMap map[int64]dissolve.BizDetail) (
	map[int64]dissolve.BizDetail, error) {

	if res == nil {
		return nil, fmt.Errorf("res param is nil")
	}

	for _, host := range res.Details {
		bizDetail, ok := bizMap[host.BizID]
		if !ok {
			bizDetail = dissolve.BizDetail{
				BizID: host.BizID, BizName: host.AppName, ModuleHostCount: make(map[string]int),
			}
		}

		var hostCount int64 = 1
		if bizDetail.Total.Current.HostCount != nil {
			hostCount += bizDetail.Total.Current.HostCount.(int64)
		}
		bizDetail.Total.Current.HostCount = hostCount
		bizDetail.Total.Current.CpuCount += host.MaxCPUCoreAmount

		if _, ok = bizDetail.ModuleHostCount[host.ModuleName]; !ok {
			bizDetail.ModuleHostCount[host.ModuleName] = 0
		}
		bizDetail.ModuleHostCount[host.ModuleName]++

		bizMap[host.BizID] = bizDetail
	}

	// 补充当前业务以机房裁撤申领的已交付CPU核心数
	bizIDs := maps.Keys(bizMap)
	bizDeliveredCpuCoreMap, err := l.listBizDeliveredCpuCore(kt, bizIDs)
	if err != nil {
		logs.Errorf("list business delivered cpu core failed, err: %v, bizIDs: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}
	for bizID, bizDetail := range bizMap {
		deliveredCpuCore, ok := bizDeliveredCpuCoreMap[bizID]
		if !ok {
			logs.Errorf("list business delivered cpu core not found, bizID: %d, rid: %s", bizID, kt.Rid)
			return nil, fmt.Errorf("list business delivered cpu core not found, bizID: %d", bizID)
		}
		bizDetail.Total.DeliveredCore = deliveredCpuCore
		bizMap[bizID] = bizDetail
	}

	return bizMap, nil
}

func calculateBizData(bizMap map[int64]dissolve.BizDetail) ([]dissolve.BizDetail, error) {
	result := make([]dissolve.BizDetail, 0)
	var curHostNum, originHostNum, curCpuNum, originCpuNum, deliveredCore int64
	moduleHostCount := make(map[string]int)

	for _, data := range bizMap {
		var curBizHostNum, originBizHostNum int64
		if data.Total.Current.HostCount != nil {
			curBizHostNum = data.Total.Current.HostCount.(int64)
		}
		if data.Total.Origin.HostCount != nil {
			originBizHostNum = data.Total.Origin.HostCount.(int64)
		}

		curHostNum += curBizHostNum
		originHostNum += originBizHostNum
		curCpuNum += data.Total.Current.CpuCount
		originCpuNum += data.Total.Origin.CpuCount
		deliveredCore += data.Total.DeliveredCore

		for module, count := range data.ModuleHostCount {
			if _, ok := moduleHostCount[module]; !ok {
				moduleHostCount[module] = count
				continue
			}

			moduleHostCount[module] += count
		}

		data.Progress = getProgress(originBizHostNum, curBizHostNum)
		result = append(result, data)
	}

	total := dissolve.BizDetail{
		BizID:           "total",
		BizName:         "总数",
		ModuleHostCount: moduleHostCount,
		Total: dissolve.Total{
			Origin:        dissolve.TotalData{HostCount: originHostNum, CpuCount: originCpuNum},
			Current:       dissolve.TotalData{HostCount: curHostNum, CpuCount: curCpuNum},
			DeliveredCore: deliveredCore,
		},
	}
	result = append(result, total)

	val := getProgress(originHostNum, curHostNum)
	progress := dissolve.BizDetail{
		BizID:   "recycle-progress",
		BizName: "裁撤进度",
		Total: dissolve.Total{
			Origin:  dissolve.TotalData{HostCount: val},
			Current: dissolve.TotalData{HostCount: val},
		},
		Progress: val,
	}
	result = append(result, progress)

	return result, nil
}

func getProgress(origin, current int64) string {
	if origin == 0 {
		return ""
	}

	return fmt.Sprintf("%.2f%%", (float64(origin-current))/float64(origin)*100)
}

// ListBizCpuCoreSummary list business cpu core summary.
func (l *logics) ListBizCpuCoreSummary(kt *kit.Kit, bizIDs []int64) (map[int64]dissolve.CpuCoreSummary, error) {
	// 1. 初始化数据，确保每个业务都有值
	bizCpuCoreSummaryMap := make(map[int64]dissolve.CpuCoreSummary, len(bizIDs))
	for _, bizID := range bizIDs {
		bizCpuCoreSummaryMap[bizID] = dissolve.CpuCoreSummary{}
	}

	// 2. 获取业务需要裁撤的总核数
	bizTotalCpuCoreMap, err := l.listBizTotalCpuCore(kt, bizIDs)
	if err != nil {
		logs.Errorf("list biz total cpu core failed, err: %v, bizIDs: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	// 3. 获取业务机房裁撤已交付的总核数
	bizDeliveredCpuCoreMap, err := l.listBizDeliveredCpuCore(kt, bizIDs)
	if err != nil {
		logs.Errorf("list biz delivered cpu core failed, err: %v, bizIDs: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	// 4.获取统计机房裁撤主机的开始时间
	hostApplyTime, err := l.dissolveConfig.GetDissolveHostApplyTime(kt)
	if err != nil {
		logs.Errorf("get host apply time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 5. 组装业务数据
	for _, bizID := range bizIDs {
		totalCpuCore, ok := bizTotalCpuCoreMap[bizID]
		if !ok {
			logs.Errorf("biz total cpu core is invalid, bizID: %d, rid: %s", bizID, kt.Rid)
			return nil, fmt.Errorf("biz total cpu core is invalid, bizID: %d", bizID)
		}

		deliveredCpuCore, ok := bizDeliveredCpuCoreMap[bizID]
		if !ok {
			logs.Errorf("biz delivered cpu core is invalid, bizID: %d, rid: %s", bizID, kt.Rid)
			return nil, fmt.Errorf("biz delivered cpu core is invalid, bizID: %d", bizID)
		}

		bizCpuCoreSummaryMap[bizID] = dissolve.CpuCoreSummary{
			TotalCore:     totalCpuCore,
			DeliveredCore: deliveredCpuCore,
			HostApplyTime: cvt.PtrToVal(hostApplyTime),
		}
	}

	return bizCpuCoreSummaryMap, nil
}

// listBizTotalCore list business dissolve total cpu core.
func (l *logics) listBizTotalCpuCore(kt *kit.Kit, bizIDs []int64) (map[int64]int64, error) {
	bizCpuCoreMap := make(map[int64]int64, len(bizIDs))
	for _, bizID := range bizIDs {
		bizCpuCoreMap[bizID] = 0
	}

	// 1. 获取全部需要裁撤的模块
	moduleNames, err := l.getAllModuleName(kt)
	if err != nil {
		logs.Errorf("list all module name failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(moduleNames) == 0 {
		return bizCpuCoreMap, nil
	}

	// 2. 查询业务需要裁撤的所有主机
	bizIDNameMap, err := l.getBizIDNameByID(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz id name failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	req := &dissolve.HostListReq{
		ResDissolveReq: dissolve.ResDissolveReq{
			BizNames:    maps.Values(bizIDNameMap),
			ModuleNames: moduleNames,
		},
		Page: &core.BasePage{Limit: noLimit},
	}
	hosts, err := l.FindOriginHost(kt, req, ReqForGetDissolveTable)
	if err != nil {
		logs.Errorf("find origin host failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 3. 统计每个业务需要裁撤的cup总核数
	for _, host := range hosts.Details {
		if _, ok := bizCpuCoreMap[host.BizID]; !ok {
			logs.Errorf("host is invalid, host: %+v, rid: %s", host, kt.Rid)
			return nil, fmt.Errorf("host is invalid, host: %+v", host)
		}

		bizCpuCoreMap[host.BizID] += host.MaxCPUCoreAmount
	}

	return bizCpuCoreMap, nil
}

func (l *logics) getAllModuleName(kt *kit.Kit) ([]string, error) {
	moduleNames := make([]string, 0)
	req := &types.ListOption{
		Fields: []string{"name"},
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		list, err := l.recycledModule.List(kt, req)
		if err != nil {
			logs.Errorf("list module failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, module := range list.Details {
			moduleNames = append(moduleNames, cvt.PtrToVal(module.Name))
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return moduleNames, nil
}

// listBizDeliveredCore list business dissolve delivered cpu core.
func (l *logics) listBizDeliveredCpuCore(kt *kit.Kit, bizIDs []int64) (map[int64]int64, error) {
	bizCpuCoreMap := make(map[int64]int64, len(bizIDs))
	for _, bizID := range bizIDs {
		bizCpuCoreMap[bizID] = 0
	}

	// 1. 查询“统计机房裁撤需求类型主机”开始时间
	time, err := l.dissolveConfig.GetDissolveHostApplyTime(kt)
	if err != nil {
		logs.Errorf("get host apply time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 2. 查询从统计时间开始，申请的机房裁撤类型的主机
	bizIDDeviceTypeHostCountMap := make(map[int64]map[string]int64)
	deviceTypeMap := make(map[string]struct{})
	page := metadata.BasePage{Start: 0, Limit: int(core.DefaultMaxPageLimit)}
	filter := map[string]interface{}{
		"bk_biz_id":    map[string]interface{}{condition.BKDBIN: bizIDs},
		"require_type": enumor.RequireTypeDissolve,
		"is_delivered": true,
		"create_at":    map[string]interface{}{condition.BKDBGTE: time},
	}
	for {
		hosts, err := model.Operation().DeviceInfo().FindManyDeviceInfo(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("list device info failed, err: %v, filter: %+v, rid: %s", err, filter, kt.Rid)
			return nil, err
		}
		for _, host := range hosts {
			bizID := int64(host.BkBizId)
			if _, ok := bizIDDeviceTypeHostCountMap[bizID]; !ok {
				bizIDDeviceTypeHostCountMap[bizID] = make(map[string]int64)
			}
			deviceType := host.DeviceType
			if _, ok := bizIDDeviceTypeHostCountMap[bizID][deviceType]; !ok {
				bizIDDeviceTypeHostCountMap[bizID][deviceType] = 0
			}
			bizIDDeviceTypeHostCountMap[bizID][deviceType]++
			deviceTypeMap[deviceType] = struct{}{}
		}
		if len(hosts) < int(core.DefaultMaxPageLimit) {
			break
		}
		page.Start += int(core.DefaultMaxPageLimit)
	}

	// 3. 查询机型对应的核心数
	deviceTypes := maps.Keys(deviceTypeMap)
	deviceTypeInfos, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("list device type info failed, err: %v, deviceTypes: %v, rid: %s", err, deviceTypes, kt.Rid)
		return nil, err
	}
	deviceTypeCpuCoreMap := make(map[string]int64, len(deviceTypeInfos))
	for _, info := range deviceTypeInfos {
		deviceTypeCpuCoreMap[info.DeviceType] = info.CPUAmount
	}

	// 4. 统计已交付给业务的核心数
	for bizID := range bizCpuCoreMap {
		deviceTypeHostCount, ok := bizIDDeviceTypeHostCountMap[bizID]
		if !ok {
			continue
		}
		for deviceType, hostCount := range deviceTypeHostCount {
			if _, ok = deviceTypeCpuCoreMap[deviceType]; !ok {
				logs.Errorf("can not find device type info, type: %s, bizID: %d", deviceType, bizID)
				return nil, fmt.Errorf("can not find device type info, type: %s, bizID: %d", deviceType, bizID)
			}
			bizCpuCoreMap[bizID] += deviceTypeCpuCoreMap[deviceType] * hostCount
		}
	}

	return bizCpuCoreMap, nil
}
