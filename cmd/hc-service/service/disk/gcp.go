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
	syncgcp "hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/service/disk/datasvc"
	"hcm/pkg/adaptor/types/disk"
	proto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateGcpDisk ...
func (svc *service) CreateGcpDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int64(req.DiskSize)
	opt := &disk.GcpDiskCreateOption{
		DiskName:  *req.DiskName,
		Region:    req.Region,
		Zone:      req.Zone,
		DiskType:  req.DiskType,
		DiskSize:  diskSize,
		DiskCount: converter.ValToPtr(uint64(req.DiskCount)),
	}
	result, err := client.CreateDisk(cts.Kit, opt)
	if err != nil {
		logs.Errorf("create gcp cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respData := &proto.BatchCreateResult{
		UnknownCloudIDs: result.UnknownCloudIDs,
		SuccessCloudIDs: result.SuccessCloudIDs,
		FailedCloudIDs:  result.FailedCloudIDs,
		FailedMessage:   result.FailedMessage,
	}

	if len(result.SuccessCloudIDs) == 0 {
		return respData, nil
	}

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: req.AccountID,
		CloudIDs:  result.SuccessCloudIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncgcp.SyncDiskOption{BootMap: nil,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// DeleteGcpDisk ...
func (svc *service) DeleteGcpDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskData, err := svc.DataCli.Gcp.RetrieveDisk(cts.Kit, req.DiskID)
	if err != nil {
		return nil, err
	}

	opt := &disk.GcpDiskDeleteOption{Zone: diskData.Zone, DiskName: diskData.Name}

	client, err := svc.Adaptor.Gcp(cts.Kit, diskData.AccountID)
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

// AttachGcpDisk ...
func (svc *service) AttachGcpDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	dataCli := svc.DataCli.Gcp

	diskInfo, err := dataCli.RetrieveDisk(cts.Kit, req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmInfo, err := dataCli.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	opt := &disk.GcpDiskAttachOption{
		Zone:       diskInfo.Zone,
		CvmName:    cvmInfo.Name,
		DiskName:   diskInfo.Name,
		DeviceName: req.DeviceName,
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, diskInfo.AccountID)
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

	diskData, err := svc.DataCli.Gcp.RetrieveDisk(cts.Kit, req.DiskID)
	if err != nil {
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: diskInfo.AccountID,
		CloudIDs:  []string{diskData.CloudID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncgcp.SyncDiskOption{BootMap: nil,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.Cvm(cts.Kit, params, &syncgcp.SyncCvmOption{Region: cvmData.Region,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DetachGcpDisk ...
func (svc *service) DetachGcpDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskInfo, err := svc.DataCli.Gcp.RetrieveDisk(cts.Kit, req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmInfo, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	var deviceName string
	for _, d := range cvmInfo.Extension.Disks {
		if d.CloudID == diskInfo.CloudID {
			deviceName = d.DeviceName
			break
		}
	}

	opt := &disk.GcpDiskDetachOption{
		Zone:       diskInfo.Zone,
		CvmName:    cvmInfo.Name,
		DeviceName: deviceName,
		DiskName:   diskInfo.Name,
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, diskInfo.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DetachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	diskData, err := svc.DataCli.Gcp.RetrieveDisk(cts.Kit, req.DiskID)
	if err != nil {
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: diskInfo.AccountID,
		CloudIDs:  []string{diskData.CloudID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncgcp.SyncDiskOption{BootMap: nil,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params,
		&syncgcp.SyncCvmWithRelResOption{Zone: cvmInfo.Zone, Region: cvmInfo.Region})
	if err != nil {
		logs.Errorf("sync gcp cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
