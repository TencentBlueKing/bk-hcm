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

package aideductconfig

import (
	"encoding/json"
	"fmt"

	actbill "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// GetAIDeductConfig 获取AI账单抵扣配置
func (s *aiDeductConfigSvc) GetAIDeductConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 权限校验
	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	// 获取配置key
	configKey, err := getConfigKeyByVendor(vendor)
	if err != nil {
		logs.Errorf("fail to get config key, err: %v, vendor: %s, rid: %s", err, vendor, cts.Kit.Rid)
		return nil, err
	}
	// 构建查询条件
	flt := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeBillAIDeduct),
			tools.RuleEqual("config_key", configKey),
		},
	}
	listReq := &datagconf.ListReq{
		Filter: flt,
		Page:   core.NewDefaultBasePage(),
	}
	// 查询配置
	resp, err := s.client.DataService().Global.GlobalConfig.List(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to get ai deduct config, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}

	// 如果配置不存在，返回空列表
	if len(resp.Details) == 0 {
		return &actbill.GetAIDeductConfigResp{MainAccountIDs: []string{}}, nil
	}

	// 解析配置值
	configValue := resp.Details[0].ConfigValue
	var accountIDs []string
	if err = json.Unmarshal([]byte(configValue), &accountIDs); err != nil {
		logs.Errorf("fail to parse ai deduct config, vendor: %s, config_value: %s, err: %v, rid: %s", vendor,
			string(configValue), err, cts.Kit.Rid)
		return nil, err
	}

	return &actbill.GetAIDeductConfigResp{MainAccountIDs: accountIDs}, nil
}

// getConfigKeyByVendor 根据云厂商获取对应的配置key
func getConfigKeyByVendor(vendor enumor.Vendor) (string, error) {
	switch vendor {
	case enumor.Gcp:
		return string(enumor.GlobalConfigKeyExcludedGcpMainAccountIDs), nil
	case enumor.Aws:
		return string(enumor.GlobalConfigKeyExcludedAwsMainAccountIDs), nil
	default:
		return "", fmt.Errorf("invalid vendor: %s", vendor)
	}
}
