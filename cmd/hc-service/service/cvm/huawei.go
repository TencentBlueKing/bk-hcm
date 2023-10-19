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

	synchuawei "hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/cmd/hc-service/service/capability"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataeip "hcm/pkg/api/data-service/cloud/eip"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *cvmSvc) initHuaWeiCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/create", svc.BatchCreateHuaWeiCvm)
	h.Add("BatchStartHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/start", svc.BatchStartHuaWeiCvm)
	h.Add("BatchStopHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/stop", svc.BatchStopHuaWeiCvm)
	h.Add("BatchRebootHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/reboot", svc.BatchRebootHuaWeiCvm)
	h.Add("BatchDeleteHuaWeiCvm", http.MethodDelete, "/vendors/huawei/cvms/batch", svc.BatchDeleteHuaWeiCvm)
	h.Add("BatchResetHuaWeiCvmPwd", http.MethodPost, "/vendors/huawei/cvms/batch/reset/pwd", svc.BatchResetHuaWeiCvmPwd)

	h.Load(cap.WebService)
}

// BatchCreateHuaWeiCvm ...
func (svc *cvmSvc) BatchCreateHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	huawei, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	createOpt := &typecvm.HuaWeiCreateOption{
		DryRun:                req.DryRun,
		Region:                req.Region,
		Name:                  req.Name,
		Zone:                  req.Zone,
		InstanceType:          req.InstanceType,
		CloudImageID:          req.CloudImageID,
		Password:              req.Password,
		RequiredCount:         req.RequiredCount,
		CloudSecurityGroupIDs: req.CloudSecurityGroupIDs,
		ClientToken:           req.ClientToken,
		CloudVpcID:            req.CloudVpcID,
		CloudSubnetID:         req.CloudSubnetID,
		Description:           req.Description,
		RootVolume:            req.RootVolume,
		DataVolume:            req.DataVolume,
		InstanceCharge:        req.InstanceCharge,
	}
	result, err := huawei.CreateCvm(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create huawei cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

	syncClient := synchuawei.NewClient(svc.dataCli, huawei)

	params := &synchuawei.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  result.SuccessCloudIDs,
	}

	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &synchuawei.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync huawei cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// BatchResetHuaWeiCvmPwd ...
func (svc *cvmSvc) BatchResetHuaWeiCvmPwd(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchResetPwdReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiResetPwdOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Password: req.Password,
	}
	if err = client.ResetCvmPwd(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reset huawei cvm pwd failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synchuawei.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync huawei cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStartHuaWeiCvm ...
func (svc *cvmSvc) BatchStartHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchStartReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synchuawei.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync huawei cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStopHuaWeiCvm ...
func (svc *cvmSvc) BatchStopHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiStopOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Force:    req.Force,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synchuawei.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync huawei cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRebootHuaWeiCvm ...
func (svc *cvmSvc) BatchRebootHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchRebootReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Force:    req.Force,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synchuawei.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync huawei cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteHuaWeiCvm ...
func (svc *cvmSvc) BatchDeleteHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	delCloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		delCloudIDs = append(delCloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiDeleteOption{
		Region:         req.Region,
		CloudIDs:       delCloudIDs,
		DeletePublicIP: req.DeletePublicIP,
		DeleteVolume:   req.DeleteDisk,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.IDs),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete huawei cvm failed, err: %v, ids: %v, rid: %s", err,
			req.IDs, cts.Kit.Rid)
		return nil, err
	}

	if req.DeletePublicIP {
		if err = svc.syncCvmRelEip(cts.Kit, req.AccountID, req.Region, req.IDs); err != nil {
			logs.Errorf("delete cvm success, but delete relation eip failed, err: %v, req: %v, rid: %s", err,
				req, cts.Kit.Rid)
			return nil, err
		}
	}

	if req.DeleteDisk {
		if err = svc.syncCvmRelDisk(cts.Kit, req.AccountID, req.Region, req.IDs); err != nil {
			logs.Errorf("delete cvm success, but delete relation disk failed, err: %v, req: %v, rid: %s", err,
				req, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (svc cvmSvc) syncCvmRelEip(kt *kit.Kit, accountID, region string, cvmIDs []string) error {
	listEipRel := &dataproto.EipCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rels, err := svc.dataCli.Global.ListEipCvmRel(kt.Ctx, kt.Header(), listEipRel)
	if err != nil {
		logs.Errorf("list eip_cvm_rel from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	eipIDs := make([]string, 0, len(rels.Details))
	for _, one := range rels.Details {
		eipIDs = append(eipIDs, one.EipID)
	}

	listEip := &dataeip.EipListReq{
		Filter: tools.ContainersExpression("id", eipIDs),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"cloud_id"},
	}
	eips, err := svc.dataCli.Global.ListEip(kt.Ctx, kt.Header(), listEip)
	if err != nil {
		logs.Errorf("list eip from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cloudIDs := make([]string, len(eips.Details))
	for _, one := range eips.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(kt, accountID)
	if err != nil {
		return err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Eip(kt, params, &synchuawei.SyncEipOption{})
	if err != nil {
		logs.Errorf("sync huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (svc cvmSvc) syncCvmRelDisk(kt *kit.Kit, accountID, region string, cvmIDs []string) error {
	listEipRel := &dataproto.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rels, err := svc.dataCli.Global.ListDiskCvmRel(kt.Ctx, kt.Header(), listEipRel)
	if err != nil {
		logs.Errorf("list disk_cvm_rel from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	diskIDs := make([]string, 0, len(rels.Details))
	for _, one := range rels.Details {
		diskIDs = append(diskIDs, one.DiskID)
	}

	listEip := &datadisk.DiskListReq{
		Filter: tools.ContainersExpression("id", diskIDs),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"cloud_id"},
	}
	disks, err := svc.dataCli.Global.ListDisk(kt.Ctx, kt.Header(), listEip)
	if err != nil {
		logs.Errorf("list disk from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cloudIDs := make([]string, len(disks.Details))
	for _, one := range disks.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(kt, accountID)
	if err != nil {
		return err
	}

	syncClient := synchuawei.NewClient(svc.dataCli, client)

	params := &synchuawei.SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Disk(kt, params, &synchuawei.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync huawei disk failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
