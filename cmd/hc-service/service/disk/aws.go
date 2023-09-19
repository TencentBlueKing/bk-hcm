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
	syncaws "hcm/cmd/hc-service/logics/res-sync/aws"
	"hcm/cmd/hc-service/service/disk/datasvc"
	"hcm/pkg/adaptor/types/disk"
	proto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateAwsDisk ...
func (svc *service) CreateAwsDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int64(req.DiskSize)
	opt := &disk.AwsDiskCreateOption{
		Region:    req.Region,
		Zone:      req.Zone,
		DiskType:  &req.DiskType,
		DiskSize:  diskSize,
		DiskCount: converter.ValToPtr(uint64(req.DiskCount)),
	}
	result, err := client.CreateDisk(cts.Kit, opt)
	if err != nil {
		logs.Errorf("create aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  result.SuccessCloudIDs,
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncaws.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// DeleteAwsDisk ...
func (svc *service) DeleteAwsDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskData, err := svc.DataCli.Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AwsDiskDeleteOption{
		Region:  diskData.Region,
		CloudID: diskData.CloudID,
	}

	client, err := svc.Adaptor.Aws(cts.Kit, diskData.AccountID)
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

// AttachAwsDisk ...
func (svc *service) AttachAwsDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取磁盘数据和cvm 数据，构造挂载磁盘请求
	diskData, err := svc.DataCli.Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := svc.DataCli.Aws.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AwsDiskAttachOption{
		Region:      diskData.Region,
		CloudCvmID:  cvmData.CloudID,
		CloudDiskID: diskData.CloudID,
		DeviceName:  req.DeviceName,
	}

	client, err := svc.Adaptor.Aws(cts.Kit, diskData.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.AttachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: diskData.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{opt.CloudDiskID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncaws.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudCvmID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncaws.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync aws cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DetachAwsDisk ...
func (svc *service) DetachAwsDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskInfo, err := svc.DataCli.Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmInfo, err := svc.DataCli.Aws.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	opt := &disk.AwsDiskDetachOption{
		Region:      diskInfo.Region,
		CloudCvmID:  cvmInfo.CloudID,
		CloudDiskID: diskInfo.CloudID,
	}
	client, err := svc.Adaptor.Aws(cts.Kit, diskInfo.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DetachDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.DiskCvmRelManager{CvmID: req.CvmID, DiskID: req.DiskID, DataCli: svc.DataCli}
	err = manager.Delete(cts.Kit)
	if err != nil {
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: diskInfo.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{opt.CloudDiskID},
	}

	_, err = syncClient.Disk(cts.Kit, params, &syncaws.SyncDiskOption{BootMap: nil})
	if err != nil {
		logs.Errorf("sync aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudCvmID}
	_, err = syncClient.Cvm(cts.Kit, params, &syncaws.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
