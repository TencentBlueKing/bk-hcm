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
	"sync"

	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/es"

	"golang.org/x/sync/errgroup"
)

func (l *logics) findHostFromES(kt *kit.Kit, cond map[string][]interface{}, index string, page *core.BasePage,
	fields []string) (*dissolve.ListHostDetails, error) {
	if page == nil {
		return nil, errors.New("page is nil")
	}
	if page.Count {
		count, err := l.esCli.CountWithCond(kt.Ctx, cond, index)
		if err != nil {
			logs.Errorf("get host count by condition failed, err: %v, index: %s, cond: %v, rid: %s", err, index, cond,
				kt.Rid)
			return nil, err
		}
		return &dissolve.ListHostDetails{Count: count}, nil
	}
	if page.Sort == "" {
		page.Sort = "_id"
	}
	var lock sync.Mutex
	var err error
	pipeline := make(chan struct{}, 10)
	res := make([]es.Host, 0)
	requestEnd := page.Start + uint32(page.Limit)
	doFunc := func(start, limit int) error {
		defer func() {
			<-pipeline
		}()
		if start >= int(requestEnd) {
			return nil
		}

		var hosts []es.Host
		hosts, err = l.esCli.SearchWithCond(kt.Ctx, cond, index, start, limit, page.Sort, fields)
		if err != nil {
			logs.Errorf("get host by condition failed, err: %v, index: %s, cond: %v, rid: %s", err, index, cond,
				kt.Rid)
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		res = append(res, hosts...)
		// start+len(hosts) < int(requestEnd) 的判断逻辑是为了防止多个协程查不到数据后，不断累加requestEnd，所以应该取它们的最小值
		if len(hosts) < limit && start+len(hosts) < int(requestEnd) {
			requestEnd = uint32(start + len(hosts))
		}
		return nil
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	start := page.Start
	var limit uint32 = 3000
	for start < requestEnd {
		if start+limit > requestEnd {
			limit = requestEnd - start
		}
		pipeline <- struct{}{}
		curStart := start
		curLimit := limit
		eg.Go(func() error { return doFunc(int(curStart), int(curLimit)) })
		// 这里错误只跳出当前循环，由下面eg.Wait()的时候统一处理异常
		if err != nil {
			break
		}
		start += limit
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	details := make([]dissolve.Host, len(res))
	for i, host := range res {
		detail, err := dissolve.ConvertHost(&host)
		if err != nil {
			logs.Errorf("convert host failed, err: %v, host: %v, rid: %s", err, host, kt.Rid)
			return nil, err
		}
		details[i] = *detail
	}

	return &dissolve.ListHostDetails{Details: details}, nil
}

func (l *logics) fillHostDataByES(kt *kit.Kit, ccHosts []cmdb.Host, esHostMap map[string]dissolve.Host,
	hostBizIDMap map[int64]int64) ([]dissolve.Host, error) {

	bizIDs := make([]int64, 0)
	for _, bizID := range hostBizIDMap {
		bizIDs = append(bizIDs, bizID)
	}
	bizIDName, err := l.getBizIDNameByID(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, ids: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	result := make([]dissolve.Host, len(ccHosts))
	for idx, ccHost := range ccHosts {
		bizID, ok := hostBizIDMap[ccHost.BkHostID]
		if !ok {
			logs.Errorf("can not find biz id, host id: %d, map: %+v,rid: %s", ccHost.BkHostID, hostBizIDMap, kt.Rid)
			return nil, errors.New("host is invalid")
		}

		bizName, ok := bizIDName[bizID]
		if !ok {
			logs.Errorf("can not find biz name, biz id: %d, map: %+v,rid: %s", bizID, bizIDName, kt.Rid)
			return nil, errors.New("biz is invalid")
		}

		esHost, ok := esHostMap[ccHost.BkAssetID]
		if ok {
			esHost.AppName = bizName
			esHost.ModuleName = ccHost.ModuleName
			esHost.BizID = bizID
			esHost.ServerOperator = ccHost.Operator
			esHost.ServerBakOperator = ccHost.BkBakOperator
			esHost.SvrTypeName = ccHost.SvrTypeName
			esHost.InnerIP = ccHost.BkHostInnerIP
			result[idx] = esHost
			continue
		}

		result[idx] = dissolve.Host{
			ServerAssetID:     ccHost.BkAssetID,
			InnerIP:           ccHost.BkHostInnerIP,
			OuterIP:           ccHost.BkHostOuterIP,
			AppName:           bizName,
			BizID:             bizID,
			DeviceType:        ccHost.SvrDeviceClass,
			SvrTypeName:       ccHost.SvrTypeName,
			ModuleName:        ccHost.ModuleName,
			IdcUnitName:       ccHost.IdcUnitName,
			SfwNameVersion:    ccHost.BkOsVersion,
			GoUpDate:          ccHost.SvrInputTime,
			RaidName:          ccHost.RaidName,
			LogicArea:         ccHost.LogicDomain,
			ServerBakOperator: ccHost.BkBakOperator,
			ServerOperator:    ccHost.Operator,
			DiskTotal:         ccHost.BkDisk,
			MaxCPUCoreAmount:  ccHost.BkCpu,
		}
	}

	return result, nil
}
