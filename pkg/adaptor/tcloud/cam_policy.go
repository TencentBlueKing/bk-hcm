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

package tcloud

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
)

// CreatePolicy creates a CAM policy.
// reference: https://cloud.tencent.com/document/product/598/34578
func (t *TCloudImpl) CreatePolicy(kt *kit.Kit, opt *typeaccount.TCloudCreatePolicyOption) (
	*typeaccount.TCloudCreatePolicyResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewCreatePolicyRequest()
	req.PolicyName = converter.ValToPtr(opt.PolicyName)
	req.PolicyDocument = converter.ValToPtr(opt.PolicyDocument)
	if opt.Description != "" {
		req.Description = converter.ValToPtr(opt.Description)
	}

	resp, err := client.CreatePolicyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create cam policy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &typeaccount.TCloudCreatePolicyResult{
		PolicyID: converter.PtrToVal(resp.Response.PolicyId),
	}, nil
}

// UpdatePolicy updates a CAM policy's document.
// reference: https://cloud.tencent.com/document/product/598/34569
func (t *TCloudImpl) UpdatePolicy(kt *kit.Kit, opt *typeaccount.TCloudUpdatePolicyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewUpdatePolicyRequest()
	req.PolicyId = converter.ValToPtr(opt.PolicyID)
	if opt.PolicyDocument != nil {
		req.PolicyDocument = opt.PolicyDocument
	}
	if opt.Description != nil {
		req.Description = opt.Description
	}

	if _, err = client.UpdatePolicyWithContext(kt.Ctx, req); err != nil {
		logs.Errorf("update cam policy failed, policyID: %d, err: %v, rid: %s", opt.PolicyID, err, kt.Rid)
		return err
	}

	return nil
}

