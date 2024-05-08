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

package securitygroup

import (
	"fmt"
	"net/http"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// InitGcpFirewallRuleService initial the gcp firewall rule service
func InitGcpFirewallRuleService(cap *capability.Capability) {
	svc := &gcpFirewallRuleSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/batch/create",
		svc.BatchCreateGcpFirewallRule)
	h.Add("BatchUpdateGcpFirewallRule", http.MethodPatch, "/vendors/gcp/firewalls/rules/batch/update",
		svc.BatchUpdateGcpFirewallRule)
	h.Add("ListGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/list", svc.ListGcpFirewallRule)
	h.Add("BatchDeleteGcpFirewallRule", http.MethodDelete, "/vendors/gcp/firewalls/rules/batch",
		svc.BatchDeleteGcpFirewallRule)

	h.Load(cap.WebService)
}

type gcpFirewallRuleSvc struct {
	dao dao.Set
}

// BatchCreateGcpFirewallRule batch create gcp firewall rule.
func (svc gcpFirewallRuleSvc) BatchCreateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpFirewallRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		createRules := make([]*tablecloud.GcpFirewallRuleTable, 0, len(req.FirewallRules))
		for _, rule := range req.FirewallRules {
			createRules = append(createRules, &tablecloud.GcpFirewallRuleTable{
				CloudID:               rule.CloudID,
				AccountID:             rule.AccountID,
				Name:                  rule.Name,
				Priority:              rule.Priority,
				Memo:                  rule.Memo,
				CloudVpcID:            rule.CloudVpcID,
				VpcSelfLink:           rule.VpcSelfLink,
				VpcID:                 rule.VpcId,
				SourceRanges:          rule.SourceRanges,
				DestinationRanges:     rule.DestinationRanges,
				SourceTags:            rule.SourceTags,
				TargetTags:            rule.TargetTags,
				SourceServiceAccounts: rule.SourceServiceAccounts,
				TargetServiceAccounts: rule.TargetServiceAccounts,
				Denied:                rule.Denied,
				Allowed:               rule.Allowed,
				BkBizID:               rule.BkBizID,
				Type:                  rule.Type,
				LogEnable:             rule.LogEnable,
				Disabled:              rule.Disabled,
				SelfLink:              rule.SelfLink,
				Creator:               cts.Kit.User,
				Reviser:               cts.Kit.User,
			})
		}

		ids, err := svc.dao.GcpFirewallRule().BatchCreateWithTx(cts.Kit, txn, createRules)
		if err != nil {
			return nil, fmt.Errorf("create gcp firewall rule failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create gcp firewall rule but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateGcpFirewallRule batch update gcp firewall rule.
func (svc gcpFirewallRuleSvc) BatchUpdateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpFirewallRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, rule := range req.FirewallRules {
			update := &tablecloud.GcpFirewallRuleTable{
				AccountID:             rule.AccountID,
				CloudID:               rule.CloudID,
				Name:                  rule.Name,
				Priority:              rule.Priority,
				Memo:                  rule.Memo,
				CloudVpcID:            rule.CloudVpcID,
				VpcSelfLink:           rule.VpcSelfLink,
				VpcID:                 rule.VpcId,
				SourceRanges:          rule.SourceRanges,
				DestinationRanges:     rule.DestinationRanges,
				SourceTags:            rule.SourceTags,
				TargetTags:            rule.TargetTags,
				SourceServiceAccounts: rule.SourceServiceAccounts,
				TargetServiceAccounts: rule.TargetServiceAccounts,
				Denied:                rule.Denied,
				Allowed:               rule.Allowed,
				BkBizID:               rule.BkBizID,
				Type:                  rule.Type,
				LogEnable:             rule.LogEnable,
				Disabled:              rule.Disabled,
				SelfLink:              rule.SelfLink,
				Reviser:               cts.Kit.User,
			}

			if err := svc.dao.GcpFirewallRule().UpdateByIDWithTx(cts.Kit, txn, rule.ID, update); err != nil {
				logs.Errorf("UpdateByIDWithTx failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update gcp firewall rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListGcpFirewallRule list gcp firewall rule.
func (svc gcpFirewallRuleSvc) ListGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpFirewallRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.GcpFirewallRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list gcp firewall rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.GcpFirewallRuleListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.GcpFirewallRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.GcpFirewallRule{
			ID:                    one.ID,
			CloudID:               one.CloudID,
			Name:                  one.Name,
			Priority:              one.Priority,
			Memo:                  one.Memo,
			CloudVpcID:            one.CloudVpcID,
			VpcSelfLink:           one.VpcSelfLink,
			SourceRanges:          one.SourceRanges,
			BkBizID:               one.BkBizID,
			AccountID:             one.AccountID,
			VpcId:                 one.VpcID,
			DestinationRanges:     one.DestinationRanges,
			SourceTags:            one.SourceTags,
			TargetTags:            one.TargetTags,
			SourceServiceAccounts: one.SourceServiceAccounts,
			TargetServiceAccounts: one.TargetServiceAccounts,
			Denied:                one.Denied,
			Allowed:               one.Allowed,
			Type:                  one.Type,
			LogEnable:             one.LogEnable,
			Disabled:              one.Disabled,
			SelfLink:              one.SelfLink,
			Creator:               one.Creator,
			Reviser:               one.Reviser,
			CreatedAt:             one.CreatedAt.String(),
			UpdatedAt:             one.UpdatedAt.String(),
		})
	}

	return &protocloud.GcpFirewallRuleListResult{Details: details}, nil
}

// BatchDeleteGcpFirewallRule batch delete gcp firewall rule.
func (svc gcpFirewallRuleSvc) BatchDeleteGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpFirewallRuleBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.GcpFirewallRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list gcp firewall rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if err := svc.dao.GcpFirewallRule().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		// TODO: add delete relation operation.

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
