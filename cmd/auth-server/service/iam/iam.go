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

// Package iam ...
package iam

import (
	"hcm/cmd/auth-server/service/capability"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/sys"
	"hcm/pkg/rest"
)

// IAM related operate.
type IAM struct {
	// data service's iamSys api
	ds *dataservice.Client
	// iam client.
	iamSys *sys.Sys
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
}

// NewIAM new iam.
func NewIAM(ds *dataservice.Client, iamSys *sys.Sys, disableAuth bool) (*IAM, error) {
	if ds == nil {
		return nil, errf.New(errf.InvalidParameter, "data client is nil")
	}

	if iamSys == nil {
		return nil, errf.New(errf.InvalidParameter, "iam sys is nil")
	}

	i := &IAM{
		ds:          ds,
		iamSys:      iamSys,
		disableAuth: disableAuth,
	}

	return i, nil
}

// InitIAMService initialize the iam get resource service
func (i *IAM) InitIAMService(c *capability.Capability) {
	h := rest.NewHandler()

	h.Add("PullResource", "POST", "/iam/find/resource", i.PullResource)

	h.Load(c.WebService)
}
