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

package lblogic

import (
	"fmt"
	"strings"

	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/zip"
)

func writeTCloudLayer4Listener(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer4ListenerMap map[string]loadbalancer.TCloudListener,
	layer4RuleMap map[string]loadbalancer.TCloudLbUrlRule) error {

	if len(layer4ListenerMap) == 0 {
		return nil
	}

	// 四层规则对于监听器来说是一对一的
	lblIDLayer4RuleMap := make(map[string]loadbalancer.TCloudLbUrlRule)
	for _, layer4Rule := range maps.Values(layer4RuleMap) {
		lblIDLayer4RuleMap[layer4Rule.LblID] = layer4Rule
	}

	clbListenerMap := make(map[string][]Layer4ListenerDetail)
	for _, listener := range layer4ListenerMap {
		lbID := listener.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id, lb id: %s", lbID)
		}

		layer4Rule, ok := lblIDLayer4RuleMap[listener.ID]
		if !ok {
			logs.Errorf("can not get layer4 rule by listener id, vendor: %s, listener id: %s, rid: %s", vendor,
				listener.ID, kt.Rid)
			return fmt.Errorf("can not get layer4 rule by listener id, vendor: %s, listener id: %s", vendor,
				listener.ID)
		}

		healthCheck := enumor.DisableListenerHealthCheck
		if layer4Rule.HealthCheck != nil {
			healthCheck = enumor.EnableListenerHealthCheck
		}
		listenerPortStr := getListenerPortStr(listener)
		clbVipDomain, err := getLbVipOrDomain(lb)
		if err != nil {
			logs.Errorf("get clb vip or domain failed, err: %v, lb id: %s, rid: %s", err, lbID, kt.Rid)
			return err
		}

		clbListenerMap[clbVipDomain] = append(clbListenerMap[clbVipDomain], Layer4ListenerDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			Scheduler:       enumor.Scheduler(layer4Rule.Scheduler),
			Session:         int(layer4Rule.SessionExpire),
			HealthCheckStr:  healthCheck,
			Name:            listener.Name,
		})
	}

	if err := writeLayer4Listeners(kt, vendor, zipOperator, clbListenerMap); err != nil {
		return err
	}

	return nil
}

func writeTCloudLayer7Listener(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer7ListenerMap map[string]loadbalancer.TCloudListener) error {

	if len(layer7ListenerMap) == 0 {
		return nil
	}

	clbListenerMap := make(map[string][]Layer7ListenerDetail)
	for _, listener := range layer7ListenerMap {
		lbID := listener.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id, lb id: %s", lbID)
		}
		listenerPortStr := getListenerPortStr(listener)
		var sslMode, certCloudID, caCloudID string
		if listener.Extension != nil && listener.Extension.Certificate != nil {
			sslMode = converter.PtrToVal(listener.Extension.Certificate.SSLMode)
			if len(listener.Extension.Certificate.CertCloudIDs) == 1 {
				certCloudID = listener.Extension.Certificate.CertCloudIDs[0]
			}
			if len(listener.Extension.Certificate.CertCloudIDs) > 1 {
				certCloudID = fmt.Sprintf("[%s]", strings.Join(listener.Extension.Certificate.CertCloudIDs, ","))
			}

			caCloudID = converter.PtrToVal(listener.Extension.Certificate.CaCloudID)
		}
		clbVipDomain, err := getLbVipOrDomain(lb)
		if err != nil {
			logs.Errorf("get clb vip or domain failed, err: %v, lb id: %s, rid: %s", err, lbID, kt.Rid)
			return err
		}

		clbListenerMap[clbVipDomain] = append(clbListenerMap[clbVipDomain], Layer7ListenerDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			SSLMode:         sslMode,
			CertCloudID:     certCloudID,
			CACloudID:       caCloudID,
			Name:            listener.Name,
		})
	}

	if err := writeLayer7Listeners(kt, vendor, zipOperator, clbListenerMap); err != nil {
		return err
	}

	return nil
}

