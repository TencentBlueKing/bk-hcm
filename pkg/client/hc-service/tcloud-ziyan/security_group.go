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

package hcziyancli

import (
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/hc-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudSecurityGroupClient create a new security group api client.
func NewCloudSecurityGroupClient(client rest.ClientInterface) *SecurityGroupClient {
	return &SecurityGroupClient{
		client: client,
	}
}

// SecurityGroupClient is data service security group api client.
type SecurityGroupClient struct {
	client rest.ClientInterface
}

// SyncSecurityGroup security group.
func (cli *SecurityGroupClient) SyncSecurityGroup(kt *kit.Kit, request *sync.TCloudSyncReq) error {

	return common.RequestNoResp[sync.TCloudSyncReq](cli.client, rest.POST, kt, request, "/security_groups/sync")
}

// CreateSecurityGroup create security group.
func (cli *SecurityGroupClient) CreateSecurityGroup(kt *kit.Kit,
	request *proto.TCloudSecurityGroupCreateReq) (*core.CreateResult, error) {

	return common.Request[proto.TCloudSecurityGroupCreateReq, core.CreateResult](cli.client, rest.POST, kt, request,
		"/security_groups/create")
}

// UpdateSecurityGroup update security group rule.
func (cli *SecurityGroupClient) UpdateSecurityGroup(kt *kit.Kit, id string,
	request *proto.SecurityGroupUpdateReq) error {

	return common.RequestNoResp[proto.SecurityGroupUpdateReq](cli.client, rest.PATCH, kt, request,
		"/security_groups/%s", id)
}

// DeleteSecurityGroup delete security group.
func (cli *SecurityGroupClient) DeleteSecurityGroup(kt *kit.Kit, id string) error {

	return common.RequestNoResp[common.Empty](cli.client, rest.DELETE, kt, nil,
		"/security_groups/%s", id)
}

// BatchCreateSecurityGroupRule batch create security group rule.
func (cli *SecurityGroupClient) BatchCreateSecurityGroupRule(kt *kit.Kit, sgID string,
	request *proto.TCloudSGRuleCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[proto.TCloudSGRuleCreateReq, core.BatchCreateResult](cli.client, rest.POST, kt, request,
		"/security_groups/%s/rules/batch/create", sgID)
}

// UpdateSecurityGroupRule update security group rule.
func (cli *SecurityGroupClient) UpdateSecurityGroupRule(kt *kit.Kit, sgID, id string,
	request *proto.TCloudSGRuleUpdateReq) error {

	return common.RequestNoResp[proto.TCloudSGRuleUpdateReq](cli.client, rest.PUT, kt, request,
		"/security_groups/%s/rules/%s", sgID, id)
}

// DeleteSecurityGroupRule delete security group rule.
func (cli *SecurityGroupClient) DeleteSecurityGroupRule(kt *kit.Kit, sgID, id string) error {

	return common.RequestNoResp[common.Empty](cli.client, rest.DELETE, kt, nil,
		"/security_groups/%s/rules/%s", sgID, id)
}

// BatchAssociateCvm 根据cvm云id绑定安全组
func (cli *SecurityGroupClient) BatchAssociateCvm(kt *kit.Kit, sgID string, cvmIDs []string) error {

	req := &proto.SecurityGroupBatchAssociateCvmReq{
		SecurityGroupID: sgID,
		CvmIDs:          cvmIDs,
	}
	return common.RequestNoResp[proto.SecurityGroupBatchAssociateCvmReq](cli.client, rest.POST, kt, req,
		"/security_groups/associate/cvms/batch")
}

// BatchDisassociateCvm 根据cvm云id解绑安全组
func (cli *SecurityGroupClient) BatchDisassociateCvm(kt *kit.Kit, sgID string, cvmIDs []string) error {

	req := &proto.SecurityGroupBatchAssociateCvmReq{
		SecurityGroupID: sgID,
		CvmIDs:          cvmIDs,
	}
	return common.RequestNoResp[proto.SecurityGroupBatchAssociateCvmReq](cli.client, rest.POST, kt, req,
		"/security_groups/disassociate/cvms/batch")
}

// AssociateLb ...
func (cli *SecurityGroupClient) AssociateLb(kt *kit.Kit, req *hclb.TCloudSetLbSecurityGroupReq) error {

	return common.RequestNoResp[hclb.TCloudSetLbSecurityGroupReq](cli.client, rest.POST, kt, req,
		"/security_groups/associate/load_balancers")
}

// DisassociateLb ...
func (cli *SecurityGroupClient) DisassociateLb(kt *kit.Kit, req *hclb.TCloudDisAssociateLbSecurityGroupReq) error {

	return common.RequestNoResp[hclb.TCloudDisAssociateLbSecurityGroupReq](cli.client, rest.POST, kt, req,
		"/security_groups/disassociate/load_balancers")
}

// BatchUpdateSecurityGroupRule batch update security group rule.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroupRule(kt *kit.Kit, sgID string,
	request *proto.TCloudSGRuleBatchUpdateReq) error {

	return common.RequestNoResp[proto.TCloudSGRuleBatchUpdateReq](cli.client, rest.PUT, kt, request,
		"/security_groups/%s/rules/batch/update", sgID)
}

