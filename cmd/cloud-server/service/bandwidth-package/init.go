/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package bandwidthpackage

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	bwpkg "hcm/pkg/api/hc-service/bandwidth-packages"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {
	svc := &bandSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// clb apis in res
	h.Add("QueryBandPackage", http.MethodPost, "/bandwidth_packages/query", svc.QueryBandPackage)

	h.Load(c.WebService)
}

type bandSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// QueryBandPackage query bandwidth package from cloud
func (svc *bandSvc) QueryBandPackage(cts *rest.Contexts) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	req.AccountID = strings.TrimSpace(req.AccountID)
	if len(req.AccountID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("account id is required"))
	}

	// list authorized
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.LoadBalancer, Action: meta.Create,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("list bandwidth package auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	listReq := &cloud.AccountListReq{
		Filter: tools.EqualExpression("id", req.AccountID),
		Page:   core.NewDefaultBasePage(),
	}
	accountResp, err := svc.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}
	if len(accountResp.Details) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("account not found"))
	}
	account := accountResp.Details[0]

	switch account.Vendor {
	case enumor.TCloud:
		return svc.ListTCloudBwPkg(cts.Kit, req.Data)
	default:
		return nil, errors.New("unsupported vendor: " + string(account.Vendor))
	}
}

// ListTCloudBwPkg ...
func (svc *bandSvc) ListTCloudBwPkg(kt *kit.Kit, data json.RawMessage) (any, error) {
	req := new(bwpkg.ListTCloudBwPkgOption)

	if err := json.Unmarshal(data, req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	bandwidthPackage, err := svc.client.HCService().TCloud.BandPkg.ListBandwidthPackage(kt, req)
	if err != nil {
		logs.Errorf("fail to list bandwidth package, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return bandwidthPackage, nil
}
