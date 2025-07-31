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
	"fmt"
	"strings"

	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateAzureSecurityGroup create azure security group.
func (g *securityGroup) CreateAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureOption{
		ResourceGroupName: req.ResourceGroupName,
		Region:            req.Region,
		Name:              req.Name,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create azure security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{
			{
				CloudID:   *sg.ID,
				BkBizID:   req.BkBizID,
				Region:    req.Region,
				Name:      *sg.Name,
				Memo:      req.Memo,
				AccountID: req.AccountID,
				Extension: &corecloud.AzureSecurityGroupExtension{
					ResourceGroupName: req.ResourceGroupName,
					Etag:              sg.Etag,
					FlushConnection:   sg.FlushConnection,
					ResourceGUID:      sg.ResourceGUID,
				},
				// Tags:        core.NewTagMap(req.Tags...),
				MgmtType:    req.MgmtType,
				MgmtBizID:   req.MgmtBizID,
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				UsageBizIds: req.UsageBizIds},
		},
	}
	result, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteAzureSecurityGroup delete azure security group.
func (g *securityGroup) DeleteAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureOption{
		ResourceGroupName: sg.Extension.ResourceGroupName,
		Region:            sg.Region,
		Name:              sg.Name,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete azure security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateAzureSecurityGroup update azure security group.
func (g *securityGroup) UpdateAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.AzureSecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{
			{
				ID:   id,
				Memo: req.Memo,
			},
		},
	}
	if err := g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AzureListSecurityGroupStatistic ...
func (g *securityGroup) AzureListSecurityGroupStatistic(cts *rest.Contexts) (any, error) {
	req := new(proto.ListSecurityGroupStatisticReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgMap, err := g.getAzureSecurityGroupMap(cts.Kit, req.SecurityGroupIDs)
	if err != nil {
		logs.Errorf("get security group map failed, sgID: %v, err: %v, rid: %s", req.SecurityGroupIDs, err, cts.Kit.Rid)
		return nil, err
	}

	cloudIDToSgIDMap := make(map[string]string)
	resGroupToCloudIDsMap := make(map[string][]string)
	sgIDToResourceCountMap := make(map[string]map[string]int64)
	for _, sgID := range req.SecurityGroupIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			return nil, fmt.Errorf("azure security group: %s not found", sgID)
		}
		cloudIDToSgIDMap[sg.CloudID] = sgID
		ResGroupName := sg.Extension.ResourceGroupName
		resGroupToCloudIDsMap[ResGroupName] = append(resGroupToCloudIDsMap[ResGroupName], sg.CloudID)
		sgIDToResourceCountMap[sgID] = make(map[string]int64)
	}

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for resourceGroupName, cloudIDs := range resGroupToCloudIDsMap {
		resp, err := g.listAzureSecurityGroupFromCloud(cts.Kit, client, resourceGroupName, cloudIDs)
		if err != nil {
			logs.Errorf("request adaptor to list azure security group failed, err: %v, resourceGroupName: %s,"+
				" cloudIDs: %v, rid: %s", err, resourceGroupName, cloudIDs, cts.Kit.Rid)
			return nil, err
		}

		if err = g.countAzureSecurityGroupStatistic(resp, sgIDToResourceCountMap, cloudIDToSgIDMap); err != nil {
			logs.Errorf("count azure security group statistic failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return resCountMapToSGStatisticResp(sgIDToResourceCountMap), nil
}

func (g *securityGroup) countAzureSecurityGroupStatistic(details []*armnetwork.SecurityGroup,
	sgIDToResourceCountMap map[string]map[string]int64, cloudIDToSgIDMap map[string]string) error {

	for _, one := range details {
		cloudID := strings.ToLower(converter.PtrToVal(one.ID))
		sgID, ok := cloudIDToSgIDMap[cloudID]
		if !ok {
			logs.Warnf("azure security group: %s not found in cloudIDToSgIDMap", cloudID)
			continue
		}
		for _, networkInterface := range one.Properties.NetworkInterfaces {
			if networkInterface == nil {
				continue
			}
			resType := converter.PtrToVal(networkInterface.Type)
			if resType == "" {
				resType = "network_interface"
			}
			sgIDToResourceCountMap[sgID][resType]++
		}
	}
	return nil
}

func (g *securityGroup) getAzureSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	AzureSecurityGroupRule, error) {

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, id: %s, err: %v, rid: %s", id, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}
