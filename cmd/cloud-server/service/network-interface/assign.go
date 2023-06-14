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
	"fmt"

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	datacloudniproto "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
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

	listReq := &core.ListReq{
		Fields: []string{"id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: req.NetworkInterfaceIDs,
				},
				&filter.AtomRule{
					Field: "bk_biz_id",
					Op:    filter.NotEqual.Factory(),
					Value: constant.UnassignedBiz,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.NetworkInterface.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list network_interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) != 0 {
		ids := make([]string, len(result.Details))
		for index, one := range result.Details {
			ids[index] = one.ID
		}
		return nil, fmt.Errorf("network_interface(ids=%v) already assigned", ids)
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.NetworkInterfaceAuditResType, req.NetworkInterfaceIDs,
		req.BkBizID)
	if err != nil {
		logs.Errorf("create network_interface assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	update := &datacloudniproto.NetworkInterfaceCommonInfoBatchUpdateReq{
		IDs:     req.NetworkInterfaceIDs,
		BkBizID: req.BkBizID,
	}
	if err := svc.client.DataService().Global.NetworkInterface.BatchUpdateNICommonInfo(
		cts.Kit.Ctx, cts.Kit.Header(), update); err != nil {
		logs.Errorf("batch update network_interface common info failed, req: %+v, err: %v, rid: %s", req, err,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *netSvc) authorizeNICAssignOp(kt *kit.Kit, ids []string, bizID int64) error {
	// authorize
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.NetworkInterfaceCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(kt.Ctx, kt.Header(), basicInfoReq)
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
