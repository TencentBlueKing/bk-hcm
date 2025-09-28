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

package bkcc

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/hooks"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var allFields = make([]string, 0)

// watchCCEvent 监听cmdb事件
func (w *Watcher) watchCCEvent(sd serviced.ServiceDiscover, resType cmdb.CursorType, eventTypes []cmdb.EventType,
	fields []string, consumeFunc func(kt *kit.Kit, events []cmdb.WatchEventDetail) error) {

	param := &cmdb.WatchEventParams{
		EventTypes: eventTypes,
		Resource:   resType,
	}
	if len(fields) != 0 {
		param.Fields = fields
	}

	for {
		if !sd.IsMaster() {
			time.Sleep(10 * time.Second)
			continue
		}

		kt := core.NewBackendKit()
		cursor, err := w.getEventCursor(kt, resType)
		if err != nil {
			logs.Errorf("get event cursor failed, err: %v, type: %s, rid: %s", err, resType, kt.Rid)
			continue
		}
		param.Cursor = cursor

		result, err := cmdb.CmdbClient().ResourceWatch(kt, param)
		if err != nil {
			logs.Errorf("watch cmdb host resource failed, err: %v, req: %+v, rid: %s", err, param, kt.Rid)
			// 如果事件节点不存在，cc会返回该错误码，此时需要将cursor设置为""，从当前时间开始监听事件
			if strings.Contains(err.Error(), cmdb.CCErrEventChainNodeNotExist) {
				if err = w.setEventCursor(kt, resType, ""); err != nil {
					logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
						"", kt.Rid)
				}
			}
			continue
		}

		if !result.Watched {
			if len(result.Events) != 0 {
				newCursor := result.Events[0].Cursor
				if err = w.setEventCursor(kt, resType, newCursor); err != nil {
					logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
						newCursor, kt.Rid)
				}
			}
			continue
		}

		if err = consumeFunc(kt, result.Events); err != nil {
			logs.Errorf("consume event failed, err: %+v, type: %s, res: %+v, rid: %s", err, resType, result, kt.Rid)
		}

		if len(result.Events) != 0 {
			newCursor := result.Events[len(result.Events)-1].Cursor
			if err = w.setEventCursor(kt, resType, newCursor); err != nil {
				logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
					newCursor, kt.Rid)
			}
		}
	}
}

// WatchHostEvent 监听主机事件，增量同步主机
func (w *Watcher) WatchHostEvent(sd serviced.ServiceDiscover) {
	w.watchCCEvent(sd, cmdb.HostType, []cmdb.EventType{cmdb.Create, cmdb.Update, cmdb.Delete}, cmdb.HostFields,
		w.consumeHostEvent)
}

func (w *Watcher) consumeHostEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
	if len(events) == 0 {
		return nil
	}

	idHostMap := make(map[int64]cmdb.Host)
	deleteHosts := make([]cmdb.Host, 0)

	// 1. 获取需要创建、更新、删除的主机
	for _, event := range events {
		host, err := convertHost(kt, event.Detail)
		if err != nil {
			logs.Errorf("convert host failed, err: %v, event: %+v, rid: %s", err, event, kt.Rid)
			continue
		}

		if event.EventType == cmdb.Delete {
			deleteHosts = append(deleteHosts, converter.PtrToVal(host))
			delete(idHostMap, host.BkHostID)
			continue
		}

		idHostMap[host.BkHostID] = converter.PtrToVal(host)
	}

	// 2. 创建或更新主机
	upsertHosts := make([]cmdb.Host, 0)
	for _, host := range idHostMap {
		upsertHosts = append(upsertHosts, host)
	}
	if len(upsertHosts) != 0 {
		if err := w.upsertHost(kt, upsertHosts); err != nil {
			logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, upsertHosts, kt.Rid)
		}
	}

	// 3. 删除需要删除的主机
	if len(deleteHosts) != 0 {
		if err := w.deleteHost(kt, deleteHosts); err != nil {
			logs.Errorf("delete host failed, err: %v, ids: %+v, rid: %s", err, deleteHosts, kt.Rid)
		}
	}

	return nil
}

