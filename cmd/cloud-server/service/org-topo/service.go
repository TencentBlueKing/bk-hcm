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

// Package orgtopo ...
package orgtopo

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/usermgr"
)

// InitService initialize service.
func InitService(c *capability.Capability) {
	svc := &orgTopoSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		userMgrCli: c.UserMgrCli,
	}

	h := rest.NewHandler()

	h.Add("SyncOrgTopos", http.MethodPost, "/org_topos/sync", svc.SyncOrgTopos)

	h.Load(c.WebService)
}

type orgTopoSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	state      serviced.State
	userMgrCli usermgr.Client
}

// SyncOrgTopos sync org topos.
func (ots *orgTopoSvc) SyncOrgTopos(cts *rest.Contexts) (any, error) {
	if err := ots.SyncOrgTopo(cts.Kit); err != nil {
		logs.Errorf("sync usermgr org topo resource failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
