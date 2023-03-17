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

package cmdb

import (
	"fmt"
	"strconv"

	"hcm/cmd/web-server/service/capability"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// InitCmdbService initial the cmdbSvc service
func InitCmdbService(c *capability.Capability) {
	svr := &cmdbSvc{
		esbClient:  c.EsbClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("ListBiz", "POST", "/bk_bizs/list", svr.ListBiz)
	h.Add("ListAuthorizedBiz", "POST", "/authorized/bizs/list", svr.ListAuthorizedBiz)
	h.Add("ListCloudArea", "POST", "/cloud_areas/list", svr.ListCloudArea)

	h.Load(c.WebService)
}

type cmdbSvc struct {
	esbClient  esb.Client
	authorizer auth.Authorizer
}

// ListBiz list all biz from cmdb
func (c *cmdbSvc) ListBiz(cts *rest.Contexts) (interface{}, error) {
	return c.listBiz(cts.Kit, nil)
}

// ListAuthorizedBiz list authorized biz with biz access permission from cmdb
func (c *cmdbSvc) ListAuthorizedBiz(cts *rest.Contexts) (interface{}, error) {
	authInstReq := &meta.ListAuthResInput{Type: meta.Biz, Action: meta.Access}
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
	resp, err := c.esbClient.Cmdb().SearchBusiness(kt.Ctx, params)
	if err != nil {
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	infos := resp.SearchBizResult.Info
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

	res, err := c.esbClient.Cmdb().SearchCloudArea(cts.Kit.Ctx, params)
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
