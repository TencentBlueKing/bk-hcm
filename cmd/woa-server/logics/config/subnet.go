/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// SubnetIf provides management interface for operations of subnet config
type SubnetIf interface {
	// GetAllSubnet get all subnet
	GetAllSubnet(kt *kit.Kit, req *types.GetAllSubnetReq) (*types.GetSubnetResult, error)
	// GetSubnetList get subnet detail config list
	GetSubnetList(kt *kit.Kit, input *types.GetSubnetListParam) (*types.GetSubnetResult, error)
	// UpdateSubnetBatch update subnet batch
	UpdateSubnetBatch(kt *kit.Kit, ids []string, update map[string]interface{}) error
}

// NewSubnetOp creates a subnet interface
func NewSubnetOp(client *client.ClientSet, thirdCli *thirdparty.Client) SubnetIf {
	return &subnet{
		cvm:    thirdCli.OldCVM,
		client: client,
	}
}

type subnet struct {
	cvm    cvmapi.CVMClientInterface
	client *client.ClientSet
}

// GetAllSubnet get all subnet
func (s *subnet) GetAllSubnet(kt *kit.Kit, req *types.GetAllSubnetReq) (*types.GetSubnetResult, error) {
	filterRules := []filter.RuleFactory{
		tools.RuleEqual("region", req.Region),
		tools.RuleJSONEqual("extension.enable_cvm", "true"),
	}
	if len(req.Zones) > 0 {
		filterRules = append(filterRules, tools.RuleIn("zone", req.Zones))
	}
	if len(req.CloudVpcID) > 0 {
		filterRules = append(filterRules, tools.RuleEqual("cloud_vpc_id", req.CloudVpcID))
	}
	if len(req.CloudID) > 0 {
		filterRules = append(filterRules, tools.RuleEqual("cloud_id", req.CloudID))
	}
	if len(req.Name) > 0 {
		filterRules = append(filterRules, tools.RuleCis("name", req.Name))
	}

	subnetReq := &types.GetSubnetListParam{
		Filter: &filter.Expression{Op: filter.And, Rules: filterRules},
		Page:   core.NewDefaultBasePage(),
	}
	subnetList := make([]*types.Subnet, 0)
	for {
		tmpSubnets, err := s.GetSubnetList(kt, subnetReq)
		if err != nil {
			return nil, err
		}
		subnetList = append(subnetList, tmpSubnets.Info...)
		if len(tmpSubnets.Info) < int(subnetReq.Page.Limit) {
			break
		}
		subnetReq.Page.Start += uint32(subnetReq.Page.Limit)
	}

	rst := &types.GetSubnetResult{
		Count: int64(len(subnetList)),
		Info:  subnetList,
	}

	return rst, nil
}

// GetSubnetList get subnet detail config list
func (s *subnet) GetSubnetList(kt *kit.Kit, input *types.GetSubnetListParam) (*types.GetSubnetResult, error) {
	// 查询账号信息
	accountID, err := getTCloudZiyanAccount(kt, s.client)
	if err != nil {
		return nil, err
	}

	filterRules := []filter.RuleFactory{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
		tools.RuleEqual("account_id", accountID),
	}
	if input.Filter != nil {
		filterRules = append(filterRules, input.Filter)
	}

	// 从MySQL查询子网列表
	listReq := &core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: filterRules},
		Page:   input.Page,
	}
	subnetList, err := s.client.DataService().TCloudZiyan.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("failed to list subnet by page, err: %v, vendor: %s, accountID: %s, rid: %s",
			err, enumor.TCloudZiyan, accountID, kt.Rid)
		return nil, err
	}

	if input.Page.Count {
		return &types.GetSubnetResult{Count: int64(subnetList.Count)}, nil
	}

	// 查询VPC列表
	vpcIDNameMap, err := s.getVpcNameMap(kt, accountID, subnetList)
	if err != nil {
		return nil, err
	}

	subnetResult := make([]*types.Subnet, 0, len(subnetList.Details))
	for _, subnetDetail := range subnetList.Details {
		var enableCvm bool
		if subnetDetail.Extension != nil {
			enableCvm = subnetDetail.Extension.EnableCvm
		}
		var vpcName string
		if _, ok := vpcIDNameMap[subnetDetail.CloudVpcID]; ok {
			vpcName = vpcIDNameMap[subnetDetail.CloudVpcID]
		}
		subnetResult = append(subnetResult, &types.Subnet{
			BkInstId:   subnetDetail.ID,
			Region:     subnetDetail.Region,
			Zone:       subnetDetail.Zone,
			VpcId:      subnetDetail.CloudVpcID,
			VpcName:    vpcName,
			SubnetId:   subnetDetail.CloudID,
			SubnetName: subnetDetail.Name,
			Enable:     enableCvm,
			Comment:    cvt.PtrToVal(subnetDetail.Memo),
		})
	}

	rst := &types.GetSubnetResult{Count: int64(subnetList.Count), Info: subnetResult}
	return rst, nil
}

