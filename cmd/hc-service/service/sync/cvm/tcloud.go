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
	"hcm/pkg/adaptor/tcloud"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
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

	cloudCvmTotalMap := make(map[string]struct{}, 0)
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
			cloudCvmTotalMap[*one.InstanceId] = struct{}{}
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

	if err = svc.removeDBNotExistCvm(cts.Kit, cli, req, cloudCvmTotalMap); err != nil {
		logs.Errorf("remove db not exist cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncCvmSvc) removeDBNotExistCvm(kt *kit.Kit, tcloud *tcloud.TCloud, req *sync.TCloudSyncReq,
	cloudCvmMap map[string]struct{}) error {

	// 找出云上已经不存在的主机ID
	start := uint32(0)
	delCloudCvmIDs := make([]string, 0)
	for {
		listReq := &dataproto.CvmListReq{
			Field: []string{"cloud_id"},
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{
				Start: start,
				Limit: core.DefaultMaxPageLimit,
			},
		}
		result, err := svc.dataCli.Global.Cvm.ListCvm(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list cvm from db failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		for _, one := range result.Details {
			if _, exist := cloudCvmMap[one.CloudID]; !exist {
				delCloudCvmIDs = append(delCloudCvmIDs, one.CloudID)
			}
		}

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	// 再用这部分主机ID去云上确认是否存在，如果不存在，删除db数据，存在的忽略，等同步修复
	start, end := 0, typecore.TCloudQueryLimit
	for {
		if int(start)+end > len(delCloudCvmIDs) {
			end = len(delCloudCvmIDs)
		} else {
			end = int(start) + typecore.TCloudQueryLimit
		}
		tmpCloudIDs := delCloudCvmIDs[start:end]

		if len(tmpCloudIDs) == 0 {
			break
		}

		listOpt := &typecvm.TCloudListOption{
			Region:   req.Region,
			CloudIDs: tmpCloudIDs,
		}
		cvms, err := tcloud.ListCvm(kt, listOpt)
		if err != nil {
			logs.Errorf("list cvm from tcloud failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
			return err
		}

		tmpMap := converter.StringSliceToMap(tmpCloudIDs)
		for _, instance := range cvms {
			delete(tmpMap, *instance.InstanceId)
		}

		if len(tmpMap) == 0 {
			start = uint32(end)
			continue
		}

		if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), &dataproto.CvmBatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", converter.MapKeyToStringSlice(tmpMap)),
		}); err != nil {
			logs.Errorf("batch delete db cvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		start = uint32(end)
		if start == uint32(len(delCloudCvmIDs)) {
			break
		}
	}

	return nil
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
	cloudCvmTotalMap := make(map[string]struct{}, 0)
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
			cloudCvmTotalMap[*one.InstanceId] = struct{}{}
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

	if err = svc.removeDBNotExistCvm(cts.Kit, cli, req, cloudCvmTotalMap); err != nil {
		logs.Errorf("remove db not exist cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
