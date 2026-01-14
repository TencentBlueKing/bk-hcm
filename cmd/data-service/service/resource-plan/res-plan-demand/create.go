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

// Package resplandemand ...
package resplandemand

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanDemand create resource plan demand
func (svc *service) BatchCreateResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	demandIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.batchCreateResPlanDemandWithTx(cts.Kit, txn, req.Demands)
		if err != nil {
			logs.Errorf("failed to batch create resource plan demand with tx, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create resource plan demand failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(demandIDs)
	if err != nil {
		logs.Errorf("create resource plan demand but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create resource plan demand but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *service) batchCreateResPlanDemandWithTx(kt *kit.Kit, txn *sqlx.Tx,
	createReqs []rpproto.ResPlanDemandCreateReq) (
	[]string, error) {

	models := make([]tablers.ResPlanDemandTable, len(createReqs))
	for idx, item := range createReqs {
		// 把字符串类型的[期望交付时间转]为符合格式的Int类型
		expectTimeInt, err := times.ConvStrTimeToInt(item.ExpectTime, constant.DateLayout)
		if err != nil {
			logs.Errorf("convert expect time to int64 failed, expect time: %s, err: %v, rid: %s", item.ExpectTime, err,
				kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		returnPlanTimeInt := 0
		if item.ObsProject == enumor.ObsProjectShortLease {
			returnPlanTimeInt, err = times.ConvStrTimeToInt(item.ReturnPlanTime, constant.DateLayout)
			if err != nil {
				logs.Errorf("convert return plan time to int64 failed, return plan time: %s, err: %v, rid: %s",
					item.ReturnPlanTime, err, kt.Rid)
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
		}

		coreType := enumor.CoreType(item.CoreType)
		if err := coreType.Validate(); err != nil {
			logs.Errorf("invalid core type: %s, rid: %s", coreType, kt.Rid)
			return nil, err
		}

		createT := tablers.ResPlanDemandTable{
			Locked:          cvt.ValToPtr(enumor.CrpDemandUnLocked),
			LockedCPUCore:   cvt.ValToPtr(int64(0)),
			BkBizID:         item.BkBizID,
			BkBizName:       item.BkBizName,
			OpProductID:     item.OpProductID,
			OpProductName:   item.OpProductName,
			PlanProductID:   item.PlanProductID,
			PlanProductName: item.PlanProductName,
			VirtualDeptID:   item.VirtualDeptID,
			VirtualDeptName: item.VirtualDeptName,
			DemandClass:     item.DemandClass,
			DemandResType:   item.DemandResType,
			ResMode:         item.ResMode,
			ObsProject:      item.ObsProject,
			ExpectTime:      expectTimeInt,
			ReturnPlanTime:  returnPlanTimeInt,
			PlanType:        item.PlanType,
			AreaName:        item.AreaName,
			RegionID:        item.RegionID,
			RegionName:      item.RegionName,
			ZoneID:          item.ZoneID,
			ZoneName:        item.ZoneName,
			TechnicalClass:  item.TechnicalClass,
			DeviceFamily:    item.DeviceFamily,
			DeviceClass:     item.DeviceClass,
			DeviceType:      item.DeviceType,
			CoreType:        coreType,
			DiskType:        item.DiskType,
			DiskTypeName:    item.DiskTypeName,
			OS:              &types.Decimal{Decimal: cvt.PtrToVal(item.OS)},
			CpuCore:         item.CpuCore,
			Memory:          item.Memory,
			DiskSize:        item.DiskSize,
			DiskIO:          item.DiskIO,
			Creator:         kt.User,
			Reviser:         kt.User,
		}

		// 系统创建的情况下，creator需要从参数中取
		if item.Creator != "" {
			createT.Creator = item.Creator
			createT.Reviser = item.Creator
		}

		models[idx] = createT
	}
	recordIDs, err := svc.dao.ResPlanDemand().CreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("create resource plan demand failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create resource plan demand failed, err: %v", err)
	}
	return recordIDs, nil
}