func (s *subnet) getVpcNameMap(kt *kit.Kit, accountID string,
	subnetList *cloud.SubnetExtListResult[corecloud.TCloudSubnetExtension]) (map[string]string, error) {

	cloudVpcIDs := slice.Map(subnetList.Details, func(cls corecloud.Subnet[corecloud.TCloudSubnetExtension]) string {
		return cls.CloudVpcID
	})
	var vpcIDNameMap map[string]string
	if len(cloudVpcIDs) > 0 {
		vpcListReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", slice.Unique(cloudVpcIDs)),
				tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleEqual("account_id", accountID),
			),
			Page: core.NewDefaultBasePage(),
		}
		vpcList, err := s.client.DataService().TCloudZiyan.Vpc.ListVpcExt(kt.Ctx, kt.Header(), vpcListReq)
		if err != nil {
			logs.Errorf("failed to list vpc, err: %+v, vendor: %s, cloudVpcIDs: %v, rid: %s",
				err, enumor.TCloudZiyan, cloudVpcIDs, kt.Rid)
			return nil, err
		}
		vpcIDNameMap = slice.FuncToMap(vpcList.Details, func(vpcItem corecloud.Vpc[corecloud.TCloudVpcExtension]) (
			string, string) {
			return vpcItem.CloudID, vpcItem.Name
		})
	}
	return vpcIDNameMap, nil
}

// UpdateSubnetBatch updates subnet batch
func (s *subnet) UpdateSubnetBatch(kt *kit.Kit, ids []string, update map[string]interface{}) error {
	if len(ids) == 0 {
		logs.Errorf("failed to batch update subnet, ids is empty, rid: %s", kt.Rid)
		return errf.Newf(errf.InvalidParameter, "ids is empty")
	}

	var enableCvm *bool
	if update["enable"] != nil {
		enableCvm = cvt.ValToPtr(metadata.GetBool(update["enable"]))
	}

	var memo string
	if update["comment"] != nil {
		memo = metadata.GetString(update["comment"])
	}

	subnets := make([]cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], 0)
	for _, id := range ids {
		tmpRes := cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{ID: id}
		if len(memo) > 0 {
			tmpRes.SubnetUpdateBaseInfo.Memo = cvt.ValToPtr(memo)
		}
		if enableCvm != nil {
			tmpRes.Extension = &cloud.TCloudSubnetUpdateExt{EnableCvm: enableCvm}
		}
		subnets = append(subnets, tmpRes)
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
		Subnets: subnets,
	}
	if err := s.client.DataService().TCloudZiyan.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("failed to update subnet, err: %v, ids: %v, enableCvm: %v, memo: %s, rid: %s",
			err, ids, enableCvm, memo, kt.Rid)
		return err
	}

	return nil
}
