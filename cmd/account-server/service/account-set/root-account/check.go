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

package rootaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// CheckDuplicateRootAccount 检查主账号是否重复
func CheckDuplicateRootAccount(cts *rest.Contexts, client *client.ClientSet, vendor enumor.Vendor,
	mainAccountIDFieldValue string) error {

	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	mainAccountIDFieldName := vendor.GetMainAccountIDField()
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", string(vendor)),
			tools.RuleJSONEqual(fmt.Sprintf("extension.%s", mainAccountIDFieldName), mainAccountIDFieldValue),
		),
		Page: core.NewCountPage(),
	}
	result, err := client.DataService().Global.RootAccount.List(cts.Kit, listReq)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("%s[%s] should be not duplicate", mainAccountIDFieldName, mainAccountIDFieldValue)
	}

	return nil
}
