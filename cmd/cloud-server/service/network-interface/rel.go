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

package networkinterface

import (
	"fmt"

	datarelproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// ListNetworkInterfaceExtByCvmID ...
func (svc *netSvc) ListNetworkInterfaceExtByCvmID(cts *rest.Contexts) (interface{}, error) {
	cvmID := cts.Request.PathParameter("cvm_id")
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CvmCloudResType,
		cvmID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if basicInfo.Vendor != vendor {
		return nil, errf.NewFromErr(
			errf.InvalidParameter,
			fmt.Errorf(
				"the vendor(%s) of the cvm does not match the vendor(%s) in url path",
				basicInfo.Vendor,
				vendor,
			),
		)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.NetworkInterface, Action: meta.Find,
		ResourceID: basicInfo.AccountID,
	}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	reqData := &datarelproto.NetworkInterfaceCvmRelWithListReq{CvmIDs: []string{cvmID}}

	switch basicInfo.Vendor {
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.ListNetworkCvmRelWithExt(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.ListNetworkCvmRelWithExt(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.Azure:
		return svc.client.DataService().Azure.ListNetworkCvmRelWithExt(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}
