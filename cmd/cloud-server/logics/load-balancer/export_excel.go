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
	"path/filepath"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
	"hcm/pkg/zip"
	"hcm/pkg/zip/excel"
)

// Exporter ...
type Exporter interface {
	PreCheck(kt *kit.Kit) error
	Export(kt *kit.Kit) (string, error)
}

// listenerExporter ...
type listenerExporter struct {
	client *client.ClientSet
	vendor enumor.Vendor
	params *cslb.ExportListenerReq
	path   string
}

// NewListenerExporter ...
func NewListenerExporter(client *client.ClientSet, vendor enumor.Vendor, params *cslb.ExportListenerReq) (Exporter,
	error) {

	return &listenerExporter{
		client: client,
		vendor: vendor,
		params: params,
		path:   cc.CloudServer().TmpFileDir,
	}, nil
}

// PreCheck ...
func (l *listenerExporter) PreCheck(kt *kit.Kit) error {
	// 1. 如果入参传入了监听器id，判断监听器是否在负载均衡下
	if err := l.checkClbListenerRel(kt); err != nil {
		logs.Errorf("check clb listener rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 判断监听器数量是否超过限制
	if err := l.checkListenerCount(kt); err != nil {
		logs.Errorf("check listener count failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 3. 判断规则数量是否超过限制
	if err := l.checkRuleCount(kt); err != nil {
		logs.Errorf("check rule count failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 4. 判断RS是否超过限制
	if err := l.checkRsCount(kt); err != nil {
		logs.Errorf("check rs count failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (l *listenerExporter) checkClbListenerRel(kt *kit.Kit) error {
	for _, one := range l.params.Listeners {
		if len(one.LblIDs) == 0 {
			continue
		}

		lblReq := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", one.LblIDs), tools.RuleEqual("lb_id", one.LbID)),
			Fields: []string{"id"},
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := l.client.DataService().Global.LoadBalancer.ListListener(kt, &lblReq)
		if err != nil {
			logs.Errorf("check clb listener rel failed, err: %v, req: %+v, rid: %s", err, lblReq, kt.Rid)
			return err
		}
		existIDMap := make(map[string]struct{})
		for _, detail := range resp.Details {
			existIDMap[detail.ID] = struct{}{}
		}

		for _, id := range one.LblIDs {
			if _, ok := existIDMap[id]; !ok {
				return errf.New(errf.InvalidParameter, fmt.Sprintf("listener: %s not belong to lb: %s", id, one.LbID))
			}
		}
	}

	return nil
}

func (l *listenerExporter) checkListenerCount(kt *kit.Kit) error {
	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()

	layer4Count, err := l.getListenerCount(kt, lbIDs, lblIDs, enumor.GetLayer4Protocol())
	if err != nil {
		return err
	}
	if layer4Count > constant.ExportLayer4ListenerLimit {
		return fmt.Errorf("导出的4层监听器数量为：%d, 超过限制：%d", layer4Count, constant.ExportLayer4ListenerLimit)
	}

	layer7Count, err := l.getListenerCount(kt, lbIDs, lblIDs, enumor.GetLayer7Protocol())
	if err != nil {
		return err
	}
	if layer7Count > constant.ExportLayer7ListenerLimit {
		return fmt.Errorf("导出的7层监听器数量为：%d, 超过限制：%d", layer7Count, constant.ExportLayer7ListenerLimit)
	}

	return nil
}

func (l *listenerExporter) getListenerCount(kt *kit.Kit, lbIDs []string, lblIDs []string,
	protocols []enumor.ProtocolType) (uint64, error) {

	if len(lbIDs) > int(core.DefaultMaxPageLimit) {
		return 0, fmt.Errorf("lb id length must less than %d", core.DefaultMaxPageLimit)
	}

	if len(lblIDs) > int(core.DefaultMaxPageLimit) {
		return 0, fmt.Errorf("lbl id length must less than %d", core.DefaultMaxPageLimit)
	}

	var count uint64

	if len(lbIDs) != 0 {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("lb_id", lbIDs), tools.RuleIn("protocol", protocols)),
			Page:   core.NewCountPage(),
		}
		resp, err := l.client.DataService().Global.LoadBalancer.ListListener(kt, &req)
		if err != nil {
			logs.Errorf("get listener count by lb id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return 0, err
		}
		count += resp.Count
	}

	if len(lblIDs) != 0 {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", lblIDs), tools.RuleIn("protocol", protocols)),
			Page:   core.NewCountPage(),
		}
		resp, err := l.client.DataService().Global.LoadBalancer.ListListener(kt, &req)
		if err != nil {
			logs.Errorf("get listener count by listener id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return 0, err
		}
		count += resp.Count
	}

	return count, nil
}

func (l *listenerExporter) checkRuleCount(kt *kit.Kit) error {
	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()
	var count uint64

	if len(lbIDs) != 0 {
		curCount, err := l.getRuleCount(kt, tools.RuleIn("lb_id", lbIDs))
		if err != nil {
			return err
		}
		count += curCount
	}

	if len(lblIDs) != 0 {
		curCount, err := l.getRuleCount(kt, tools.RuleIn("lbl_id", lblIDs))
		if err != nil {
			return err
		}
		count += curCount
	}

	if count > constant.ExportRuleLimit {
		return fmt.Errorf("导出规则数量超过限制，当前数量: %d, 限制数量: %d", count, constant.ExportRuleLimit)
	}

	return nil
}

func (l *listenerExporter) getRuleCount(kt *kit.Kit, rule *filter.AtomRule) (uint64, error) {
	req := core.ListReq{
		Filter: tools.ExpressionAnd(rule, tools.RuleEqual("rule_type", enumor.Layer7RuleType)),
		Page:   core.NewCountPage(),
	}
	switch l.vendor {
	case enumor.TCloud:
		resp, err := l.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &req)
		if err != nil {
			logs.Errorf("get listener rule count failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return 0, err
		}
		return resp.Count, nil
	default:
		return 0, errf.New(errf.InvalidParameter, "不支持该厂商的导出")
	}
}

func (l *listenerExporter) checkRsCount(kt *kit.Kit) error {
	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()

	layer4Count, err := l.getRsCount(kt, lbIDs, lblIDs, enumor.Layer4RuleType)
	if err != nil {
		logs.Errorf("get layer4 rs count failed, err: %v, lbIDs: %+v, lblIDs: %+v, rid: %s", err, lbIDs, lblIDs, kt.Rid)
		return err
	}
	if layer4Count > constant.ExportLayer4RsLimit {
		return fmt.Errorf("导出4层RS数量超过限制，当前数量: %d, 限制数量: %d", layer4Count, constant.ExportLayer4RsLimit)
	}

	layer7Count, err := l.getRsCount(kt, lbIDs, lblIDs, enumor.Layer7RuleType)
	if err != nil {
		logs.Errorf("get layer7 rs count failed, err: %v, lbIDs: %+v, lblIDs: %+v, rid: %s", err, lbIDs, lblIDs, kt.Rid)
		return err
	}
	if layer7Count > constant.ExportLayer7RsLimit {
		return fmt.Errorf("导出7层RS数量超过限制，当前数量: %d, 限制数量: %d", layer7Count, constant.ExportLayer7RsLimit)
	}

	return nil
}

func (l *listenerExporter) getRsCount(kt *kit.Kit, lbIDs []string, lblIDs []string, ruleType enumor.RuleType) (uint64,
	error) {

	if len(lbIDs) > int(core.DefaultMaxPageLimit) {
		return 0, fmt.Errorf("lb id length must less than %d", core.DefaultMaxPageLimit)
	}

	if len(lblIDs) > int(core.DefaultMaxPageLimit) {
		return 0, fmt.Errorf("lbl id length must less than %d", core.DefaultMaxPageLimit)
	}

	rules := make([]filter.RuleFactory, 0)
	rules = append(rules, tools.RuleEqual("listener_rule_type", ruleType))

	if len(lbIDs) != 0 && len(lblIDs) != 0 {
		rules = append(rules, tools.ExpressionOr(tools.RuleIn("lb_id", lbIDs), tools.RuleIn("lbl_id", lblIDs)))
	}
	if len(lbIDs) != 0 {
		rules = append(rules, tools.ExpressionAnd(tools.RuleIn("lb_id", lbIDs)))
	}
	if len(lblIDs) != 0 {
		rules = append(rules, tools.ExpressionAnd(tools.RuleIn("lbl_id", lblIDs)))
	}

	relReq := core.ListReq{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: core.NewDefaultBasePage(),
	}
	targetGroupIDMap := make(map[string]struct{})
	for {
		relResp, err := l.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &relReq)
		if err != nil {
			logs.Errorf("get target group listener rel failed, err: %v, req: %+v, rid: %s", err, relReq, kt.Rid)
			return 0, err
		}
		for _, detail := range relResp.Details {
			targetGroupIDMap[detail.TargetGroupID] = struct{}{}
		}
		if len(relResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		relReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	targetGroupIDs := maps.Keys(targetGroupIDMap)
	var count uint64
	for _, batch := range slice.Split(targetGroupIDs, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("target_group_id", batch)),
			Page:   core.NewCountPage(),
		}
		resp, err := l.client.DataService().Global.LoadBalancer.ListTarget(kt, &req)
		if err != nil {
			logs.Errorf("get target count by target group id failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return 0, err
		}
		count += resp.Count
	}

	return count, nil
}

// Export ...
func (l *listenerExporter) Export(kt *kit.Kit) (string, error) {
	fileName := zip.GenFileName(constant.CLBFilePrefix)
	zipOperator, err := excel.NewOperator(l.path, fileName)
	if err != nil {
		logs.Errorf("create zip operator failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	defer zipOperator.Close()

	switch l.vendor {
	case enumor.TCloud:
		err = l.exportTCloud(kt, zipOperator)
	default:
		return "", errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", l.vendor)
	}
	if err != nil {
		logs.Errorf("export file failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return "", err
	}

	if err = zipOperator.Save(); err != nil {
		logs.Errorf("save zip file failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return filepath.Join(l.path, fileName), nil
}

func (l *listenerExporter) getLbs(kt *kit.Kit) (map[string]corelb.BaseLoadBalancer, error) {
	lbIDs := l.params.GetAllLbIDs()
	result := make(map[string]corelb.BaseLoadBalancer)

	for _, split := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", split)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "cloud_id", "private_ipv4_addresses", "private_ipv6_addresses",
				"public_ipv4_addresses", "public_ipv6_addresses", "domain"},
		}
		resp, err := l.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, &req)
		if err != nil {
			logs.Errorf("list load balancer failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = detail
		}
	}

	return result, nil
}

func (l *listenerExporter) getTgLblRelClassifyProtocol(kt *kit.Kit) ([]corelb.BaseTargetListenerRuleRel,
	[]corelb.BaseTargetListenerRuleRel, error) {

	layer4TgLblRel := make([]corelb.BaseTargetListenerRuleRel, 0)
	layer7TgLblRel := make([]corelb.BaseTargetListenerRuleRel, 0)

	relReq := core.ListReq{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: l.getTgLblRelRule(),
		},
		Page: core.NewDefaultBasePage(),
	}
	for {
		relResp, err := l.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &relReq)
		if err != nil {
			logs.Errorf("get target group listener rel failed, err: %v, req: %+v, rid: %s", err, relReq, kt.Rid)
			return nil, nil, err
		}
		for _, detail := range relResp.Details {
			switch detail.ListenerRuleType {
			case enumor.Layer4RuleType:
				layer4TgLblRel = append(layer4TgLblRel, detail)
			case enumor.Layer7RuleType:
				layer7TgLblRel = append(layer7TgLblRel, detail)
			default:
				return nil, nil, fmt.Errorf("invalid listener rule type: %s", detail.ListenerRuleType)
			}
		}
		if len(relResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		relReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return layer4TgLblRel, layer7TgLblRel, nil
}

func (l *listenerExporter) getTgLblRelRule() []filter.RuleFactory {
	lbIDs, lblIDs := l.params.GetPartLbAndLblIDs()
	rules := make([]filter.RuleFactory, 0)

	if len(lbIDs) != 0 && len(lblIDs) != 0 {
		rules = append(rules, tools.ExpressionOr(tools.RuleIn("lb_id", lbIDs), tools.RuleIn("lbl_id", lblIDs)))
		return rules
	}

	if len(lbIDs) != 0 {
		rules = append(rules, tools.ExpressionAnd(tools.RuleIn("lb_id", lbIDs)))
		return rules
	}

	if len(lblIDs) != 0 {
		rules = append(rules, tools.ExpressionAnd(tools.RuleIn("lbl_id", lblIDs)))
		return rules
	}
	return rules
}

func (l *listenerExporter) getRsClassifyProtocol(kt *kit.Kit, layer4TgLblRel,
	layer7TgLblRel []corelb.BaseTargetListenerRuleRel) ([]corelb.BaseTarget, []corelb.BaseTarget, error) {

	layer4TgIDMap := make(map[string]struct{})
	for _, rel := range layer4TgLblRel {
		layer4TgIDMap[rel.TargetGroupID] = struct{}{}
	}
	layer7TgIDMap := make(map[string]struct{})
	for _, rel := range layer7TgLblRel {
		layer7TgIDMap[rel.TargetGroupID] = struct{}{}
	}

	tgIDs := make([]string, 0)
	if len(layer4TgIDMap) != 0 {
		tgIDs = append(tgIDs, maps.Keys(layer4TgIDMap)...)
	}
	if len(layer7TgIDMap) != 0 {
		tgIDs = append(tgIDs, maps.Keys(layer7TgIDMap)...)
	}

	layer4Rs := make([]corelb.BaseTarget, 0)
	layer7Rs := make([]corelb.BaseTarget, 0)
	if len(tgIDs) == 0 {
		return layer4Rs, layer7Rs, nil
	}

	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					tools.RuleIn("target_group_id", batch),
				},
			},
			Page: core.NewDefaultBasePage(),
		}
		for {
			resp, err := l.client.DataService().Global.LoadBalancer.ListTarget(kt, &req)
			if err != nil {
				logs.Errorf("get rs failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, nil, err
			}
			for _, detail := range resp.Details {
				tgID := detail.TargetGroupID
				if _, ok := layer4TgIDMap[tgID]; ok {
					layer4Rs = append(layer4Rs, detail)
					continue
				}
				if _, ok := layer7TgIDMap[tgID]; ok {
					layer7Rs = append(layer7Rs, detail)
				}
			}
			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			req.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return layer4Rs, layer7Rs, nil
}

func (l *listenerExporter) writerLayer4Listeners(kt *kit.Kit, zipOperator zip.OperatorI,
	clbListenerMap map[string][]Layer4ListenerDetail) error {

	firstRow, err := getFirstRow(l.vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return err
	}

	for clbFlag, listeners := range clbListenerMap {
		for i, batch := range slice.Split(listeners, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, layer4ListenerHeaders...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", constant.Layer4ListenerFilePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, constant.Layer4ListenerSheetName)
			if err := zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (l *listenerExporter) writerLayer7Listeners(kt *kit.Kit, zipOperator zip.OperatorI,
	clbListenerMap map[string][]Layer7ListenerDetail) error {

	firstRow, err := getFirstRow(l.vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return err
	}

	for clbFlag, listeners := range clbListenerMap {
		for i, batch := range slice.Split(listeners, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, layer7ListenerHeaders...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", constant.Layer7ListenerFilePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, constant.Layer7ListenerSheetName)
			if err := zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (l *listenerExporter) writerRules(kt *kit.Kit, zipOperator zip.OperatorI,
	clbRuleMap map[string][]RuleDetail) error {

	firstRow, err := getFirstRow(l.vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return err
	}

	for clbFlag, rules := range clbRuleMap {
		for i, batch := range slice.Split(rules, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, ruleHeaders...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", constant.RuleFilePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, constant.RuleSheetName)
			if err := zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (l *listenerExporter) writerLayer4Rs(kt *kit.Kit, zipOperator zip.OperatorI,
	clbRsMap map[string][]Layer4RsDetail) error {

	firstRow, err := getFirstRow(l.vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return err
	}

	for clbFlag, rs := range clbRsMap {
		for i, batch := range slice.Split(rs, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, layer4RsHeaders...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", constant.Layer4RsFilePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, constant.Layer4RsSheetName)
			if err := zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}

func (l *listenerExporter) writerLayer7Rs(kt *kit.Kit, zipOperator zip.OperatorI,
	clbRsMap map[string][]Layer7RsDetail) error {

	firstRow, err := getFirstRow(l.vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, l.vendor, kt.Rid)
		return err
	}

	for clbFlag, rs := range clbRsMap {
		for i, batch := range slice.Split(rs, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, layer7RsHeaders...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", constant.Layer7RsFilePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, constant.Layer7RsSheetName)
			if err := zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
