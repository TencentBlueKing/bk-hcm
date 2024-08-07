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

// Package mainaccount ...
package mainaccount

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/cryptography"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initial the main account service
func InitService(cap *capability.Capability) {
	svc := &service{
		dao:    cap.Dao,
		cipher: cap.Cipher,
	}

	h := rest.NewHandler()

	h.Add("CreateMainAccount", http.MethodPost, "/vendors/{vendor}/main_accounts/create", svc.CreateMainAccount)
	h.Add("GetMainAccount", http.MethodGet, "/vendors/{vendor}/main_accounts/{account_id}", svc.GetMainAccount)
	h.Add("UpdateMainAccount", http.MethodPatch, "/main_accounts/{account_id}", svc.UpdateMainAccount)
	h.Add("ListMainAccount", http.MethodPost, "/main_accounts/list", svc.ListMainAccount)

	h.Add("GetMainAccountBasicInfo", http.MethodGet, "/main_accounts/basic_info/{account_id}", svc.GetMainAccountBasicInfo)

	h.Load(cap.WebService)
}

type service struct {
	dao    dao.Set
	cipher cryptography.Crypto
}
