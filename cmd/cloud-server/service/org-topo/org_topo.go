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

// Package orgtopo ...
package orgtopo

import (
	"errors"
	"strconv"
	"time"

	"hcm/pkg/api/core"
	dataorgtopo "hcm/pkg/api/data-service/org_topo"
	table "hcm/pkg/dal/table/org-topo"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/usermgr"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncOrgTopo sync org topo
func (ots *orgTopoSvc) SyncOrgTopo(kt *kit.Kit) error {
	startTime := time.Now()
	logs.Infof("ready to sync usermgr org topo data, startTime: %v, rid: %s", startTime, kt.Rid)

	deptMap, err := ots.userMgrCli.ListAllDepartment(kt)
	if err != nil {
		logs.Errorf("call list all department api failed, rid: %s", kt.Rid)
		return err
	}

	logs.Infof("call list all department api success, deptNum: %d, rid: %s", len(deptMap), kt.Rid)

	cloudOrgTopos, err := ots.convertDeptMapToDB(kt, deptMap)
	if err != nil {
		return err
	}

	dbOrgTopos := make([]table.OrgTopo, 0)
	split := slice.Split(cloudOrgTopos, int(filter.DefaultMaxInLimit))
	for _, parts := range split {
		// list all existed items
		partDeptIDs := make([]string, 0)
		for _, partItem := range parts {
			partDeptIDs = append(partDeptIDs, partItem.DeptID)
		}
		partDeptIDsReq := &dataorgtopo.ListByDeptIDsReq{
			DeptIDs: partDeptIDs,
		}
		exists, err := ots.client.DataService().Global.OrgTopo.ListByDeptIDs(kt, partDeptIDsReq)
		if err != nil {
			logs.Errorf("failed to list among org topo, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		dbOrgTopos = append(dbOrgTopos, exists.Details...)
	}

	adds, updates, deletes, err := ots.diffOrgTopo(kt, cloudOrgTopos, dbOrgTopos)
	if err != nil {
		logs.Errorf("failed to diff org topo, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(deletes) > 0 {
		if err = ots.batchDeleteOrgTopo(kt, deletes); err != nil {
			logs.Errorf("batch delete org topo failed, err: %+v, deletes: %v, rid: %s", err, deletes, kt.Rid)
			return err
		}
	}

	upsertReq := &dataorgtopo.BatchUpsertOrgTopoReq{
		AddOrgTopos:    adds,
		UpdateOrgTopos: updates,
	}
	_, err = ots.client.DataService().Global.OrgTopo.BatchUpsert(kt, upsertReq)
	endTime := time.Now()
	if err != nil {
		logs.Errorf("batch upsert org topo failed, err: %+v, endTime: %v, cost: %fs, rid: %s", err, endTime,
			endTime.Sub(startTime).Seconds(), kt.Rid)
		return err
	}

	logs.Infof("sync usermgr org topo success, total: %d, addNum: %d, updateNum: %d, delNum: %d, endTime: %v, "+
		"cost: %fs, rid: %s", len(cloudOrgTopos), len(adds), len(updates), len(deletes), endTime,
		endTime.Sub(startTime).Seconds(), kt.Rid)

	return nil
}

func (ots *orgTopoSvc) convertDeptMapToDB(kt *kit.Kit, deptMap map[string]*usermgr.DeptInfo) ([]table.OrgTopo, error) {
	if len(deptMap) == 0 {
		return nil, errors.New("dept map is empty")
	}

	orgTopos := make([]table.OrgTopo, 0, len(deptMap))
	for deptID, deptInfo := range deptMap {
		if deptInfo == nil {
			logs.Warnf("dept info is empty, deptID: %+v, rid: %s", deptID, kt.Rid)
			continue
		}

		hasChildren := int64(0)
		if deptInfo.HasChildren {
			hasChildren = 1
		}

		orgTopos = append(orgTopos, table.OrgTopo{
			DeptID:      strconv.FormatInt(deptInfo.ID, 10),
			DeptName:    deptInfo.Name,
			FullName:    deptInfo.FullName,
			Level:       deptInfo.Level,
			Parent:      strconv.FormatInt(deptInfo.Parent, 10),
			HasChildren: cvt.ValToPtr(hasChildren),
			Memo:        nil,
			Creator:     kt.User,
			Reviser:     kt.User,
		})
	}

	return orgTopos, nil
}

func (ots *orgTopoSvc) diffOrgTopo(kt *kit.Kit, topos []table.OrgTopo, dbOrgTopos []table.OrgTopo) (
	[]table.OrgTopo, []table.OrgTopo, []string, error) {

	add := make([]table.OrgTopo, 0)
	update := make([]table.OrgTopo, 0)
	deletes := make([]string, 0)

	split := slice.Split(topos, int(filter.DefaultMaxInLimit))
	for _, parts := range split {
		partAdd, partUpdate, partDelete := ots.compareOrgTopo(kt, parts, dbOrgTopos)

		add = append(add, partAdd...)
		update = append(update, partUpdate...)
		deletes = append(deletes, partDelete...)
	}

	return add, update, deletes, nil
}

func (ots *orgTopoSvc) compareOrgTopo(kt *kit.Kit, topos []table.OrgTopo, exists []table.OrgTopo) (
	[]table.OrgTopo, []table.OrgTopo, []string) {

	addMap := make(map[string]table.OrgTopo)
	for _, topo := range topos {
		addMap[topo.DeptID] = table.OrgTopo{
			DeptID:      topo.DeptID,
			DeptName:    topo.DeptName,
			FullName:    topo.FullName,
			Level:       topo.Level,
			Parent:      topo.Parent,
			HasChildren: topo.HasChildren,
			Memo:        topo.Memo,
			Creator:     kt.User,
			Reviser:     kt.User,
		}
	}

	updates := make([]table.OrgTopo, 0)
	deletes := make([]string, 0)
	for _, exist := range exists {
		// 将已存在的条目，从待新增的列表中剔除
		candidate, ok := addMap[exist.DeptID]
		if !ok {
			deletes = append(deletes, exist.DeptID)
			continue
		}

		delete(addMap, exist.DeptID)

		// 对比条目，判断是否更新
		if ots.isOrgTopoUpdate(candidate, exist) {
			// do not update memo and creator
			candidate.ID = exist.ID
			candidate.Memo = exist.Memo
			candidate.Creator = ""
			updates = append(updates, candidate)
		}
	}

	adds := make([]table.OrgTopo, 0)
	for _, item := range addMap {
		adds = append(adds, item)
	}

	return adds, updates, deletes
}

func (ots *orgTopoSvc) isOrgTopoUpdate(candidate table.OrgTopo, exist table.OrgTopo) bool {
	if candidate.DeptName != exist.DeptName {
		return true
	}

	if candidate.FullName != exist.FullName {
		return true
	}

	if candidate.Level != exist.Level {
		return true
	}

	if candidate.Parent != exist.Parent {
		return true
	}

	if cvt.PtrToVal(candidate.HasChildren) != cvt.PtrToVal(exist.HasChildren) {
		return true
	}


	return false
}

// batchDeleteOrgTopo batch delete org topo.
func (ots *orgTopoSvc) batchDeleteOrgTopo(kt *kit.Kit, ids []string) error {
	split := slice.Split(ids, int(filter.DefaultMaxInLimit))
	for _, spartIDs := range split {
		tmpReq := &dataorgtopo.BatchDeleteOrgTopoReq{
			BatchDeleteReq: core.BatchDeleteReq{
				IDs: spartIDs,
			},
		}
		err := ots.client.DataService().Global.OrgTopo.BatchDelete(kt, tmpReq)
		if err != nil {
			logs.Errorf("batch delete org topo table failed, allNum: %d, tmpNum: %d, err: %+v, rid: %s",
				len(ids), len(spartIDs), err, kt.Rid)
			return err
		}
	}

	return nil
}
