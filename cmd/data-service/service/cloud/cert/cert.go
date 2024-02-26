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

// Package cert 托管证书的DB接口
package cert

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

var svc *certSvc

// InitService initial the cert service
func InitService(cap *capability.Capability) {
	svc = &certSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("ListCert", http.MethodPost, "/certs/list", svc.ListCert)
	h.Add("ListCertExt", http.MethodPost, "/vendors/{vendor}/certs/list", svc.ListCertExt)
	h.Add("CreateCert", http.MethodPost, "/vendors/{vendor}/certs/create", svc.CreateCert)
	h.Add("BatchUpdateCert", http.MethodPatch, "/certs", svc.BatchUpdateCert)
	h.Add("BatchUpdateCertExt", http.MethodPatch, "/vendors/{vendor}/certs", svc.BatchUpdateCertExt)
	h.Add("BatchDeleteCert", http.MethodDelete, "/certs/batch", svc.BatchDeleteCert)

	h.Load(cap.WebService)
}

type certSvc struct {
	dao dao.Set
}
