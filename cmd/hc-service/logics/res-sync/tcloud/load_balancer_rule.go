/*
 *
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

package tcloud

import (
	"hcm/cmd/hc-service/logics/res-sync/common"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// LoadBalancerLayer4Rule 同步负载均衡下的4层监听器规则，四层规则一次同步
func (cli *client) LoadBalancerLayer4Rule(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) (*SyncResult, error) {
	cloudListeners, err := cli.listListenerFromCloud(kt, opt)
	if err != nil {
		logs.Errorf("fail to list listener for sync, err: %v, opt:%+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}
	// 过滤四层规则
	l4Listeners := slice.Filter(cloudListeners, func(cloud typeslb.TCloudListener) bool {
		return !cloud.GetProtocol().IsLayer7Protocol()
	})
	dbRules, err := cli.listL4RuleFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(l4Listeners) == 0 && len(dbRules) == 0 {
		return new(SyncResult), nil
	}

	// 新增实例应该在同步监听器的时候附带创建，云上已删除的规则应该在监听器同步时被删除
	_, updateMap, _ := common.Diff[typeslb.TCloudListener, corelb.TCloudLbUrlRule](
		l4Listeners, dbRules, isLayer4RuleChange)

	// 更新变更监听器，更新对应四层/七层 规则
	if err = cli.updateLayer4Rule(kt, updateMap); err != nil {
		return nil, err
	}

	return new(SyncResult), nil

}
func (cli *client) listL4RuleFromDB(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) ([]corelb.TCloudLbUrlRule, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", opt.LBID),
			tools.RuleEqual("rule_type", enumor.Layer4RuleType)),
		Page: core.NewDefaultBasePage(),
	}

	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of lb(%s) for sync, err: %v, rid: %s", opt.LBID, err, kt.Rid)
		return nil, err
	}
	return ruleResp.Details, nil
}

// ListenerLayer7Rule 同步指定监听器下的7层规则，7层按监听器同步
func (cli *client) ListenerLayer7Rule(kt *kit.Kit, lblOpt *SyncListenerOfSingleLBOption,
	ruleOpt *SyncLayer7RuleOption) (*SyncResult, error) {
	return nil, nil
}

func (cli *client) updateLayer4Rule(kt *kit.Kit, updateMap map[string]typeslb.TCloudListener) error {

	if len(updateMap) == 0 {
		return nil
	}
	updateReq := &cloud.TCloudUrlRuleBatchUpdateReq{}
	for id, listener := range updateMap {
		updateReq.UrlRules = append(updateReq.UrlRules,
			&cloud.TCloudUrlRuleUpdate{
				ID:            id,
				Scheduler:     cvt.PtrToVal(listener.Scheduler),
				SessionType:   cvt.PtrToVal(listener.SessionType),
				SessionExpire: listener.SessionExpireTime,
				HealthCheck:   convHealthCheck(listener.HealthCheck),
				Certificate:   convCert(listener.Certificate),
			},
		)
	}
	err := cli.dbCli.TCloud.LoadBalancer.BatchUpdateTCloudUrlRule(kt, updateReq)
	if err != nil {
		logs.Errorf("fail to update tcloud url rule, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func convHealthCheck(cloud *tclb.HealthCheck) *corelb.TCloudHealthCheckInfo {
	if cloud == nil {
		return nil
	}
	db := &corelb.TCloudHealthCheckInfo{
		HealthSwitch:    cloud.HealthSwitch,
		TimeOut:         cloud.TimeOut,
		IntervalTime:    cloud.IntervalTime,
		HealthNum:       cloud.HealthNum,
		UnHealthNum:     cloud.UnHealthNum,
		HttpCode:        cloud.HttpCode,
		CheckPort:       cloud.CheckPort,
		CheckType:       cloud.CheckType,
		HttpVersion:     cloud.HttpVersion,
		HttpCheckPath:   cloud.HttpCheckPath,
		HttpCheckDomain: cloud.HttpCheckDomain,
		HttpCheckMethod: cloud.HttpCheckMethod,
		SourceIpType:    cloud.SourceIpType,
		ContextType:     cloud.ContextType,
		SendContext:     cloud.SendContext,
		RecvContext:     cloud.RecvContext,
		ExtendedCode:    cloud.ExtendedCode,
	}

	return db
}

// 四层监听器的健康检查这些信息保存在规则里，需要检查对应的规则
func isLayer4RuleChange(cloud typeslb.TCloudListener, db corelb.TCloudLbUrlRule) bool {

	if cvt.PtrToVal(cloud.Scheduler) != db.Scheduler {
		return true
	}
	if cvt.PtrToVal(cloud.SessionType) != db.SessionType {
		return true
	}

	if isHealthCheckChange(cloud.HealthCheck, db.HealthCheck) {
		return true
	}
	if isListenerCertChange(cloud.Certificate, db.Certificate) {
		return true
	}
	return false
}

// SyncLayer7RuleOption 同步7层规则选项，包含 监听器信息
type SyncLayer7RuleOption struct {
}
