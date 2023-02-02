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
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// decodeSecurityGroupSyncReq get par from body
func (g *securityGroup) decodeSecurityGroupSyncReq(cts *rest.Contexts) (*proto.SecurityGroupSyncReq, error) {

	req := new(proto.SecurityGroupSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return req, nil
}

// getDatasFromDSForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromDSForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncDS, error) {

	start := 0
	resultsHcm := make([]corecloud.BaseSecurityGroup, 0)
	for {
		dataReq := &dataproto.SecurityGroupListReq{
			Filter: tools.EqualExpression("account_id", req.AccountID),
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		results, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			dataReq)

		if err != nil {
			logs.Errorf("from data-service list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(results.Details) == 0 {
			break
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*proto.SecurityGroupSyncDS)
	for _, result := range resultsHcm {
		sg := new(proto.SecurityGroupSyncDS)
		sg.IsUpdated = false
		sg.HcSecurityGroup = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// diffSecurityGroupSyncDelete for delete
func (g *securityGroup) diffSecurityGroupSyncDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	batchDeleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

// getAddCloudIDs
func getAddCloudIDs[T any](cloudMap map[string]T, dsMap map[string]*proto.SecurityGroupSyncDS) []string {

	addCloudIDs := []string{}
	for id := range cloudMap {
		if _, ok := dsMap[id]; !ok {
			addCloudIDs = append(addCloudIDs, id)
		} else {
			dsMap[id].IsUpdated = true
		}
	}

	return addCloudIDs
}

// getDeleteAndUpdateCloudIDs
func getDeleteAndUpdateCloudIDs(dsMap map[string]*proto.SecurityGroupSyncDS) ([]string, []string) {

	deleteCloudIDs := []string{}
	updateCloudIDs := []string{}
	for id, one := range dsMap {
		if !one.IsUpdated {
			deleteCloudIDs = append(deleteCloudIDs, id)
		} else {
			updateCloudIDs = append(updateCloudIDs, id)
		}
	}

	return deleteCloudIDs, updateCloudIDs
}
