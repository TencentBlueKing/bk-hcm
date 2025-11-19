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

// Package cc 自研云相关配置放在这个文件中，避免冲突
package cc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

const (
	// WoaServerName is woa server's name
	WoaServerName Name = "woa-server"
)

// WoaServer return woa server Setting.
func WoaServer() WoaServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty task server setting")
		return WoaServerSetting{}
	}

	s, ok := rt.settings.(*WoaServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get woa server setting", ServiceName())
		return WoaServerSetting{}
	}

	return *s
}

// MongoDB mongodb config
type MongoDB struct {
	Host                 string `yaml:"host"`
	Port                 string `yaml:"port"`
	Usr                  string `yaml:"usr"`
	Pwd                  string `yaml:"pwd"`
	Database             string `yaml:"database"`
	MaxOpenConns         uint64 `yaml:"maxOpenConns"`
	MaxIdleConns         uint64 `yaml:"maxIdleConns"`
	Mechanism            string `yaml:"mechanism"`
	RsName               string `yaml:"rsName"`
	SocketTimeoutSeconds int    `yaml:"socketTimeoutSeconds"`
}

// validate mongodb.
func (m MongoDB) validate() error {
	if len(m.Host) == 0 {
		return errors.New("mongodb host is not set")
	}
	if len(m.Usr) == 0 {
		return errors.New("mongodb usr is not set")
	}
	if len(m.Pwd) == 0 {
		return errors.New("mongodb pwd is not set")
	}
	if len(m.Database) == 0 {
		return errors.New("mongodb database is not set")
	}
	if len(m.RsName) == 0 {
		return errors.New("mongodb rsName is not set")
	}

	return nil
}

// Redis config
type Redis struct {
	Host         string `yaml:"host"`
	Pwd          string `yaml:"pwd"`
	SentinelPwd  string `yaml:"sentinelPwd"`
	Database     string `yaml:"database"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MasterName   string `yaml:"masterName"`
}

// validate redis.
func (r Redis) validate() error {
	if len(r.Host) == 0 {
		return errors.New("redis host is not set")
	}
	if len(r.Pwd) == 0 {
		return errors.New("redis pwd is not set")
	}
	if len(r.Database) == 0 {
		return errors.New("redis database is not set")
	}

	return nil
}

// ClientConfig third-party api client config set
type ClientConfig struct {
	CvmOpt        CVMCliConf `yaml:"cvm"`
	TjjOpt        TjjCli     `yaml:"tjj"`
	XshipOpt      XshipCli   `yaml:"xship"`
	TCloudOpt     TCloudCli  `yaml:"tencentcloud"`
	DvmOpt        DVMCli     `yaml:"dvm"`
	ErpOpt        ErpCli     `yaml:"erp"`
	TmpOpt        TmpCli     `yaml:"tmp"`
	Xray          XrayCli    `yaml:"xray"`
	Tcaplus       TcaplusCli `yaml:"tcaplus"`
	TGW           TGWCli     `yaml:"tgw"`
	L5            L5Cli      `yaml:"l5"`
	Safety        SafetyCli  `yaml:"safety"`
	BkChat        BkChatCli  `yaml:"bkchat"`
	Sops          SopsCli    `yaml:"sops"`
	ITSM          ApiGateway `yaml:"itsm"`
	Ngate         NgateCli   `yaml:"ngate"`
	CaiChe        CaiCheCli  `yaml:"caiche"`
	BkDbm         ApiGateway `yaml:"bkdbm"`
	BkBotApproval ApiGateway `yaml:"bkbotapproval"`
}

func (c ClientConfig) validate() error {
	if err := c.CvmOpt.validate(); err != nil {
		return err
	}

	if err := c.TjjOpt.validate(); err != nil {
		return err
	}

	if err := c.XshipOpt.validate(); err != nil {
		return err
	}

	if err := c.TCloudOpt.validate(); err != nil {
		return err
	}

	if err := c.DvmOpt.validate(); err != nil {
		return err
	}

	if err := c.ErpOpt.validate(); err != nil {
		return err
	}

	if err := c.TmpOpt.validate(); err != nil {
		return err
	}

	if err := c.Xray.validate(); err != nil {
		return err
	}

	if err := c.Tcaplus.validate(); err != nil {
		return err
	}

	if err := c.TGW.validate(); err != nil {
		return err
	}

	if err := c.L5.validate(); err != nil {
		return err
	}

	if err := c.Safety.validate(); err != nil {
		return err
	}

	if err := c.BkChat.validate(); err != nil {
		return err
	}

	if err := c.Sops.validate(); err != nil {
		return err
	}

	if err := c.CaiChe.Validate(); err != nil {
		return err
	}

	if err := c.BkDbm.validate(); err != nil {
		return err
	}

	if err := c.BkBotApproval.validate(); err != nil {
		return err
	}

	return nil
}

// CVMCliConf yunti client config
type CVMCliConf struct {
	CvmApiAddr        string `yaml:"host"`
	CvmOldApiAddr     string `yaml:"old_host"`
	CvmLaunchPassword string `yaml:"launch_password"`
}

func (c CVMCliConf) validate() error {
	if len(c.CvmApiAddr) == 0 {
		return errors.New("cvm.host is not set")
	}

	if len(c.CvmOldApiAddr) == 0 {
		return errors.New("cvm.old_host is not set")
	}

	if len(c.CvmLaunchPassword) == 0 {
		return errors.New("cvm.launch_password is not set")
	}

	return nil
}

// TjjCli tjj client options
type TjjCli struct {
	// tjj api address
	TjjApiAddr string `yaml:"host"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Operator   string `yaml:"operator"`
}

