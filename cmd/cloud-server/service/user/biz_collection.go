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

package user

import (
	"strconv"

	csuser "hcm/pkg/api/cloud-server/user"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	dsuser "hcm/pkg/api/data-service/user"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateBizCollection create biz collection.
func (svc *service) CreateBizCollection(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(csuser.BizCollectionReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.BizCollection, Action: meta.Create}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create biz collection auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &dsuser.UserCollectionCreateReq{
		ResType: enumor.BizCollResType,
		ResID:   strconv.FormatInt(req.BkBizID, 10),
	}
	if _, err = svc.client.DataService().Global.UserCollection.Create(cts.Kit, createReq); err != nil {
		logs.Errorf("create biz collection failed, err: %v, req: %+v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteBizCollection delete biz collection.
func (svc *service) DeleteBizCollection(cts *rest.Contexts) (interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(csuser.BizCollectionReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.BizCollection, Action: meta.Delete}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("delete biz collection auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataservice.BatchDeleteReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: enumor.BizCollResType,
				},
				&filter.AtomRule{
					Field: "res_id",
					Op:    filter.Equal.Factory(),
					Value: strconv.FormatInt(req.BkBizID, 10),
				},
				&filter.AtomRule{
					Field: "user",
					Op:    filter.Equal.Factory(),
					Value: cts.Kit.User,
				},
			},
		},
	}
	if err = svc.client.DataService().Global.UserCollection.BatchDelete(cts.Kit, delReq); err != nil {
		logs.Errorf("delete biz collection failed, err: %v, bizID: %d, rid: %s", err, req.BkBizID, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetBizCollection get biz collection.
func (svc *service) GetBizCollection(cts *rest.Contexts) (interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.BizCollection, Action: meta.Delete}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("get biz collection auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bizIDs := make([]int64, 0)
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: enumor.BizCollResType,
				},
				&filter.AtomRule{
					Field: "user",
					Op:    filter.Equal.Factory(),
					Value: cts.Kit.User,
				},
			},
		},
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"res_id"},
	}
	for {
		result, err := svc.client.DataService().Global.UserCollection.List(cts.Kit, req)
		if err != nil {
			logs.Errorf("list user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			bizID, err := strconv.ParseInt(one.ResID, 10, 64)
			if err != nil {
				logs.Errorf("get biz collection parse res_id failed, err: %v, resID: %s, rid: %s",
					err, one.ResID, cts.Kit.Rid)
				return nil, err
			}

			bizIDs = append(bizIDs, bizID)
		}

		if len(result.Details) < int(req.Page.Limit) {
			break
		}

		req.Page.Start += uint32(req.Page.Limit)
	}

	return bizIDs, nil
}
