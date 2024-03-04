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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
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
	if err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
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

	sgList, clbInfo, err := g.getSecurityGroupAndClb(cts.Kit, req.SecurityGroupIDs, req.LoadBalancerID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, clbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	sgCloudIDs := make([]string, len(sgList))
	for _, sg := range sgList {
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
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

	createReq := &protocloud.SGCommonRelBatchCreateReq{Rels: make([]protocloud.SGCommonRelCreate, 0)}
	for _, sgID := range req.SecurityGroupIDs {
		createReq.Rels = append(createReq.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: sgID,
			ResID:           req.LoadBalancerID,
			ResType:         enumor.ClbCloudResType,
			Priority:        constant.LoadBalancerBindSecurityGroupMaxLimit,
		})
	}

	if err = g.dataCli.Global.SGCommonRel.BatchCreate(cts.Kit, createReq); err != nil {
		logs.Errorf("request dataservice create security group clb rels failed, err: %v, req: %+v, rid: %s",
			err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getSecurityGroupAndClb(kt *kit.Kit, sgIDs []string, clbID string) (
	[]corecloud.BaseSecurityGroup, *coreclb.BaseClb, error) {

	sgReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("id", sgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, ids: %v, rid: %s",
			err, sgIDs, kt.Rid)
		return nil, nil, err
	}

	if len(sgResult.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "security group: %v not found", sgIDs)
	}

	clbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", clbID),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := g.dataCli.Global.LoadBalancer.ListClb(kt, clbReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud clb failed, err: %v, id: %s, rid: %s", err, clbID, kt.Rid)
		return nil, nil, err
	}

	if len(result.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "clb: %s not found", clbID)
	}

	return sgResult.Details, &result.Details[0], nil
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

	_, clbInfo, err := g.getSecurityGroupAndClb(cts.Kit, []string{req.SecurityGroupID}, req.LoadBalancerID)
	if err != nil {
		return nil, err
	}

	sgReq := &protocloud.SGCommonRelWithSecurityGroupListReq{
		ResIDs:  []string{req.LoadBalancerID},
		ResType: enumor.ClbCloudResType,
	}
	sgComList, err := g.dataCli.Global.SGCommonRel.ListWithSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	allSGIDs := make([]string, 0)
	existSG := false
	for _, sg := range sgComList {
		if sg.ID == req.SecurityGroupID {
			existSG = true
			continue
		}
		allSGIDs = append(allSGIDs, sg.CloudID)
	}
	if !existSG {
		return nil, errf.Newf(errf.RecordNotFound, "not found sg id: %s", req.SecurityGroupID)
	}

	client, err := g.ad.TCloud(cts.Kit, clbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &clb.TCloudSetClbSecurityGroupOption{
		Region:         clbInfo.Region,
		LoadBalancerID: clbInfo.CloudID,
		SecurityGroups: allSGIDs,
	}
	if _, err = client.SetClbSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group disAssociate clb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(req.SecurityGroupID, req.LoadBalancerID, enumor.ClbCloudResType)
	if err = g.dataCli.Global.SGCvmRel.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group clb rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
