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
 */

// Package argstpl ...
package argstpl

import (
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/hc-service/argument-template"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// UpdateBizArgsTpl update biz argument template.
func (svc *argsTplSvc) UpdateBizArgsTpl(cts *rest.Contexts) (interface{}, error) {
	return svc.updateArgsTpl(cts, handler.BizOperateAuth, true)
}

func (svc *argsTplSvc) updateArgsTpl(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler, bizRequired bool) (
	interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update argument template request decode failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	var bkBizID int64 = constant.UnassignedBiz
	var err error
	if bizRequired {
		bkBizID, err = cts.PathParameter("bk_biz_id").Int64()
		if err != nil {
			return nil, err
		}
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.ArgumentTemplate,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("update argument template auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, accID: %s, id: %s, err: %v, rid: %s",
			req.AccountID, id, err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.updateTCloudArgumentTemplate(cts.Kit, req.Data, id, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *argsTplSvc) updateTCloudArgumentTemplate(kt *kit.Kit, body json.RawMessage, id string, bkBizID int64) (
	interface{}, error) {

	req := new(hcargstpl.TCloudUpdateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req.BkBizID = bkBizID
	err := svc.client.HCService().TCloud.ArgsTpl.UpdateArgsTpl(kt, id, req)
	if err != nil {
		logs.Errorf("update tcloud argument template failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	return nil, nil
}
