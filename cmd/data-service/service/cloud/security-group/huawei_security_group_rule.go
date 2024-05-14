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

// initHuaWeiSGRuleService initial the huawei security group rule service
func initHuaWeiSGRuleService(cap *capability.Capability) {
	svc := &huaweiSGRuleSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateHuaWeiRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/batch/create",
		svc.BatchCreateHuaWeiRule)
	h.Add("BatchUpdateHuaWeiRule", "PUT", "/vendors/huawei/security_groups/{security_group_id}/rules/batch",
		svc.BatchUpdateHuaWeiRule)
	h.Add("ListHuaWeiRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/list",
		svc.ListHuaWeiRule)
	h.Add("DeleteHuaWeiRule", "DELETE", "/vendors/huawei/security_groups/{security_group_id}/rules/batch",
		svc.DeleteHuaWeiRule)

	h.Load(cap.WebService)
}

type huaweiSGRuleSvc struct {
	dao dao.Set
}

// BatchCreateHuaWeiRule create huawei rule.
func (svc *huaweiSGRuleSvc) BatchCreateHuaWeiRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.HuaWeiSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rules := make([]*tablecloud.HuaWeiSecurityGroupRuleTable, 0, len(req.Rules))
	for _, rule := range req.Rules {
		rules = append(rules, &tablecloud.HuaWeiSecurityGroupRuleTable{
			Region:                    rule.Region,
			CloudID:                   rule.CloudID,
			Type:                      string(rule.Type),
			CloudSecurityGroupID:      rule.CloudSecurityGroupID,
			SecurityGroupID:           rule.SecurityGroupID,
			AccountID:                 rule.AccountID,
			CloudProjectID:            rule.CloudProjectID,
			Memo:                      rule.Memo,
			Protocol:                  rule.Protocol,
			Ethertype:                 rule.Ethertype,
			CloudRemoteGroupID:        rule.CloudRemoteGroupID,
			RemoteIPPrefix:            rule.RemoteIPPrefix,
			Action:                    rule.Action,
			CloudRemoteAddressGroupID: rule.CloudRemoteAddressGroupID,
			Port:                      rule.Port,
			Priority:                  rule.Priority,
			Creator:                   cts.Kit.User,
			Reviser:                   cts.Kit.User,
		})
	}
	ruleIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ruleIDs, err := svc.dao.HuaWeiSGRule().BatchCreateWithTx(cts.Kit, txn, rules)
		if err != nil {
			return nil, fmt.Errorf("batch create huawei security group rule failed, err: %v", err)
		}

		return ruleIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := ruleIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create huawei security group rule but return id type is not string, id type: %v",
			reflect.TypeOf(ruleIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateHuaWeiRule update huawei rule.
func (svc *huaweiSGRuleSvc) BatchUpdateHuaWeiRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.HuaWeiSGRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Rules {
			rule := &tablecloud.HuaWeiSecurityGroupRuleTable{
				Region:                    one.Region,
				CloudID:                   one.CloudID,
				Type:                      string(one.Type),
				CloudSecurityGroupID:      one.CloudSecurityGroupID,
				SecurityGroupID:           one.SecurityGroupID,
				AccountID:                 one.AccountID,
				CloudProjectID:            one.CloudProjectID,
				Memo:                      one.Memo,
				Protocol:                  one.Protocol,
				Action:                    one.Action,
				Ethertype:                 one.Ethertype,
				CloudRemoteGroupID:        one.CloudRemoteGroupID,
				RemoteIPPrefix:            one.RemoteIPPrefix,
				CloudRemoteAddressGroupID: one.CloudRemoteAddressGroupID,
				Port:                      one.Port,
				Priority:                  one.Priority,
				Reviser:                   cts.Kit.User,
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
			if err := svc.dao.HuaWeiSGRule().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update huawei security group rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListHuaWeiRule list huawei rule.
func (svc *huaweiSGRuleSvc) ListHuaWeiRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.HuaWeiSGRuleListReq)
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
	result, err := svc.dao.HuaWeiSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei security group rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.HuaWeiSGRuleListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.HuaWeiSecurityGroupRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.HuaWeiSecurityGroupRule{
			ID:                        one.ID,
			Region:                    one.Region,
			CloudID:                   one.CloudID,
			Memo:                      one.Memo,
			Protocol:                  one.Protocol,
			Ethertype:                 one.Ethertype,
			CloudRemoteGroupID:        one.CloudRemoteGroupID,
			RemoteIPPrefix:            one.RemoteIPPrefix,
			Action:                    one.Action,
			CloudRemoteAddressGroupID: one.CloudRemoteAddressGroupID,
			Port:                      one.Port,
			Priority:                  one.Priority,
			Type:                      enumor.SecurityGroupRuleType(one.Type),
			CloudSecurityGroupID:      one.CloudSecurityGroupID,
			CloudProjectID:            one.CloudProjectID,
			AccountID:                 one.AccountID,
			SecurityGroupID:           one.SecurityGroupID,
			Creator:                   one.Creator,
			Reviser:                   one.Reviser,
			CreatedAt:                 one.CreatedAt.String(),
			UpdatedAt:                 one.UpdatedAt.String(),
		})
	}

	return &protocloud.HuaWeiSGRuleListResult{Details: details}, nil
}

// DeleteHuaWeiRule delete huawei rule.
func (svc *huaweiSGRuleSvc) DeleteHuaWeiRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.HuaWeiSGRuleBatchDeleteReq)
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
	listResp, err := svc.dao.HuaWeiSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei security group rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	delFilter := tools.ContainersExpression("id", delIDs)
	if err := svc.dao.HuaWeiSGRule().Delete(cts.Kit, delFilter); err != nil {
		logs.Errorf("delete huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
