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
	"sort"
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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// PolicyLibraryApplier provides the complete apply logic for permission policy libraries.
type PolicyLibraryApplier struct {
	client *client.ClientSet
	audit  audit.Interface
}

// CreateTmplBaseInfo represents the create permission template information.
type CreateTmplBaseInfo struct {
	Name string  `json:"name"`
	Memo *string `json:"memo"`
}

// UpdateTmplBaseInfo represents the update permission template information.
type UpdateTmplBaseInfo struct {
	Memo *string `json:"memo"`
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
		return a.applyTCloudCreate(kt, library, accountIDs,
			CreateTmplBaseInfo{Name: library.Name, Memo: library.Memo}), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// ApplyCreateWithTmplInfo applies a permission policy library (create) to the given accounts with template info.
func (a *PolicyLibraryApplier) ApplyCreateWithTmplInfo(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
	accountIDs []string, tmplInfo CreateTmplBaseInfo) (*proto.ApplyPermissionPolicyLibraryResult, error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	if err := a.CheckAccountsBizInScope(kt, library.BkBizIDs, accountIDs); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return a.applyTCloudCreate(kt, library, accountIDs, tmplInfo), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (a *PolicyLibraryApplier) applyTCloudCreate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountIDs []string, tmplInfo CreateTmplBaseInfo) *proto.ApplyPermissionPolicyLibraryResult {

	results := make([]proto.ApplyAccountResult, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		results = append(results, a.applyTCloudCreateForAccount(kt, library, accountID, tmplInfo))
	}
	return &proto.ApplyPermissionPolicyLibraryResult{Results: results}
}

func (a *PolicyLibraryApplier) applyTCloudCreateForAccount(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountID string, tmplInfo CreateTmplBaseInfo) proto.ApplyAccountResult {

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

	camResult, err := a.createTCloudCAMPolicy(kt, library, accountID, tmplInfo)
	if err != nil {
		return proto.ApplyAccountResult{
			AccountID: accountID, Status: proto.ApplyStatusFailed, Reason: err.Error(),
		}
	}

	if err = a.createTCloudLocalTemplate(kt, library, accountID, camResult.PolicyID, tmplInfo); err != nil {
		return proto.ApplyAccountResult{
			AccountID: accountID,
			Status:    proto.ApplyStatusFailed,
			Reason:    fmt.Sprintf("云策略已创建(id=%d), 但本地模板创建失败: %v", camResult.PolicyID, err),
		}
	}

	if err = a.RecordApplyCreateAudit(kt, library.ID, accountID); err != nil {
		logs.Errorf("record apply create audit failed, library_id: %s, account_id: %s, err: %v, rid: %s",
			library.ID, accountID, err, kt.Rid)
		return proto.ApplyAccountResult{
			AccountID: accountID,
			Status:    proto.ApplyStatusFailed,
			Reason:    fmt.Sprintf("云策略已创建(id=%d), 但记录审计失败: %v", camResult.PolicyID, err),
		}
	}

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
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch), tools.RuleEqual("type", enumor.ResourceAccount)),
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
			"some accounts not found, expected %d but got %d", len(accountIDs), len(accounts))
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

