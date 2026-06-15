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

package subaccountsecret

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

var svc *subAccountSecretSvc

// InitService initial the sub account secret service
func InitService(cap *capability.Capability) {
	svc = &subAccountSecretSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateSubAccountSecret", http.MethodPost, "/vendors/{vendor}/sub_account_secrets/batch/create",
		svc.BatchCreateSubAccountSecret)
	h.Add("BatchUpdateSubAccountSecret", http.MethodPatch, "/vendors/{vendor}/sub_account_secrets/batch/update",
		svc.BatchUpdateSubAccountSecret)
	h.Add("BatchDeleteSubAccountSecret", http.MethodDelete, "/sub_account_secrets/batch",
		svc.BatchDeleteSubAccountSecret)
	h.Add("ListSubAccountSecret", http.MethodPost, "/sub_account_secrets/list", svc.ListSubAccountSecret)

	h.Add("ListSubAccountSecretWithExtension", "POST", "/vendors/{vendor}/sub_account_secrets/extensions/list",
		svc.ListSubAccountSecretWithExtension)
	h.Add("ListSubAccountSecretJoinExt", http.MethodPost, "/vendors/{vendor}/sub_account_secrets/list/join",
		svc.ListSubAccountSecretJoinExt)

	h.Load(cap.WebService)
}

type subAccountSecretSvc struct {
	dao dao.Set
}