func (t TjjCli) validate() error {
	if len(t.TjjApiAddr) == 0 {
		return errors.New("tjj.host is not set")
	}

	if len(t.SecretID) == 0 {
		return errors.New("tjj.secret_id is not set")
	}

	if len(t.SecretKey) == 0 {
		return errors.New("tjj.secret_key is not set")
	}

	if len(t.Operator) == 0 {
		return errors.New("tjj.operator is not set")
	}

	return nil
}

// XshipCli xship client options
type XshipCli struct {
	// Xship api address
	XshipApiAddr string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	SecretKey    string `yaml:"secret_key"`
}

func (x XshipCli) validate() error {
	if len(x.XshipApiAddr) == 0 {
		return errors.New("xship.host is not set")
	}

	if len(x.ClientID) == 0 {
		return errors.New("xship.client_id is not set")
	}

	if len(x.SecretKey) == 0 {
		return errors.New("xship.secret_key is not set")
	}

	return nil
}

// TCloudCli  tencent cloud client options
type TCloudCli struct {
	Endpoints  TCloudEndpoints  `yaml:"endpoints"`
	Credential TCloudCredential `yaml:"credential"`
}

func (t TCloudCli) validate() error {
	if err := t.Endpoints.validate(); err != nil {
		return err
	}

	if err := t.Credential.validate(); err != nil {
		return err
	}

	return nil
}

// TCloudEndpoints tencent cloud endpoints
type TCloudEndpoints struct {
	Cvm string `yaml:"cvm"`
	Vpc string `yaml:"vpc"`
	Clb string `yaml:"clb"`
}

func (e TCloudEndpoints) validate() error {
	if len(e.Cvm) == 0 {
		return errors.New("tencentcloud.endpoints.cvm is not set")
	}

	if len(e.Vpc) == 0 {
		return errors.New("tencentcloud.endpoints.vpc is not set")
	}

	if len(e.Clb) == 0 {
		return errors.New("tencentcloud.endpoints.clb is not set")
	}

	return nil
}

// TCloudCredential tencent cloud credential
type TCloudCredential struct {
	ID  string `yaml:"id"`
	Key string `yaml:"key"`
}

func (e TCloudCredential) validate() error {
	if len(e.ID) == 0 {
		return errors.New("tencentcloud.credential.id is not set")
	}

	if len(e.Key) == 0 {
		return errors.New("tencentcloud.credential.key is not set")
	}

	return nil
}

// ErpCli erp client options
type ErpCli struct {
	// erp api address
	ErpApiAddr string `yaml:"host"`
}

func (c ErpCli) validate() error {
	if len(c.ErpApiAddr) == 0 {
		return errors.New("erp.host is not set")
	}

	return nil
}

// TmpCli tmp client options
type TmpCli struct {
	// tmp api address
	TMPApiAddr string `yaml:"host"`
}

func (c TmpCli) validate() error {
	if len(c.TMPApiAddr) == 0 {
		return errors.New("tmp.host is not set")
	}

	return nil
}

