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

// BatchCreateLoadBalancer 批量创建负载均衡
func (svc *lbSvc) BatchCreateLoadBalancer(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateLoadBalancer[corelb.TCloudClbExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}

}
func batchCreateLoadBalancer[T corelb.Extension](cts *rest.Contexts, svc *lbSvc, vendor enumor.Vendor) (any, error) {
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
			logs.Errorf("[%s]fail to batch create load balancer, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create load balancer failed, err: %v", err)
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

func convClbReqToTable[T corelb.Extension](kt *kit.Kit, vendor enumor.Vendor, lb dataproto.LbBatchCreate[T]) (
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
		IPVersion:            string(lb.IPVersion),
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
		SyncTime:             lb.SyncTime,
		Extension:            types.JsonField(extension),
		Tags:                 types.StringMap(lb.Tags),
		Creator:              kt.User,
		Reviser:              kt.User,
		BandWidth:            lb.BandWidth,
		Isp:                  lb.Isp,
	}, nil
}

func getVpcMapByIDs(kt *kit.Kit, cloudIDs []string) (
	map[string]cloud.VpcTable, error) {

	vpcOpt := &typesdao.ListOption{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.NewDefaultBasePage(),
	}
	vpcResult, err := svc.dao.Vpc().List(kt, vpcOpt)
	if err != nil {
		logs.Errorf("list vpc by ids failed, vpcCloudIDs: %v, err: %v, rid: %s", cloudIDs, err, kt.Rid)
		return nil, fmt.Errorf("list vpc by cloudIDs failed, err: %v", err)
	}

	idMap := make(map[string]cloud.VpcTable, len(vpcResult.Details))
	for _, item := range vpcResult.Details {
		idMap[item.CloudID] = item
	}

	return idMap, nil
}

// CreateTargetGroupListenerRel 批量创建目标组与监听器的绑定关系
func (svc *lbSvc) CreateTargetGroupListenerRel(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TargetGroupListenerRelCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if len(req.CloudTargetGroupID) == 0 {
			return nil, errf.Newf(errf.InvalidParameter, "cloud_target_group_id can not empty")
		}
		ruleModel := &tablelb.TCloudLbUrlRuleTable{
			TargetGroupID:      req.TargetGroupID,
			CloudTargetGroupID: req.CloudTargetGroupID,
			Reviser:            cts.Kit.User,
		}
		err := svc.dao.LoadBalancerTCloudUrlRule().UpdateByIDWithTx(cts.Kit, txn, req.ListenerRuleID, ruleModel)
		if err != nil {
			return nil, err
		}

		models := make([]*tablelb.TargetGroupListenerRuleRelTable, 0)
		models = append(models, &tablelb.TargetGroupListenerRuleRelTable{
			Vendor:              req.Vendor,
			ListenerRuleID:      req.ListenerRuleID,
			CloudListenerRuleID: req.CloudListenerRuleID,
			ListenerRuleType:    req.ListenerRuleType,
			TargetGroupID:       req.TargetGroupID,
			CloudTargetGroupID:  req.CloudTargetGroupID,
			LbID:                req.LbID,
			CloudLbID:           req.CloudLbID,
			LblID:               req.LblID,
			CloudLblID:          req.CloudLblID,
			BindingStatus:       req.BindingStatus,
			Detail:              req.Detail,
			Creator:             cts.Kit.User,
			Reviser:             cts.Kit.User,
		})
		ids, err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create target group listener rel, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create target group listener rel failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create target group listener rel but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
