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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncAwsCvm ...
func (svc *syncCvmSvc) SyncAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.SyncAwsCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &cvm.SyncAwsCvmOption{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	if _, err := cvm.SyncAwsCvm(cts.Kit, svc.adaptor, svc.dataCli, opt); err != nil {
		logs.Errorf("request to sync aws cvm failed, err: %v, opt: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncAwsCvmWithRelResource ...
func (svc *syncCvmSvc) SyncAwsCvmWithRelResource(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.AwsSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := svc.adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typecvm.AwsListOption{
		Region:   req.Region,
		CloudIDs: nil,
		Page: &typecore.AwsPage{
			MaxResults: converter.ValToPtr(int64(constant.BatchOperationMaxLimit)),
			NextToken:  nil,
		},
	}
	for {
		resp, err := cli.ListCvm(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list aws cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resp.Reservations {
			for _, instance := range one.Instances {
				cloudIDs = append(cloudIDs, *instance.InstanceId)
			}
		}

		if len(cloudIDs) == 0 {
			break
		}

		syncOpt := &cvm.SyncAwsCvmOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err = cvm.SyncAwsCvmWithRelResource(cts.Kit, syncOpt, svc.adaptor, svc.dataCli); err != nil {
			logs.Errorf("request to sync aws cvm with relation resource failed, err: %v, opt: %v, rid: %s", err,
				syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if resp.NextToken == nil {
			break
		}

		listOpt.Page.NextToken = resp.NextToken
	}

	return nil, nil
}
