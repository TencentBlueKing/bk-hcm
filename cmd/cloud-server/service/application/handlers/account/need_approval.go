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

package account

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// autoApproveAccountTypes 免审白名单：命中的账号类型无需走ITSM，也不落申请单，同步交付
// 后续如需按云厂商/业务/责任人等进一步收窄免审范围，在此扩展判断条件即可
var autoApproveAccountTypes = map[enumor.AccountType]struct{}{
	enumor.RegistrationAccount: {},
}

// NeedApproval 登记账号录入免审：命中白名单账号类型时跳过ITSM与申请单，同步交付账号
func (a *ApplicationOfAddAccount) NeedApproval(kt *kit.Kit) (bool, error) {
	if _, ok := autoApproveAccountTypes[a.req.Type]; ok {
		return false, nil
	}
	return true, nil
}
