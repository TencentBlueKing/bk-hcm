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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/concurrence"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// Listener 同步多个负载均衡下的监听器
func (cli *client) Listener(kt *kit.Kit, params *SyncListenerParams) (*SyncResult, error) {

	if err := validator.ValidateTool(params); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 并发同步多个监听器
	var syncResult *SyncResult
	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.LbInfos,
		func(lb corelb.TCloudLoadBalancer) error {
			syncOpt := &SyncListenerOfSingleLBOption{
				AccountID: params.AccountID,
				Region:    params.Region,
				BizID:     lb.BkBizID,
				LBID:      lb.ID,
				CloudLBID: lb.CloudID,
			}
			var err error
			if syncResult, err = cli.listener(kt, syncOpt); err != nil {
				logs.ErrorDepthf(1, "[%s] account: %s lb: %s sync listener failed, err: %v, rid: %s",
					enumor.TCloud, params.AccountID, lb.CloudID, err, kt.Rid)
				return err
			}

			return nil
		})
	if err != nil {
		return nil, err
	}
	return syncResult, nil
}

// LoadBalancerListener 同步指定负载均衡均衡下的监听器
func (cli *client) listener(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) (
	*SyncResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudListeners, err := cli.listListenerFromCloud(kt, opt)
	if err != nil {
		logs.Errorf("fail to list listener for sync, err: %v, opt:%+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	dbListeners, err := cli.listListenerFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(cloudListeners) == 0 && len(dbListeners) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TCloudListener, common.TCloudComposedListener](
		cloudListeners, dbListeners, isListenerChange)

	// 删除云上已经删除的监听器实例
	if err = cli.deleteListener(kt, opt, delCloudIDs); err != nil {
		return nil, err
	}

	// 创建云上新增监听器实例
	_, err = cli.createListener(kt, opt, addSlice)
	if err != nil {
		return nil, err
	}
	// 更新变更监听器
	if err = cli.updateListener(kt, opt.BizID, updateMap); err != nil {
		return nil, err
	}

	//  TODO: 同步七层规则和目标组

	return new(SyncResult), nil
}

// 获取云上监听器列表
func (cli *client) listListenerFromCloud(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) ([]typeslb.TCloudListener,
	error) {
	listOpt := &typeslb.TCloudListListenersOption{
		Region:         opt.Region,
		LoadBalancerId: opt.CloudLBID,
	}
	return cli.cloudCli.ListListener(kt, listOpt)
}

// 获取本地监听器列表
func (cli *client) listListenerFromDB(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) ([]common.TCloudComposedListener,
	error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("lb_id", opt.LBID),
		Page:   core.NewDefaultBasePage(),
	}
	lblResp, err := cli.dbCli.TCloud.LoadBalancer.ListListener(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list listener of lb(%s) for sync, err: %v, rid: %s", opt.LBID, err, kt.Rid)
		return nil, err
	}
	listReq.Filter = tools.ExpressionAnd(
		tools.RuleEqual("lb_id", opt.LBID),
		tools.RuleEqual("rule_type", enumor.Layer4RuleType))
	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of lb(%s) for sync, err: %v, rid: %s", opt.LBID, err, kt.Rid)
		return nil, err
	}
	// lb id as key
	ruleMap := make(map[string]*corelb.TCloudLbUrlRule)
	for _, r := range ruleResp.Details {
		ruleMap[r.LbID] = cvt.ValToPtr(r)
	}

	// merge to one type
	result := make([]common.TCloudComposedListener, 0, len(lblResp.Details))
	for _, lbl := range lblResp.Details {
		result = append(result, common.TCloudComposedListener{Listener: cvt.ValToPtr(lbl), Rule: ruleMap[lbl.ID]})
	}
	return result, nil
}

