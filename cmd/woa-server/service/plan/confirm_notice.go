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

package plan

import (
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// PushResPlanConfirmNotice 手动触发资源预测确认通知
func (s *service) PushResPlanConfirmNotice(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.PushResPlanConfirmNoticeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// biz_ids 为空时通知所有业务
	successIDs, failedIDs, err := s.planController.PushResPlanConfirmNotice(cts.Kit, req.BkBizIDs)
	if err != nil {
		logs.Errorf("failed to push res plan confirm notice, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &ptypes.PushResPlanConfirmNoticeResp{
		SuccessIDs: successIDs,
		FailedIDs:  failedIDs,
	}, nil
}

// ConfirmResPlanDemands 确认资源预测需求
func (s *service) ConfirmResPlanDemands(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ConfirmResPlanDemandsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode confirm res plan demands request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate confirm res plan demands request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	successIDs, failedIDs, err := s.planController.ConfirmResPlanDemands(cts.Kit, req.BkBizID, req.DemandIDs)
	if err != nil {
		logs.Errorf("failed to confirm res plan demands, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &ptypes.ConfirmResPlanDemandsResp{
		SuccessIDs: successIDs,
		FailedIDs:  failedIDs,
	}, nil
}

// ConfirmBizResPlanDemands 确认业务资源预测需求
func (s *service) ConfirmBizResPlanDemands(cts *rest.Contexts) (interface{}, error) {

	req := new(ptypes.ConfirmBizResPlanDemandsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode confirm biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate confirm biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Update}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	successIDs, failedIDs, err := s.planController.ConfirmResPlanDemands(cts.Kit, bkBizID, req.DemandIDs)
	if err != nil {
		logs.Errorf("failed to confirm biz resource plan demands, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &ptypes.ConfirmResPlanDemandsResp{
		SuccessIDs: successIDs,
		FailedIDs:  failedIDs,
	}, nil
}
