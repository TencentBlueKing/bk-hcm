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

	typecvm "hcm/pkg/adaptor/types/cvm"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// TCloudSGAssociateLoadBalancer ...
func (g *securityGroup) TCloudSGAssociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
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

	sgCloudIDs, sgComReq, err := g.getUpsertSGIDsParams(cts.Kit, req, sgComList)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, lbInfo.AccountID)
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

// TCloudSGDisassociateLoadBalancer ...
func (g *securityGroup) TCloudSGDisassociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.TCloud(cts.Kit, lbInfo.AccountID)
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
		enumor.TCloud, req.LbID, []string{req.SecurityGroupID}, enumor.LoadBalancerCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice tcloud delete security group lb rels failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getUpsertSGIDsParams get update or insert security group IDs and parameters for upsert operation.
func (g *securityGroup) getUpsertSGIDsParams(kt *kit.Kit, req *hclb.TCloudSetLbSecurityGroupReq,
	sgComList *protocloud.SGCommonRelListResult) ([]string, *protocloud.SGCommonRelBatchUpsertReq, error) {

	delSGIDs := make([]string, 0)
	for _, sg := range sgComList.Details {
		delSGIDs = append(delSGIDs, sg.SecurityGroupID)
	}

	sgComReq := &protocloud.SGCommonRelBatchUpsertReq{
		Rels: make([]protocloud.SGCommonRelCreate, 0, len(req.SecurityGroupIDs)),
	}
	if len(delSGIDs) > 0 {
		sgComReq.DeleteReq = buildSGCommonRelDeleteReq(
			enumor.TCloud, req.LbID, delSGIDs, enumor.LoadBalancerCloudResType)
	}

	tmpPriority := int64(0)
	for _, newSGID := range req.SecurityGroupIDs {
		tmpPriority++
		sgComReq.Rels = append(sgComReq.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: newSGID,
			ResVendor:       enumor.TCloud,
			ResID:           req.LbID,
			ResType:         enumor.LoadBalancerCloudResType,
			Priority:        tmpPriority,
		})
	}

	sgMap, err := g.getSecurityGroupMap(kt, req.SecurityGroupIDs)
	if err != nil {
		return nil, nil, err
	}

	// 安全组的云端ID数组
	sgCloudIDs := make([]string, 0)
	for _, sgID := range req.SecurityGroupIDs {
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

// TCloudSGBatchAssociateCvm 批量绑定安全组
func (g *securityGroup) TCloudSGBatchAssociateCvm(cts *rest.Contexts) (any, error) {

	req := new(proto.SecurityGroupBatchAssociateCvmReq)
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

	cvmList, err := g.getCvms(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}
	cvmCloudIDToIDMap := make(map[string]string, len(req.CvmIDs))
	for _, baseCvm := range cvmList {
		cvmCloudIDToIDMap[baseCvm.CloudID] = baseCvm.ID
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          converter.MapKeyToSlice(cvmCloudIDToIDMap),
	}
	if err = client.SecurityGroupCvmBatchAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	err = g.createSGCommonRelsForTCloud(cts.Kit, client, sg.Region, cvmCloudIDToIDMap)
	if err != nil {
		logs.Errorf("create sg common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
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

	err = g.createSGCommonRelsForTCloud(cts.Kit, client, sg.Region, map[string]string{cvm.CloudID: cvm.ID})
	if err != nil {
		logs.Errorf("create security group cvm rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

	deleteReq := buildSGCommonRelDeleteReq(
		enumor.TCloud, req.CvmID, []string{req.SecurityGroupID}, enumor.CvmCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

// TCloudSGBatchDisassociateCvm  批量解绑安全组
func (g *securityGroup) TCloudSGBatchDisassociateCvm(cts *rest.Contexts) (any, error) {
	req := new(proto.SecurityGroupBatchAssociateCvmReq)
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

	cvmList, err := g.getCvms(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}
	cloudCvmIDs := make([]string, 0, len(req.CvmIDs))
	for _, baseCvm := range cvmList {
		cloudCvmIDs = append(cloudCvmIDs, baseCvm.CloudID)
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          cloudCvmIDs,
	}
	if err = client.SecurityGroupCvmBatchDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq, err := buildSGCommonRelDeleteReqForMultiResource(enumor.CvmCloudResType, req.SecurityGroupID, req.CvmIDs...)
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
