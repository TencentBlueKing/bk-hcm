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

// Package ressync 资源同步相关接口
package ressync

import (
	"net/http"

	ressync "hcm/cmd/woa-server/logics/res-sync"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cron/core"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		client:       c.Client,
		authorizer:   c.Authorizer,
		resSyncLogic: c.ResSyncLogic,
		tasks:        c.Tasks,
	}
	h := rest.NewHandler()
	h.Path("/res_syncs")

	s.initService(h)
	h.Load(c.WebService)
}

type service struct {
	authorizer   auth.Authorizer
	client       *client.ClientSet
	resSyncLogic ressync.Logics
	tasks        map[enumor.CronTask]core.Task
}

// initService 资源下的接口
func (s *service) initService(h *rest.Handler) {
	h.Add("SyncCapacities", http.MethodPost, s.tasks[enumor.CronTaskSyncDeviceCapacity].GetURL(), s.SyncCapacities)
	h.Add("SyncLeftIPs", http.MethodPost, "/left_ips/sync", s.SyncLeftIPs)
}
