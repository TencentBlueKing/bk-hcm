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

package aws

import (
	syncaws "hcm/cmd/hc-service/logics/res-sync/aws"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/eip/datasvc"
	"hcm/pkg/adaptor/types/eip"
	apicore "hcm/pkg/api/core"
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

	client, err := svc.Adaptor.Aws(cts.Kit, req.AccountID)
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
	req := new(proto.AwsEipAssociateReq)
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

	opt, err := svc.makeEipAssociateOption(cts.Kit, req)
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

	eipData, err := svc.DataCli.Aws.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{eipData.CloudID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncaws.SyncEipOption{})
	if err != nil {
		logs.Errorf("sync aws eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudCvmID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncaws.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync aws cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateEip ...
func (svc *EipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsEipDisassociateReq)
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

	client, err := svc.Adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DisassociateEip(cts.Kit, opt)
	if err != nil {
		logs.Errorf("aws cloud disassociate eip failed, req: %+v, opt: %+v, err: %+v", req, opt, err)
		return nil, err
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	err = manager.Delete(cts.Kit)

	eipData, err := svc.DataCli.Aws.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{eipData.CloudID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncaws.SyncEipOption{})
	if err != nil {
		logs.Errorf("sync aws eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Aws.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncaws.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync aws cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateEip ...
func (svc *EipSvc) CreateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsEipCreateReq)
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

	opt, err := svc.makeEipCreateOption(req)
	if err != nil {
		return nil, err
	}

	eipPtr, err := client.CreateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	cloudIDs := []string{*eipPtr}

	syncClient := syncaws.NewClient(svc.DataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    opt.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncaws.SyncEipOption{
		BkBizID: req.BkBizID,
	})
	if err != nil {
		logs.Errorf("sync aws eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
					Value: string(enumor.Aws),
				},
			},
		}, Page: &apicore.BasePage{Limit: uint(len(cloudIDs))}, Fields: []string{"id"}},
	)

	eipIDs := make([]string, len(cloudIDs))
	for idx, eipData := range resp.Details {
		eipIDs[idx] = eipData.ID
	}

	return &apicore.BatchCreateResult{IDs: eipIDs}, nil
}

func (svc *EipSvc) makeEipDeleteOption(
	kt *kit.Kit,
	req *proto.EipDeleteReq,
) (*eip.AwsEipDeleteOption, error) {
	eipData, err := svc.DataCli.Aws.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}
	return &eip.AwsEipDeleteOption{
		Region:  eipData.Region,
		CloudID: eipData.CloudID,
	}, nil
}

func (svc *EipSvc) makeEipAssociateOption(
	kt *kit.Kit,
	req *proto.AwsEipAssociateReq,
) (*eip.AwsEipAssociateOption, error) {
	dataCli := svc.DataCli.Aws

	eipData, err := dataCli.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	cvmData, err := dataCli.Cvm.GetCvm(kt.Ctx, kt.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	return &eip.AwsEipAssociateOption{
		Region:     eipData.Region,
		PublicIp:   eipData.PublicIp,
		CloudCvmID: cvmData.CloudID,
	}, nil
}

func (svc *EipSvc) makeEipDisassociateOption(
	kt *kit.Kit,
	req *proto.AwsEipDisassociateReq,
) (*eip.AwsEipDisassociateOption, error) {
	eipData, err := svc.DataCli.Aws.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	return &eip.AwsEipDisassociateOption{Region: eipData.Region, PublicIp: eipData.PublicIp}, nil
}

func (svc *EipSvc) makeEipCreateOption(req *proto.AwsEipCreateReq) (*eip.AwsEipCreateOption, error) {
	return &eip.AwsEipCreateOption{
		Region:             req.Region,
		PublicIpv4Pool:     req.PublicIpv4Pool,
		NetworkBorderGroup: req.NetworkBorderGroup,
	}, nil
}
