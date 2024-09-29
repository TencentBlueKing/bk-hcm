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

// Package capability ...
package capability

import (
	"hcm/cmd/cloud-server/logics"
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/client"
	"hcm/pkg/cryptography"
	"hcm/pkg/iam/auth"
	"hcm/pkg/thirdparty/api-gateway/bkbase"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/esb"

	"github.com/emicklei/go-restful/v3"
)

// Capability defines the service's capability
type Capability struct {
	WebService *restful.WebService
	ApiClient  *client.ClientSet
	Authorizer auth.Authorizer
	Audit      audit.Interface
	Cipher     cryptography.Crypto
	EsbClient  esb.Client
	Logics     *logics.Logics
	ItsmCli    itsm.Client
	BKBaseCli  bkbase.Client
	CmsiCli    cmsi.Client
}
