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

package subaccount

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// GetSubAccount get sub account.
func (svc *service) GetSubAccount(cts *rest.Contexts) (interface{}, error) {
	return svc.getSubAccount(cts, handler.ResOperateAuth)
}

func (svc *service) getSubAccount(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SubAccountCloudResType, id, "id", "vendor")
	if err != nil {
		logs.Errorf("request ds to get resource basic info failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SubAccount,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("get sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	var result interface{}
	// get vpc detail info
	switch basicInfo.Vendor {
	case enumor.TCloud:
		result, err = svc.client.DataService().TCloud.SubAccount.Get(cts.Kit, id)
	case enumor.Aws:
		result, err = svc.client.DataService().Aws.SubAccount.Get(cts.Kit, id)
	case enumor.Gcp:
		result, err = svc.client.DataService().Gcp.SubAccount.Get(cts.Kit, id)
	case enumor.HuaWei:
		result, err = svc.client.DataService().HuaWei.SubAccount.Get(cts.Kit, id)
	case enumor.Azure:
		result, err = svc.client.DataService().Azure.SubAccount.Get(cts.Kit, id)
	default:
		return nil, fmt.Errorf("vendor: %s not support", basicInfo.Vendor)
	}
	if err != nil {
		logs.Errorf("request ds to get sub account failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