func (cli *client) deleteListener(kt *kit.Kit, opt *SyncListenerOfSingleLBOption, cloudIds []string) error {
	if len(cloudIds) == 0 {
		return nil
	}
	delReq := &dataproto.LoadBalancerBatchDeleteReq{Filter: tools.ContainersExpression("cloud_id", cloudIds)}
	err := cli.dbCli.Global.LoadBalancer.DeleteListener(kt, delReq)
	if err != nil {
		logs.Errorf("fail to delete listeners(ids:%v) while sync, err: %v, syncOpt: %+v, rid: %s",
			cloudIds, err, opt, kt.Rid)
		return err
	}
	return nil
}

func (cli *client) createListener(kt *kit.Kit, syncOpt *SyncListenerOfSingleLBOption,
	addSlice []typeslb.TCloudListener) ([]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}
	dbListeners := make([]dataproto.ListenersCreateReq, 0, len(addSlice))
	dbRules := make([]dataproto.ListenerWithRuleCreateReq, 0)
	for _, lbl := range addSlice {
		if cvt.PtrToVal((*enumor.ProtocolType)(lbl.Protocol)).IsLayer7Protocol() {
			dbListeners = append(dbListeners, dataproto.ListenersCreateReq{
				CloudID:       lbl.GetCloudID(),
				Name:          cvt.PtrToVal(lbl.ListenerName),
				Vendor:        enumor.TCloud,
				AccountID:     syncOpt.AccountID,
				BkBizID:       syncOpt.BizID,
				LbID:          syncOpt.LBID,
				CloudLbID:     syncOpt.CloudLBID,
				Protocol:      cvt.PtrToVal((*enumor.ProtocolType)(lbl.Protocol)),
				Port:          cvt.PtrToVal(lbl.Port),
				DefaultDomain: getDefaultDomain(lbl),
			})
			// for layer 7 only create listeners itself
			continue
		}
		// layer 4 create with rule
		dbRules = append(dbRules, dataproto.ListenerWithRuleCreateReq{
			CloudID:       lbl.GetCloudID(),
			Name:          cvt.PtrToVal(lbl.ListenerName),
			Vendor:        enumor.TCloud,
			AccountID:     syncOpt.AccountID,
			BkBizID:       syncOpt.BizID,
			LbID:          syncOpt.LBID,
			CloudLbID:     syncOpt.CloudLBID,
			Protocol:      cvt.PtrToVal((*enumor.ProtocolType)(lbl.Protocol)),
			Port:          cvt.PtrToVal(lbl.Port),
			CloudRuleID:   lbl.GetCloudID(),
			Scheduler:     cvt.PtrToVal(lbl.Scheduler),
			RuleType:      enumor.Layer4RuleType,
			SessionType:   cvt.PtrToVal(lbl.SessionType),
			SessionExpire: cvt.PtrToVal(lbl.SessionExpireTime),
			SniSwitch:     enumor.SniType(cvt.PtrToVal(lbl.SniSwitch)),
			Certificate:   convCert(lbl.Certificate),
		})
	}
	createdIDs := make([]string, 0, len(addSlice))
	if len(dbListeners) > 0 {
		lblCreated, err := cli.dbCli.TCloud.LoadBalancer.BatchCreateTCloudListener(kt,
			&dataproto.ListenerBatchCreateReq{Listeners: dbListeners})
		if err != nil {
			logs.Errorf("fail to create listener while sync, err: %v syncOpt: %+v, rid: %s",
				err, syncOpt, kt.Rid)
			return nil, err
		}
		createdIDs = append(createdIDs, lblCreated.IDs...)
	}

	if len(dbRules) > 0 {
		ruleCreated, err := cli.dbCli.TCloud.LoadBalancer.BatchCreateTCloudListenerWithRule(kt,
			&dataproto.ListenerWithRuleBatchCreateReq{ListenerWithRules: dbRules})
		if err != nil {
			logs.Errorf("fail to create listener with rule while sync, err: %v syncOpt: %+v, rid: %s",
				err, syncOpt, kt.Rid)
			return nil, err
		}
		createdIDs = append(createdIDs, ruleCreated.IDs...)
	}

	return createdIDs, nil
}

