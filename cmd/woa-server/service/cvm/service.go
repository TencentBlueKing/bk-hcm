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

	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/cvm"
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/finops"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		authorizer:   c.Authorizer,
		logics:       c.CvmLogic,
		configLogics: c.ConfigLogics,
		gcLogics:     c.GcLogic,
		finOpsCli:    c.FinOpsCli,
		client:       c.Client,
	}
	h := rest.NewHandler()

	s.initCvmService(h)

	h.Load(c.WebService)

	// 业务下的接口
	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}")
	bizService(bizH, s)

	bizH.Load(c.WebService)
}

type service struct {
	logics       cvm.Logics
	configLogics config.Logics
	authorizer   auth.Authorizer
	gcLogics     gclogics.Logics
	finOpsCli    finops.Client
	client       *client.ClientSet
}

func (s *service) initCvmService(h *rest.Handler) {
	h.Add("CreateApplyOrder", http.MethodPost, "/cvm/create/apply/order", s.CreateApplyOrder)
	h.Add("GetApplyOrderById", http.MethodPost, "/cvm/find/apply/order", s.GetApplyOrderById)
	h.Add("GetApplyOrder", http.MethodPost, "/cvm/findmany/apply/order", s.GetApplyOrder)
	h.Add("GetApplyDevice", http.MethodPost, "/cvm/findmany/apply/device", s.GetApplyDevice)
	h.Add("GetCapacity", http.MethodPost, "/cvm/find/capacity", s.GetCapacity)

	h.Add("GetApplyStatusCfg", http.MethodGet, "/cvm/find/config/apply/status", s.GetApplyStatusCfg)
}

// bizService 业务下的接口
func bizService(h *rest.Handler, s *service) {
	h.Add("GetDeviceLoadUsage", http.MethodPost, "/device/load_usage", s.GetDeviceLoadUsage)
	h.Add("ListDeviceCPUUsageTrend", http.MethodPost, "/device/cpu_usage/trend", s.ListDeviceCPUUsageTrend)
}
