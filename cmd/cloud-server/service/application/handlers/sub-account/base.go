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

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
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
func (a *ApplicationBaseSubAccount) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return a.GetItsmPlatformAndAccountApprover(managers, a.accountID)
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
		return fmt.Errorf("query sub account failed, err: %w", err)
	}

	if result.Count == 0 {
		return fmt.Errorf("sub account(id=%s) not found", subAccountID)
	}

	return nil
}