func (cli *client) updateListener(kt *kit.Kit, bizID int64, updateMap map[string]typeslb.TCloudListener) error {

	if len(updateMap) == 0 {
		return nil
	}
	updates := make([]*dataproto.TCloudListenerUpdate, 0, len(updateMap))

	for id, lbl := range updateMap {

		updates = append(updates, &dataproto.TCloudListenerUpdate{
			ID:            id,
			Name:          cvt.PtrToVal(lbl.ListenerName),
			BkBizID:       bizID,
			SniSwitch:     enumor.SniType(cvt.PtrToVal(lbl.SniSwitch)),
			DefaultDomain: getDefaultDomain(lbl),
			Extension: &corelb.TCloudListenerExtension{
				Certificate: convCert(lbl.Certificate),
			},
		})
	}

	err := cli.dbCli.TCloud.LoadBalancer.BatchUpdateTCloudListener(kt,
		&dataproto.TCloudListenerUpdateReq{Listeners: updates})
	if err != nil {
		logs.Errorf("fail to update listener while sync, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 更新规则
	return nil
}
func convCert(cloud *tclb.CertificateOutput) *corelb.TCloudCertificateInfo {
	if cloud == nil {
		return nil
	}
	db := &corelb.TCloudCertificateInfo{
		SSLMode:   cloud.SSLMode,
		CaCloudID: cloud.CertCaId,
	}
	if cloud.CertId != nil {
		db.CertCloudIDs = append(db.CertCloudIDs, cvt.PtrToVal(cloud.CertId))
	}
	for _, cloudCertID := range cloud.ExtCertIds {
		db.CertCloudIDs = append(db.CertCloudIDs, cvt.PtrToVal(cloudCertID))
	}
	return db
}

// isListenerChange 四层规则有健康检查这类信息在监听器上，七层规则可能有0-n条规则，对应字段在规则同步时处理
func isListenerChange(cloud typeslb.TCloudListener, db common.TCloudComposedListener) bool {

	// 通用字段
	if cvt.PtrToVal(cloud.ListenerName) != db.Name {
		return true
	}
	protocol := enumor.ProtocolType(cvt.PtrToVal(cloud.Protocol))
	switch protocol {
	case enumor.HttpProtocol:
		// http 只有名称和默认域名可以变
		if getDefaultDomain(cloud) != db.DefaultDomain {
			return true
		}
	case enumor.HttpsProtocol:
		if isHttpsListenerChanged(cloud, db) {
			return true
		}
	default:
		// 	其他为4层协议
		if isLayer4Changed(cloud, db) {
			return true
		}
	}

	return false
}

func isLayer4Changed(cloud typeslb.TCloudListener, db common.TCloudComposedListener) bool {

	if isListenerCertChange(cloud.Certificate, db.Extension.Certificate) {
		return true
	}
	// 规则单独检查

	return false
}

func isHealthCheckChange(cloud *tclb.HealthCheck, db *corelb.TCloudHealthCheckInfo) bool {
	if cloud == nil || db == nil {
		// 云上和本地都为空 则是未变化，否则需要更新本地
		return !(cloud == nil && db == nil)
	}
	if assert.IsPtrInt64Equal(cloud.HealthSwitch, db.HealthSwitch) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.TimeOut, db.TimeOut) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.IntervalTime, db.IntervalTime) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.HealthNum, db.HealthNum) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.UnHealthNum, db.UnHealthNum) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.HttpCode, db.HttpCode) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.HttpCheckPath, db.HttpCheckPath) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.HttpCheckDomain, db.HttpCheckDomain) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.HttpCheckMethod, db.HttpCheckMethod) {
		return true
	}
	if assert.IsPtrInt64Equal(cloud.CheckPort, db.CheckPort) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.ContextType, db.ContextType) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.SendContext, db.SendContext) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.RecvContext, db.RecvContext) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.CheckType, db.CheckType) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.HttpVersion, db.HttpVersion) {
		return true
	}

	if assert.IsPtrInt64Equal(cloud.SourceIpType, db.SourceIpType) {
		return true
	}
	if assert.IsPtrStringEqual(cloud.ExtendedCode, db.ExtendedCode) {
		return true
	}

	return false
}

