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

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// CreateHuaWeiSecurityGroup create huawei security group.
func (g *securityGroup) CreateHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.HuaWeiSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSecurityGroupCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to CreateSecurityGroup failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.HuaWeiSecurityGroupExtension]{
			{
				CloudID:   sg.Id,
				BkBizID:   req.BkBizID,
				Region:    req.Region,
				Name:      sg.Name,
				Memo:      &sg.Description,
				AccountID: req.AccountID,
				Extension: &corecloud.HuaWeiSecurityGroupExtension{
					CloudProjectID:           sg.ProjectId,
					CloudEnterpriseProjectID: sg.EnterpriseProjectId,
				},
			},
		},
	}
	result, err := g.dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteHuaWeiSecurityGroup delete huawei security group.
func (g *securityGroup) DeleteHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.SecurityGroupDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateHuaWeiSecurityGroup update huawei security group.
func (g *securityGroup) UpdateHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
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

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSecurityGroupUpdateOption{
		CloudID:     sg.CloudID,
		Region:      sg.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	if err := client.UpdateSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.HuaWeiSecurityGroupExtension]{
			{
				ID:   sg.ID,
				Name: req.Name,
				Memo: req.Memo,
			},
		},
	}
	if err := g.dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncHuaWeiSecurityGroup sync security group from huawei to hcm.
func (g *securityGroup) SyncHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	cloudMap, err := g.getDatasFromHuaWeiForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := g.getDatasFromDSForSecurityGroupSync(cts, req)
	if err != nil {
		return nil, err
	}

	err = g.diffHWSecurityGroupSync(cts, cloudMap, dsMap, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getDatasFromHuaWeiForSecurityGroupSync get datas from cloud
func (g *securityGroup) getDatasFromHuaWeiForSecurityGroupSync(cts *rest.Contexts,
	req *proto.SecurityGroupSyncReq) (map[string]*proto.SecurityGroupSyncHuaWeiDiff, error) {

	client, err := g.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	datasCloud := []model.SecurityGroup{}
	limit := int32(typcore.HuaweiQueryLimit)
	var marker *string = nil
	for {
		opt := &types.HuaWeiSecurityGroupListOption{
			Region: req.Region,
			Page:   &typcore.HuaweiPage{Limit: &limit, Marker: marker},
		}
		datas, pageInfo, err := client.ListSecurityGroup(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
				cts.Kit.Rid)
			return nil, err
		}
		datasCloud = append(datasCloud, *datas...)
		marker = pageInfo.NextMarker
		if len(*datas) == 0 || pageInfo.NextMarker == nil {
			break
		}
	}

	cloudMap := make(map[string]*proto.SecurityGroupSyncHuaWeiDiff)
	for _, data := range datasCloud {
		sg := new(proto.SecurityGroupSyncHuaWeiDiff)
		sg.SecurityGroup = data
		cloudMap[data.Id] = sg
	}

	return cloudMap, nil
}

// diffHWSecurityGroupSync make huawei and data-service diff, process resources according to diff
func (g *securityGroup) diffHWSecurityGroupSync(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncHuaWeiDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, req *proto.SecurityGroupSyncReq) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup delete operate")
		err := g.diffSecurityGroupSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("sync delete huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffHuaWeiSGRuleSyncDelete(cts, deleteCloudIDs, dsMap)
		if err != nil {
			logs.Errorf("sync delete huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup update operate")
		err := g.diffHWSecurityGroupSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("sync update huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffHuaWeiSGRuleSyncUpdate(cts, updateCloudIDs, req, dsMap)
		if err != nil {
			logs.Errorf("sync update huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup add operate")
		ids, err := g.diffHWSecurityGroupSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("sync add huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		err = g.diffHuaWeiSGRuleSyncAdd(cts, ids, req)
		if err != nil {
			logs.Errorf("sync add huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffSecurityGroupSyncAdd for add
func (g *securityGroup) diffHWSecurityGroupSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncHuaWeiDiff,
	req *proto.SecurityGroupSyncReq, addCloudIDs []string) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.HuaWeiSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.HuaWeiSecurityGroupExtension]{
			CloudID:   cloudMap[id].SecurityGroup.Id,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      cloudMap[id].SecurityGroup.Name,
			Memo:      &cloudMap[id].SecurityGroup.Description,
			AccountID: req.AccountID,
			Extension: &corecloud.HuaWeiSecurityGroupExtension{
				CloudProjectID:           cloudMap[id].SecurityGroup.ProjectId,
				CloudEnterpriseProjectID: cloudMap[id].SecurityGroup.EnterpriseProjectId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return nil, nil
	}

	ids, err := g.dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return ids.IDs, nil
}

// diffSecurityGroupSyncUpdate for update
func (g *securityGroup) diffHWSecurityGroupSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.SecurityGroupSyncHuaWeiDiff,
	dsMap map[string]*proto.SecurityGroupSyncDS, updateCloudIDs []string) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.HuaWeiSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if cloudMap[id].SecurityGroup.Name == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.Description == *dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.HuaWeiSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: cloudMap[id].SecurityGroup.Name,
			Memo: &cloudMap[id].SecurityGroup.Description,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) <= 0 {
		return nil
	}

	if err := g.dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {
		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}
