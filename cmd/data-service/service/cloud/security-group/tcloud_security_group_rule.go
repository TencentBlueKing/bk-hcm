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

// Package securitygroup 提供腾讯云安全组规则的数据服务接口
// 包含安全组规则的增删改查等核心功能
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

// initTCloudSGRuleService initial the tcloud security group rule service
func initTCloudSGRuleService(cap *capability.Capability) {
	svc := &tcloudSGRuleSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateTCloudRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch/create",
		svc.BatchCreateTCloudRule)
	h.Add("BatchUpdateTCloudRule", "PUT", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch",
		svc.BatchUpdateTCloudRule)
	h.Add("ListTCloudRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/list",
		svc.ListTCloudRule)
	h.Add("DeleteTCloudRule", "DELETE", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch",
		svc.DeleteTCloudRule)
	h.Add("ListTCloudRuleExt", "POST", "/vendors/tcloud/security_groups/rules/list", svc.ListTCloudRuleExt)

	h.Load(cap.WebService)
}

// tcloudSGRuleSvc 腾讯云安全组规则服务结构体
// 封装了安全组规则相关的业务逻辑处理
type tcloudSGRuleSvc struct {
	dao dao.Set // 数据访问对象集合，用于数据库操作
}

// BatchCreateTCloudRule 批量创建腾讯云安全组规则
// 支持一次性创建多个安全组规则，提高操作效率
// cts: REST上下文，包含请求参数和用户信息
// 返回: 创建成功的规则ID列表或错误信息
func (svc *tcloudSGRuleSvc) BatchCreateTCloudRule(cts *rest.Contexts) (interface{}, error) {
	// 从URL路径中获取安全组ID
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	// 解析请求体中的创建规则参数
	req := new(protocloud.TCloudSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 验证请求参数的有效性
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 将请求中的规则数据转换为数据库表结构
	rules := make([]*tablecloud.TCloudSecurityGroupRuleTable, 0, len(req.Rules))
	for _, rule := range req.Rules {
		rules = append(rules, &tablecloud.TCloudSecurityGroupRuleTable{
			Region:                     rule.Region,
			CloudPolicyIndex:           rule.CloudPolicyIndex,
			Version:                    rule.Version,
			Type:                       string(rule.Type),
			CloudSecurityGroupID:       rule.CloudSecurityGroupID,
			SecurityGroupID:            rule.SecurityGroupID,
			AccountID:                  rule.AccountID,
			Action:                     rule.Action,
			Protocol:                   rule.Protocol,
			Port:                       rule.Port,
			ServiceID:                  rule.ServiceID,
			CloudServiceID:             rule.CloudServiceID,
			ServiceGroupID:             rule.ServiceGroupID,
			CloudServiceGroupID:        rule.CloudServiceGroupID,
			IPv4Cidr:                   rule.IPv4Cidr,
			IPv6Cidr:                   rule.IPv6Cidr,
			CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			AddressID:                  rule.AddressID,
			CloudAddressID:             rule.CloudAddressID,
			AddressGroupID:             rule.AddressGroupID,
			CloudAddressGroupID:        rule.CloudAddressGroupID,
			Memo:                       rule.Memo,
			Creator:                    cts.Kit.User,
			Reviser:                    cts.Kit.User,
		})
	}

	ruleIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ruleIDs, err := svc.dao.TCloudSGRule().BatchCreateOrUpdateWithTx(cts.Kit, txn, rules)
		if err != nil {
			return nil, fmt.Errorf("batch create tcloud security group rule failed, err: %v", err)
		}

		return ruleIDs, nil
	})
	if err != nil {
		return nil, err
	}

	// 类型断言，确保返回的ID列表是字符串类型
	ids, ok := ruleIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create tcloud security group rule but return id type is not string, id type: %v",
			reflect.TypeOf(ruleIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateTCloudRule 批量更新腾讯云安全组规则
// 支持一次性更新多个安全组规则的配置信息
// cts: REST上下文，包含请求参数和用户信息
// 返回: 更新结果或错误信息
func (svc *tcloudSGRuleSvc) BatchUpdateTCloudRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.TCloudSGRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Rules {
			rule := &tablecloud.TCloudSecurityGroupRuleTable{
				Region:                     one.Region,
				CloudPolicyIndex:           one.CloudPolicyIndex,
				Version:                    one.Version,
				Type:                       string(one.Type),
				CloudSecurityGroupID:       one.CloudSecurityGroupID,
				SecurityGroupID:            one.SecurityGroupID,
				AccountID:                  one.AccountID,
				Action:                     one.Action,
				Protocol:                   one.Protocol,
				Port:                       one.Port,
				ServiceID:                  one.ServiceID,
				CloudServiceID:             one.CloudServiceID,
				ServiceGroupID:             one.ServiceGroupID,
				CloudServiceGroupID:        one.CloudServiceGroupID,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
				AddressID:                  one.AddressID,
				CloudAddressID:             one.CloudAddressID,
				AddressGroupID:             one.AddressGroupID,
				CloudAddressGroupID:        one.CloudAddressGroupID,
				Memo:                       one.Memo,
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
			if err := svc.dao.TCloudSGRule().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update tcloud security group rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListTCloudRule 查询指定安全组的规则列表
// 根据安全组ID查询其下所有的安全组规则
// cts: REST上下文，包含请求参数和用户信息
// 返回: 安全组规则列表或错误信息
func (svc *tcloudSGRuleSvc) ListTCloudRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.TCloudSGRuleListReq)
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
	result, err := svc.dao.TCloudSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud security group rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TCloudSGRuleListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.TCloudSecurityGroupRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.TCloudSecurityGroupRule{
			ID:                         one.ID,
			Region:                     one.Region,
			CloudPolicyIndex:           one.CloudPolicyIndex,
			Version:                    one.Version,
			Protocol:                   one.Protocol,
			Port:                       one.Port,
			ServiceID:                  one.ServiceID,
			CloudServiceID:             one.CloudServiceID,
			ServiceGroupID:             one.ServiceGroupID,
			CloudServiceGroupID:        one.CloudServiceGroupID,
			IPv4Cidr:                   one.IPv4Cidr,
			IPv6Cidr:                   one.IPv6Cidr,
			CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
			AddressID:                  one.AddressID,
			CloudAddressID:             one.CloudAddressID,
			AddressGroupID:             one.AddressGroupID,
			CloudAddressGroupID:        one.CloudAddressGroupID,
			Action:                     one.Action,
			Memo:                       one.Memo,
			Type:                       enumor.SecurityGroupRuleType(one.Type),
			CloudSecurityGroupID:       one.CloudSecurityGroupID,
			SecurityGroupID:            one.SecurityGroupID,
			AccountID:                  one.AccountID,
			Creator:                    one.Creator,
			Reviser:                    one.Reviser,
			CreatedAt:                  one.CreatedAt.String(),
			UpdatedAt:                  one.UpdatedAt.String(),
		})
	}

	return &protocloud.TCloudSGRuleListResult{Details: details}, nil
}

