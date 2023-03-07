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

package firewall

import (
	"fmt"
	"net/http"
	"strconv"

	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"google.golang.org/api/compute/v1"
)

// InitFirewallService initial the security group service
func InitFirewallService(cap *capability.Capability) {
	sg := &firewall{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("CreateGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/create", sg.CreateGcpFirewallRule)
	h.Add("DeleteGcpFirewallRule", http.MethodDelete, "/vendors/gcp/firewalls/rules/{id}", sg.DeleteGcpFirewallRule)
	h.Add("UpdateGcpFirewallRule", http.MethodPut, "/vendors/gcp/firewalls/rules/{id}", sg.UpdateGcpFirewallRule)

	h.Load(cap.WebService)
}

type firewall struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
}

// DeleteGcpFirewallRule delete gcp firewall rule.
func (f *firewall) DeleteGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := f.getGcpFirewallRuleByID(cts, id)
	if err != nil {
		logs.Errorf("request dataservice get gcp firewall rule failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := f.ad.Gcp(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &firewallrule.DeleteOption{
		CloudID: rule.CloudID,
	}
	if err := client.DeleteFirewallRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete gcp firewall rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.GcpFirewallRuleBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := f.dataCli.Gcp.Firewall.BatchDeleteFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete gcp firewall rule failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateGcpFirewallRule update gcp firewall rule.
func (f *firewall) UpdateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.GcpFirewallRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rule, err := f.getGcpFirewallRuleByID(cts, id)
	if err != nil {
		logs.Errorf("request dataservice get gcp firewall rule failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := f.ad.Gcp(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &firewallrule.UpdateOption{
		CloudID: rule.CloudID,
		GcpFirewallRule: &firewallrule.Update{
			Description:           req.Memo,
			Priority:              req.Priority,
			SourceTags:            req.SourceTags,
			TargetTags:            req.TargetTags,
			Denied:                req.Denied,
			Allowed:               req.Allowed,
			SourceRanges:          req.SourceRanges,
			DestinationRanges:     req.DestinationRanges,
			Disabled:              req.Disabled,
			SourceServiceAccounts: nil,
			TargetServiceAccounts: nil,
		},
	}
	if err := client.UpdateFirewallRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update gcp firewall rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.GcpFirewallRuleBatchUpdateReq{
		FirewallRules: []protocloud.GcpFirewallRuleBatchUpdate{
			{
				ID:                id,
				Priority:          req.Priority,
				Memo:              req.Memo,
				SourceRanges:      req.SourceRanges,
				DestinationRanges: req.DestinationRanges,
				SourceTags:        req.SourceTags,
				TargetTags:        req.TargetTags,
				Denied:            req.Denied,
				Allowed:           req.Allowed,
				Disabled:          req.Disabled,
			},
		},
	}
	if err := f.dataCli.Gcp.Firewall.BatchUpdateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateFirewallRule failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (f *firewall) getGcpFirewallRuleByID(cts *rest.Contexts, id string) (*corecloud.GcpFirewallRule, error) {

	listReq := &protocloud.GcpFirewallRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	listResp, err := f.dataCli.Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice get gcp firewall rule failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "gcp firewall rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// CreateGcpFirewallRule ...
func (f *firewall) CreateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpFirewallRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	item, err := f.createFirewallRule(cts, req)
	if err != nil {
		return nil, err
	}

	vpcID, err := f.getVpcIDByCloudVpcID(cts.Kit, req.CloudVpcID)
	if err != nil {
		return nil, err
	}

	rule := protocloud.GcpFirewallRuleBatchCreate{
		CloudID:               strconv.FormatUint(item.Id, 10),
		AccountID:             req.AccountID,
		Name:                  item.Name,
		Priority:              item.Priority,
		Memo:                  item.Description,
		CloudVpcID:            item.Network,
		VpcId:                 vpcID,
		SourceRanges:          item.SourceRanges,
		BkBizID:               req.BkBizID,
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

	createReq := &protocloud.GcpFirewallRuleBatchCreateReq{
		FirewallRules: []protocloud.GcpFirewallRuleBatchCreate{rule},
	}
	result, err := f.dataCli.Gcp.Firewall.BatchCreateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create firewall rule failed, err: %v, req: %v, rid: %s",
			err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (f *firewall) createFirewallRule(cts *rest.Contexts, req *proto.GcpFirewallRuleCreateReq) (*compute.Firewall,
	error) {

	client, err := f.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &firewallrule.CreateOption{
		Name:              req.Name,
		CloudVpcID:        req.CloudVpcID,
		Description:       req.Memo,
		Priority:          req.Priority,
		SourceTags:        req.SourceTags,
		TargetTags:        req.TargetTags,
		Denied:            req.Denied,
		Allowed:           req.Allowed,
		SourceRanges:      req.SourceRanges,
		DestinationRanges: req.DestinationRanges,
		Disabled:          req.Disabled,
	}
	ruleID, err := client.CreateFirewallRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request gcp create firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	items, _, err := client.ListFirewallRule(cts.Kit, &firewallrule.ListOption{
		CloudIDs: []uint64{ruleID},
	})
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("create firewall rule success, but list failed, ruleID: %d", ruleID)
	}

	return items[0], nil
}

func (f *firewall) getVpcIDByCloudVpcID(kt *kit.Kit, cloudVpcID string) (string, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudVpcID),
		Page:   core.DefaultBasePage,
		Fields: []string{"id"},
	}
	result, err := f.dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("request dataservice to list vpc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return "", err
	}

	if len(result.Details) == 0 {
		return "", errf.Newf(errf.RecordNotFound, "vpc(cloud_id=%s) not found", cloudVpcID)
	}

	return result.Details[0].CloudID, nil
}

// SyncGcpFirewallRule sync gcp to hcm
func (f *firewall) SyncGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpFirewallSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &protocloud.GcpFirewallRuleListReq{
		Filter: tools.EqualExpression("account_id", req.AccountID),
		Page: &core.BasePage{
			Start: uint32(0),
			Limit: core.DefaultMaxPageLimit,
		},
	}
	ids := make([]string, 0)

	for {
		results, err := f.dataCli.Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), listReq)
		if err != nil {
			logs.Errorf("request dataservice list gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		if err := f.dataCli.Gcp.Firewall.BatchDeleteFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
			logs.Errorf("request dataservice delete gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	_, err := f.SyncCreateGcpFirewallRule(cts, req.AccountID)
	if err != nil {
		logs.Errorf("create gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncCreateGcpFirewallRule create GcpFirewallRule for sync
func (f *firewall) SyncCreateGcpFirewallRule(cts *rest.Contexts, accountID string) (interface{}, error) {
	client, err := f.ad.Gcp(cts.Kit, accountID)
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

		resp, token, err := client.ListFirewallRule(cts.Kit, opt)
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
	result, err := f.dataCli.Gcp.Firewall.BatchCreateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
