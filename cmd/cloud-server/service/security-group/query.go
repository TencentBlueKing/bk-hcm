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

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// GetSecurityGroup get security group.
func (svc *securityGroupSvc) GetSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.getSecurityGroup(cts, handler.ResOperateAuth)
}

// GetBizSecurityGroup get biz security group.
func (svc *securityGroupSvc) GetBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.getSecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) getSecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	cvmCount := uint64(0)

	if baseInfo.Vendor != enumor.Azure {
		cvmCount, err = svc.queryAssociateCvmCount(cts.Kit, id)
		if err != nil {
			return nil, err
		}
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		sg, err := svc.client.DataService().TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}

		return &proto.SecurityGroup[corecloud.TCloudSecurityGroupExtension]{
			BaseSecurityGroup: sg.BaseSecurityGroup,
			CvmCount:          cvmCount,
			Extension:         sg.Extension,
		}, nil

	case enumor.Aws:
		sg, err := svc.client.DataService().Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}

		return &proto.SecurityGroup[corecloud.AwsSecurityGroupExtension]{
			BaseSecurityGroup: sg.BaseSecurityGroup,
			CvmCount:          cvmCount,
			Extension:         sg.Extension,
		}, nil

	case enumor.HuaWei:
		sg, err := svc.client.DataService().HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}

		return &proto.SecurityGroup[corecloud.HuaWeiSecurityGroupExtension]{
			BaseSecurityGroup: sg.BaseSecurityGroup,
			CvmCount:          cvmCount,
			Extension:         sg.Extension,
		}, nil

	case enumor.Azure:
		sg, err := svc.client.DataService().Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}

		subnetCount, niCount, err := svc.queryAssociateSubnetAndNICount(cts.Kit, id)
		if err != nil {
			return nil, err
		}

		return &proto.SecurityGroup[corecloud.AzureSecurityGroupExtension]{
			BaseSecurityGroup:     sg.BaseSecurityGroup,
			NetworkInterfaceCount: niCount,
			SubnetCount:           subnetCount,
			Extension:             sg.Extension,
		}, nil

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
	}
}

func (svc *securityGroupSvc) queryAssociateCvmCount(kt *kit.Kit, id string) (uint64, error) {
	cvmRelOpt := &core.ListReq{
		Filter: tools.EqualExpression("security_group_id", id),
		Page:   core.NewCountPage(),
	}
	cvmRelResult, err := svc.client.DataService().Global.SGCvmRel.ListSgCvmRels(kt.Ctx, kt.Header(), cvmRelOpt)
	if err != nil {
		return 0, err
	}

	return cvmRelResult.Count, nil
}

func (svc *securityGroupSvc) queryAssociateSubnetAndNICount(kt *kit.Kit, id string) (uint64, uint64, error) {
	listOpt := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "extension.security_group_id",
					Op:    filter.JSONEqual.Factory(),
					Value: id,
				},
			},
		},
		Page: core.NewCountPage(),
	}
	subnetResult, err := svc.client.DataService().Global.Subnet.List(kt.Ctx, kt.Header(), listOpt)
	if err != nil {
		return 0, 0, err
	}

	niResult, err := svc.client.DataService().Global.NetworkInterface.List(kt, listOpt)
	if err != nil {
		return 0, 0, err
	}

	return subnetResult.Count, niResult.Count, nil
}

// ListSecurityGroup list security group.
func (svc *securityGroupSvc) ListSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listSecurityGroup(cts, handler.ListResourceAuthRes)
}

// ListBizSecurityGroup list biz security group.
func (svc *securityGroupSvc) ListBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listSecurityGroup(cts, handler.ListBizAuthRes)
}

func (svc *securityGroupSvc) listSecurityGroup(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(proto.SecurityGroupListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SecurityGroup, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	dataReq := &dataproto.SecurityGroupListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		dataReq)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.SecurityGroupListResult{
		Count:   result.Count,
		Details: result.Details,
	}, nil
}

// ListSecurityGroupsByCvmID list security groups by cvm_id.
func (svc *securityGroupSvc) ListSecurityGroupsByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGByCvmID(cts, handler.ResOperateAuth)
}

