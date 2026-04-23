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

package subaccount

import (
	"fmt"
	"strconv"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protoaudit "hcm/pkg/api/data-service/audit"
	protocloud "hcm/pkg/api/data-service/cloud"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

// ActionHandlerFactory creates an ApplicationHandler for a specific sub-account action.
type ActionHandlerFactory func(opt *handlers.HandlerOption, base *BaseSubAccountContent, content string,
) (handlers.ApplicationHandler, error)

var actionHandlerRegistry = make(map[enumor.SubAccountAction]ActionHandlerFactory)

// RegisterActionHandler registers a handler factory for a sub-account action.
// Each action sub-package calls this in init() to self-register.
func RegisterActionHandler(action enumor.SubAccountAction, factory ActionHandlerFactory) {
	actionHandlerRegistry[action] = factory
}

// BaseSubAccountContent is the common header embedded in all sub-account application content structs.
// Each action's content struct embeds this base and adds action-specific fields.
type BaseSubAccountContent struct {
	Action    enumor.SubAccountAction `json:"action"`
	Vendor    enumor.Vendor           `json:"vendor"`
	BkBizID   int64                   `json:"bk_biz_id"`
	AccountID string                  `json:"account_id"`
}

// NewHandlerFromApplication dispatches to the registered action handler factory
// based on the action field in the application content.
func NewHandlerFromApplication(opt *handlers.HandlerOption, content string) (handlers.ApplicationHandler, error) {
	// 解析申请单公共内容，并根据Action构建对应handler
	ac := new(BaseSubAccountContent)
	if err := json.UnmarshalFromString(content, ac); err != nil {
		return nil, fmt.Errorf("unmarshal sub account action content failed, err: %w", err)
	}

	handler, ok := actionHandlerRegistry[ac.Action]
	if !ok {
		return nil, fmt.Errorf("unsupported sub account action: %s", ac.Action)
	}

	return handler(opt, ac, content)
}

// ApplicationBaseSubAccount is the shared base for all subaccount operation handlers.
// Each action-specific handler (create/update/delete subaccount, secret-key operations)
// embeds this base and only implements action-specific methods.
type ApplicationBaseSubAccount struct {
	handlers.BaseApplicationHandler

	action    enumor.SubAccountAction
	bkBizID   int64
	accountID string
}

// NewApplicationBaseSubAccount create a new base subaccount handler.
func NewApplicationBaseSubAccount(opt *handlers.HandlerOption, base *BaseSubAccountContent) ApplicationBaseSubAccount {
	return ApplicationBaseSubAccount{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(
			opt, enumor.OperateSubAccount, base.Vendor,
		),
		action:    base.Action,
		bkBizID:   base.BkBizID,
		accountID: base.AccountID,
	}
}

// Action return the subaccount action type.
func (a *ApplicationBaseSubAccount) Action() enumor.SubAccountAction {
	return a.action
}

// BkBizID return the biz ID.
func (a *ApplicationBaseSubAccount) BkBizID() int64 {
	return a.bkBizID
}

// AccountID return the parent account ID.
func (a *ApplicationBaseSubAccount) AccountID() string {
	return a.accountID
}

// SetAccountID set the parent account ID, typically called during CheckReq when the ID is resolved dynamically.
func (a *ApplicationBaseSubAccount) SetAccountID(accountID string) {
	a.accountID = accountID
}

// PrepareReq no pre-processing needed for subaccount operations.
func (a *ApplicationBaseSubAccount) PrepareReq() error {
	return nil
}

// PrepareReqFromContent no pre-processing needed when restoring from DB content.
// Action-specific handlers may override this if needed.
func (a *ApplicationBaseSubAccount) PrepareReqFromContent() error {
	return nil
}

// GetBkBizIDs get related biz IDs.
func (a *ApplicationBaseSubAccount) GetBkBizIDs() []int64 {
	if a.bkBizID > 0 {
		return []int64{a.bkBizID}
	}
	return nil
}

// GetItsmApprover get ITSM approvers for subaccount operations.
// The approvers are the parent 2nd-level account's managers queried by accountID,
// who serve as the approvers for the subaccount approval flow.
func (a *ApplicationBaseSubAccount) GetItsmApprover(kt *kit.Kit, managers []string) ([]itsm.VariableApprover, error) {
	return a.GetItsmPlatformAndAccountApprover(kt, managers, a.accountID)
}

// CheckSubAccountExists checks if the sub account exists.
func (a *ApplicationBaseSubAccount) CheckSubAccountExists(subAccountID string) error {
	result, err := a.Client.DataService().Global.SubAccount.List(
		a.Cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("id", subAccountID)),
			Page:   core.NewCountPage(),
		},
	)
	if err != nil {
		logs.Errorf("query sub account failed, id: %s, err: %v, rid: %s", subAccountID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("query sub account failed, err: %w", err)
	}

	if result.Count == 0 {
		logs.Errorf("sub account(id=%s) not found, rid: %s", subAccountID, a.Cts.Kit.Rid)
		return fmt.Errorf("sub account(id=%s) not found", subAccountID)
	}

	return nil
}

