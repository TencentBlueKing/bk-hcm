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

package huawei

import (
	"hcm/cmd/hc-service/logics/sync/cvm"
	syncdisk "hcm/cmd/hc-service/logics/sync/disk"
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

// CreateDisk ...
func (svc *DiskSvc) CreateDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.HuaWeiDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.HuaWei(cts.Kit, req.Base.AccountID)
	if err != nil {
		return nil, err
	}

	diskCount := int32(req.Base.DiskCount)
	opt := &disk.HuaWeiDiskCreateOption{
		Region:         req.Base.Region,
		Zone:           req.Base.Zone,
		DiskType:       req.Base.DiskType,
		DiskSize:       int32(req.Base.DiskSize),
		DiskCount:      &diskCount,
		DiskChargeType: &req.Extension.DiskChargeType,
	}

	if prepaid := req.Extension.DiskChargePrepaid; prepaid != nil {
		opt.DiskChargePrepaid = &disk.HuaWeiDiskChargePrepaid{
			PeriodNum:   prepaid.PeriodNum,
			PeriodType:  prepaid.PeriodType,
			IsAutoRenew: prepaid.IsAutoRenew,
		}
	}

	client.CreateDisk(cts.Kit, opt)

	// TODO save to data-service

	return nil, nil
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

	client, err := svc.Adaptor.HuaWei(cts.Kit, req.AccountID)
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
	req := new(proto.HuaWeiDiskAttachReq)
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

	client, err := svc.Adaptor.HuaWei(cts.Kit, req.AccountID)
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

	_, err = syncdisk.SyncHuaWeiDisk(
		cts.Kit,
		&syncdisk.SyncHuaWeiDiskOption{
			AccountID: req.AccountID,
			Region:    opt.Region,
			CloudIDs:  []string{opt.CloudDiskID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncHuaWeiDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return cvm.SyncHuaWeiCvm(
		cts.Kit,
		&cvm.SyncHuaWeiCvmOption{AccountID: req.AccountID, Region: opt.Region, CloudIDs: []string{opt.CloudCvmID}},
		svc.Adaptor,
		svc.DataCli,
	)
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

	client, err := svc.Adaptor.HuaWei(cts.Kit, req.AccountID)
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

	_, err = syncdisk.SyncHuaWeiDisk(
		cts.Kit,
		&syncdisk.SyncHuaWeiDiskOption{
			AccountID: req.AccountID,
			Region:    opt.Region,
			CloudIDs:  []string{opt.CloudDiskID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncHuaWeiDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return cvm.SyncHuaWeiCvm(
		cts.Kit,
		&cvm.SyncHuaWeiCvmOption{AccountID: req.AccountID, Region: opt.Region, CloudIDs: []string{opt.CloudCvmID}},
		svc.Adaptor,
		svc.DataCli,
	)
}

func (svc *DiskSvc) makeDiskAttachOption(
	kt *kit.Kit,
	req *proto.HuaWeiDiskAttachReq,
) (*disk.HuaWeiDiskAttachOption, error) {
	dataCli := svc.DataCli.HuaWei

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.HuaWeiDiskAttachOption{
		Region:      diskData.Region,
		CloudCvmID:  cvmData.CloudID,
		CloudDiskID: diskData.CloudID,
		DeviceName:  req.DeviceName,
	}, nil
}

func (svc *DiskSvc) makeDiskDetachOption(
	kt *kit.Kit,
	req *proto.DiskDetachReq,
) (*disk.HuaWeiDiskDetachOption, error) {
	dataCli := svc.DataCli.HuaWei

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.HuaWeiDiskDetachOption{
		Region:      diskData.Region,
		CloudCvmID:  cvmData.CloudID,
		CloudDiskID: diskData.CloudID,
	}, nil
}

func (svc *DiskSvc) makeDiskDeleteOption(
	kt *kit.Kit,
	req *proto.DiskDeleteReq,
) (*disk.HuaWeiDiskDeleteOption, error) {
	diskData, err := svc.DataCli.HuaWei.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	return &disk.HuaWeiDiskDeleteOption{Region: diskData.Region, CloudID: diskData.CloudID}, nil
}
