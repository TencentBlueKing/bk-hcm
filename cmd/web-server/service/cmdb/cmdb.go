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

// Package cmdb ...
package cmdb

import (
	"fmt"
	"strconv"

	"hcm/cmd/web-server/service/capability"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// InitCmdbService initial the cmdbSvc service
func InitCmdbService(c *capability.Capability) {
	svr := &cmdbSvc{
		esbClient:  c.EsbClient,
		authorizer: c.Authorizer,
		client:     c.ApiClient,
	}

	h := rest.NewHandler()
	h.Add("ListBiz", "POST", "/bk_bizs/list", svr.ListBiz)
	h.Add("ListAuthorizedBiz", "POST", "/authorized/bizs/list", svr.ListAuthorizedBiz)
	h.Add("ListAuthorizedBizAudit", "POST", "/authorized/audit/bizs/list", svr.ListAuthorizedBizAudit)
	h.Add("ListCloudArea", "POST", "/cloud_areas/list", svr.ListCloudArea)
	h.Add("ListAllCloudArea", "POST", "/all/cloud_areas/list", svr.ListAllCloudArea)
	h.Add("GetBizBriefCacheTopo", "GET", "/bizs/{bk_biz_id}/brief/cache/topo", svr.GetBizBriefCacheTopo)
	h.Add("ListCloudHost", "POST", "/bizs/{bk_biz_id}/cloud/hosts/list", svr.ListCloudHost)

	h.Load(c.WebService)
}

type cmdbSvc struct {
	client     *client.ClientSet
	esbClient  esb.Client
	authorizer auth.Authorizer
}

// ListBiz list all biz from cmdb
func (c *cmdbSvc) ListBiz(cts *rest.Contexts) (interface{}, error) {
	return c.listBiz(cts.Kit, nil)
}

// ListAuthorizedBiz list authorized biz with biz access permission from cmdb
func (c *cmdbSvc) ListAuthorizedBiz(cts *rest.Contexts) (interface{}, error) {
	return c.listAuthorizedBiz(cts, meta.Biz, meta.Access)
}

// ListAuthorizedBizAudit list authorized biz audit with biz access permission from cmdb
func (c *cmdbSvc) ListAuthorizedBizAudit(cts *rest.Contexts) (interface{}, error) {
	return c.listAuthorizedBiz(cts, meta.Audit, meta.Find)
}

// ListAuthorizedBiz list authorized biz with biz access permission from cmdb
func (c *cmdbSvc) listAuthorizedBiz(cts *rest.Contexts, typ meta.ResourceType,
	action meta.Action) (interface{}, error) {
	authInstReq := &meta.ListAuthResInput{Type: typ, Action: action}
	authInstRes, err := c.authorizer.ListAuthorizedInstances(cts.Kit, authInstReq)
	if err != nil {
		return nil, err
	}

	if authInstRes.IsAny {
		return c.listBiz(cts.Kit, nil)
	}

	if len(authInstRes.IDs) == 0 {
		return make([]map[string]interface{}, 0), nil
	}

	ids := make([]int64, len(authInstRes.IDs))
	for i, id := range authInstRes.IDs {
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, errf.Newf(errf.InvalidParameter, "parse id %s failed, err: %v", id, err)
		}
		ids[i] = intID
	}

	filter := &cmdb.QueryFilter{
		Rule: &cmdb.CombinedRule{
			Condition: "AND",
			Rules: []cmdb.Rule{
				&cmdb.AtomRule{
					Field:    "bk_biz_id",
					Operator: "in",
					Value:    ids,
				},
			},
		},
	}

	return c.listBiz(cts.Kit, filter)
}

func (c *cmdbSvc) listBiz(kt *kit.Kit, filter *cmdb.QueryFilter) (interface{}, error) {
	params := &cmdb.SearchBizParams{
		BizPropertyFilter: filter,
		Fields:            []string{"bk_biz_id", "bk_biz_name"},
	}
	resp, err := c.esbClient.Cmdb().SearchBusiness(kt, params)
	if err != nil {
		logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	infos := resp.Info
	data := make([]map[string]interface{}, 0, len(infos))
	for _, biz := range infos {
		data = append(data, map[string]interface{}{
			"id":   biz.BizID,
			"name": biz.BizName,
		})
	}

	return data, nil
}

// ListCloudArea list all cloud area basic info from cmdb.
func (c *cmdbSvc) ListCloudArea(cts *rest.Contexts) (interface{}, error) {
	req := new(webserver.ListCloudAreaOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	params := &cmdb.SearchCloudAreaParams{
		Fields: []string{"bk_cloud_id", "bk_cloud_name"},
		Page: cmdb.BasePage{
			Limit: int64(req.Page.Limit),
			Start: int64(req.Page.Start),
			Sort:  "bk_cloud_id",
		},
		Condition: map[string]interface{}{"bk_cloud_id": map[string]interface{}{"$ne": 0}},
	}

	if req.Name != "" {
		params.Condition["bk_cloud_name"] = map[string]interface{}{"$regex": req.Name}
	}

	if req.ID != 0 {
		params.Condition["bk_cloud_id"] = req.ID
	}

	res, err := c.esbClient.Cmdb().SearchCloudArea(cts.Kit, params)
	if err != nil {
		return nil, fmt.Errorf("call cmdb search cloud area api failed, err: %v", err)
	}

	result := &webserver.ListCloudAreaResult{
		Count: res.Count,
		Info:  make([]webserver.CloudArea, 0, len(res.Info)),
	}

	for _, cloudArea := range res.Info {
		result.Info = append(result.Info, webserver.CloudArea{
			ID:   cloudArea.CloudID,
			Name: cloudArea.CloudName,
		})
	}

	return result, nil
}

// ListAllCloudArea list all cloud area.
func (c *cmdbSvc) ListAllCloudArea(cts *rest.Contexts) (interface{}, error) {

	limit := int64(webserver.ListCloudAreaLimit)
	start := int64(0)
	params := &cmdb.SearchCloudAreaParams{
		Fields: []string{"bk_cloud_id", "bk_cloud_name"},
		Page: cmdb.BasePage{
			Limit: limit,
			Start: int64(0),
			Sort:  "bk_cloud_id",
		},
	}
	result := &webserver.ListCloudAreaResult{
		Count: 0,
		Info:  make([]webserver.CloudArea, 0),
	}
	for {
		params.Page.Start = start

		res, err := c.esbClient.Cmdb().SearchCloudArea(cts.Kit, params)
		if err != nil {
			logs.Errorf("call cmdb search cloud area api failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("call cmdb search cloud area api failed, err: %v", err)
		}

		result.Count = res.Count
		for _, cloudArea := range res.Info {
			result.Info = append(result.Info, webserver.CloudArea{
				ID:   cloudArea.CloudID,
				Name: cloudArea.CloudName,
			})
		}

		if int64(len(res.Info)) < limit {
			break
		}

		start += limit
	}

	return result, nil
}

// GetBizBriefCacheTopo define
func (c *cmdbSvc) GetBizBriefCacheTopo(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	params := &cmdb.GetBizBriefCacheTopoParams{
		BkBizID: bizID,
	}
	result, err := c.esbClient.Cmdb().GetBizBriefCacheTopo(cts.Kit, params)
	if err != nil {
		logs.Errorf("call cmdb get biz brief cache topo failed, err: %v, bizID: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
