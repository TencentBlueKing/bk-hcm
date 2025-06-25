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
	apicore "hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudLoadBalancerCreateReq tcloud batch create req.
type TCloudLoadBalancerCreateReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id"`
	Region    string  `json:"region" validate:"required"`
	Name      *string `json:"name" validate:"required,max=60"`

	LoadBalancerType typelb.TCloudLoadBalancerType   `json:"load_balancer_type" validate:"required"`
	AddressIPVersion typelb.TCloudIPVersionForCreate `json:"address_ip_version" validate:"omitempty"`

	// 公网	单可用区		传递zones（单元素数组）
	// 公网	主备可用区	传递zones（单元素数组），以及backup_zones
	Zones                   []string `json:"zones" validate:"omitempty"`
	BackupZones             []string `json:"backup_zones" validate:"omitempty"`
	CloudVpcID              *string  `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           *string  `json:"cloud_subnet_id" validate:"omitempty"`
	Vip                     *string  `json:"vip" validate:"omitempty"`
	CloudEipID              *string  `json:"cloud_eip_id" validate:"omitempty"`
	VipIsp                  *string  `json:"vip_isp" validate:"omitempty"`
	InternetMaxBandwidthOut *int64   `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthPackageID      *string  `json:"bandwidth_package_id" validate:"omitempty"`
	BandwidthpkgSubType     *string  `json:"bandwidthpkg_sub_type" validate:"omitempty"`
	Egress                  *string  `json:"egress" validate:"omitempty"`

	SlaType      *string `json:"sla_type" validate:"omitempty"`
	AutoRenew    *bool   `json:"auto_renew" validate:"omitempty"`
	RequireCount *uint64 `json:"require_count" validate:"omitempty,max=20"`
	Memo         string  `json:"memo" validate:"omitempty"`

	InternetChargeType *typelb.TCloudLoadBalancerNetworkChargeType `json:"internet_charge_type" validate:"omitempty"`
	// LoadBalancerPassToTarget 安全组放通模式
	LoadBalancerPassToTarget *bool `json:"load_balancer_pass_to_target" validate:"required"`

	Tags []apicore.TagPair `json:"tags,omitempty"`
}

