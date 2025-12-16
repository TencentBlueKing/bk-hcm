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
	"fmt"
	"slices"
	"strings"
	"sync"

	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"

	"golang.org/x/sync/errgroup"
)

// getBizIDNameByName 此方法如果没有传name,那么会查询所有的业务
func (l *logics) getBizIDNameByName(kt *kit.Kit, names []string, groupIDs []int64) (map[int64]string, error) {
	groupIDMap := make(map[int64]struct{})
	for _, id := range groupIDs {
		groupIDMap[id] = struct{}{}
	}

	if len(names) == 0 {
		return l.getAllBizIDName(kt, groupIDMap)
	}

	names = util.StrArrayUnique(names)
	bizIDName := make(map[int64]string, 0)

	req := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_oper_grp_name_id"},
		Page:   cmdb.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit},
	}
	for _, batch := range slice.Split(names, pkg.BKMaxInstanceLimit) {
		req.BizPropertyFilter = &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    "bk_biz_name",
						Operator: cmdb.OperatorIn,
						Value:    batch,
					},
				},
			},
		}

		resp, err := l.cmdbCli.SearchBusiness(kt, req)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, info := range resp.Info {
			if _, ok := groupIDMap[info.BkOperGrpNameID]; !ok && len(groupIDMap) != 0 {
				continue
			}

			bizIDName[info.BizID] = info.BizName
		}
	}

	return bizIDName, nil
}

// getBizIDNameByID 此方法如果没有传id,那么会查询所有的业务
func (l *logics) getBizIDNameByID(kt *kit.Kit, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return l.getAllBizIDName(kt, make(map[int64]struct{}))
	}

	ids = util.IntArrayUnique(ids)
	bizIDName := make(map[int64]string, 0)

	req := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page:   cmdb.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit},
	}
	start := 0
	end := len(ids)
	if len(ids) > pkg.BKMaxInstanceLimit {
		end = pkg.BKMaxInstanceLimit
	}

	for {
		req.BizPropertyFilter = &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    "bk_biz_id",
						Operator: cmdb.OperatorIn,
						Value:    ids[start:end],
					},
				},
			},
		}

		resp, err := l.cmdbCli.SearchBusiness(kt, req)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, info := range resp.Info {
			bizIDName[info.BizID] = info.BizName
		}

		if len(resp.Info) < pkg.BKMaxInstanceLimit {
			break
		}

		start = end
		if end+pkg.BKMaxInstanceLimit > len(ids) {
			end = len(ids)
			continue
		}

		end += pkg.BKMaxInstanceLimit
	}

	return bizIDName, nil
}

func (l *logics) getAllBizIDName(kt *kit.Kit, groupIDMap map[int64]struct{}) (map[int64]string, error) {
	req := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_oper_grp_name_id"},
		Page:   cmdb.BasePage{Start: 0, Limit: noLimit},
	}

	resp, err := l.cmdbCli.SearchBusiness(kt, req)
	if err != nil {
		logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	bizIDName := make(map[int64]string)
	for _, info := range resp.Info {
		if _, ok := groupIDMap[info.BkOperGrpNameID]; !ok && len(groupIDMap) != 0 {
			continue
		}
		bizIDName[info.BizID] = info.BizName
	}

	return bizIDName, nil
}

func (l *logics) getBlackBizIDName(kt *kit.Kit) (map[int64]string, error) {
	if len(l.blacklist) == 0 {
		return make(map[int64]string), nil
	}

	bizNames := make([]string, 0)
	for _, v := range strings.Split(l.blacklist, ",") {
		bizNames = append(bizNames, v)
	}

	return l.getBizIDNameByName(kt, bizNames, make([]int64, 0))
}

func (l *logics) getHostByIDFromCC(kt *kit.Kit, hostIDs []int64, page *core.BasePage, source ReqSourceI) ([]cmdb.Host,
	int64, error) {
	if page.Count {
		req := &cmdb.ListHostReq{
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: querybuilder.CombinedRule{
					Condition: querybuilder.ConditionAnd,
					Rules: []querybuilder.Rule{
						querybuilder.AtomRule{
							Field: pkg.BKHostIDField, Operator: querybuilder.OperatorIn, Value: hostIDs,
						},
					},
				},
			},
		}
		count, err := l.getHostCountFromCC(kt, req)
		if err != nil {
			logs.Errorf("get host count from cc failed, err: %v, cond: %+v, rid: %s", err, req, kt.Rid)
			return nil, 0, err
		}

		return nil, count, nil
	}

	if page.Start > uint32(len(hostIDs)) {
		return make([]cmdb.Host, 0), 0, nil
	}

	var lock sync.Mutex
	hosts := make([]cmdb.Host, 0)
	doFunc := func(ids []int64) error {
		req := &cmdb.ListHostReq{
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: querybuilder.CombinedRule{
					Condition: querybuilder.ConditionAnd,
					Rules: []querybuilder.Rule{
						querybuilder.AtomRule{
							Field: pkg.BKHostIDField, Operator: querybuilder.OperatorIn, Value: ids,
						},
					},
				},
			},
			Fields: source.GetCCHostFields(),
			Page:   cmdb.BasePage{Start: 0, Limit: int64(len(ids)), Sort: pkg.BKHostIDField},
		}
		subHosts, err := l.getHostFromCC(kt, req)
		if err != nil {
			logs.Errorf("get host failed, err: %v, host ids: %v, rid: %s", err, ids, kt.Rid)
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		hosts = append(hosts, subHosts...)
		return nil
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	eg.SetLimit(20)
	slices.Sort(hostIDs)
	end := min(len(hostIDs), int(page.Start+uint32(page.Limit)))
	for _, batch := range slice.Split(hostIDs[page.Start:end], pkg.BKMaxPageSize) {
		curBatch := batch
		eg.Go(func() error { return doFunc(curBatch) })
	}

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}

	return hosts, 0, nil
}

