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

package hclb

import (
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudBatchCreateReq tcloud batch create req.
type TCloudBatchCreateReq struct {
	AccountID        string                        `json:"account_id" validate:"required"`
	Region           string                        `json:"region" validate:"required"`
	LoadBalancerType typelb.TCloudLoadBalancerType `json:"load_balancer_type" validate:"required"`
	Name             *string                       `json:"name" validate:"required,max=60"`
	// 公网	单可用区		传递zones（单元素数组）
	// 公网	主备可用区	传递zones（单元素数组），以及backup_zones
	Zones                   []string                         `json:"zones" validate:"omitempty"`
	BackupZones             []string                         `json:"backup_zones" validate:"omitempty"`
	AddressIPVersion        *typelb.TCloudIPVersionForCreate `json:"address_ip_version" validate:"required"`
	CloudVpcID              *string                          `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           *string                          `json:"cloud_subnet_id" validate:"omitempty"`
	Vip                     *string                          `json:"vip" validate:"omitempty"`
	CloudEipID              *string                          `json:"cloud_eip_id" validate:"omitempty"`
	VipIsp                  *string                          `json:"vip_isp" validate:"omitempty"`
	InternetChargeType      *string                          `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64                           `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthPackageID      *string                          `json:"bandwidth_package_id" validate:"omitempty"`
	SlaType                 *string                          `json:"sla_type" validate:"omitempty"`
	AutoRenew               *bool                            `json:"auto_renew" validate:"omitempty"`
	RequireCount            *uint64                          `json:"require_count" validate:"omitempty"`
	Memo                    string                           `json:"memo" validate:"omitempty"`
}

// Validate request.
func (req *TCloudBatchCreateReq) Validate() error {
	switch req.LoadBalancerType {
	case typelb.InternalLoadBalancerType:
		// 内网校验
		if converter.PtrToVal(req.CloudSubnetID) == "" {
			return errors.New("subnet id is required for load balancer type 'INTERNAL'")
		}
	case typelb.OpenLoadBalancerType:
		if converter.PtrToVal(req.CloudEipID) != "" {
			return errors.New("eip id only support load balancer type 'INTERNAL'")
		}
	default:
		return fmt.Errorf("unknown load balancer type: '%s'", req.LoadBalancerType)
	}

	return validator.Validate.Struct(req)
}

// BatchCreateResult ...
type BatchCreateResult struct {
	UnknownCloudIDs []string `json:"unknown_cloud_ids"`
	SuccessCloudIDs []string `json:"success_cloud_ids"`
	FailedCloudIDs  []string `json:"failed_cloud_ids"`
	FailedMessage   string   `json:"failed_message"`
}

// -------------------------- List Clb--------------------------

// TCloudListOption defines options to list tcloud clb instances.
type TCloudListOption struct {
	AccountID string           `json:"account_id" validate:"required"`
	Region    string           `json:"region" validate:"required"`
	CloudIDs  []string         `json:"cloud_ids" validate:"omitempty,max=200"`
	Page      *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud clb list option.
func (opt TCloudListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// TCloudDescribeResourcesOption ...
type TCloudDescribeResourcesOption struct {
	AccountID                             string `json:"account_id" validate:"required"`
	*typelb.TCloudDescribeResourcesOption `json:",inline" validate:"required"`
}

// Validate tcloud clb list option.
func (opt TCloudDescribeResourcesOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// --------------------------[Associate 设置负载均衡实例的安全组]--------------------------

// TCloudSetLbSecurityGroupReq defines options to set tcloud lb security-group request.
type TCloudSetLbSecurityGroupReq struct {
	LbID             string   `json:"lb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate tcloud lb security-group option.
func (opt TCloudSetLbSecurityGroupReq) Validate() error {
	if len(opt.LbID) == 0 {
		return errors.New("lb_id is required")
	}

	if len(opt.SecurityGroupIDs) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("invalid security_group_ids max value: %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return validator.Validate.Struct(opt)
}

// --------------------------[DisAssociate 设置负载均衡实例的安全组]--------------------------

// TCloudDisAssociateLbSecurityGroupReq defines options to DisAssociate tcloud lb security-group request.
type TCloudDisAssociateLbSecurityGroupReq struct {
	LbID            string `json:"lb_id" validate:"required"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
}

// Validate tcloud lb security-group option.
func (opt TCloudDisAssociateLbSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudLBUpdateReq tcloud load balancer  update request.
type TCloudLBUpdateReq struct {
	ClbID string `json:"load_balancer_id" validate:"omitempty"`

	Name *string `json:"name" validate:"omitempty"`
	Memo *string `json:"memo" validate:"omitempty"`

	// 网络计费相关参数
	InternetChargeType      *string `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64  `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthpkgSubType     *string `json:"bandwidthpkg_sub_type" validate:"omitempty"`

	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组。
	LoadBalancerPassToTarget *bool `json:"load_balancer_pass_to_target" validate:"omitempty"`
	SnatPro                  *bool `json:"snat_pro" validate:"omitempty"`
	DeleteProtect            *bool `json:"delete_protect" validate:"omitempty"`
	ModifyClassicDomain      *bool `json:"modify_classic_domain" validate:"omitempty"`
}

// Validate tcloud security group update request.
func (req *TCloudLBUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRuleBatchCreateReq tcloud lb url rule batch create req.
type TCloudRuleBatchCreateReq struct {
	Rules []TCloudRuleCreate `json:"rules" validate:"min=1"`
}

// Validate request.
func (req *TCloudRuleBatchCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRuleCreate 腾讯云url规则创建
type TCloudRuleCreate struct {
	Url               string   `json:"url,omitempty" validate:"required"`
	Domains           []string `json:"domains,omitempty"`
	SessionExpireTime *int64   `json:"session_expire_time,omitempty"`
	Scheduler         *string  `json:"scheduler,omitempty"`
	ForwardType       *string  `json:"forward_type,omitempty"`
	DefaultServer     *bool    `json:"default_server,omitempty"`
	Http2             *bool    `json:"http2,omitempty"`
	TargetType        *string  `json:"target_type,omitempty"`
	Quic              *bool    `json:"quic,omitempty"`
	TrpcFunc          *string  `json:"trpc_func,omitempty"`
	TrpcCallee        *string  `json:"trpc_callee,omitempty"`

	HealthCheck  *corelb.TCloudHealthCheckInfo `json:"health_check,omitempty"`
	Certificates *corelb.TCloudCertificateInfo `json:"certificates,omitempty"`
}

// Validate request.
func (req *TCloudRuleCreate) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRuleUpdateReq 腾讯云url规则更新
type TCloudRuleUpdateReq struct {
	Url               *string `json:"url,omitempty"`
	SessionExpireTime *int64  `json:"session_expire_time,omitempty"`
	Scheduler         *string `json:"scheduler,omitempty"`
	ForwardType       *string `json:"forward_type,omitempty"`
	DefaultServer     *bool   `json:"default_server,omitempty"`
	Http2             *bool   `json:"http2,omitempty"`
	TargetType        *string `json:"target_type,omitempty"`
	Quic              *bool   `json:"quic,omitempty"`
	TrpcFunc          *string `json:"trpc_func,omitempty"`
	TrpcCallee        *string `json:"trpc_callee,omitempty"`

	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check,omitempty"`
}

// Validate request.
func (req *TCloudRuleUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudBatchDeleteRuleReq 批量删除规则,支持按id删除 或 按域名删除,id删除优先级更高
type TCloudBatchDeleteRuleReq struct {
	// 需要删除的规则id
	RuleIDs          []string `json:"rule_ids"`
	Domain           *string  `json:"domain"`
	NewDefaultDomain *string  `json:"new_default_domain"`
}

// Validate ...
func (r TCloudBatchDeleteRuleReq) Validate() error {
	if len(r.RuleIDs) == 0 && len(converter.PtrToVal(r.Domain)) == 0 {
		return errors.New("both rule_ids and domain are empty")
	}
	return validator.Validate.Struct(r)
}

// --------------------------[创建监听器及规则]--------------------------

// ListenerWithRuleCreateReq listener with rule create req.
type ListenerWithRuleCreateReq struct {
	Name          string                        `json:"name" validate:"required"`
	BkBizID       int64                         `json:"bk_biz_id" validate:"omitempty"`
	LbID          string                        `json:"lb_id" validate:"omitempty"`
	Protocol      enumor.ProtocolType           `json:"protocol" validate:"required"`
	Port          int64                         `json:"port" validate:"required"`
	Scheduler     string                        `json:"scheduler" validate:"required"`
	SessionType   string                        `json:"session_type" validate:"required"`
	SessionExpire int64                         `json:"session_expire" validate:"required"`
	TargetGroupID string                        `json:"target_group_id" validate:"required"`
	Domain        string                        `json:"domain" validate:"omitempty"`
	Url           string                        `json:"url" validate:"omitempty"`
	SniSwitch     enumor.SniType                `json:"sni_switch" validate:"omitempty"`
	Certificate   *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
}

// Validate 校验创建监听器的参数
func (req *ListenerWithRuleCreateReq) Validate() error {
	if req.Protocol.IsLayer7Protocol() {
		if len(req.Domain) == 0 || len(req.Url) == 0 {
			return errors.New("domain and url is required")
		}
	}
	if req.SessionExpire > 0 && req.SessionExpire < constant.ListenerMinSessionExpire {
		return fmt.Errorf("invalid session_expire min value: %d", constant.ListenerMinSessionExpire)
	}
	return validator.Validate.Struct(req)
}

// --------------------------[更新监听器]--------------------------

// ListenerWithRuleUpdateReq listener update req.
type ListenerWithRuleUpdateReq struct {
	Name      string                          `json:"name" validate:"required"`
	BkBizID   int64                           `json:"bk_biz_id" validate:"omitempty"`
	SniSwitch enumor.SniType                  `json:"sni_switch" validate:"omitempty"`
	Extension *corelb.TCloudListenerExtension `json:"extension"`
}

// Validate 校验更新监听器的参数
func (req *ListenerWithRuleUpdateReq) Validate() error {
	if err := req.SniSwitch.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(req)
}

// --------------------------[更新域名属性]--------------------------

// DomainAttrUpdateReq domain attr update req.
type DomainAttrUpdateReq struct {
	Domain    string `json:"domain" validate:"required"`
	NewDomain string `json:"new_domain" validate:"omitempty"`
	// 域名相关的证书信息，注意，仅对启用SNI的监听器适用。
	Certificate *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
	// 是否开启Http2，注意，只有HTTPS域名才能开启Http2。
	Http2 bool `json:"http2" validate:"omitempty"`
	// 是否设为默认域名，注意，一个监听器下只能设置一个默认域名。
	DefaultServer *bool `json:"default_server" validate:"omitempty"`
	// 是否开启Quic，注意，只有HTTPS域名才能开启Quic
	Quic bool `json:"quic" validate:"omitempty"`
}

// Validate 校验更新域名的参数
func (req *DomainAttrUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
