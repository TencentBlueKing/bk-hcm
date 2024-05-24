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

package application

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cryptography"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/esb"
)

// InitApplicationService ...
func InitApplicationService(c *capability.Capability, bkHcmUrl string) {
	svc := &applicationSvc{
		client:     c.ApiClient,
		audit:      c.Audit,
		authorizer: c.Authorizer,
		cipher:     c.Cipher,
		itsmCli:    c.ItsmCli,
		esbCli:     c.EsbClient,
		bkHcmUrl:   bkHcmUrl,
	}
	h := rest.NewHandler()
	h.Add("List", "POST", "/applications/list", svc.List)
	h.Add("Get", "GET", "/applications/{application_id}", svc.Get)
	h.Add("Cancel", "PATCH", "/applications/{application_id}/cancel", svc.Cancel)
	h.Add("Approve", "POST", "/applications/approve", svc.Approve)

	h.Add("CreateForAddAccount", "POST", "/applications/types/add_account", svc.CreateForAddAccount)
	h.Add("CreateForCreateCvm", "POST", "/vendors/{vendor}/applications/types/create_cvm", svc.CreateForCreateCvm)
	h.Add("CreateForCreateVpc", "POST", "/vendors/{vendor}/applications/types/create_vpc", svc.CreateForCreateVpc)
	h.Add("CreateForCreateDisk", "POST", "/vendors/{vendor}/applications/types/create_disk", svc.CreateForCreateDisk)
	h.Add("CreateForCreateLB", "POST",
		"/vendors/{vendor}/applications/types/create_load_balancer", svc.CreateForCreateLB)

	initApplicationServiceHooks(svc, h)
	h.Load(c.WebService)
}

type applicationSvc struct {
	client     *client.ClientSet
	audit      audit.Interface
	authorizer auth.Authorizer
	cipher     cryptography.Crypto
	itsmCli    itsm.Client
	esbCli     esb.Client
	bkHcmUrl   string
}

func (a *applicationSvc) getCallbackUrl() string {
	return fmt.Sprintf("%s/api/v1/cloud/applications/approve", strings.TrimRight(a.bkHcmUrl, "/"))
}

func (a *applicationSvc) getHandlerOption(cts *rest.Contexts) *handlers.HandlerOption {
	return &handlers.HandlerOption{
		Cts:       cts,
		Client:    a.client,
		ItsmCli:   a.itsmCli,
		EsbClient: a.esbCli,
		Cipher:    a.cipher,
		Audit:     a.audit,
	}
}

func (a *applicationSvc) getApprovalProcessInfo(
	cts *rest.Contexts, applicationType enumor.ApplicationType,
) (int64, []string, error) {
	// DB中添加4条记录，分别对应add_account、create_cvm、create_vpc、create_disk
	// Note：目前4条记录对应一个itsm流程id，后续如果要使用其它流程可直接修改数据库适配
	// 新增类型只需要增加对应的tye和DB记录
	result, err := a.client.DataService().Global.ApprovalProcess.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApprovalProcessListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "application_type",
						Op:    filter.Equal.Factory(),
						Value: string(applicationType),
					},
				},
			},
			Page: &core.BasePage{
				Count: false,
				Start: 0,
				Limit: 1,
			},
		},
	)
	if err != nil {
		return 0, nil, err
	}
	if result.Details == nil || len(result.Details) != 1 {
		return 0, nil, fmt.Errorf("approval process of [%s] not init", applicationType)
	}

	return result.Details[0].ServiceID, strings.Split(result.Details[0].Managers, ","), nil
}

func (a *applicationSvc) updateStatusWithDetail(
	cts *rest.Contexts, applicationID string, status enumor.ApplicationStatus, deliveryDetail string,
) error {
	req := &dataproto.ApplicationUpdateReq{Status: status}
	if deliveryDetail != "" {
		req.DeliveryDetail = &deliveryDetail
	}
	_, err := a.client.DataService().Global.Application.Update(cts.Kit, applicationID, req)
	return err
}

func (a *applicationSvc) getApplicationBySN(cts *rest.Contexts, sn string) (*dataproto.ApplicationResp, error) {
	// 构造过滤条件，只能查询自己的单据
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "sn", Op: filter.Equal.Factory(), Value: sn},
		},
	}
	// 查询
	resp, err := a.client.DataService().Global.Application.List(
		cts.Kit,
		&dataproto.ApplicationListReq{
			Filter: reqFilter,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found application by sn(%s)", sn)
	}

	return resp.Details[0], nil
}

func (a *applicationSvc) checkApplyResPermission(cts *rest.Contexts, resType meta.ResourceType) error {
	body, err := cts.RequestBody()
	if err != nil {
		logs.Errorf("get request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	bizID := gjson.GetBytes(body, "bk_biz_id").Int()
	if bizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: resType, Action: meta.Apply}, BizID: bizID}
	if err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return err
	}

	return nil
}