func convertHost(kt *kit.Kit, data json.RawMessage) (*cmdb.Host, error) {
	host := &cmdb.Host{}
	if err := json.Unmarshal(data, host); err != nil {
		logs.Errorf("unmarshal host failed, err: %v, data: %v, rid: %s", err, data, kt.Rid)
		return nil, err
	}

	return host, nil
}

func (w *Watcher) upsertHost(kt *kit.Kit, upsertHosts []cmdb.Host) error {
	if len(upsertHosts) == 0 {
		return nil
	}

	vendors := []enumor.Vendor{enumor.Other}
	vendors = hooks.AdjustWatcherVendor(kt, vendors)
	vendorAccountIDMap, err := w.getVendorAccountID(kt, vendors)
	if err != nil {
		logs.Errorf("get vendor account id failed, err: %v, vendors: %v, rid: %s", err, vendors, kt.Rid)
		return err
	}

	bizIDVendorHostIDsMap, err := w.classifyHost(kt, upsertHosts, false)
	if err != nil {
		logs.Errorf("classify host failed, err: %v, hosts: %v, rid: %s", err, upsertHosts, kt.Rid)
		return err
	}

	for bizID, vendorHostIDsMap := range bizIDVendorHostIDsMap {
		for vendor, hostIDs := range vendorHostIDsMap {
			accountID, ok := vendorAccountIDMap[vendor]
			if !ok {
				logs.Errorf("get vendor account id failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
				continue
			}
			switch vendor {
			case enumor.Other:
				for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
					req := &sync.OtherSyncHostByCondReq{BizID: bizID, HostIDs: batch, AccountID: accountID}
					err = w.CliSet.HCService().Other.Host.SyncHostWithRelResByCond(kt.Ctx, kt.Header(), req)
					if err != nil {
						logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, batch, kt.Rid)
						continue
					}
				}
			// todo add other case
			default:
				logs.Errorf("not support vendor: %s, hostIDs: %v, rid: %s", vendor, hostIDs, kt.Rid)
			}
		}
	}

	return nil
}

const ignoreBizID int64 = 0

func (w *Watcher) classifyHost(kt *kit.Kit, hosts []cmdb.Host, isIgnoreBizID bool) (
	map[int64]map[enumor.Vendor][]int64, error) {

	hostIDs := make([]int64, 0, len(hosts))
	for _, host := range hosts {
		hostIDs = append(hostIDs, host.BkHostID)
	}
	hostBizIDMap := make(map[int64]int64)
	if !isIgnoreBizID {
		var err error
		hostBizIDMap, err = w.getHostBizID(kt, hostIDs)
		if err != nil {
			logs.Errorf("get host bizID map failed, err: %v, ids: %v, rid: %s", err, hostIDs, kt.Rid)
			return nil, err
		}
	}

	bizIDVendorHostIDsMap := make(map[int64]map[enumor.Vendor][]int64)
	for _, host := range hosts {
		bizID := ignoreBizID
		if !isIgnoreBizID {
			var ok bool
			bizID, ok = hostBizIDMap[host.BkHostID]
			if !ok {
				logs.Errorf("get host bizID failed, hostID: %v, rid: %s", host.BkHostID, kt.Rid)
				continue
			}
		}

		if _, ok := bizIDVendorHostIDsMap[bizID]; !ok {
			bizIDVendorHostIDsMap[bizID] = make(map[enumor.Vendor][]int64)
		}

		match, vendor, err := hooks.MatchWatcherUpsertHost(kt, host)
		if err != nil {
			logs.Errorf("match watcher upsert host failed, err: %v, host: %+v, rid: %s", err, host, kt.Rid)
			continue
		}
		if match {
			if _, ok := bizIDVendorHostIDsMap[bizID][vendor]; !ok {
				bizIDVendorHostIDsMap[bizID][vendor] = make([]int64, 0)
			}
			bizIDVendorHostIDsMap[bizID][vendor] = append(bizIDVendorHostIDsMap[bizID][vendor], host.BkHostID)
			continue
		}

		if _, ok := bizIDVendorHostIDsMap[bizID][enumor.Other]; !ok {
			bizIDVendorHostIDsMap[bizID][enumor.Other] = make([]int64, 0)
		}
		bizIDVendorHostIDsMap[bizID][enumor.Other] = append(bizIDVendorHostIDsMap[bizID][enumor.Other], host.BkHostID)
	}

	return bizIDVendorHostIDsMap, nil
}

