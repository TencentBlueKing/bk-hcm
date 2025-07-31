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
	"net/http"

	syncgcp "hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/gcp"
	typecvm "hcm/pkg/adaptor/types/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *cvmSvc) initGcpCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateGcpCvm", http.MethodPost, "/vendors/gcp/cvms/batch/create", svc.BatchCreateGcpCvm)
	h.Add("StartGcpCvm", http.MethodPost, "/vendors/gcp/cvms/{id}/start", svc.StartGcpCvm)
	h.Add("StopGcpCvm", http.MethodPost, "/vendors/gcp/cvms/{id}/stop", svc.StopGcpCvm)
	h.Add("RebootGcpCvm", http.MethodPost, "/vendors/gcp/cvms/{id}/reboot", svc.RebootGcpCvm)
	h.Add("DeleteGcpCvm", http.MethodDelete, "/vendors/gcp/cvms/{id}", svc.DeleteGcpCvm)

	h.Load(cap.WebService)
}

// BatchCreateGcpCvm ...
func (svc *cvmSvc) BatchCreateGcpCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.GcpBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	gcpCli, err := svc.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	vpcSelfLink, err := svc.getVpcSelfLinkByCloudID(cts.Kit, req.CloudVpcID)
	if err != nil {
		return nil, err
	}

	subnetSelfLink, err := svc.getSubnetSelfLinkByCloudID(cts.Kit, req.CloudSubnetID)
	if err != nil {
		return nil, err
	}

	image, err := svc.getImageByCloudID(cts.Kit, req.CloudImageID)
	if err != nil {
		return nil, err
	}

	platform, err := gcp.GetSystemPlatformFromImagePlatforms(image.Extension.ProjectID)
	if err != nil {
		return nil, err
	}

	createOpt := &typecvm.GcpCreateOption{
		NamePrefix:          req.NamePrefix,
		Zone:                req.Zone,
		InstanceType:        req.InstanceType,
		CloudImageSelfLink:  image.Extension.SelfLink,
		Password:            req.Password,
		RequiredCount:       req.RequiredCount,
		RequestID:           req.RequestID,
		CloudVpcSelfLink:    vpcSelfLink,
		CloudSubnetSelfLink: subnetSelfLink,
		Description:         req.Description,
		ImageProjectType:    platform,
		SystemDisk:          req.SystemDisk,
		DataDisk:            req.DataDisk,
	}
	result, err := gcpCli.CreateCvm(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respData := &protocvm.BatchCreateResult{
		UnknownCloudIDs: result.UnknownCloudIDs,
		SuccessCloudIDs: result.SuccessCloudIDs,
		FailedCloudIDs:  result.FailedCloudIDs,
		FailedMessage:   result.FailedMessage,
	}
	if len(result.SuccessCloudIDs) == 0 {
		return respData, nil
	}

	syncClient := syncgcp.NewClient(svc.dataCli, gcpCli)
	params := &syncgcp.SyncBaseParams{
		AccountID: req.AccountID,
		CloudIDs:  result.SuccessCloudIDs,
	}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncgcp.SyncCvmWithRelResOption{Region: req.Region,
		Zone: req.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// StartGcpCvm ...
func (svc *cvmSvc) StartGcpCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvmFromDB, err := svc.dataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Gcp(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.GcpStartOption{
		Zone: cvmFromDB.Zone,
		Name: cvmFromDB.Name,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start gcp cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.dataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: cvmFromDB.AccountID,
		CloudIDs:  []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncgcp.SyncCvmOption{Region: cvmFromDB.Region,
		Zone: cvmFromDB.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// StopGcpCvm ...
func (svc *cvmSvc) StopGcpCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvmFromDB, err := svc.dataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Gcp(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.GcpStopOption{
		Zone: cvmFromDB.Zone,
		Name: cvmFromDB.Name,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop gcp cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.dataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: cvmFromDB.AccountID,
		CloudIDs:  []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncgcp.SyncCvmOption{Region: cvmFromDB.Region,
		Zone: cvmFromDB.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// RebootGcpCvm ...
func (svc *cvmSvc) RebootGcpCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvmFromDB, err := svc.dataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Gcp(cts.Kit, cvmFromDB.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.GcpResetOption{
		Zone: cvmFromDB.Zone,
		Name: cvmFromDB.Name,
	}
	if err = client.ResetCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reset gcp cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.dataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: cvmFromDB.AccountID,
		CloudIDs:  []string{cvmFromDB.CloudID},
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncgcp.SyncCvmOption{Region: cvmFromDB.Region,
		Zone: cvmFromDB.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteGcpCvm ...
func (svc *cvmSvc) DeleteGcpCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvm, err := svc.dataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Gcp(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.GcpDeleteOption{
		Zone: cvm.Zone,
		Name: cvm.Name,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete gcp cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete gcp cvm failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
