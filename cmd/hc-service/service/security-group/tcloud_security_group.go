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
	typcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateTCloudSecurityGroup create tcloud security group.
func (g *securityGroup) CreateTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.TCloudSecurityGroupCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{
			{
				CloudID:   *sg.SecurityGroupId,
				BkBizID:   req.BkBizID,
				Region:    req.Region,
				Name:      *sg.SecurityGroupName,
				Memo:      sg.SecurityGroupDesc,
				AccountID: req.AccountID,
				Extension: &corecloud.TCloudSecurityGroupExtension{
					CloudProjectID: sg.ProjectId,
				},
			},
		},
	}
	result, err := g.dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteTCloudSecurityGroup delete tcloud security group.
func (g *securityGroup) DeleteTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.SecurityGroupDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateTCloudSecurityGroup update tcloud security group.
func (g *securityGroup) UpdateTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.SecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.TCloudSecurityGroupUpdateOption{
		CloudID:     sg.CloudID,
		Region:      sg.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	if err := client.UpdateSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to UpdateSecurityGroup failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{
			{
				ID:   sg.ID,
				Name: req.Name,
				Memo: req.Memo,
			},
		},
	}
	if err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncTCloudSecurityGroup sync tcloud security group to hcm.
func (g *securityGroup) SyncTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	cloudMap, err := g.getDatasFromTCloudForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := g.getDatasFromDSForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	err = g.diffTCloudSecurityGroupSync(cts, cloudMap, dsMap, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getDatasFromTCloudForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromTCloudForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncTCloudDiff, error) {

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	offset := 0
	datasCloud := []*vpc.SecurityGroup{}
	for {
		opt := &types.TCloudSecurityGroupListOption{
			Region: req.Region,
			Page:   &typcore.TCloudPage{Offset: uint64(offset), Limit: uint64(typcore.TCloudQueryLimit)},
		}
		datas, err := client.ListSecurityGroup(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt,
				cts.Kit.Rid)
			return nil, err
		}
		offset += len(datas)
		datasCloud = append(datasCloud, datas...)
		if len(datas) == 0 || uint(len(datas)) < typcore.TCloudQueryLimit {
			break
		}
	}

	cloudMap := make(map[string]*proto.SecurityGroupSyncTCloudDiff)
	for _, data := range datasCloud {
		sg := new(proto.SecurityGroupSyncTCloudDiff)
		sg.SecurityGroup = data
		cloudMap[*data.SecurityGroupId] = sg
	}

	return cloudMap, nil
}

// diffTCloudSecurityGroupSync make tcloud and data-service diff, process resources according to diff
func (g *securityGroup) diffTCloudSecurityGroupSync(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncTCloudDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, req *proto.SecurityGroupSyncReq) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup delete operate")
		err := g.diffSecurityGroupSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("sync delete tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffTCloudSGRuleSyncDelete(cts, deleteCloudIDs, dsMap)
		if err != nil {
			logs.Errorf("sync delete tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup update operate")
		err := g.diffTCloudSecurityGroupSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("sync update tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffTCloudSGRuleSyncUpdate(cts, updateCloudIDs, req, dsMap)
		if err != nil {
			logs.Errorf("sync update tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup add operate")
		ids, err := g.diffTCloudSecurityGroupSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("sync add tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffTCloudSGRuleSyncAdd(cts, ids, req)
		if err != nil {
			logs.Errorf("sync add tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffSecurityGroupSyncAdd for add
func (g *securityGroup) diffTCloudSecurityGroupSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncTCloudDiff,
	req *proto.SecurityGroupSyncReq, addCloudIDs []string) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.SecurityGroupId,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      *cloudMap[id].SecurityGroup.SecurityGroupName,
			Memo:      cloudMap[id].SecurityGroup.SecurityGroupDesc,
			AccountID: req.AccountID,
			Extension: &corecloud.TCloudSecurityGroupExtension{
				CloudProjectID: cloudMap[id].SecurityGroup.ProjectId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return nil, nil
	}

	results, err := g.dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffSecurityGroupSyncUpdate for update
func (g *securityGroup) diffTCloudSecurityGroupSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncTCloudDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, updateCloudIDs []string) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.SecurityGroupName == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.SecurityGroupDesc == dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.SecurityGroupName,
			Memo: cloudMap[id].SecurityGroup.SecurityGroupDesc,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) <= 0 {
		return nil
	}

	if err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {
		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}
