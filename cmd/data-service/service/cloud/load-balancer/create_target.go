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

package loadbalancer

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	typesdao "hcm/pkg/dal/dao/types"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchCreateTarget 批量创建目标
func (svc *lbSvc) BatchCreateTarget(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TargetBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		rsIDs, err := svc.batchCreateTargetWithGroupID(cts.Kit, txn, "", "", req.Targets)
		if err != nil {
			logs.Errorf("fail to batch create target, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create target failed, err: %v", err)
		}
		return rsIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create target but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// accountID 参数和tgID 参数 会覆盖rsList 中指定的参数. 对于cvm 类型数据会尝试查询对应的的cvm信息
func (svc *lbSvc) batchCreateTargetWithGroupID(kt *kit.Kit, txn *sqlx.Tx, accountID, tgID string,
	rsList []*dataproto.TargetBaseReq) ([]string, error) {

	rsModels := make([]*tablelb.LoadBalancerTargetTable, 0)
	cloudCvmIDs := make([]string, 0)
	for _, item := range rsList {
		if item.InstType == enumor.CvmInstType {
			cloudCvmIDs = append(cloudCvmIDs, item.CloudInstID)
		}
		if len(tgID) > 0 {
			item.TargetGroupID = tgID
		}
	}

	// 查询Cvm信息
	cvmMap := make(map[string]tablecvm.Table)
	for _, batchIds := range slice.Split(cloudCvmIDs, constant.BatchOperationMaxLimit) {
		cvmReq := &typesdao.ListOption{
			Filter: tools.ContainersExpression("cloud_id", batchIds),
			Page:   core.NewDefaultBasePage(),
		}
		cvmList, err := svc.dao.Cvm().List(kt, cvmReq)
		if err != nil {
			logs.Errorf("failed to list cvm, cloudIDs: %v, err: %v, rid: %s", batchIds, err, kt.Rid)
			return nil, err
		}

		for _, item := range cvmList.Details {
			cvmMap[item.CloudID] = item
		}
	}

	for _, item := range rsList {
		tmpRs := &tablelb.LoadBalancerTargetTable{
			AccountID:     item.AccountID,
			TargetGroupID: item.TargetGroupID,
			// for local target group its cloud id is same as local id
			CloudTargetGroupID: item.TargetGroupID,
			IP:                 item.IP,
			Port:               item.Port,
			Weight:             item.Weight,
			InstType:           item.InstType,
			InstID:             "",
			CloudInstID:        item.CloudInstID,
			InstName:           item.InstName,
			TargetGroupRegion:  item.TargetGroupRegion,
			PrivateIPAddress:   item.PrivateIPAddress,
			PublicIPAddress:    item.PublicIPAddress,
			CloudVpcIDs:        item.CloudVpcIDs,
			Zone:               item.Zone,
			Memo:               nil,
			Creator:            kt.User,
			Reviser:            kt.User,
		}
		// 实例类型-CVM
		if dbCvm, exists := cvmMap[item.CloudInstID]; exists && item.InstType == enumor.CvmInstType {
			tmpRs.InstID = dbCvm.ID
			tmpRs.InstName = dbCvm.Name
			tmpRs.PrivateIPAddress = dbCvm.PrivateIPv4Addresses
			tmpRs.PublicIPAddress = dbCvm.PublicIPv4Addresses
			tmpRs.Zone = dbCvm.Zone
			tmpRs.AccountID = dbCvm.AccountID
			tmpRs.CloudVpcIDs = dbCvm.CloudVpcIDs
		}
		if item.InstType == enumor.CcnInstType {
			tmpRs.InstID = tmpRs.CloudInstID
			tmpRs.AccountID = item.AccountID
		}

		rsModels = append(rsModels, tmpRs)
	}
	ids := make([]string, 0, len(rsModels))
	for batchIdx, rsBatch := range slice.Split(rsModels, constant.BatchOperationMaxLimit) {
		batchCreated, err := svc.dao.LoadBalancerTarget().BatchCreateWithTx(kt, txn, rsBatch)
		if err != nil {
			logs.Errorf("batch create target failed, batch idx: %d, err: %v, rid: %s", batchIdx, err, kt.Rid)
			return nil, err
		}
		ids = append(ids, batchCreated...)
	}
	return ids, nil
}