func (w *Watcher) deleteHost(kt *kit.Kit, deleteHosts []cmdb.Host) error {
	if len(deleteHosts) == 0 {
		return nil
	}

	vendors := []enumor.Vendor{enumor.Other}
	vendors = hooks.AdjustWatcherVendor(kt, vendors)
	vendorAccountIDMap, err := w.getVendorAccountID(kt, vendors)
	if err != nil {
		logs.Errorf("get vendor account id failed, err: %v, vendors: %v, rid: %s", err, vendors, kt.Rid)
		return err
	}

	bizIDVendorHostIDsMap, err := w.classifyHost(kt, deleteHosts, true)
	if err != nil {
		logs.Errorf("classify host failed, err: %v, hosts: %v, rid: %s", err, deleteHosts, kt.Rid)
		return err
	}
	vendorHostIDsMap, ok := bizIDVendorHostIDsMap[ignoreBizID]
	if !ok {
		logs.Errorf("can not get vendor host ids map, map: %v, rid: %s", bizIDVendorHostIDsMap, kt.Rid)
		return errors.New("can not get vendor host ids map")
	}

	for vendor, hostIDs := range vendorHostIDsMap {
		accountID, ok := vendorAccountIDMap[vendor]
		if !ok {
			logs.Errorf("get vendor account id failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
			continue
		}
		switch vendor {
		case enumor.Other:
			for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
				req := &sync.OtherDelHostByCondReq{HostIDs: batch, AccountID: accountID}
				if err = w.CliSet.HCService().Other.Host.DeleteHostByCond(kt.Ctx, kt.Header(), req); err != nil {
					logs.Errorf("delete host failed, err: %v, account id: %s, ids: %+v, rid: %s", err, accountID, batch,
						kt.Rid)
					continue
				}
			}
		// todo add other case
		default:
			logs.Errorf("not support vendor: %s, hostIDs: %v, rid: %s", vendor, hostIDs, kt.Rid)
		}
	}

	return nil
}

func (w *Watcher) getVendorAccountID(kt *kit.Kit, vendors []enumor.Vendor) (map[enumor.Vendor]string, error) {
	req := &cloud.AccountListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("vendor", vendors)),
		Page:   &core.BasePage{Start: 0, Limit: constant.BatchOperationMaxLimit},
	}

	accounts, err := w.CliSet.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("get account failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(accounts.Details) == 0 {
		logs.Errorf("can not get account, req: %+v, rid: %s", req, kt.Rid)
		return nil, errors.New("can not get account")
	}

	vendorAccountIDMap := make(map[enumor.Vendor]string)
	for _, account := range accounts.Details {
		vendorAccountIDMap[account.Vendor] = account.ID
	}

	return vendorAccountIDMap, nil
}

func (w *Watcher) getHostBizID(kt *kit.Kit, hostIDs []int64) (map[int64]int64, error) {
	if len(hostIDs) == 0 {
		return make(map[int64]int64), nil
	}

	hostBizIDMap := make(map[int64]int64)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		req := &cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := cmdb.CmdbClient().FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("fail to find cmdb topo relation, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, relation := range converter.PtrToVal(relationRes) {
			bizID := relation.BizID
			if bizID == w.ccHostPoolBiz {
				bizID = constant.HostPoolBiz
			}

			hostBizIDMap[relation.HostID] = bizID
		}
	}

	return hostBizIDMap, nil
}

// WatchHostRelationEvent 监听主机关系事件，增量修改主机关系
func (w *Watcher) WatchHostRelationEvent(sd serviced.ServiceDiscover) {
	w.watchCCEvent(sd, cmdb.HostRelation, []cmdb.EventType{cmdb.Create}, allFields, w.consumeHostRelationEvent)
}

