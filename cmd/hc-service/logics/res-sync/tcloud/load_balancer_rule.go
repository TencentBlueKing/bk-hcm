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
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// LoadBalancerRule 规则同步
func (cli *client) loadBalancerRule(kt *kit.Kit, opt *SyncListenerOfSingleLBOption,
	cloudListeners []typeslb.TCloudListener) (any, error) {

	var l4Listeners, l7Listeners []typeslb.TCloudListener
	for _, listener := range cloudListeners {
		if listener.GetProtocol().IsLayer7Protocol() {
			l7Listeners = append(l7Listeners, listener)
			continue
		}
		l4Listeners = append(l4Listeners, listener)
	}
	_, err := cli.LoadBalancerLayer4Rule(kt, opt.LBID, l4Listeners)
	if err != nil {
		return nil, err
	}
	l7Opt := &SyncLayer7RuleOption{
		LBID:            opt.LBID,
		CloudLBID:       opt.CloudLBID,
		ListenerID:      "",
		CloudListenerID: "",
	}
	dbListeners, err := cli.listListenerFromDB(kt, opt)
	if err != nil {
		return nil, err
	}
	dbListenerMap := make(map[string]*corelb.TCloudListener)
	for _, dbLbl := range dbListeners {
		dbListenerMap[dbLbl.CloudID] = cvt.ValToPtr(dbLbl)
	}

	// 逐个同步监听器下的规则
	for _, listener := range l7Listeners {
		l7Opt.CloudListenerID = cvt.PtrToVal(listener.ListenerId)
		dbLbl := dbListenerMap[listener.GetCloudID()]
		if dbLbl == nil {
			// 	云上新建的监听器，等待下次同步
			logs.Infof("found new listener from cloud, id: %s", listener.GetCloudID())
			continue
		}
		l7Opt.ListenerID = dbLbl.ID
		_, err := cli.ListenerLayer7Rule(kt, l7Opt, listener)
		if err != nil {
			logs.Errorf("fail to sync rules of listener, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// LoadBalancerLayer4Rule 同步负载均衡下的4层监听器规则，四层规则一次同步
func (cli *client) LoadBalancerLayer4Rule(kt *kit.Kit, lbID string, l4Listeners []typeslb.TCloudListener) (
	*SyncResult, error) {

	dbRules, err := cli.listL4RuleFromDB(kt, lbID)
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

// ListenerLayer7Rule 同步指定监听器下的7层规则，7层按监听器同步
func (cli *client) ListenerLayer7Rule(kt *kit.Kit, opt *SyncLayer7RuleOption, cloudListener typeslb.TCloudListener) (
	*SyncResult, error) {
	// 对于七层规则逐个监听器进行同步

	dbRules, err := cli.listL7RuleFromDB(kt, cvt.PtrToVal(cloudListener.ListenerId))
	if err != nil {
		return nil, err
	}

	if len(cloudListener.Rules) == 0 && len(dbRules) == 0 {
		return new(SyncResult), nil
	}

	cloudRules := make([]typeslb.TCloudUrlRule, 0, len(cloudListener.Rules))
	for _, rule := range cloudListener.Rules {
		cloudRules = append(cloudRules, typeslb.TCloudUrlRule{RuleOutput: rule})
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TCloudUrlRule, corelb.TCloudLbUrlRule](
		cloudRules, dbRules, isLayer7RuleChange)

	if err = cli.deleteLayer7Rule(kt, delCloudIDs); err != nil {
		return nil, err
	}

	if err = cli.updateLayer7Rule(kt, updateMap); err != nil {
		return nil, err
	}

	if _, err = cli.createLayer7Rule(kt, opt, addSlice); err != nil {
		return nil, err
	}
	return nil, nil
}

func (cli *client) listL4RuleFromDB(kt *kit.Kit, lbID string) ([]corelb.TCloudLbUrlRule, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("rule_type", enumor.Layer4RuleType)),
		Page: core.NewDefaultBasePage(),
	}

	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of lb(%s) for sync, err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, err
	}
	return ruleResp.Details, nil
}

func (cli *client) listL7RuleFromDB(kt *kit.Kit, cloudLBLID string) ([]corelb.TCloudLbUrlRule, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_lbl_id", cloudLBLID),
			tools.RuleEqual("rule_type", enumor.Layer7RuleType)),
		Page: core.NewDefaultBasePage(),
	}

	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of lbl(%s) for sync, err: %v, rid: %s", cloudLBLID, err, kt.Rid)
		return nil, err
	}
	return ruleResp.Details, nil
}
func (cli *client) updateLayer4Rule(kt *kit.Kit, updateMap map[string]typeslb.TCloudListener) error {

	if len(updateMap) == 0 {
		return nil
	}
	updateReq := &dataproto.TCloudUrlRuleBatchUpdateReq{}
	for id, listener := range updateMap {
		updateReq.UrlRules = append(updateReq.UrlRules,
			&dataproto.TCloudUrlRuleUpdate{
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

func (cli *client) deleteLayer7Rule(kt *kit.Kit, cloudIds []string) error {

	if len(cloudIds) == 0 {
		return nil
	}
	delReq := &dataproto.LoadBalancerBatchDeleteReq{Filter: tools.ContainersExpression("cloud_id", cloudIds)}
	err := cli.dbCli.TCloud.LoadBalancer.BatchDeleteTCloudUrlRule(kt, delReq)
	if err != nil {
		logs.Errorf("fail to delete listeners(ids:%v) while sync, err: %v, rid: %s",
			cloudIds, err, kt.Rid)
		return err
	}
	return nil
}

func (cli *client) createLayer7Rule(kt *kit.Kit, opt *SyncLayer7RuleOption,
	addSlice []typeslb.TCloudUrlRule) ([]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}

	dbRules := make([]dataproto.TCloudUrlRuleCreate, 0)
	for _, cloud := range addSlice {

		dbRules = append(dbRules, dataproto.TCloudUrlRuleCreate{
			LbID:       opt.LBID,
			CloudLbID:  opt.CloudLBID,
			LblID:      opt.ListenerID,
			CloudLBLID: opt.CloudListenerID,
			CloudID:    cloud.GetCloudID(),
			RuleType:   enumor.Layer7RuleType,

			Domain:    cvt.PtrToVal(cloud.Domain),
			URL:       cvt.PtrToVal(cloud.Url),
			Scheduler: cvt.PtrToVal(cloud.Scheduler),

			SessionExpire: cvt.PtrToVal(cloud.SessionExpireTime),
			HealthCheck:   convHealthCheck(cloud.HealthCheck),
			Certificate:   convCert(cloud.Certificate),
		})
	}

	ruleCreated, err := cli.dbCli.TCloud.LoadBalancer.BatchCreateTCloudUrlRule(kt,
		&dataproto.TCloudUrlRuleBatchCreateReq{UrlRules: dbRules})
	if err != nil {
		logs.Errorf("fail to create rule while sync, err: %v syncOpt: %+v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	return ruleCreated.IDs, nil
}

func (cli *client) updateLayer7Rule(kt *kit.Kit, updateMap map[string]typeslb.TCloudUrlRule) error {

	if len(updateMap) == 0 {
		return nil
	}
	updates := make([]*dataproto.TCloudUrlRuleUpdate, 0, len(updateMap))

	for id, rule := range updateMap {

		updates = append(updates, &dataproto.TCloudUrlRuleUpdate{
			ID:            id,
			Domain:        cvt.PtrToVal(rule.Domain),
			URL:           cvt.PtrToVal(rule.Url),
			Scheduler:     cvt.PtrToVal(rule.Scheduler),
			SessionExpire: rule.SessionExpireTime,
			HealthCheck:   convHealthCheck(rule.HealthCheck),
			Certificate:   convCert(rule.Certificate),
		})
	}

	err := cli.dbCli.TCloud.LoadBalancer.BatchUpdateTCloudUrlRule(kt,
		&dataproto.TCloudUrlRuleBatchUpdateReq{UrlRules: updates})
	if err != nil {
		logs.Errorf("fail to update rule while sync, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
func convHealthCheck(cloud *tclb.HealthCheck) *corelb.TCloudHealthCheckInfo {
	if cloud == nil {
		return nil
	}

	db := &corelb.TCloudHealthCheckInfo{
		// 确保0值时不会存储为nil
		HealthSwitch:    cvt.ValToPtr(cvt.PtrToVal(cloud.HealthSwitch)),
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

	if isHealthCheckChange(cloud.HealthCheck, db.HealthCheck, false) {
		return true
	}
	if isListenerCertChange(cloud.Certificate, db.Certificate) {
		return true
	}
	return false
}

func isLayer7RuleChange(cloud typeslb.TCloudUrlRule, db corelb.TCloudLbUrlRule) bool {

	if cvt.PtrToVal(cloud.Url) != db.URL {
		return true
	}
	if cvt.PtrToVal(cloud.SessionExpireTime) != db.SessionExpire {
		return true
	}
	if cvt.PtrToVal(cloud.Scheduler) != db.Scheduler {
		return true
	}
	if cvt.PtrToVal(cloud.Domain) != db.Domain {
		return true
	}

	if isHealthCheckChange(cloud.HealthCheck, db.HealthCheck, true) {
		return true
	}
	if isListenerCertChange(cloud.Certificate, db.Certificate) {
		return true
	}

	return false
}

// SyncLayer7RuleOption 同步7层规则选项，包含 监听器信息
type SyncLayer7RuleOption struct {

	// 对应的负载均衡
	LBID      string `json:"lb_id" validate:"required"`
	CloudLBID string `json:"cloud_lb_id" validate:"required"`

	// 对应的监听器
	ListenerID      string `json:"lbl_id" validate:"required"`
	CloudListenerID string `json:"cloud_lbl_id" validate:"required"`
}
