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

package cvm

import (
	cvm "hcm/cmd/hc-service/logics/sync/cvm"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncAzureCvm ...
func (svc *syncCvmSvc) SyncAzureCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.AzureSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := svc.adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typecvm.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	cvms, err := cli.ListCvm(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list azure cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
		return nil, err
	}

	if len(cvms) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0)
	for _, one := range cvms {
		cloudIDs = append(cloudIDs, *one.ID)
	}

	start := 0
	end := 0
	syncOpt := &cvm.SyncAzureCvmOption{
		AccountID:         req.AccountID,
		Region:            req.Region,
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          cloudIDs,
	}
	for {
		if start+constant.BatchOperationMaxLimit > len(cloudIDs) {
			end = len(cloudIDs)
		} else {
			end = start + constant.BatchOperationMaxLimit
		}

		syncOpt.CloudIDs = cloudIDs[start:end]
		if _, err = cvm.SyncAzureCvm(cts.Kit, svc.adaptor, svc.dataCli, syncOpt); err != nil {
			logs.Errorf("request to sync azure cvm failed, err: %v, opt: %v, rid: %s", err, syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if end == len(cloudIDs) {
			break
		}
	}

	return nil, nil
}

// SyncAzureCvmWithRelResource ...
func (svc *syncCvmSvc) SyncAzureCvmWithRelResource(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.AzureSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := svc.adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typecvm.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	cvms, err := cli.ListCvm(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list azure cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
		return nil, err
	}

	if len(cvms) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0)
	for _, one := range cvms {
		cloudIDs = append(cloudIDs, *one.ID)
	}

	start := 0
	end := 0
	syncOpt := &cvm.SyncAzureCvmOption{
		AccountID:         req.AccountID,
		Region:            req.Region,
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          cloudIDs,
	}
	for {
		if start+constant.BatchOperationMaxLimit > len(cloudIDs) {
			end = len(cloudIDs)
		} else {
			end = start + constant.BatchOperationMaxLimit
		}

		syncOpt.CloudIDs = cloudIDs[start:end]
		if _, err = cvm.SyncAzureCvmWithRelResource(cts.Kit, svc.adaptor, svc.dataCli, syncOpt); err != nil {
			logs.Errorf("request to sync azure cvm with relation resource failed, err: %v, opt: %v, rid: %s",
				err, syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if end == len(cloudIDs) {
			break
		}
	}

	return nil, nil
}