func writeTCloudRule(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer7ListenerMap map[string]loadbalancer.TCloudListener,
	layer7RuleMap map[string]loadbalancer.TCloudLbUrlRule) error {

	if len(layer7RuleMap) == 0 {
		return nil
	}

	clbRuleMap := make(map[string][]RuleDetail)
	for _, rule := range layer7RuleMap {
		lbID := rule.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id, lb id: %s", lbID)
		}

		listenerID := rule.LblID
		listener, ok := layer7ListenerMap[listenerID]
		if !ok {
			logs.Errorf("can not get listener by listener id, listener id: %s, rid: %s", listenerID, kt.Rid)
			return fmt.Errorf("can not get listener by listener id, listener id: %s", listenerID)
		}

		healthCheck := enumor.DisableListenerHealthCheck
		if rule.HealthCheck != nil {
			healthCheck = enumor.EnableListenerHealthCheck
		}
		listenerPortStr := getListenerPortStr(listener)
		isDefaultDomain := false
		if rule.Domain == listener.DefaultDomain {
			isDefaultDomain = true
		}
		clbVipDomain, err := getLbVipOrDomain(lb)
		if err != nil {
			logs.Errorf("get clb vip or domain failed, err: %v, lb id: %s, rid: %s", err, lbID, kt.Rid)
			return err
		}

		clbRuleMap[clbVipDomain] = append(clbRuleMap[clbVipDomain], RuleDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			Domain:          rule.Domain,
			DefaultDomain:   isDefaultDomain,
			UrlPath:         rule.URL,
			Scheduler:       enumor.Scheduler(rule.Scheduler),
			Session:         int(rule.SessionExpire),
			HealthCheckStr:  healthCheck,
		})
	}

	if err := writeRules(kt, vendor, zipOperator, clbRuleMap); err != nil {
		return err
	}

	return nil
}

func writeTCloudLayer4Rs(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, listenerMap map[string]loadbalancer.TCloudListener,
	tgLblRel []loadbalancer.BaseTargetListenerRuleRel, layer4Rs []loadbalancer.BaseTarget) error {

	if len(layer4Rs) == 0 {
		return nil
	}

	tgIDLblIDMap := make(map[string]string)
	for _, tgLblRel := range tgLblRel {
		tgIDLblIDMap[tgLblRel.TargetGroupID] = tgLblRel.LblID
	}

	clbRsMap := make(map[string][]Layer4RsDetail)
	for _, rs := range layer4Rs {
		tgID := rs.TargetGroupID
		lblID, ok := tgIDLblIDMap[tgID]
		if !ok {
			logs.Errorf("can not get lbl by tg id, tg id: %s, rid: %s", tgID, kt.Rid)
			return fmt.Errorf("can not get lbl by tg id, tg id: %s", tgID)
		}
		listener, ok := listenerMap[lblID]
		if !ok {
			logs.Errorf("can not get listener by lbl id, lbl id: %s, rid: %s", lblID, kt.Rid)
			return fmt.Errorf("can not get listener by lbl id, lbl id: %s", lblID)
		}
		lbID := listener.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id, lb id: %s", lbID)
		}
		listenerPortStr := getListenerPortStr(listener)
		rsPortStr := getRsPortStr(listener, rs)
		clbVipDomain, err := getLbVipOrDomain(lb)
		if err != nil {
			logs.Errorf("get clb vip or domain failed, err: %v, lb id: %s, rid: %s", err, lbID, kt.Rid)
			return err
		}

		clbRsMap[clbVipDomain] = append(clbRsMap[clbVipDomain], Layer4RsDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			InstType:        rs.InstType,
			RsIp:            rs.IP,
			RsPortStr:       rsPortStr,
			Weight:          rs.Weight,
		})
	}

	if err := writeLayer4Rs(kt, vendor, zipOperator, clbRsMap); err != nil {
		return err
	}

	return nil
}

