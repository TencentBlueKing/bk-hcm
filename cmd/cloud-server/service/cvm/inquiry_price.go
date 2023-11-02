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

package cvm

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	cloudserver "hcm/pkg/api/cloud-server"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InquiryPriceCvm inquiry price cvm.
func (svc *cvmSvc) InquiryPriceCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("inquiry price cvm request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm, Action: meta.Create,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("inquiry price cvm auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.inquiryPriceTCloudCvm(cts.Kit, req.Data)
	case enumor.HuaWei:
		return svc.inquiryPriceHuaWeiCvm(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *cvmSvc) inquiryPriceTCloudCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(cscvm.TCloudCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cvm.InquiryPrice(kt, common.ConvTCloudCvmCreateReq(req))
	if err != nil {
		logs.Errorf("inquiry price tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *cvmSvc) inquiryPriceHuaWeiCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(cscvm.HuaWeiCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().HuaWei.Cvm.InquiryPrice(kt, common.ConvHuaWeiCvmCreateReq(req))
	if err != nil {
		logs.Errorf("inquiry price huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return result, err
	}

	return result, nil
}
