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

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/pkg/adaptor/tcloud"
	typecvm "hcm/pkg/adaptor/types/cvm"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

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

	opt := &securitygroup.TCloudCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
		Tags:        req.Tags,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud security group failed, err: %v, opt: %+v, rid: %s",
			err, opt, cts.Kit.Rid)
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
				Tags:        core.NewTagMap(req.Tags...),
				MgmtType:    req.MgmtType,
				MgmtBizID:   req.MgmtBizID,
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				UsageBizIds: req.UsageBizIds,
			}},
	}
	result, err := g.dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, req: %v, rid: %s",
			err, req, cts.Kit.Rid)
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

	opt := &securitygroup.TCloudDeleteOption{
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
	if err = g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
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

	opt := &securitygroup.TCloudUpdateOption{
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
	if err = g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// createSGCommonRelsForTCloud creates security group common relations for TCloud.
func (g *securityGroup) createSGCommonRelsForTCloud(kt *kit.Kit, client tcloud.TCloud, region string,
	cvmCloudIDToIDMap map[string]string) error {

	cloudCvms, err := g.listTCloudCvmFromCloud(kt, client, region, converter.MapKeyToSlice(cvmCloudIDToIDMap))
	if err != nil {
		logs.Errorf("list cvm from cloud failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sgCloudIDs := make([]string, 0)
	for _, one := range cloudCvms {
		sgCloudIDs = append(sgCloudIDs, converter.PtrToSlice(one.SecurityGroupIds)...)
	}

	sgCloudIDToIDMap, err := g.getSecurityGroupMapByCloudIDs(kt, enumor.TCloud, region, sgCloudIDs)
	if err != nil {
		logs.Errorf("get security group map by cloud ids failed, err: %v, cloudIDs: %v, rid: %s",
			err, sgCloudIDs, kt.Rid)
		return err
	}

	for _, one := range cloudCvms {
		cvmID, ok := cvmCloudIDToIDMap[converter.PtrToVal(one.InstanceId)]
		if !ok {
			logs.Errorf("cvm cloud id to id not found, cvmID: %s, rid: %s", converter.PtrToVal(one.InstanceId), kt.Rid)
			return fmt.Errorf("cvm cloud id to id not found, cvmID: %s", converter.PtrToVal(one.InstanceId))
		}

		sgIDs := make([]string, 0, len(one.SecurityGroupIds))
		for _, sgCloudID := range converter.PtrToSlice(one.SecurityGroupIds) {
			sgID, ok := sgCloudIDToIDMap[sgCloudID]
			if !ok {
				logs.Errorf("cloud id(%s) not found in security group map, rid: %s", sgCloudID, kt.Rid)
				return fmt.Errorf("cloud id(%s) not found in security group map", sgCloudID)
			}
			sgIDs = append(sgIDs, sgID)
		}

		err = g.createSGCommonRels(kt, enumor.TCloud, enumor.CvmCloudResType, cvmID, sgIDs)
		if err != nil {
			// 不抛出err, 尽最大努力交付
			logs.Errorf("create sg common rels failed, err: %v, cvmID: %s, sgIDs: %v, rid: %s",
				err, cvmID, converter.MapValueToSlice(sgCloudIDToIDMap), kt.Rid)
		}
	}

	return nil
}

// listTCloudCvmFromCloud lists TCloud CVMs from the cloud using the provided client and region.
func (g *securityGroup) listTCloudCvmFromCloud(kt *kit.Kit, client tcloud.TCloud, region string, cvmCloudIDs []string) (
	[]typecvm.TCloudCvm, error) {

	listOpt := &typecvm.TCloudListOption{
		CloudIDs: cvmCloudIDs,
		Region:   region,
	}
	cloudCvms, err := client.ListCvm(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, nil
	}

	return cloudCvms, err
}

// TCloudListSecurityGroupStatistic ...
func (g *securityGroup) TCloudListSecurityGroupStatistic(cts *rest.Contexts) (any, error) {
	req := new(proto.ListSecurityGroupStatisticReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgMap, err := g.getSecurityGroupMap(cts.Kit, req.SecurityGroupIDs)
	if err != nil {
		logs.Errorf("get security group map failed, sgID: %v, err: %v, rid: %s", req.SecurityGroupIDs, err, cts.Kit.Rid)
		return nil, err
	}

	cloudIDToSgIDMap := make(map[string]string)
	for _, sgID := range req.SecurityGroupIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			return nil, fmt.Errorf("tcloud security group: %s not found", sgID)
		}
		cloudIDToSgIDMap[sg.CloudID] = sgID
	}

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.TCloudListOption{
		Region:   req.Region,
		CloudIDs: converter.MapKeyToSlice(cloudIDToSgIDMap),
	}
	resp, err := client.DescribeSGAssociationStatistics(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to tcloud security group statistic failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	sgIDToResourceCountMap := make(map[string]map[string]int64)
	for _, one := range resp {
		sgID := cloudIDToSgIDMap[converter.PtrToVal(one.SecurityGroupId)]
		sgIDToResourceCountMap[sgID] = tcloudSGAssociateStatisticToResourceCountMap(one)
	}

	return resCountMapToSGStatisticResp(sgIDToResourceCountMap), nil
}

const (
	tcloudStatisticResTypeCVM = "CVM"
	tcloudStatisticResTypeCDB = "CDB"
	tcloudStatisticResTypeENI = "ENI"
	tcloudStatisticResTypeSG  = "SG"
	tcloudStatisticResTypeCLB = "CLB"
)

// defaultResourceCountMap initializes a map with default values for TCloud security group statistics.
func tcloudSGAssociateStatisticToResourceCountMap(
	statistic securitygroup.TCloudSecurityGroupAssociationStatistic) map[string]int64 {

	return map[string]int64{
		tcloudStatisticResTypeCVM: int64(converter.PtrToVal(statistic.CVM)),
		tcloudStatisticResTypeCDB: int64(converter.PtrToVal(statistic.CDB)),
		tcloudStatisticResTypeENI: int64(converter.PtrToVal(statistic.ENI)),
		tcloudStatisticResTypeSG:  int64(converter.PtrToVal(statistic.SG)),
		tcloudStatisticResTypeCLB: int64(converter.PtrToVal(statistic.CLB)),
	}
}

// TCloudCloneSecurityGroup ...
func (g *securityGroup) TCloudCloneSecurityGroup(cts *rest.Contexts) (any, error) {

	req := new(proto.TCloudSecurityGroupCloneReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	securityGroupMap, err := g.getSecurityGroupMap(cts.Kit, []string{req.SecurityGroupID})
	if err != nil {
		logs.Errorf("get security group map failed, sgID: %s, err: %v, rid: %s", req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}
	sg, ok := securityGroupMap[req.SecurityGroupID]
	if !ok {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", req.SecurityGroupID)
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	// 如果目标地域为空，则默认指定为源安全组的地域
	if req.TargetRegion == "" {
		req.TargetRegion = sg.Region
	}
	opt := &securitygroup.TCloudSecurityGroupCloneOption{
		Region:          req.TargetRegion,
		SecurityGroupID: sg.CloudID,
		Tags:            req.Tags,
		RemoteRegion:    sg.Region,
		GroupName:       req.GroupName,
	}
	newSecurityGroup, err := client.CloneSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to clone tcloud security group failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}
	sgID, err := g.createSecurityGroupForData(cts.Kit, req, sg.AccountID, newSecurityGroup)
	if err != nil {
		logs.Errorf("create security group for data-service failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	syncParam := &synctcloud.SyncBaseParams{AccountID: sg.AccountID, Region: req.TargetRegion, CloudIDs: []string{sgID}}
	_, syncErr := g.syncSGRule(cts.Kit, syncParam)
	if syncErr != nil {
		logs.Warnf("sync security group rule failed, err: %v, sg: %s, rid: %s", syncErr, sgID, cts.Kit.Rid)
	}
	return core.CreateResult{ID: sgID}, nil
}

// createSecurityGroupForData creates a security group in the data service.
func (g *securityGroup) createSecurityGroupForData(kt *kit.Kit, req *proto.TCloudSecurityGroupCloneReq,
	accountID string, sg *vpc.SecurityGroup) (string, error) {

	tags := make([]core.TagPair, 0, len(sg.TagSet))
	for _, tag := range sg.TagSet {
		tags = append(tags, core.TagPair{
			Key:   converter.PtrToVal(tag.Key),
			Value: converter.PtrToVal(tag.Value),
		})
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{
			{
				CloudID:   *sg.SecurityGroupId,
				BkBizID:   req.ManagementBizID,
				Region:    req.TargetRegion,
				Name:      *sg.SecurityGroupName,
				Memo:      sg.SecurityGroupDesc,
				AccountID: accountID,
				Extension: &corecloud.TCloudSecurityGroupExtension{
					CloudProjectID: sg.ProjectId,
				},
				Tags:        core.NewTagMap(tags...),
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				MgmtType:    enumor.MgmtTypeBiz,
				MgmtBizID:   req.ManagementBizID,
				UsageBizIds: []int64{req.ManagementBizID},
			}},
	}
	result, err := g.dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, req: %v, rid: %s",
			err, createReq, kt.Rid)
		return "", err
	}
	return result.IDs[0], nil
}
