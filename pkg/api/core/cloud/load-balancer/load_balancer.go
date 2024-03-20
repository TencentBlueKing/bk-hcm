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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

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

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// LoadBalancer define clb.
type LoadBalancer[Ext Extension] struct {
	BaseLoadBalancer `json:",inline"`
	Extension        *Ext `json:"extension"`
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
	Extension     *T
}

// ListenerExtension 监听器拓展
type ListenerExtension interface {
	TCloudListenerExtension
}

// BaseTCloudLbUrlRule define base tcloud lb url rule.
type BaseTCloudLbUrlRule struct {
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

// BaseClbTarget define base clb target.
type BaseClbTarget struct {
	ID                 string            `json:"id"`
	AccountID          string            `json:"account_id"`
	InstType           string            `json:"inst_type"`
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

// BaseClbTargetGroup define base clb target group.
type BaseClbTargetGroup struct {
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