// QueryPermissionTemplateNames queries permission template names by IDs.
func (a *ApplicationBaseSubAccount) QueryPermissionTemplateNames(ids []string) ([]string, error) {
	result, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(
		a.Cts.Kit,
		&protocloud.PermissionTemplateListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", ids)),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		logs.Errorf("query permission template names failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return nil, err
	}
	if len(result.Details) != len(ids) {
		logs.Errorf("permission template names count mismatch, expected %d, got %d, rid: %s",
			len(ids), len(result.Details), a.Cts.Kit.Rid)
		return nil, fmt.Errorf("permission template names count mismatch, expected %d, got %d",
			len(ids), len(result.Details))
	}

	names := make([]string, 0, len(result.Details))
	for _, tmpl := range result.Details {
		names = append(names, tmpl.Name)
	}

	return names, nil
}

// CheckSubSecretExists checks if the sub account secret exists.
func (a *ApplicationBaseSubAccount) CheckSubSecretExists(subAccountID string) error {
	result, err := a.Client.DataService().Global.SubAccountSecret.ListSubAccountSecret(a.Cts.Kit,
		&protocloud.SubAccountSecretListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("sub_account_id", subAccountID)),
			Page:   core.NewCountPage(),
		},
	)
	if err != nil {
		logs.Errorf("query sub account secret failed, id: %s, err: %v, rid: %s", subAccountID,
			err, a.Cts.Kit.Rid)
		return fmt.Errorf("query sub account secret failed, err: %w", err)
	}
	if result.Count > 0 {
		logs.Errorf("sub account(%s) has sub account secrets, please delete the secrets first, rid: %s",
			subAccountID, a.Cts.Kit.Rid)
		return fmt.Errorf("sub account(%s) has sub account secrets, please delete the secrets first", subAccountID)
	}

	return nil
}

// ListPermissionTemplate lists permission templates by IDs.
func (a *ApplicationBaseSubAccount) ListPermissionTemplate(ids []string) ([]corecloud.BasePermissionTemplate, error) {
	result, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(
		a.Cts.Kit,
		&protocloud.PermissionTemplateListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", ids)),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		logs.Errorf("list permission templates failed, ids: %v, err: %v, rid: %s", ids, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("list permission templates failed, ids: %v, err: %w", ids, err)
	}

	if len(result.Details) != len(ids) {
		logs.Errorf("permission templates count mismatch, expected %d, got %d, rid: %s",
			len(ids), len(result.Details), a.Cts.Kit.Rid)
		return nil, fmt.Errorf("permission templates count mismatch, expected %d, got %d",
			len(ids), len(result.Details))
	}

	return result.Details, nil
}

