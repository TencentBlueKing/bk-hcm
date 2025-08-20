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
	"hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/hc-service/sync"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncCvmCCInfo 同步cvm的cc相关信息，这里的同步和云上资源同步不同，不完全适配ResourceSync的机制（非云资源）
func (svc *service) SyncCvmCCInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.AzureSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	syncCli, err := svc.syncCli.Azure(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("init azure sync client failed for account %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}
	hd := &cvmCCInfoSyncHandler{
		accountID:         req.AccountID,
		resourceGroupName: req.ResourceGroupName,
		dbCli:             svc.dataCli,
		syncCli:           syncCli,
	}

	return nil, hd.Sync(cts.Kit)
}

// SyncCvmCCInfoByCond sync cvm cc info by condition.
func (svc *service) SyncCvmCCInfoByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.SyncCvmByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("bk_host_id", req.HostIDs),
			tools.RuleNotEqual("bk_biz_id", constant.UnassignedBiz),
		),
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	resp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list azure cvm failed, err: %v, req: %v, rid: %s", err, converter.PtrToVal(listReq), cts.Kit.Rid)
		return nil, err
	}

	syncCli, err := svc.syncCli.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	op := &cvmCCInfoSyncHandler{syncCli: syncCli}

	if err = op.syncCCInfo(cts.Kit, resp.Details); err != nil {
		logs.Errorf("sync cc info failed, err: %v, cvm: %v, rid: %s", err, resp.Details, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// cvmCCInfoSyncHandler ...
type cvmCCInfoSyncHandler struct {
	accountID         string
	resourceGroupName string
	offset            uint
	dbCli             *dataservice.Client
	syncCli           azure.Interface
}

// Sync 同步cvm的cc相关信息
func (h *cvmCCInfoSyncHandler) Sync(kt *kit.Kit) error {
	for {
		cvms, err := h.nextCvms(kt)
		if err != nil {
			logs.Errorf("list cvm for sync cc info failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		if len(cvms) == 0 {
			break
		}
		if err := h.syncCCInfo(kt, cvms); err != nil {
			logs.Errorf("sync cc info failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func (h *cvmCCInfoSyncHandler) nextCvms(kt *kit.Kit) ([]cvm.BaseCvm, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", h.accountID),
			tools.RuleJSONEqual("extension.resource_group_name", h.resourceGroupName),
			tools.RuleNotEqual("bk_biz_id", constant.UnassignedBiz),
		),
		Page: &core.BasePage{
			Start: uint32(h.offset),
			Limit: core.DefaultMaxPageLimit,
		},
	}
	resp, err := h.dbCli.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list azure cvm failed, err: %v, account id: %s, resource_group_name: %s, rid: %s", err,
			h.accountID, h.resourceGroupName, kt.Rid)
		return nil, err
	}

	h.offset += uint(len(resp.Details))

	return resp.Details, nil
}

func (h *cvmCCInfoSyncHandler) syncCCInfo(kt *kit.Kit, cvms []cvm.BaseCvm) error {
	params := &azure.SyncCvmCCInfoParams{
		Cvms: cvms,
	}
	return h.syncCli.CvmCCInfo(kt, params)
}