const noLimit = 999999999

func (l *logics) getAllHostIDFromCC(kt *kit.Kit, conds []*cmdb.QueryFilter) ([]int64, error) {
	if len(conds) == 0 {
		logs.Errorf("get all host id from cc failed, cond is nil, rid: %s", kt.Rid)
		return nil, fmt.Errorf("conds is nil")
	}

	hostIDs := make([]int64, 0)
	var lock sync.Mutex
	pipeline := make(chan struct{}, 20)
	doFunc := func(cond cmdb.QueryFilter) error {
		defer func() {
			<-pipeline
		}()

		req := &cmdb.ListHostReq{
			Fields:             []string{pkg.BKHostIDField},
			HostPropertyFilter: &cond,
			Page:               cmdb.BasePage{Start: 0, Limit: noLimit, Sort: pkg.BKHostIDField},
		}
		hosts, err := l.getHostFromCC(kt, req)
		if err != nil {
			logs.Errorf("get host id from cc failed, err: %v, cond: %+v, rid: %s", err, req, kt.Rid)
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		for _, host := range hosts {
			hostIDs = append(hostIDs, host.BkHostID)
		}
		return nil
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	for _, cond := range conds {
		pipeline <- struct{}{}
		curCond := cond
		eg.Go(func() error { return doFunc(converter.PtrToVal(curCond)) })
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return hostIDs, nil
}

func (l *logics) getHostFromCC(kt *kit.Kit, req *cmdb.ListHostReq) ([]cmdb.Host, error) {
	if req == nil {
		return nil, fmt.Errorf("call cc req is nil")
	}

	limit := req.Page.Limit
	if req.Page.Limit > pkg.BKMaxPageSize {
		req.Page.Limit = pkg.BKMaxPageSize
	}

	result := make([]cmdb.Host, 0)
	for {
		resp, err := l.cmdbCli.ListHost(kt, req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, host := range resp.Info {
			result = append(result, host)
		}

		if int64(len(resp.Info)) < req.Page.Limit || int64(len(result)) == limit {
			break
		}

		req.Page.Start += req.Page.Limit

		if int64(len(result))+req.Page.Limit > limit {
			req.Page.Limit = limit - int64(len(result))
		}
	}

	return result, nil
}

func (l *logics) getHostCountFromCC(kt *kit.Kit, req *cmdb.ListHostReq) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("call cc req is nil")
	}

	req.Page = cmdb.BasePage{Start: 0, Limit: 1, Sort: pkg.BKHostIDField}
	hosts, err := l.cmdbCli.ListHost(kt, req)
	if err != nil {
		logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return 0, err
	}

	return int64(hosts.Count), nil
}

// getHostIDByBizCond 筛选hostIDs, 获取在bizIDName对应的业务中，但是不在blackBizIDName业务的host id结果
func (l *logics) getHostIDByBizCond(kt *kit.Kit, originHostIDs []int64, bizIDName, blackBizIDName map[int64]string) (
	[]int64, map[int64]int64, error) {

	var lock sync.Mutex
	pipeline := make(chan struct{}, 30)
	originHostBizIDMap := make(map[int64]int64)
	doFunc := func(hostIDs []int64) error {
		defer func() {
			<-pipeline
		}()

		subHostBizIDMap, err := l.cmdbCli.GetHostBizIds(kt, hostIDs)
		if err != nil {
			logs.Errorf("get host biz id failed, err: %v, originHostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		for key, val := range subHostBizIDMap {
			originHostBizIDMap[key] = val
		}

		return nil
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	pageSize := 1000
	for _, hostIDs := range slice.Split(originHostIDs, pageSize) {
		pipeline <- struct{}{}
		curHostIDs := hostIDs
		eg.Go(func() error { return doFunc(curHostIDs) })
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	bizMap := make(map[int64]struct{})
	for bizID := range bizIDName {
		bizMap[bizID] = struct{}{}
	}

	blackBizIDMap := make(map[int64]struct{})
	for bizID := range blackBizIDName {
		blackBizIDMap[bizID] = struct{}{}
	}

	hostIDs := make([]int64, 0)
	hostBizIDMap := make(map[int64]int64)

	for hostID, bizID := range originHostBizIDMap {
		if _, ok := bizMap[bizID]; !ok {
			continue
		}

		if _, ok := blackBizIDMap[bizID]; ok {
			continue
		}

		hostIDs = append(hostIDs, hostID)
		hostBizIDMap[hostID] = bizID
	}

	return hostIDs, hostBizIDMap, nil
}
