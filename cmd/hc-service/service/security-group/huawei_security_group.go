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
	"fmt"

	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/adaptor/types"
	typecvm "hcm/pkg/adaptor/types/cvm"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
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

	opt := &securitygroup.HuaWeiCreateOption{
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
				// Tags:        core.NewTagMap(req.Tags...),
				MgmtType:    req.MgmtType,
				MgmtBizID:   req.MgmtBizID,
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				UsageBizIds: req.UsageBizIds,
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

// HuaWeiSecurityGroupAssociateCvm ...
func (g *securityGroup) HuaWeiSecurityGroupAssociateCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, cvm, err := g.getSecurityGroupAndCvm(cts.Kit, req.SecurityGroupID, req.CvmID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.HuaWeiAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to huawei security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	// create security group cvm rels in db
	err = g.createSGCommonRelsForHuawei(cts.Kit, client, sg.Region, cvm)
	if err != nil {
		logs.Errorf("create security group common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

func (g *securityGroup) createSGCommonRelsForHuawei(kt *kit.Kit, client *huawei.HuaWei, region string,
	cvm *corecvm.BaseCvm) error {

	huaweiCvmFromCloud, err := g.listHuaweiCvmFromCloud(kt, client, region, cvm.CloudID)
	if err != nil {
		logs.Errorf("request adaptor to list cvm from cloud failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sgCloudIDs := make([]string, 0, len(huaweiCvmFromCloud[0].SecurityGroups))
	for _, cur := range huaweiCvmFromCloud[0].SecurityGroups {
		sgCloudIDs = append(sgCloudIDs, cur.Id)
	}

	sgCloudIDToIDMap, err := g.getSecurityGroupMapByCloudIDs(kt, enumor.HuaWei, region, sgCloudIDs)
	if err != nil {
		logs.Errorf("get security group map by cloud ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sgIDs := make([]string, 0, len(sgCloudIDs))
	for _, sgCloudID := range sgCloudIDs {
		sgID, ok := sgCloudIDToIDMap[sgCloudID]
		if !ok {
			logs.Errorf("cloud id(%s) not found in security group map, rid: %s", sgCloudID, kt.Rid)
			return fmt.Errorf("cloud id(%s) not found in security group map", sgCloudID)
		}
		sgIDs = append(sgIDs, sgID)
	}

	err = g.createSGCommonRels(kt, enumor.HuaWei, enumor.CvmCloudResType, cvm.ID, sgIDs)
	if err != nil {
		logs.Errorf("create security group common rels failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (g *securityGroup) listHuaweiCvmFromCloud(kt *kit.Kit, client *huawei.HuaWei, region, cvmCloudID string) (
	[]typecvm.HuaWeiCvm, error) {

	listOpt := &typecvm.HuaWeiListOption{
		Region:   region,
		CloudIDs: []string{cvmCloudID},
	}
	cvms, err := client.ListCvm(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}
	if len(cvms) == 0 {
		return nil, fmt.Errorf("cvm(%s) not found from cloud", cvmCloudID)
	}
	return cvms, nil
}

// HuaWeiSecurityGroupDisassociateCvm ...
func (g *securityGroup) HuaWeiSecurityGroupDisassociateCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, cvm, err := g.getSecurityGroupAndCvm(cts.Kit, req.SecurityGroupID, req.CvmID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	listCvmOpt := &typecvm.HuaWeiListOption{
		Region:   sg.Region,
		CloudIDs: []string{cvm.CloudID},
	}
	cvms, err := client.ListCvm(cts.Kit, listCvmOpt)
	if err != nil {
		logs.Errorf("request adaptor to list cvm failed, err: %v, opt: %v, rid: %s", err, listCvmOpt, cts.Kit)
		return nil, err
	}

	if cvms == nil || len(cvms) == 0 {
		return nil, errf.New(errf.RecordNotFound, "cvm not found from cloud")
	}

	cvmCloud := cvms
	if len(cvmCloud[0].SecurityGroups) <= 1 {
		return nil, errors.New("the last security group of the cvm is not allowed to disassociate")
	}

	opt := &securitygroup.HuaWeiAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to huawei security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	// delete security group cvm rels in db
	deleteReq := buildSGCommonRelDeleteReq(
		enumor.HuaWei, req.CvmID, []string{req.SecurityGroupID}, enumor.CvmCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
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

	opt := &securitygroup.HuaWeiDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	// delete security group in db
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

	opt := &securitygroup.HuaWeiUpdateOption{
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

// HuaweiListSecurityGroupStatistic result a list of *proto.HuaweiListSecurityGroupStatisticItem.
func (g *securityGroup) HuaweiListSecurityGroupStatistic(cts *rest.Contexts) (any, error) {
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
			return nil, fmt.Errorf("huawei security group: %s not found", sgID)
		}
		cloudIDToSgIDMap[sg.CloudID] = sgID
	}

	ports, err := g.listHuaweiPorts(cts.Kit, req.Region, req.AccountID, cloudIDToSgIDMap)
	if err != nil {
		logs.Errorf("list ports failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	sgIDToResourceCountMap, err := g.countHuaweiSecurityGroupStatistic(cts.Kit, ports, cloudIDToSgIDMap)
	if err != nil {
		logs.Errorf("count huawei security group statistic failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// Convert the map to the response format
	return resCountMapToSGStatisticResp(sgIDToResourceCountMap), nil
}

// countHuaweiSecurityGroupStatistic count huawei security group statistic.
func (g *securityGroup) countHuaweiSecurityGroupStatistic(kt *kit.Kit, ports []model.Port,
	cloudIDToSgIDMap map[string]string) (map[string]map[string]int64, error) {

	sgIDToResourceCountMap := make(map[string]map[string]int64)
	for _, sgID := range cloudIDToSgIDMap {
		sgIDToResourceCountMap[sgID] = make(map[string]int64)
	}
	for _, port := range ports {
		for _, cloudID := range port.SecurityGroups {
			sgID, ok := cloudIDToSgIDMap[cloudID]
			if !ok {
				logs.Warnf("cloudID: %s not found in cloudIDToSgIDMap, vendor: %s, rid: %s",
					cloudID, enumor.HuaWei, kt.Rid)
				continue
			}
			sgIDToResourceCountMap[sgID][port.DeviceOwner.Value()]++
		}
	}
	return sgIDToResourceCountMap, nil
}

// listHuaweiPorts list huawei ports by security group cloud IDs from cloud.
func (g *securityGroup) listHuaweiPorts(kt *kit.Kit, region, accountID string, cloudIDToSgIDMap map[string]string) (
	[]model.Port, error) {

	client, err := g.ad.HuaWei(kt, accountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaweiListPortOption{
		Region:           region,
		SecurityGroupIDs: converter.MapKeyToSlice(cloudIDToSgIDMap),
	}

	result := make([]model.Port, 0)
	for {
		resp, err := client.ListPorts(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to huawei security group statistic failed, err: %v, opt: %v, rid: %s",
				err, opt, kt.Rid)
			return nil, err
		}
		if len(resp) == 0 {
			break
		}
		result = append(result, resp...)
		opt.Marker = resp[len(resp)-1].Id
	}
	return result, nil
}
