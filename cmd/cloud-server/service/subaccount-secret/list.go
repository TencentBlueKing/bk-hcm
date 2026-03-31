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
	proto "hcm/pkg/api/cloud-server/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ListSubAccountSecretJoinExt lists sub account secrets under a business (join + extension via data-service).
func (svc *service) ListSubAccountSecretJoinExt(cts *rest.Contexts) (interface{}, error) {
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

	if err := svc.authorizeBizSubAccountSecretList(cts.Kit, bizID); err != nil {
		return nil, err
	}

	dsReq := &protocloud.SubAccountSecretJoinExtListReq{
		BkBizID: bizID,
		SubAccountSecretFilters: protocloud.SubAccountSecretFilters{
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
		return svc.client.DataService().TCloud.SubAccountSecret.ListSubAccountSecretJoinExt(cts.Kit, dsReq)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc *service) authorizeBizSubAccountSecretList(kt *kit.Kit, bizID int64) error {
	attr := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.SubAccountSecret, Action: meta.Find},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(kt, attr)
	if err != nil {
		return err
	}
	if !authorized {
		return errf.New(errf.PermissionDenied, "permission denied")
	}
	return nil
}
