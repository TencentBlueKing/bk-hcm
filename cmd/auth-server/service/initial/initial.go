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

// Package initial ...
package initial

import (
	"hcm/cmd/auth-server/service/capability"
	authserver "hcm/pkg/api/auth-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Initial iam init related operate.
type Initial struct {
	// iam client.
	iamSys *sys.Sys
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
}

// NewInitial new initial.
func NewInitial(iamSys *sys.Sys, disableAuth bool) (*Initial, error) {
	if iamSys == nil {
		return nil, errf.New(errf.InvalidParameter, "iam sys is nil")
	}

	i := &Initial{
		iamSys:      iamSys,
		disableAuth: disableAuth,
	}

	return i, nil
}

// InitInitialService initialize the iam Initial service
func (i *Initial) InitInitialService(c *capability.Capability) {
	h := rest.NewHandler()

	h.Add("InitAuthCenter", "POST", "/init/authcenter", i.InitAuthCenter)

	h.Load(c.WebService)
}

// InitAuthCenter init auth center's auth model.
func (i *Initial) InitAuthCenter(cts *rest.Contexts) (interface{}, error) {
	req := new(authserver.InitAuthCenterReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	// if auth is disabled, returns error if user wants to init auth center
	if i.disableAuth {
		logs.Errorf("authorize function is disabled, can not init auth center, rid: %s", cts.Kit.Rid)
		return nil, errf.New(errf.Aborted, "authorize function is disabled, can not init auth center.")
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("request param validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := i.iamSys.Register(cts.Kit, req.Host); err != nil {
		logs.Errorf("register to iam failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
