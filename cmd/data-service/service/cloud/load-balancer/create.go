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
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

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
	return targetGroup, nil
}

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
	if len(cloudCvmIDs) > 0 {
		cvmReq := &typesdao.ListOption{
			Filter: tools.ContainersExpression("cloud_id", cloudCvmIDs),
			Page:   core.NewDefaultBasePage(),
		}
		cvmList, err := svc.dao.Cvm().List(kt, cvmReq)
		if err != nil {
			logs.Errorf("failed to list cvm, cloudIDs: %v, err: %v, rid: %s", cloudCvmIDs, err, kt.Rid)
			return nil, err
		}

		for _, item := range cvmList.Details {
			cvmMap[item.CloudID] = item
		}
	}

	for _, item := range rsList {
		tmpRs := &tablelb.LoadBalancerTargetTable{
			AccountID:     item.AccountID,
			InstType:      item.InstType,
			CloudInstID:   item.CloudInstID,
			TargetGroupID: item.TargetGroupID,
			// for local target group its cloud id is same as local id
			CloudTargetGroupID: item.TargetGroupID,
			Port:               item.Port,
			Weight:             item.Weight,
			Creator:            kt.User,
			Reviser:            kt.User,
		}
		if len(accountID) > 0 {
			tmpRs.AccountID = accountID
		}
		// 实例类型-CVM
		if item.InstType == enumor.CvmInstType {
			tmpRs.InstID = cvmMap[item.CloudInstID].ID
			tmpRs.InstName = cvmMap[item.CloudInstID].Name
			tmpRs.PrivateIPAddress = cvmMap[item.CloudInstID].PrivateIPv4Addresses
			tmpRs.PublicIPAddress = cvmMap[item.CloudInstID].PublicIPv4Addresses
			tmpRs.Zone = cvmMap[item.CloudInstID].Zone
		}

		rsModels = append(rsModels, tmpRs)
	}
	return svc.dao.LoadBalancerTarget().BatchCreateWithTx(kt, txn, rsModels)
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

