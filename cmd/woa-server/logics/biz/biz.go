/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package biz

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/slice"
)

// ListAuthorizedBiz list authorized biz with biz access permission from cmdb.
func (l *logics) ListAuthorizedBiz(kt *kit.Kit) ([]int64, error) {
	authReq := &meta.ListAuthResInput{Type: meta.Biz, Action: meta.Access}
	authResp, err := l.authorizer.ListAuthorizedInstances(kt, authReq)
	if err != nil {
		logs.Errorf("failed to list authorized instance, err: %v, rid: %d", err, kt.Rid)
		return nil, err
	}

	// search cmdb biz with biz access permission.
	cmdbReq := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
	}
	if !authResp.IsAny {
		ids := make([]int64, 0, len(authResp.IDs))
		for _, id := range authResp.IDs {
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse id %s failed, err: %v", id, err)
			}
			ids = append(ids, intID)
		}

		cmdbReq.BizPropertyFilter = &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    "bk_biz_id",
						Operator: cmdb.OperatorIn,
						Value:    ids,
					},
				},
			},
		}
	}
	resp, err := l.cmdbCli.SearchBusiness(kt, cmdbReq)
	if err != nil {
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	bkBizIDs := make([]int64, 0, len(resp.Info))
	for _, info := range resp.Info {
		bkBizIDs = append(bkBizIDs, info.BizID)
	}

	return bkBizIDs, nil
}

// GetBizOrgRel get biz org relation.
func (l *logics) GetBizOrgRel(kt *kit.Kit, bkBizID int64) (*mtypes.BizOrgRel, error) {
	// search cmdb business belonging.
	req := &cmdb.SearchBizCompanyCmdbInfoParams{
		BizIDs: []int64{bkBizID},
	}

	resp, err := l.cmdbCli.SearchBizCompanyCmdbInfo(kt, req)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return nil, errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	// convert search biz belonging response to biz org relation response.
	bizBelong := (*resp)[0]
	rst := &mtypes.BizOrgRel{
		BkBizID:         bizBelong.BkBizID,
		BkBizName:       bizBelong.BizName,
		OpProductID:     bizBelong.BkProductID,
		OpProductName:   bizBelong.BkProductName,
		PlanProductID:   bizBelong.PlanProductID,
		PlanProductName: bizBelong.PlanProductName,
		VirtualDeptID:   bizBelong.VirtualDeptID,
		VirtualDeptName: bizBelong.VirtualDeptName,
	}

	return rst, nil
}

// ListBizsOrgRel list bizs org rel.
func (l *logics) ListBizsOrgRel(kt *kit.Kit, bkBizIDs []int64) (map[int64]*mtypes.BizOrgRel, error) {
	// search cmdb business belonging.
	req := &cmdb.SearchBizCompanyCmdbInfoParams{
		BizIDs: bkBizIDs,
	}

	resp, err := l.cmdbCli.SearchBizCompanyCmdbInfo(kt, req)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || len(*resp) != len(bkBizIDs) {
		logs.Errorf("search biz belonging, but resp is empty or resp != biz list length, rid: %s", kt.Rid)
		return nil, errors.New("search biz belonging, but resp is empty or resp != biz list length")
	}

	rst := make(map[int64]*mtypes.BizOrgRel)
	for _, bizBelong := range *resp {
		rst[bizBelong.BkBizID] = &mtypes.BizOrgRel{
			BkBizID:         bizBelong.BkBizID,
			BkBizName:       bizBelong.BizName,
			OpProductID:     bizBelong.BkProductID,
			OpProductName:   bizBelong.BkProductName,
			PlanProductID:   bizBelong.PlanProductID,
			PlanProductName: bizBelong.PlanProductName,
			VirtualDeptID:   bizBelong.VirtualDeptID,
			VirtualDeptName: bizBelong.VirtualDeptName,
		}
	}
	return rst, nil
}

