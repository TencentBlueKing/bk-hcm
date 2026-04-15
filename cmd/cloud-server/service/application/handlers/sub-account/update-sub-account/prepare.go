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

package updatesubaccount

import (
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/converter"
)

// updateSubAccountContent is the content stored in application.content for updating sub-account.
type updateSubAccountContent struct {
	subaccount.BaseSubAccountContent `json:",inline"`
	Req                              proto.SubAccountUpdateReq `json:"req"`
	SubAccountName                   string                    `json:"sub_account_name"`
}

// GenerateApplicationContent generate the content to be stored in DB.
func (a *ApplicationOfUpdateSubAccount) GenerateApplicationContent() interface{} {
	return &updateSubAccountContent{
		BaseSubAccountContent: subaccount.BaseSubAccountContent{
			Action:    enumor.SubAccountActionUpdate,
			Vendor:    a.Vendor(),
			BkBizID:   a.BkBizID(),
			AccountID: a.AccountID(),
		},
		Req:            converter.PtrToVal(a.req),
		SubAccountName: a.subAccountName,
	}
}
