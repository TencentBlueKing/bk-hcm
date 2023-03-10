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

	"hcm/cmd/hc-service/logics/sync/vpc"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	apicloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// FireWallGcpDiff gcp diff for sync
type FireWallGcpDiff struct {
	FireWall *compute.Firewall
}

// FireWallSyncDS data-service diff for sync
type FireWallSyncDS struct {
	IsUpdated  bool
	HcFireWall cloud.GcpFirewallRule
}

// SyncGcpFirewallRule sync gcp firewall rules to hcm.
func SyncGcpFirewallRule(kt *kit.Kit, req *hcservice.GcpFirewallSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromGcpForFireWallSync(kt, adaptor, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := GetDatasFromDSForGcpFireWallSync(kt, req, dataCli)
	if err != nil {
		return nil, err
	}

	err = diffGcpFireWallSync(kt, cloudMap, dsMap, req, dataCli, adaptor)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetDatasFromDSForGcpFireWallSync get gcp firewall datas from hc
func GetDatasFromDSForGcpFireWallSync(kt *kit.Kit, req *hcservice.GcpFirewallSyncReq,
	dataCli *dataservice.Client) (map[string]*FireWallSyncDS, error) {

	dsMap := make(map[string]*FireWallSyncDS)
	start := 0
	for {
		listReq := &apicloud.GcpFirewallRuleListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			listReq.Filter.Rules = append(listReq.Filter.Rules, filter)
		}

		results, err := dataCli.Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("request dataservice list gcp firewall rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, result := range results.Details {
			sg := new(FireWallSyncDS)
			sg.IsUpdated = false
			sg.HcFireWall = result
			dsMap[result.CloudID] = sg
		}

		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	return dsMap, nil
}

// DiffFireWallSyncDelete ...
func DiffFireWallSyncDelete(kt *kit.Kit, deleteCloudIDs []string,
	dataCli *dataservice.Client) error {

	batchDeleteReq := &apicloud.GcpFirewallRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if err := dataCli.Gcp.Firewall.BatchDeleteFirewallRule(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func getDatasFromGcpForFireWallSync(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient,
	req *hcservice.GcpFirewallSyncReq) (map[string]*FireWallGcpDiff, error) {

	client, err := ad.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := new(firewallrule.ListOption)
	if len(req.CloudIDs) > 0 {
		opt.CloudIDs = converter.StringSliceToUint64Slice(req.CloudIDs)
	}

	results, _, err := client.ListFirewallRule(kt, opt)
	if err != nil {
		return nil, err
	}

	cloudMap := make(map[string]*FireWallGcpDiff)
	for _, one := range results {
		sg := new(FireWallGcpDiff)
		sg.FireWall = one
		cloudMap[fmt.Sprint(one.Id)] = sg
	}

	return cloudMap, nil
}

func diffGcpFireWallSync(kt *kit.Kit, cloudMap map[string]*FireWallGcpDiff, dsMap map[string]*FireWallSyncDS,
	req *hcservice.GcpFirewallSyncReq, dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		err := DiffFireWallSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request DiffFireWallSyncDelete failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := diffFireWallSyncUpdate(kt, cloudMap, req, dsMap, updateCloudIDs, dataCli, adaptor)
		if err != nil {
			logs.Errorf("request diffGcpSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffFireWallSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli, adaptor)
		if err != nil {
			logs.Errorf("request diffGcpDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func isGcpChange(db *FireWallSyncDS, cloud *FireWallGcpDiff, vpcID string) bool {

	if db.HcFireWall.Name != cloud.FireWall.Name {
		return true
	}

	if db.HcFireWall.CloudID != fmt.Sprint(cloud.FireWall.Id) {
		return true
	}

	if db.HcFireWall.VpcId != vpcID {
		return true
	}

	if db.HcFireWall.Priority != cloud.FireWall.Priority {
		return true
	}

	if db.HcFireWall.Memo != cloud.FireWall.Description {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.SourceRanges, cloud.FireWall.SourceRanges) {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.DestinationRanges, cloud.FireWall.DestinationRanges) {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.SourceTags, cloud.FireWall.SourceTags) {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.TargetTags, cloud.FireWall.TargetTags) {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.SourceServiceAccounts, cloud.FireWall.SourceServiceAccounts) {
		return true
	}

	if !assert.IsStringSliceEqual(db.HcFireWall.TargetServiceAccounts, cloud.FireWall.TargetServiceAccounts) {
		return true
	}

	if db.HcFireWall.Type != cloud.FireWall.Direction {
		return true
	}

	if db.HcFireWall.LogEnable != cloud.FireWall.LogConfig.Enable {
		return true
	}

	if db.HcFireWall.Disabled != cloud.FireWall.Disabled {
		return true
	}

	if db.HcFireWall.SelfLink != cloud.FireWall.SelfLink {
		return true
	}

	return false
}

func diffFireWallSyncUpdate(kt *kit.Kit, cloudMap map[string]*FireWallGcpDiff, req *hcservice.GcpFirewallSyncReq,
	dsMap map[string]*FireWallSyncDS, updateCloudIDs []string, dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) error {

	rulesUpdate := make([]apicloud.GcpFirewallRuleBatchUpdate, 0)

	for _, id := range updateCloudIDs {

		vpcID := ""
		vpcHCID, _, err := queryVpcIDBySelfLink(kt, dataCli, cloudMap[id].FireWall.Network)
		if err != nil {
			logs.Errorf("request queryVpcIDBySelfLink failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			vpcID = vpcHCID
		}

		if vpcID == "" {
			req := &vpc.SyncGcpOption{
				AccountID: req.AccountID,
				SelfLinks: []string{cloudMap[id].FireWall.Network},
			}
			_, err := vpc.GcpVpcSync(kt, req, adaptor, dataCli)
			if err != nil {
				logs.Errorf("request to sync gcp vpc logic failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			vpcHCID, _, err := queryVpcIDBySelfLink(kt, dataCli, cloudMap[id].FireWall.Network)
			if err != nil {
				logs.Errorf("request queryVpcIDBySelfLink failed, err: %v, rid: %s", err, kt.Rid)
				return err
			} else {
				vpcID = vpcHCID
			}
		}

		if !isGcpChange(dsMap[id], cloudMap[id], vpcID) {
			continue
		}

		rule := apicloud.GcpFirewallRuleBatchUpdate{
			ID:                    dsMap[id].HcFireWall.ID,
			CloudID:               fmt.Sprint(cloudMap[id].FireWall.Id),
			AccountID:             req.AccountID,
			Name:                  cloudMap[id].FireWall.Name,
			Priority:              cloudMap[id].FireWall.Priority,
			Memo:                  cloudMap[id].FireWall.Description,
			CloudVpcID:            cloudMap[id].FireWall.Network,
			VpcId:                 vpcID,
			SourceRanges:          cloudMap[id].FireWall.SourceRanges,
			BkBizID:               constant.UnassignedBiz,
			DestinationRanges:     cloudMap[id].FireWall.DestinationRanges,
			SourceTags:            cloudMap[id].FireWall.SourceTags,
			TargetTags:            cloudMap[id].FireWall.TargetTags,
			SourceServiceAccounts: cloudMap[id].FireWall.SourceServiceAccounts,
			TargetServiceAccounts: cloudMap[id].FireWall.TargetServiceAccounts,
			Type:                  cloudMap[id].FireWall.Direction,
			LogEnable:             cloudMap[id].FireWall.LogConfig.Enable,
			Disabled:              cloudMap[id].FireWall.Disabled,
			SelfLink:              cloudMap[id].FireWall.SelfLink,
		}

		if len(cloudMap[id].FireWall.Denied) != 0 {
			sets := make([]cloud.GcpProtocolSet, 0, len(cloudMap[id].FireWall.Denied))
			for _, one := range cloudMap[id].FireWall.Denied {
				sets = append(sets, cloud.GcpProtocolSet{
					Protocol: one.IPProtocol,
					Port:     one.Ports,
				})
			}
			rule.Denied = sets
		}

		if len(cloudMap[id].FireWall.Allowed) != 0 {
			sets := make([]cloud.GcpProtocolSet, 0, len(cloudMap[id].FireWall.Allowed))
			for _, one := range cloudMap[id].FireWall.Allowed {
				sets = append(sets, cloud.GcpProtocolSet{
					Protocol: one.IPProtocol,
					Port:     one.Ports,
				})
			}
			rule.Allowed = sets
		}

		rulesUpdate = append(rulesUpdate, rule)
	}

	batchCreateReq := &apicloud.GcpFirewallRuleBatchUpdateReq{
		FirewallRules: rulesUpdate,
	}
	err := dataCli.Gcp.Firewall.BatchUpdateFirewallRule(kt.Ctx, kt.Header(), batchCreateReq)
	if err != nil {
		return err
	}

	return nil
}

func diffFireWallSyncAdd(kt *kit.Kit, cloudMap map[string]*FireWallGcpDiff, req *hcservice.GcpFirewallSyncReq, addCloudIDs []string,
	dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) ([]string, error) {

	rulesCreate := make([]apicloud.GcpFirewallRuleBatchCreate, 0)

	for _, id := range addCloudIDs {

		vpcID := ""
		vpcHCID, _, err := queryVpcIDBySelfLink(kt, dataCli, cloudMap[id].FireWall.Network)
		if err != nil {
			logs.Errorf("request queryVpcIDBySelfLink failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			vpcID = vpcHCID
		}

		if vpcID == "" {
			req := &vpc.SyncGcpOption{
				AccountID: req.AccountID,
				SelfLinks: []string{cloudMap[id].FireWall.Network},
			}
			_, err := vpc.GcpVpcSync(kt, req, adaptor, dataCli)
			if err != nil {
				logs.Errorf("request to sync gcp vpc logic failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			vpcHCID, _, err := queryVpcIDBySelfLink(kt, dataCli, cloudMap[id].FireWall.Network)
			if err != nil {
				logs.Errorf("request queryVpcIDBySelfLink failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			} else {
				vpcID = vpcHCID
			}
		}

		rule := apicloud.GcpFirewallRuleBatchCreate{
			CloudID:               fmt.Sprint(cloudMap[id].FireWall.Id),
			AccountID:             req.AccountID,
			Name:                  cloudMap[id].FireWall.Name,
			Priority:              cloudMap[id].FireWall.Priority,
			Memo:                  cloudMap[id].FireWall.Description,
			CloudVpcID:            cloudMap[id].FireWall.Network,
			VpcId:                 vpcID,
			SourceRanges:          cloudMap[id].FireWall.SourceRanges,
			BkBizID:               constant.UnassignedBiz,
			DestinationRanges:     cloudMap[id].FireWall.DestinationRanges,
			SourceTags:            cloudMap[id].FireWall.SourceTags,
			TargetTags:            cloudMap[id].FireWall.TargetTags,
			SourceServiceAccounts: cloudMap[id].FireWall.SourceServiceAccounts,
			TargetServiceAccounts: cloudMap[id].FireWall.TargetServiceAccounts,
			Type:                  cloudMap[id].FireWall.Direction,
			LogEnable:             cloudMap[id].FireWall.LogConfig.Enable,
			Disabled:              cloudMap[id].FireWall.Disabled,
			SelfLink:              cloudMap[id].FireWall.SelfLink,
		}

		if len(cloudMap[id].FireWall.Denied) != 0 {
			sets := make([]cloud.GcpProtocolSet, 0, len(cloudMap[id].FireWall.Denied))
			for _, one := range cloudMap[id].FireWall.Denied {
				sets = append(sets, cloud.GcpProtocolSet{
					Protocol: one.IPProtocol,
					Port:     one.Ports,
				})
			}
			rule.Denied = sets
		}

		if len(cloudMap[id].FireWall.Allowed) != 0 {
			sets := make([]cloud.GcpProtocolSet, 0, len(cloudMap[id].FireWall.Allowed))
			for _, one := range cloudMap[id].FireWall.Allowed {
				sets = append(sets, cloud.GcpProtocolSet{
					Protocol: one.IPProtocol,
					Port:     one.Ports,
				})
			}
			rule.Allowed = sets
		}

		rulesCreate = append(rulesCreate, rule)
	}

	batchCreateReq := &apicloud.GcpFirewallRuleBatchCreateReq{
		FirewallRules: rulesCreate,
	}

	if len(rulesCreate) <= 0 {
		return make([]string, 0), nil
	}

	result, err := dataCli.Gcp.Firewall.BatchCreateFirewallRule(kt.Ctx, kt.Header(), batchCreateReq)
	if err != nil {
		return nil, err
	}

	return result.IDs, nil
}

func getAddCloudIDs[T any](cloudMap map[string]T, dsMap map[string]*FireWallSyncDS) []string {
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

func getDeleteAndUpdateCloudIDs(dsMap map[string]*FireWallSyncDS) ([]string, []string) {
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

func queryVpcIDBySelfLink(kt *kit.Kit, dataCli *dataservice.Client, selfLink string) (
	string, int64, error) {

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "extension.self_link",
					Op:    filter.JSONEqual.Factory(),
					Value: selfLink,
				},
			},
		},
		Page:   core.DefaultBasePage,
		Fields: []string{"id"},
	}
	vpcResult, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return "", 0, err
	}

	if len(vpcResult.Details) != 1 {
		return "", 0, errf.Newf(errf.RecordNotFound, "vpc: %s not found", selfLink)
	}

	return vpcResult.Details[0].ID, vpcResult.Details[0].BkCloudID, nil
}
