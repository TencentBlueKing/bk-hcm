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

// Package permissiontemplate provides hc-service handlers for permission template cloud operations.
package permissiontemplate

import (
	"net/http"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/rest"
)

// InitService initialize the permission template service.
func InitService(cap *capability.Capability) {
	svc := &service{
		ad: cap.CloudAdaptor,
	}

	h := rest.NewHandler()

	h.Add("TCloudCreateCAMPolicy", http.MethodPost,
		"/vendors/tcloud/permission_templates/cam/create_policy", svc.TCloudCreateCAMPolicy)
	h.Add("TCloudUpdateCAMPolicy", http.MethodPatch,
		"/vendors/tcloud/permission_templates/cam/update_policy", svc.TCloudUpdateCAMPolicy)
	h.Add("TCloudDeleteCAMPolicy", http.MethodDelete,
		"/vendors/tcloud/permission_templates/cam/delete_policy", svc.TCloudDeleteCAMPolicy)

	h.Load(cap.WebService)
}

type service struct {
	ad *cloudadaptor.CloudAdaptorClient
}
