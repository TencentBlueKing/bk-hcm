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

package account

import (
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(c *capability.Capability) {
	svc := &accountSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("Create", "POST", "/accounts/create", svc.Create)
	h.Add("Check", "POST", "/accounts/check", svc.Check)
	h.Add("CheckByID", "POST", "/accounts/{account_id}/check", svc.CheckByID)
	h.Add("List", "POST", "/accounts/list", svc.List)
	h.Add("Get", "GET", "/accounts/{account_id}", svc.Get)
	h.Add("Update", "PATCH", "/accounts/{account_id}", svc.Update)

	h.Load(c.WebService)
}

type accountSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

func (a *accountSvc) Create(cts *rest.Contexts) (interface{}, error) {
	return map[string]interface{}{
		"id": 1,
	}, nil
}

func (a *accountSvc) Check(cts *rest.Contexts) (interface{}, error) {
	return nil, nil
}

func (a *accountSvc) CheckByID(cts *rest.Contexts) (interface{}, error) {
	return nil, nil
}

type PageReq struct {
	Page *types.BasePage `json:"page" validate:"required"`
}

func (p *PageReq) Validate() error {
	return validator.Validate.Struct(p)
}

func (a *accountSvc) List(cts *rest.Contexts) (interface{}, error) {

	req := new(PageReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page.Count {
		return map[string]interface{}{
			"count":   1,
			"details": []interface{}{},
		}, nil
	}

	details := []interface{}{
		map[string]interface{}{
			"id":     1,
			"vendor": "tcloud", // 云厂商，枚举值有：tcloud 、aws、azure、gcp、huawei
			"spec": map[string]interface{}{
				"name":          "qcloud-account",
				"type":          "resource",                         // resource表示资源账号，register表示登记账号
				"managers":      []string{"jiananzhang", "jamesge"}, // 负责人
				"department_id": 1,                                  // 组织架构，选择的部门ID
				"price":         500.01,                             // 余额
				"price_unit":    "",                                 // 余额单位，可能是美元、人民币等
				"memo":          "测试账号",                             // 备注
			},
		},
	}
	return map[string]interface{}{
		"count":   0,
		"details": details,
	}, nil
}

// Get create account with options
func (a *accountSvc) Get(cts *rest.Contexts) (interface{}, error) {
	return map[string]interface{}{
		"id":     1,
		"vendor": "tcloud", // 云厂商，枚举值有：tcloud 、aws、azure、gcp、huawei
		"spec": map[string]interface{}{
			"name":          "qcloud-account",
			"type":          "resource",                         // resource表示资源账号，register表示登记账号
			"managers":      []string{"jiananzhang", "jamesge"}, // 负责人
			"department_id": 1,                                  // 组织架构，选择的部门ID
			"price":         500.01,                             // 余额
			"price_unit":    "",                                 // 余额单位，可能是美元、人民币等
			"memo":          "测试账号",                             // 备注
		},
		"extension": map[string]interface{}{
			"cloud_main_account": "112224234",             // 主账号
			"cloud_sub_account":  "121435343333",          // 子账号
			"cloud_secret_id":    "AIDDy324DY23423424hdj", // SecretID
			"cloud_secret_key":   "AKEYdsfewerwerewrwe",   // SecretKey
		},
		"revision": map[string]interface{}{
			"creator":   "tom",
			"reviser":   "tom",
			"create_at": "2019-07-29 11:57:20",
			"update_at": "2019-07-29 11:57:20",
		},
	}, nil
}

func (a *accountSvc) Update(cts *rest.Contexts) (interface{}, error) {
	return nil, nil
}

// Create defines to create account with options
// func (a *accountSvc) Create(cts *rest.Contexts) (interface{}, error) {
// 	req := new(protocloudserver.CreateAccountReq)
// 	if err := cts.DecodeInto(req); err != nil {
// 		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
// 	}
//
// 	if err := req.Validate(); err != nil {
// 		return nil, errf.Newf(errf.InvalidParameter, err.Error())
// 	}
//
// 	// 校验权限
// 	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Create}}
// 	err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// 转换数据结构，调用DataService
// 	createAccountReq := &protodataservice.CreateAccountReq{
// 		Vendor:     req.Vendor,
// 		Spec:       req.Spec,
// 		Extension:  req.Extension,
// 		Attachment: req.Attachment,
// 	}
// 	resp, err := a.client.DataService().CloudAccount().Create(cts.Kit.Ctx, cts.Kit.Header(), createAccountReq)
// 	if err != nil {
// 		return nil, fmt.Errorf("create account failed, err: %v", err)
// 	}
//
// 	return &core.CreateResult{ID: resp.ID}, nil
// }
//
// // List accounts
// func (a *accountSvc) List(cts *rest.Contexts) (interface{}, error) {
// 	req := new(protocloudserver.ListAccountReq)
// 	if err := cts.DecodeInto(req); err != nil {
// 		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
// 	}
//
// 	if err := req.Validate(); err != nil {
// 		return nil, errf.Newf(errf.InvalidParameter, err.Error())
// 	}
//
// 	// TODO: 校验权限，这里应该只能拉取用户有权的账号列表
//
// 	// 转换数据结构，调用DataService
// 	listAccountReq := &protodataservice.ListAccountReq{
// 		Filter: req.Filter,
// 		Page:   req.Page,
// 	}
// 	resp, err := a.client.DataService().CloudAccount().List(cts.Kit.Ctx, cts.Kit.Header(), listAccountReq)
// 	if err != nil {
// 		return nil, fmt.Errorf("create account failed, err: %v", err)
// 	}
//
// 	return &protocloudserver.ListAccountResult{
// 		Count:   resp.Count,
// 		Details: resp.Details,
// 	}, nil
// 	return nil, nil
// }

// Check 根据云账号信息校验
// func (a *accountSvc) Check(cts *rest.Contexts) error {
// req := new(protocloudserver.CheckAccountReq)
// if err := cts.DecodeInto(req); err != nil {
// 	return errf.New(errf.DecodeRequestFailed, err.Error())
// }
// reqExtension := req.Extension
//
// checkReq := &protohcservice.AccountCheckReq{Vendor: req.Vendor}
// switch req.Vendor {
// case enumor.TCloud:
// 	checkReq.Secret = &types.Secret{
// 		TCloud: &types.BaseSecret{
// 			ID:  reqExtension.TCloud.Secret.Cid,
// 			Key: reqExtension.TCloud.Secret.Key,
// 		},
// 	}
// 	checkReq.AccountInfo = &types.AccountCheckOption{
// 		Tcloud: &types.TcloudAccountInfo{
// 			AccountCid:     reqExtension.TCloud.SubAccountCid,
// 			MainAccountCid: reqExtension.TCloud.MainAccountCid,
// 		},
// 	}
// case enumor.AWS:
// 	checkReq.Secret = &types.Secret{
// 		Aws: &types.BaseSecret{
// 			ID:  reqExtension.Aws.Secret.Cid,
// 			Key: reqExtension.Aws.Secret.Key,
// 		},
// 	}
// 	checkReq.AccountInfo = &types.AccountCheckOption{
// 		Aws: &types.AwsAccountInfo{
// 			AccountCid:  reqExtension.Aws.AccountCid,
// 			IamUserName: reqExtension.Aws.IamUserName,
// 		},
// 	}
// 	// TODO: 待多云异构如何处理讨论完后补齐其他情况
// }
//
// err := a.client.HCService().Account().Check(cts.Kit.Ctx, cts.Kit.Header(), checkReq)
// if err != nil {
// 	return fmt.Errorf("check account failed, err: %v", err)
// }
//
// return nil
// 	return nil
// }
