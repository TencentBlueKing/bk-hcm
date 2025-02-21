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

	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// CountSecurityGroupRules list security group rules count
func (svc *securityGroupSvc) CountSecurityGroupRules(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloud.CountSecurityGroupRuleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.listTCloudSecurityGroupRulesCount(cts.Kit, req.SecurityGroupIDs)
	case enumor.Aws:
		return svc.listAwsSecurityGroupRulesCount(cts.Kit, req.SecurityGroupIDs)
	case enumor.HuaWei:
		return svc.listHuaweiSecurityGroupRulesCount(cts.Kit, req.SecurityGroupIDs)
	case enumor.Azure:
		return svc.listAzureSecurityGroupRulesCount(cts.Kit, req.SecurityGroupIDs)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for CountSecurityGroupRules", vendor)
	}
}

func (svc *securityGroupSvc) listTCloudSecurityGroupRulesCount(kt *kit.Kit, ids []string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		resp, err := svc.dao.TCloudSGRule().CountBySecurityGroupIDs(kt,
			tools.ContainersExpression("security_group_id", sgIDs))
		if err != nil {
			logs.Errorf("listTCloudSecurityGroupRulesCount failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for k, v := range resp {
			result[k] = v
		}
	}
	return result, nil
}

func (svc *securityGroupSvc) listHuaweiSecurityGroupRulesCount(kt *kit.Kit, ids []string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		resp, err := svc.dao.HuaWeiSGRule().CountBySecurityGroupIDs(kt,
			tools.ContainersExpression("security_group_id", sgIDs))
		if err != nil {
			logs.Errorf("listHuaweiSecurityGroupRulesCount failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for k, v := range resp {
			result[k] = v
		}
	}
	return result, nil
}

func (svc *securityGroupSvc) listAwsSecurityGroupRulesCount(kt *kit.Kit, ids []string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		resp, err := svc.dao.AwsSGRule().CountBySecurityGroupIDs(kt,
			tools.ContainersExpression("security_group_id", sgIDs))
		if err != nil {
			logs.Errorf("listAwsSecurityGroupRulesCount failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for k, v := range resp {
			result[k] = v
		}
	}
	return result, nil
}

func (svc *securityGroupSvc) listAzureSecurityGroupRulesCount(kt *kit.Kit, ids []string) (map[string]int64, error) {
	// split ids to 100 each
	result := make(map[string]int64)
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		resp, err := svc.dao.AzureSGRule().CountBySecurityGroupIDs(kt,
			tools.ContainersExpression("security_group_id", sgIDs))
		if err != nil {
			logs.Errorf("listAzureSecurityGroupRulesCount failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for k, v := range resp {
			result[k] = v
		}
	}
	return result, nil
}
