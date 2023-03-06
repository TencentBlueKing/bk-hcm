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
	"strconv"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncGcpFirewallRule sync gcp firewall rules to hcm.
func SyncGcpFirewallRule(kt *kit.Kit, req *proto.GcpFirewallSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	listReq := &protocloud.GcpFirewallRuleListReq{
		Filter: tools.EqualExpression("account_id", req.AccountID),
		Page: &core.BasePage{
			Start: uint32(0),
			Limit: core.DefaultMaxPageLimit,
		},
	}
	ids := make([]string, 0)

	for {
		results, err := dataCli.Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("request dataservice list gcp firewall rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, result := range results.Details {
			ids = append(ids, result.ID)
		}

		listReq.Page.Start += uint32(len(results.Details))
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	if len(ids) > 0 {
		req := &protocloud.GcpFirewallRuleBatchDeleteReq{
			Filter: tools.ContainersExpression("id", ids),
		}
		if err := dataCli.Gcp.Firewall.BatchDeleteFirewallRule(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("request dataservice delete gcp firewall rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	_, err := syncCreateGcpFirewallRule(kt, req.AccountID, adaptor, dataCli)
	if err != nil {
		logs.Errorf("create gcp firewall rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func syncCreateGcpFirewallRule(kt *kit.Kit, accountID string,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	client, err := ad.Gcp(kt, accountID)
	if err != nil {
		return nil, err
	}

	ruleCreates := make([]protocloud.GcpFirewallRuleBatchCreate, 0)
	nextToken := ""
	for {
		opt := &firewallrule.ListOption{
			Page: &adcore.GcpPage{
				PageSize: int64(adcore.GcpQueryLimit),
			},
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		resp, token, err := client.ListFirewallRule(kt, opt)
		if err != nil {
			return nil, err
		}

		for _, item := range resp {
			rule := protocloud.GcpFirewallRuleBatchCreate{
				CloudID:    strconv.FormatUint(item.Id, 10),
				AccountID:  accountID,
				Name:       item.Name,
				Priority:   item.Priority,
				Memo:       item.Description,
				CloudVpcID: item.Network,
				// TODO: 待处理和vpc关联字段
				VpcId:                 "todo",
				SourceRanges:          item.SourceRanges,
				BkBizID:               constant.UnassignedBiz,
				DestinationRanges:     item.DestinationRanges,
				SourceTags:            item.SourceTags,
				TargetTags:            item.TargetTags,
				SourceServiceAccounts: item.SourceServiceAccounts,
				TargetServiceAccounts: item.TargetServiceAccounts,
				Type:                  item.Direction,
				LogEnable:             item.LogConfig.Enable,
				Disabled:              item.Disabled,
				SelfLink:              item.SelfLink,
			}

			if len(item.Denied) != 0 {
				sets := make([]corecloud.GcpProtocolSet, 0, len(item.Denied))
				for _, one := range item.Denied {
					sets = append(sets, corecloud.GcpProtocolSet{
						Protocol: one.IPProtocol,
						Port:     one.Ports,
					})
				}
				rule.Denied = sets
			}

			if len(item.Allowed) != 0 {
				sets := make([]corecloud.GcpProtocolSet, 0, len(item.Allowed))
				for _, one := range item.Allowed {
					sets = append(sets, corecloud.GcpProtocolSet{
						Protocol: one.IPProtocol,
						Port:     one.Ports,
					})
				}
				rule.Allowed = sets
			}

			ruleCreates = append(ruleCreates, rule)
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	req := &protocloud.GcpFirewallRuleBatchCreateReq{
		FirewallRules: ruleCreates,
	}
	result, err := dataCli.Gcp.Firewall.BatchCreateFirewallRule(kt.Ctx, kt.Header(), req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
