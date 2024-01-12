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

// Package argstpl 参数模版的DB接口
package argstpl

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

var svc *argsTplSvc

// InitService initial the argument template service
func InitService(cap *capability.Capability) {
	svc = &argsTplSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("ListArgsTpl", http.MethodPost, "/argument_templates/list", svc.ListArgsTpl)
	h.Add("ListArgsTplExt", http.MethodPost, "/vendors/{vendor}/argument_templates/list", svc.ListArgsTplExt)
	h.Add("CreateArgsTpl", http.MethodPost, "/vendors/{vendor}/argument_templates/create", svc.CreateArgsTpl)
	h.Add("BatchUpdateArgsTpl", http.MethodPut, "/argument_templates", svc.BatchUpdateArgsTpl)
	h.Add("BatchDeleteArgsTpl", http.MethodDelete, "/argument_templates/batch", svc.BatchDeleteArgsTpl)

	h.Load(cap.WebService)
}

type argsTplSvc struct {
	dao dao.Set
}