// BatchCheckUserBizAccessAuth batch check user biz access auth.
func (l *logics) BatchCheckUserBizAccessAuth(kt *kit.Kit, bkBizID int64, userNames []string) (map[string]bool, error) {
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	userAuthMap, err := l.authorizer.AuthorizeByUsers(kt, userNames, authRes)
	if err != nil {
		logs.Errorf("failed to check authorize by users, bkBizID: %d, err: %v, userNames: %v, rid: %s",
			bkBizID, err, userNames, kt.Rid)
		return nil, err
	}

	processorAuth := make(map[string]bool)
	for _, userName := range userNames {
		if authInfo, ok := userAuthMap[userName]; ok {
			processorAuth[userName] = authInfo.IsAuth
		}
	}
	return processorAuth, nil
}

// GetBkBizMaintainer get bk biz maintainer.
func (l *logics) GetBkBizMaintainer(kt *kit.Kit, bkBizIDs []int64) (map[int64][]string, error) {
	data := make(map[int64][]string)

	bkBizIDs = slice.Unique(bkBizIDs)
	if len(bkBizIDs) == 0 {
		return data, nil
	}

	for _, split := range slice.Split(bkBizIDs, cmdb.BusinessSearchMaxLimit) {
		rules := []cmdb.Rule{
			&cmdb.AtomRule{
				Field:    "bk_biz_id",
				Operator: cmdb.OperatorIn,
				Value:    split,
			},
		}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}

		params := &cmdb.SearchBizParams{
			BizPropertyFilter: expression,
			Fields:            []string{"bk_biz_id", "bk_biz_name", "bk_biz_maintainer"},
		}
		resp, err := l.cmdbCli.SearchBusiness(kt, params)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
		}

		for _, biz := range resp.Info {
			data[biz.BizID] = strings.Split(biz.BkBizMaintainer, ",")
		}
	}

	return data, nil
}

// GetBizNames get biz names.
func (l *logics) GetBizNames(kt *kit.Kit, bkBizIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)

	bkBizIDs = slice.Unique(bkBizIDs)
	if len(bkBizIDs) == 0 {
		return result, nil
	}

	for _, split := range slice.Split(bkBizIDs, cmdb.BusinessSearchMaxLimit) {
		rules := []cmdb.Rule{
			&cmdb.AtomRule{
				Field:    "bk_biz_id",
				Operator: cmdb.OperatorIn,
				Value:    split,
			},
		}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}

		params := &cmdb.SearchBizParams{
			BizPropertyFilter: expression,
			Fields:            []string{"bk_biz_id", "bk_biz_name"},
		}
		resp, err := l.cmdbCli.SearchBusiness(kt, params)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
		}

		for _, info := range resp.Info {
			result[info.BizID] = info.BizName
		}
	}

	return result, nil
}

// GetBkBizIDsByOpProductName get bk biz ids list by op product name.
func (l *logics) GetBkBizIDsByOpProductName(kt *kit.Kit, opProductNames []string) (map[string][]int64, error) {
	data := make(map[string][]int64)

	opProductNames = slice.Unique(opProductNames)
	if len(opProductNames) == 0 {
		return data, nil
	}

	for _, split := range slice.Split(opProductNames, cmdb.BusinessSearchMaxLimit) {
		rules := []cmdb.Rule{
			&cmdb.AtomRule{
				Field:    "bk_product_name",
				Operator: cmdb.OperatorIn,
				Value:    split,
			},
		}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}
		params := &cmdb.SearchBizParams{
			BizPropertyFilter: expression,
			Fields:            []string{"bk_biz_id", "bk_biz_name", "bk_product_id", "bk_product_name"},
			Page: cmdb.BasePage{
				Start: 0,
				Limit: int64(cmdb.BusinessSearchMaxLimit),
			},
		}

		for {
			resp, err := l.cmdbCli.SearchBusiness(kt, params)
			if err != nil {
				logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
				return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
			}

			for _, biz := range resp.Info {
				data[biz.BkProductName] = append(data[biz.BkProductName], biz.BizID)
			}

			if int64(len(resp.Info)) < params.Page.Limit {
				break
			}
			params.Page.Start += params.Page.Limit
		}
	}

	return data, nil
}
