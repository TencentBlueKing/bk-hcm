/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package azure

import (
	"fmt"
	gosync "sync"

	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// CondSyncParams 条件同步选项
type CondSyncParams struct {
	AccountID          string   `json:"account_id" validate:"required"`
	CloudIDs           []string `json:"cloud_ids,omitempty" validate:"omitempty,max=20"`
	ResourceGroupNames []string `json:"resource_group_names,omitempty" validate:"max=20"`
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

var condSyncFuncMap = map[enumor.CloudResourceType]CondSyncFunc{
	enumor.SecurityGroupCloudResType: CondSyncSecurityGroup,
	enumor.SubnetCloudResType:        CondSyncSubnet,
}

// GetCondSyncFunc ...
func GetCondSyncFunc(res enumor.CloudResourceType) (syncFunc CondSyncFunc, ok bool) {
	syncFunc, ok = condSyncFuncMap[res]
	return syncFunc, ok
}

// CondSyncSecurityGroup ...
func CondSyncSecurityGroup(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.AzureSyncReq{
		AccountID: params.AccountID,
		CloudIDs:  params.CloudIDs,
	}
	for i := range params.ResourceGroupNames {
		syncReq.ResourceGroupName = params.ResourceGroupNames[i]
		err := cliSet.HCService().Azure.SecurityGroup.SyncSecurityGroup(kt.Ctx, kt.Header(), &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync security group failed, err: %v, req: %+v, rid: %s",
				enumor.Azure, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync security group end, req: %+v, rid: %s", enumor.Azure, syncReq, kt.Rid)
	}
	return nil
}

type azureSubnetVpcKey struct {
	resourceGroupName string
	cloudVpcID        string
}

// CondSyncSubnet sync azure subnet by condition.
func CondSyncSubnet(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	vpcKeys, err := listAzureCondSyncSubnetVpcKeys(kt, cliSet, params)
	if err != nil {
		return err
	}

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for key := range vpcKeys {
		pipeline <- true
		wg.Add(1)
		go func(resourceGroupName, cloudVpcID string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			req := &sync.AzureSubnetSyncReq{
				AccountID:         params.AccountID,
				ResourceGroupName: resourceGroupName,
				CloudVpcID:        cloudVpcID,
			}
			err := cliSet.HCService().Azure.Subnet.SyncSubnet(kt.Ctx, kt.Header(), req)
			if firstErr == nil && err != nil {
				logs.Errorf("[%s] conditional sync subnet failed, err: %v, req: %+v, rid: %s",
					enumor.Azure, err, req, kt.Rid)
				firstErr = err
			}
		}(key.resourceGroupName, key.cloudVpcID)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	logs.Infof("[%s] conditional sync subnet end, account: %s, resource_group_names: %v, cloud_ids: %v, rid: %s",
		enumor.Azure, params.AccountID, params.ResourceGroupNames, params.CloudIDs, kt.Rid)
	return nil
}

func listAzureCondSyncSubnetVpcKeys(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) (
	map[azureSubnetVpcKey]struct{}, error) {

	if len(params.CloudIDs) > 0 {
		return listAzureSubnetVpcKeysByCloudIDs(kt, cliSet, params.AccountID, params.CloudIDs)
	}

	vpcKeys := make(map[azureSubnetVpcKey]struct{})
	for _, name := range params.ResourceGroupNames {
		keys, err := listAzureSubnetVpcKeysInResourceGroup(kt, cliSet, params.AccountID, name)
		if err != nil {
			return nil, err
		}
		for key := range keys {
			vpcKeys[key] = struct{}{}
		}
	}
	return vpcKeys, nil
}

func listAzureSubnetVpcKeysByCloudIDs(kt *kit.Kit, cliSet *client.ClientSet, accountID string,
	cloudIDs []string) (map[azureSubnetVpcKey]struct{}, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleIn("cloud_id", cloudIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := cliSet.DataService().Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("[%s] list subnet for cond sync failed, err: %v, req: %+v, rid: %s",
			enumor.Azure, err, listReq, kt.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		return nil, fmt.Errorf("subnet not found by cloud_ids: %v", cloudIDs)
	}

	vpcKeys := make(map[azureSubnetVpcKey]struct{}, len(resp.Details))
	for _, one := range resp.Details {
		vpcKeys[azureSubnetVpcKey{
			resourceGroupName: one.Extension.ResourceGroupName,
			cloudVpcID:        one.CloudVpcID,
		}] = struct{}{}
	}
	return vpcKeys, nil
}

func listAzureSubnetVpcKeysInResourceGroup(kt *kit.Kit, cliSet *client.ClientSet, accountID,
	resourceGroupName string) (map[azureSubnetVpcKey]struct{}, error) {

	vpcKeys := make(map[azureSubnetVpcKey]struct{})
	filterRules := []filter.RuleFactory{
		&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
		&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(), Value: resourceGroupName},
	}
	listReq := &core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: filterRules},
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		Fields: []string{"cloud_id"},
	}
	for start := uint32(0); ; start += uint32(core.DefaultMaxPageLimit) {
		listReq.Page.Start = start
		vpcResult, err := cliSet.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("[%s] list vpc for cond sync subnet failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
			return nil, err
		}
		for _, vpc := range vpcResult.Details {
			vpcKeys[azureSubnetVpcKey{resourceGroupName: resourceGroupName, cloudVpcID: vpc.CloudID}] = struct{}{}
		}
		if len(vpcResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
	}
	return vpcKeys, nil
}
