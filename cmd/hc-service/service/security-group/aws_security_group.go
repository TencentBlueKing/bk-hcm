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

	"hcm/pkg/adaptor/aws"
	typescore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
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

	"github.com/aws/aws-sdk-go/service/ec2"
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

	opt := &securitygroup.AwsCreateOption{
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

	listOpt := &securitygroup.AwsListOption{
		Region:   req.Region,
		CloudIDs: []string{cloudID},
	}
	_, result, err := client.ListSecurityGroup(cts.Kit, listOpt)
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
				// Tags:        core.NewTagMap(req.Tags...),
				MgmtType:    req.MgmtType,
				MgmtBizID:   req.MgmtBizID,
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				UsageBizIds: req.UsageBizIds,
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

// AwsSecurityGroupAssociateCvm ...
func (g *securityGroup) AwsSecurityGroupAssociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.Aws(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AwsAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to aws security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	err = g.createSGCommonRelsForAws(cts.Kit, client, sg.Region, cvm)
	if err != nil {
		logs.Errorf("create security group common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

func (g *securityGroup) createSGCommonRelsForAws(kt *kit.Kit, client *aws.Aws, region string,
	cvm *corecvm.BaseCvm) error {

	awsCvmFromCloud, err := g.listAwsCvmFromCloud(kt, client, region, cvm)
	if err != nil {
		logs.Errorf("request adaptor to list cvm from cloud failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sgCloudIDs := make([]string, 0, len(awsCvmFromCloud[0].SecurityGroups))
	for _, cur := range awsCvmFromCloud[0].SecurityGroups {
		sgCloudIDs = append(sgCloudIDs, converter.PtrToVal(cur.GroupId))
	}

	sgCloudIDToIDMap, err := g.getSecurityGroupMapByCloudIDs(kt, enumor.Aws, region, sgCloudIDs)
	if err != nil {
		logs.Errorf("get security group map by cloud ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sgIDs := make([]string, 0, len(sgCloudIDs))
	for _, cur := range sgCloudIDs {
		sgID, ok := sgCloudIDToIDMap[cur]
		if !ok {
			logs.Errorf("cloud id(%s) not found in security group map, rid: %s", cur, kt.Rid)
			return fmt.Errorf("cloud id(%s) not found in security group map", cur)
		}
		sgIDs = append(sgIDs, sgID)
	}

	err = g.createSGCommonRels(kt, enumor.Aws, enumor.CvmCloudResType, cvm.ID, sgIDs)
	if err != nil {
		logs.Errorf("create security group common rels failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (g *securityGroup) listAwsCvmFromCloud(kt *kit.Kit, client *aws.Aws, region string, cvm *corecvm.BaseCvm) (
	[]typecvm.AwsCvm, error) {
	listOpt := &typecvm.AwsListOption{
		Region:   region,
		CloudIDs: []string{cvm.CloudID},
	}
	awsCvms, _, err := client.ListCvm(kt, listOpt)
	if err != nil {
		logs.Errorf("list aws cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}
	if len(awsCvms) == 0 {
		logs.Errorf("aws cvm(%s) not found, rid: %s", cvm.CloudID, kt.Rid)
		return nil, fmt.Errorf("aws cvm(%s) not found", cvm.CloudID)
	}
	return awsCvms, nil
}

// AwsSecurityGroupDisassociateCvm ...
func (g *securityGroup) AwsSecurityGroupDisassociateCvm(cts *rest.Contexts) (interface{}, error) {
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

	client, err := g.ad.Aws(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AwsAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmID:           cvm.CloudID,
	}
	if err = client.SecurityGroupCvmDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to aws security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(enumor.Aws, req.CvmID, []string{req.SecurityGroupID}, enumor.CvmCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice delete security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, deleteReq, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 同步主机数据

	return nil, nil
}

func (g *securityGroup) getVpcIDByCloudVpcID(kt *kit.Kit, cloudVpcID string) (string, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudVpcID),
		Page:   core.NewDefaultBasePage(),
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

	opt := &securitygroup.AwsDeleteOption{
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

// AwsListSecurityGroupStatistic ...
func (g *securityGroup) AwsListSecurityGroupStatistic(cts *rest.Contexts) (any, error) {
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
			return nil, fmt.Errorf("aws security group: %s not found", sgID)
		}
		cloudIDToSgIDMap[sg.CloudID] = sgID
	}

	resp, err := g.listAwsCvmNetworkInterfaceFromCloud(cts.Kit, req.Region, req.AccountID, cloudIDToSgIDMap)
	if err != nil {
		logs.Errorf("list aws cvm network interface from cloud failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	sgIDToResourceCountMap, err := g.countAwsSecurityGroupStatistic(cts.Kit, resp, cloudIDToSgIDMap)
	if err != nil {
		logs.Errorf("count aws security group statistic failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return resCountMapToSGStatisticResp(sgIDToResourceCountMap), nil
}

func (g *securityGroup) countAwsSecurityGroupStatistic(kt *kit.Kit, details []networkinterface.AwsNetworkInterface,
	cloudIDToSgIDMap map[string]string) (map[string]map[string]int64, error) {

	sgIDToResourceCountMap := make(map[string]map[string]int64)
	for _, sgID := range cloudIDToSgIDMap {
		sgIDToResourceCountMap[sgID] = make(map[string]int64)
	}
	for _, one := range details {
		/**
		InterfaceType 可选值:
		api_gateway_managed | aws_codestar_connections_managed | branch | ec2_instance_connect_endpoint |
		efa | efa-only | efs | gateway_load_balancer | gateway_load_balancer_endpoint | global_accelerator_managed |
		interface | iot_rules_managed | lambda | load_balancer | nat_gateway | network_load_balancer | quicksight |
		transit_gateway | trunk | vpc_endpoint
		*/
		interfaceType := converter.PtrToVal(one.InterfaceType)
		for _, group := range one.Groups {
			sgID, ok := cloudIDToSgIDMap[converter.PtrToVal(group.GroupId)]
			if !ok {
				logs.Warnf("cloud id: %s not found in cloud id to sg id map, rid: %s",
					group.GroupId, kt.Rid)
				continue
			}
			sgIDToResourceCountMap[sgID][interfaceType]++
		}
	}
	return sgIDToResourceCountMap, nil
}

func (g *securityGroup) listAwsCvmNetworkInterfaceFromCloud(kt *kit.Kit, region, accountID string,
	cloudIDToIDMap map[string]string) ([]networkinterface.AwsNetworkInterface, error) {

	cli, err := g.ad.Aws(kt, accountID)
	if err != nil {
		return nil, err
	}

	result := make([]networkinterface.AwsNetworkInterface, 0)
	var nextToken *string
	for {
		opt := &networkinterface.AwsNetworkInterfaceListOption{
			Region: region,
			Page: &typescore.AwsPage{
				NextToken:  nextToken,
				MaxResults: converter.ValToPtr(int64(typescore.AwsQueryLimit)),
			},
			Filters: []*ec2.Filter{
				{
					Name:   converter.ValToPtr("group-id"),
					Values: converter.SliceToPtr(converter.MapKeyToSlice(cloudIDToIDMap)),
				},
			},
		}

		resp, err := cli.DescribeNetworkInterfaces(kt, opt)
		if err != nil {
			logs.Errorf("describe network interfaces failed, err: %v, sgCloudIDs: %v, rid: %s",
				err, converter.MapKeyToSlice(cloudIDToIDMap), kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			result = append(result, detail)
		}
		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return result, nil
}
