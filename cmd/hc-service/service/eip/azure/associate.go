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
	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/cmd/hc-service/service/eip/datasvc"
	proto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AssociateEip ...
func (svc *EipSvc) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureEipAssociateReq)
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

	opt, err := svc.makeEipAssociateOption(cts.Kit, req, client)
	if err != nil {
		return nil, err
	}

	err = client.AssociateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.DataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         req.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{opt.CloudEipID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncazure.SyncEipOption{})
	if err != nil {
		logs.Errorf("sync azure eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncazure.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync azure cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	networkInterface, err := svc.DataCli.Azure.NetworkInterface.Get(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		req.NetworkInterfaceID,
	)
	if err != nil {
		return nil, err
	}

	params.CloudIDs = []string{networkInterface.CloudID}
	_, err = syncClient.NetworkInterface(cts.Kit, params, &syncazure.SyncNIOption{})
	if err != nil {
		logs.Errorf("sync azure ni failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateEip ...
func (svc *EipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureEipDisassociateReq)
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

	opt, err := svc.makeEipDisassociateOption(cts.Kit, req, client)
	if err != nil {
		return nil, err
	}

	if opt.NetworkInterface != nil {
		err = client.DisassociateEip(cts.Kit, opt)
		if err != nil {
			logs.Errorf("azure cloud disassociate eip failed, rgName: %s, cloudEipID: %s, err: %+v",
				opt.ResourceGroupName, opt.CloudEipID, err)
			return nil, err
		}
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	if err = manager.Delete(cts.Kit); err != nil {
		logs.Errorf("delete azure eip cvm rel db failed, eipID: %s, cvmID: %s, err: %+v",
			req.EipID, req.CvmID, err)
		return nil, err
	}

	syncClient := syncazure.NewClient(svc.DataCli, client)

	params := &syncazure.SyncBaseParams{
		AccountID:         req.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{opt.CloudEipID},
	}

	_, err = syncClient.Eip(cts.Kit, params, &syncazure.SyncEipOption{})
	if err != nil {
		logs.Errorf("sync azure eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		logs.Errorf("azure disassociate eip get cvm failed, cvmID: %s, err: %+v", req.CvmID, err)
		return nil, err
	}

	params.CloudIDs = []string{cvmData.CloudID}
	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncazure.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync azure cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	params.CloudIDs = []string{opt.CloudNetworkInterfaceID}
	_, err = syncClient.NetworkInterface(cts.Kit, params, &syncazure.SyncNIOption{})
	if err != nil {
		logs.Errorf("sync azure ni failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
