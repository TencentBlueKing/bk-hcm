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

// Package deletesecretkey provides the handler for deleting sub account secret key.
package deletesecretkey

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfDeleteSecretKey)(nil)

func init() {
	subaccount.RegisterOperationHandler(enumor.OpDeleteSubAccountSecret, newHandlerFromContent)
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent, content string,
) (handlers.ApplicationHandler, error) {

	ct := new(deleteSecretKeyContent)
	if err := json.UnmarshalFromString(content, ct); err != nil {
		return nil, fmt.Errorf("unmarshal delete secret key content failed, err: %w", err)
	}

	h := NewApplicationOfDeleteSecretKey(opt, base, ct.SecretID)
	h.SetAccountID(ct.AccountID)
	return h, nil
}

// ApplicationOfDeleteSecretKey handler for deleting sub account secret key.
type ApplicationOfDeleteSecretKey struct {
	subaccount.ApplicationBaseSubAccount

	secretID string
}

// NewApplicationOfDeleteSecretKey create a new handler.
func NewApplicationOfDeleteSecretKey(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent,
	secretID string) *ApplicationOfDeleteSecretKey {

	return &ApplicationOfDeleteSecretKey{
		ApplicationBaseSubAccount: subaccount.NewApplicationBaseSubAccount(opt, base),
		secretID:                  secretID,
	}
}
