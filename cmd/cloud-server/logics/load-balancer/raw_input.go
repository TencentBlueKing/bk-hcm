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

package lblogic

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/enumor"
)

var (
	supportedProtocols = map[enumor.ProtocolType]struct{}{
		enumor.TcpProtocol:   {},
		enumor.UdpProtocol:   {},
		enumor.HttpProtocol:  {},
		enumor.HttpsProtocol: {},
	}

	supportInstTypes = map[enumor.InstType]struct{}{
		enumor.CvmInstType: {},
		enumor.EniInstType: {},
	}

	supportActions = map[string]enumor.BatchOperationActionType{
		"追加RS":           enumor.AppendRS,
		"新增监听器&绑定RS":     enumor.CreateListenerAndAppendRS,
		"新增URL&绑定RS":     enumor.CreateURLAndAppendRS,
		"新增监听器&URL&绑定RS": enumor.CreateListenerWithURLAndAppendRS,
		"修改权重":           enumor.ModifyRSWeight,
	}
)

// BindRSRawInput 解析Excel中每一行的输入数据的原始结构
type BindRSRawInput struct {
	Action       enumor.BatchOperationActionType
	ListenerName string
	Protocol     string

	IPDomainType string // IP域名类型
	VIPs         []string
	VPorts       []int
	haveEndPort  bool // 是否是端口端

	Domain  string // 域名
	URLPath string // URL路径

	RSIPs   []string
	RSPorts []int
	Weight  []int

	Scheduler      string // 均衡方式
	SessionExpired int64  // 会话保持时间，单位秒
	InstType       string // 后端类型 CVM、ENI
	HealthCheck    bool   // 是否开启健康检查

	ServerCert []string // ref: pkg/api/core/cloud/load-balancer/tcloud.go:188
	ClientCert string   // 客户端证书
}

// SplitRecord 根据VIPs和VPorts进行记录的拆分, 最终的记录应该是一个VIP对应一个VPort的形式=>一个监听器
func (l *BindRSRawInput) SplitRecord() ([]*BindRSRecord, error) {
	if len(l.VIPs) == 0 || len(l.VPorts) == 0 {
		return nil, fmt.Errorf("VIP and VPort should not be empty")
	}

	records := make([]*BindRSRecord, 0)
	if l.haveEndPort {
		for i := 0; i < len(l.VIPs); i++ {
			record := l.initRecord()
			record.VIP = l.VIPs[i]
			record.VPorts = l.VPorts
			records = append(records, record)
		}
		return records, nil
	}

	for i := 0; i < len(l.VIPs); i++ {
		for j := 0; j < len(l.VPorts); j++ {
			record := l.initRecord()
			record.VIP = l.VIPs[i]
			record.VPorts = []int{l.VPorts[j]}
			records = append(records, record)
		}
	}
	return records, nil
}

func (l *BindRSRawInput) initRecord() *BindRSRecord {
	return &BindRSRecord{
		Action:         l.Action,
		ListenerName:   l.ListenerName,
		Protocol:       enumor.ProtocolType(strings.ToUpper(l.Protocol)),
		HaveEndPort:    l.haveEndPort,
		IPDomainType:   l.IPDomainType,
		Domain:         l.Domain,
		URLPath:        l.URLPath,
		ServerCerts:    l.ServerCert,
		ClientCert:     l.ClientCert,
		Scheduler:      l.Scheduler,
		SessionExpired: l.SessionExpired,
		HealthCheck:    l.HealthCheck,
		RSIPs:          l.RSIPs,
		RSPorts:        l.RSPorts,
		Weights:        l.Weight,
		InstType:       enumor.InstType(l.InstType),
	}
}

// ModifyWeightRawInput excel中的原始输入数据结构
type ModifyWeightRawInput struct {
	Action       enumor.BatchOperationActionType
	ListenerName string
	Protocol     string

	IPDomainType string // IP域名类型
	VIPs         []string
	VPorts       []int
	haveEndPort  bool // 是否是端口端

	Domain  string // 域名
	URLPath string // URL路径

	RSIPs     []string
	RSPorts   []int
	OldWeight []int
	Weight    []int
}

// SplitRecord 根据VIPs和VPorts进行记录的拆分, 最终的记录应该是一个VIP对应一个VPort的形式=>一个监听器
func (l *ModifyWeightRawInput) SplitRecord() ([]*ModifyWeightRecord, error) {
	if len(l.VIPs) == 0 || len(l.VPorts) == 0 {
		return nil, fmt.Errorf("VIP and VPort should not be empty")
	}

	records := make([]*ModifyWeightRecord, 0)
	if l.haveEndPort {
		for i := 0; i < len(l.VIPs); i++ {
			record := l.initRecord()
			record.VIP = l.VIPs[i]
			record.VPorts = l.VPorts
			records = append(records, record)
		}
		return records, nil
	}

	for i := 0; i < len(l.VIPs); i++ {
		for j := 0; j < len(l.VPorts); j++ {
			record := l.initRecord()
			record.VIP = l.VIPs[i]
			record.VPorts = []int{l.VPorts[j]}
			records = append(records, record)
		}
	}

	return records, nil
}

func (l *ModifyWeightRawInput) initRecord() *ModifyWeightRecord {
	return &ModifyWeightRecord{
		Action:       l.Action,
		ListenerName: l.ListenerName,
		Protocol:     enumor.ProtocolType(strings.ToUpper(l.Protocol)),
		HaveEndPort:  l.haveEndPort,
		IPDomainType: l.IPDomainType,
		Domain:       l.Domain,
		URLPath:      l.URLPath,
		RSIPs:        l.RSIPs,
		RSPorts:      l.RSPorts,
		Weights:      l.Weight,
		OldWeight:    l.OldWeight,
	}
}