// createTCloudCAMPolicy calls hc-service to create a CAM policy on TCloud.
func (a *PolicyLibraryApplier) createTCloudCAMPolicy(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountID string, tmplInfo CreateTmplBaseInfo) (*hspermissiontemplate.CreateCAMPolicyResult, error) {

	camReq := &hspermissiontemplate.CreateCAMPolicyReq{
		AccountID:      accountID,
		PolicyName:     tmplInfo.Name,
		PolicyDocument: library.PolicyDocument,
		Description:    cvt.PtrToVal(tmplInfo.Memo),
	}

	result, err := a.client.HCService().TCloud.PermissionTemplate.CreateCAMPolicy(kt, camReq)
	if err != nil {
		logs.Errorf("tcloud create cam policy failed, accountID: %s, err: %v, rid: %s",
			accountID, err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// createTCloudLocalTemplate creates a local permission_template record after the cloud policy is created.
func (a *PolicyLibraryApplier) createTCloudLocalTemplate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountID string, cloudPolicyID uint64, tmplInfo CreateTmplBaseInfo) error {

	now := time.Now().UTC().Format(time.RFC3339)
	dsReq := &protocloud.PermissionTemplateBatchCreateReq[corecloud.TCloudPermissionTemplateExtension]{
		PermissionTemplates: []protocloud.PermissionTemplateCreate[corecloud.TCloudPermissionTemplateExtension]{
			{
				CloudID:               strconv.FormatUint(cloudPolicyID, 10),
				Name:                  tmplInfo.Name,
				AccountID:             accountID,
				PolicyLibraryID:       cvt.ValToPtr(library.ID),
				PolicyLibraryVersion:  cvt.ValToPtr(library.Version),
				PolicyLibrarySyncTime: cvt.ValToPtr(now),
				PolicyDocument:        library.PolicyDocument,
				Memo:                  tmplInfo.Memo,
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

// RecordApplyCreateAudit records an apply audit log.
func (a *PolicyLibraryApplier) RecordApplyCreateAudit(kt *kit.Kit, libraryID, accountID string) error {
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
		return err
	}

	return nil
}

// ApplyUpdate applies a permission policy library (update) to the given template IDs.
func (a *PolicyLibraryApplier) ApplyUpdate(kt *kit.Kit, vendor enumor.Vendor, libraryID string, templateIDs []string) (
	*proto.ApplyPermissionPolicyLibraryUpdateResult, error) {

	if err := a.CheckPermTmplUpdatability(kt, vendor, templateIDs, libraryID); err != nil {
		logs.Errorf("check permission template updatability failed, libraryID: %s, templateIDs: %v, err: %v, rid: %s",
			libraryID, templateIDs, err, kt.Rid)
		return nil, err
	}

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	accountIDs, err := a.GetPermTmplAccountIDs(kt, templateIDs)
	if err != nil {
		return nil, err
	}
	if err = a.CheckAccountsBizInScope(kt, library.BkBizIDs, accountIDs); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return a.applyTCloudUpdate(kt, library, templateIDs, UpdateTmplBaseInfo{Memo: library.Memo}), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// ApplyUpdateWithTmplInfo applies a permission policy library (update) to the given accounts with template info.
func (a *PolicyLibraryApplier) ApplyUpdateWithTmplInfo(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
	templateIDs []string, tmplInfo UpdateTmplBaseInfo) (*proto.ApplyPermissionPolicyLibraryUpdateResult, error) {

	if err := a.CheckPermTmplUpdatability(kt, vendor, templateIDs, libraryID); err != nil {
		logs.Errorf("check permission template updatability failed, libraryID: %s, templateIDs: %v, err: %v, rid: %s",
			libraryID, templateIDs, err, kt.Rid)
		return nil, err
	}

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	accountIDs, err := a.GetPermTmplAccountIDs(kt, templateIDs)
	if err != nil {
		return nil, err
	}
	if err = a.CheckAccountsBizInScope(kt, library.BkBizIDs, accountIDs); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return a.applyTCloudUpdate(kt, library, templateIDs, tmplInfo), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// GetPermTmplAccountIDs gets the account IDs of the given template IDs.
func (a *PolicyLibraryApplier) GetPermTmplAccountIDs(kt *kit.Kit, templateIDs []string) ([]string, error) {
	accountIDMap := make(map[string]struct{})
	for _, batch := range slice.Split(templateIDs, constant.BatchOperationMaxLimit) {
		req := &protocloud.PermissionTemplateListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := a.client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, req)
		if err != nil {
			logs.Errorf("get permission template failed, err: %v, ids: %v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		for _, one := range result.Details {
			accountIDMap[one.AccountID] = struct{}{}
		}
	}
	return maps.Keys(accountIDMap), nil
}

func (a *PolicyLibraryApplier) applyTCloudUpdate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	templateIDs []string, tmplInfo UpdateTmplBaseInfo) *proto.ApplyPermissionPolicyLibraryUpdateResult {

	results := make([]proto.ApplyTemplateResult, 0, len(templateIDs))
	for _, templateID := range templateIDs {
		results = append(results, a.applyTCloudUpdateForTemplate(kt, library, templateID, tmplInfo))
	}
	return &proto.ApplyPermissionPolicyLibraryUpdateResult{Results: results}
}

func (a *PolicyLibraryApplier) applyTCloudUpdateForTemplate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	templateID string, tmplInfo UpdateTmplBaseInfo) proto.ApplyTemplateResult {

	templates, err := a.getTCloudTemplateByIDs(kt, []string{templateID})
	if err != nil {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID, Status: proto.ApplyStatusFailed, Reason: err.Error(),
		}
	}
	if len(templates) != 1 {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID, Status: proto.ApplyStatusFailed, Reason: "template not found",
		}
	}
	tmpl := templates[0]

	cloudPolicyID, err := strconv.ParseUint(tmpl.CloudID, 10, 64)
	if err != nil {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID,
			Status:               proto.ApplyStatusFailed,
			Reason:               fmt.Sprintf("parse cloud policy id failed: %v", err),
		}
	}

	if err = a.updateTCloudCAMPolicy(kt, library, tmpl.AccountID, cloudPolicyID, tmplInfo); err != nil {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID, Status: proto.ApplyStatusFailed, Reason: err.Error(),
		}
	}

	if err = a.updateTCloudLocalTemplate(kt, library, templateID, tmplInfo); err != nil {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID,
			Status:               proto.ApplyStatusFailed,
			Reason: fmt.Sprintf("云策略已更新(cloudPolicyID=%d), 但本地模板更新失败: %v", cloudPolicyID,
				err),
		}
	}

	if err = a.RecordApplyUpdateAudit(kt, library.ID, tmpl.ID); err != nil {
		return proto.ApplyTemplateResult{
			PermissionTemplateID: templateID,
			Status:               proto.ApplyStatusFailed,
			Reason:               fmt.Sprintf("云策略已更新(cloudPolicyID=%d), 但审计创建失败: %v", cloudPolicyID, err),
		}
	}

	return proto.ApplyTemplateResult{PermissionTemplateID: templateID, Status: proto.ApplyStatusSuccess}
}

