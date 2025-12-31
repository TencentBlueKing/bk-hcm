/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetTCloudZiyanAccount 获取自研云账号信息
func getTCloudZiyanAccount(kt *kit.Kit, client *client.ClientSet) (string, error) {
	accountReq := &protocloud.AccountListReq{
		Page: core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleEqual("type", enumor.ResourceAccount),
		),
	}

	accountList, err := client.DataService().Global.Account.List(kt.Ctx, kt.Header(), accountReq)
	if err != nil {
		logs.Errorf("get %s account list failed, err: %v, rid: %s", enumor.TCloudZiyan, err, kt.Rid)
		return "", err
	}

	if len(accountList.Details) == 0 {
		return "", errf.Newf(errf.RecordNotFound, "vendor: %s, account not found", enumor.TCloudZiyan)
	}

	return accountList.Details[0].ID, nil
}
