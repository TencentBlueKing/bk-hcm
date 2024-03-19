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
	ID        string        `json:"id"`
	CloudID   string        `json:"cloud_id"`
	Name      string        `json:"name"`
	Vendor    enumor.Vendor `json:"vendor"`
	AccountID string        `json:"account_id"`
	BkBizID   int64         `json:"bk_biz_id"`

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

	LbID          string   `json:"lb_id"`
	CloudLbID     string   `json:"cloud_lb_id"`
	Protocol      string   `json:"protocol"`
	Port          int64    `json:"port"`
	DefaultDomain string   `json:"default_domain"`
	Zones         []string `json:"zones"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// BaseTCloudLbUrlRule define base tcloud lb url rule.
type BaseTCloudLbUrlRule struct {
	ID      string `json:"id"`
	CloudID string `json:"cloud_id"`
	Name    string `json:"name"`

	RuleType           enumor.RuleType        `json:"rule_type"`
	LbID               string                 `json:"lb_id"`
	CloudLbID          string                 `json:"cloud_lb_id"`
	LblID              string                 `json:"lbl_id"`
	CloudLBLID         string                 `json:"cloud_lbl_id"`
	TargetGroupID      string                 `json:"target_group_id"`
	CloudTargetGroupID string                 `json:"cloud_target_group_id"`
	Domain             string                 `json:"domain"`
	URL                string                 `json:"url"`
	Scheduler          string                 `json:"scheduler"`
	SniSwitch          int64                  `json:"sni_switch"`
	SessionType        string                 `json:"session_type"`
	SessionExpire      int64                  `json:"session_expire"`
	HealthCheck        *TCloudHealthCheckInfo `json:"health_check"`
	Certificate        *TCloudCertificateInfo `json:"certificate"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// TCloudHealthCheckInfo define health check.
type TCloudHealthCheckInfo struct {
	// 是否开启健康检查：1（开启）、0（关闭）
	HealthSwitch *int64 `json:"health_switch,omitempty"`
	// 健康检查的响应超时时间（仅适用于四层监听器），可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间
	TimeOut *int64 `json:"time_out,omitempty"`
	// 健康检查探测间隔时间，默认值：5，IPv4 CLB实例的取值范围为：2-300，IPv6 CLB 实例的取值范围为：5-300。单位：秒
	// 说明：部分老旧 IPv4 CLB实例的取值范围为：5-300
	IntervalTime *int64 `json:"interval_time,omitempty"`
	// 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次
	HealthNum *int64 `json:"health_num,omitempty"`
	// 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
	UnHealthNum *int64 `json:"un_health_num,omitempty"`
	// 健康检查状态码（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。可选值：1~31，默认 31。
	// 1 表示探测后返回值 1xx 代表健康，2 表示返回 2xx 代表健康，4 表示返回 3xx 代表健康，8 表示返回 4xx 代表健康，
	// 16 表示返回 5xx 代表健康。若希望多种返回码都可代表健康，则将相应的值相加。
	HttpCode *int64 `json:"http_code"`
	// 自定义探测相关参数。健康检查端口，默认为后端服务的端口，除非您希望指定特定端口，否则建议留空。（仅适用于TCP/UDP监听器）
	CheckPort *int64 `json:"check_port,omitempty"`
	// 健康检查使用的协议。取值 TCP | HTTP | HTTPS | GRPC | PING | CUSTOM，UDP监听器支持PING/CUSTOM，
	// TCP监听器支持TCP/HTTP/CUSTOM，TCP_SSL/QUIC监听器支持TCP/HTTP，HTTP规则支持HTTP/GRPC，HTTPS规则支持HTTP/HTTPS/GRPC
	CheckType *string `json:"check_type,omitempty"`
	// HTTP版本。健康检查协议CheckType的值取HTTP时，必传此字段，代表后端服务的HTTP版本：HTTP/1.0、HTTP/1.1；（仅适用于TCP监听器）
	HttpVersion *string `json:"http_version,omitempty"`
	// 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）
	HttpCheckPath *string `json:"http_check_path,omitempty"`
	// 健康检查域名（仅适用于HTTP/HTTPS监听器和TCP监听器的HTTP健康检查方式。针对TCP监听器，当使用HTTP健康检查方式时，该参数为必填项）
	HttpCheckDomain *string `json:"http_check_domain,omitempty"`
	// 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET
	HttpCheckMethod *string `json:"http_check_method,omitempty"`
	// 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP），默认值：0
	SourceIpType *int64 `json:"source_ip_type,omitempty"`
	// 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查的输入格式，可取值：HEX或TEXT；
	// 取值为HEX时，SendContext和RecvContext的字符只能在0123456789ABCDEF中选取且长度必须是偶数位。（仅适用于TCP/UDP监听器）
	ContextType *string `json:"context_type"`
}

// TCloudCertificateInfo define certificate.
type TCloudCertificateInfo struct {
	// 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证
	SSLMode *string `json:"ssl_mode,omitempty"`
	// 服务端证书的 ID，如果不填写此项则必须上传证书，包括 CertContent，CertKey，CertName
	CertId *string `json:"cert_id,omitempty"`
	// 客户端证书的 ID，当监听器采用双向认证，即 SSLMode=MUTUAL 时，如果不填写此项则必须上传客户端证书，包括 CertCaContent，CertCaName
	CertCaId *string `json:"cert_ca_id,omitempty"`
	// 上传服务端证书的名称，如果没有 CertId，则此项必传。
	CertName *string `json:"cert_name"`
	// 上传服务端证书的 key，如果没有 CertId，则此项必传。
	CertKey *string `json:"cert_key"`
	// 上传服务端证书的内容，如果没有 CertId，则此项必传。
	CertContent *string `json:"cert_content"`
	// 上传客户端 CA 证书的名称，如果 SSLMode=mutual，如果没有 CertCaId，则此项必传。
	CertCaName *string `json:"cert_ca_name"`
	// 上传客户端证书的内容，如果 SSLMode=mutual，如果没有 CertCaId，则此项必传。
	CertCaContent *string   `json:"cert_ca_content"`
	ExtCertIds    []*string `json:"ext_cert_ids,omitempty"`
}

// MultiCertInfo 多证书结构体
type MultiCertInfo struct {
	// 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证
	SSLMode *string `json:"ssl_mode"`
	// 监听器或规则证书列表，单双向认证，多本服务端证书算法类型不能重复;若SSLMode为双向认证，证书列表必须包含一本ca证书。
	CertList []*TCloudCertificateInfo `json:"cert_list"`
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
	Protocol        string                 `json:"protocol"`
	Region          string                 `json:"region"`
	Port            int64                  `json:"port"`
	Weight          int64                  `json:"weight"`
	HealthCheck     *TCloudHealthCheckInfo `json:"health_check"`
	Memo            *string                `json:"memo"`
	*core.Revision  `json:",inline"`
}