// XrayCli xray client options
type XrayCli struct {
	// xray api address
	XrayApiAddr string `yaml:"host"`
	ClientID    string `yaml:"client_id"`
	SecretKey   string `yaml:"secret_key"`
}

func (x XrayCli) validate() error {
	if len(x.XrayApiAddr) == 0 {
		return errors.New("xray.host is not set")
	}

	if len(x.ClientID) == 0 {
		return errors.New("xray.client_id is not set")
	}

	if len(x.SecretKey) == 0 {
		return errors.New("xray.secret_key is not set")
	}

	return nil
}

// TcaplusCli tcaplus client options
type TcaplusCli struct {
	// tcaplus api address
	TcaplusApiAddr string `yaml:"host"`
}

func (c TcaplusCli) validate() error {
	if len(c.TcaplusApiAddr) == 0 {
		return errors.New("tcaplus.host is not set")
	}

	return nil
}

// TGWCli tgw client options
type TGWCli struct {
	// tgw api address
	TgwApiAddr string `yaml:"host"`
}

func (c TGWCli) validate() error {
	if len(c.TgwApiAddr) == 0 {
		return errors.New("tgw.host is not set")
	}

	return nil
}

// L5Cli l5 client options
type L5Cli struct {
	// l5 api address
	L5ApiAddr string `yaml:"host"`
}

func (c L5Cli) validate() error {
	if len(c.L5ApiAddr) == 0 {
		return errors.New("l5.host is not set")
	}

	return nil
}

// SafetyCli Safety client options
type SafetyCli struct {
	// safety api address
	SafetyApiAddr string `yaml:"host"`
}

func (c SafetyCli) validate() error {
	if len(c.SafetyApiAddr) == 0 {
		return errors.New("safety.host is not set")
	}

	return nil
}

