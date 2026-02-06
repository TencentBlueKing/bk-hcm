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

package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetSpringResPoolChargeType get spring resource pool charge type config
func (s *service) GetSpringResPoolChargeType(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	chargeType, err := s.logics.SpringResPool().GetChargeType(cts.Kit, bizID)
	if err != nil {
		logs.Errorf("failed to get spring res pool charge type, bk_biz_id: %d, err: %v, rid: %s",
			bizID, err, cts.Kit.Rid)
		return nil, err
	}

	return map[string]interface{}{"charge_type": chargeType}, nil
}

// UpsertSpringResPoolChargeType upsert spring resource pool charge type config
func (s *service) UpsertSpringResPoolChargeType(cts *rest.Contexts) (interface{}, error) {
	req := new(types.UpsertSpringResPoolChargeTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.GlobalConfig, Action: meta.Create}}); err != nil {
		logs.Errorf("upsert global config auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.SpringResPool().UpsertChargeType(cts.Kit, req); err != nil {
		logs.Errorf("failed to upsert spring res pool charge type, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteSpringResPoolChargeType delete spring resource pool charge type config
func (s *service) DeleteSpringResPoolChargeType(cts *rest.Contexts) (interface{}, error) {
	req := new(types.DelSpringResPoolChargeTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.GlobalConfig, Action: meta.Delete}}); err != nil {
		logs.Errorf("delete global config auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.SpringResPool().DeleteChargeType(cts.Kit, req.BizID); err != nil {
		logs.Errorf("failed to delete spring res pool charge type, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
