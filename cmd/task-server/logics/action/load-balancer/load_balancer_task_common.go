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

// Package actionlb ...
package actionlb

import (
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
)

// getListenerWithLb 根据监听器ID获取负载均衡及监听器详情
func getListenerWithLb(kt *kit.Kit, lblID string) (*corelb.BaseLoadBalancer,
	*corelb.BaseListener, error) {

	// 查询监听器数据
	listenerReq := &core.ListReq{
		Filter: tools.EqualExpression("id", lblID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	lblResp, err := actcli.GetDataService().Global.LoadBalancer.ListListener(kt, listenerReq)
	if err != nil {
		logs.Errorf("fail to list tcloud listener, err: %v, id: %s, rid: %s", err, lblID, kt.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.InvalidParameter, "lbl not found")
	}
	listener := lblResp.Details[0]

	// 查询负载均衡
	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", listener.LbID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	lbResp, err := actcli.GetDataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("fail to list tcloud load balancer, err: %v, id: %s, rid: %s", err, listener.LbID, kt.Rid)
		return nil, nil, err
	}
	if len(lbResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "lb not found")
	}
	lb := lbResp.Details[0]
	return &lb, &listener, nil
}

// isHealthCheckChange 七层规则不支持设置检查端口
func isHealthCheckChange(req *corelb.TCloudHealthCheckInfo, db *corelb.TCloudHealthCheckInfo, isL7 bool) bool {
	if req == nil || db == nil {
		// 请求或数据库为空，默认参数
		return false
	}
	if !assert.IsPtrInt64Equal(req.HealthSwitch, db.HealthSwitch) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.TimeOut, db.TimeOut) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.IntervalTime, db.IntervalTime) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.HealthNum, db.HealthNum) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.UnHealthNum, db.UnHealthNum) {
		return true
	}
	// 七层规则不支持设置检查端口, 这里不比较该数据
	if isL7 && !assert.IsPtrInt64Equal(req.CheckPort, db.CheckPort) {
		return true
	}
	if !assert.IsPtrStringEqual(req.ContextType, db.ContextType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.SendContext, db.SendContext) {
		return true
	}
	if !assert.IsPtrStringEqual(req.RecvContext, db.RecvContext) {
		return true
	}
	if !assert.IsPtrStringEqual(req.CheckType, db.CheckType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpVersion, db.HttpVersion) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.SourceIpType, db.SourceIpType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.ExtendedCode, db.ExtendedCode) {
		return true
	}

	if isHealthHttpCheckChange(req, db) {
		return true
	}

	return false
}

// http的健康检查
func isHealthHttpCheckChange(req *corelb.TCloudHealthCheckInfo, db *corelb.TCloudHealthCheckInfo) bool {
	if req == nil || db == nil {
		// 请求或数据库为空，默认参数
		return false
	}
	if !assert.IsPtrInt64Equal(req.HttpCode, db.HttpCode) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckPath, db.HttpCheckPath) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckDomain, db.HttpCheckDomain) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckMethod, db.HttpCheckMethod) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpVersion, db.HttpVersion) {
		return true
	}
	return false
}

// isListenerCertChange 比较两个负载均衡证书是否需要更新
func isListenerCertChange(want *corelb.TCloudCertificateInfo, db *corelb.TCloudCertificateInfo) bool {
	if want == nil {
		// 请求为空，默认参数
		return false
	}
	if db == nil {
		// 数据库为空，认为是默认参数
		return false
	}

	if !assert.IsPtrStringEqual(want.SSLMode, db.SSLMode) {
		return true
	}

	if !assert.IsPtrStringEqual(want.CaCloudID, db.CaCloudID) {
		return true
	}

	// 都有，但是数量不相等
	if len(db.CertCloudIDs) != len(want.CertCloudIDs) {
		// 数量不相等
		return true
	}
	// 要求证书按顺序相等。
	for i := range want.CertCloudIDs {
		if db.CertCloudIDs[i] != want.CertCloudIDs[i] {
			return true
		}
	}
	return false
}

// batchListListenerByIDs 根据监听器ID数组，批量获取监听器列表
func batchListListenerByIDs(kt *kit.Kit, lblIDs []string) ([]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "listener ids is required")
	}

	// 查询监听器列表
	req := &core.ListReq{
		Filter: tools.ContainersExpression("id", lblIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lblList := make([]corelb.BaseListener, 0)
	for {
		lblResp, err := actcli.GetDataService().Global.LoadBalancer.ListListener(kt, req)
		if err != nil {
			logs.Errorf("failed to list tcloud listener, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
			return nil, err
		}

		lblList = append(lblList, lblResp.Details...)
		if uint(len(lblResp.Details)) < core.DefaultMaxPageLimit {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return lblList, nil
}