// getTCloudTemplateByIDs retrieves the Cloud permission template by IDs.
func (a *PolicyLibraryApplier) getTCloudTemplateByIDs(kt *kit.Kit, templateIDs []string) (
	[]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], error) {

	result := make([]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], 0)
	for _, batch := range slice.Split(templateIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.PermissionTemplateExtListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch), tools.RuleEqual("vendor", enumor.TCloud)),
			Page:   core.NewDefaultBasePage(),
		}
		batchResult, err := a.client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
		if err != nil {
			return nil, err
		}
		result = append(result, batchResult.Details...)
	}

	if len(result) != len(templateIDs) {
		logs.Errorf("get tcloud template failed, expected: %d, got: %d, templateIDs: %v, rid: %s", len(templateIDs),
			len(result), templateIDs, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter,
			fmt.Errorf("get tcloud template failed, expected: %d, got: %d", len(templateIDs), len(result)))
	}

	return result, nil
}

// updateTCloudCAMPolicy calls hc-service to update a CAM policy on TCloud.
func (a *PolicyLibraryApplier) updateTCloudCAMPolicy(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	accountID string, cloudPolicyID uint64, tmplInfo UpdateTmplBaseInfo) error {

	camReq := &hspermissiontemplate.UpdateCAMPolicyReq{
		AccountID:      accountID,
		PolicyID:       cloudPolicyID,
		PolicyDocument: cvt.ValToPtr(library.PolicyDocument),
		Description:    tmplInfo.Memo,
	}

	if err := a.client.HCService().TCloud.PermissionTemplate.UpdateCAMPolicy(kt, camReq); err != nil {
		logs.Errorf("tcloud update cam policy failed, accountID: %s, policyID: %d, err: %v, rid: %s",
			accountID, cloudPolicyID, err, kt.Rid)
		return err
	}

	return nil
}

