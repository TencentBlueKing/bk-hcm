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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
	tcloud "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// AwsSGRuleSync ...
type AwsSGRuleSync struct {
	IsUpdate     bool
	IsRealUpdate bool
	SGRule       *ec2.SecurityGroupRule
}

// AzureSGRuleSync ...
type AzureSGRuleSync struct {
	IsUpdate     bool
	IsRealUpdate bool
	SGRule       *armnetwork.SecurityRule
}

// HuaWeiSGRuleSync ...
type HuaWeiSGRuleSync struct {
	IsUpdate     bool
	IsRealUpdate bool
	SGRule       model.SecurityGroupRule
}

// TCloudSGRuleSync ...
type TCloudSGRuleSync struct {
	Version      string
	IsUpdate     bool
	IsRealUpdate bool
	SGRuleID     string
	Typ          enumor.SecurityGroupRuleType
	SGRule       *vpc.SecurityGroupPolicy
}

// SecurityGroupSyncDS data-service diff for sync
type SecurityGroupSyncDS struct {
	IsUpdated       bool
	HcSecurityGroup cloud.BaseSecurityGroup
}

// SecurityGroupSyncHuaWeiDiff huawei cloud diff for sync
type SecurityGroupSyncHuaWeiDiff struct {
	SecurityGroup model.SecurityGroup
}

// SecurityGroupSyncTCloudDiff tcloud diff for sync
type SecurityGroupSyncTCloudDiff struct {
	SecurityGroup *tcloud.SecurityGroup
}

// SecurityGroupSyncAwsDiff aws diff for sync
type SecurityGroupSyncAwsDiff struct {
	SecurityGroup *ec2.SecurityGroup
}

// SecurityGroupSyncAzureDiff azure diff for sync
type SecurityGroupSyncAzureDiff struct {
	SecurityGroup *armnetwork.SecurityGroup
}

// GetDatasFromDSForSecurityGroupSync get sg datas from hc
func GetDatasFromDSForSecurityGroupSync(kt *kit.Kit, req *proto.SecurityGroupSyncReq,
	dataCli *dataservice.Client) (map[string]*SecurityGroupSyncDS, error) {

	start := 0
	resultsHcm := make([]corecloud.BaseSecurityGroup, 0)
	for {
		dataReq := &dataproto.SecurityGroupListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(),
			dataReq)

		if err != nil {
			logs.Errorf("from data-service list security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(results.Details) == 0 {
			break
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*SecurityGroupSyncDS)
	for _, result := range resultsHcm {
		sg := new(SecurityGroupSyncDS)
		sg.IsUpdated = false
		sg.HcSecurityGroup = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

func diffSecurityGroupSyncDelete(kt *kit.Kit, deleteCloudIDs []string,
	dataCli *dataservice.Client) error {

	batchDeleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if err := dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func getAddCloudIDs[T any](cloudMap map[string]T, dsMap map[string]*SecurityGroupSyncDS) []string {
	addCloudIDs := make([]string, 0)
	for id := range cloudMap {
		if _, ok := dsMap[id]; !ok {
			addCloudIDs = append(addCloudIDs, id)
		} else {
			dsMap[id].IsUpdated = true
		}
	}

	return addCloudIDs
}

func getDeleteAndUpdateCloudIDs(dsMap map[string]*SecurityGroupSyncDS) ([]string, []string) {
	deleteCloudIDs := make([]string, 0)
	updateCloudIDs := make([]string, 0)
	for id, one := range dsMap {
		if !one.IsUpdated {
			deleteCloudIDs = append(deleteCloudIDs, id)
		} else {
			updateCloudIDs = append(updateCloudIDs, id)
		}
	}

	return deleteCloudIDs, updateCloudIDs
}
