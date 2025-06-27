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
	"fmt"
	"net/http"

	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/azure"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// initAzureCvmService 初始化 Azure CVM 服务，注册相关的 HTTP 处理函数。
func (svc *cvmSvc) initAzureCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("CreateAzureCvm", http.MethodPost, "/vendors/azure/cvms/create", svc.CreateAzureCvm)
	h.Add("StartAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/start", svc.StartAzureCvm)
	h.Add("StopAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/stop", svc.StopAzureCvm)
	h.Add("RebootAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/reboot", svc.RebootAzureCvm)
	h.Add("DeleteAzureCvm", http.MethodDelete, "/vendors/azure/cvms/{id}", svc.DeleteAzureCvm)

	h.Load(cap.WebService)
}

// CreateAzureCvm 创建 Azure CVM 实例。
// 该函数首先解码并验证请求参数，然后调用 Azure 适配器创建 CVM，
// 创建成功后，会同步 CVM 及其关联资源到数据服务层。
// cts: REST 请求上下文。
// 返回: 创建的 CVM 的云上 ID 及可能发生的错误。
func (svc *cvmSvc) CreateAzureCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AzureCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	azureCli, err := svc.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudID, err := svc.createAzureCvm(cts.Kit, azureCli, req)
	if err != nil {
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.dataCli, azureCli)

	params := &syncazure.SyncBaseParams{
		AccountID:         req.AccountID,
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          []string{cloudID},
	}

	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncazure.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync azure cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &protocvm.AzureCreateResp{CloudID: cloudID}, nil
}

func (svc *cvmSvc) createAzureCvm(kt *kit.Kit, azureCli *azure.Azure, req *protocvm.AzureCreateReq) (
	string, error) {

	listImageReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Azure,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.Equal.Factory(),
					Value: req.CloudImageID,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	imageResult, err := svc.dataCli.Azure.ListImage(kt, listImageReq)
	if err != nil {
		return "", err
	}

	if len(imageResult.Details) == 0 {
		return "", fmt.Errorf("image: %s not found", req.CloudImageID)
	}

	image := imageResult.Details[0]
	createOpt := &typecvm.AzureCreateOption{
		ResourceGroupName: req.ResourceGroupName,
		Region:            req.Region,
		Name:              req.Name,
		Zones:             req.Zones,
		InstanceType:      req.InstanceType,
		Image: &typecvm.AzureImage{
			Offer:     image.Extension.Offer,
			Publisher: image.Extension.Publisher,
			Sku:       image.Extension.Sku,
			Version:   image.Name,
		},
		Username:             req.Username,
		Password:             req.Password,
		CloudSubnetID:        req.CloudSubnetID,
		CloudSecurityGroupID: req.CloudSecurityGroupID,
		OSDisk: &typecvm.AzureOSDisk{
			Name:   req.OSDisk.Name,
			SizeGB: req.OSDisk.SizeGB,
			Type:   req.OSDisk.Type,
		},
		DataDisk:         make([]typecvm.AzureDataDisk, len(req.DataDisk)),
		PublicIPAssigned: req.PublicIPAssigned,
	}
	for j, one := range req.DataDisk {
		createOpt.DataDisk[j] = typecvm.AzureDataDisk{
			Name:   one.Name,
			SizeGB: one.SizeGB,
			Type:   one.Type,
		}
	}
	cloudID, err := azureCli.CreateCvm(kt, createOpt)
	if err != nil {
		logs.Errorf("create cvm failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return cloudID, nil
}

// StartAzureCvm 启动指定的 Azure CVM 实例。
// 该函数首先根据 CVM ID 从数据服务层获取 CVM 信息，然后调用 Azure 适配器启动 CVM，
// 启动成功后，会同步 CVM 信息到数据服务层。
// cts: REST 请求上下文。
// 返回: nil 及可能发生的错误。
func (svc *cvmSvc) StartAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvmFromDB, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureStartOption{
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		Name:              cvmFromDB.Name,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.dataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         cvmFromDB.AccountID,
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		CloudIDs:          []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncazure.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync azure cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// StopAzureCvm 停止指定的 Azure CVM 实例。
// 该函数首先根据 CVM ID 从数据服务层获取 CVM 信息，然后调用 Azure 适配器停止 CVM，
// 停止成功后，会同步 CVM 信息到数据服务层。
// cts: REST 请求上下文。
// 返回: nil 及可能发生的错误。
func (svc *cvmSvc) StopAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protocvm.AzureStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmFromDB, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureStopOption{
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		Name:              cvmFromDB.Name,
		SkipShutdown:      req.SkipShutdown,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.dataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         cvmFromDB.AccountID,
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		CloudIDs:          []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncazure.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync azure cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// RebootAzureCvm 重启指定的 Azure CVM 实例。
// 该函数首先根据 CVM ID 从数据服务层获取 CVM 信息，然后调用 Azure 适配器重启 CVM，
// 重启成功后，会同步 CVM 信息到数据服务层。
// cts: REST 请求上下文。
// 返回: nil 及可能发生的错误。
func (svc *cvmSvc) RebootAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvmFromDB, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureRebootOption{
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		Name:              cvmFromDB.Name,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.dataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         cvmFromDB.AccountID,
		ResourceGroupName: cvmFromDB.Extension.ResourceGroupName,
		CloudIDs:          []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncazure.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync azure cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteAzureCvm 删除指定的 Azure CVM 实例。
// 该函数首先根据 CVM ID 从数据服务层获取 CVM 信息，然后调用 Azure 适配器删除 CVM，
// 删除成功后，会从数据服务层删除该 CVM 的记录。
// cts: REST 请求上下文。
// 返回: nil 及可能发生的错误。
func (svc *cvmSvc) DeleteAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protocvm.AzureDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvm, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureDeleteOption{
		ResourceGroupName: cvm.Extension.ResourceGroupName,
		Name:              cvm.Name,
		Force:             req.Force,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete azure cvm failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
