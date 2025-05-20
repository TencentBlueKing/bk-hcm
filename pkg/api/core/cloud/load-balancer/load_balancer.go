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

package loadbalancer

import (
	"encoding/json"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

// SummaryBalancer define summary clb.
type SummaryBalancer struct {
	ID               string               `json:"id"`
	CloudID          string               `json:"cloud_id"`
	Name             string               `json:"name"`
	Vendor           enumor.Vendor        `json:"vendor"`
	BkBizID          int64                `json:"bk_biz_id"`
	IPVersion        enumor.IPAddressType `json:"ip_version"`
	LoadBalancerType string               `json:"lb_type"`

	Region      string   `json:"region"`
	Zones       []string `json:"zones"`
	BackupZones []string `json:"backup_zones"`

	VpcID      string `json:"vpc_id"`
	CloudVpcID string `json:"cloud_vpc_id"`

	Domain string  `json:"domain"`
	Status string  `json:"status"`
	Memo   *string `json:"memo"`

	// PrivateIPv4Addresses 内网IP
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	// PublicIPv6Addresses 公网IP
	PublicIPv4Addresses []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses []string `json:"public_ipv6_addresses"`
}

// BaseLoadBalancer define base clb.
type BaseLoadBalancer struct {
	ID               string               `json:"id"`
	CloudID          string               `json:"cloud_id"`
	Name             string               `json:"name"`
	Vendor           enumor.Vendor        `json:"vendor"`
	AccountID        string               `json:"account_id"`
	BkBizID          int64                `json:"bk_biz_id"`
	IPVersion        enumor.IPAddressType `json:"ip_version"`
	LoadBalancerType string               `json:"lb_type"`

	Region               string   `json:"region" validate:"omitempty"`
	Zones                []string `json:"zones"`
	BackupZones          []string `json:"backup_zones"`
	VpcID                string   `json:"vpc_id" validate:"omitempty"`
	CloudVpcID           string   `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID             string   `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID        string   `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []string `json:"public_ipv6_addresses"`
	Domain               string   `json:"domain"`
	Status               string   `json:"status"`
	CloudCreatedTime     string   `json:"cloud_created_time"`
	CloudStatusTime      string   `json:"cloud_status_time"`
	CloudExpiredTime     string   `json:"cloud_expired_time"`

	Tags core.TagMap `json:"tags"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// LoadBalancer define clb.
type LoadBalancer[Ext Extension] struct {
	BaseLoadBalancer `json:",inline"`
	Extension        *Ext `json:"extension"`
}

// LoadBalancerRaw define clb.
type LoadBalancerRaw struct {
	BaseLoadBalancer `json:",inline"`
	Extension        json.RawMessage `json:"extension"`
}

// LoadBalancerWithDeleteProtect define clb with load balancer delete protect
type LoadBalancerWithDeleteProtect struct {
	BaseLoadBalancer `json:",inline"`
	DeleteProtect    bool `json:"delete_protect"`
}

// GetID ...
func (lb LoadBalancer[T]) GetID() string {
	return lb.BaseLoadBalancer.ID
}

// GetCloudID ...
func (lb LoadBalancer[T]) GetCloudID() string {
	return lb.BaseLoadBalancer.CloudID
}

// Extension extension.
type Extension interface {
	TCloudClbExtension
}

// BaseListener define base listener.
type BaseListener struct {
	ID        string        `json:"id"`
	CloudID   string        `json:"cloud_id"`
	Name      string        `json:"name"`
	Vendor    enumor.Vendor `json:"vendor"`
	AccountID string        `json:"account_id"`
	BkBizID   int64         `json:"bk_biz_id"`

	LbID          string              `json:"lb_id"`
	CloudLbID     string              `json:"cloud_lb_id"`
	Protocol      enumor.ProtocolType `json:"protocol"`
	Port          int64               `json:"port"`
	DefaultDomain string              `json:"default_domain"`
	Region        string              `json:"region"`
	Zones         []string            `json:"zones"`
	// 腾讯云 CLB 的七层 HTTPS 监听器支持 SNI，即支持绑定多个证书，监听规则中的不同域名可使用不同证书。
	// SNI关闭，证书在监听器上；SNI关闭，证书在对应规则上
	SniSwitch enumor.SniType `json:"sni_switch"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// Listener 监听器带拓展
type Listener[T ListenerExtension] struct {
	*BaseListener `json:",inline"`
	Extension     *T `json:"extension"`
}

// GetID ...
func (lbl Listener[T]) GetID() string {
	return lbl.BaseListener.ID
}

// GetCloudID ...
func (lbl Listener[T]) GetCloudID() string {
	return lbl.BaseListener.CloudID
}

// ListenerExtension 监听器拓展
type ListenerExtension interface {
	TCloudListenerExtension
}

// TCloudLbUrlRule define base tcloud lb url rule.
type TCloudLbUrlRule struct {
	ID      string `json:"id"`
	CloudID string `json:"cloud_id"`
	Name    string `json:"name"`

	RuleType           enumor.RuleType `json:"rule_type"`
	LbID               string          `json:"lb_id"`
	CloudLbID          string          `json:"cloud_lb_id"`
	LblID              string          `json:"lbl_id"`
	CloudLBLID         string          `json:"cloud_lbl_id"`
	TargetGroupID      string          `json:"target_group_id"`
	CloudTargetGroupID string          `json:"cloud_target_group_id"`
	Region             string          `json:"region"`
	Domain             string          `json:"domain"`
	URL                string          `json:"url"`
	Scheduler          string          `json:"scheduler"`

	SessionType   string                 `json:"session_type"`
	SessionExpire int64                  `json:"session_expire"`
	HealthCheck   *TCloudHealthCheckInfo `json:"health_check"`
	Certificate   *TCloudCertificateInfo `json:"certificate"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// GetID ...
func (r TCloudLbUrlRule) GetID() string {
	return r.ID
}

// GetCloudID ...
func (r TCloudLbUrlRule) GetCloudID() string {
	return r.CloudID
}

// BaseLoadBalancerTarget define base load balancer target.
type BaseLoadBalancerTarget struct {
	ID                 string            `json:"id"`
	AccountID          string            `json:"account_id"`
	InstType           enumor.InstType   `json:"inst_type"`
	InstID             string            `json:"inst_id"`
	CloudInstID        string            `json:"cloud_inst_id"`
	InstName           string            `json:"inst_name"`
	TargetGroupID      string            `json:"target_group_id"`
	CloudTargetGroupID string            `json:"cloud_target_group_id"`
	Port               int64             `json:"port"`
	Weight             int64             `json:"weight"`
	PrivateIPAddress   types.StringArray `json:"private_ip_address"`
	PublicIPAddress    types.StringArray `json:"public_ip_address"`
	Zone               string            `json:"zone"`
	Memo               *string           `json:"memo"`
	*core.Revision     `json:",inline"`
}

// BaseLoadBalancerTargetGroup define base load balancer target group.
type BaseLoadBalancerTargetGroup struct {
	ID              string                 `json:"id"`
	CloudID         string                 `json:"cloud_id"`
	Name            string                 `json:"name"`
	Vendor          enumor.Vendor          `json:"vendor"`
	AccountID       string                 `json:"account_id"`
	BkBizID         int64                  `json:"bk_biz_id"`
	TargetGroupType string                 `json:"target_group_type"`
	VpcID           string                 `json:"vpc_id"`
	CloudVpcID      string                 `json:"cloud_vpc_id"`
	Protocol        enumor.ProtocolType    `json:"protocol"`
	Region          string                 `json:"region"`
	Port            int64                  `json:"port"`
	Weight          int64                  `json:"weight"`
	HealthCheck     *TCloudHealthCheckInfo `json:"health_check"`
	Memo            *string                `json:"memo"`
	*core.Revision  `json:",inline"`
}

// BaseResFlowLock define base res flow lock.
type BaseResFlowLock struct {
	ResID          string                   `json:"res_id"`
	ResType        enumor.CloudResourceType `json:"res_type"`
	Owner          string                   `json:"owner"`
	*core.Revision `json:",inline"`
}

// BaseResFlowRel define base res flow rel.
type BaseResFlowRel struct {
	ID             string                   `json:"id"`
	ResID          string                   `json:"res_id"`
	ResType        enumor.CloudResourceType `json:"res_type" `
	FlowID         string                   `json:"flow_id"`
	TaskType       enumor.TaskType          `json:"task_type"`
	Status         enumor.ResFlowStatus     `json:"status"`
	*core.Revision `json:",inline"`
}