// ListPolicies 分页拉取 CAM 策略列表（含预设策略和自定义策略）。
// reference: https://cloud.tencent.com/document/product/598/34570
func (t *TCloudImpl) ListPolicies(kt *kit.Kit, opt *typeaccount.TCloudListPoliciesOption) (
	[]typeaccount.TCloudPolicyItem, uint64, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(opt.Region)
	if err != nil {
		return nil, 0, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewListPoliciesRequest()
	req.Page = converter.ValToPtr(opt.Page)
	req.Rp = converter.ValToPtr(opt.Rp)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	resp, err := LimitExceededErrRetry(client.ListPoliciesWithContext, kt, req, constant.TCloudClientErrRetryTimes,
		rangeMS)
	if err != nil {
		logs.Errorf("fail to get policy from tcloud after retry, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	total := converter.PtrToVal(resp.Response.TotalNum)

	items := make([]typeaccount.TCloudPolicyItem, 0, len(resp.Response.List))
	for _, p := range resp.Response.List {
		item := typeaccount.TCloudPolicyItem{
			PolicyID:    converter.PtrToVal(p.PolicyId),
			PolicyName:  converter.PtrToVal(p.PolicyName),
			Description: converter.PtrToVal(p.Description),
			PolicyType:  enumor.TCloudPolicyType(converter.PtrToVal(p.Type)),
			CreateTime:  converter.PtrToVal(p.AddTime),
		}
		items = append(items, item)
	}

	return items, total, nil
}

// ListAttachedUserAllPolicies 获取子用户绑定的所有策略列表（支持分页和限流重试）。
// reference: https://cloud.tencent.com/document/product/598/67728
func (t *TCloudImpl) ListAttachedUserAllPolicies(kt *kit.Kit, opt *typeaccount.TCloudListAttachedUserAllPoliciesOption,
) (*typeaccount.TCloudListAttachedUserAllPoliciesResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewListAttachedUserAllPoliciesRequest()
	req.TargetUin = converter.ValToPtr(opt.TargetUin)
	req.Page = converter.ValToPtr(opt.Page)
	req.Rp = converter.ValToPtr(opt.Rp)
	req.AttachType = opt.AttachType

	rangeMS := [2]uint{constant.MinRetryInterval, constant.MaxRetryInterval}
	resp, err := LimitExceededErrRetry(client.ListAttachedUserAllPoliciesWithContext, kt, req,
		constant.TCloudClientErrRetryTimes, rangeMS)
	if err != nil {
		logs.Errorf("fail to list attached user all policies from tcloud after retry, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	totalNum := converter.PtrToVal(resp.Response.TotalNum)
	items := make([]typeaccount.TCloudAttachedPolicy, 0, len(resp.Response.PolicyList))
	for _, p := range resp.Response.PolicyList {
		item := typeaccount.TCloudAttachedPolicy{
			PolicyID:     converter.PtrToVal(p.PolicyId),
			PolicyName:   converter.PtrToVal(p.PolicyName),
			Description:  converter.PtrToVal(p.Description),
			AddTime:      converter.PtrToVal(p.AddTime),
			StrategyType: converter.PtrToVal(p.StrategyType),
		}
		items = append(items, item)
	}

	return &typeaccount.TCloudListAttachedUserAllPoliciesResult{
		PolicyList: items,
		TotalNum:   totalNum,
	}, nil
}

// AttachUserPolicy 为子用户绑定策略。
// reference: https://cloud.tencent.com/document/product/598/34579
func (t *TCloudImpl) AttachUserPolicy(kt *kit.Kit, opt *typeaccount.TCloudAttachUserPolicyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewAttachUserPolicyRequest()
	req.AttachUin = converter.ValToPtr(opt.TargetUin)
	req.PolicyId = converter.ValToPtr(opt.PolicyId)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	if _, err = LimitExceededErrRetry(client.AttachUserPolicyWithContext, kt, req, constant.TCloudClientErrRetryTimes,
		rangeMS); err != nil {
		logs.Errorf("fail to attach user policy after retry, targetUin: %d, policyId: %d, err: %v, rid: %s",
			opt.TargetUin, opt.PolicyId, err, kt.Rid)
		return err
	}

	return nil
}

// DetachUserPolicy 解除绑定到子用户的策略。
// reference: https://cloud.tencent.com/document/product/598/34575
func (t *TCloudImpl) DetachUserPolicy(kt *kit.Kit, opt *typeaccount.TCloudDetachUserPolicyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewDetachUserPolicyRequest()
	req.DetachUin = converter.ValToPtr(opt.DetachUin)
	req.PolicyId = converter.ValToPtr(opt.PolicyId)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	if _, err = LimitExceededErrRetry(client.DetachUserPolicyWithContext, kt, req, constant.TCloudClientErrRetryTimes,
		rangeMS); err != nil {
		logs.Errorf("fail to detach user policy after retry, detachUin: %d, policyId: %d, err: %v, rid: %s",
			opt.DetachUin, opt.PolicyId, err, kt.Rid)
		return err
	}

	return nil
}

// GetPolicyDetail 获取单个CAM策略的完整详情（含 PolicyDocument）。
// reference: https://cloud.tencent.com/document/product/598/34574
func (t *TCloudImpl) GetPolicyDetail(kt *kit.Kit, opt *typeaccount.TCloudGetPolicyDetailOption) (
	*typeaccount.TCloudPolicyDetail, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewGetPolicyRequest()
	req.PolicyId = converter.ValToPtr(opt.PolicyID)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	resp, err := LimitExceededErrRetry(client.GetPolicyWithContext, kt, req, constant.TCloudClientErrRetryTimes,
		rangeMS)
	if err != nil {
		logs.Errorf("fail to get policy from tcloud after retry, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	detail := &typeaccount.TCloudPolicyDetail{
		PolicyID:       opt.PolicyID,
		PolicyName:     converter.PtrToVal(resp.Response.PolicyName),
		PolicyDocument: converter.PtrToVal(resp.Response.PolicyDocument),
		Description:    converter.PtrToVal(resp.Response.Description),
		PolicyType:     enumor.TCloudPolicyType(converter.PtrToVal(resp.Response.Type)),
		CreateTime:     converter.PtrToVal(resp.Response.AddTime),
	}

	return detail, nil
}

// DeletePolicy deletes a CAM policy.
// reference: https://cloud.tencent.com/document/product/598/34577
func (t *TCloudImpl) DeletePolicy(kt *kit.Kit, opt *typeaccount.TCloudDeletePolicyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CamServiceClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewDeletePolicyRequest()
	for _, id := range opt.PolicyIDs {
		req.PolicyId = append(req.PolicyId, converter.ValToPtr(id))
	}

	if _, err = client.DeletePolicyWithContext(kt.Ctx, req); err != nil {
		logs.Errorf("delete cam policies failed, policyIDs: %v, err: %v, rid: %s", opt.PolicyIDs, err, kt.Rid)
		return err
	}

	return nil
}
