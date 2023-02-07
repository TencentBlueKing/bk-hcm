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

package audit

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService initialize the audit service.
func InitService(c *capability.Capability) {
	svc := &svc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("GetAudit", http.MethodGet, "/audits/{id}", svc.GetAudit)
	h.Add("ListAudit", http.MethodPost, "/audits/list", svc.ListAudit)

	h.Load(c.WebService)
}

type svc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// GetAudit get audit.
func (svc svc) GetAudit(cts *rest.Contexts) (interface{}, error) {
	id, err := cts.PathParameter("id").Uint64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	audit, err := svc.client.DataService().Global.Audit.GetAudit(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("get audit failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return audit, nil
}

// ListAudit list audit.
func (svc svc) ListAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AuditListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.Audit.ListAudit(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}
