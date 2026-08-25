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

package huawei

import (
	"fmt"
	gosync "sync"

	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/region"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/api/hc-service/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// CondSyncParams 条件同步选项
type CondSyncParams struct {
	AccountID string   `json:"account_id" validate:"required"`
	Regions   []string `json:"regions,omitempty" validate:"max=20"`
	CloudIDs  []string `json:"cloud_ids,omitempty" validate:"omitempty,max=20"`
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

var condSyncFuncMap = map[enumor.CloudResourceType]CondSyncFunc{
	enumor.RegionCloudResType:        CondSyncRegion,
	enumor.ZoneCloudResType:          CondSyncZone,
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
	syncReq := sync.HuaWeiSyncReq{
		AccountID: params.AccountID,
		CloudIDs:  params.CloudIDs,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().HuaWei.SecurityGroup.SyncSecurityGroup(kt.Ctx, kt.Header(), &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync security group failed, err: %v, req: %+v, rid: %s",
				enumor.HuaWei, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync security group end, req: %+v, rid: %s", enumor.HuaWei, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncRegion sync region
func CondSyncRegion(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := &region.HuaWeiRegionSyncReq{
		AccountID: params.AccountID,
	}
	err := cliSet.HCService().HuaWei.Region.SyncRegion(kt.Ctx, kt.Header(), syncReq)
	if err != nil {
		logs.Errorf("[%s] conditional sync region failed, err: %v, req: %+v, rid: %s",
			enumor.HuaWei, err, syncReq, kt.Rid)
		return err
	}
	logs.Infof("[%s] conditional sync region end, req: %+v, rid: %s", enumor.HuaWei, syncReq, kt.Rid)
	return nil
}

// CondSyncZone sync zone
func CondSyncZone(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	// zone不能根据cloudID进行部分同步
	syncReq := &zone.HuaWeiZoneSyncReq{
		AccountID: params.AccountID,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().HuaWei.Zone.SyncZone(kt.Ctx, kt.Header(), syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync zone failed, err: %v, req: %+v, rid: %s",
				enumor.HuaWei, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync zone end, req: %+v, rid: %s", enumor.HuaWei, syncReq, kt.Rid)
	}
	return nil
}

type huaweiSubnetVpcKey struct {
	region     string
	cloudVpcID string
}

// CondSyncSubnet sync huawei subnet by condition.
func CondSyncSubnet(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	vpcKeys, err := listHuaweiCondSyncSubnetVpcKeys(kt, cliSet, params)
	if err != nil {
		return err
	}

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for key := range vpcKeys {
		pipeline <- true
		wg.Add(1)
		go func(region, cloudVpcID string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			req := &sync.HuaWeiSubnetSyncReq{
				AccountID:  params.AccountID,
				Region:     region,
				CloudVpcID: cloudVpcID,
			}
			err := cliSet.HCService().HuaWei.Subnet.SyncSubnet(kt.Ctx, kt.Header(), req)
			if firstErr == nil && Error(err) != nil {
				logs.Errorf("[%s] conditional sync subnet failed, err: %v, req: %+v, rid: %s",
					enumor.HuaWei, err, req, kt.Rid)
				firstErr = err
			}
		}(key.region, key.cloudVpcID)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	logs.Infof("[%s] conditional sync subnet end, account: %s, regions: %v, cloud_ids: %v, rid: %s",
		enumor.HuaWei, params.AccountID, params.Regions, params.CloudIDs, kt.Rid)
	return nil
}

func listHuaweiCondSyncSubnetVpcKeys(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) (
	map[huaweiSubnetVpcKey]struct{}, error) {

	if len(params.CloudIDs) > 0 {
		return listHuaweiSubnetVpcKeysByCloudIDs(kt, cliSet, params.AccountID, params.CloudIDs)
	}

	vpcKeys := make(map[huaweiSubnetVpcKey]struct{})
	for _, region := range params.Regions {
		keys, err := listHuaweiSubnetVpcKeysInRegion(kt, cliSet, params.AccountID, region)
		if err != nil {
			return nil, err
		}
		for key := range keys {
			vpcKeys[key] = struct{}{}
		}
	}
	return vpcKeys, nil
}

func listHuaweiSubnetVpcKeysByCloudIDs(kt *kit.Kit, cliSet *client.ClientSet, accountID string,
	cloudIDs []string) (map[huaweiSubnetVpcKey]struct{}, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleIn("cloud_id", cloudIDs),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"region", "cloud_vpc_id"},
	}
	resp, err := cliSet.DataService().Global.Subnet.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("[%s] list subnet for cond sync failed, err: %v, req: %+v, rid: %s",
			enumor.HuaWei, err, listReq, kt.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		return nil, fmt.Errorf("subnet not found by cloud_ids: %v", cloudIDs)
	}

	vpcKeys := make(map[huaweiSubnetVpcKey]struct{}, len(resp.Details))
	for _, one := range resp.Details {
		if one.Vendor != enumor.HuaWei {
			continue
		}
		vpcKeys[huaweiSubnetVpcKey{region: one.Region, cloudVpcID: one.CloudVpcID}] = struct{}{}
	}
	if len(vpcKeys) == 0 {
		return nil, fmt.Errorf("subnet not found by cloud_ids: %v", cloudIDs)
	}
	return vpcKeys, nil
}

func listHuaweiSubnetVpcKeysInRegion(kt *kit.Kit, cliSet *client.ClientSet, accountID, region string) (
	map[huaweiSubnetVpcKey]struct{}, error) {

	vpcKeys := make(map[huaweiSubnetVpcKey]struct{})
	accountRegionRules := []filter.RuleFactory{
		&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
		&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
	}
	listReq := &core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: accountRegionRules},
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		Fields: []string{"cloud_id"},
	}
	for start := uint32(0); ; start += uint32(core.DefaultMaxPageLimit) {
		listReq.Page.Start = start
		vpcResult, err := cliSet.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("[%s] list vpc for cond sync subnet failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
			return nil, err
		}
		for _, vpc := range vpcResult.Details {
			vpcKeys[huaweiSubnetVpcKey{region: region, cloudVpcID: vpc.CloudID}] = struct{}{}
		}
		if len(vpcResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
	}
	return vpcKeys, nil
}