// Validate request.
func (req *TCloudLoadBalancerCreateReq) Validate(bizRequired bool) error {

	if bizRequired && req.BkBizID <= 0 {
		return errors.New("bk_biz_id is required")
	}

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
		// 	公网不能指定子网
		if converter.PtrToVal(req.CloudSubnetID) != "" {
			return errors.New("subnet id is not supported for load balancer type 'OPEN'")
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
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,min=1,max=5"`
}

// Validate tcloud lb security-group option.
func (opt TCloudSetLbSecurityGroupReq) Validate() error {
	if len(opt.LbID) == 0 {
		return errors.New("lb_id is required")
	}

	if len(opt.SecurityGroupIDs) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("load balancer only allows binding to %d security groups by default",
			constant.LoadBalancerBindSecurityGroupMaxLimit)
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

	// 跨域1.0 region 非空表示支持跨域
	TargetRegion *string `json:"target_region,omitempty"`
	// 跨域1.0 为0表示基础网络
	TargetCloudVpcID *string `json:"target_vpc,omitempty"`
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
	Url string `json:"url,omitempty" validate:"required"`

	TargetGroupID      string `json:"target_group_id" validate:"omitempty"`
	CloudTargetGroupID string `json:"cloud_target_group_id" validate:"omitempty"`

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

	Memo *string `json:"memo,omitempty"`
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

// TCloudRuleDeleteByIDReq 批量按ID删除规则
type TCloudRuleDeleteByIDReq struct {
	RuleIDs []string `json:"rule_ids" validate:"required,min=1"`
}

// Validate ...
func (r TCloudRuleDeleteByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// TCloudRuleDeleteByDomainReq 批量按域名删除规则
type TCloudRuleDeleteByDomainReq struct {
	Domains          []string `json:"domains" validate:"required,min=1"`
	NewDefaultDomain *string  `json:"new_default_domain"`
}

// Validate ...
func (r TCloudRuleDeleteByDomainReq) Validate() error {
	return validator.Validate.Struct(r)
}

// --------------------------[创建监听器及规则]--------------------------

// TCloudListenerCreateReq listener only create req.
type TCloudListenerCreateReq struct {
	Name          string                        `json:"name" validate:"required"`
	BkBizID       int64                         `json:"bk_biz_id" validate:"omitempty"`
	LbID          string                        `json:"lb_id" validate:"required"`
	Protocol      enumor.ProtocolType           `json:"protocol" validate:"required"`
	Port          int64                         `json:"port" validate:"required"`
	Scheduler     string                        `json:"scheduler" validate:"omitempty"`
	SessionExpire int64                         `json:"session_expire" validate:"omitempty"`
	SniSwitch     enumor.SniType                `json:"sni_switch" validate:"omitempty"`
	Certificate   *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
	HealthCheck   *corelb.TCloudHealthCheckInfo `json:"health_check,omitempty"`
	SessionType   *string                       `json:"session_type" validate:"omitempty"`
	EndPort       *int64                        `json:"end_port" validate:"omitempty,min=1"`
}

// Validate 校验创建监听器的参数
func (req *TCloudListenerCreateReq) Validate() error {

	if req.SessionExpire > 0 && (req.SessionExpire < 30 || req.SessionExpire > 3600) {
		return errors.New("session_expire must be '0' or between `30` and `3600`")
	}

	// 7层HTTPS 监听器 SNI开启，不能传入证书，SNI关闭时，需要传入证书
	if req.Protocol == enumor.HttpsProtocol {
		if req.SniSwitch == enumor.SniTypeClose && req.Certificate == nil {
			return errf.New(errf.InvalidParameter, "certificate is required for non-sni https listener")
		}
		if req.SniSwitch == enumor.SniTypeClose && converter.PtrToVal(req.Certificate.CaCloudID) == "" && len(req.
			Certificate.CertCloudIDs) == 0 {
			return errf.New(errf.InvalidParameter,
				"certificate.ca_cloud_id/certificate.cert_cloud_ids is required for non-sni https listener")
		}
		if req.SniSwitch == enumor.SniTypeOpen && req.Certificate != nil {
			return errf.New(errf.InvalidParameter, "certificate should not exists for sni https listener")
		}
	}

	return validator.Validate.Struct(req)
}

// ListenerWithRuleCreateReq listener with rule create req.
type ListenerWithRuleCreateReq struct {
	Name          string                        `json:"name" validate:"required"`
	BkBizID       int64                         `json:"bk_biz_id" validate:"omitempty"`
	LbID          string                        `json:"lb_id" validate:"required"`
	Protocol      enumor.ProtocolType           `json:"protocol" validate:"required"`
	Port          int64                         `json:"port" validate:"required"`
	Scheduler     string                        `json:"scheduler" validate:"required"`
	SessionType   string                        `json:"session_type" validate:"required"`
	SessionExpire int64                         `json:"session_expire" validate:"omitempty"`
	TargetGroupID string                        `json:"target_group_id" validate:"required"`
	Domain        string                        `json:"domain" validate:"omitempty"`
	Url           string                        `json:"url" validate:"omitempty"`
	SniSwitch     enumor.SniType                `json:"sni_switch" validate:"omitempty"`
	Certificate   *corelb.TCloudCertificateInfo `json:"certificate" validate:"omitempty"`
	EndPort       uint64                        `json:"end_port" validate:"omitempty"`
}

// Validate 校验创建监听器的参数
func (req *ListenerWithRuleCreateReq) Validate() error {
	if req.Protocol.IsLayer7Protocol() {
		if len(req.Domain) == 0 || len(req.Url) == 0 {
			return errors.New("domain and url is required")
		}
	}
	if req.SessionExpire > 0 && (req.SessionExpire < 30 || req.SessionExpire > 3600) {
		return errors.New("session_expire must be '0' or between `30` and `3600`")
	}
	return validator.Validate.Struct(req)
}

// ListenerWithRuleCreateResult ...
type ListenerWithRuleCreateResult struct {
	CloudLblID  string `json:"cloud_lbl_id"`
	CloudRuleID string `json:"cloud_rule_id"`
}

// ListenerCreateResult 监听器创建结果
type ListenerCreateResult = apicore.CloudCreateResult

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

// --------------------------[批量操作RS]--------------------------

// TCloudBatchOperateTargetReq tcloud batch operate rs req.
type TCloudBatchOperateTargetReq struct {
	TargetGroupID string                 `json:"target_group_id" validate:"required"`
	LbID          string                 `json:"lb_id" validate:"required"`
	RsList        []*cloud.TargetBaseReq `json:"targets" validate:"required,min=1,max=100,dive"`
}

// Validate RsList最大支持100个.
func (req *TCloudBatchOperateTargetReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchRegisterTCloudTargetReq 批量注册云上RS
type BatchRegisterTCloudTargetReq struct {
	CloudListenerID string            `json:"cloud_listener_id"  validate:"required"`
	CloudRuleID     string            `json:"cloud_rule_id"  validate:"omitempty"`
	TargetGroupID   string            `json:"target_group_id"  validate:"omitempty"`
	RuleType        enumor.RuleType   `json:"rule_type" validate:"required"`
	Targets         []*RegisterTarget `json:"targets"  validate:"required,min=1,dive"`
}

// Validate ...
func (r BatchRegisterTCloudTargetReq) Validate() error {
	for _, target := range r.Targets {
		err := target.Validate()
		if err != nil {
			return err
		}
	}
	return validator.Validate.Struct(r)
}

// RegisterTarget ...
type RegisterTarget struct {
	CloudInstID      string          `json:"cloud_inst_id,omitempty" validate:"omitempty"`
	TargetType       enumor.InstType `json:"inst_type,omitempty" validate:"required"`
	EniIp            string          `json:"eni_ip,omitempty" validate:"omitempty"`
	Port             int64           `json:"port" validate:"required"`
	Weight           *int64          `json:"weight" validate:"required,min=0,max=100"`
	Zone             string          `json:"zone,omitempty" validate:"omitempty"`
	InstName         string          `json:"inst_name,omitempty" validate:"omitempty"`
	PrivateIPAddress []string        `json:"private_ip_address,omitempty" validate:"omitempty"`
	PublicIPAddress  []string        `json:"public_ip_address,omitempty" validate:"omitempty"`
}

// Validate ...
func (r RegisterTarget) Validate() error {
	if r.TargetType == enumor.EniInstType && len(r.EniIp) == 0 {
		return errors.New("eni_ip not set for eni type target")
	}
	if r.TargetType == enumor.CvmInstType && len(r.CloudInstID) == 0 {
		return errors.New("cloud_inst_id not set for cvm type target")
	}
	return validator.Validate.Struct(r)
}

// BatchDeleteLoadBalancerReq ...
type BatchDeleteLoadBalancerReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required,min=1"`
}

// Validate ...
func (r *BatchDeleteLoadBalancerReq) Validate() error {
	if len(r.IDs) > constant.BatchListenerMaxLimit {
		return errors.New("batch delete limit is 20")
	}
	return validator.Validate.Struct(r)
}

// HealthCheckUpdateReq 健康检查更新接口
type HealthCheckUpdateReq struct {
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check" validate:"required"`
}

// Validate HealthCheckUpdateReq
func (h *HealthCheckUpdateReq) Validate() error {

	return validator.Validate.Struct(h)
}

// --------------------------[批量查询RS健康]--------------------------

// TCloudTargetHealthReq tcloud target health req.
type TCloudTargetHealthReq struct {
	AccountID  string   `json:"account_id" validate:"omitempty"`
	Region     string   `json:"region" validate:"omitempty"`
	CloudLbIDs []string `json:"cloud_lb_ids" validate:"required,min=1,max=20,dive"`
}

// Validate 最大支持20个.
func (req *TCloudTargetHealthReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudTargetHealthResp ...
type TCloudTargetHealthResp struct {
	Details []TCloudTargetHealthResult `json:"details"`
}

// TCloudTargetHealthResult ...
type TCloudTargetHealthResult struct {
	CloudLbID string                         `json:"cloud_lb_id"`
	Listeners []*TCloudTargetHealthLblResult `json:"listeners"`
}

// TCloudTargetHealthLblResult ...
type TCloudTargetHealthLblResult struct {
	CloudLblID   string                          `json:"cloud_lbl_id"`
	Protocol     enumor.ProtocolType             `json:"protocol"`
	ListenerName string                          `json:"listener_name"`
	HealthCheck  *corelb.TCloudHealthCheckInfo   `json:"health_check"`
	Rules        []*TCloudTargetHealthRuleResult `json:"rules"`
}

// TCloudTargetHealthRuleResult ...
type TCloudTargetHealthRuleResult struct {
	CloudRuleID string                        `json:"cloud_rule_id"`
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check"`
}

// QueryTCloudListenerTargets ...
type QueryTCloudListenerTargets struct {
	AccountID           string              `json:"account_id" validate:"required"`
	Region              string              `json:"region" validate:"required"`
	LoadBalancerCloudId string              `json:"load_balancer_cloud_id" validate:"required"`
	ListenerCloudIDs    []string            `json:"listener_cloud_ids" validate:"omitempty"`
	Protocol            enumor.ProtocolType `json:"protocol" validate:"omitempty"`
	Port                int64               `json:"port" validate:"omitempty"`
}

// Validate ...
func (t *QueryTCloudListenerTargets) Validate() error {
	return validator.Validate.Struct(t)
}

// TCloudListLoadBalancerQuotaReq tcloud list load balancer quota req.
type TCloudListLoadBalancerQuotaReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate request.
func (req *TCloudListLoadBalancerQuotaReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudCreateSnatIpReq ...
type TCloudCreateSnatIpReq struct {
	AccountID           string           `json:"account_id" validate:"required"`
	Region              string           `json:"region" validate:"required"`
	LoadBalancerCloudId string           `json:"load_balancer_cloud_id" validate:"required"`
	SnatIPs             []*corelb.SnatIp `json:"snat_ips" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *TCloudCreateSnatIpReq) Validate() error {
	return validator.Validate.Struct(r)
}

// TCloudDeleteSnatIpReq ...
type TCloudDeleteSnatIpReq struct {
	AccountID           string   `json:"account_id" validate:"required"`
	Region              string   `json:"region" validate:"required"`
	LoadBalancerCloudId string   `json:"load_balancer_cloud_id" validate:"required"`
	Ips                 []string `json:"ips" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *TCloudDeleteSnatIpReq) Validate() error {
	return validator.Validate.Struct(r)
}

// --------------------------[按负载均衡批量解绑RS]--------------------------

// TCloudBatchUnbindRsReq tcloud batch unbind rs req.
type TCloudBatchUnbindRsReq struct {
	AccountID           string                           `json:"account_id" validate:"required"`
	Region              string                           `json:"region" validate:"required"`
	Vendor              enumor.Vendor                    `json:"vendor" validate:"required"`
	LoadBalancerCloudId string                           `json:"load_balancer_cloud_id" validate:"required"`
	Details             []*cloud.ListBatchListenerResult `json:"details"`
}

// Validate validate tcloud batch unbind rs.
func (req *TCloudBatchUnbindRsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// --------------------------[按负载均衡批量调整RS权重]--------------------------

// TCloudBatchModifyRsWeightReq tcloud batch modify rs weight req.
type TCloudBatchModifyRsWeightReq struct {
	AccountID           string                           `json:"account_id" validate:"required"`
	Region              string                           `json:"region" validate:"required"`
	Vendor              enumor.Vendor                    `json:"vendor" validate:"required"`
	LoadBalancerCloudId string                           `json:"load_balancer_cloud_id" validate:"required"`
	Details             []*cloud.ListBatchListenerResult `json:"details"`
	NewRsWeight         *int64                           `json:"new_rs_weight" validate:"required,min=0,max=100"`
}

// Validate validate tcloud batch modify rs weight.
func (req *TCloudBatchModifyRsWeightReq) Validate() error {
	return validator.Validate.Struct(req)
}
