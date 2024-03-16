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
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	typesdao "hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateCLB 批量创建clb
func (svc *lbSvc) BatchCreateCLB(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateClb[corelb.TCloudClbExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}

}
func batchCreateClb[T corelb.Extension](cts *rest.Contexts, svc *lbSvc, vendor enumor.Vendor) (any, error) {
	req := new(dataproto.LoadBalancerBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.LoadBalancerTable, 0, len(req.Lbs))
		for _, lb := range req.Lbs {
			lbTable, err := convClbReqToTable(cts.Kit, vendor, lb)
			if err != nil {
				return nil, err
			}
			models = append(models, lbTable)
		}

		ids, err := svc.dao.LoadBalancer().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create clb, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create clb failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create clb but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func convClbReqToTable[T corelb.Extension](kt *kit.Kit, vendor enumor.Vendor, lb dataproto.ClbBatchCreate[T]) (
	*tablelb.LoadBalancerTable, error) {
	extension, err := json.MarshalToString(lb.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return &tablelb.LoadBalancerTable{
		CloudID:              lb.CloudID,
		Name:                 lb.Name,
		Vendor:               vendor,
		AccountID:            lb.AccountID,
		BkBizID:              lb.BkBizID,
		Region:               lb.Region,
		Zones:                lb.Zones,
		BackupZones:          lb.BackupZones,
		LBType:               lb.LoadBalancerType,
		VpcID:                lb.VpcID,
		CloudVpcID:           lb.CloudVpcID,
		SubnetID:             lb.SubnetID,
		CloudSubnetID:        lb.CloudSubnetID,
		PrivateIPv4Addresses: lb.PrivateIPv4Addresses,
		PrivateIPv6Addresses: lb.PrivateIPv6Addresses,
		PublicIPv4Addresses:  lb.PublicIPv4Addresses,
		PublicIPv6Addresses:  lb.PublicIPv6Addresses,
		Domain:               lb.Domain,
		Status:               lb.Status,
		Memo:                 lb.Memo,
		CloudCreatedTime:     lb.CloudCreatedTime,
		CloudStatusTime:      lb.CloudStatusTime,
		CloudExpiredTime:     lb.CloudExpiredTime,
		Extension:            types.JsonField(extension),
		Creator:              kt.User,
		Reviser:              kt.User,
	}, nil
}

// BatchCreateTargetGroup 批量创建目标组
func (svc *lbSvc) BatchCreateTargetGroup(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateTargetGroup[corelb.TCloudTargetGroupExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateTargetGroup[T corelb.TargetGroupExtension](cts *rest.Contexts,
	svc *lbSvc, vendor enumor.Vendor) (any, error) {

	req := new(dataproto.TargetGroupBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		vpcInfoMap, err := getVpcMapByIDs(cts.Kit, req.TargetGroups)
		if err != nil {
			return nil, err
		}
		models := make([]*tablelb.LoadBalancerTargetGroupTable, 0, len(req.TargetGroups))
		for _, lb := range req.TargetGroups {
			tgTable, err := convTargetGroupCreateReqToTable(cts.Kit, vendor, lb, vpcInfoMap)
			if err != nil {
				return nil, err
			}
			models = append(models, tgTable)
		}

		ids, err := svc.dao.LoadBalancerTargetGroup().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create target group, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create target group failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create target group but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func convTargetGroupCreateReqToTable[T corelb.TargetGroupExtension](kt *kit.Kit, vendor enumor.Vendor,
	tg dataproto.TargetGroupBatchCreate[T], vpcInfoMap map[string]cloud.VpcTable) (
	*tablelb.LoadBalancerTargetGroupTable, error) {

	extensionJSON, err := types.NewJsonField(tg.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	targetGroup := &tablelb.LoadBalancerTargetGroupTable{
		Name:            tg.Name,
		Vendor:          vendor,
		AccountID:       tg.AccountID,
		BkBizID:         tg.BkBizID,
		TargetGroupType: tg.TargetGroupType,
		VpcID:           tg.VpcID,
		CloudVpcID:      tg.CloudVpcID,
		Region:          tg.Region,
		Protocol:        tg.Protocol,
		Port:            tg.Port,
		Weight:          tg.Weight,
		HealthCheck:     tg.HealthCheck,
		Memo:            tg.Memo,
		Extension:       extensionJSON,
		Creator:         kt.User,
		Reviser:         kt.User,
	}
	if len(tg.TargetGroupType) == 0 {
		targetGroup.TargetGroupType = enumor.LocalTargetGroupType
	}
	if tg.Weight == 0 {
		targetGroup.Weight = -1
	}
	if len(tg.VpcID) > 0 && len(tg.CloudVpcID) == 0 {
		if vpcInfo, ok := vpcInfoMap[tg.VpcID]; ok {
			targetGroup.CloudVpcID = vpcInfo.CloudID
		}
	}
	return targetGroup, nil
}

func getVpcMapByIDs[T corelb.TargetGroupExtension](kt *kit.Kit, tgList []dataproto.TargetGroupBatchCreate[T]) (
	map[string]cloud.VpcTable, error) {

	vpcIDs := make([]string, 0)
	for _, item := range tgList {
		if len(item.VpcID) > 0 {
			vpcIDs = append(vpcIDs, item.VpcID)
		}
	}
	vpcOpt := &typesdao.ListOption{
		Filter: tools.ContainersExpression("id", vpcIDs),
		Page:   core.NewDefaultBasePage(),
	}
	vpcResult, err := svc.dao.Vpc().List(kt, vpcOpt)
	if err != nil {
		logs.Errorf("list vpc by ids failed, vpcIDs: %v, err: %v, rid: %s", vpcIDs, err, kt.Rid)
		return nil, fmt.Errorf("list vpc by ids failed, err: %v", err)
	}

	idMap := make(map[string]cloud.VpcTable, len(vpcResult.Details))
	for _, item := range vpcResult.Details {
		idMap[item.ID] = item
	}

	return idMap, nil
}
