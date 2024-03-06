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
	"fmt"
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	typeclb "hcm/pkg/adaptor/types/clb"
	adcore "hcm/pkg/adaptor/types/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	"hcm/pkg/api/data-service/cloud"
	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

func (svc *clbSvc) initTCloudClbService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateTCloudClb", http.MethodPost, "/vendors/tcloud/clbs/batch/create", svc.BatchCreateTCloudClb)
	h.Add("ListTCloudClb", http.MethodPost, "/vendors/tcloud/clbs/list", svc.ListTCloudClb)
	h.Add("TCloudDescribeResources", http.MethodPost, "/vendors/tcloud/clbs/resources/describe", svc.TCloudDescribeResources)

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

		InternetChargeType:      req.InternetChargeType,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,

		BandwidthPackageID: req.BandwidthPackageID,
		SlaType:            req.SlaType,
		Number:             req.RequireCount,
		ClientToken:        converter.StrNilPtr(cts.Kit.Rid),
	}
	// 负载均衡实例的网络类型-公网属性
	if req.LoadBalancerType == typeclb.OpenLoadBalancerType {
		// IP版本-仅适用于公网负载均衡
		createOpt.AddressIPVersion = req.AddressIPVersion
		// 静态单线IP 线路类型-仅适用于公网负载均衡, 如果不指定本参数，则默认使用BGP
		createOpt.VipIsp = req.VipIsp
		// 可用区ID-仅适用于公网负载均衡
		if len(req.Zones) > 0 {
			createOpt.ZoneID = converter.ValToPtr(req.Zones[0])
		}
		// 设置跨可用区容灾时的主可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 0 {
			createOpt.MasterZoneID = converter.ValToPtr(req.BackupZones[0])
		}
		// 设置跨可用区容灾时的备可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 1 {
			createOpt.SlaveZoneID = converter.ValToPtr(req.BackupZones[1])
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
	dbCreateReq := &cloud.TCloudCLBCreateReq{
		Clbs: make([]cloud.ClbBatchCreate[coreclb.TCloudClbExtension], 0, len(result.SuccessCloudIDs)),
	}
	// 预创建数据库记录
	for i, cloudID := range result.SuccessCloudIDs {
		var name = converter.PtrToVal(createOpt.LoadBalancerName)
		if converter.PtrToVal(req.RequireCount) > 1 {
			name = name + fmt.Sprintf("-%d", i+1)
		}
		dbCreateReq.Clbs = append(dbCreateReq.Clbs, cloud.ClbBatchCreate[coreclb.TCloudClbExtension]{
			BkBizID:          constant.UnassignedBiz,
			CloudID:          cloudID,
			Name:             name,
			Vendor:           string(enumor.TCloud),
			LoadBalancerType: string(req.LoadBalancerType),
			AccountID:        req.AccountID,
			Zones:            req.Zones,
			Region:           req.Region,
		})
	}

	_, err = svc.dataCli.Global.LoadBalancer.BatchCreateTCloudClb(cts.Kit, dbCreateReq)
	if err != nil {
		logs.Errorf("fail to pre-insert clb record to db, err: %v , rid: %s", err, cts.Kit.Rid)
		// still try to sync
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
		logs.Errorf("[%s] list tcloud clb failed, req: %+v, err: %v, rid: %s",
			enumor.TCloud, req, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *clbSvc) TCloudDescribeResources(cts *rest.Contexts) (any, error) {
	req := new(protoclb.TCloudDescribeResourcesOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	return client.DescribeResources(cts.Kit, req.TCloudDescribeResourcesOption)
}
