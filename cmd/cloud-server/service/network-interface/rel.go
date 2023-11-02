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
	"hcm/pkg/tools/hooks/handler"
)

// ListNetworkInterfaceExtByCvmID list network interface with extension by cvm id.
func (svc *netSvc) ListNetworkInterfaceExtByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listNICExtByCvmID(cts, handler.ResValidWithAuth)
}

// ListBizNICExtByCvmID list biz network interface with extension by cvm id.
func (svc *netSvc) ListBizNICExtByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listNICExtByCvmID(cts, handler.BizValidWithAuth)
}

func (svc *netSvc) listNICExtByCvmID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	cvmID := cts.Request.PathParameter("cvm_id")
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CvmCloudResType, cvmID)
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

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.NetworkInterface,
		Action: meta.Find, BasicInfo: basicInfo})
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
