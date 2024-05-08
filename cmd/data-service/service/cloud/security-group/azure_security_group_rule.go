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
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// initAzureSGRuleService initial the azure security group rule service
func initAzureSGRuleService(cap *capability.Capability) {
	svc := &azureSGRuleSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateAzureRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/batch/create",
		svc.BatchCreateAzureRule)
	h.Add("BatchUpdateAzureRule", "PUT", "/vendors/azure/security_groups/{security_group_id}/rules/batch",
		svc.BatchUpdateAzureRule)
	h.Add("ListAzureRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/list",
		svc.ListAzureRule)
	h.Add("DeleteAzureRule", "DELETE", "/vendors/azure/security_groups/{security_group_id}/rules/batch",
		svc.DeleteAzureRule)

	h.Load(cap.WebService)
}

type azureSGRuleSvc struct {
	dao dao.Set
}

// BatchCreateAzureRule create azure rule.
func (svc *azureSGRuleSvc) BatchCreateAzureRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AzureSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rules := make([]*tablecloud.AzureSecurityGroupRuleTable, 0, len(req.Rules))
	for _, rule := range req.Rules {
		rules = append(rules, &tablecloud.AzureSecurityGroupRuleTable{
			Region:                              rule.Region,
			CloudID:                             rule.CloudID,
			CloudSecurityGroupID:                rule.CloudSecurityGroupID,
			AccountID:                           rule.AccountID,
			SecurityGroupID:                     rule.SecurityGroupID,
			Type:                                string(rule.Type),
			ProvisioningState:                   rule.ProvisioningState,
			Etag:                                rule.Etag,
			Name:                                rule.Name,
			Memo:                                rule.Memo,
			Protocol:                            rule.Protocol,
			DestinationAddressPrefix:            rule.DestinationAddressPrefix,
			DestinationAddressPrefixes:          convStringSlice(rule.DestinationAddressPrefixes),
			CloudDestinationAppSecurityGroupIDs: convStringSlice(rule.CloudDestinationAppSecurityGroupIDs),
			DestinationPortRange:                rule.DestinationPortRange,
			DestinationPortRanges:               convStringSlice(rule.DestinationPortRanges),
			SourceAddressPrefix:                 rule.SourceAddressPrefix,
			SourceAddressPrefixes:               convStringSlice(rule.SourceAddressPrefixes),
			CloudSourceAppSecurityGroupIDs:      convStringSlice(rule.CloudSourceAppSecurityGroupIDs),
			SourcePortRange:                     rule.SourcePortRange,
			SourcePortRanges:                    convStringSlice(rule.SourcePortRanges),
			Priority:                            rule.Priority,
			Access:                              rule.Access,
			Creator:                             cts.Kit.User,
			Reviser:                             cts.Kit.User,
		})
	}
	ruleIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ruleIDs, err := svc.dao.AzureSGRule().BatchCreateWithTx(cts.Kit, txn, rules)
		if err != nil {
			return nil, fmt.Errorf("batch create azure security group rule failed, err: %v", err)
		}

		return ruleIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := ruleIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create azure security group rule but return id type is not string, id type: %v",
			reflect.TypeOf(ruleIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func convStringSlice(data []*string) []string {
	if data == nil {
		return nil
	}

	return converter.PtrToSlice(data)
}

func convStringPtrSlice(data []string) []*string {
	if data == nil {
		return nil
	}

	return converter.SliceToPtr(data)
}

// BatchUpdateAzureRule update azure rule.
func (svc *azureSGRuleSvc) BatchUpdateAzureRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AzureSGRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Rules {
			rule := &tablecloud.AzureSecurityGroupRuleTable{
				Region:                              one.Region,
				CloudID:                             one.CloudID,
				CloudSecurityGroupID:                one.CloudSecurityGroupID,
				AccountID:                           one.AccountID,
				SecurityGroupID:                     one.SecurityGroupID,
				Type:                                string(one.Type),
				ProvisioningState:                   one.ProvisioningState,
				Etag:                                one.Etag,
				Name:                                one.Name,
				Memo:                                one.Memo,
				Protocol:                            one.Protocol,
				DestinationAddressPrefix:            one.DestinationAddressPrefix,
				DestinationAddressPrefixes:          convStringSlice(one.DestinationAddressPrefixes),
				CloudDestinationAppSecurityGroupIDs: convStringSlice(one.CloudDestinationAppSecurityGroupIDs),
				DestinationPortRange:                one.DestinationPortRange,
				DestinationPortRanges:               convStringSlice(one.DestinationPortRanges),
				SourceAddressPrefix:                 one.SourceAddressPrefix,
				SourceAddressPrefixes:               convStringSlice(one.SourceAddressPrefixes),
				CloudSourceAppSecurityGroupIDs:      convStringSlice(one.CloudSourceAppSecurityGroupIDs),
				SourcePortRange:                     one.SourcePortRange,
				SourcePortRanges:                    convStringSlice(one.SourcePortRanges),
				Priority:                            one.Priority,
				Access:                              one.Access,
				Reviser:                             cts.Kit.User,
			}

			flt := &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "id",
						Op:    filter.Equal.Factory(),
						Value: one.ID,
					},
					&filter.AtomRule{
						Field: "security_group_id",
						Op:    filter.Equal.Factory(),
						Value: sgID,
					},
				},
			}
			if err := svc.dao.AzureSGRule().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update azure security group rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListAzureRule list azure rule.
