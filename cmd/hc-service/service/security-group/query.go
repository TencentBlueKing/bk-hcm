/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package securitygroup

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// getCvms 根据cvmIDs获取云服务器信息
func (g *securityGroup) getCvms(kt *kit.Kit, cvmIDs []string) ([]cvm.BaseCvm, error) {

	result := make([]cvm.BaseCvm, 0, len(cvmIDs))
	for _, ids := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := g.dataCli.Global.Cvm.ListCvm(kt, listReq)
		if err != nil {
			logs.Errorf("list cvm failed, req: %+v, err: %v, rid: %s", listReq, err, kt.Rid)
			return nil, err
		}
		result = append(result, resp.Details...)
	}

	if len(result) != len(cvmIDs) {
		logs.Errorf("list cvm failed, got %d, but expect %d, rid: %s", len(result), len(cvmIDs), kt.Rid)
		return nil, fmt.Errorf("list cvm failed, got %d, but expect %d", len(result), len(cvmIDs))
	}
	return result, nil
}

func (g *securityGroup) getSecurityGroupMap(kt *kit.Kit, sgIDs []string) (
	map[string]corecloud.BaseSecurityGroup, error) {

	sgReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("id", sgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, ids: %v, rid: %s",
			err, sgIDs, kt.Rid)
		return nil, err
	}

	sgMap := make(map[string]corecloud.BaseSecurityGroup, len(sgResult.Details))
	for _, sg := range sgResult.Details {
		sgMap[sg.ID] = sg
	}

	return sgMap, nil
}

func (g *securityGroup) getLoadBalancerInfoAndSGComRels(kt *kit.Kit, lbID string) (
	*corelb.BaseLoadBalancer, *protocloud.SGCommonRelListResult, error) {

	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", lbID),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := g.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, id: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, nil, err
	}

	if len(lbList.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "not found lb id: %s", lbID)
	}

	lbInfo := lbList.Details[0]
	// 查询目前绑定的安全组
	sgcomReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_vendor", lbInfo.Vendor),
			tools.RuleEqual("res_id", lbID),
			tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
		),
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "priority", Order: "ASC"},
	}
	sgComList, err := g.dataCli.Global.SGCommonRel.ListSgCommonRels(kt, sgcomReq)
	if err != nil {
		logs.Errorf("call dataserver to list sg common failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, nil, err
	}

	return &lbInfo, sgComList, nil
}