// DeleteTCloudRule delete tcloud rule.
func (svc *tcloudSGRuleSvc) DeleteTCloudRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(protocloud.TCloudSGRuleBatchDeleteReq)
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
	listResp, err := svc.dao.TCloudSGRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud security group rule failed, err: %v", err)
	}

	// 如果没有找到符合条件的规则，直接返回
	if len(listResp.Details) == 0 {
		return nil, nil
	}

	// 提取所有需要删除的规则ID
	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	// 构建删除过滤条件
	delFilter := tools.ContainersExpression("id", delIDs)
	if err := svc.dao.TCloudSGRule().Delete(cts.Kit, delFilter); err != nil {
		logs.Errorf("delete tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListTCloudRuleExt list tcloud rule ext.
func (svc *tcloudSGRuleSvc) ListTCloudRuleExt(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.TCloudSGRuleListReq)
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
	result, err := svc.dao.TCloudSGRule().ListExt(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud security group rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TCloudSGRuleListExtResult{Count: result.Count}, nil
	}

	existSG := make(map[string]struct{})
	sgDetails := make([]corecloud.BaseSecurityGroup, 0)
	details := make([]corecloud.TCloudSecurityGroupRule, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.TCloudSecurityGroupRule{
			ID:                         one.ID,
			Region:                     one.Region,
			CloudPolicyIndex:           one.CloudPolicyIndex,
			Version:                    one.Version,
			Protocol:                   one.Protocol,
			Port:                       one.Port,
			ServiceID:                  one.ServiceID,
			CloudServiceID:             one.CloudServiceID,
			ServiceGroupID:             one.ServiceGroupID,
			CloudServiceGroupID:        one.CloudServiceGroupID,
			IPv4Cidr:                   one.IPv4Cidr,
			IPv6Cidr:                   one.IPv6Cidr,
			CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
			AddressID:                  one.AddressID,
			CloudAddressID:             one.CloudAddressID,
			AddressGroupID:             one.AddressGroupID,
			CloudAddressGroupID:        one.CloudAddressGroupID,
			Action:                     one.Action,
			Memo:                       one.Memo,
			Type:                       enumor.SecurityGroupRuleType(one.Type),
			CloudSecurityGroupID:       one.CloudSecurityGroupID,
			SecurityGroupID:            one.SecurityGroupID,
			AccountID:                  one.AccountID,
			Creator:                    one.Creator,
			Reviser:                    one.Reviser,
			CreatedAt:                  one.CreatedAt.String(),
			UpdatedAt:                  one.UpdatedAt.String(),
		})
		if _, ok := existSG[one.SecurityGroupID]; !ok {
			existSG[one.SecurityGroupID] = struct{}{}
			sgDetails = append(sgDetails, corecloud.BaseSecurityGroup{
				ID:      one.SecurityGroupID,
				CloudID: one.CloudSecurityGroupID,
			})
		}
	}

	return &protocloud.TCloudSGRuleListExtResult{SecurityGroupRule: details, SecurityGroup: sgDetails}, nil
}
