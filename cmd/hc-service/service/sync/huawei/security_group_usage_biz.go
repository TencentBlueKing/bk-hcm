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

package huawei

import (
	"hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/sync"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncSecurityGroupUsageBiz 同步安全组使用业务id列表，这里的同步和云上资源同步不同，不完全适配ResourceSync的机制（非云资源）
func (svc *service) SyncSecurityGroupUsageBiz(cts *rest.Contexts) (interface{}, error) {

	req := new(sync.HuaWeiSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	syncCli, err := svc.syncCli.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("init huawei sync client failed for account %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}
	hd := &SGUsageBizRelSyncHandler{
		accountID: req.AccountID,
		region:    req.Region,
		offset:    0,
		dbCli:     svc.dataCli,
		syncCli:   syncCli,
	}

	return nil, hd.Sync(cts.Kit)
}

// SGUsageBizRelSyncHandler ...
type SGUsageBizRelSyncHandler struct {
	accountID string
	region    string
	offset    uint
	dbCli     *dataservice.Client
	syncCli   huawei.Interface
}

// Sync 同步安全组使用业务id列表
func (h *SGUsageBizRelSyncHandler) Sync(kt *kit.Kit) error {
	for {

		sgList, err := h.nextSGID(kt)
		if err != nil {
			logs.Errorf("list security group for sync usage bizs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		if len(sgList) == 0 {
			break
		}
		if err := h.syncUsageBizs(kt, sgList); err != nil {
			logs.Errorf("sync usage bizs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}
func (h *SGUsageBizRelSyncHandler) nextSGID(kt *kit.Kit) ([]corecloud.BaseSecurityGroup, error) {

	listReq := &cloud.SecurityGroupListReq{
		Field: []string{"id", "cloud_id", "region", "vendor",
			"usage_biz_ids", "mgmt_biz_id", "mgmt_type", "bk_biz_id", "manager", "bak_manager"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", h.accountID),
			tools.RuleEqual("region", h.region),
		),
		Page: &core.BasePage{
			Start: uint32(h.offset),
			Limit: core.DefaultMaxPageLimit,
		},
	}
	sgResp, err := h.dbCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list sg of Huawei %s at %s failed, err: %v, rid: %s", h.accountID, h.region, err, kt.Rid)
		return nil, err
	}

	h.offset += uint(len(sgResp.Details))

	return sgResp.Details, nil
}

func (h *SGUsageBizRelSyncHandler) syncUsageBizs(kt *kit.Kit, sgList []corecloud.BaseSecurityGroup) error {
	params := &huawei.SyncSGUsageBizParams{
		AccountID: h.accountID,
		Region:    h.region,
		SGList:    sgList,
	}
	return h.syncCli.SecurityGroupUsageBiz(kt, params)
}
