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

package permissionpolicylibrary

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/cloud-server/logics/audit"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protoaudit "hcm/pkg/api/data-service/audit"
	protocloud "hcm/pkg/api/data-service/cloud"
	hspermissiontemplate "hcm/pkg/api/hc-service/permission-template"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// PolicyLibraryApplier provides the complete apply logic for permission policy libraries.
type PolicyLibraryApplier struct {
	client *client.ClientSet
	audit  audit.Interface
}

// NewPolicyLibraryApplier creates a new PolicyLibraryApplier.
func NewPolicyLibraryApplier(cli *client.ClientSet, audit audit.Interface) *PolicyLibraryApplier {
	return &PolicyLibraryApplier{client: cli, audit: audit}
}

// ApplyCreate applies a permission policy library (create) to the given accounts.
func (a *PolicyLibraryApplier) ApplyCreate(kt *kit.Kit, vendor enumor.Vendor, libraryID string, accountIDs []string) (
	*proto.ApplyPermissionPolicyLibraryResult, error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	if err := a.CheckAccountsBizInScope(kt, library.BkBizIDs, accountIDs); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return a.tcloudApplyCreate(kt, library, accountIDs), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (a *PolicyLibraryApplier) tcloudApplyCreate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountIDs []string) *proto.ApplyPermissionPolicyLibraryResult {

	results := make([]proto.ApplyAccountResult, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		results = append(results, a.tcloudApplyCreateForAccount(kt, library, accountID))
	}
	return &proto.ApplyPermissionPolicyLibraryResult{Results: results}
}

func (a *PolicyLibraryApplier) tcloudApplyCreateForAccount(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountID string) proto.ApplyAccountResult {

	applied, err := a.CheckAccountApplied(kt, library.ID, accountID)
	if err != nil {
		return proto.ApplyAccountResult{
			AccountID: accountID, Status: proto.ApplyStatusFailed, Reason: err.Error(),
		}
	}
	if applied {
		return proto.ApplyAccountResult{
			AccountID: accountID,
			Status:    proto.ApplyStatusFailed,
			Reason:    "该二级账号已应用此权限策略库",
		}
	}

	camResult, err := a.TCloudCreateCAMPolicy(kt, library, accountID)
	if err != nil {
		return proto.ApplyAccountResult{
			AccountID: accountID, Status: proto.ApplyStatusFailed, Reason: err.Error(),
		}
	}

	if err = a.TCloudCreateLocalTemplate(kt, library, accountID, camResult.PolicyID); err != nil {
		return proto.ApplyAccountResult{
			AccountID: accountID,
			Status:    proto.ApplyStatusFailed,
			Reason:    fmt.Sprintf("云策略已创建(id=%d), 但本地模板创建失败: %v", camResult.PolicyID, err),
		}
	}

	a.RecordApplyAudit(kt, library.ID, accountID)

	return proto.ApplyAccountResult{AccountID: accountID, Status: proto.ApplyStatusSuccess}
}

// GetPolicyLibraryDetail retrieves the permission policy library detail by ID.
func (a *PolicyLibraryApplier) GetPolicyLibraryDetail(kt *kit.Kit, id string) (
	*corecloud.BasePermissionPolicyLibrary, error) {

	req := &protocloud.PermissionPolicyLibraryListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}

	result, err := a.client.DataService().Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(kt, req)
	if err != nil {
		logs.Errorf("list permission policy library failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	if result == nil || len(result.Details) == 0 {
		return nil, fmt.Errorf("permission policy library not found, id: %s", id)
	}

	return &result.Details[0], nil
}

// CheckAccountsBizInScope checks if all accounts' management biz IDs are within the allowed biz scope.
func (a *PolicyLibraryApplier) CheckAccountsBizInScope(kt *kit.Kit, allowedBkBizIDs []int64,
	accountIDs []string) error {

	allowedSet := make(map[int64]struct{}, len(allowedBkBizIDs))
	for _, bizID := range allowedBkBizIDs {
		allowedSet[bizID] = struct{}{}
	}

	accounts := make([]*corecloud.BaseAccount, 0)
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.AccountListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := a.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list accounts for biz scope check failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		accounts = append(accounts, result.Details...)
	}

	if len(accounts) != len(accountIDs) {
		return errf.Newf(errf.InvalidParameter,
			"some accounts not found, expected %d but got %d", len(accountIDs), accounts)
	}

	outOfScope := make([]string, 0)
	for _, account := range accounts {
		if _, ok := allowedSet[account.BkBizID]; !ok {
			outOfScope = append(outOfScope, fmt.Sprintf("%s(bk_biz_id=%d)", account.ID, account.BkBizID))
		}
	}

	if len(outOfScope) > 0 {
		return errf.Newf(errf.InvalidParameter,
			"accounts' management biz not in policy library scope: %s", strings.Join(outOfScope, ", "))
	}

	return nil
}

