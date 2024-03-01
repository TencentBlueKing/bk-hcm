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

package clb

import (
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	typeclb "hcm/pkg/adaptor/types/clb"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func (svc *clbSvc) initTCloudClbService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateTCloudClb", http.MethodPost, "/vendors/tcloud/clbs/batch/create", svc.BatchCreateTCloudClb)
	h.Add("ListTCloudClb", http.MethodPost, "/vendors/tcloud/clbs/list", svc.ListTCloudClb)
	h.Add("BatchCreateTCloudClb", http.MethodPost, "/vendors/tcloud/clbs/batch/security_groups",
		svc.BatchSetTCloudClbSecurityGroup)

	h.Load(cap.WebService)
}

// BatchCreateTCloudClb ...
func (svc *clbSvc) BatchCreateTCloudClb(cts *rest.Contexts) (interface{}, error) {
	req := new(protoclb.TCloudBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	createOpt := &typeclb.TCloudCreateClbOption{
		Region:           req.Region,
		LoadBalancerType: req.LoadBalancerType,
		LoadBalancerName: req.Name,
		VpcID:            req.CloudVpcID,
		SubnetID:         req.CloudSubnetID,
		Vip:              req.Vip,
		VipIsp:           req.VipIsp,
		InternetAccessible: &tclb.InternetAccessible{
			InternetChargeType:      common.StringPtr(req.InternetChargeType),
			InternetMaxBandwidthOut: common.Int64Ptr(req.InternetMaxBandwidthOut),
		},
		BandwidthPackageID: req.BandwidthPackageID,
		SlaType:            req.SlaType,
		Number:             req.RequireCount,
	}
	// 负载均衡实例的网络类型-公网属性
	if req.LoadBalancerType == typeclb.OpenLoadBalancerType {
		// IP版本-仅适用于公网负载均衡
		createOpt.AddressIPVersion = req.AddressIPVersion
		// 静态单线IP 线路类型-仅适用于公网负载均衡, 如果不指定本参数，则默认使用BGP
		createOpt.VipIsp = req.VipIsp
		// 可用区ID-仅适用于公网负载均衡
		if len(req.Zones) > 0 {
			createOpt.ZoneID = req.Zones[0]
		}
		// 设置跨可用区容灾时的主可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 0 {
			createOpt.MasterZoneID = req.BackupZones[0]
		}
		// 设置跨可用区容灾时的备可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 1 {
			createOpt.SlaveZoneID = req.BackupZones[1]
		}
	}

	result, err := tcloud.CreateClb(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create tcloud clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respData := &protoclb.BatchCreateResult{
		UnknownCloudIDs: result.UnknownCloudIDs,
		SuccessCloudIDs: result.SuccessCloudIDs,
		FailedCloudIDs:  result.FailedCloudIDs,
		FailedMessage:   result.FailedMessage,
	}

	if len(result.SuccessCloudIDs) == 0 {
		return respData, nil
	}

	// TODO 补充CLB同步逻辑

	return respData, nil
}

// ListTCloudClb list tcloud clb
func (svc *clbSvc) ListTCloudClb(cts *rest.Contexts) (interface{}, error) {
	req := new(protoclb.TCloudListOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeclb.TCloudListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	result, err := tcloud.ListClb(cts.Kit, opt)
	if err != nil {
		logs.Errorf("[%s] list tcloud clb failed, req: %+v, err: %v, rid: %s", enumor.TCloud, req, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// BatchSetTCloudClbSecurityGroup 设置负载均衡实例的安全组
func (svc *clbSvc) BatchSetTCloudClbSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protoclb.TCloudSetClbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// 根据负载均衡ID，查询CLB基本信息
	clbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.LoadBalancerID),
		Page:   core.NewDefaultBasePage(),
	}
	clbResults, err := svc.dataCli.Global.LoadBalancer.ListClb(cts.Kit, clbReq)
	if err != nil {
		logs.Errorf("[%s] batch list clb failed. err: %v, accountID: %s, req: %+v, rid: %s",
			enumor.TCloud, err, req.AccountID, req, cts.Kit.Rid)
		return nil, err
	}

	if len(clbResults.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "[%s]clb id is not found", req.LoadBalancerID)
	}

	region := clbResults.Details[0].Region
	setOpt := &typeclb.TCloudSetClbSecurityGroupOption{
		Region:         region,
		LoadBalancerID: req.LoadBalancerID,
	}
	if len(req.SecurityGroups) > 0 {
		setOpt.SecurityGroups = req.SecurityGroups
	}

	result, err := tcloud.SetClbSecurityGroups(cts.Kit, setOpt)
	if err != nil {
		logs.Errorf("create tcloud clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
