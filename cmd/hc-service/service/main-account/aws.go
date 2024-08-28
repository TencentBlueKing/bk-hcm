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

// Package mainaccount Package service defines service.
package mainaccount

import (
	proto "hcm/pkg/api/hc-service/main-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AwsCreateMainAccount 创建aws账号
func (s *service) AwsCreateMainAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.CreateAwsMainAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1、获取一级账号aws Client
	client, err := s.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	// 2、在组织中创建AWS账号
	resp, err := client.CreateAccount(cts.Kit, req)
	if err != nil {
		logs.Errorf("fail to create aws main account, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("create aws account succeed, id: %s, name: %s, email: %s,  rid: %s",
		req.Email, req.CloudAccountName, resp.AccountID, cts.Kit.Rid)

	return resp, nil
}
