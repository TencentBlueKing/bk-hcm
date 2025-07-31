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
	"hcm/pkg/dal/table/cloud"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

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

	vpcCloudIDs := slice.Map(req.TargetGroups,
		func(g dataproto.TargetGroupBatchCreate[T]) string { return g.CloudVpcID })

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		vpcInfoMap, err := getVpcMapByIDs(cts.Kit, vpcCloudIDs)
		if err != nil {
			return nil, err
		}

		tgIDs := make([]string, 0, len(req.TargetGroups))
		for _, tgReq := range req.TargetGroups {
			// 创建目标组
			tgTable, err := convTargetGroupCreateReqToTable(cts.Kit, vendor, tgReq, vpcInfoMap)
			if err != nil {
				return nil, err
			}

			models := []*tablelb.LoadBalancerTargetGroupTable{tgTable}
			tgNewIDs, err := svc.dao.LoadBalancerTargetGroup().BatchCreateWithTx(cts.Kit, txn, models)
			if err != nil {
				logs.Errorf("[%s]fail to batch create target group, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
				return nil, fmt.Errorf("batch create target group failed, err: %v", err)
			}
			tgIDs = append(tgIDs, tgNewIDs...)

			// 添加RS
			if tgReq.RsList != nil {
				_, err = svc.batchCreateTargetWithGroupID(cts.Kit, txn, tgReq.AccountID, tgNewIDs[0], tgReq.RsList)
				if err != nil {
					logs.Errorf("[%s]fail to batch create target, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
					return nil, fmt.Errorf("batch create target failed, err: %v", err)
				}
			}
		}

		return tgIDs, nil
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
	vpcInfo, ok := vpcInfoMap[tg.CloudVpcID]
	if !ok {
		return nil, errf.Newf(errf.RecordNotFound, "cloudVpcID[%s] not found", tg.CloudVpcID)
	}

	targetGroup := &tablelb.LoadBalancerTargetGroupTable{
		Name:            tg.Name,
		Vendor:          vendor,
		AccountID:       tg.AccountID,
		BkBizID:         tg.BkBizID,
		TargetGroupType: tg.TargetGroupType,
		VpcID:           vpcInfo.ID,
		CloudVpcID:      vpcInfo.CloudID,
		Region:          tg.Region,
		Protocol:        tg.Protocol,
		Port:            tg.Port,
		Weight:          cvt.ValToPtr(tg.Weight),
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
		targetGroup.Weight = cvt.ValToPtr(int64(-1))
	}
	return targetGroup, nil
}