// updateTCloudLocalTemplate updates the local permission_template record after the cloud policy is updated.
func (a *PolicyLibraryApplier) updateTCloudLocalTemplate(kt *kit.Kit, library *corecloud.BasePermissionPolicyLibrary,
	templateID string, tmplInfo UpdateTmplBaseInfo) error {

	now := time.Now().UTC().Format(time.RFC3339)
	updateFields := map[string]interface{}{
		"policy_document":          library.PolicyDocument,
		"policy_library_id":        library.ID,
		"policy_library_version":   library.Version,
		"policy_library_sync_time": now,
		"memo":                     tmplInfo.Memo,
	}
	if err := a.audit.ResUpdateAudit(kt, enumor.PermissionTemplateAuditResType, templateID, updateFields); err != nil {
		logs.Errorf("tcloud update permission template failed, templateID: %s, err: %v, rid: %s",
			templateID, err, kt.Rid)
		return err
	}

	dsReq := &protocloud.PermissionTemplateBatchUpdateReq[corecloud.TCloudPermissionTemplateExtension]{
		PermissionTemplates: []protocloud.PermissionTemplateUpdate[corecloud.TCloudPermissionTemplateExtension]{
			{
				ID:                    templateID,
				PolicyDocument:        library.PolicyDocument,
				PolicyLibraryID:       cvt.ValToPtr(library.ID),
				PolicyLibraryVersion:  cvt.ValToPtr(library.Version),
				PolicyLibrarySyncTime: cvt.ValToPtr(now),
				Memo:                  tmplInfo.Memo,
			},
		},
	}

	if err := a.client.DataService().TCloud.PermissionTemplate.BatchUpdate(kt, dsReq); err != nil {
		logs.Errorf("tcloud update permission template failed, templateID: %s, err: %v, rid: %s",
			templateID, err, kt.Rid)
		return err
	}

	return nil
}

// RecordApplyUpdateAudit records an apply audit log. Errors are logged but not returned.
func (a *PolicyLibraryApplier) RecordApplyUpdateAudit(kt *kit.Kit, libraryID, permTmplID string) error {
	err := a.audit.ResOperationAudit(kt, protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.PermissionPolicyLibraryAuditResType,
		ResID:             libraryID,
		Action:            protoaudit.ApplyOp,
		AssociatedResType: enumor.PermissionTemplateAuditResType,
		AssociatedResID:   permTmplID,
	})
	if err != nil {
		logs.Errorf("record apply audit failed, libraryID: %s, permTmplID: %s, err: %v, rid: %s",
			libraryID, permTmplID, err, kt.Rid)
		return err
	}

	return nil
}

// ListAllAppliedAccountIDs scans permission_template table for all account IDs applied to the given library.
func (a *PolicyLibraryApplier) ListAllAppliedAccountIDs(kt *kit.Kit, libraryID string) ([]string, error) {
	accountIDSet := make(map[string]struct{})
	start := uint32(0)
	for {
		req := &protocloud.PermissionTemplateListReq{
			Filter: tools.EqualExpression("policy_library_id", libraryID),
			Page:   &core.BasePage{Start: start, Limit: core.DefaultMaxPageLimit},
		}
		result, err := a.client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, req)
		if err != nil {
			logs.Errorf("list permission template failed, libraryID: %s, err: %v, rid: %s", libraryID, err, kt.Rid)
			return nil, err
		}
		for _, tmpl := range result.Details {
			accountIDSet[tmpl.AccountID] = struct{}{}
		}
		if uint(len(result.Details)) < core.DefaultMaxPageLimit {
			break
		}
		start += uint32(core.DefaultMaxPageLimit)
	}
	return maps.Keys(accountIDSet), nil
}

