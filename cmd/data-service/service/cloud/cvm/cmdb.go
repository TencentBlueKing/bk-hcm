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

package cvm

import (
	"hcm/cmd/data-service/service/cloud/logics/cmdb"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// upsertCmdbHosts upsert cmdb hosts. operators 由调用方按"创建/修改"语义构建（修改场景传 nil）。
// TODO add previous hosts params to transfer across biz when supported.
func upsertCmdbHosts[T corecvm.Extension](svc *cvmSvc, kt *kit.Kit, vendor enumor.Vendor, models []*cvm.Table,
	cvmIDOperatorMap map[string]*string) error {

	bizHostMap := make(map[int64][]corecvm.Cvm[T])
	for _, model := range models {
		if model.BkBizID == constant.UnassignedBiz || model.Vendor == enumor.Other {
			// ignore unassigned host. TODO delete unassigned host from cmdb when transfer back to resource supported.
			continue
		}

		host, err := convCvmGetResult[T](convTableToBaseCvm(model), model.Extension)
		if err != nil {
			logs.Errorf("conv cvm get result failed, err: %v, model: %+v, extension: %s, rid: %s", err, model,
				model.Extension, kt.Rid)
			return err
		}
		bizHostMap[model.BkBizID] = append(bizHostMap[model.BkBizID], converter.PtrToVal(host))
	}

	needCheckHostIDs := make([]int64, 0)
	for bizID, hosts := range bizHostMap {
		addCmdbReq := &cmdb.AddCloudHostToBizReq[T]{
			Vendor: vendor, BizID: bizID, Hosts: hosts, Operators: cvmIDOperatorMap}
		hostIDs, err := cmdb.AddCloudHostToBiz[T](svc.cmdbLogics, kt, addCmdbReq)
		if err != nil {
			logs.Errorf("[%s] add cmdb cloud hosts failed, err: %v, req: %+v, rid: %s", constant.CmdbSyncFailed, err,
				addCmdbReq, kt.Rid)
			return err
		}

		if len(hostIDs) != len(hosts) {
			logs.Errorf("[%s] add cmdb cloud hosts len(hostIDs[%d]) != len(hosts[%d]), req: %+v, rid: %s",
				constant.CmdbSyncFailed, len(hostIDs), len(hosts), addCmdbReq, kt.Rid)
			return err
		}
		needCheckHostIDs = append(needCheckHostIDs, hostIDs...)
		for i, host := range hosts {
			updateFilter := tools.EqualExpression("id", host.ID)
			updateField := &cvm.Table{BkHostID: hostIDs[i]}
			if err = svc.dao.Cvm().Update(kt, updateFilter, updateField); err != nil {
				logs.Errorf("[%s] update cvm failed, err: %v, filter: %+v, field: %+v, rid: %s",
					constant.CmdbSyncFailed, err, *updateFilter, updateField, kt.Rid)
				return err
			}
		}
	}

	if err := deleteOtherVendorHost(svc, kt, needCheckHostIDs); err != nil {
		logs.Errorf("delete other vendor host failed, err: %v, hostIDs: %+v, rid: %s", err, needCheckHostIDs, kt.Rid)
		return err
	}

	return nil
}

// upsertCmdbBaseHosts upsert cmdb hosts' basic info. operators 由调用方按"创建/修改"语义构建（修改场景传 nil）。
// TODO add previous hosts params to transfer across biz when supported.
func upsertBaseCmdbHosts(svc *cvmSvc, kt *kit.Kit, models []*cvm.Table, cvmIDOperatorMap map[string]*string) error {
	bizHostMap := make(map[int64][]corecvm.BaseCvm)
	for _, model := range models {
		if model.BkBizID == constant.UnassignedBiz || model.Vendor == enumor.Other {
			// ignore unassigned host. TODO delete unassigned host from cmdb when transfer back to resource supported.
			continue
		}

		bizHostMap[model.BkBizID] = append(bizHostMap[model.BkBizID], converter.PtrToVal(convTableToBaseCvm(model)))
	}

	logs.Infof("upsert cmdb base hosts, hostCount: %d, operatorCount: %d, rid: %s",
		len(models), len(cvmIDOperatorMap), kt.Rid)

	needCheckHostIDs := make([]int64, 0)
	for bizID, hosts := range bizHostMap {
		addCmdbReq := &cmdb.AddBaseCloudHostToBizReq{BizID: bizID, Hosts: hosts, Operators: cvmIDOperatorMap}
		hostIDs, err := cmdb.AddBaseCloudHostToBiz(svc.cmdbLogics, kt, addCmdbReq)
		if err != nil {
			logs.Errorf("[%s] add cmdb base cloud hosts failed, err: %v, req: %+v, rid: %s", constant.CmdbSyncFailed,
				err, addCmdbReq, kt.Rid)
			return err
		}

		if len(hostIDs) != len(hosts) {
			logs.Errorf("[%s] add cmdb base cloud hosts len(hostIDs[%d]) != len(hosts[%d]), req: %+v, rid: %s",
				constant.CmdbSyncFailed, len(hostIDs), len(hosts), addCmdbReq, kt.Rid)
			return err
		}
		needCheckHostIDs = append(needCheckHostIDs, hostIDs...)

		for i, host := range hosts {
			updateFilter := tools.EqualExpression("id", host.ID)
			updateField := &cvm.Table{BkHostID: hostIDs[i]}
			if err = svc.dao.Cvm().Update(kt, updateFilter, updateField); err != nil {
				logs.Errorf("[%s] update cvm failed, err: %v, filter: %+v, field: %+v, rid: %s",
					constant.CmdbSyncFailed, err, *updateFilter, updateField, kt.Rid)
				return err
			}
		}
	}

	if err := deleteOtherVendorHost(svc, kt, needCheckHostIDs); err != nil {
		logs.Errorf("delete other vendor host failed, err: %v, hostIDs: %+v, rid: %s", err, needCheckHostIDs, kt.Rid)
		return err
	}

	return nil
}

// buildCvmIDOperatorMap 为"分配主机到业务"链路按主机推导同步到 CMDB 的 operator（以 cvm id 为键），
// 由 SyncCvmToCmdb 与 BatchUpdateCvmCommonInfo（上层置 SetOperator 时）复用：购买机（creator 非后台用户）
// 取 creator；云上同步的增量机（creator 为后台用户）取二级账号第一个负责人。
//
// 保留 bk_host_id <= 0 的主机级守卫：分配链路可能重推账号下全部主机，已在 CMDB 的主机（bk_host_id>0）不入 map，
// 避免覆盖 CMDB 侧已有 operator。返回值仅包含需要下发 operator 的主机，未出现的经下层 omitempty 不下发该字段。
func buildCvmIDOperatorMap(svc *cvmSvc, kt *kit.Kit, models []*cvm.Table) (map[string]*string, error) {
	cvmIDOperatorMap := make(map[string]*string)
	// accountCvmIDs 记录后台创建（云上同步）主机所属二级账号到其 cvm id 列表的映射
	accountCvmIDs := make(map[string][]string)
	for _, model := range models {
		if model.BkBizID == constant.UnassignedBiz || model.Vendor == enumor.Other {
			// ignore unassigned host. TODO delete unassigned host from cmdb when transfer back to resource supported.
			continue
		}
		// bk_host_id>0 表示已在 CMDB，不下发 operator，避免覆盖 CMDB 侧已有值。
		if model.BkHostID > 0 {
			logs.Infof("skip build cmdb operator for host already in cmdb, bk_host_id: %d, "+
				"cvmID: %s, cloudID: %s, rid: %s", model.BkHostID, model.ID, model.CloudID, kt.Rid)
			continue
		}
		if model.Creator == "" {
			continue
		}

		if model.Creator == constant.BackendOperationUserKey {
			if model.AccountID == "" {
				logs.Warnf("skip cmdb operator for backend host without account, cvmID: %s, cloudID: %s, rid: %s",
					model.ID, model.CloudID, kt.Rid)
				continue
			}
			accountCvmIDs[model.AccountID] = append(accountCvmIDs[model.AccountID], model.ID)
			continue
		}
		cvmIDOperatorMap[model.ID] = &model.Creator
	}

	if len(accountCvmIDs) == 0 {
		return cvmIDOperatorMap, nil
	}

	// 需要用二级账号负责人替代BackendOperationUserKey作为operator
	accountIDs := make([]string, 0, len(accountCvmIDs))
	for accountID := range accountCvmIDs {
		accountIDs = append(accountIDs, accountID)
	}

	for _, subIDs := range slice.Split(accountIDs, constant.BatchOperationMaxLimit) {
		opt := &types.ListOption{
			Fields: []string{"id", "managers"},
			Filter: tools.ContainersExpression("id", subIDs),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := svc.dao.Account().List(kt, opt)
		if err != nil {
			logs.Errorf("list account for cmdb operator failed, err: %v, ids: %+v, rid: %s", err, subIDs, kt.Rid)
			return nil, err
		}
		for _, account := range result.Details {
			if len(account.Managers) == 0 {
				logs.Warnf("account managers empty for cmdb operator, accountID: %s, cvmIDs: %+v, rid: %s",
					account.ID, accountCvmIDs[account.ID], kt.Rid)
				continue
			}
			// 二级账号负责人取第一个 manager 作为 operator
			manager := account.Managers[0]
			for _, cvmID := range accountCvmIDs[account.ID] {
				cvmIDOperatorMap[cvmID] = &manager
			}
		}
	}

	return cvmIDOperatorMap, nil
}

func deleteOtherVendorHost(svc *cvmSvc, kt *kit.Kit, hostIDs []int64) error {
	if len(hostIDs) == 0 {
		return nil
	}

	deleteIDs := make([]string, 0)
	for _, subHostIDs := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		filter := tools.ExpressionAnd(tools.RuleIn("bk_host_id", subHostIDs), tools.RuleEqual("vendor", enumor.Other))
		opt := &types.ListOption{
			Fields: []string{"id"},
			Filter: filter,
			Page:   core.NewDefaultBasePage(),
		}
		result, err := svc.dao.Cvm().List(kt, opt)
		if err != nil {
			logs.Errorf("list cvm failed, err: %v, filter: %+v, rid: %s", err, converter.PtrToVal(filter), kt.Rid)
			return err
		}
		for _, detail := range result.Details {
			deleteIDs = append(deleteIDs, detail.ID)
		}
	}

	_, err := svc.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, ids := range slice.Split(deleteIDs, constant.BatchOperationMaxLimit) {
			filter := tools.ContainersExpression("id", ids)
			if err := svc.dao.Cvm().DeleteWithTx(kt, txn, filter); err != nil {
				logs.Errorf("delete cvm failed, err: %v, filter: %+v, rid: %s", err, converter.PtrToVal(filter), kt.Rid)
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		logs.Errorf("delete cvm failed, err: %v, delete hostIDs: %+v, rid: %s", err, deleteIDs, kt.Rid)
		return err
	}

	return nil
}

// deleteCmdbHosts delete cmdb hosts.
func deleteCmdbHosts(svc *cvmSvc, kt *kit.Kit, models []cvm.Table) error {
	delBizMap := make(map[int64]map[enumor.Vendor][]string)
	for _, one := range models {
		if one.BkBizID == constant.UnassignedBiz || one.Vendor == enumor.Other {
			continue
		}
		vendorMap, exists := delBizMap[one.BkBizID]
		if !exists {
			vendorMap = make(map[enumor.Vendor][]string)
		}
		vendorMap[one.Vendor] = append(vendorMap[one.Vendor], one.CloudID)
		delBizMap[one.BkBizID] = vendorMap
	}

	for bizID, vendorMap := range delBizMap {
		delCmdbFilter := &cmdb.DeleteCloudHostFromBizReq{BizID: bizID, VendorCloudIDs: vendorMap}
		if err := svc.cmdbLogics.DeleteCloudHostFromBiz(kt, delCmdbFilter); err != nil {
			logs.Errorf("[%s] delete cmdb cloud hosts failed, err: %v, req: %+v, rid: %s", constant.CmdbSyncFailed,
				err, delCmdbFilter, kt.Rid)
			return err
		}
	}

	return nil
}
