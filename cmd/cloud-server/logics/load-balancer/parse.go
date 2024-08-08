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
	"strconv"
	"strings"

	"hcm/pkg/criteria/enumor"
)

const (
	bindRSTableLen         = 17
	modifyWeightRSTableLen = 12
)

// ParseBindRSRawInput parse excel row data to BindRSRawInput struct
func ParseBindRSRawInput(row []string) (*BindRSRawInput, error) {
	if len(row) < bindRSTableLen {
		return nil, fmt.Errorf("excel cell number should be %d, but got %d", bindRSTableLen, len(row))
	}

	raw := new(BindRSRawInput)
	var err error
	raw.Action, err = parseAction(row[0])
	if err != nil {
		return nil, fmt.Errorf("Error parsing Action: %v\n", err)
	}
	raw.ListenerName = row[1]
	raw.Protocol = row[2]
	raw.IPDomainType, err = parseIPDomainType(row[3])
	if err != nil {
		return nil, fmt.Errorf("Error parsing IPDomainType: %v\n", err)
	}
	raw.VIPs = parseIPs(row[4])
	raw.VPorts, err = parsePorts(row[5], ";")
	if err != nil {
		return nil, fmt.Errorf("Error parsing VPorts: %v\n", err)
	}
	if strings.HasPrefix(row[5], "[") && strings.HasSuffix(row[5], "]") {
		raw.haveEndPort = true
	}
	raw.Domain = row[6]
	raw.URLPath = row[7]

	raw.RSIPs = parseIPs(row[8])

	raw.RSPorts, err = parsePorts(row[9], "\n")
	if err != nil {
		return nil, fmt.Errorf("Error parsing RSPORTs: %v\n", err)
	}

	raw.Weight, err = parseIntSlice(row[10], "\n")
	if err != nil {
		return nil, fmt.Errorf("Error parsing NewWeight: %v\n", err)

	}
	raw.Scheduler = parseScheduler(row[11])
	raw.SessionExpired, err = strconv.ParseInt(row[12], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Sticky: %v\n", err)
	}
	raw.InstType = row[13]
	raw.ServerCert = parseServerCert(row[14])
	raw.ClientCert = row[15]
	raw.HealthCheck, err = parseHealthCheck(row[16])
	if err != nil {
		return nil, fmt.Errorf("Error parsing HealthCheck: %v\n", err)
	}

	return raw, nil
}

// ParseModifyWeightRawInput parse excel row data to ModifyWeightRawInput struct
func ParseModifyWeightRawInput(row []string) (*ModifyWeightRawInput, error) {
	if len(row) < modifyWeightRSTableLen {
		return nil, fmt.Errorf("excel cell number should be %d, but got %d", bindRSTableLen, len(row))
	}

	raw := new(ModifyWeightRawInput)
	var err error
	raw.Action, err = parseAction(row[0])
	if err != nil {
		return nil, fmt.Errorf("Error parsing Action: %v\n", err)
	}
	raw.ListenerName = row[1]
	raw.Protocol = row[2]
	raw.IPDomainType, err = parseIPDomainType(row[3])
	if err != nil {
		return nil, fmt.Errorf("Error parsing IPDomainType: %v\n", err)
	}
	raw.VIPs = parseIPs(row[4])
	raw.VPorts, err = parsePorts(row[5], ";")
	if err != nil {
		return nil, fmt.Errorf("Error parsing VPorts: %v\n", err)
	}
	if strings.HasPrefix(row[5], "[") && strings.HasSuffix(row[5], "]") {
		raw.haveEndPort = true
	}
	raw.Domain = row[6]
	raw.URLPath = row[7]

	raw.RSIPs = parseIPs(row[8])

	raw.RSPorts, err = parsePorts(row[9], "\n")
	if err != nil {
		return nil, fmt.Errorf("Error parsing RSPORTs: %v\n", err)
	}

	raw.OldWeight, err = parseIntSlice(row[10], "\n")
	if err != nil {
		return nil, fmt.Errorf("Error parsing Old NewWeight: %v\n", err)
	}
	raw.Weight, err = parseIntSlice(row[11], "\n")
	if err != nil {
		return nil, fmt.Errorf("Error parsing NewWeight: %v\n", err)
	}
	return raw, nil
}

func parseAction(actionStr string) (enumor.BatchOperationActionType, error) {
	action, ok := supportActions[actionStr]
	if !ok {
		return "", fmt.Errorf("unsupported action: %s", actionStr)
	}
	return action, nil
}

// parseIPs 解析使用回车分隔的IP地址字符串
func parseIPs(ipStr string) []string {
	split := strings.Split(ipStr, "\n")
	for i := range split {
		split[i] = strings.TrimSpace(split[i])
		split[i] = strings.Trim(split[i], "\t")
	}
	return split
}

// parsePorts 解析端口字符串，支持换行符分隔和端口范围格式
func parsePorts(portStr, sep string) ([]int, error) {
	if strings.HasPrefix(portStr, "[") && strings.HasSuffix(portStr, "]") {
		// 端口段配置
		portsStr := portStr[1 : len(portStr)-1] // 移除括号
		portsStr = strings.TrimSpace(portsStr)
		// 分割端口范围
		ports := strings.Split(portsStr, ",")
		if len(ports) != 2 {
			return nil, fmt.Errorf("invalid port range format: %s", portStr)
		}
		startPort, err := strconv.Atoi(strings.TrimSpace(ports[0]))
		if err != nil {
			return nil, err
		}
		endPort, err := strconv.Atoi(strings.TrimSpace(ports[1]))
		if err != nil {
			return nil, err
		}

		if startPort > endPort {
			return nil, fmt.Errorf("start port bigger than end port: %s", portStr)
		}
		return []int{startPort, endPort}, nil
	} else {
		// 换行符分隔的多个端口
		return parseIntSlice(portStr, sep)
	}
}

// parseIntSlice 用于将分隔符分隔的字符串转换为整数切片
func parseIntSlice(s, sep string) ([]int, error) {
	parts := strings.Split(s, sep)
	ints := make([]int, 0, len(parts))
	for _, part := range parts {
		port, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		ints = append(ints, port)
	}
	return ints, nil
}

// parseHealthCheck 解析健康检查字段，将"是"转换为true，"否"转换为false
func parseHealthCheck(healthCheckStr string) (bool, error) {
	switch strings.TrimSpace(healthCheckStr) {
	case "是":
		return true, nil
	case "否":
		return false, nil
	default:
		return false, fmt.Errorf("invalid value for health check: %s, should be '是' or '否'", healthCheckStr)
	}
}

func parseServerCert(certStr string) []string {
	tmp := strings.Split(certStr, ",")
	// ignore empty string
	result := make([]string, 0, len(tmp))
	for _, cert := range tmp {
		if len(cert) > 0 {
			result = append(result, cert)
		}
	}
	return result
}

func parseScheduler(schedulerStr string) string {
	switch schedulerStr {
	case "按权重轮询":
		return "WRR"
	case "最小连接数":
		return "LEAST_CONN"
	case "IP Hash":
		return "IP_HASH"
	default:
		return schedulerStr
	}
}

func parseIPDomainType(ipDomainTypeStr string) (string, error) {
	t, ok := supportIPDomainType[ipDomainTypeStr]
	if !ok {
		return "", fmt.Errorf("unsupported IP domain type: %s", ipDomainTypeStr)
	}
	return t, nil
}
