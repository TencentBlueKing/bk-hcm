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
	csuser "hcm/pkg/api/cloud-server/user"
	"hcm/pkg/api/core"
	coreuser "hcm/pkg/api/core/user"
	dataservice "hcm/pkg/api/data-service"
	dsuser "hcm/pkg/api/data-service/user"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateCollection create collection.
func (svc *service) CreateCollection(cts *rest.Contexts) (interface{}, error) {
	req := new(csuser.CreateCollectionReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReq := &dsuser.UserCollectionCreateReq{
		ResType: req.ResType,
		ResID:   req.ResID,
	}
	result, err := svc.client.DataService().Global.UserCollection.Create(cts.Kit, createReq)
	if err != nil {
		logs.Errorf("create collection failed, err: %v, req: %+v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// DeleteCollection delete collection.
func (svc *service) DeleteCollection(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	// 只能自己删除自己的收藏
	delReq := &dataservice.BatchDeleteReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.Equal.Factory(),
					Value: id,
				},
				&filter.AtomRule{
					Field: "user",
					Op:    filter.Equal.Factory(),
					Value: cts.Kit.User,
				},
			},
		},
	}
	if err := svc.client.DataService().Global.UserCollection.BatchDelete(cts.Kit, delReq); err != nil {
		logs.Errorf("delete biz collection failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListResourceCollection list resource collection.
func (svc *service) ListResourceCollection(cts *rest.Contexts) (interface{}, error) {

	resType := enumor.UserCollectionResType(cts.PathParameter("res_type").String())
	collections := make([]coreuser.UserCollection, 0)
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: resType,
				},
				&filter.AtomRule{
					Field: "user",
					Op:    filter.Equal.Factory(),
					Value: cts.Kit.User,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	for {
		result, err := svc.client.DataService().Global.UserCollection.List(cts.Kit, req)
		if err != nil {
			logs.Errorf("list user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		collections = append(collections, result.Details...)

		if len(result.Details) < int(req.Page.Limit) {
			break
		}

		req.Page.Start += uint32(req.Page.Limit)
	}

	return collections, nil
}
