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

package monthtask

import (
	"encoding/json"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// GenAIFilterRules generate ai filter rules
func GenAIFilterRules(kt *kit.Kit, vendor enumor.Vendor) (rules []filter.RuleFactory, err error) {
	rules = []filter.RuleFactory{tools.RuleStartsWith("hc_product_name", constant.BillItemAIPrefix)}

	// 获取排除的主账号列表
	excludedAccountIDs, err := getExcludedMainAccountIDs(kt, vendor)
	if err != nil {
		logs.Errorf("fail to get excluded main account ids, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
		return nil, err
	}
	if len(excludedAccountIDs) > 0 {
		rules = append(rules, tools.RuleNotIn("main_account_id", excludedAccountIDs))
	}

	return rules, nil
}

// getExcludedMainAccountIDs 从 global_config 获取排除的主账号列表
func getExcludedMainAccountIDs(kt *kit.Kit, vendor enumor.Vendor) ([]string, error) {
	configKey := getConfigKeyByVendor(vendor)
	if len(configKey) == 0 {
		// 不支持的云厂商，返回空列表
		return nil, nil
	}

	// 构建查询条件
	flt := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeBillAIDeduct),
			tools.RuleEqual("config_key", configKey),
		},
	}

	req := &datagconf.ListReq{
		Filter: flt,
		Page:   core.NewDefaultBasePage(),
	}

	// 查询配置
	resp, err := actcli.GetDataService().Global.GlobalConfig.List(kt, req)
	if err != nil {
		logs.Errorf("fail to get excluded main account ids from global config, vendor: %s, err: %v, rid: %s",
			vendor, err, kt.Rid)
		return nil, err
	}

	if len(resp.Details) == 0 {
		// 配置不存在，返回空列表
		return nil, nil
	}

	// 解析配置值
	configValue := resp.Details[0].ConfigValue
	var accountIDs []string
	if err = json.Unmarshal([]byte(configValue), &accountIDs); err != nil {
		logs.Warnf("fail to parse excluded main account ids config, vendor: %s, config_value: %s, err: %v, rid: %s",
			vendor, string(configValue), err, kt.Rid)
		return nil, err
	}

	return accountIDs, nil
}

// getConfigKeyByVendor 根据云厂商获取对应的配置key
func getConfigKeyByVendor(vendor enumor.Vendor) string {
	switch vendor {
	case enumor.Gcp:
		return string(enumor.GlobalConfigKeyExcludedGcpMainAccountIDs)
	case enumor.Aws:
		return string(enumor.GlobalConfigKeyExcludedAwsMainAccountIDs)
	default:
		return ""
	}
}