func (svc *azureSGRuleSvc) ListAzureRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AzureSGRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.SGRuleListOption{
		SecurityGroupID: sgID,
		Fields:          req.Field,
		Filter:          req.Filter,
		Page:            req.Page,
	}
	result, err := svc.dao.AzureSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure security group rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.AzureSGRuleListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.AzureSecurityGroupRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.AzureSecurityGroupRule{
			ID:                                  one.ID,
			Region:                              one.Region,
			CloudID:                             one.CloudID,
			Etag:                                one.Etag,
			Name:                                one.Name,
			Memo:                                one.Memo,
			DestinationAddressPrefix:            one.DestinationAddressPrefix,
			DestinationAddressPrefixes:          convStringPtrSlice(one.DestinationAddressPrefixes),
			CloudDestinationAppSecurityGroupIDs: convStringPtrSlice(one.CloudDestinationAppSecurityGroupIDs),
			DestinationPortRange:                one.DestinationPortRange,
			DestinationPortRanges:               convStringPtrSlice(one.DestinationPortRanges),
			Protocol:                            one.Protocol,
			ProvisioningState:                   one.ProvisioningState,
			SourceAddressPrefix:                 one.SourceAddressPrefix,
			SourceAddressPrefixes:               convStringPtrSlice(one.SourceAddressPrefixes),
			CloudSourceAppSecurityGroupIDs:      convStringPtrSlice(one.CloudSourceAppSecurityGroupIDs),
			SourcePortRange:                     one.SourcePortRange,
			SourcePortRanges:                    convStringPtrSlice(one.SourcePortRanges),
			Priority:                            one.Priority,
			Type:                                enumor.SecurityGroupRuleType(one.Type),
			Access:                              one.Access,
			CloudSecurityGroupID:                one.CloudSecurityGroupID,
			AccountID:                           one.AccountID,
			SecurityGroupID:                     one.SecurityGroupID,
			Creator:                             one.Creator,
			Reviser:                             one.Reviser,
			CreatedAt:                           one.CreatedAt.String(),
			UpdatedAt:                           one.UpdatedAt.String(),
		})
	}

	return &protocloud.AzureSGRuleListResult{Details: details}, nil
}

// DeleteAzureRule delete azure rule.
func (svc *azureSGRuleSvc) DeleteAzureRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AzureSGRuleBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.SGRuleListOption{
		SecurityGroupID: sgID,
		Fields:          []string{"id"},
		Filter:          req.Filter,
		Page:            core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.AzureSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure security group rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	delFilter := tools.ContainersExpression("id", delIDs)
	if err := svc.dao.AzureSGRule().Delete(cts.Kit, delFilter); err != nil {
		logs.Errorf("delete azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
