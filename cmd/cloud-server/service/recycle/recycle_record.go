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

package recycle

import (
	proto "hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListRecycleRecord list recycle record.
func (svc *svc) ListRecycleRecord(cts *rest.Contexts) (interface{}, error) {
	return svc.listRecycleRecord(cts, handler.ListResourceRecycleAuthRes)
}

// ListBizRecycleRecord list biz recycle record.
func (svc *svc) ListBizRecycleRecord(cts *rest.Contexts) (interface{}, error) {
	return svc.listRecycleRecord(cts, handler.ListBizRecycleAuthRes)
}

func (svc *svc) listRecycleRecord(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.RecycleBin, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return new(proto.RecycleRecordListResult), nil
	}
	req.Filter = expr

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	return svc.client.DataService().Global.RecycleRecord.ListRecycleRecord(cts.Kit, listReq)
}
