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
	"fmt"

	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/TencentBlueKing/gopkg/conv"
)

func canListAccountExtension(appCode string) error {
	// TODO: 校验App来源, 这里只校验了非Web来请求，需要改造从配置文件读取，允许访问该接口的AppCode白名单（目前暂时可以借助APIGateway的应用认证白名单）
	if appCode == "hcm-web-server" {
		return fmt.Errorf("app[%s] has no permission to list account with extension", appCode)
	}

	return nil
}

// ListWithExtension 该接口返回了Extension，不包括SecretKey，只提供给安全使用
func (a *accountSvc) ListWithExtension(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountListWithExtReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验应用权限
	if err := canListAccountExtension(cts.Kit.AppCode); err != nil {
		return nil, err
	}

	// 校验用户是否有查看权限，有权限的ID列表
	accountIDs, isAny, err := a.listAuthorized(cts, meta.Find, meta.Account)
	if err != nil {
		return nil, err
	}
	// 无任何账号权限
	if len(accountIDs) == 0 && !isAny {
		return []map[string]interface{}{}, nil
	}

	// 构造权限过滤条件
	var reqFilter *filter.Expression
	if isAny {
		reqFilter = req.Filter
	} else {
		reqFilter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: accountIDs},
			},
		}
		// 加上请求里过滤条件
		if req.Filter != nil && !req.Filter.IsEmpty() {
			reqFilter.Rules = append(reqFilter.Rules, req.Filter)
		}
	}

	resp, err := a.client.DataService().Global.Account.ListWithExtension(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
	if err != nil || resp == nil {
		return nil, err
	}

	// 去除SecretKey
	if resp.Details != nil {
		for _, detail := range resp.Details {
			secretKeyField := detail.Vendor.GetSecretField()
			// 存在SecretKey则删除
			// Note: 资源账号、安全审计账号必然存在，登录账号大部分时候是不存在的
			if _, ok := detail.Extension[secretKeyField]; ok {
				delete(detail.Extension, secretKeyField)
			}
		}
	}

	return resp, nil
}

// ListSecretKey  批量获取Secret，只给安全提供
func (a *accountSvc) ListSecretKey(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListSecretKeyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验应用权限
	if err := canListAccountExtension(cts.Kit.AppCode); err != nil {
		return nil, err
	}

	// 校验用户是否有查询Secret的权限
	if err := a.checkPermissions(cts, meta.KeyAccess, req.IDs); err != nil {
		return nil, err
	}

	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs},
			filter.AtomRule{Field: "type", Op: filter.NotEqual.Factory(), Value: enumor.ResourceAccount},
		},
	}

	// 查询账号信息，带Extension的
	resp, err := a.client.DataService().Global.Account.ListWithExtension(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil || resp == nil || resp.Details == nil {
		return nil, err
	}

	secretKeyData := make([]proto.SecretKeyData, 0, len(resp.Details))
	for _, detail := range resp.Details {
		// 根据vendor获取SecretKey
		secretKeyField := detail.Vendor.GetSecretField()
		secretKeyData = append(secretKeyData, proto.SecretKeyData{
			ID:        detail.ID,
			SecretKey: conv.ToString(detail.Extension[secretKeyField]),
		})
	}

	return secretKeyData, nil
}
