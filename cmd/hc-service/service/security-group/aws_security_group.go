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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateAwsSecurityGroup create aws security group.
func (g *securityGroup) CreateAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AwsSecurityGroupCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
		CloudVpcID:  req.CloudVpcID,
	}
	cloudID, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	listOpt := &types.AwsSecurityGroupListOption{
		Region:   req.Region,
		CloudIDs: []string{cloudID},
	}
	result, err := client.ListSecurityGroup(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	if len(result.SecurityGroups) != 1 {
		logs.Errorf("create aws security group succeeds, but query failed, cloud_id: %s, rid: %s", cloudID, cts.Kit.Rid)
		return nil, fmt.Errorf("create aws security group succeeds, but query failed")
	}

	vpcID, err := g.getVpcIDByCloudVpcID(cts.Kit, req.CloudVpcID)
	if err != nil {
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{
			{
				CloudID:   *result.SecurityGroups[0].GroupId,
				BkBizID:   req.BkBizID,
				Region:    req.Region,
				Name:      *result.SecurityGroups[0].GroupName,
				Memo:      result.SecurityGroups[0].Description,
				AccountID: req.AccountID,
				Extension: &corecloud.AwsSecurityGroupExtension{
					VpcID:        vpcID,
					CloudVpcID:   result.SecurityGroups[0].VpcId,
					CloudOwnerID: result.SecurityGroups[0].OwnerId,
				},
			},
		},
	}
	createResp, err := g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: createResp.IDs[0]}, nil
}

func (g *securityGroup) getVpcIDByCloudVpcID(kt *kit.Kit, cloudVpcID string) (string, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudVpcID),
		Page:   core.DefaultBasePage,
		Fields: []string{"id"},
	}
	result, err := g.dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("request dataservice to list vpc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return "", err
	}

	if len(result.Details) == 0 {
		return "", errf.Newf(errf.RecordNotFound, "vpc(cloud_id=%s) not found", cloudVpcID)
	}

	return result.Details[0].CloudID, nil
}

// DeleteAwsSecurityGroup delete aws security group.
func (g *securityGroup) DeleteAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.Aws(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.SecurityGroupDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice BatchDeleteSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncAwsSecurityGroup delete aws security group.
func (g *securityGroup) SyncAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	cloudMap, err := g.getDatasFromAwsForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := g.getDatasFromDSForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	err = g.diffAwsSecurityGroupSync(cts, cloudMap, dsMap, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getDatasFromAwsForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromAwsForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncAwsDiff, error) {

	client, err := g.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &types.AwsSecurityGroupListOption{
		Region: req.Region,
	}
	result, err := client.ListSecurityGroup(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.SecurityGroupSyncAwsDiff)
	for _, one := range result.SecurityGroups {
		sg := new(proto.SecurityGroupSyncAwsDiff)
		sg.SecurityGroup = one
		cloudMap[*one.GroupId] = sg
	}

	return cloudMap, nil
}

// diffAwsSecurityGroupSync make aws and data-service diff, process resources according to diff
func (g *securityGroup) diffAwsSecurityGroupSync(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAwsDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, req *proto.SecurityGroupSyncReq) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync aws SecurityGroup delete operate rid: %s", cts.Kit.Rid)
		err := g.diffSecurityGroupSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("sync delete aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAwsSGRuleSyncDelete(cts, deleteCloudIDs, dsMap)
		if err != nil {
			logs.Errorf("sync delete aws security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync aws SecurityGroup update operate rid: %s", cts.Kit.Rid)
		err := g.diffAwsSecurityGroupSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("sync update aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAwsSGRuleSyncUpdate(cts, updateCloudIDs, req, dsMap)
		if err != nil {
			logs.Errorf("sync update aws security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync aws SecurityGroup add operate rid: %s", cts.Kit.Rid)
		ids, err := g.diffAwsSecurityGroupSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("sync add aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffAwsSGRuleSyncAdd(cts, ids, req)
		if err != nil {
			logs.Errorf("sync add aws security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffSecurityGroupSyncAdd for add
func (g *securityGroup) diffAwsSecurityGroupSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAwsDiff,
	req *proto.SecurityGroupSyncReq, addCloudIDs []string) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.GroupId,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      *cloudMap[id].SecurityGroup.GroupName,
			Memo:      cloudMap[id].SecurityGroup.Description,
			AccountID: req.AccountID,
			Extension: &corecloud.AwsSecurityGroupExtension{
				CloudVpcID:   cloudMap[id].SecurityGroup.VpcId,
				CloudOwnerID: cloudMap[id].SecurityGroup.OwnerId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	results, err := g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffSecurityGroupSyncUpdate for update
func (g *securityGroup) diffAwsSecurityGroupSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncAwsDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, updateCloudIDs []string) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.GroupName == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.Description == dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.GroupName,
			Memo: cloudMap[id].SecurityGroup.Description,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := g.dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