// listAllInScopeAccountIDs scans account table for all account IDs matching the given vendor and biz IDs.
func (a *PolicyLibraryApplier) listAllInScopeAccountIDs(kt *kit.Kit, vendor enumor.Vendor, bizIDs []int64) (
	[]string, error) {

	if len(bizIDs) == 0 {
		return nil, nil
	}

	accountIDs := make([]string, 0)
	for _, batch := range slice.Split(bizIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.AccountListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("type", enumor.ResourceAccount), tools.RuleIn("bk_biz_id", batch)),
			Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		}
		for {
			result, err := a.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
			if err != nil {
				logs.Errorf("list accounts in scope failed, vendor: %s, bizID: %v, err: %v, rid: %s", vendor, batch,
					err, kt.Rid)
				return nil, err
			}
			for _, acc := range result.Details {
				accountIDs = append(accountIDs, acc.ID)
			}
			if uint(len(result.Details)) < core.DefaultMaxPageLimit {
				break
			}
			req.Page.Start += uint32(req.Page.Limit)
		}
	}

	return slice.Unique(accountIDs), nil
}

// ListUnAppliedAccountIDs returns account IDs that are in scope but have not applied the given policy library.
func (a *PolicyLibraryApplier) ListUnAppliedAccountIDs(kt *kit.Kit, vendor enumor.Vendor, libraryID string) (
	[]string, error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	return a.computeUnAppliedAccountIDs(kt, vendor, libraryID, library.BkBizIDs)
}

// ListBizUnAppliedAccountIDs returns account IDs that belong to bizID, are in the library's biz scope,
// and have not applied the given policy library.
func (a *PolicyLibraryApplier) ListBizUnAppliedAccountIDs(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
	bizID int64) ([]string, error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	inScope := false
	for _, biz := range library.BkBizIDs {
		if biz == bizID {
			inScope = true
			break
		}
	}
	if !inScope {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id %d is not in policy library scope", bizID)
	}

	return a.computeUnAppliedAccountIDs(kt, vendor, libraryID, []int64{bizID})
}

// computeUnAppliedAccountIDs lists in-scope account IDs for the given bizIDs, then subtracts the already-applied
// ones, returning a sorted slice of account IDs that have not yet applied the policy library.
func (a *PolicyLibraryApplier) computeUnAppliedAccountIDs(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
	bizIDs []int64) ([]string, error) {

	inScopeAccountIDs, err := a.listAllInScopeAccountIDs(kt, vendor, bizIDs)
	if err != nil {
		return nil, err
	}

	appliedAccountIDs, err := a.ListAllAppliedAccountIDs(kt, libraryID)
	if err != nil {
		return nil, err
	}

	unAppliedAccountIDs := slice.NotIn(appliedAccountIDs, inScopeAccountIDs)
	sort.Strings(unAppliedAccountIDs)
	return unAppliedAccountIDs, nil
}

