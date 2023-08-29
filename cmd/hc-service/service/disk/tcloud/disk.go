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

package tcloud

import (
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/disk/datasvc"
	"hcm/pkg/adaptor/types/disk"
	proto "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// DiskSvc ...
type DiskSvc struct {
	Adaptor *cloudclient.CloudAdaptorClient
	DataCli *dataservice.Client
}

// CountDisk ...
func (svc *DiskSvc) CountDisk(cts *rest.Contexts) (interface{}, error) {
	return nil, nil
}

// CreateDisk ...
func (svc *DiskSvc) CreateDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	diskCount := uint64(req.DiskCount)
	opt := &disk.TCloudDiskCreateOption{
		DiskName:       req.DiskName,
		Region:         req.Region,
		Zone:           req.Zone,
		DiskType:       req.DiskType,
		DiskSize:       &req.DiskSize,
		DiskCount:      &diskCount,
		DiskChargeType: req.Extension.DiskChargeType,
	}

	if prepaid := req.Extension.DiskChargePrepaid; prepaid != nil {
		opt.DiskChargePrepaid = &disk.TCloudDiskChargePrepaid{
			Period:    prepaid.Period,
			RenewFlag: prepaid.RenewFlag,
		}
	}

	result, err := client.CreateDisk(cts.Kit, opt)
	if err != nil {
		logs.Errorf("create tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

	syncClient := synctcloud.NewClient(svc.DataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  result.SuccessCloudIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &synctcloud.SyncDiskOption{})
	if err != nil {
		logs.Errorf("sync tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// DeleteDisk ...
func (svc *DiskSvc) DeleteDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeDiskDeleteOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.TCloud(cts.Kit, req.AccountID)
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

// AttachDisk ...
func (svc *DiskSvc) AttachDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeDiskAttachOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.TCloud(cts.Kit, req.AccountID)
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

	syncClient := synctcloud.NewClient(svc.DataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    opt.Region,
		CloudIDs:  opt.CloudDiskIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &synctcloud.SyncDiskOption{})
	if err != nil {
		logs.Errorf("sync tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudCvmID}
	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DetachDisk ...
func (svc *DiskSvc) DetachDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeDiskDetachOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DetachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.DataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    opt.Region,
		CloudIDs:  opt.CloudDiskIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &synctcloud.SyncDiskOption{})
	if err != nil {
		logs.Errorf("sync tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudCvmID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &synctcloud.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *DiskSvc) makeDiskAttachOption(
	kt *kit.Kit,
	req *proto.TCloudDiskAttachReq,
) (*disk.TCloudDiskAttachOption, error) {
	dataCli := svc.DataCli.TCloud

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.TCloudDiskAttachOption{
		Region:       diskData.Region,
		CloudCvmID:   cvmData.CloudID,
		CloudDiskIDs: []string{diskData.CloudID},
	}, nil
}

func (svc *DiskSvc) makeDiskDetachOption(
	kt *kit.Kit,
	req *proto.DiskDetachReq,
) (*disk.TCloudDiskDetachOption, error) {
	dataCli := svc.DataCli.TCloud

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.TCloudDiskDetachOption{
		Region:       diskData.Region,
		CloudCvmID:   cvmData.CloudID,
		CloudDiskIDs: []string{diskData.CloudID},
	}, nil
}

func (svc *DiskSvc) makeDiskDeleteOption(
	kt *kit.Kit,
	req *proto.DiskDeleteReq,
) (*disk.TCloudDiskDeleteOption, error) {
	diskData, err := svc.DataCli.TCloud.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}
	return &disk.TCloudDiskDeleteOption{Region: diskData.Region, CloudIDs: []string{diskData.CloudID}}, nil
}
