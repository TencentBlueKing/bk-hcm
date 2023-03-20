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

package azure

import (
	"hcm/cmd/hc-service/logics/sync/cvm"
	syncdisk "hcm/cmd/hc-service/logics/sync/disk"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/disk/datasvc"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	proto "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// DiskSvc ...
type DiskSvc struct {
	Adaptor *cloudclient.CloudAdaptorClient
	DataCli *dataservice.Client
}

// CreateDisk ...
func (svc *DiskSvc) CreateDisk(cts *rest.Contexts) (interface{}, error) {
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
	}
	cloudIDs, err := client.CreateDisk(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	_, err = syncdisk.SyncAzureDisk(
		cts.Kit,
		&syncdisk.SyncAzureDiskOption{
			AccountID:         req.AccountID,
			ResourceGroupName: req.Extension.ResourceGroupName,
			CloudIDs:          cloudIDs,
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		return nil, err
	}

	resp, err := svc.DataCli.Global.ListDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.DiskListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: cloudIDs,
				}, &filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: string(enumor.Azure),
				},
			},
		}, Page: &core.BasePage{Limit: uint(len(cloudIDs))}, Fields: []string{"id"}},
	)

	diskIDs := make([]string, len(cloudIDs))
	for idx, diskData := range resp.Details {
		diskIDs[idx] = diskData.ID
	}
	return &core.BatchCreateResult{IDs: diskIDs}, nil
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

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
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
	req := new(proto.AzureDiskAttachReq)
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

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
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
	_, err = syncdisk.SyncAzureDisk(
		cts.Kit,
		&syncdisk.SyncAzureDiskOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{diskData.CloudID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncAzureDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}
	return cvm.SyncAzureCvm(
		cts.Kit,
		svc.Adaptor,
		svc.DataCli,
		&cvm.SyncAzureCvmOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{cvmData.CloudID},
		},
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

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
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

	diskData, err := svc.DataCli.Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}
	_, err = syncdisk.SyncAzureDisk(
		cts.Kit,
		&syncdisk.SyncAzureDiskOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{diskData.CloudID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncAzureDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}
	return cvm.SyncAzureCvm(
		cts.Kit,
		svc.Adaptor,
		svc.DataCli,
		&cvm.SyncAzureCvmOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{cvmData.CloudID},
		},
	)
}

func (svc *DiskSvc) makeDiskAttachOption(
	kt *kit.Kit,
	req *proto.AzureDiskAttachReq,
) (*disk.AzureDiskAttachOption, error) {
	dataCli := svc.DataCli.Azure

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.AzureDiskAttachOption{
		ResourceGroupName: diskData.Extension.ResourceGroupName,
		CvmName:           cvmData.Name,
		DiskName:          diskData.Name,
		CachingType:       req.CachingType,
	}, nil
}

func (svc *DiskSvc) makeDiskDetachOption(
	kt *kit.Kit,
	req *proto.DiskDetachReq,
) (*disk.AzureDiskDetachOption, error) {
	dataCli := svc.DataCli.Azure

	diskData, err := dataCli.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &disk.AzureDiskDetachOption{
		ResourceGroupName: diskData.Extension.ResourceGroupName,
		CvmName:           cvmData.Name,
		DiskName:          diskData.Name,
	}, nil
}

func (svc *DiskSvc) makeDiskDeleteOption(
	kt *kit.Kit,
	req *proto.DiskDeleteReq,
) (*disk.AzureDiskDeleteOption, error) {
	diskData, err := svc.DataCli.Azure.RetrieveDisk(kt.Ctx, kt.Header(), req.DiskID)
	if err != nil {
		return nil, err
	}

	return &disk.AzureDiskDeleteOption{
		ResourceGroupName: diskData.Extension.ResourceGroupName,
		DiskName:          diskData.Name,
	}, nil
}