// ListTemplatesInScope returns all permission templates applied from the given library
// whose associated accounts are still within the library's current biz scope.
func (a *PolicyLibraryApplier) ListTemplatesInScope(kt *kit.Kit, vendor enumor.Vendor, libraryID string) (any, error) {
	switch vendor {
	case enumor.TCloud:
		return a.listTCloudTemplatesInScope(kt, libraryID)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// listTCloudTemplatesInScope returns TCloud permission templates applied from the given library
// whose associated accounts are still within the library's current biz scope.
func (a *PolicyLibraryApplier) listTCloudTemplatesInScope(kt *kit.Kit, libraryID string) (
	[]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	return a.listTCloudBizTemplatesInScope(kt, libraryID, library.BkBizIDs)
}

// ListBizTemplatesInScope returns all permission templates applied from the given library
// whose associated accounts have their management biz equal to bizID.
func (a *PolicyLibraryApplier) ListBizTemplatesInScope(kt *kit.Kit, vendor enumor.Vendor, libraryID string,
	bizID int64) (any, error) {

	library, err := a.GetPolicyLibraryDetail(kt, libraryID)
	if err != nil {
		return nil, err
	}

	inScope := false
	for _, biz := range library.BkBizIDs {
		if biz == bizID {
			inScope = true
			break
		}
	}
	if !inScope {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id %d is not in policy library scope", bizID)
	}

	switch vendor {
	case enumor.TCloud:
		return a.listTCloudBizTemplatesInScope(kt, libraryID, []int64{bizID})
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// listTCloudBizTemplatesInScope returns TCloud permission templates applied from the given library
// whose associated accounts have their management biz equal to bizID.
func (a *PolicyLibraryApplier) listTCloudBizTemplatesInScope(kt *kit.Kit, libraryID string, bizIDs []int64) (
	[]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], error) {

	accountIDs, err := a.listAllInScopeAccountIDs(kt, enumor.TCloud, bizIDs)
	if err != nil {
		return nil, err
	}

	conditions := make([]*filter.AtomRule, 0)
	conditions = append(conditions, tools.RuleEqual("policy_library_id", libraryID))
	for _, batch := range slice.Split(accountIDs, int(filter.DefaultMaxInLimit)) {
		conditions = append(conditions, tools.RuleIn("account_id", batch))
	}

	req := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(conditions...),
		Page:   core.NewDefaultBasePage(),
	}
	details := make([]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], 0)
	for {
		result, err := a.client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
		if err != nil {
			logs.Errorf("list permission template ext failed, libraryID: %s, err: %v, rid: %s", libraryID, err, kt.Rid)
			return nil, err
		}
		details = append(details, result.Details...)
		if uint(len(result.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return details, nil
}

// CheckPermTmplUpdatability checks whether the given permission template can be updated.
func (a *PolicyLibraryApplier) CheckPermTmplUpdatability(kt *kit.Kit, vendor enumor.Vendor, templateIDs []string,
	policyLibraryID string) error {

	switch vendor {
	case enumor.TCloud:
		return a.checkTCloudPermTmplUpdatability(kt, templateIDs, policyLibraryID)
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (a *PolicyLibraryApplier) checkTCloudPermTmplUpdatability(kt *kit.Kit, templateIDs []string,
	policyLibraryID string) error {

	if policyLibraryID == "" {
		logs.Errorf("policy library id is required, templateIDs: %v, rid: %s", templateIDs, kt.Rid)
		return errf.Newf(errf.InvalidParameter, "policy library id is required")
	}

	templates, err := a.getTCloudTemplateByIDs(kt, templateIDs)
	if err != nil {
		logs.Errorf("get tcloud template failed, templateID: %s, err: %v, rid: %s", templateIDs[0], err, kt.Rid)
		return err
	}

	for _, template := range templates {
		if template.Extension == nil {
			logs.Errorf("tcloud template extension not found, templateID: %s, rid: %s", template.ID, kt.Rid)
			return errf.Newf(errf.InvalidParameter, "tcloud template %s extension not found", template.ID)
		}
		// 权限模版的权限策略库id为空时，只有自定义的权限模版可以更新
		if cvt.PtrToVal(template.PolicyLibraryID) == "" && template.Extension.CloudType == enumor.TCloudCustomPolicy {
			continue
		}

		// 权限模版的权限策略库id不为空时, 权限模版的权限策略库id必须和想要应用的权限策略库id一致
		if cvt.PtrToVal(template.PolicyLibraryID) != "" && cvt.PtrToVal(template.PolicyLibraryID) == policyLibraryID {
			continue
		}

		logs.Errorf("tcloud template permission policy library id not match, templateID: %s, policy library id: %s, "+
			"rid: %s", template.ID, template.PolicyLibraryID, kt.Rid)
		return errf.Newf(errf.InvalidParameter, "tcloud template permission policy library id not match, "+
			"templateID: %s, policy library id: %s", template.ID, cvt.PtrToVal(template.PolicyLibraryID))
	}

	return nil
}
