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

package permissiontemplate

import (
	typeaccount "hcm/pkg/adaptor/types/account"
	proto "hcm/pkg/api/hc-service/permission-template"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudCreateCAMPolicy creates a CAM policy for the specified account.
func (svc *service) TCloudCreateCAMPolicy(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.CreateCAMPolicyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudCli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("get tcloud adaptor failed, accountID: %s, err: %v, rid: %s",
			req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typeaccount.TCloudCreatePolicyOption{
		PolicyName:     req.PolicyName,
		PolicyDocument: req.PolicyDocument,
		Description:    req.Description,
	}

	result, err := tcloudCli.CreatePolicy(cts.Kit, opt)
	if err != nil {
		logs.Errorf("create cam policy failed, accountID: %s, policyName: %s, err: %v, rid: %s",
			req.AccountID, req.PolicyName, err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.CreateCAMPolicyResult{PolicyID: result.PolicyID}, nil
}

// TCloudDeleteCAMPolicy deletes a CAM policy for the specified account.
func (svc *service) TCloudDeleteCAMPolicy(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.DeleteCAMPolicyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudCli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("get tcloud adaptor failed, accountID: %s, err: %v, rid: %s",
			req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typeaccount.TCloudDeletePolicyOption{
		PolicyIDs: req.PolicyIDs,
	}

	if err = tcloudCli.DeletePolicy(cts.Kit, opt); err != nil {
		logs.Errorf("delete cam policies failed, accountID: %s, policyIDs: %v, err: %v, rid: %s",
			req.AccountID, req.PolicyIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TCloudUpdateCAMPolicy updates a CAM policy for the specified account.
func (svc *service) TCloudUpdateCAMPolicy(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.UpdateCAMPolicyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudCli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("get tcloud adaptor failed, accountID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typeaccount.TCloudUpdatePolicyOption{
		PolicyID:       req.PolicyID,
		PolicyDocument: req.PolicyDocument,
		Description:    req.Description,
	}

	if err = tcloudCli.UpdatePolicy(cts.Kit, opt); err != nil {
		logs.Errorf("update cam policy failed, accountID: %s, policyID: %d, err: %v, rid: %s",
			req.AccountID, req.PolicyID, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