// DVMCli dvm client options
type DVMCli struct {
	DvmApiAddr string `yaml:"host"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Operator   string `yaml:"operator"`
}

func (c DVMCli) validate() error {
	if len(c.DvmApiAddr) == 0 {
		return errors.New("dvm.host is not set")
	}

	if len(c.SecretID) == 0 {
		return errors.New("dvm.secret_id is not set")
	}

	if len(c.SecretKey) == 0 {
		return errors.New("dvm.secret_key is not set")
	}

	if len(c.Operator) == 0 {
		return errors.New("dvm.operator is not set")
	}

	return nil
}

// BkChatCli bkchat client options
type BkChatCli struct {
	BkChatApiAddr string `yaml:"host"`
	NoticeFmt     string `yaml:"notify_format"`
}

func (c BkChatCli) validate() error {
	if len(c.BkChatApiAddr) == 0 {
		return errors.New("bkchat.host is not set")
	}

	return nil
}

// SopsCli sops client options
type SopsCli struct {
	SopsApiAddr string `yaml:"host"`
	AppCode     string `yaml:"app_code"`
	AppSecret   string `yaml:"app_secret"`
	Operator    string `yaml:"operator"`
	DevnetIP    string `yaml:"devnet_download_ip"`
}

func (c SopsCli) validate() error {
	if len(c.SopsApiAddr) == 0 {
		return errors.New("sops.host is not set")
	}

	if len(c.AppCode) == 0 {
		return errors.New("sops.app_code is not set")
	}

	if len(c.AppSecret) == 0 {
		return errors.New("sops.app_secret is not set")
	}

	if len(c.DevnetIP) == 0 {
		return errors.New("sops.devnet_download_ip is not set")
	}

	return nil
}

// ItsmFlow defines the itsm flow related runtime.
type ItsmFlow struct {
	// ServiceName is the itsm service name.
	ServiceName string `yaml:"serviceName"`
	// ServiceID is the itsm service id.
	ServiceID int64 `yaml:"serviceID"`
	// StateNodes is the itsm state nodes.
	StateNodes []StateNode `yaml:"stateNodes"`
	// RedirectUrlTemplate is the itsm service redirect url template.
	RedirectUrlTemplate string `yaml:"redirectUrlTemplate"`
}

// validate ItsmFlow runtime.
func (i ItsmFlow) validate() error {
	if i.ServiceID == 0 {
		return errors.New("itsm service id is not set")
	}

	for _, stateNode := range i.StateNodes {
		if err := stateNode.validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResourceDissolve resource dissolve config
type ResourceDissolve struct {
	OriginDate               string   `yaml:"originDate"`
	ProjectNames             []string `yaml:"projectNames"`
	ProjectIDs               []int    `yaml:"projectIDs"`
	ListExcludedProjectNames []string `yaml:"listExcludedProjectNames"`
	SyncDissolveHost         bool     `yaml:"syncDissolveHost"`
}

func (r ResourceDissolve) validate() error {
	if len(r.OriginDate) == 0 {
		return errors.New("resourceDissolve.originDate is not set")
	}

	if len(r.ProjectNames) == 0 {
		return errors.New("resourceDissolve.projectNames is not set")
	}

	if len(r.ProjectIDs) == 0 {
		return errors.New("resourceDissolve.projectIDs is not set")
	}

	return nil
}

// StateNode defines the itsm state node related runtime.
type StateNode struct {
	ID          int64  `yaml:"id"`
	NodeName    string `yaml:"nodeName"`
	Approver    string `json:"approver"`
	ApprovalKey string `yaml:"approvalKey"`
	RemarkKey   string `yaml:"remarkKey"`
}

// validate StateNode runtime.
func (i StateNode) validate() error {
	if i.ID == 0 {
		return errors.New("state node id is not set")
	}

	if len(i.NodeName) == 0 {
		return errors.New("state node name is not set")
	}

	if len(i.ApprovalKey) == 0 {
		return errors.New("state node approval key is not set")
	}

	if len(i.RemarkKey) == 0 {
		return errors.New("state node remark key is not set")
	}
	return nil
}

// ObjectStore object store config
type ObjectStore struct {
	Type              string `yaml:"type"`
	ObjectStoreTCloud `yaml:",inline"`
}

// ObjectStoreTCloud tencent cloud cos config
type ObjectStoreTCloud struct {
	UIN             string `yaml:"uin"`
	COSPrefix       string `yaml:"prefix"`
	COSSecretID     string `yaml:"secretId"`
	COSSecretKey    string `yaml:"secretKey"`
	COSBucketURL    string `yaml:"bucketUrl"`
	CosBucketName   string `yaml:"bucketName"`
	CosBucketRegion string `yaml:"bucketRegion"`
	COSIsDebug      bool   `yaml:"isDebug"`
}

// Validate do validate
func (ost ObjectStoreTCloud) Validate() error {
	if len(ost.COSSecretID) == 0 {
		return errors.New("cos secret_id cannot be empty")
	}
	if len(ost.COSSecretKey) == 0 {
		return errors.New("cos secret_key cannot be empty")
	}
	if len(ost.COSBucketURL) == 0 {
		return errors.New("cos bucket_url cannot be empty")
	}
	if len(ost.CosBucketName) == 0 {
		return errors.New("cos bucket_name cannot be empty")
	}
	if len(ost.CosBucketRegion) == 0 {
		return errors.New("cos bucket_region cannot be empty")
	}
	if len(ost.UIN) == 0 {
		return errors.New("cos uin cannot be empty")
	}
	return nil
}

// Es elasticsearch config
type Es struct {
	Url      string    `json:"url"`
	User     string    `json:"user"`
	Password string    `json:"password"`
	TLS      TLSConfig `yaml:"tls"`
}

func (e Es) validate() error {
	if len(e.Url) == 0 {
		return errors.New("elasticsearch.url is not set")
	}

	if len(e.User) == 0 {
		return errors.New("elasticsearch.user is not set")
	}

	if len(e.Password) == 0 {
		return errors.New("elasticsearch.password is not set")
	}

	if err := e.TLS.validate(); err != nil {
		return fmt.Errorf("validate tls failed, err: %v", err)
	}

	return nil
}

// Jarvis 财务侧api配置
type Jarvis struct {
	AppID     string    `yaml:"appID"`
	AppKey    string    `yaml:"appKey"`
	Endpoints []string  `yaml:"endpoints"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate do validate
func (c *Jarvis) validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("jarvis endpoints is empty")
	}
	return nil
}

// ExchangeRate 汇率自动拉取配置
type ExchangeRate struct {
	EnablePull      bool                  `yaml:"enablePull"`
	PullIntervalMin uint64                `yaml:"pullIntervalMin"`
	ToCurrency      []enumor.CurrencyCode `yaml:"toCurrency"`
	FromCurrency    []enumor.CurrencyCode `yaml:"fromCurrency"`
}

