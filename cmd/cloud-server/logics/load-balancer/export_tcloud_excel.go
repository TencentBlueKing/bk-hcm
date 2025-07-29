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

	"hcm/pkg/api/core"
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/zip"
)

func (l *listenerExporter) exportTCloud(kt *kit.Kit, zipOperator zip.OperatorI) error {
	lbMap, err := l.getLbs(kt)
	if err != nil {
		logs.Errorf("get lbs failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	layer4ListenerMap, layer7ListenerMap, err := l.getTCloudListeners(kt)
	if err != nil {
		logs.Errorf("get tcloud listeners failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	layer4RuleMap, layer7RuleMap, err := l.getTCloudRules(kt)
	if err != nil {
		logs.Errorf("get tcloud rules failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err = l.writeTCloudLayer4Listener(kt, zipOperator, lbMap, layer4ListenerMap, layer4RuleMap); err != nil {
		logs.Errorf("build tcloud layer4 listener excel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if err = l.writeTCloudLayer7Listener(kt, zipOperator, lbMap, layer7ListenerMap); err != nil {
		logs.Errorf("build tcloud layer7 listener excel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if err = l.writeTCloudRule(kt, zipOperator, lbMap, layer7ListenerMap, layer7RuleMap); err != nil {
		logs.Errorf("build tcloud rule excel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	layer4TgLblRel, layer7TgLblRel, err := l.getTgLblRelClassifyProtocol(kt)
	if err != nil {
		logs.Errorf("get tcloud target group listener rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	layer4Rs, layer7Rs, err := l.getRsClassifyProtocol(kt, layer4TgLblRel, layer7TgLblRel)
	if err != nil {
		logs.Errorf("get tcloud rs failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err = l.writeTCloudLayer4Rs(kt, zipOperator, lbMap, layer4ListenerMap, layer4TgLblRel, layer4Rs); err != nil {
		logs.Errorf("build tcloud layer4 rs excel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	err = l.writeTCloudLayer7Rs(kt, zipOperator, lbMap, layer7ListenerMap, layer7RuleMap, layer7TgLblRel, layer7Rs)
	if err != nil {
		logs.Errorf("build tcloud layer7 rs excel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (l *listenerExporter) getTCloudListeners(kt *kit.Kit) (map[string]loadbalancer.TCloudListener,
	map[string]loadbalancer.TCloudListener, error) {

	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()

	layer4ListenerMap, err := l.getTCloudListenersByProtocol(kt, lbIDs, lblIDs, enumor.GetLayer4Protocol())
	if err != nil {
		logs.Errorf("get tcloud layer4 listener failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	layer7ListenerMap, err := l.getTCloudListenersByProtocol(kt, lbIDs, lblIDs, enumor.GetLayer7Protocol())
	if err != nil {
		logs.Errorf("get tcloud layer7 listener failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	return layer4ListenerMap, layer7ListenerMap, nil
}

func (l *listenerExporter) getTCloudListenersByProtocol(kt *kit.Kit, lbIDs []string, lblIDs []string,
	protocols []enumor.ProtocolType) (map[string]loadbalancer.TCloudListener, error) {

	if len(lbIDs) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("lb id length must less than %d", core.DefaultMaxPageLimit)
	}

	if len(lblIDs) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("lbl id length must less than %d", core.DefaultMaxPageLimit)
	}

	result := make(map[string]loadbalancer.TCloudListener)

	if len(lbIDs) != 0 {
		for {
			req := core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleIn("lb_id", lbIDs), tools.RuleIn("protocol", protocols)),
				Page:   core.NewDefaultBasePage(),
			}
			resp, err := l.client.DataService().TCloud.LoadBalancer.ListListener(kt, &req)
			if err != nil {
				logs.Errorf("get listener by lb id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, err
			}
			for _, detail := range resp.Details {
				result[detail.ID] = detail
			}

			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}

			req.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	if len(lblIDs) != 0 {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", lblIDs), tools.RuleIn("protocol", protocols)),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := l.client.DataService().TCloud.LoadBalancer.ListListener(kt, &req)
		if err != nil {
			logs.Errorf("get listener by listener id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = detail
		}
	}

	return result, nil
}

func (l *listenerExporter) writeTCloudLayer4Listener(kt *kit.Kit, zipOperator zip.OperatorI,
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
			logs.Errorf("can not get tcloud layer4 rule by listener id, listener id: %s, rid: %s", listener.ID,
				kt.Rid)
			return fmt.Errorf("can not get tcloud layer4 rule by listener id, listener id: %s", listener.ID)
		}

		healthCheck := enumor.DisableListenerHealthCheck
		if layer4Rule.HealthCheck != nil {
			healthCheck = enumor.EnableListenerHealthCheck
		}
		listenerPortStr := getListenerPortStr(listener)
		clbVipDomain := getLbVipOrDomain(lb)

		clbListenerMap[clbVipDomain] = append(clbListenerMap[clbVipDomain], Layer4ListenerDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			Scheduler:       enumor.Scheduler(layer4Rule.Scheduler),
			Session:         int(layer4Rule.SessionExpire),
			HealthCheckStr:  healthCheck,
		})
	}

	if err := l.writerLayer4Listeners(kt, zipOperator, clbListenerMap); err != nil {
		return err
	}

	return nil
}

func (l *listenerExporter) getTCloudRules(kt *kit.Kit) (map[string]loadbalancer.TCloudLbUrlRule,
	map[string]loadbalancer.TCloudLbUrlRule, error) {

	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()

	layer4RuleMap, err := l.getTCloudRulesByRuleType(kt, lbIDs, lblIDs, enumor.Layer4RuleType)
	if err != nil {
		logs.Errorf("get tcloud layer4 rules failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	layer7RuleMap, err := l.getTCloudRulesByRuleType(kt, lbIDs, lblIDs, enumor.Layer7RuleType)
	if err != nil {
		logs.Errorf("get tcloud layer7 rules failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	return layer4RuleMap, layer7RuleMap, nil
}

func (l *listenerExporter) getTCloudRulesByRuleType(kt *kit.Kit, lbIDs []string, lblIDs []string,
	ruleType enumor.RuleType) (map[string]loadbalancer.TCloudLbUrlRule, error) {

	if len(lbIDs) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("lb id length must less than %d", core.DefaultMaxPageLimit)
	}

	if len(lblIDs) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("lbl id length must less than %d", core.DefaultMaxPageLimit)
	}

	result := make(map[string]loadbalancer.TCloudLbUrlRule)

	if len(lbIDs) != 0 {
		for {
			req := core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleIn("lb_id", lbIDs), tools.RuleEqual("rule_type", ruleType)),
				Page:   core.NewDefaultBasePage(),
			}
			resp, err := l.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &req)
			if err != nil {
				logs.Errorf("get rule by lb id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, err
			}
			for _, detail := range resp.Details {
				result[detail.ID] = detail
			}

			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}

			req.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	if len(lblIDs) != 0 {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("lbl_id", lblIDs), tools.RuleEqual("rule_type", ruleType)),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := l.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &req)
		if err != nil {
			logs.Errorf("get rule by listener id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = detail
		}
	}

	return result, nil
}

func (l *listenerExporter) writeTCloudLayer7Listener(kt *kit.Kit, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer7ListenerMap map[string]loadbalancer.TCloudListener) error {

	if len(layer7ListenerMap) == 0 {
		return nil
	}

	clbListenerMap := make(map[string][]Layer7ListenerDetail)
	for _, listener := range layer7ListenerMap {
		lbID := listener.LbID
		lb, ok := lbMap[lbID]
		if !ok {
			logs.Errorf("can not get clb by lb id failed, lb id: %s, rid: %s", lbID, kt.Rid)
			return fmt.Errorf("can not get clb by lb id failed, lb id: %s", lbID)
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
		clbVipDomain := getLbVipOrDomain(lb)

		clbListenerMap[clbVipDomain] = append(clbListenerMap[clbVipDomain], Layer7ListenerDetail{
			ClbVipDomain:    clbVipDomain,
			CloudClbID:      lb.CloudID,
			Protocol:        listener.Protocol,
			ListenerPortStr: listenerPortStr,
			SSLMode:         sslMode,
			CertCloudID:     certCloudID,
			CACloudID:       caCloudID,
		})
	}

	if err := l.writerLayer7Listeners(kt, zipOperator, clbListenerMap); err != nil {
		return err
	}

	return nil
}

func (l *listenerExporter) writeTCloudRule(kt *kit.Kit, zipOperator zip.OperatorI,
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
			logs.Errorf("can not get listener by listener id, listener id: %s, rid: %s", listener.ID, kt.Rid)
			return fmt.Errorf("can not get listener by listener id, listener id: %s", listener.ID)
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
		clbVipDomain := getLbVipOrDomain(lb)

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

	if err := l.writerRules(kt, zipOperator, clbRuleMap); err != nil {
		return err
	}

	return nil
}

func (l *listenerExporter) writeTCloudLayer4Rs(kt *kit.Kit, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer4ListenerMap map[string]loadbalancer.TCloudListener,
	layer4TgLblRel []loadbalancer.BaseTargetListenerRuleRel, layer4Rs []loadbalancer.BaseTarget) error {

	if len(layer4Rs) == 0 {
		return nil
	}

	tgIDLblIDMap := make(map[string]string)
	for _, tgLblRel := range layer4TgLblRel {
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
		listener, ok := layer4ListenerMap[lblID]
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
		clbVipDomain := getLbVipOrDomain(lb)

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

	if err := l.writerLayer4Rs(kt, zipOperator, clbRsMap); err != nil {
		return err
	}

	return nil
}

func (l *listenerExporter) writeTCloudLayer7Rs(kt *kit.Kit, zipOperator zip.OperatorI,
	lbMap map[string]loadbalancer.BaseLoadBalancer, layer7ListenerMap map[string]loadbalancer.TCloudListener,
	layer7RuleMap map[string]loadbalancer.TCloudLbUrlRule, layer7TgLblRel []loadbalancer.BaseTargetListenerRuleRel,
	layer7Rs []loadbalancer.BaseTarget) error {

	if len(layer7Rs) == 0 {
		return nil
	}

	tgIDLblIDMap := make(map[string]string)
	tgIDRuleIDMap := make(map[string]string)
	for _, tgLblRel := range layer7TgLblRel {
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
		listener, ok := layer7ListenerMap[lblID]
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
		rule, ok := layer7RuleMap[ruleID]
		if !ok {
			logs.Errorf("can not get rule by rule id, rule id: %s, rid: %s", ruleID, kt.Rid)
			return fmt.Errorf("can not get rule by rule id, rule id: %s", ruleID)
		}
		listenerPortStr := getListenerPortStr(listener)
		rsPortStr := getRsPortStr(listener, rs)
		clbVipDomain := getLbVipOrDomain(lb)

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

	if err := l.writerLayer7Rs(kt, zipOperator, clbRsMap); err != nil {
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