// ParseTmlIDsFromCloudID parses permission template cloud_id to policy_id.
func (a *ApplicationBaseSubAccount) ParseTmlIDsFromCloudID(details []corecloud.BasePermissionTemplate,
) ([]uint64, error) {

	policyIDs := make([]uint64, 0, len(details))
	for _, tmpl := range details {
		cloudPolicyID, err := strconv.ParseUint(tmpl.CloudID, 10, 64)
		if err != nil {
			logs.Errorf("parse permission template cloud_id failed, cloud_id: %s, err: %v, rid: %s",
				tmpl.CloudID, err, a.Cts.Kit.Rid)
			return nil, err
		}

		policyIDs = append(policyIDs, cloudPolicyID)
	}

	return policyIDs, nil
}

// AttachPolicies attaches permission templates to sub account.
func (a *ApplicationBaseSubAccount) AttachPolicies(uin uint64, tmplIDs []string) error {
	templates, err := a.ListPermissionTemplate(tmplIDs)
	if err != nil {
		return fmt.Errorf("list permission templates to attach failed, ids: %v, err: %w", tmplIDs, err)
	}

	policyIDs, err := a.ParseTmlIDsFromCloudID(templates)
	if err != nil {
		return fmt.Errorf("parse permission template cloud_id for attach failed, err: %w", err)
	}

	if err = a.Client.HCService().TCloud.Account.AttachUserPolicies(
		a.Cts.Kit,
		&hssubaccount.TCloudAttachUserPoliciesReq{AccountID: a.AccountID(), TargetUin: uin, PolicyIDs: policyIDs},
	); err != nil {
		logs.Errorf("attach user policies failed, uin: %d, policy_ids: %v, err: %v, rid: %s",
			uin, policyIDs, err, a.Cts.Kit.Rid)
		return fmt.Errorf("attach user policies failed, uin: %d, policy_ids: %v, err: %w", uin, policyIDs, err)
	}

	return nil
}

// DetachPolicies detaches permission templates from sub account.
func (a *ApplicationBaseSubAccount) DetachPolicies(uin uint64, tmplIDs []string) error {
	templates, err := a.ListPermissionTemplate(tmplIDs)
	if err != nil {
		return fmt.Errorf("list permission templates to detach failed, ids: %v, err: %w", tmplIDs, err)
	}

	policyIDs, err := a.ParseTmlIDsFromCloudID(templates)
	if err != nil {
		return fmt.Errorf("parse permission template cloud_id for detach failed, err: %w", err)
	}

	if err = a.Client.HCService().TCloud.Account.DetachUserPolicies(
		a.Cts.Kit,
		&hssubaccount.TCloudDetachUserPoliciesReq{AccountID: a.AccountID(), DetachUin: uin, PolicyIDs: policyIDs},
	); err != nil {
		logs.Errorf("detach user policies failed, uin: %d, policy_ids: %v, err: %v, rid: %s",
			uin, policyIDs, err, a.Cts.Kit.Rid)
		return fmt.Errorf("detach user policies failed, uin: %d, policy_ids: %v, err: %w", uin, policyIDs, err)
	}

	return nil
}

// CreateAudit 创建审计记录，账号可能拥有不同的业务，所以需要放在上层，获取路由的业务作为BizID。
func (a *ApplicationBaseSubAccount) CreateAudit(action enumor.AuditAction, resType enumor.AuditResourceType,
	resID, resName string, detail interface{}) error {

	return a.Audit.BatchCreateAudit(a.Cts.Kit, &protoaudit.BatchCreateAuditReq{
		Audits: []protoaudit.BatchCreateAuditInfo{
			{
				ResID:     resID,
				ResName:   resName,
				ResType:   resType,
				Action:    action,
				BkBizID:   a.BkBizID(),
				Vendor:    a.Vendor(),
				AccountID: a.AccountID(),
				Detail:    detail,
			},
		},
	})
}
