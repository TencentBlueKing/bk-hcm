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

package gcp

import (
	syncgcp "hcm/cmd/hc-service/logics/res-sync/gcp"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/eip/datasvc"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	proto "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// EipSvc ...
type EipSvc struct {
	Adaptor *cloudclient.CloudAdaptorClient
	DataCli *dataservice.Client
}

// DeleteEip ...
func (svc *EipSvc) DeleteEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.EipDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeEipDeleteOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DeleteEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.EipManager{DataCli: svc.DataCli}
	return nil, manager.Delete(cts.Kit, []string{req.EipID})
}

// AssociateEip ...
func (svc *EipSvc) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpEipAssociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeEipAssociateOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.AssociateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	err = manager.Create(cts.Kit)
	if err != nil {
		return nil, err
	}

	eipData, err := svc.DataCli.Gcp.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: req.AccountID,
		CloudIDs:  []string{eipData.CloudID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncgcp.SyncEipOption{Region: opt.Region})
	if err != nil {
		logs.Errorf("sync gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncgcp.SyncCvmWithRelResOption{Region: eipData.Region,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm with rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, err = syncClient.NetworkInterface(cts.Kit, params, &syncgcp.SyncNIOption{Region: eipData.Region,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp ni failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateEip ...
func (svc *EipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpEipDisassociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeEipDisassociateOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	if len(opt.AccessConfigName) > 0 {
		err = client.DisassociateEip(cts.Kit, opt)
		if err != nil {
			logs.Errorf("gcp cloud disassociate eip failed, req: %+v, opt: %+v, err: %+v", req, opt, err)
			return nil, err
		}
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	err = manager.Delete(cts.Kit)
	if err != nil {
		return nil, err
	}

	eipData, err := svc.DataCli.Gcp.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: req.AccountID,
		CloudIDs:  []string{eipData.CloudID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncgcp.SyncEipOption{Region: opt.Region})
	if err != nil {
		logs.Errorf("sync gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncgcp.SyncCvmWithRelResOption{Region: eipData.Region,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp cvm with rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, err = syncClient.NetworkInterface(cts.Kit, params, &syncgcp.SyncNIOption{Region: eipData.Region,
		Zone: opt.Zone})
	if err != nil {
		logs.Errorf("sync gcp ni failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateEip ...
func (svc *EipSvc) CreateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpEipCreateReq)
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

	opt, err := svc.makeEipCreateOption(req)
	if err != nil {
		return nil, err
	}

	result, err := client.CreateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if len(result.UnknownCloudIDs) > 0 {
		logs.Errorf("eip(%v) is unknown, rid: %s", result.UnknownCloudIDs, cts.Kit.Rid)
	}

	cloudIDs := result.SuccessCloudIDs

	syncClient := syncgcp.NewClient(svc.DataCli, client)

	params := &syncgcp.SyncBaseParams{
		AccountID: req.AccountID,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncgcp.SyncEipOption{Region: opt.Region, BkBizID: req.BkBizID})
	if err != nil {
		logs.Errorf("sync gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	resp, err := svc.DataCli.Global.ListEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: cloudIDs,
				}, &filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: string(enumor.Gcp),
				},
			},
		}, Page: &core.BasePage{Limit: uint(len(cloudIDs))}, Fields: []string{"id"}},
	)

	eipIDs := make([]string, len(cloudIDs))
	for idx, eipData := range resp.Details {
		eipIDs[idx] = eipData.ID
	}

	return &core.BatchCreateResult{IDs: eipIDs}, nil
}

func (svc *EipSvc) makeEipDeleteOption(
	kt *kit.Kit,
	req *proto.EipDeleteReq,
) (*eip.GcpEipDeleteOption, error) {
	eipData, err := svc.DataCli.Gcp.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}
	return &eip.GcpEipDeleteOption{Region: eipData.Region, EipName: *eipData.Name}, nil
}

func (svc *EipSvc) makeEipAssociateOption(
	kt *kit.Kit,
	req *proto.GcpEipAssociateReq,
) (*eip.GcpEipAssociateOption, error) {
	eipData, err := svc.DataCli.Gcp.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	networkInterface, err := svc.DataCli.Gcp.NetworkInterface.Get(kt.Ctx, kt.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}

	return &eip.GcpEipAssociateOption{
		Zone:                 cvmData.Zone,
		CvmName:              cvmData.Name,
		NetworkInterfaceName: networkInterface.Name,
		PublicIp:             eipData.PublicIp,
		CloudID:              eipData.CloudID,
		Region:               eipData.Region,
	}, nil
}

func (svc *EipSvc) makeEipDisassociateOption(
	kt *kit.Kit,
	req *proto.GcpEipDisassociateReq,
) (*eip.GcpEipDisassociateOption, error) {
	dataCli := svc.DataCli.Gcp

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	networkInterface, err := dataCli.NetworkInterface.Get(kt.Ctx, kt.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}

	eipData, err := dataCli.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	var accessConfigName string
	for _, config := range networkInterface.Extension.AccessConfigs {
		if config.NatIP == eipData.PublicIp {
			accessConfigName = config.Name
			break
		}
	}

	return &eip.GcpEipDisassociateOption{
		Zone:                 cvmData.Zone,
		CvmName:              cvmData.Name,
		NetworkInterfaceName: networkInterface.Name,
		AccessConfigName:     accessConfigName,
		CloudID:              eipData.CloudID,
		Region:               eipData.Region,
	}, nil
}

func (svc *EipSvc) makeEipCreateOption(req *proto.GcpEipCreateReq) (*eip.GcpEipCreateOption, error) {
	return &eip.GcpEipCreateOption{
		EipName:     req.EipName,
		Region:      req.Region,
		NetworkTier: req.NetworkTier,
		IpVersion:   req.IpVersion,
	}, nil
}