// CloneSecurityGroup 克隆安全组
func (cli *SecurityGroupClient) CloneSecurityGroup(kt *kit.Kit, req *proto.TCloudSecurityGroupCloneReq) (
	*core.CreateResult, error) {

	return common.Request[proto.TCloudSecurityGroupCloneReq, core.CreateResult](cli.client, rest.POST, kt, req,
		"/security_groups/clone")
}

// ListSecurityGroupStatistic 查询安全组关联的云上资源数量
func (cli *SecurityGroupClient) ListSecurityGroupStatistic(kt *kit.Kit, req *proto.ListSecurityGroupStatisticReq) (
	*proto.ListSecurityGroupStatisticResp, error) {

	return common.Request[proto.ListSecurityGroupStatisticReq, proto.ListSecurityGroupStatisticResp](
		cli.client, rest.POST, kt, req, "/security_groups/statistic")
}

// SyncSecurityGroupUsageBizRel ...
func (cli *SecurityGroupClient) SyncSecurityGroupUsageBizRel(kt *kit.Kit, req *sync.TCloudSyncReq) error {
	return common.RequestNoResp[sync.TCloudSyncReq](cli.client, rest.POST, kt, req,
		"/security_groups/usage_biz_rels/sync")
}

// DisassociateCvm ...
func (cli *SecurityGroupClient) DisassociateCvm(kt *kit.Kit, req *proto.SecurityGroupAssociateCvmReq) error {

	return common.RequestNoResp[proto.SecurityGroupAssociateCvmReq](cli.client, rest.POST, kt, req,
		"/security_groups/disassociate/cvms")

}

// BatchAssociateLoadBalancers 批量绑定负载均衡
func (cli *SecurityGroupClient) BatchAssociateLoadBalancers(kt *kit.Kit,
	req *proto.SecurityGroupBatchOperateLoadBalancerReq) error {

	return common.RequestNoResp[proto.SecurityGroupBatchOperateLoadBalancerReq](cli.client, rest.POST, kt, req,
		"/security_groups/associate/load_balancers/batch")
}

// BatchDisassociateLoadBalancers 批量解绑负载均衡
func (cli *SecurityGroupClient) BatchDisassociateLoadBalancers(kt *kit.Kit,
	req *proto.SecurityGroupBatchOperateLoadBalancerReq) error {

	return common.RequestNoResp[proto.SecurityGroupBatchOperateLoadBalancerReq](cli.client, rest.POST, kt, req,
		"/security_groups/disassociate/load_balancers/batch")
}