func isHttpsListenerChanged(cloud typeslb.TCloudListener, db common.TCloudComposedListener) bool {
	if db.DefaultDomain != getDefaultDomain(cloud) {
		return true
	}
	if cvt.PtrToVal(cloud.SniSwitch) != int64(db.SniSwitch) {
		return true
	}
	if db.Extension == nil {
		return true
	}

	if isListenerCertChange(cloud.Certificate, db.Extension.Certificate) {
		return true
	}
	return false
}

func isListenerCertChange(cloud *tclb.CertificateOutput, db *corelb.TCloudCertificateInfo) bool {
	if cloud == nil || db == nil {
		// 云上和本地都为空 则是未变化
		return !(cloud == nil && db == nil)
	}

	if !assert.IsPtrStringEqual(cloud.SSLMode, db.SSLMode) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CertCaId, db.CaCloudID) {
		return true
	}
	// 云上有，本地没有
	if len(cvt.PtrToVal(cloud.CertId)) != 0 && len(db.CertCloudIDs) == 0 {
		return true
	}
	// 云上没有，本地有
	if len(cvt.PtrToVal(cloud.CertId)) == 0 && len(db.CertCloudIDs) > 0 {
		return true
	}

	// 本地和云上都有，但是和云上不相等
	if len(db.CertCloudIDs) > 0 && cvt.PtrToVal(cloud.CertId) != db.CertCloudIDs[0] {
		return true
	}
	// 本地和云上都有，但是数量不相等
	if len(db.CertCloudIDs) != (len(cloud.ExtCertIds) + 1) {
		// 数量不相等
		return true
	}
	// 要求证书按顺序相等。
	for i := range cloud.ExtCertIds {
		if db.CertCloudIDs[i+1] != cvt.PtrToVal(cloud.ExtCertIds[i]) {
			return true
		}
	}
	return false
}

// 四层监听器的健康检查这些信息保存在规则里，需要检查对应的规则
func isLayer4RuleChange(cloud typeslb.TCloudListener, db *corelb.TCloudLbUrlRule) bool {
	if db == nil {
		return true
	}
	if cvt.PtrToVal(cloud.Scheduler) != db.Scheduler {
		return true
	}
	if cvt.PtrToVal(cloud.SessionType) != db.SessionType {
		return true
	}

	if isHealthCheckChange(cloud.HealthCheck, db.HealthCheck) {
		return true
	}
	return false
}

func getDefaultDomain(cloud typeslb.TCloudListener) string {
	// 需要去规则中捞
	for _, rule := range cloud.Rules {
		if rule != nil && cvt.PtrToVal(rule.DefaultServer) {
			return cvt.PtrToVal(rule.Domain)
		}
	}
	return ""
}

// SyncListenerOfSingleLBOption ...
type SyncListenerOfSingleLBOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	BizID     int64  `json:"biz_id" validate:"required"`

	// 对应的负载均衡
	LBID      string `json:"lbid" validate:"required"`
	CloudLBID string `json:"cloud_lbid" validate:"required"`
}

// Validate ...
func (o *SyncListenerOfSingleLBOption) Validate() error {
	return validator.Validate.Struct(o)
}

// SyncListenerParams ...
type SyncListenerParams struct {
	AccountID string                      `json:"account_id" validate:"required"`
	Region    string                      `json:"region" validate:"required"`
	LbInfos   []corelb.TCloudLoadBalancer `json:"lb_infos" validate:"required,min=1"`
}

// Validate ...
func (o *SyncListenerParams) Validate() error {
	return validator.Validate.Struct(o)
}