func writeTCloudLayer7Rs(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, listenerMap map[string]loadbalancer.TCloudListener,
	ruleMap map[string]loadbalancer.TCloudLbUrlRule, tgLblRel []loadbalancer.BaseTargetListenerRuleRel,
	layer7Rs []loadbalancer.BaseTarget) error {

	if len(layer7Rs) == 0 {
		return nil
	}

	tgIDLblIDMap := make(map[string]string)
	tgIDRuleIDMap := make(map[string]string)
	for _, tgLblRel := range tgLblRel {
		tgIDLblIDMap[tgLblRel.TargetGroupID] = tgLblRel.LblID
		tgIDRuleIDMap[tgLblRel.TargetGroupID] = tgLblRel.ListenerRuleID
	}

	clbRsMap := make(map[string][]Layer7RsDetail)
	for _, rs := range layer7Rs {
		tgID := rs.TargetGroupID
		lblID, ok := tgIDLblIDMap[tgID]
		if !ok {
			logs.Errorf("can not get lbl by tg id, tg id: %s, rid: %s", tgID, kt.Rid)
			return fmt.Errorf("can not get lbl by tg id, tg id: %s", tgID)
		}
		listener, ok := listenerMap[lblID]
		if !ok {
			logs.Errorf("can not get listener by lbl id, lbl id: %s, rid: %s", lblID, kt.Rid)
			return fmt.Errorf("can not get listener by lbl id, lbl id: %s", lblID)
		}
		lbID := listener.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id, lb id: %s", lbID)
		}
		ruleID, ok := tgIDRuleIDMap[tgID]
		if !ok {
			logs.Errorf("can not get rule id by tg id, tg id: %s, rid: %s", tgID, kt.Rid)
			return fmt.Errorf("can not get rule id by tg id, tg id: %s", tgID)
		}
		rule, ok := ruleMap[ruleID]
		if !ok {
			logs.Errorf("can not get rule by rule id, rule id: %s, rid: %s", ruleID, kt.Rid)
			return fmt.Errorf("can not get rule by rule id, rule id: %s", ruleID)
		}
		listenerPortStr := getListenerPortStr(listener)
		rsPortStr := getRsPortStr(listener, rs)
		clbVipDomain, err := getLbVipOrDomain(lb)
		if err != nil {
			logs.Errorf("get clb vip or domain failed, err: %v, lb id: %s, rid: %s", err, lbID, kt.Rid)
			return err
		}

		clbRsMap[clbVipDomain] = append(clbRsMap[clbVipDomain], Layer7RsDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			Domain:          rule.Domain,
			URLPath:         rule.URL,
			InstType:        rs.InstType,
			RsIp:            rs.IP,
			RsPortStr:       rsPortStr,
			Weight:          rs.Weight,
		})
	}

	if err := writeLayer7Rs(kt, vendor, zipOperator, clbRsMap); err != nil {
		return err
	}

	return nil
}

func getListenerPortStr(listener loadbalancer.TCloudListener) string {
	listenerPortStr := fmt.Sprintf("%d", listener.Port)
	if listener.Extension != nil && converter.PtrToVal(listener.Extension.EndPort) != 0 {
		listenerPortStr = fmt.Sprintf("[%d,%d]", listener.Port, converter.PtrToVal(listener.Extension.EndPort))
	}
	return listenerPortStr
}

func getRsPortStr(listener loadbalancer.TCloudListener, rs loadbalancer.BaseTarget) string {
	rsPortStr := fmt.Sprintf("%d", rs.Port)
	if listener.Extension != nil && converter.PtrToVal(listener.Extension.EndPort) != 0 {
		lblEndPort := converter.PtrToVal(listener.Extension.EndPort)
		rsPortStr = fmt.Sprintf("[%d,%d]", rs.Port, lblEndPort-listener.Port+rs.Port)
	}
	return rsPortStr
}
