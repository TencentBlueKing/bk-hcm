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

// CreateCvm create cvm.
func (svc *cvmSvc) CreateCvm(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create cvm request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm, Action: meta.Create,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create cvm auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.createTCloudCvm(cts.Kit, req.Data)
	case enumor.Aws:
		return svc.createAwsCvm(cts.Kit, req.Data)
	case enumor.HuaWei:
		return svc.createHuaWeiCvm(cts.Kit, req.Data)
	case enumor.Gcp:
		return svc.createGcpCvm(cts.Kit, req.Data)
	case enumor.Azure:
		return svc.createAzureCvm(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *cvmSvc) createAzureCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(cscvm.AzureCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Azure.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(),
		common.ConvAzureCvmCreateReq(req))
	if err != nil {
		logs.Errorf("batch create azure cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *cvmSvc) createHuaWeiCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(cscvm.HuaWeiCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().HuaWei.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(),
		common.ConvHuaWeiCvmCreateReq(req))
	if err != nil {
		logs.Errorf("batch create huawei cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *cvmSvc) createGcpCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(cscvm.GcpCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Gcp.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), common.ConvGcpCvmCreateReq(req))
	if err != nil {
		logs.Errorf("batch create gcp cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *cvmSvc) createAwsCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(cscvm.AwsCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Aws.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), common.ConvAwsCvmCreateReq(req))
	if err != nil {
		logs.Errorf("batch create aws cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *cvmSvc) createTCloudCvm(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(cscvm.TCloudCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(),
		common.ConvTCloudCvmCreateReq(req))
	if err != nil {
		logs.Errorf("batch create tcloud cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}
