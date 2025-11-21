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

package host

import (
	"fmt"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/dissolve/host"
	define "hcm/pkg/dal/table/dissolve/host"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/caiche"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// RecycledHost provides interface for operations of recycle host.
type RecycledHost interface {
	Create(kt *kit.Kit, hosts []define.RecycleHostTable) ([]string, error)
	Update(kt *kit.Kit, host *define.RecycleHostTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*host.ListRecycleHostDetails, error)
	Delete(kt *kit.Kit, ids []string) error
	IsDissolveHost(kt *kit.Kit, assetIDs []string) (map[string]bool, error)
	Sync(kt *kit.Kit) error
}

type logics struct {
	dao          dao.Set
	thirdCli     *thirdparty.Client
	projectIDs   []int
	svrTypeNames []string
}

// New create recycle host logics.
func New(dao dao.Set, thirdCli *thirdparty.Client, projectIDs []int, svrTypeNames []string) RecycledHost {

	return &logics{
		dao:          dao,
		thirdCli:     thirdCli,
		projectIDs:   projectIDs,
		svrTypeNames: svrTypeNames,
	}
}

// Create recycle host.
func (l *logics) Create(kt *kit.Kit, hosts []define.RecycleHostTable) ([]string, error) {
	result, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		res, err := l.dao.RecycleHost().CreateWithTx(kt, txn, hosts)
		if err != nil {
			logs.Errorf("create recycle host failed, err: %v, data: %+v, rid: %s", err, hosts, kt.Rid)
			return nil, err
		}

		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// Update recycle host.
func (l *logics) Update(kt *kit.Kit, host *define.RecycleHostTable) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleHost().UpdateWithTx(kt, txn, tools.EqualExpression("id", host.ID), host)
		if err != nil {
			logs.Errorf("update recycle host failed, err: %v, data: %+v, rid: %s", err, host, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// List recycle host.
func (l *logics) List(kt *kit.Kit, opt *types.ListOption) (
	*host.ListRecycleHostDetails, error) {

	return l.dao.RecycleHost().List(kt, opt)
}

// Delete recycle host.
func (l *logics) Delete(kt *kit.Kit, ids []string) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleHost().DeleteWithTx(kt, txn, tools.ContainersExpression("id", ids))
		if err != nil {
			logs.Errorf("delete recycle host failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// IsDissolveHost check if host is dissolve host.
func (l *logics) IsDissolveHost(kt *kit.Kit, assetIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(assetIDs))
	for _, id := range assetIDs {
		result[id] = false
	}

	for _, ids := range slice.Split(assetIDs, int(core.DefaultMaxPageLimit)) {
		req := &types.ListOption{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("asset_id", ids),
				tools.RuleNotEqual("abolish_phase", enumor.Complete),
			),
			Fields: []string{"asset_id"},
			Page:   core.NewDefaultBasePage(),
		}

		list, err := l.List(kt, req)
		if err != nil {
			logs.Errorf("list recycle host failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		for _, one := range list.Details {
			result[converter.PtrToVal(one.AssetID)] = true
		}
	}

	return result, nil
}

// Sync recycle host.
func (l *logics) Sync(kt *kit.Kit) error {
	start := time.Now()
	logs.Infof("start sync recycle host, time: %v, rid: %s", start, kt.Rid)

	dbHosts, err := l.getAllHostFromDB(kt)
	if err != nil {
		logs.Errorf("get recycle host from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	caiCheHosts, err := l.getAllHostFromCaiChe(kt)
	if err != nil {
		logs.Errorf("get recycle host from caiche failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	create, update, deleteIDs := diff(dbHosts, caiCheHosts)
	if len(deleteIDs) != 0 {
		for _, split := range slice.Split(deleteIDs, int(core.DefaultMaxPageLimit)) {
			if err = l.Delete(kt, split); err != nil {
				logs.Errorf("delete recycle host failed, err: %v, ids: %+v, rid: %s", err, split, kt.Rid)
				return err
			}
		}
	}

	for _, one := range update {
		if err = l.Update(kt, &one); err != nil {
			logs.Errorf("update recycle host failed, err: %v, data: %+v, rid: %s", err, one, kt.Rid)
			return err
		}
	}

	if len(create) != 0 {
		for _, split := range slice.Split(create, int(core.DefaultMaxPageLimit)) {
			if _, err = l.Create(kt, split); err != nil {
				logs.Errorf("create recycle host failed, err: %v, ids: %+v, rid: %s", err, split, kt.Rid)
				return err
			}
		}
	}

	end := time.Now()
	logs.Infof("end sync recycle host, time: %v, cost: %v, create: %d, update: %d, delete: %d, rid: %s", end,
		end.Sub(start), len(create), len(update), len(deleteIDs), kt.Rid)

	return nil
}

func (l *logics) getAllHostFromDB(kt *kit.Kit) ([]define.RecycleHostTable, error) {
	hosts := make([]define.RecycleHostTable, 0)
	req := &types.ListOption{Filter: tools.AllExpression(),
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}}

	for {
		result, err := l.List(kt, req)
		if err != nil {
			logs.Errorf("get recycle host failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Details...)

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return hosts, nil
}

type version string

const (
	v1 version = "v1"
	v2 version = "v2"
)

type hostWithOtherInfo struct {
	define.RecycleHostTable
	abolishTime string
	version     version
}

// getAllHostFromCaiChe 从裁撤系统同步机器，目前通过旧接口同步年度裁撤，通过新接口同步云徙裁撤，后续待裁撤系统重构完，统一成一个接口
func (l *logics) getAllHostFromCaiChe(kt *kit.Kit) ([]define.RecycleHostTable, error) {
	hostMap := make(map[string]hostWithOtherInfo)
	virtualDepartmentName := []string{"IEG_Global", "IEG技术运营部"}
	abolishPhase := []enumor.AbolishPhase{enumor.Incomplete, enumor.Complete, enumor.BsiComplete}

	v2Req := &caiche.ListDeviceV2Req{
		PageNo:   1, // 裁撤系统这个api是从1开始的
		PageSize: core.DefaultMaxPageLimit,
		Filter: map[string]interface{}{
			"virtualDepartmentName": virtualDepartmentName,
			"projectId":             l.projectIDs,
			"abolishPhase":          abolishPhase,
			"svrTypeName":           l.svrTypeNames,
		},
	}
	start := time.Now()
	logs.Infof("start sync recycle host v2, time: %v, rid: %s", start, kt.Rid)
	for {
		resp, err := l.thirdCli.CaiChe.ListDeviceV2(kt, v2Req)
		if err != nil {
			logs.Errorf("get recycle v2 host failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(v2Req),
				kt.Rid)
			return nil, err
		}

		hostMap, err = addHost(kt, hostMap, transferToHostV2(resp.Data))
		if err != nil {
			logs.Errorf("add v2 host failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(resp.Data) < int(core.DefaultMaxPageLimit) {
			break
		}

		v2Req.PageNo++
	}
	end := time.Now()
	logs.Infof("end sync recycle host v2, time: %v, cost: %v, rid: %s", end, end.Sub(start), kt.Rid)

	result := make([]define.RecycleHostTable, 0, len(hostMap))
	for _, val := range hostMap {
		result = append(result, val.RecycleHostTable)
	}
	return result, nil
}

func transferToHost(devices []caiche.Device) []hostWithOtherInfo {
	hosts := make([]hostWithOtherInfo, 0, len(devices))
	for _, device := range devices {
		data := hostWithOtherInfo{
			RecycleHostTable: define.RecycleHostTable{
				AssetID:      converter.ValToPtr(device.SvrAssetId),
				InnerIP:      converter.ValToPtr(device.ServerLanIP),
				Module:       converter.ValToPtr(device.Module),
				AbolishPhase: converter.ValToPtr(device.AbolishPhase),
				ProjectName:  converter.ValToPtr(device.ProjectName),
			},
			abolishTime: device.AbolishDate,
			version:     v1,
		}
		hosts = append(hosts, data)
	}

	return hosts
}

func transferToHostV2(devices []caiche.DeviceV2) []hostWithOtherInfo {
	hosts := make([]hostWithOtherInfo, 0, len(devices))
	for _, device := range devices {
		data := hostWithOtherInfo{
			RecycleHostTable: define.RecycleHostTable{
				AssetID:      converter.ValToPtr(device.SerAssetID),
				InnerIP:      converter.ValToPtr(device.ServerLanIP),
				Module:       converter.ValToPtr(device.ModName),
				AbolishPhase: converter.ValToPtr(device.AbolishPhase),
				ProjectName:  converter.ValToPtr(device.ProjectName),
			},
			abolishTime: device.AbolishTime,
			version:     v2,
		}

		hosts = append(hosts, data)
	}

	return hosts
}

func addHost(kt *kit.Kit, hostMap map[string]hostWithOtherInfo, hosts []hostWithOtherInfo) (
	map[string]hostWithOtherInfo, error) {

	for _, newHost := range hosts {
		assetID := converter.PtrToVal(newHost.AssetID)
		oldHost, ok := hostMap[assetID]
		if !ok {
			hostMap[assetID] = newHost
			continue
		}

		newHostAbolishPhase := converter.PtrToVal(newHost.AbolishPhase)
		oldHostAbolishPhase := converter.PtrToVal(oldHost.AbolishPhase)

		if newHostAbolishPhase != enumor.Complete && oldHostAbolishPhase != enumor.Complete {
			logs.Errorf("host is invalid, two data abolishPhase not equal complete, assetID: %s, rid: %s", assetID,
				kt.Rid)
			return nil, fmt.Errorf("host is invalid, two data abolishPhase not equal complete, assetID: %s", assetID)
		}

		if newHostAbolishPhase != enumor.Complete {
			hostMap[assetID] = newHost
			continue
		}

		if oldHostAbolishPhase != enumor.Complete {
			continue
		}

		newHostAbolishTime, err := parseTime(newHost.version, newHost.abolishTime)
		if err != nil {
			logs.Errorf("parse new host abolish time failed, err: %v, abolish time: %s, assetID: %s, rid: %s", err,
				newHost.abolishTime, assetID, kt.Rid)
			return nil, err
		}
		oldHostAbolishTime, err := parseTime(oldHost.version, oldHost.abolishTime)
		if err != nil {
			logs.Errorf("parse old host abolish time failed, err: %v, abolish time: %s, assetID: %s, rid: %s", err,
				oldHost.abolishTime, assetID, kt.Rid)
			return nil, err
		}

		if newHostAbolishTime.After(oldHostAbolishTime) {
			hostMap[assetID] = newHost
		}
	}

	return hostMap, nil
}

func parseTime(version version, abolishTime string) (time.Time, error) {
	switch version {
	case v1:
		return time.Parse(constant.DateLayout, abolishTime)
	case v2:
		return time.Parse(constant.DateTimeLayout, abolishTime)
	default:
		return time.Time{}, fmt.Errorf("invalid version: %s", version)
	}
}

func diff(dbHosts []define.RecycleHostTable, caiCheHosts []define.RecycleHostTable) (
	[]define.RecycleHostTable, []define.RecycleHostTable, []string) {

	dbMap := make(map[string]define.RecycleHostTable, len(dbHosts))
	for _, one := range dbHosts {
		dbMap[converter.PtrToVal(one.AssetID)] = one
	}

	create := make([]define.RecycleHostTable, 0)
	update := make([]define.RecycleHostTable, 0)
	for _, newHost := range caiCheHosts {
		dbHost, exist := dbMap[converter.PtrToVal(newHost.AssetID)]
		if !exist {
			create = append(create, newHost)
			continue
		}

		delete(dbMap, converter.PtrToVal(newHost.AssetID))

		if isChange(newHost, dbHost) {
			newHost.ID = dbHost.ID
			update = append(update, newHost)
		}
	}

	deleteIDs := make([]string, 0)
	for _, one := range dbMap {
		deleteIDs = append(deleteIDs, one.ID)
	}

	return create, update, deleteIDs
}

func isChange(new, old define.RecycleHostTable) bool {
	if converter.PtrToVal(new.InnerIP) != converter.PtrToVal(old.InnerIP) {
		return true
	}

	if converter.PtrToVal(new.Module) != converter.PtrToVal(old.Module) {
		return true
	}

	if converter.PtrToVal(new.AbolishPhase) != converter.PtrToVal(old.AbolishPhase) {
		return true
	}

	if converter.PtrToVal(new.ProjectName) != converter.PtrToVal(old.ProjectName) {
		return true
	}

	return false
}