func (r *ExchangeRate) trySetDefault() {
	if r.PullIntervalMin == 0 {
		r.PullIntervalMin = 120
	}
	if len(r.ToCurrency) == 0 {
		r.ToCurrency = []enumor.CurrencyCode{enumor.CurrencyCNY}
	}
}

// IEGObsOption OBS 账单拉取配置
type IEGObsOption struct {
	Endpoints []string  `yaml:"endpoints"`
	APIKey    string    `yaml:"apiKey"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate hcm runtime.
func (gt IEGObsOption) validate() error {
	if len(gt.Endpoints) == 0 {
		return errors.New("obs endpoints is not set")
	}
	if len(gt.APIKey) == 0 {
		return errors.New("obs api key is not set")
	}

	if err := gt.TLS.validate(); err != nil {
		return fmt.Errorf("validate obs tls failed, err: %v", err)
	}
	return nil
}

// AwsSavingsPlansOption savings plans allocation option
type AwsSavingsPlansOption struct {
	// RootAccountCloudID which root account these savings plans belongs to
	RootAccountCloudID string `yaml:"rootAccountCloudID" validate:"required"`
	// SpArnPrefix arn prefix to match savings plans, empty for no filter
	SpArnPrefix string `yaml:"spArnPrefix" validate:"omitempty"`
	// SpPurchaseAccountCloudID which account purchase this saving plans,
	// the cost of savings plans will be added to this account as income
	SpPurchaseAccountCloudID string `yaml:"SpPurchaseAccountCloudID" validate:"required"`
}

func (opt *AwsSavingsPlansOption) validate() error {
	if opt.RootAccountCloudID == "" {
		return errors.New("root account cloud id cannot be empty for aws savings plans")
	}

	if opt.SpPurchaseAccountCloudID == "" {
		return errors.New("sp purchase account cloud id cannot be empty for aws savings plans")
	}
	return nil
}

// NgateCli sops client options
type NgateCli struct {
	Host      string `yaml:"host"`
	AppCode   string `yaml:"app_code"`
	AppSecret string `yaml:"app_secret"`
}

func (c NgateCli) validate() error {
	if len(c.Host) == 0 {
		return errors.New("ngate.host is not set")
	}

	if len(c.AppCode) == 0 {
		return errors.New("ngate.app_code is not set")
	}

	if len(c.AppSecret) == 0 {
		return errors.New("ngate.app_secret is not set")
	}

	return nil
}

// RollingServer 滚服相关配置
type RollingServer struct {
	SyncBill           bool                      `yaml:"syncBill"`
	ReturnNotification RollingServerReturnNotice `yaml:"returnNotification"`
}

// RollingServerReturnNotice 滚服资源退换通知配置
type RollingServerReturnNotice struct {
	Enable           bool     `yaml:"enable"`
	SendToBusiness   bool     `yaml:"sendToBusiness"`
	DefaultReceivers []string `yaml:"defaultReceivers"`
}

// ResPlan 资源预测相关配置
type ResPlan struct {
	ReportPenaltyRatio   bool                       `yaml:"reportPenaltyRatio"`
	ExpireNotification   ResPlanExpireNotice        `yaml:"expireNotification"`
	NearExpiredTransfer  ResPlanNearExpiredTransfer `yaml:"nearExpiredTransfer"`
	RefreshTransferQuota ResPlanTransferQuota       `yaml:"refreshTransferQuota"`
	AdminAuditor         []string                   `yaml:"adminAuditor"`
	CRPOverLimitContact  []string                   `yaml:"crpOverLimitContact"`
}

// ResPlanNearExpiredTransfer 资源预测临期转移配置
type ResPlanNearExpiredTransfer struct {
	Enable bool `yaml:"enable"`
}

// ResPlanExpireNotice 资源预测过期通知配置
type ResPlanExpireNotice struct {
	Enable           bool     `yaml:"enable"`
	SendToBusiness   bool     `yaml:"sendToBusiness"`
	DefaultReceivers []string `yaml:"defaultReceivers"`
}

// ResPlanTransferQuota 资源预测转移额度配置
type ResPlanTransferQuota struct {
	Enable     bool  `yaml:"enable"`
	Quota      int64 `yaml:"quota"`
	AuditQuota int64 `yaml:"audit_quota"`
}

// MOA 太湖/MOA api配置
type MOA struct {
	PaasID          string      `yaml:"paasID"`
	Token           string      `yaml:"token"`
	Endpoints       []string    `yaml:"endpoints"`
	TLS             TLSConfig   `yaml:"tls"`
	DefaultTemplate MoaTplCfg   `yaml:"defaultTemplate"`
	Templates       []MoaTplCfg `yaml:"templates"`
}

// validate do validate
func (c *MOA) validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("moa endpoints is empty")
	}
	if err := c.DefaultTemplate.ValidateDefaultTemplate(); err != nil {
		return fmt.Errorf("moa default template config err: %w", err)
	}
	for i, template := range c.Templates {
		if err := template.Validate(false); err != nil {
			return fmt.Errorf("moa templates[%d] scene: %s, title: %s, err: %w",
				i, template.Scene, template.ZH.Title, err)
		}
	}

	return nil
}

// MoaTplCfg moa payload with i18n
type MoaTplCfg struct {
	Scene   enumor.MoaScene      `yaml:"scene" json:"-"`
	Channel enumor.Moa2FAChannel `yaml:"channel" json:"-" `
	Timeout time.Duration        `yaml:"timeout" json:"-" `
	ZH      MoaPayload           `yaml:"zh" json:"zh"`
	EN      MoaPayload           `yaml:"en" json:"en"`
}

// BuildPayload build payload to string
func (p *MoaTplCfg) BuildPayload(affectCount int) (string, error) {
	tmp := *p
	tmp.ZH.Desc = fmt.Sprintf(p.ZH.Desc, affectCount)
	tmp.EN.Desc = fmt.Sprintf(p.EN.Desc, affectCount)
	result, err := json.Marshal(tmp)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ValidateDefaultTemplate validate default template, all required fields must be set
func (p *MoaTplCfg) ValidateDefaultTemplate() error {
	return p.Validate(true)
}

// Validate only name is required
func (p *MoaTplCfg) Validate(defaultTemplateCheck bool) error {
	if defaultTemplateCheck {
		// 默认模板超时时间必填
		if p.Timeout == 0 {
			return errors.New("timeout is required for default template")
		}
		// 默认模板channel必填
		if len(p.Channel) == 0 {
			return errors.New("channel is not set for default template")
		}
	} else {
		// 默认模板不需要校验场景
		if len(p.Scene) == 0 {
			return errors.New("scene is not set")
		}
	}
	if p.Timeout < 0 {
		return errors.New("timeout should not be less than 0")
	}

	if len(p.Channel) > 0 {
		if err := p.Channel.Validate(); err != nil {
			return err
		}
	}
	if err := p.ZH.Validate(defaultTemplateCheck); err != nil {
		return fmt.Errorf("zh: %w", err)
	}
	if err := p.EN.Validate(defaultTemplateCheck); err != nil {
		return fmt.Errorf("en: %w", err)
	}

	return nil
}

// MoaPayload moa payload
type MoaPayload struct {
	Title     string      `json:"title" yaml:"title"`
	Desc      string      `json:"desc" yaml:"desc"`
	Navigator string      `json:"navigator" yaml:"navigator"`
	IconURL   *string     `json:"icon_url" yaml:"iconUrl"`
	Footer    string      `json:"footer" yaml:"footer"`
	Buttons   []MoaButton `json:"buttons" yaml:"buttons"`
}

// Over 生成用当前配置覆盖给定配置的副本
func (p *MoaPayload) Over(defaultPayload *MoaPayload) MoaPayload {
	overlay := cvt.PtrToVal(defaultPayload)
	if p.Title != "" {
		overlay.Title = p.Title
	}
	if p.Desc != "" {
		overlay.Desc = p.Desc
	}
	if p.Navigator != "" {
		overlay.Navigator = p.Navigator
	}

	if p.IconURL != nil {
		overlay.IconURL = p.IconURL
	}
	if p.Footer != "" {
		overlay.Footer = p.Footer
	}
	if p.Buttons != nil {
		overlay.Buttons = p.Buttons
	}
	return overlay
}

// Validate ...
func (p *MoaPayload) Validate(defaultTemplateCheck bool) error {
	if defaultTemplateCheck {
		if len(p.Title) == 0 {
			return errors.New("title is not set")
		}

		if len(p.Desc) == 0 {
			return errors.New("desc is not set")
		}

		if len(p.Buttons) == 0 {
			return errors.New("buttons is not set")
		}
	}

	for i, button := range p.Buttons {
		if err := button.Validate(); err != nil {
			return fmt.Errorf("buttons[%d]: %w", i, err)
		}
	}

	return nil

}

// MoaButton moa button
type MoaButton struct {
	Desc       string               `json:"desc" yaml:"desc" validate:"required"`
	ButtonType enumor.MoaButtonType `json:"button_type" yaml:"button_type"`
}

// Validate ...
func (b *MoaButton) Validate() error {
	if len(b.Desc) == 0 {
		return errors.New("desc is not set")
	}

	if len(b.ButtonType) == 0 {
		return errors.New("button_type is not set")
	}

	if err := b.ButtonType.Validate(); err != nil {
		return err
	}
	return nil
}

// AlarmCli alarm client options
type AlarmCli struct {
	AlarmApiAddr string `yaml:"host"`
	AppCode      string `yaml:"app_code"`
	AppSecret    string `yaml:"app_secret"`
}

func (c AlarmCli) validate() error {
	if len(c.AlarmApiAddr) == 0 {
		return errors.New("alarm.host is not set")
	}
	if len(c.AppCode) == 0 {
		return errors.New("alarm.app_code is not set")
	}
	if len(c.AppSecret) == 0 {
		return errors.New("alarm.app_secret is not set")
	}
	return nil
}

// Secret ...
type Secret struct {
	SubAccountID string `yaml:"sub_account_id"`
	ID           string `yaml:"id"`
	Key          string `yaml:"key"`
}

// Validate ...
func (s Secret) Validate() error {
	if len(s.SubAccountID) == 0 {
		return errors.New("sub account id is not set")
	}

	if len(s.ID) == 0 {
		return errors.New("secret id is not set")
	}

	if len(s.Key) == 0 {
		return errors.New("secret key is not set")
	}

	return nil
}

// CaiCheCli caiche client options
type CaiCheCli struct {
	Host      string `yaml:"host"`
	AppKey    string `yaml:"app_key"`
	AppSecret string `yaml:"app_secret"`
}

// Validate ...
func (c CaiCheCli) Validate() error {
	if c.Host == "" {
		return errors.New("caiche host is not set")
	}

	if c.AppKey == "" {
		return errors.New("caiche app_key is not set")
	}

	if c.AppSecret == "" {
		return errors.New("caiche app_secret is not set")
	}

	return nil
}

// ResourceSync 自研云-资源同步相关配置
type ResourceSync struct {
	SyncVpc      int `yaml:"syncVpc"`
	SyncSubnet   int `yaml:"syncSubnet"`
	SyncCapacity int `yaml:"syncCapacity"`
	SyncLeftIP   int `yaml:"syncLeftIP"`
}

// Validate ...
func (c ResourceSync) Validate() error {
	if c.SyncVpc < 0 {
		return errors.New("resourceSync.syncVpc is illegality")
	}

	if c.SyncSubnet < 0 {
		return errors.New("resourceSync.syncSubnet is illegality")
	}

	if c.SyncCapacity < 0 {
		return errors.New("resourceSync.syncCapacity is illegality")
	}

	if c.SyncLeftIP < 0 {
		return errors.New("resourceSync.syncLeftIP is illegality")
	}

	return nil
}

// StuckCheckCfg stuck check config.
type StuckCheckCfg struct {
	Enable       bool          `yaml:"enable"`
	Interval     time.Duration `yaml:"interval"`
	StartUpDelay time.Duration `yaml:"startUpDelay"`
	MaxTime      time.Duration `yaml:"maxTime"`
	MinTime      time.Duration `yaml:"minTime"`
}

// StuckCheck order stuck check.
type StuckCheck struct {
	Recycle StuckCheckCfg `yaml:"recycle"`
}

func (c *StuckCheck) trySetDefault() {
	c.Recycle.trySetDefault()
}

func (c *StuckCheckCfg) trySetDefault() {
	if c.Interval <= 0 {
		c.Interval = 24 * time.Hour
	}

	if c.MaxTime <= 0 {
		c.MaxTime = 14 * 24 * time.Hour
	}

	if c.MinTime <= 0 {
		c.MinTime = 24 * time.Hour
	}
	if c.StartUpDelay <= 0 {
		c.StartUpDelay = 2 * time.Minute
	}
}
