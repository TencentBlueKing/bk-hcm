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

package disk

import (
	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/cmd/hc-service/service/disk/datasvc"
	"hcm/pkg/adaptor/types/disk"
	proto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateAzureDisk ...
func (svc *service) CreateAzureDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int32(req.DiskSize)
	opt := &disk.AzureDiskCreateOption{
		DiskName:          *req.DiskName,
		ResourceGroupName: req.Extension.ResourceGroupName,
		Region:            req.Region,
		Zone:              req.Zone,
		DiskType:          req.DiskType,
		DiskSize:          diskSize,
		DiskCount:         converter.ValToPtr(uint64(req.DiskCount)),
	}
	cloudIDs, err := client.CreateDisk(cts.Kit, opt)
	if err != nil {
		logs.Errorf("create azure cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.DataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         req.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          cloudIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncazure.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.BatchCreateResult{SuccessCloudIDs: cloudIDs}, nil
}

// DeleteAzureDisk ...
func (svc *service) DeleteAzureDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskData, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AzureDiskDeleteOption{
		ResourceGroupName: diskData.Extension.ResourceGroupName,
		DiskName:          diskData.Name,
	}

	client, err := svc.Adaptor.Azure(cts.Kit, diskData.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DeleteDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.DiskManager{DataCli: svc.DataCli}
	return nil, manager.Delete(cts.Kit, []string{req.DiskID})
}

// AttachAzureDisk ...
func (svc *service) AttachAzureDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskInfo, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmInfo, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AzureDiskAttachOption{
		ResourceGroupName: diskInfo.Extension.ResourceGroupName,
		CvmName:           cvmInfo.Name,
		DiskName:          diskInfo.Name,
		CachingType:       req.CachingType,
	}

	client, err := svc.Adaptor.Azure(cts.Kit, diskInfo.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.AttachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.DiskCvmRelManager{CvmID: req.CvmID, DiskID: req.DiskID, DataCli: svc.DataCli}
	err = manager.Create(cts.Kit)
	if err != nil {
		return nil, err
	}

	diskData, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.DataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         diskInfo.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{diskData.CloudID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncazure.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.Cvm(cts.Kit, params, &syncazure.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DetachAzureDisk ...
func (svc *service) DetachAzureDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskInfo, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmInfo, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AzureDiskDetachOption{
		ResourceGroupName: diskInfo.Extension.ResourceGroupName,
		CvmName:           cvmInfo.Name,
		DiskName:          diskInfo.Name,
	}

	client, err := svc.Adaptor.Azure(cts.Kit, diskInfo.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DetachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	diskData, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.DataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         diskInfo.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{diskData.CloudID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncazure.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncazure.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync azure cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
