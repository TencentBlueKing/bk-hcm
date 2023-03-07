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

package cvm

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitCvmService initialize the cvm service.
func InitCvmService(c *capability.Capability) {
	svc := &cvmSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("GetCvm", http.MethodGet, "/cvms/{id}", svc.GetCvm)
	h.Add("ListCvmExt", http.MethodPost, "/cvms/list", svc.ListCvm)
	h.Add("BatchDeleteCvm", http.MethodDelete, "/cvms/batch", svc.BatchDeleteCvm)
	h.Add("AssignCvmToBiz", http.MethodPost, "/cvms/assign/bizs", svc.AssignCvmToBiz)
	h.Add("BatchStartCvm", http.MethodPost, "/cvms/batch/start", svc.BatchStartCvm)
	h.Add("BatchStopCvm", http.MethodPost, "/cvms/batch/stop", svc.BatchStopCvm)
	h.Add("BatchRebootCvm", http.MethodPost, "/cvms/batch/reboot", svc.BatchRebootCvm)

	h.Load(c.WebService)
}

type cvmSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

func cvmClassificationByVendor(infoMap map[string]types.CloudResourceBasicInfo) map[enumor.Vendor][]types.
	CloudResourceBasicInfo {

	cvmVendorMap := make(map[enumor.Vendor][]types.CloudResourceBasicInfo, 0)
	for _, info := range infoMap {
		if _, exist := cvmVendorMap[info.Vendor]; !exist {
			cvmVendorMap[info.Vendor] = []types.CloudResourceBasicInfo{info}
			continue
		}

		cvmVendorMap[info.Vendor] = append(cvmVendorMap[info.Vendor], info)
	}

	return cvmVendorMap
}

func cvmClassification(infoMap []types.CloudResourceBasicInfo,
) map[string] /*account_id*/ map[string] /*regin*/ []string {

	cvmMap := make(map[string]map[string][]string, 0)
	for _, one := range infoMap {
		if _, exist := cvmMap[one.AccountID]; !exist {
			cvmMap[one.AccountID] = map[string][]string{
				one.Region: {one.ID},
			}

			continue
		}

		if _, exist := cvmMap[one.AccountID][one.Region]; !exist {
			cvmMap[one.AccountID][one.Region] = []string{one.ID}

			continue
		}

		cvmMap[one.AccountID][one.Region] = append(cvmMap[one.AccountID][one.Region], one.ID)
	}

	return cvmMap
}