// ListBizSecurityGroupsByCvmID list biz security groups by cvm_id.
func (svc *securityGroupSvc) ListBizSecurityGroupsByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGByCvmID(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGByCvmID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	cvmID := cts.PathParameter("cvm_id").String()
	if len(cvmID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cvm_id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CvmCloudResType, cvmID)
	if err != nil {
		logs.Errorf("get resource vendor failed, err: %s, cvmID: %s, rid: %s", err, cvmID, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	// azure的安全组绑定在网络接口和子网上面，查询规则与其他云不同
	if baseInfo.Vendor == enumor.Azure {
		return svc.listSGByCvmIDForAzure(cts.Kit, cvmID)
	}

	if baseInfo.Vendor == enumor.Gcp {
		return nil, errors.New("gcp not support security group")
	}

	listReq := &dataproto.SGCvmRelWithSecurityGroupListReq{
		CvmIDs: []string{cvmID},
	}
	result, err := svc.client.DataService().Global.SGCvmRel.ListWithSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list security group by cvm_id failed, err: %v, req: %v, rid: %s", err, cts.Kit.Rid, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *securityGroupSvc) listSGByCvmIDForAzure(kt *kit.Kit, cvmID string) (interface{}, error) {

	cvm, err := svc.client.DataService().Azure.Cvm.GetCvm(kt.Ctx, kt.Header(), cvmID)
	if err != nil {
		logs.Errorf("get cvm failed, err: %v, cvmID: %s, rid: %s", err, cvmID, kt.Rid)
		return nil, err
	}

	listSubnetReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", cvm.SubnetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	subnetResult, err := svc.client.DataService().Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), listSubnetReq)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, subnetIDs: %v, rid: %s", err, cvm.SubnetIDs, kt.Rid)
		return nil, err
	}

	listNIReq := &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", cvm.Extension.CloudNetworkInterfaceIDs),
		Page:   core.NewDefaultBasePage(),
	}
	niResult, err := svc.client.DataService().Azure.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(),
		listNIReq)
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, niIDs: %v, rid: %s", err,
			cvm.Extension.CloudNetworkInterfaceIDs, kt.Rid)
		return nil, err
	}

	sgIDs := make([]string, 0)
	for _, one := range subnetResult.Details {
		if len(one.Extension.SecurityGroupID) != 0 {
			sgIDs = append(sgIDs, one.Extension.SecurityGroupID)
		}
	}

	for _, one := range niResult.Details {
		if len(converter.PtrToVal(one.Extension.SecurityGroupID)) != 0 {
			sgIDs = append(sgIDs, *one.Extension.SecurityGroupID)
		}
	}

	if len(sgIDs) == 0 {
		return make([]corecloud.SGCvmRelWithBaseSecurityGroup, 0), nil
	}

	listSGReq := &dataproto.SecurityGroupListReq{
		Filter: tools.ContainersExpression("id", sgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listSGReq)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, sgIDs: %v, rid: %s", err, sgIDs, kt.Rid)
		return nil, err
	}

	sgs := make([]corecloud.SGCvmRelWithBaseSecurityGroup, 0, len(sgResult.Details))
	for _, one := range sgResult.Details {
		sgs = append(sgs, corecloud.SGCvmRelWithBaseSecurityGroup{
			BaseSecurityGroup: one,
			CvmID:             cvmID,
		})
	}

	return sgs, nil
}

// ListSecurityGroupsByResID list security groups by res_id.
func (svc *securityGroupSvc) ListSecurityGroupsByResID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGByResID(cts, handler.ResOperateAuth)
}

// ListBizSecurityGroupsByResID list biz security groups by res_id.
func (svc *securityGroupSvc) ListBizSecurityGroupsByResID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGByResID(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGByResID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	resID := cts.PathParameter("res_id").String()
	if len(resID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "res_id is required")
	}

	resType := enumor.CloudResourceType(cts.PathParameter("res_type").String())
	if len(resType) == 0 {
		return nil, errf.New(errf.InvalidParameter, "res_type is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, resType, resID)
	if err != nil {
		logs.Errorf("get resource vendor failed, err: %s, resID: %s, resType: %s, rid: %s",
			err, resID, resType, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		logs.Errorf("list security group by resID failed, id: %s, err: %v, rid: %s", resID, err, cts.Kit.Rid)
		return nil, err
	}

	listReq := &dataproto.SGCommonRelWithSecurityGroupListReq{
		ResIDs:  []string{resID},
		ResType: resType,
	}
	result, err := svc.client.DataService().Global.SGCommonRel.ListWithSecurityGroup(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list security group by res_id failed, resID: %s, err: %v, req: %v, rid: %s",
			resID, err, cts.Kit.Rid, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
