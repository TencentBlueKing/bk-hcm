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

package account

import (
	typeaccount "hcm/pkg/adaptor/types/account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// TCloudAttachUserPolicies attaches multiple CAM policies to a TCloud sub-user.
func (svc *service) TCloudAttachUserPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudAttachUserPoliciesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, policyID := range req.PolicyIDs {
		if err = client.AttachUserPolicy(cts.Kit, &typeaccount.TCloudAttachUserPolicyOption{TargetUin: req.TargetUin,
			PolicyId: policyID,
		}); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// TCloudDetachUserPolicies detaches multiple CAM policies from a TCloud sub-user.
func (svc *service) TCloudDetachUserPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDetachUserPoliciesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, policyID := range req.PolicyIDs {
		if err = client.DetachUserPolicy(cts.Kit, &typeaccount.TCloudDetachUserPolicyOption{
			DetachUin: req.DetachUin,
			PolicyId:  policyID,
		}); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
