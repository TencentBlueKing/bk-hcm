/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package systeminit

import (
	"fmt"
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/cloud-server/system-init"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// InitSystemInitService initialize the system init service.
func InitSystemInitService(c *capability.Capability) {
	svc := &systemInitSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("OtherAccountInit", http.MethodPost, "/system-init/accounts/other/init", svc.OtherAccountInit)

	h.Load(c.WebService)
}

type systemInitSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// OtherAccountInit 查找是否存在vendor为other的用户，若有则返回，没有则创建
func (s *systemInitSvc) OtherAccountInit(cts *rest.Contexts) (any, error) {

	if err := s.checkAdmin(cts); err != nil {
		return nil, fmt.Errorf("check admin failed, err: %v", err)
	}

	// 查找是否存在vendor为other的用户，若有则返回，没有则创建
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("vendor", "other"),
		Page:   core.NewDefaultBasePage(),
	}
	accResp, err := s.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("fail to list other vendor account, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(accResp.Details) > 0 {
		return apisysteminit.OtherAccountInitResp{ExistsAccountID: accResp.Details[0].ID}, nil
	}
	// 创建other vendor用户
	createReq := &protocloud.AccountCreateReq[protocloud.OtherAccountExtensionCreateReq]{
		Name:     "内置账号",
		Managers: []string{"admin"},
		Type:     enumor.ResourceAccount,
		Site:     enumor.InternationalSite,
		Memo:     cvt.ValToPtr("内置账号"),
		Extension: &protocloud.OtherAccountExtensionCreateReq{
			CloudID:     "other",
			CloudSecKey: "",
		},
		BkBizIDs: []int64{constant.AttachedAllBiz},
	}
	createResp, err := s.client.DataService().Other.Account.Create(cts.Kit, createReq)
	if err != nil {
		logs.Errorf("fail to create other vendor account, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return apisysteminit.OtherAccountInitResp{CreatedAccountID: createResp.ID}, nil
}

func (s *systemInitSvc) checkAdmin(cts *rest.Contexts) error {
	// TODO 鉴权 仅管理员、后台可以调用
	if cts.Kit.User != "admin" {
		return fmt.Errorf("only admin can call this api")
	}
	return nil
}
