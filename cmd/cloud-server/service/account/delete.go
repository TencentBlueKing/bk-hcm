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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// DeleteValidate ...
func (a *accountSvc) DeleteValidate(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// auth
	if err := a.checkPermission(cts, meta.Delete, accountID); err != nil {
		return nil, err
	}

	return a.client.DataService().Global.Account.DeleteValidate(cts.Kit.Ctx, cts.Kit.Header(), accountID)
}

// Delete ...
func (a *accountSvc) Delete(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// TODO 添加审计，最终是在dao去记录，因为账号删除有校验，存在删除不了的情况，且账号不需要在云上删除。

	// auth
	if err := a.checkPermission(cts, meta.Delete, accountID); err != nil {
		return nil, err
	}

	req := &protocloud.AccountDeleteReq{
		Filter: tools.EqualExpression("id", accountID),
	}
	return a.client.DataService().Global.Account.Delete(cts.Kit.Ctx, cts.Kit.Header(), req)
}
