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

package subaccountsecret

import (
	logicaccount "hcm/cmd/cloud-server/logics/account"
	proto "hcm/pkg/api/cloud-server/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ListSubAccountSecret lists sub account secrets under a business (join + extension via data-service).
func (svc *service) ListSubAccountSecret(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.ListSubAccountSecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, errf.NewFromErr(errf.PermissionDenied, err)
	}

	return svc.listBizSubAccountSecretJoinExt(cts.Kit, bizID, vendor, req)
}

// （最大查询范围）查询符合以下条件的三级账号的密钥：三级账号的业务是当前业务，或三级账号所属二级账号的管理业务是当前业务
func (svc *service) listBizSubAccountSecretJoinExt(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *proto.ListSubAccountSecretReq) (interface{}, error) {

	dsReq := &protocloud.SubAccountSecretJoinExtListReq{
		BkBizID: bizID,
		SubAccountSecretFilters: protocloud.SubAccountSecretFilters{
			IDs:                req.IDs,
			Status:             req.Status,
			AccountIDs:         req.AccountIDs,
			SubAccountIDs:      req.SubAccountIDs,
			AccountManagers:    req.AccountManagers,
			SubAccountManagers: req.SubAccountManagers,
			Extension:          req.Extension,
		},
		Page: req.Page,
	}

	switch vendor {
	case enumor.TCloud:
		dsRes, err := svc.client.DataService().TCloud.SubAccountSecret.ListSubAccountSecretJoinExt(kt, dsReq)
		if err != nil {
			return nil, err
		}
		if req.Page.Count {
			return &proto.BizSubAccountSecretJoinExtListResult{Count: dsRes.Count}, nil
		}
		return svc.convertBizSubAccountSecretJoinExtList(kt, bizID, dsRes)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc *service) convertBizSubAccountSecretJoinExtList(kt *kit.Kit, bkBizID int64,
	listResult *protocloud.SubAccountSecretJoinExtListResult) (*proto.BizSubAccountSecretJoinExtListResult, error) {

	if listResult == nil {
		return &proto.BizSubAccountSecretJoinExtListResult{Details: []proto.BizSubAccountSecretJoinExtDetail{}}, nil
	}

	accountIDs := extractAccountIDsFromSubAccountSecretJoinList(listResult.Details)
	_, operableMap, err := logicaccount.BatchBuildOperableAndNameMap(kt, svc.client.DataService(), bkBizID, accountIDs)
	if err != nil {
		return nil, err
	}

	return &proto.BizSubAccountSecretJoinExtListResult{
		Count:   listResult.Count,
		Details: buildBizSubAccountSecretJoinExtDetails(listResult.Details, operableMap),
	}, nil
}

func extractAccountIDsFromSubAccountSecretJoinList(details []protocloud.SubAccountSecretJoinExtDetail) []string {
	result := make([]string, 0, len(details))
	for _, item := range details {
		if item.AccountID == "" {
			continue
		}
		result = append(result, item.AccountID)
	}
	return result
}

func buildBizSubAccountSecretJoinExtDetails(details []protocloud.SubAccountSecretJoinExtDetail,
	operableMap map[string]bool) []proto.BizSubAccountSecretJoinExtDetail {

	result := make([]proto.BizSubAccountSecretJoinExtDetail, 0, len(details))
	for _, item := range details {
		result = append(result, proto.BizSubAccountSecretJoinExtDetail{
			SubAccountSecretJoinExtDetail: item,
			Operable:                      operableMap[item.AccountID],
		})
	}
	return result
}
