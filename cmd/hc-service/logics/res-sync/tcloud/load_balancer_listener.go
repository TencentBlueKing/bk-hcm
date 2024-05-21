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

// listenerByLbBatch 同步多个负载均衡下的监听器：
func (cli *client) listenerByLbBatch(kt *kit.Kit, params *SyncListenerOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 并发同步多个负载均衡下的监听器
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
			if syncResult, err = cli.Listener(kt, syncOpt); err != nil {
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

// Listener 2. 同步指定负载均衡均衡下的监听器
func (cli *client) Listener(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) (
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

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TCloudListener, corelb.TCloudListener](
		cloudListeners, dbListeners, isListenerChange)

	// 删除云上已经删除的监听器实例
	if err = cli.deleteListener(kt, opt, delCloudIDs); err != nil {
		return nil, err
	}

	// 创建云上新增监听器实例， 对于四层规则一起创建对应的规则
	_, err = cli.createListener(kt, opt, addSlice)
	if err != nil {
		return nil, err
	}
	// 更新变更监听器，不更新对应四层/七层 规则
	if err = cli.updateListener(kt, opt.BizID, updateMap); err != nil {
		return nil, err
	}

	// 同步监听器下的四层/七层规则
	_, err = cli.loadBalancerRule(kt, opt, cloudListeners)
	if err != nil {
		logs.Errorf("fail to sync listener rule for sync listener, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}
	targetParam := &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  nil,
	}

	// 同步相关目标组中的rs
	err = cli.ListenerTargets(kt, targetParam, opt)
	if err != nil {
		logs.Errorf("fail to sync listener targets for sync listener, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	// 同步本地目标组
	err = cli.LocalTargetGroup(kt, targetParam, opt, cloudListeners)
	if err != nil {
		logs.Errorf("fail to sync target group for listener, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

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
func (cli *client) listListenerFromDB(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) ([]corelb.TCloudListener,
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
	return lblResp.Details, nil
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
	dbListeners := make([]dataproto.ListenersCreateReq[corelb.TCloudListenerExtension], 0, len(addSlice))
	dbRules := make([]dataproto.ListenerWithRuleCreateReq, 0)
	for _, lbl := range addSlice {
		if lbl.GetProtocol().IsLayer7Protocol() {
			dbListeners = append(dbListeners, dataproto.ListenersCreateReq[corelb.TCloudListenerExtension]{
				CloudID:       lbl.GetCloudID(),
				Name:          cvt.PtrToVal(lbl.ListenerName),
				Vendor:        enumor.TCloud,
				AccountID:     syncOpt.AccountID,
				BkBizID:       syncOpt.BizID,
				LbID:          syncOpt.LBID,
				CloudLbID:     syncOpt.CloudLBID,
				Protocol:      lbl.GetProtocol(),
				Port:          cvt.PtrToVal(lbl.Port),
				DefaultDomain: getDefaultDomain(lbl),
				Extension: &corelb.TCloudListenerExtension{
					Certificate: convCert(lbl.Certificate),
					EndPort:     lbl.EndPort,
				}})
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
			Protocol:      lbl.GetProtocol(),
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
			&dataproto.TCloudListenerBatchCreateReq{Listeners: dbListeners})
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
				EndPort:     lbl.EndPort,
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
func isListenerChange(cloud typeslb.TCloudListener, db corelb.TCloudListener) bool {

	// 通用字段
	if cvt.PtrToVal(cloud.ListenerName) != db.Name {
		return true
	}
	if cvt.PtrToVal(cloud.EndPort) != cvt.PtrToVal(db.Extension.EndPort) {
		return true
	}
	switch cloud.GetProtocol() {
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
		if isLayer4ListenerChanged(cloud, db) {
			return true
		}
	}

	return false
}

func isLayer4ListenerChanged(cloud typeslb.TCloudListener, db corelb.TCloudListener) bool {

	if isListenerCertChange(cloud.Certificate, db.Extension.Certificate) {
		return true
	}
	// 规则单独检查

	return false
}

// 七层规则不支持设置检查端口
func isHealthCheckChange(cloud *tclb.HealthCheck, db *corelb.TCloudHealthCheckInfo, isL7 bool) bool {
	if cloud == nil || db == nil {
		// 云上和本地都为空 则是未变化，否则需要更新本地
		return !(cloud == nil && db == nil)
	}
	if !assert.IsPtrInt64Equal(cloud.HealthSwitch, db.HealthSwitch) {
		return true
	}
	if !assert.IsPtrInt64Equal(cloud.TimeOut, db.TimeOut) {
		return true
	}
	if !assert.IsPtrInt64Equal(cloud.IntervalTime, db.IntervalTime) {
		return true
	}
	if !assert.IsPtrInt64Equal(cloud.HealthNum, db.HealthNum) {
		return true
	}
	if !assert.IsPtrInt64Equal(cloud.UnHealthNum, db.UnHealthNum) {
		return true
	}
	if !assert.IsPtrInt64Equal(cloud.HttpCode, db.HttpCode) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.HttpCheckPath, db.HttpCheckPath) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.HttpCheckDomain, db.HttpCheckDomain) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.HttpCheckMethod, db.HttpCheckMethod) {
		return true
	}
	// 七层规则不支持设置检查端口, 这里不比较该数据
	if isL7 && !assert.IsPtrInt64Equal(cloud.CheckPort, db.CheckPort) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.ContextType, db.ContextType) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.SendContext, db.SendContext) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.RecvContext, db.RecvContext) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.CheckType, db.CheckType) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.HttpVersion, db.HttpVersion) {
		return true
	}

	if !assert.IsPtrInt64Equal(cloud.SourceIpType, db.SourceIpType) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.ExtendedCode, db.ExtendedCode) {
		return true
	}

	return false
}

func isHttpsListenerChanged(cloud typeslb.TCloudListener, db corelb.TCloudListener) bool {
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

// SyncListenerOption ...
type SyncListenerOption struct {
	AccountID string                      `json:"account_id" validate:"required"`
	Region    string                      `json:"region" validate:"required"`
	LbInfos   []corelb.TCloudLoadBalancer `json:"lb_infos" validate:"required,min=1"`
}

// Validate ...
func (o *SyncListenerOption) Validate() error {
	return validator.Validate.Struct(o)
}
