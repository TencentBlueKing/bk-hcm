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
	"hcm/pkg/tools/slice"
	"hcm/pkg/zip"
	"hcm/pkg/zip/excel"
)

// targetExporter ...
type targetExporter struct {
	client *client.ClientSet
	vendor enumor.Vendor
	params *cslb.ExportTargetReq
	path   string
}

// NewTargetExporter ...
func NewTargetExporter(client *client.ClientSet, vendor enumor.Vendor, params *cslb.ExportTargetReq) (Exporter, error) {
	return &targetExporter{
		client: client,
		vendor: vendor,
		params: params,
		path:   cc.CloudServer().TmpFileDir,
	}, nil
}

// PreCheck ...
func (t *targetExporter) PreCheck(kt *kit.Kit) error {
	return t.params.Validate()
}

// Export ...
func (t *targetExporter) Export(kt *kit.Kit) (string, error) {
	fileName := zip.GenFileName(constant.CLBFilePrefix)
	zipOperator, err := excel.NewOperator(t.path, fileName)
	if err != nil {
		logs.Errorf("create zip operator failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	defer func(zipOperator zip.OperatorI) {
		if err = zipOperator.Close(); err != nil {
			logs.Errorf("close zip operator failed, err: %v, rid: %s", err, kt.Rid)
		}
	}(zipOperator)

	switch t.vendor {
	case enumor.TCloud:
		err = t.exportTCloud(kt, zipOperator)
	default:
		return "", errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", t.vendor)
	}
	if err != nil {
		logs.Errorf("export file failed, err: %v, vendor: %s, rid: %s", err, t.vendor, kt.Rid)
		return "", err
	}

	if err = zipOperator.Save(); err != nil {
		logs.Errorf("save zip file failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return filepath.Join(t.path, fileName), nil
}

func (t *targetExporter) exportTCloud(kt *kit.Kit, zipOperator zip.OperatorI) error {
	// 查询RS信息
	targets, err := t.getTargetByIDs(kt, t.params.TargetIDs)
	if err != nil {
		logs.Errorf("get target by ids failed, err: %v, ids: %v, rid: %s", err, t.params.TargetIDs, kt.Rid)
		return err
	}

	// 查询目标组与监听器关系
	tgIDs := make([]string, 0)
	for _, target := range targets {
		tgIDs = append(tgIDs, target.TargetGroupID)
	}
	tgLblRel, err := t.getTgLblRelByTgIDs(kt, tgIDs)
	if err != nil {
		logs.Errorf("get tg lbl rel by tg ids failed, err: %v, tg ids: %v, rid: %s", err, tgIDs, kt.Rid)
		return err
	}

	// 数据处理
	ruleIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	lbIDs := make([]string, 0)
	tgIDTypeMap := make(map[string]enumor.RuleType)
	for _, rel := range tgLblRel {
		ruleIDs = append(ruleIDs, rel.ListenerRuleID)
		lblIDs = append(lblIDs, rel.LblID)
		lbIDs = append(lbIDs, rel.LbID)
		tgIDTypeMap[rel.TargetGroupID] = rel.ListenerRuleType
	}
	layer4Targets, layer7Targets := make([]corelb.BaseTarget, 0), make([]corelb.BaseTarget, 0)
	for _, target := range targets {
		typeVal := tgIDTypeMap[target.TargetGroupID]
		switch typeVal {
		case enumor.Layer4RuleType:
			layer4Targets = append(layer4Targets, target)
		case enumor.Layer7RuleType:
			layer7Targets = append(layer7Targets, target)
		default:
			logs.Errorf("invalid target group type, id: %s, type: %s, rid: %s", target.TargetGroupID, typeVal, kt.Rid)
			return errf.Newf(errf.InvalidParameter, "invalid target group type, id: %s, type: %s", target.TargetGroupID,
				typeVal)
		}
	}

	// 查询规则
	ruleMap, err := t.getRuleByIDs(kt, ruleIDs)
	if err != nil {
		logs.Errorf("get rule by ids failed, err: %v, ids: %v, rid: %s", err, ruleIDs, kt.Rid)
		return err
	}

	// 查询监听器
	lblMap, err := t.getLblByIDs(kt, lblIDs)
	if err != nil {
		logs.Errorf("get lbl by ids failed, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
		return err
	}

	// 查询clb
	lbMap, err := t.getLbByIDs(kt, lbIDs)
	if err != nil {
		logs.Errorf("get lb by ids failed, err: %v, lbIDs: %v, rid: %s", err, lbIDs, kt.Rid)
		return err
	}

	// 将数据写入excel
	if err = writeTCloudLayer4Rs(kt, t.vendor, zipOperator, lbMap, lblMap, tgLblRel, layer4Targets); err != nil {
		logs.Errorf("write tcloud layer4 target failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	err = writeTCloudLayer7Rs(kt, t.vendor, zipOperator, lbMap, lblMap, ruleMap, tgLblRel, layer7Targets)
	if err != nil {
		logs.Errorf("write tcloud layer7 target failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (t *targetExporter) getTargetByIDs(kt *kit.Kit, ids []string) ([]corelb.BaseTarget, error) {
	ids = slice.Unique(ids)

	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		resp, err := t.client.DataService().Global.LoadBalancer.ListTarget(kt, &req)
		if err != nil {
			logs.Errorf("list target failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		targets = append(targets, resp.Details...)
	}

	return targets, nil
}

func (t *targetExporter) getTgLblRelByTgIDs(kt *kit.Kit, tgIDs []string) ([]corelb.BaseTargetListenerRuleRel, error) {
	tgIDs = slice.Unique(tgIDs)

	rels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("target_group_id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		resp, err := t.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &req)
		if err != nil {
			logs.Errorf("list tg lbl rel failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		rels = append(rels, resp.Details...)
	}

	return rels, nil
}

func (t *targetExporter) getRuleByIDs(kt *kit.Kit, ids []string) (map[string]corelb.TCloudLbUrlRule, error) {
	ids = slice.Unique(ids)

	ruleMap := make(map[string]corelb.TCloudLbUrlRule)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		switch t.vendor {
		case enumor.TCloud:
			resp, err := t.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &req)
			if err != nil {
				logs.Errorf("list rule failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, err
			}
			for _, rule := range resp.Details {
				ruleMap[rule.ID] = rule
			}

		default:
			return nil, fmt.Errorf("unsupported vendor: %s", t.vendor)
		}
	}

	return ruleMap, nil
}

func (t *targetExporter) getLblByIDs(kt *kit.Kit, ids []string) (map[string]corelb.TCloudListener, error) {
	ids = slice.Unique(ids)

	lblMap := make(map[string]corelb.TCloudListener)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		switch t.vendor {
		case enumor.TCloud:
			resp, err := t.client.DataService().TCloud.LoadBalancer.ListListener(kt, &req)
			if err != nil {
				logs.Errorf("list listener failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, err
			}
			for _, lbl := range resp.Details {
				lblMap[lbl.ID] = lbl
			}
		default:
			return nil, fmt.Errorf("unsupported vendor: %s", t.vendor)
		}
	}

	return lblMap, nil
}

func (t *targetExporter) getLbByIDs(kt *kit.Kit, ids []string) (map[string]corelb.BaseLoadBalancer, error) {
	ids = slice.Unique(ids)

	lbMap := make(map[string]corelb.BaseLoadBalancer)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		resp, err := t.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, &req)
		if err != nil {
			logs.Errorf("list lb failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, lb := range resp.Details {
			lbMap[lb.ID] = lb
		}
	}

	return lbMap, nil
}
