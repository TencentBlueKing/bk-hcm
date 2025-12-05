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

	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	tziyan "hcm/pkg/adaptor/tcloud-ziyan"
	typecvm "hcm/pkg/adaptor/types/cvm"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	adptsg "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/core/ziyan"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateZiyanSecurityGroup create tcloud ziyan security group.
func (g *securityGroup) CreateZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
		Tags:        req.Tags,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud ziyan security group failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{{
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
	result, err := g.dataCli.TCloudZiyan.SecurityGroup.BatchCreateSecurityGroup(cts.Kit, createReq)
	if err != nil {

		bpaasSN := errf.GetBPaasSNFromErr(err)
		if len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.CreateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            req.AccountID,
				BkBizID:              req.BkBizID,
				Region:               req.Region,
				SecurityGroupCloudID: converter.PtrToVal(sg.SecurityGroupId),
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}

		logs.Errorf("request dataservice to create tcloud ziyan security group failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

// ZiyanSecurityGroupAssociateCvm ...
func (g *securityGroup) ZiyanSecurityGroupAssociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmAssociate(cts.Kit, opt); err != nil {

		if bpaasSN := errf.GetBPaasSNFromErr(err); len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.AssociateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupCloudID: sg.CloudID,
				SecurityGroupID:      sg.ID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}

		logs.Errorf("request adaptor to tcloud security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	err = g.createSGCommonRelsForTCloudZiyan(cts.Kit, client, sg.Region, map[string]string{cvm.CloudID: cvm.ID})
	if err != nil {
		logs.Errorf("create security group cvm rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// TODO: 同步主机数据

	return nil, nil
}

// ZiyanSecurityGroupDisassociateCvm ...
func (g *securityGroup) ZiyanSecurityGroupDisassociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	listCvmOpt := &typecvm.TCloudListOption{
		Region:   sg.Region,
		CloudIDs: []string{cvm.CloudID},
	}
	cvms, err := client.ListCvm(cts.Kit, listCvmOpt)
	if err != nil {
		logs.Errorf("request adaptor to list cvm failed, err: %v, opt: %v, rid: %s", err, listCvmOpt, cts.Kit)
		return nil, err
	}

	if len(cvms) == 0 {
		return nil, errf.New(errf.RecordNotFound, "cvm not found from cloud")
	}

	if len(cvms[0].SecurityGroupIds) <= 1 {
		return nil, errors.New("the last security group of the cvm is not allowed to disassociate")
	}

	opt := &adptsg.TCloudAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmDisassociate(cts.Kit, opt); err != nil {

		if bpaasSN := errf.GetBPaasSNFromErr(err); len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.DisassociateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupID:      sg.ID,
				SecurityGroupCloudID: sg.CloudID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}

		logs.Errorf("request adaptor to tcloud security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(enumor.Ziyan, req.CvmID, []string{req.SecurityGroupID},
		enumor.CvmCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

// DeleteZiyanSecurityGroup delete tcloud ziyan security group.
func (g *securityGroup) DeleteZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.TCloudZiyan.SecurityGroup.GetSecurityGroup(cts.Kit, id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		bpaasSN := errf.GetBPaasSNFromErr(err)
		if len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.DeleteSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupID:      sg.ID,
				SecurityGroupCloudID: sg.CloudID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}

		logs.Errorf("request adaptor to delete tcloud ziyan security group failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateZiyanSecurityGroup update tcloud ziyan security group.
func (g *securityGroup) UpdateZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
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

	sg, err := g.dataCli.TCloudZiyan.SecurityGroup.GetSecurityGroup(cts.Kit, id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudUpdateOption{
		CloudID:     sg.CloudID,
		Region:      sg.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	if err := client.UpdateSecurityGroup(cts.Kit, opt); err != nil {

		bpaasSN := errf.GetBPaasSNFromErr(err)
		if len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.UpdateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupID:      sg.ID,
				SecurityGroupCloudID: sg.CloudID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}

		logs.Errorf("request adaptor to UpdateSecurityGroup failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{{
			ID:   sg.ID,
			Name: req.Name,
			Memo: req.Memo,
		}},
	}
	if err := g.dataCli.TCloudZiyan.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit, updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ZiyanSecurityGroupAssociateLoadBalancer ...
func (g *securityGroup) ZiyanSecurityGroupAssociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	req := new(hclb.TCloudSetLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据LbID查询负载均衡基本信息
	lbInfo, sgComList, err := g.getLoadBalancerInfoAndSGComRels(cts.Kit, req.LbID)
	if err != nil {
		return nil, err
	}

	sgCloudIDs, sgComReq, err := g.getUpsertSGIDsParams(cts.Kit, enumor.Ziyan, req, sgComList)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudSetClbSecurityGroupOption{
		Region:         lbInfo.Region,
		LoadBalancerID: lbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetLoadBalancerSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate lb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	if err = g.dataCli.Global.SGCommonRel.BatchUpsertSgCommonRels(cts.Kit, sgComReq); err != nil {
		logs.Errorf("request dataservice upsert security group lb rels failed, err: %v, req: %+v, rid: %s",
			err, sgComReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ZiyanSecurityGroupDisassociateLoadBalancer ...
func (g *securityGroup) ZiyanSecurityGroupDisassociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	req := new(hclb.TCloudDisAssociateLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据LbID查询负载均衡基本信息
	lbInfo, sgComList, err := g.getLoadBalancerInfoAndSGComRels(cts.Kit, req.LbID)
	if err != nil {
		return nil, err
	}

	allSGIDs := make([]string, 0)
	existSG := false
	for _, rel := range sgComList.Details {
		if rel.SecurityGroupID == req.SecurityGroupID {
			existSG = true
		}
		allSGIDs = append(allSGIDs, rel.SecurityGroupID)
	}
	if !existSG {
		return nil, errf.Newf(errf.RecordNotFound, "not found sg id: %s", req.SecurityGroupID)
	}

	sgMap, err := g.getSecurityGroupMap(cts.Kit, allSGIDs)
	if err != nil {
		return nil, err
	}

	// 安全组的云端ID数组
	sgCloudIDs := make([]string, 0)
	for _, sgID := range allSGIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			continue
		}
		if sg.ID == req.SecurityGroupID {
			// 跳过用户需要解绑的安全组ID
			continue
		}
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudSetClbSecurityGroupOption{
		Region:         lbInfo.Region,
		LoadBalancerID: lbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetLoadBalancerSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group disAssociate lb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(
		enumor.TCloudZiyan, req.LbID, []string{req.SecurityGroupID}, enumor.LoadBalancerCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice tcloud delete security group lb rels failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ZiyanSGBatchAssociateCvm 根据cvmID 绑定安全组
func (g *securityGroup) ZiyanSGBatchAssociateCvm(cts *rest.Contexts) (any, error) {

	req := new(proto.SecurityGroupBatchAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.getSecurityGroupByID(cts, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, req.SecurityGroupID, cts.Kit.Rid)
		return nil, err
	}
	cvmList, err := g.getCvms(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}
	cvmCloudIDToIDMap := make(map[string]string, len(req.CvmIDs))
	for _, baseCvm := range cvmList {
		cvmCloudIDToIDMap[baseCvm.CloudID] = baseCvm.ID
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          converter.MapKeyToSlice(cvmCloudIDToIDMap),
	}
	if err = client.SecurityGroupCvmBatchAssociate(cts.Kit, opt); err != nil {
		bpaasSN := errf.GetBPaasSNFromErr(err)
		if len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.AssociateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupID:      sg.ID,
				SecurityGroupCloudID: sg.CloudID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}
		logs.Errorf("request adaptor to tcloud ziyan security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	err = g.createSGCommonRelsForTCloudZiyan(cts.Kit, client, sg.Region, cvmCloudIDToIDMap)
	if err != nil {
		logs.Errorf("create sg common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) createSGCommonRelsForTCloudZiyan(kt *kit.Kit, client tziyan.TCloudZiyan, region string,
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

	sgCloudIDToIDMap, err := g.getSecurityGroupMapByCloudIDs(kt, enumor.TCloudZiyan, region, sgCloudIDs)
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

		err = g.createSGCommonRels(kt, enumor.Ziyan, enumor.CvmCloudResType, cvmID, sgIDs)
		if err != nil {
			// 不抛出err, 尽最大努力交付
			logs.Errorf("create sg common rels failed, err: %v, cvmID: %s, sgIDs: %v, rid: %s",
				err, cvmID, converter.MapValueToSlice(sgCloudIDToIDMap), kt.Rid)
		}
	}

	return nil
}

// ZiyanListSecurityGroupStatistic ...
func (g *securityGroup) ZiyanListSecurityGroupStatistic(cts *rest.Contexts) (any, error) {
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
			logs.Errorf("security group: %s not found, rid: %s", sgID, cts.Kit.Rid)
			return nil, fmt.Errorf("tcloud-ziyan security group: %s not found", sgID)
		}
		cloudIDToSgIDMap[sg.CloudID] = sgID
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudListOption{
		Region:   req.Region,
		CloudIDs: converter.MapKeyToSlice(cloudIDToSgIDMap),
	}
	resp, err := client.DescribeSGAssociationStatistics(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to tcloud-ziyan security group statistic failed, err: %v, opt: %v, rid: %s",
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

// ZiyanSGBatchDisassociateCvm  根据cvm云id 解绑安全组
func (g *securityGroup) ZiyanSGBatchDisassociateCvm(cts *rest.Contexts) (any, error) {
	req := new(proto.SecurityGroupBatchAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.getSecurityGroupByID(cts, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, req.SecurityGroupID, cts.Kit.Rid)
		return nil, err
	}
	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	cvmList, err := g.getCvms(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}
	cloudCvmIDs := make([]string, 0, len(req.CvmIDs))
	for _, baseCvm := range cvmList {
		cloudCvmIDs = append(cloudCvmIDs, baseCvm.CloudID)
	}
	opt := &adptsg.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          cloudCvmIDs,
	}
	if err = client.SecurityGroupCvmBatchDisassociate(cts.Kit, opt); err != nil {

		bpaasSN := errf.GetBPaasSNFromErr(err)
		if len(bpaasSN) > 0 {
			content := &coreziyan.BPaasApplicationContent{
				Action:               enumor.DisassociateSecurityGroup,
				SN:                   bpaasSN,
				AccountID:            sg.AccountID,
				BkBizID:              sg.BkBizID,
				Region:               sg.Region,
				SecurityGroupID:      sg.ID,
				SecurityGroupCloudID: sg.CloudID,
				Params:               opt,
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, content)
		}
		logs.Errorf("request adaptor to tcloud ziyan security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq, err := buildSGCommonRelDeleteReqForMultiResource(
		enumor.CvmCloudResType, req.SecurityGroupID, req.CvmIDs...)
	if err != nil {
		logs.Errorf("build sg cvm rel delete req failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// ZiyanCloneSecurityGroup ...
func (g *securityGroup) ZiyanCloneSecurityGroup(cts *rest.Contexts) (any, error) {

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

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	// 如果目标地域为空，则默认指定为源安全组的地域
	if req.TargetRegion == "" {
		req.TargetRegion = sg.Region
	}
	opt := &adptsg.TCloudSecurityGroupCloneOption{
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
	sgID, err := g.createTZiyanSecurityGroupForData(cts.Kit, req, sg.AccountID, newSecurityGroup)
	if err != nil {
		logs.Errorf("create security group for data-service failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	syncParam := &ziyan.SyncBaseParams{AccountID: sg.AccountID, Region: req.TargetRegion, CloudIDs: []string{sgID}}
	_, syncErr := g.syncZiyanSGRule(cts.Kit, syncParam)
	if syncErr != nil {
		logs.Warnf("sync security group rule failed, err: %v, sg: %s, rid: %s", syncErr, sgID, cts.Kit.Rid)
	}
	return core.CreateResult{ID: sgID}, nil
}

func (g *securityGroup) createTZiyanSecurityGroupForData(kt *kit.Kit, req *proto.TCloudSecurityGroupCloneReq,
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
	result, err := g.dataCli.TCloudZiyan.SecurityGroup.BatchCreateSecurityGroup(kt, createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud-ziyan security group failed, err: %v, req: %v, rid: %s",
			err, createReq, kt.Rid)
		return "", err
	}
	return result.IDs[0], nil
}

// ZiyanSGBatchAssociateLoadBalancers 绑定一个安全组到多个负载均衡实例
func (g *securityGroup) ZiyanSGBatchAssociateLoadBalancers(cts *rest.Contexts) (any, error) {
	req := new(proto.SecurityGroupBatchOperateLoadBalancerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgMap, err := g.getSecurityGroupMap(cts.Kit, []string{req.SecurityGroupID})
	if err != nil {
		logs.Errorf("get security group map failed, sgID: %s, err: %v, rid: %s", req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}
	sg, ok := sgMap[req.SecurityGroupID]
	if !ok {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", req.SecurityGroupID)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	_, clbCloudIDs, sgComReq, err := g.getSecurityGroupIDRelReq(cts.Kit, req, sg, adptsg.AddSGOperationType)
	if err != nil {
		logs.Errorf("get security group id rel rel failed, sgID: %s, err: %v, rid: %s",
			req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}

	// 绑定一个安全组到多个负载均衡实例
	opt := &typelb.TCloudSetSecurityGroupForClbsOption{
		Region:          sg.Region,
		SecurityGroup:   sg.CloudID,
		OperationType:   adptsg.AddSGOperationType,
		LoadBalancerIDs: clbCloudIDs,
	}
	if _, err = client.SetSecurityGroupForLoadbalancers(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate clbs failed, err: %v, vendor: %s, opt: %+v, "+
			"rid: %s", err, sg.Vendor, converter.PtrToVal(opt), cts.Kit.Rid)
		return nil, err
	}

	if len(sgComReq.Rels) > 0 {
		if err = g.dataCli.Global.SGCommonRel.BatchUpsertSgCommonRels(cts.Kit, sgComReq); err != nil {
			logs.Errorf("request dataservice upsert security group clb rels failed, err: %v, vendor: %s, "+
				"req: %+v, rid: %s", err, sg.Vendor, sgComReq, cts.Kit.Rid)
			return nil, err
		}
	}
	return nil, nil
}

// ZiyanSGBatchDisassociateLoadBalancers 解绑一个安全组到多个负载均衡实例
func (g *securityGroup) ZiyanSGBatchDisassociateLoadBalancers(cts *rest.Contexts) (any, error) {
	req := new(proto.SecurityGroupBatchOperateLoadBalancerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgMap, err := g.getSecurityGroupMap(cts.Kit, []string{req.SecurityGroupID})
	if err != nil {
		logs.Errorf("get security group map failed, sgID: %s, err: %v, rid: %s", req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}
	sg, ok := sgMap[req.SecurityGroupID]
	if !ok {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", req.SecurityGroupID)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	clbIDs, clbCloudIDs, _, err := g.getSecurityGroupIDRelReq(cts.Kit, req, sg, adptsg.DelSGOperationType)
	if err != nil {
		logs.Errorf("get security group id rel rel failed, sgID: %s, err: %v, rid: %s",
			req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}

	// 解绑一个安全组到多个负载均衡实例
	opt := &typelb.TCloudSetSecurityGroupForClbsOption{
		Region:          sg.Region,
		SecurityGroup:   sg.CloudID,
		OperationType:   adptsg.DelSGOperationType,
		LoadBalancerIDs: clbCloudIDs,
	}
	if _, err = client.SetSecurityGroupForLoadbalancers(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud-ziyan security group associate clbs failed, err: %v, vendor: %s, "+
			"opt: %+v, rid: %s", err, sg.Vendor, converter.PtrToVal(opt), cts.Kit.Rid)
		return nil, err
	}

	if len(clbIDs) > 0 {
		deleteReq := buildSGCommonRelDeleteClbsReq(
			sg.Vendor, req.SecurityGroupID, clbIDs, enumor.LoadBalancerCloudResType)
		if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
			logs.Errorf("request dataservice delete security group lb rels failed, err: %v, vendor: %s, "+
				"sgID: %s, clbIDs: %v, rid: %s", err, sg.Vendor, req.SecurityGroupID, clbIDs, cts.Kit.Rid)
			return nil, err
		}
	}
	return nil, nil
}