func (w *Watcher) consumeHostRelationEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
	if len(events) == 0 {
		return nil
	}

	hostBizIDMap := make(map[int64]int64)
	hostIDs := make([]int64, 0)
	for _, event := range events {
		relation, err := convertHostRelation(kt, event.Detail)
		if err != nil {
			logs.Errorf("convert host relation failed, err: %v, event: %+v, rid: %s", err, event, kt.Rid)
			continue
		}

		if _, ok := hostBizIDMap[relation.HostID]; !ok {
			hostIDs = append(hostIDs, relation.HostID)
		}

		hostBizIDMap[relation.HostID] = relation.BizID
	}

	dbHosts, err := w.listHostFromDB(kt, hostIDs)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return err
	}

	updateHostIDs := make([]int64, 0)
	for _, host := range dbHosts {
		if hostBizIDMap[host.BkHostID] == host.BkBizID {
			continue
		}

		updateHostIDs = append(updateHostIDs, host.BkHostID)
	}

	if len(updateHostIDs) == 0 {
		return nil
	}

	updateHosts, err := w.listHostFromCC(kt, updateHostIDs)
	if err != nil {
		logs.Errorf("list host from cc failed, err: %v, hostIDs: %v, rid: %s", err, updateHostIDs, kt.Rid)
		return err
	}

	if err = w.upsertHost(kt, updateHosts); err != nil {
		logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, updateHostIDs, kt.Rid)
	}

	return nil
}

func convertHostRelation(kt *kit.Kit, data json.RawMessage) (*cmdb.HostTopoRelation, error) {
	relation := &cmdb.HostTopoRelation{}
	if err := json.Unmarshal(data, relation); err != nil {
		logs.Errorf("unmarshal host relation failed, err: %v, data: %v, rid: %s", err, data, kt.Rid)
		return nil, err
	}

	return relation, nil
}

func (w *Watcher) listHostFromDB(kt *kit.Kit, hostIDs []int64) ([]cvm.BaseCvm, error) {
	hosts := make([]cvm.BaseCvm, 0)
	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("bk_host_id", batch)),
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
				Sort:  "bk_host_id",
			},
		}
		result, err := w.CliSet.DataService().Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.ErrorJson("request dataservice to list cvm failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Details...)
	}

	return hosts, nil
}

func (w *Watcher) listHostFromCC(kt *kit.Kit, hostIDs []int64) ([]cmdb.Host, error) {
	hostBizID, err := w.getHostBizID(kt, hostIDs)
	if err != nil {
		logs.Errorf("get host biz id failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, err
	}
	bizHostIDs := make(map[int64][]int64)
	for hostID, bizID := range hostBizID {
		if _, ok := bizHostIDs[bizID]; !ok {
			bizHostIDs[bizID] = make([]int64, 0)
		}
		bizHostIDs[bizID] = append(bizHostIDs[bizID], hostID)
	}

	hosts := make([]cmdb.Host, 0)
	for bizID, ids := range bizHostIDs {
		for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
			filter := &cmdb.QueryFilter{
				Rule: &cmdb.CombinedRule{
					Condition: "AND",
					Rules: []cmdb.Rule{
						&cmdb.AtomRule{Field: "bk_host_id", Operator: cmdb.OperatorIn, Value: batch},
					},
				},
			}
			page := &cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit), Sort: "bk_host_id"}
			if bizID == constant.HostPoolBiz {
				params := &cmdb.ListResourcePoolHostsParams{
					Fields:             cmdb.HostFields,
					HostPropertyFilter: filter,
					Page:               page,
				}
				result, err := cmdb.CmdbClient().ListResourcePoolHosts(kt, params)
				if err != nil {
					logs.Errorf("failed to list resource pool host, err: %v, req: %+v, rid: %s", err, params, kt.Rid)
					return nil, err
				}
				hosts = append(hosts, result.Info...)
				continue
			}

			params := &cmdb.ListBizHostParams{
				BizID:              bizID,
				Fields:             cmdb.HostFields,
				HostPropertyFilter: filter,
				Page:               page,
			}
			result, err := cmdb.CmdbClient().ListBizHost(kt, params)
			if err != nil {
				logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, params, kt.Rid)
				return nil, err
			}
			hosts = append(hosts, result.Info...)
		}
	}

	return hosts, nil
}
