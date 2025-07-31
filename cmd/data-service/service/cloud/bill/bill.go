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

// Package bill ...
package bill

import (
	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitBillConfigService initialize the bill config service.
func InitBillConfigService(cap *capability.Capability) {
	svc := &billConfigSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("ListBillConfig", "POST", "/bills/config/list", svc.ListBillConfig)
	h.Add("ListBillConfigExt", "POST", "/vendors/{vendor}/bills/config/list", svc.ListBillConfigExt)
	h.Add("GetBillConfig", "GET", "/vendors/{vendor}/bills/config/{id}", svc.GetBillConfig)
	h.Add("BatchCreateAccountBillConfig", "POST", "/vendors/{vendor}/bills/config/batch/create",
		svc.BatchCreateAccountBillConfig)
	h.Add("BatchUpdateAccountBillConfig", "PATCH", "/vendors/{vendor}/bills/config/batch",
		svc.BatchUpdateAccountBillConfig)
	h.Add("BatchDeleteAccountBillConfig", "DELETE", "/bills/config/batch",
		svc.BatchDeleteAccountBillConfig)

	h.Load(cap.WebService)
}

type billConfigSvc struct {
	dao dao.Set
}
