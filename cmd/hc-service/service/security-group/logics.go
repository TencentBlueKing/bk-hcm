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

package securitygroup

import (
	"errors"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

func buildSGCommonRelDeleteReqForMultiResource(resType enumor.CloudResourceType, sgID string, resIDs ...string) (
	*dataproto.BatchDeleteReq, error) {

	if len(resIDs) == 0 {
		return nil, errors.New("cvmIDs is required")
	}
	return &dataproto.BatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("security_group_id", sgID),
			tools.RuleIn("res_id", resIDs),
			tools.RuleEqual("res_type", resType),
		),
	}, nil
}

func buildSGCommonRelDeleteReq(vendor enumor.Vendor, resID string, sgIDs []string,
	resType enumor.CloudResourceType) *dataproto.BatchDeleteReq {

	return &dataproto.BatchDeleteReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: resType,
				},
				&filter.AtomRule{
					Field: "res_id",
					Op:    filter.Equal.Factory(),
					Value: resID,
				},
				&filter.AtomRule{
					Field: "security_group_id",
					Op:    filter.In.Factory(),
					Value: sgIDs,
				},
			},
		},
	}
}

func (g *securityGroup) getSecurityGroupAndCvm(kt *kit.Kit, sgID, cvmID string) (*corecloud.BaseSecurityGroup,
	*corecvm.BaseCvm, error) {

	sgReq := &protocloud.SecurityGroupListReq{
		Filter: tools.EqualExpression("id", sgID),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, id: %s, rid: %s",
			err, sgID, kt.Rid)
		return nil, nil, err
	}

	if len(sgResult.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", sgID)
	}

	cvmReq := &core.ListReq{
		Filter: tools.EqualExpression("id", cvmID),
		Page:   core.NewDefaultBasePage(),
	}
	cvmResult, err := g.dataCli.Global.Cvm.ListCvm(kt, cvmReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, id: %s, rid: %s", err, cvmID, kt.Rid)
		return nil, nil, err
	}

	if len(cvmResult.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "cvm: %s not found", sgID)
	}

	return &sgResult.Details[0], &cvmResult.Details[0], nil
}

// createSGCommonRels 先删除cvmID关联的安全组关系，再创建新的安全组关系
// sgCloudIDs的入参顺序决定新建的关联关系的优先级
func (g *securityGroup) createSGCommonRels(kt *kit.Kit, vendor enumor.Vendor, resType enumor.CloudResourceType,
	cvmID string, sgIDs []string) error {

	createReq := &protocloud.SGCommonRelBatchUpsertReq{
		DeleteReq: &dataproto.BatchDeleteReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("res_id", cvmID),
				tools.RuleEqual("res_type", resType),
			),
		},
	}

	for i, sgID := range sgIDs {
		createReq.Rels = append(createReq.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: sgID,
			ResVendor:       vendor,
			ResID:           cvmID,
			ResType:         resType,
			Priority:        int64(i) + 1,
		})
	}
	if err := g.dataCli.Global.SGCommonRel.BatchUpsertSgCommonRels(kt, createReq); err != nil {
		logs.Errorf("request dataservice create security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, createReq, kt.Rid)
		return err
	}
	return nil
}

func (g *securityGroup) getSecurityGroupMapByCloudIDs(kt *kit.Kit, vendor enumor.Vendor, region string,
	cloudIDs []string) (map[string]string, error) {

	cloudIDs = slice.Unique(cloudIDs)
	m := make(map[string]string)
	for _, ids := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.SecurityGroupListReq{
			Field: []string{"id", "cloud_id"},
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", ids),
				tools.RuleEqual("region", region),
				tools.RuleEqual("vendor", vendor),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("request dataservice list security group failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, one := range resp.Details {
			m[one.CloudID] = one.ID
		}
	}
	return m, nil
}

func resCountMapToSGStatisticResp(
	sgIDToResourceCountMap map[string]map[string]int64) *proto.ListSecurityGroupStatisticResp {

	result := &proto.ListSecurityGroupStatisticResp{
		Details: make([]*proto.SecurityGroupStatisticItem, 0, len(sgIDToResourceCountMap)),
	}
	for sgID, resourceMap := range sgIDToResourceCountMap {
		one := &proto.SecurityGroupStatisticItem{
			ID:        sgID,
			Resources: make([]proto.SecurityGroupStatisticResource, 0, len(resourceMap)),
		}
		for resName, count := range resourceMap {
			one.Resources = append(one.Resources,
				proto.SecurityGroupStatisticResource{
					ResName: resName,
					Count:   count,
				},
			)
		}
		result.Details = append(result.Details, one)
	}
	return result
}
