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

	"github.com/jmoiron/sqlx"
)

// initAwsSGRuleService initial the aws security group rule service
func initAwsSGRuleService(cap *capability.Capability) {
	svc := &awsSGRuleSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateAwsRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/batch/create",
		svc.BatchCreateAwsRule)
	h.Add("BatchUpdateAwsRule", "PUT", "/vendors/aws/security_groups/{security_group_id}/rules/batch",
		svc.BatchUpdateAwsRule)
	h.Add("ListAwsRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/list",
		svc.ListAwsRule)
	h.Add("DeleteAwsRule", "DELETE", "/vendors/aws/security_groups/{security_group_id}/rules/batch",
		svc.DeleteAwsRule)

	h.Load(cap.WebService)
}

type awsSGRuleSvc struct {
	dao dao.Set
}

// BatchCreateAwsRule create aws rule.
func (svc *awsSGRuleSvc) BatchCreateAwsRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AwsSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rules := make([]*tablecloud.AwsSecurityGroupRuleTable, 0, len(req.Rules))
	for _, rule := range req.Rules {
		rules = append(rules, &tablecloud.AwsSecurityGroupRuleTable{
			Region:                     rule.Region,
			CloudID:                    rule.CloudID,
			IPv4Cidr:                   rule.IPv4Cidr,
			IPv6Cidr:                   rule.IPv6Cidr,
			Memo:                       rule.Memo,
			Type:                       string(rule.Type),
			FromPort:                   rule.FromPort,
			ToPort:                     rule.ToPort,
			Protocol:                   rule.Protocol,
			CloudPrefixListID:          rule.CloudPrefixListID,
			CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			CloudSecurityGroupID:       rule.CloudSecurityGroupID,
			CloudGroupOwnerID:          rule.CloudGroupOwnerID,
			SecurityGroupID:            rule.SecurityGroupID,
			AccountID:                  rule.AccountID,
			Creator:                    cts.Kit.User,
			Reviser:                    cts.Kit.User,
		})
	}
	ruleIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ruleIDs, err := svc.dao.AwsSGRule().BatchCreateWithTx(cts.Kit, txn, rules)
		if err != nil {
			return nil, fmt.Errorf("batch create aws security group rule failed, err: %v", err)
		}

		return ruleIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := ruleIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create aws security group rule but return id type is not string, id type: %v",
			reflect.TypeOf(ruleIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateAwsRule update aws rule.
func (svc *awsSGRuleSvc) BatchUpdateAwsRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AwsSGRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Rules {
			rule := &tablecloud.AwsSecurityGroupRuleTable{
				Region:                     one.Region,
				CloudID:                    one.CloudID,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				Memo:                       one.Memo,
				Type:                       string(one.Type),
				FromPort:                   one.FromPort,
				ToPort:                     one.ToPort,
				Protocol:                   one.Protocol,
				CloudPrefixListID:          one.CloudPrefixListID,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
				CloudSecurityGroupID:       one.CloudSecurityGroupID,
				CloudGroupOwnerID:          one.CloudGroupOwnerID,
				SecurityGroupID:            one.SecurityGroupID,
				AccountID:                  one.AccountID,
				Reviser:                    cts.Kit.User,
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
			if err := svc.dao.AwsSGRule().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update aws security group rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListAwsRule list aws rule.
func (svc *awsSGRuleSvc) ListAwsRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AwsSGRuleListReq)
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
	result, err := svc.dao.AwsSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws security group rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.AwsSGRuleListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.AwsSecurityGroupRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.AwsSecurityGroupRule{
			ID:                         one.ID,
			Region:                     one.Region,
			CloudID:                    one.CloudID,
			IPv4Cidr:                   one.IPv4Cidr,
			IPv6Cidr:                   one.IPv6Cidr,
			Memo:                       one.Memo,
			FromPort:                   one.FromPort,
			ToPort:                     one.ToPort,
			Type:                       enumor.SecurityGroupRuleType(one.Type),
			Protocol:                   one.Protocol,
			CloudPrefixListID:          one.CloudPrefixListID,
			CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
			CloudSecurityGroupID:       one.CloudSecurityGroupID,
			CloudGroupOwnerID:          one.CloudGroupOwnerID,
			AccountID:                  one.AccountID,
			SecurityGroupID:            one.SecurityGroupID,
			Creator:                    one.Creator,
			Reviser:                    one.Reviser,
			CreatedAt:                  one.CreatedAt.String(),
			UpdatedAt:                  one.UpdatedAt.String(),
		})
	}

	return &protocloud.AwsSGRuleListResult{Details: details}, nil
}

// DeleteAwsRule delete aws rule.
func (svc *awsSGRuleSvc) DeleteAwsRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.AwsSGRuleBatchDeleteReq)
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
	listResp, err := svc.dao.AwsSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws security group rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	delFilter := tools.ContainersExpression("id", delIDs)
	if err := svc.dao.AwsSGRule().Delete(cts.Kit, delFilter); err != nil {
		logs.Errorf("delete aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
