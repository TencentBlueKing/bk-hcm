///*
// * TencentBlueKing is pleased to support the open source community by making
// * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
// * Copyright (C) 2022 THL A29 Limited,
// * a Tencent company. All rights reserved.
// * Licensed under the MIT License (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at http://opensource.org/licenses/MIT
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on
// * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// * either express or implied. See the License for the
// * specific language governing permissions and limitations under the License.
// *
// * We undertake not to change the open source license (MIT license) applicable
// *
// * to the current version of the project delivered to anyone in the future.
// */

package cloud

import (
	"fmt"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/protocol/base"
	protocloud "hcm/pkg/api/protocol/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao"
	"hcm/pkg/models/cloud"
	"hcm/pkg/rest"
)

// InitAccountBizRelService ...
func InitAccountBizRelService(cap *capability.Capability) {
	svc := &accountBizRelSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	// 采用类似 iac 接口的结构简化处理, 不遵循 RESTful 风格
	h.Add("CreateAccountBizRel", "POST", "/cloud/account_biz_rels/create/", svc.Create)
	h.Add("UpdateAccountBizRels", "POST", "/cloud/account_biz_rels/update/", svc.Update)
	h.Add("ListAccountBizRels", "POST", "/cloud/account_biz_rels/list/", svc.List)
	h.Add("DeleteAccountBizRels", "POST", "/cloud/account_biz_rels/delete/", svc.Delete)

	h.Load(cap.WebService)
}

type accountBizRelSvc struct {
	dao dao.Set
}

// Create ...
func (svc *accountBizRelSvc) Create(cts *rest.Contexts) (interface{}, error) {
	reqData := new(protocloud.CreateAccountBizRelReq)

	if err := cts.DecodeInto(reqData); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	rel := &cloud.AccountBizRel{
		BkBizID:   reqData.BkBizID,
		AccountID: reqData.AccountID,
	}
	id, err := rel.Create(cts.Kit)
	if err != nil {
		return nil, fmt.Errorf("create account_biz_rel failed, err: %v", err)
	}

	return &base.CreateResult{ID: id}, nil
}

// Update ...
func (svc *accountBizRelSvc) Update(cts *rest.Contexts) (interface{}, error) {
	reqData := new(protocloud.UpdateAccountBizRelsReq)

	if err := cts.DecodeInto(reqData); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	rel := &cloud.AccountBizRel{
		BkBizID: reqData.BkBizID,
	}

	err := rel.Update(cts.Kit, &reqData.FilterExpr, validator.ExtractValidFields(reqData))

	return nil, err
}

// List ...
func (svc *accountBizRelSvc) List(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.ListAccountBizRelsReq)
	if err := cts.DecodeInto(reqData); err != nil {
		return nil, err
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}
	mData, err := svc.dao.CloudAccountBizRel().List(cts.Kit, reqData.ToListOption())
	if err != nil {
		return nil, err
	}

	var details []cloud.AccountBizRelData
	for _, m := range mData {
		details = append(details, *cloud.NewAccountBizRelData(m))
	}

	return &cloud.ListAccountBizRelsResult{Details: details}, nil
}

// Delete ...
func (svc *accountBizRelSvc) Delete(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.DeleteAccountBizRelsReq)
	if err := cts.DecodeInto(reqData); err != nil {
		return nil, err
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	err := svc.dao.CloudAccountBizRel().Delete(cts.Kit, &reqData.FilterExpr, new(tablecloud.AccountBizRelModel))

	return nil, err
}