// CheckAccountApplied checks if the given account already has a permission template from the given library.
func (a *PolicyLibraryApplier) CheckAccountApplied(kt *kit.Kit, libraryID, accountID string) (bool, error) {
	expr, err := tools.And(
		tools.EqualExpression("policy_library_id", libraryID),
		tools.EqualExpression("account_id", accountID),
	)
	if err != nil {
		return false, err
	}

	req := &protocloud.PermissionTemplateExtListReq{
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
	}

	result, err := a.client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
	if err != nil {
		logs.Errorf("list permission template failed, libraryID: %s, accountID: %s, err: %v, rid: %s",
			libraryID, accountID, err, kt.Rid)
		return false, err
	}

	return result != nil && len(result.Details) > 0, nil
}

// TCloudCreateCAMPolicy calls hc-service to create a CAM policy on TCloud.
func (a *PolicyLibraryApplier) TCloudCreateCAMPolicy(kt *kit.Kit,
	library *corecloud.BasePermissionPolicyLibrary, accountID string) (
	*hspermissiontemplate.CreateCAMPolicyResult, error) {

	camReq := &hspermissiontemplate.CreateCAMPolicyReq{
		AccountID:      accountID,
		PolicyName:     library.Name,
		PolicyDocument: library.PolicyDocument,
		Description:    fmt.Sprintf("Created from policy library %s (v%d)", library.Name, library.Version),
	}

	result, err := a.client.HCService().TCloud.PermissionTemplate.CreateCAMPolicy(kt, camReq)
	if err != nil {
		logs.Errorf("tcloud create cam policy failed, accountID: %s, err: %v, rid: %s",
			accountID, err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// TCloudCreateLocalTemplate creates a local permission_template record after the cloud policy is created.
func (a *PolicyLibraryApplier) TCloudCreateLocalTemplate(kt *kit.Kit,
	library *corecloud.BasePermissionPolicyLibrary, accountID string, cloudPolicyID uint64) error {

	now := time.Now().UTC().Format(time.RFC3339)
	dsReq := &protocloud.PermissionTemplateBatchCreateReq[corecloud.TCloudPermissionTemplateExtension]{
		PermissionTemplates: []protocloud.PermissionTemplateCreate[corecloud.TCloudPermissionTemplateExtension]{
			{
				CloudID:               strconv.FormatUint(cloudPolicyID, 10),
				Name:                  library.Name,
				AccountID:             accountID,
				PolicyLibraryID:       cvt.ValToPtr(library.ID),
				PolicyLibraryVersion:  cvt.ValToPtr(library.Version),
				PolicyLibrarySyncTime: cvt.ValToPtr(now),
				PolicyDocument:        library.PolicyDocument,
				Memo:                  cvt.ValToPtr(fmt.Sprintf("Applied from policy library: %s", library.Name)),
				Extension: &corecloud.TCloudPermissionTemplateExtension{
					CloudType: enumor.TCloudCustomPolicy,
				},
			},
		},
	}

	_, err := a.client.DataService().TCloud.PermissionTemplate.BatchCreate(kt, dsReq)
	if err != nil {
		logs.Errorf("tcloud create permission template failed, accountID: %s, cloudPolicyID: %d, err: %v, rid: %s",
			accountID, cloudPolicyID, err, kt.Rid)
		return err
	}

	return nil
}

// RecordApplyAudit records an apply audit log. Errors are logged but not returned.
func (a *PolicyLibraryApplier) RecordApplyAudit(kt *kit.Kit, libraryID, accountID string) {
	err := a.audit.ResOperationAudit(kt, protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.PermissionPolicyLibraryAuditResType,
		ResID:             libraryID,
		Action:            protoaudit.ApplyOp,
		AssociatedResType: enumor.AccountAuditResType,
		AssociatedResID:   accountID,
	})
	if err != nil {
		logs.Errorf("record apply audit failed, libraryID: %s, accountID: %s, err: %v, rid: %s",
			libraryID, accountID, err, kt.Rid)
	}
}
