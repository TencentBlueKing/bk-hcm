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

// Package rootaccountbillconfig ...
package rootaccountbillconfig

import (
	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initialize the bill config service.
func InitService(cap *capability.Capability) {
	svc := &rootBillConfigSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("ListRootAccountBillConfig", "POST", "/bills/root_account_config/list", svc.ListRootAccountBillConfig)
	h.Add("ListRootAccountBillConfigExt", "POST", "/vendors/{vendor}/bills/root_account_config/list",
		svc.ListRootAccountBillConfigExt)
	h.Add("GetRootBillConfig", "GET", "/vendors/{vendor}/bills/root_account_config/{id}", svc.GetRootAccountBillConfig)
	h.Add("BatchCreateRootAccountBillConfig", "POST", "/vendors/{vendor}/bills/root_account_config/batch/create",
		svc.BatchCreateRootAccountBillConfig)
	h.Add("BatchUpdateRootAccountBillConfig", "PATCH", "/vendors/{vendor}/bills/root_account_config/batch",
		svc.BatchUpdateRootAccountBillConfig)
	h.Add("BatchDeleteRootAccountBillConfig", "DELETE", "/bills/root_account_config/batch",
		svc.BatchDeleteRootAccountBillConfig)

	h.Load(cap.WebService)
}

type rootBillConfigSvc struct {
	dao dao.Set
}
