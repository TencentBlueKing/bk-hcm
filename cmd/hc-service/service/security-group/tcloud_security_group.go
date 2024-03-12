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

	"hcm/pkg/adaptor/types/clb"
	typecvm "hcm/pkg/adaptor/types/cvm"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	coreclb "hcm/pkg/api/core/cloud/clb"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	hcclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
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

// TCloudSecurityGroupAssociateCvm ...
func (g *securityGroup) TCloudSecurityGroupAssociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.TCloudAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SGCvmRelBatchCreateReq{
		Rels: []protocloud.SGCvmRelCreate{
			{
				SecurityGroupID: req.SecurityGroupID,
				CvmID:           req.CvmID,
			},
		},
	}
	if err = g.dataCli.Global.SGCvmRel.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
		logs.Errorf("request dataservice create security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, createReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

// TCloudSecurityGroupDisassociateCvm ...
func (g *securityGroup) TCloudSecurityGroupDisassociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
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

	opt := &securitygroup.TCloudAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCvmRelDeleteReq(req.SecurityGroupID, req.CvmID)
	if err = g.dataCli.Global.SGCvmRel.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
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

// TCloudSecurityGroupAssociateClb ...
func (g *securityGroup) TCloudSecurityGroupAssociateClb(cts *rest.Contexts) (interface{}, error) {
	req := new(hcclb.TCloudSetClbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据ClbID查询负载均衡基本信息
	clbInfo, sgComList, err := g.getClbInfoAndSGComRels(cts.Kit, req.ClbID)
	if err != nil {
		return nil, err
	}

	sgCloudIDs, sgComReq, err := g.getUpsertSGIDsParams(cts.Kit, req, sgComList)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, clbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &clb.TCloudSetClbSecurityGroupOption{
		Region:         clbInfo.Region,
		LoadBalancerID: clbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetClbSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate clb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	if err = g.dataCli.Global.SGCommonRel.BatchUpsert(cts.Kit, sgComReq); err != nil {
		logs.Errorf("request dataservice upsert security group clb rels failed, err: %v, req: %+v, rid: %s",
			err, sgComReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getUpsertSGIDsParams(kt *kit.Kit, req *hcclb.TCloudSetClbSecurityGroupReq,
	sgComList *protocloud.SGCommonRelListResult) ([]string, *protocloud.SGCommonRelBatchUpsertReq, error) {

	newSGIDsMap := converter.StringSliceToMap(req.SecurityGroupIDs)
	allSGIDs := make([]string, 0)
	delSGIDs := make([]string, 0)
	for _, sg := range sgComList.Details {
		if _, ok := newSGIDsMap[sg.SecurityGroupID]; ok {
			delSGIDs = append(delSGIDs, sg.SecurityGroupID)
			continue
		}
		allSGIDs = append(allSGIDs, sg.SecurityGroupID)
	}

	for _, newSGID := range req.SecurityGroupIDs {
		allSGIDs = append(allSGIDs, newSGID)
	}

	sgComReq := &protocloud.SGCommonRelBatchUpsertReq{
		Rels: make([]protocloud.SGCommonRelCreate, 0, len(req.SecurityGroupIDs)),
	}
	if len(delSGIDs) > 0 {
		sgComReq.DeleteReq = buildSGCommonRelDeleteReq(
			enumor.TCloud, req.ClbID, delSGIDs, enumor.LoadBalancerCloudResType)
	}

	tmpPriority := int64(0)
	for _, newSGID := range allSGIDs {
		tmpPriority++
		sgComReq.Rels = append(sgComReq.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: newSGID,
			Vendor:          enumor.TCloud,
			ResID:           req.ClbID,
			ResType:         enumor.LoadBalancerCloudResType,
			Priority:        tmpPriority,
		})
	}

	sgMap, err := g.getSecurityGroupMap(kt, allSGIDs)
	if err != nil {
		return nil, nil, err
	}

	// 安全组的云端ID数组
	sgCloudIDs := make([]string, 0)
	for _, sgID := range allSGIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			continue
		}
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
	}
	if len(sgCloudIDs) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "cloud security group ids is empty")
	}

	return sgCloudIDs, sgComReq, nil
}

// TCloudSecurityGroupDisassociateClb ...
func (g *securityGroup) TCloudSecurityGroupDisassociateClb(cts *rest.Contexts) (interface{}, error) {
	req := new(hcclb.TCloudDisAssociateClbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据ClbID查询负载均衡基本信息
	clbInfo, sgComList, err := g.getClbInfoAndSGComRels(cts.Kit, req.ClbID)
	if err != nil {
		return nil, err
	}

	allSGIDs := make([]string, 0)
	existSG := false
	for _, sg := range sgComList.Details {
		if sg.SecurityGroupID == req.SecurityGroupID {
			existSG = true
			continue
		}
		allSGIDs = append(allSGIDs, sg.SecurityGroupID)
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
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
	}

	client, err := g.ad.TCloud(cts.Kit, clbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &clb.TCloudSetClbSecurityGroupOption{
		Region:         clbInfo.Region,
		LoadBalancerID: clbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetClbSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group disAssociate clb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(
		enumor.TCloud, req.ClbID, []string{req.SecurityGroupID}, enumor.LoadBalancerCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDelete(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice tcloud delete security group clb rels failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getClbInfoAndSGComRels(kt *kit.Kit, clbID string) (
	*coreclb.BaseClb, *protocloud.SGCommonRelListResult, error) {

	clbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", clbID),
		Page:   core.NewDefaultBasePage(),
	}
	clbList, err := g.dataCli.Global.LoadBalancer.ListClb(kt, clbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, id: %s, err: %v, rid: %s", clbID, err, kt.Rid)
		return nil, nil, err
	}

	if len(clbList.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "not found clb id: %s", clbID)
	}

	clbInfo := clbList.Details[0]
	// 查询目前绑定的安全组
	sgcomReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: clbInfo.Vendor,
				},
				&filter.AtomRule{
					Field: "res_id",
					Op:    filter.Equal.Factory(),
					Value: clbID,
				},
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: meta.LoadBalancer,
				},
			},
		},
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "priority", Order: "ASC"},
	}
	sgComList, err := g.dataCli.Global.SGCommonRel.List(kt, sgcomReq)
	if err != nil {
		logs.Errorf("call dataserver to list sg common failed, clbID: %s, err: %v, rid: %s", clbID, err, kt.Rid)
		return nil, nil, err
	}

	return &clbInfo, sgComList, nil
}

func (g *securityGroup) getSecurityGroupMap(kt *kit.Kit, sgIDs []string) (
	map[string]corecloud.BaseSecurityGroup, error) {

	sgReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("id", sgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, ids: %v, rid: %s",
			err, sgIDs, kt.Rid)
		return nil, err
	}

	sgMap := make(map[string]corecloud.BaseSecurityGroup, len(sgResult.Details))
	for _, sg := range sgResult.Details {
		sgMap[sg.ID] = sg
	}

	return sgMap, nil
}
