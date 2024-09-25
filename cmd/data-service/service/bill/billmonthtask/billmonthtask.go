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

// Package billmonthtask ...
package billmonthtask

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initialize the bill puller service
func InitService(cap *capability.Capability) {
	svc := &service{
		dao: cap.Dao,
	}
	h := rest.NewHandler()
	h.Add("CreateBillMonthTask", http.MethodPost, "/bills/month_tasks/create", svc.CreateBillMonthTask)
	h.Add("DeleteBillMonthTask", http.MethodDelete, "/bills/month_tasks/batch", svc.DeleteBillMonthTask)
	h.Add("UpdateBillMonthTask", http.MethodPut, "/bills/month_tasks", svc.UpdateBillMonthTask)
	h.Add("ListBillMonthTask", http.MethodGet, "/bills/month_tasks/list", svc.ListBillMonthTask)

	h.Load(cap.WebService)
}

type service struct {
	dao dao.Set
}
