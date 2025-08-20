/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/pkg/adaptor/types/cvm"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// 主要存放腾讯云主机操作的实现，如 开关机、重装、重启

// BatchResetTCloudCvm 重装系统
func (svc *cvmSvc) BatchResetTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchResetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, cloudID := range req.CloudIDs {
		opt := &cvm.ResetInstanceOption{
			Region:   req.Region,
			CloudID:  cloudID,
			ImageID:  req.ImageID,
			Password: req.Password,
		}
		if _, err = client.ResetCvmInstance(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to tcloud reset cvm instance failed, err: %v, opt: %+v, cloudID: %s, "+
				"rid: %s", err, cvt.PtrToVal(req), cloudID, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

// BatchStartTCloudCvm ...
func (svc *cvmSvc) BatchStartTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStartReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStopTCloudCvm ...
func (svc *cvmSvc) BatchStopTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStopOption{
		Region:      req.Region,
		CloudIDs:    cloudIDs,
		StopType:    req.StopType,
		StoppedMode: req.StoppedMode,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRebootTCloudCvm ...
func (svc *cvmSvc) BatchRebootTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchRebootReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		StopType: req.StopType,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
