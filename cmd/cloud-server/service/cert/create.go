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

// Package cert ...
package cert

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	hccert "hcm/pkg/api/hc-service/cert"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateCert create cert.
func (svc *certSvc) CreateCert(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create cert request decode failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	//authRes := meta.ResourceAttribute{Basic: &meta.Basic{
	//	Type: meta.Cert, Action: meta.Create, ResourceID: req.AccountID}}
	//if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
	//	logs.Errorf("create cert auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
	//	return nil, err
	//}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, accID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.createTCloudCert(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *certSvc) createTCloudCert(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(hccert.TCloudCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	publicKey, err := base64.URLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		logs.Errorf("create tcloud cert decode publickey failed, pk: %s, err: %v, rid: %s", req.PublicKey, err, kt.Rid)
		return nil, err
	}
	privateKey, err := base64.URLEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		logs.Errorf("create tcloud cert decode privatekey failed, ik: %s, err: %v, rid: %s", req.PublicKey, err, kt.Rid)
		return nil, err
	}
	req.PublicKey = string(publicKey)
	req.PrivateKey = string(privateKey)

	if err = req.Validate(true); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cert.CreateCert(kt, req)
	if err != nil {
		logs.Errorf("create tcloud cert failed, req: %+v, result: %+v, err: %v, rid: %s", req, result, err, kt.Rid)
		return result, err
	}

	return result, nil
}
