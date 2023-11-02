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

package networkinterface

import (
	logicsni "hcm/cmd/cloud-server/logics/network-interface"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AssignNetworkInterfaceToBiz assign network interface to biz.
func (svc *netSvc) AssignNetworkInterfaceToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignNetworkInterfaceToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := svc.authorizeNICAssignOp(cts.Kit, req.NetworkInterfaceIDs, req.BkBizID)
	if err != nil {
		return nil, err
	}

	return nil, logicsni.Assign(cts.Kit, svc.client.DataService(), req.NetworkInterfaceIDs, req.BkBizID, false)
}

func (svc *netSvc) authorizeNICAssignOp(kt *kit.Kit, ids []string, bizID int64) error {
	// authorize
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.NetworkInterfaceCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
	if err != nil {
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.NetworkInterface,
				Action:     meta.Assign,
				ResourceID: info.AccountID,
			},
			BizID: bizID,
		})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}
