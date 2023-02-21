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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
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

	opt := &types.AzureSecurityGroupOption{
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
					Etag:              sg.Etag,
					FlushConnection:   sg.Properties.FlushConnection,
					ResourceGUID:      sg.Properties.ResourceGUID,
					ProvisioningState: string(*sg.Properties.ProvisioningState),
				},
			},
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

	opt := &types.AzureSecurityGroupOption{
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

// SyncAzureSecurityGroup create azure security group.
func (g *securityGroup) SyncAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	cloudMap, err := g.getDatasFromAzureForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := g.getDatasFromAzureDSForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	err = g.diffAzureSecurityGroupSync(cts, cloudMap, dsMap, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getDatasFromDSForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromAzureDSForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncDS, error) {

	start := 0
	resultsHcm := make([]corecloud.BaseSecurityGroup, 0)
	for {
		dataReq := &dataproto.SecurityGroupListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "extension.resource_group_name",
						Op:    filter.JSONEqual.Factory(),
						Value: req.ResourceGroupName,
					},
				},
			},
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

// getDatasFromAzureForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromAzureForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncAzureDiff, error) {

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &types.AzureSecurityGroupListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	result, err := client.ListSecurityGroup(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.SecurityGroupSyncAzureDiff)
	for _, one := range result {
		sg := new(proto.SecurityGroupSyncAzureDiff)
		sg.SecurityGroup = one
		cloudMap[*one.ID] = sg
	}

	return cloudMap, nil
}

// diffAzureSecurityGroupSync make azure and data-service diff, process resources according to diff
func (g *securityGroup) diffAzureSecurityGroupSync(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAzureDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, req *proto.SecurityGroupSyncReq) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup delete operate rid: %s", cts.Kit.Rid)
		err := g.diffSecurityGroupSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("sync delete azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAzureSGRuleSyncDelete(cts, deleteCloudIDs, dsMap)
		if err != nil {
			logs.Errorf("sync delete azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup update operate rid: %s", cts.Kit.Rid)
		err := g.diffAzureSecurityGroupSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("sync update azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAzureSGRuleSyncUpdate(cts, updateCloudIDs, req, dsMap)
		if err != nil {
			logs.Errorf("sync update azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup add operate rid: %s", cts.Kit.Rid)
		ids, err := g.diffAzureSecurityGroupSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("sync add azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAzureSGRuleSyncAdd(cts, ids, req)
		if err != nil {
			logs.Errorf("sync add azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffAzuerSecurityGroupSyncAdd for add
func (g *securityGroup) diffAzureSecurityGroupSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAzureDiff,
	req *proto.SecurityGroupSyncReq, addCloudIDs []string) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.ID,
			BkBizID:   constant.UnassignedBiz,
			Region:    *cloudMap[id].SecurityGroup.Location,
			Name:      *cloudMap[id].SecurityGroup.Name,
			Memo:      nil,
			AccountID: req.AccountID,
			Extension: &corecloud.AzureSecurityGroupExtension{
				ResourceGroupName: req.ResourceGroupName,
				Etag:              cloudMap[id].SecurityGroup.Etag,
				FlushConnection:   cloudMap[id].SecurityGroup.Properties.FlushConnection,
				ResourceGUID:      cloudMap[id].SecurityGroup.Properties.ResourceGUID,
				ProvisioningState: string(*cloudMap[id].SecurityGroup.Properties.ProvisioningState),
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	results, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffAzureSecurityGroupSyncUpdate for update
func (g *securityGroup) diffAzureSecurityGroupSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAzureDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, updateCloudIDs []string) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.Name == dsMap[id].HcSecurityGroup.Name {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.Name,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
