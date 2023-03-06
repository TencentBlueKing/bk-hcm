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
	"hcm/cmd/hc-service/logics/sync/cvm"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncTCloudCvm ...
func (svc *syncCvmSvc) SyncTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := svc.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typecvm.TCloudListOption{
		Region:   req.Region,
		CloudIDs: nil,
		Page: &typecore.TCloudPage{
			Offset: 0,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	for {
		cvms, err := cli.ListCvm(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list tcloud cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(cvms) == 0 {
			break
		}

		cloudIDs := make([]string, 0, len(cvms))
		for _, one := range cvms {
			cloudIDs = append(cloudIDs, *one.InstanceId)
		}

		syncOpt := &cvm.SyncTCloudCvmOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  cloudIDs,
		}
		if err = cvm.SyncTCloudCvm(cts.Kit, svc.adaptor, svc.dataCli, syncOpt); err != nil {
			logs.Errorf("request to sync tcloud cvm failed, err: %v, opt: %v, rid: %s", err, syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(cvms) < typecore.TCloudQueryLimit {
			break
		}

		listOpt.Page.Offset += typecore.TCloudQueryLimit
	}

	return nil, nil
}

// OperateSyncTCloudCvm ...
func (svc *syncCvmSvc) OperateSyncTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := svc.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typecvm.TCloudListOption{
		Region:   req.Region,
		CloudIDs: nil,
		Page: &typecore.TCloudPage{
			Offset: 0,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	for {
		cvms, err := cli.ListCvm(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list tcloud cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(cvms) == 0 {
			break
		}

		cloudIDs := make([]string, 0, len(cvms))
		for _, one := range cvms {
			cloudIDs = append(cloudIDs, *one.InstanceId)
		}

		syncOpt := &cvm.SyncTCloudCvmOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  cloudIDs,
		}
		_, err = cvm.SyncTCloudCvmWithRelResource(cts.Kit, svc.adaptor, svc.dataCli, syncOpt)
		if err != nil {
			logs.Errorf("request to sync tcloud cvm all rel failed, err: %v, opt: %v, rid: %s", err, syncOpt,
				cts.Kit.Rid)
			return nil, err
		}

		if len(cvms) < typecore.TCloudQueryLimit {
			break
		}

		listOpt.Page.Offset += typecore.TCloudQueryLimit
	}

	return nil, nil
}
