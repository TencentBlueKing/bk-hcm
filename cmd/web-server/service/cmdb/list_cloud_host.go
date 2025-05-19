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
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/slice"
)

// ListCloudHost list cloud host.
func (c *cmdbSvc) ListCloudHost(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(webserver.CloudHostListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	params := &cmdb.ListBizHostParams{
		BizID:       bizID,
		BkSetIDs:    req.BkSetIDs,
		BkModuleIDs: req.BkModuleIDs,
		Fields:      []string{"bk_cloud_inst_id"},
		Page:        req.Page,
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: "AND",
				Rules: []cmdb.Rule{
					&cmdb.AtomRule{
						Field:    "bk_cloud_host_identifier",
						Operator: "equal",
						Value:    true,
					},
				},
			},
		},
	}
	result, err := c.cmdbClient.ListBizHost(cts.Kit, params)
	if err != nil {
		logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	resp := &webserver.CloudHostListResp{
		Count:   0,
		Details: make([]corecvm.BaseCvm, 0),
	}
	if len(result.Info) == 0 {
		return resp, nil
	}

	cloudIDs := make([]string, 0, len(result.Info))
	for _, host := range result.Info {
		cloudIDs = append(cloudIDs, host.BkCloudInstID)
	}

	partIDs := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	details := make([]corecvm.BaseCvm, 0)
	for _, partID := range partIDs {
		listReq := &core.ListReq{
			Page: &core.BasePage{
				Start: 0,
				Limit: core.DefaultMaxPageLimit,
			},
			Filter: tools.ContainersExpression("cloud_id", partID),
		}
		listResult, err := c.client.CloudServer().Cvm.List(cts.Kit, bizID, listReq)
		if err != nil {
			logs.Errorf("call cloud server to list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		details = append(details, listResult.Details...)
	}

	resp.Count = result.Count
	resp.Details = details
	return resp, nil
}
