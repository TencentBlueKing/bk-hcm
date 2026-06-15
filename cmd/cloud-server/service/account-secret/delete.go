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

package accountsecret

import (
	"fmt"

	proto "hcm/pkg/api/cloud-server/account-secret"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/maps"
)

// BatchDeleteBizAccountSecret batch deletes biz account secrets.
func (s *service) BatchDeleteBizAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountSecretBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Update}, BizID: bizID}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	// Validate and collect account ids that need Extension clearing
	accountIDsToClear, err := s.validAndCollectClearExtAccountIDs(cts.Kit, vendor, bizID, req.IDs)
	if err != nil {
		return nil, err
	}

	// 记录删除审计
	if err = s.audit.ResDeleteAudit(cts.Kit, enumor.AccountSecretAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	// Batch delete secret records
	delReq := &protocloud.AccountSecretBatchDeleteReq{Filter: tools.ExpressionAnd(tools.RuleIn("id", req.IDs))}
	err = s.client.DataService().Global.AccountSecret.BatchDelete(cts.Kit, delReq)
	if err != nil {
		logs.Errorf("batch delete account secrets failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	// Clear Account Extension for resource secrets
	for _, accountID := range accountIDsToClear {
		if err = s.clearAccountExtension(cts.Kit, vendor, accountID); err != nil {
			logs.Errorf("clear account extension failed, err: %v, account_id: %s, rid: %s", err, accountID, cts.Kit.Rid)
			return nil, errf.Newf(errf.Aborted, "account secret records deleted, but failed to clear extension for "+
				"account %s: %v", accountID, err)
		}
	}

	return nil, nil
}

// validAndCollectClearExtAccountIDs validates secrets and collects account ids that need Extension clearing.
func (s *service) validAndCollectClearExtAccountIDs(kt *kit.Kit, vendor enumor.Vendor, bizID int64,
	secretIDs []string) ([]string, error) {

	listReq := &protocloud.AccountSecretListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", secretIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := s.client.DataService().Global.AccountSecret.ListAccountSecret(kt, listReq)
	if err != nil {
		logs.Errorf("list account secrets failed, err: %v, ids: %v, rid: %s", err, secretIDs, kt.Rid)
		return nil, err
	}
	if len(listResp.Details) != len(secretIDs) {
		secretIDMap := make(map[string]struct{})
		for _, secret := range listResp.Details {
			secretIDMap[secret.ID] = struct{}{}
		}
		var missingIDs []string
		for _, id := range secretIDs {
			if _, ok := secretIDMap[id]; !ok {
				missingIDs = append(missingIDs, id)
			}
		}
		return nil, errf.Newf(errf.RecordNotFound, "the following secret IDs do not exist: %v", missingIDs)
	}

	accountIDs := make([]string, 0, len(listResp.Details))
	for _, secret := range listResp.Details {
		accountIDs = append(accountIDs, secret.AccountID)
	}
	accountReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", accountIDs)),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "bk_biz_id"},
	}
	accountResp, err := s.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), accountReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, account ids: %v, rid: %s", err, accountIDs, kt.Rid)
		return nil, err
	}
	infoMap := make(map[string]*cloud.BaseAccount)
	for _, account := range accountResp.Details {
		infoMap[account.ID] = account
	}

	// Validate each secret and collect resource secret account ids
	accountIDSet := make(map[string]struct{})
	for _, secret := range listResp.Details {
		if secret.Vendor != vendor {
			return nil, errf.Newf(errf.InvalidParameter, "secret %s vendor (%s) does not match (%s)", secret.ID,
				secret.Vendor, vendor)
		}
		accountInfo, ok := infoMap[secret.AccountID]
		if !ok {
			return nil, errf.Newf(errf.RecordNotFound, "account %s associated with secret %s not found",
				secret.AccountID, secret.ID)
		}
		if accountInfo.BkBizID != bizID {
			return nil, errf.Newf(errf.PermissionDenied, "secret %s does not belong to business %d", secret.ID, bizID)
		}

		// Collect resource secret account ids
		if secret.Type == enumor.ResourceSecretType {
			accountIDSet[secret.AccountID] = struct{}{}
		}
	}

	return maps.Keys(accountIDSet), nil
}

// clearAccountExtension clears the account extension for the given vendor.
func (s *service) clearAccountExtension(kt *kit.Kit, vendor enumor.Vendor, accountID string) error {
	switch vendor {
	case enumor.TCloud:
		return s.clearTCloudAccountExtension(kt, accountID)
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}