// BatchCreateTCloudUrlRule 批量创建腾讯云url规则 纯规则条目创建，不校验监听器
func (svc *lbSvc) BatchCreateTCloudUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TCloudUrlRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("[ds] BatchCreateTCloudUrlRule request validate failed, err:%v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	models := make([]*tablelb.TCloudLbUrlRuleTable, 0, len(req.UrlRules))
	for _, rule := range req.UrlRules {

		ruleModel := &tablelb.TCloudLbUrlRuleTable{
			CloudID:            rule.CloudID,
			Name:               rule.Name,
			RuleType:           rule.RuleType,
			LbID:               rule.LbID,
			CloudLbID:          rule.CloudLbID,
			LblID:              rule.LblID,
			CloudLBLID:         rule.CloudLBLID,
			TargetGroupID:      rule.TargetGroupID,
			CloudTargetGroupID: rule.CloudTargetGroupID,
			Domain:             rule.Domain,
			URL:                rule.URL,
			Scheduler:          rule.Scheduler,
			SessionType:        rule.SessionType,
			SessionExpire:      rule.SessionExpire,
			Memo:               rule.Memo,

			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		}
		healthCheckJson, err := json.MarshalToString(rule.HealthCheck)
		if err != nil {
			logs.Errorf("fail to marshal health check into json, err: %v, healthcheck: %+v, rid: %s",
				err, rule.HealthCheck, cts.Kit.Rid)
			return nil, err
		}
		ruleModel.HealthCheck = types.JsonField(healthCheckJson)
		certJson, err := json.MarshalToString(rule.Certificate)
		if err != nil {
			logs.Errorf("fail to marshal certificate into json, err: %v, certificate: %+v, rid: %s",
				err, rule.Certificate, cts.Kit.Rid)
			return nil, err
		}
		ruleModel.Certificate = types.JsonField(certJson)
		models = append(models, ruleModel)
	}

	// 创建
	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		ids, err := svc.dao.LoadBalancerTCloudUrlRule().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("fail to batch create lb rule, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create lb rule failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create tcloud url rule but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchCreateListener 批量创建监听器
func (svc *lbSvc) BatchCreateListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateListener(cts, svc)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateListener(cts *rest.Contexts, svc *lbSvc) (any, error) {
	req := new(dataproto.ListenerBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.LoadBalancerListenerTable, 0, len(req.Listeners))
		for _, item := range req.Listeners {
			models = append(models, &tablelb.LoadBalancerListenerTable{
				CloudID:       item.CloudID,
				Name:          item.Name,
				Vendor:        item.Vendor,
				AccountID:     item.AccountID,
				BkBizID:       item.BkBizID,
				LBID:          item.LbID,
				CloudLBID:     item.CloudLbID,
				Protocol:      item.Protocol,
				Port:          item.Port,
				DefaultDomain: item.DefaultDomain,
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			})
		}
		ids, err := svc.dao.LoadBalancerListener().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("fail to batch create listener, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create listener failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create listener but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchCreateListenerWithRule 批量创建监听器及规则
func (svc *lbSvc) BatchCreateListenerWithRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudListenerWithRule(cts)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func (svc *lbSvc) batchCreateTCloudListenerWithRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ListenerWithRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids, err := svc.insertListenerWithRule(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *lbSvc) insertListenerWithRule(kt *kit.Kit, req *dataproto.ListenerWithRuleBatchCreateReq) ([]string, error) {
	result, err := svc.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		lblIDs := make([]string, 0, len(req.ListenerWithRules))
		for _, item := range req.ListenerWithRules {
			lblID, ruleID, err := svc.createListenerWithRule(kt, txn, item)
			if err != nil {
				logs.Errorf("fail to create listener with rule, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			lblIDs = append(lblIDs, lblID)
			if len(item.TargetGroupID) == 0 {
				continue
			}
			ruleRelModels := []*tablelb.TargetGroupListenerRuleRelTable{{
				ListenerRuleID:   ruleID,
				ListenerRuleType: item.RuleType,
				TargetGroupID:    item.TargetGroupID,
				LbID:             item.LbID,
				LblID:            lblID,
				BindingStatus:    enumor.SuccessBindingStatus,
				Creator:          kt.User,
				Reviser:          kt.User,
			}}
			_, err = svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(kt, txn, ruleRelModels)
			if err != nil {
				logs.Errorf("fail to batch create listener rule rel, err: %v, rid:%s", err, kt.Rid)
				return nil, fmt.Errorf("batch create listener rule rel failed, err: %v", err)
			}
		}
		return lblIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

func (svc *lbSvc) createListenerWithRule(kt *kit.Kit, txn *sqlx.Tx, item dataproto.ListenerWithRuleCreateReq) (
	lblID string, ruleID string, err error) {

	models := []*tablelb.LoadBalancerListenerTable{{
		CloudID:       item.CloudID,
		Name:          item.Name,
		Vendor:        item.Vendor,
		AccountID:     item.AccountID,
		BkBizID:       item.BkBizID,
		LBID:          item.LbID,
		CloudLBID:     item.CloudLbID,
		Protocol:      item.Protocol,
		Port:          item.Port,
		DefaultDomain: item.Domain,
		SniSwitch:     item.SniSwitch,
		Creator:       kt.User,
		Reviser:       kt.User,
	}}
	lblIDs, err := svc.dao.LoadBalancerListener().BatchCreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("fail to batch create listener, err: %v, rid:%s", err, kt.Rid)
		return "", "", fmt.Errorf("batch create listener failed, err: %v", err)
	}
	certJSON, err := json.MarshalToString(item.Certificate)
	if err != nil {
		logs.Errorf("json marshal Certificate failed, err: %v", err)
		return "", "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleModels := []*tablelb.TCloudLbUrlRuleTable{{
		CloudID:            item.CloudRuleID,
		RuleType:           item.RuleType,
		LbID:               item.LbID,
		CloudLbID:          item.CloudLbID,
		LblID:              lblIDs[0],
		CloudLBLID:         item.CloudID,
		TargetGroupID:      item.TargetGroupID,
		CloudTargetGroupID: item.CloudTargetGroupID,
		Domain:             item.Domain,
		URL:                item.Url,
		Scheduler:          item.Scheduler,
		SessionType:        item.SessionType,
		SessionExpire:      item.SessionExpire,
		Certificate:        types.JsonField(certJSON),
		Creator:            kt.User,
		Reviser:            kt.User,
	}}
	ruleIDs, err := svc.dao.LoadBalancerTCloudUrlRule().BatchCreateWithTx(kt, txn, ruleModels)
	if err != nil {
		logs.Errorf("fail to batch create listener url rule, err: %v, rid:%s", err, kt.Rid)
		return "", "", fmt.Errorf("batch create listener url rule failed, err: %v", err)
	}
	return lblIDs[0], ruleIDs[0], nil
}

// CreateResFlowLock 创建资源跟Flow的锁定关系
func (svc *lbSvc) CreateResFlowLock(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowLockCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		model := &tablelb.LoadBalancerFlowLockTable{
			ResID:   req.ResID,
			ResType: req.ResType,
			Owner:   req.Owner,
			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		}
		err := svc.dao.LoadBalancerFlowLock().CreateWithTx(cts.Kit, txn, model)
		if err != nil {
			logs.Errorf("[%s]fail to create load balancer flow lock, req: %+v, err: %v, rid:%s", req, err, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow lock failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchCreateResFlowRel 批量创建资源跟Flow的关系记录
func (svc *lbSvc) BatchCreateResFlowRel(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.LoadBalancerFlowRelTable, 0, len(req.ResFlowRels))
		for _, item := range req.ResFlowRels {
			models = append(models, &tablelb.LoadBalancerFlowRelTable{
				ResID:    item.ResID,
				FlowID:   item.FlowID,
				TaskType: item.TaskType,
				Status:   item.Status,
				Creator:  cts.Kit.User,
				Reviser:  cts.Kit.User,
			})
		}
		ids, err := svc.dao.LoadBalancerFlowRel().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create load balancer flow rel, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create load balancer flow rel failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create load balancer flow rel but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// ResFlowLock 锁定资源跟Flow
func (svc *lbSvc) ResFlowLock(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ResFlowLockReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		lockModel := &tablelb.LoadBalancerFlowLockTable{
			ResID:   req.ResID,
			ResType: req.ResType,
			Owner:   req.FlowID,
			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		}
		err := svc.dao.LoadBalancerFlowLock().CreateWithTx(cts.Kit, txn, lockModel)
		if err != nil {
			logs.Errorf("fail to create load balancer flow lock, req: %+v, err: %v, rid:%s", req, err, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow lock failed, err: %v", err)
		}

		relModels := []*tablelb.LoadBalancerFlowRelTable{{
			ResID:    req.ResID,
			FlowID:   req.FlowID,
			TaskType: req.TaskType,
			Status:   req.Status,
			Creator:  cts.Kit.User,
			Reviser:  cts.Kit.User,
		}}
		_, err = svc.dao.LoadBalancerFlowRel().BatchCreateWithTx(cts.Kit, txn, relModels)
		if err != nil {
			logs.Errorf("fail to create load balancer flow rel, err: %v, req: %+v, rid:%s", err, req, cts.Kit.Rid)
			return nil, fmt.Errorf("create load balancer flow rel failed, err: %v", err)
		}

		// 创建目标组的操作记录
		if req.ResType == enumor.TargetGroupCloudResType {
			err = svc.createTargetGroupOfResFlowAudit(cts.Kit, req, txn)
			if err != nil {
				logs.Errorf("fail to create res flow audits, err: %v, req: %+v, rid:%s", err, req, cts.Kit.Rid)
				return nil, fmt.Errorf("create res flow audits failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *lbSvc) createTargetGroupOfResFlowAudit(kt *kit.Kit, req *dataproto.ResFlowLockReq, txn *sqlx.Tx) error {
	resReq := &typesdao.ListOption{
		Filter: tools.EqualExpression("id", req.ResID),
		Page:   core.NewDefaultBasePage(),
	}
	resList, err := svc.dao.LoadBalancerTargetGroup().List(kt, resReq)
	if err != nil {
		return err
	}
	if len(resList.Details) == 0 {
		return errf.Newf(errf.RecordNotFound, "resID: %s, resType: %s, is not found", req.ResID, req.ResType)
	}

	resInfo := resList.Details[0]
	var auditData = struct {
		TargetGroup tablelb.LoadBalancerTargetGroupTable `json:"target_group"`
		ResFlow     *dataproto.ResFlowLockReq            `json:"res_flow"`
	}{
		TargetGroup: resInfo,
		ResFlow:     req,
	}

	audits := make([]*tableaudit.AuditTable, 0)
	audits = append(audits, &tableaudit.AuditTable{
		ResID:      resInfo.ID,
		CloudResID: resInfo.CloudID,
		ResName:    resInfo.Name,
		ResType:    enumor.AuditResourceType(req.ResType),
		Action:     enumor.Update,
		BkBizID:    resInfo.BkBizID,
		Vendor:     resInfo.Vendor,
		AccountID:  resInfo.AccountID,
		Operator:   kt.User,
		Source:     kt.GetRequestSource(),
		Rid:        kt.Rid,
		AppCode:    kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: auditData,
		},
	})
	if err = svc.dao.Audit().BatchCreateWithTx(kt, txn, audits); err != nil {
		logs.Errorf("batch create %s audit failed, err: %v, req: %+v, rid: %s", req.ResType, err, req, kt.Rid)
		return err
	}

	return nil
}

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
