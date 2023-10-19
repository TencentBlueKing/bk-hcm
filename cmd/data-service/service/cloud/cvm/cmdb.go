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
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// upsertCmdbHosts upsert cmdb hosts. TODO add previous hosts params to transfer across biz when supported.
func upsertCmdbHosts[T corecvm.Extension](svc *cvmSvc, kt *kit.Kit, vendor enumor.Vendor, models []*cvm.Table) error {
	bizHostMap := make(map[int64][]corecvm.Cvm[T])
	for _, model := range models {
		if model.BkBizID == constant.UnassignedBiz {
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

	for bizID, hosts := range bizHostMap {
		addCmdbReq := &cmdb.AddCloudHostToBizReq[T]{Vendor: vendor, BizID: bizID, Hosts: hosts}
		if err := cmdb.AddCloudHostToBiz[T](svc.cmdbLogics, kt, addCmdbReq); err != nil {
			logs.Errorf("[%s] add cmdb cloud hosts failed, err: %v, req: %+v, rid: %s", constant.CmdbSyncFailed, err,
				addCmdbReq, kt.Rid)
			return err
		}
	}

	return nil
}

// upsertCmdbBaseHosts upsert cmdb hosts' basic info.
// TODO add previous hosts params to transfer across biz when supported.
func upsertBaseCmdbHosts(svc *cvmSvc, kt *kit.Kit, models []*cvm.Table) error {
	bizHostMap := make(map[int64][]corecvm.BaseCvm)
	for _, model := range models {
		if model.BkBizID == constant.UnassignedBiz {
			// ignore unassigned host. TODO delete unassigned host from cmdb when transfer back to resource supported.
			continue
		}

		bizHostMap[model.BkBizID] = append(bizHostMap[model.BkBizID], converter.PtrToVal(convTableToBaseCvm(model)))
	}

	for bizID, hosts := range bizHostMap {
		addCmdbReq := &cmdb.AddBaseCloudHostToBizReq{BizID: bizID, Hosts: hosts}
		if err := cmdb.AddBaseCloudHostToBiz(svc.cmdbLogics, kt, addCmdbReq); err != nil {
			logs.Errorf("[%s] add cmdb base cloud hosts failed, err: %v, req: %+v, rid: %s", constant.CmdbSyncFailed,
				err, addCmdbReq, kt.Rid)
			return err
		}
	}

	return nil
}

// deleteCmdbHosts delete cmdb hosts.
func deleteCmdbHosts(svc *cvmSvc, kt *kit.Kit, models []cvm.Table) error {
	delBizMap := make(map[int64]map[enumor.Vendor][]string)
	for _, one := range models {
		if one.BkBizID == constant.UnassignedBiz {
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
