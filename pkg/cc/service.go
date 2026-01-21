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

package cc

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	initOnce sync.Once

	// serviceName is the runtime service's name.
	serviceName Name
)

// InitService set the initial service.
func InitService(sn Name) {
	initOnce.Do(func() {
		time.Local = time.FixedZone("UTC", 0)
		serviceName = sn
	})
}

// ServiceName return the current runtime service's name.
func ServiceName() Name {
	return serviceName
}

// Name is the name of the service
type Name string

const (
	// APIServerName is api server's name
	APIServerName Name = "api-server"
	// CloudServerName is cloud server's name
	CloudServerName Name = "cloud-server"
	// DataServiceName is data service's name
	DataServiceName Name = "data-service"
	// HCServiceName is hc service's name
	HCServiceName Name = "hc-service"
	// AuthServerName is the auth server's service name
	AuthServerName Name = "auth-server"
	// WebServerName is the web page server's name
	WebServerName Name = "web-server"
	// TaskServerName is task server's name
	TaskServerName Name = "task-server"
	// AccountServerName is account server's name
	AccountServerName Name = "account-server"
)

// Setting defines all service Setting interface.
type Setting interface {
	trySetFlagBindIP(ip net.IP) error
	trySetDefault()
	Validate() error
	TenantEnable() bool
}

// TestSetting test setting.
type TestSetting struct {
}

func (t TestSetting) trySetFlagBindIP(ip net.IP) error {
	return nil
}

func (t TestSetting) trySetDefault() {
	return
}

// Validate TestSetting option.
func (t TestSetting) Validate() error {
	return nil
}

// TenantEnable get tenant is enabled.
func (t TestSetting) TenantEnable() bool {
	return false
}

// ApiServerSetting defines api server used setting options.
type ApiServerSetting struct {
	Network Network      `yaml:"network"`
	Service Service      `yaml:"service"`
	Log     LogOption    `yaml:"log"`
	Tenant  TenantConfig `yaml:"tenant"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *ApiServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the ApiServerSetting default value if user not configured.
func (s *ApiServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()

	return
}

// Validate ApiServerSetting option.
func (s ApiServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *ApiServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// TaskManagement ...
type TaskManagement struct {
	// 关闭任务管理轮询
	Disable bool `yaml:"disable"`
}

// CloudServerSetting defines cloud server used setting options.
type CloudServerSetting struct {
	// 内部版配置
	FinOps ApiGateway `yaml:"finops"`
	MOA    MOA        `yaml:"moa"`

	Network          Network          `yaml:"network"`
	Service          Service          `yaml:"service"`
	Log              LogOption        `yaml:"log"`
	Crypto           Crypto           `yaml:"crypto"`
	BkHcmUrl         string           `yaml:"bkHcmUrl"`
	BkApigwHCMURL    string           `yaml:"bkApigwHCMUrl"`
	CloudResource    CloudResource    `yaml:"cloudResource"`
	Recycle          Recycle          `yaml:"recycle"`
	BillConfig       BillConfig       `yaml:"billConfig"`
	Itsm             ApiGateway       `yaml:"itsm"`
	CloudSelection   CloudSelection   `yaml:"cloudSelection"`
	Cmsi             CMSI             `yaml:"cmsi"`
	UserMgr          ApiGateway       `yaml:"userMgr"`
	OrgTopoConfig    BillConfig       `yaml:"orgTopoConfig"`
	TaskManagement   TaskManagement   `yaml:"taskManagement"`
	Tenant           TenantConfig     `yaml:"tenant"`
	Cmdb             ApiGateway       `yaml:"cmdb"`
	CCHostPoolBiz    int64            `yaml:"ccHostPoolBiz"`
	ConcurrentConfig ConcurrentConfig `yaml:"concurrentConfig"`
	TmpFileDir       string           `yaml:"tmpFileDir"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *CloudServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the CloudServerSetting default value if user not configured.
func (s *CloudServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.ConcurrentConfig.trySetDefault()
	if s.TmpFileDir == "" {
		s.TmpFileDir = "/tmp"
	}

	return
}

// Validate CloudServerSetting option.
func (s CloudServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Crypto.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if s.BkHcmUrl == "" {
		return fmt.Errorf("bkHcmUrl should not be empty")
	}

	if s.BkApigwHCMURL == "" {
		return fmt.Errorf("bkApigwHCMUrl should not be empty")
	}

	if err := s.CloudResource.validate(); err != nil {
		return err
	}

	if err := s.Recycle.validate(); err != nil {
		return err
	}

	if err := s.Itsm.validate(); err != nil {
		return err
	}

	if err := s.Cmsi.validate(); err != nil {
		return err
	}
	if err := s.MOA.validate(); err != nil {
		return err
	}

	if s.CCHostPoolBiz == 0 {
		return fmt.Errorf("ccHostPoolBiz should not be empty")
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *CloudServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// DataServiceSetting defines data service used setting options.
type DataServiceSetting struct {
	OBSDatabase *DataBase `yaml:"obsDatabase,omitempty"`

	Network     Network      `yaml:"network"`
	Service     Service      `yaml:"service"`
	Log         LogOption    `yaml:"log"`
	Database    DataBase     `yaml:"database"`
	Objectstore ObjectStore  `yaml:"objectstore"`
	Crypto      Crypto       `yaml:"crypto"`
	Cmdb        ApiGateway   `yaml:"cmdb"`
	Tenant      TenantConfig `yaml:"tenant"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *DataServiceSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the DataServiceSetting default value if user not configured.
func (s *DataServiceSetting) trySetDefault() {
	if s.OBSDatabase != nil {
		s.OBSDatabase.trySetDefault()
	}

	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Database.trySetDefault()

	return
}

// Validate DataServiceSetting option.
func (s DataServiceSetting) Validate() error {
	if s.OBSDatabase != nil {
		if err := s.OBSDatabase.validate(); err != nil {
			return err
		}
	}

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Database.validate(); err != nil {
		return err
	}

	if err := s.Crypto.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *DataServiceSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// HCServiceSetting defines hc service used setting options.
type HCServiceSetting struct {
	// 自研云增加的配置写在这里
	Esb                   Esb      `yaml:"esb"`
	ZiyanSecrets          []Secret `yaml:"ziyanSecrets"`
	SecurityGroupSkipList []string `yaml:"securityGroupMgmtSkipList"`

	Network       Network      `yaml:"network"`
	Service       Service      `yaml:"service"`
	Log           LogOption    `yaml:"log"`
	SyncConfig    SyncConfig   `yaml:"sync"`
	Tenant        TenantConfig `yaml:"tenant"`
	Cmdb          ApiGateway   `yaml:"cmdb"`
	CCHostPoolBiz int64        `yaml:"ccHostPoolBiz"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *HCServiceSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the HCServiceSetting default value if user not configured.
func (s *HCServiceSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.SyncConfig.trySetDefault()

	return
}

// Validate HCServiceSetting option.
func (s HCServiceSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}
	if err := s.SyncConfig.Validate(); err != nil {
		return fmt.Errorf("syncConfig validate error: %w", err)
	}

	for _, secret := range s.ZiyanSecrets {
		if err := secret.Validate(); err != nil {
			return err
		}
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if s.CCHostPoolBiz == 0 {
		return fmt.Errorf("ccHostPoolBiz should not be empty")
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *HCServiceSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// AuthServerSetting defines auth server used setting options.
type AuthServerSetting struct {
	Network Network      `yaml:"network"`
	Service Service      `yaml:"service"`
	Log     LogOption    `yaml:"log"`
	Esb     Esb          `yaml:"esb"`
	Cmdb    ApiGateway   `yaml:"cmdb"`
	Tenant  TenantConfig `yaml:"tenant"`

	IAM IAM `yaml:"iam"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *AuthServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the AuthServerSetting default value if user not configured.
func (s *AuthServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()

	return
}

// Validate AuthServerSetting option.
func (s AuthServerSetting) Validate() error {
	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Esb.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if err := s.IAM.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *AuthServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// WebServerSetting defines api server used setting options.
type WebServerSetting struct {
	Network       Network       `yaml:"network"`
	Service       Service       `yaml:"service"`
	Log           LogOption     `yaml:"log"`
	Web           Web           `yaml:"web"`
	Esb           Esb           `yaml:"esb"`
	Itsm          ApiGateway    `yaml:"itsm"`
	ChangeLogPath ChangeLogPath `yaml:"changeLogPath"`
	Notice        Notice        `yaml:"notice"`
	TemplatePath  string        `yaml:"templatePath"`
	Tenant        TenantConfig  `yaml:"tenant"`
	Cmdb          ApiGateway    `yaml:"cmdb"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *WebServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the ApiServerSetting default value if user not configured.
func (s *WebServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.ChangeLogPath.trySetDefault()
	if len(s.TemplatePath) == 0 {
		s.TemplatePath = "template"
	}

	return
}

// Validate ApiServerSetting option.
func (s WebServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Web.validate(); err != nil {
		return err
	}

	if err := s.Esb.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if err := s.Itsm.validate(); err != nil {
		return err
	}

	if err := s.Notice.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *WebServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// LabelSwitch switch for labels
type LabelSwitch struct {
	AwsCN bool `json:"awsCN" yaml:"awsCN"`
}

// TaskServerSetting defines task server used setting options.
type TaskServerSetting struct {
	// 自研云增加的配置写在这里
	OBSDatabase *DataBase   `yaml:"obsDatabase,omitempty"`
	Cmdb        ApiGateway  `yaml:"cmdb"`
	AlarmCli    *AlarmCli   `yaml:"alarm,omitempty"`
	SamPwdCli   *ApiGateway `yaml:"sampwd,omitempty"`

	Network  Network      `yaml:"network"`
	Service  Service      `yaml:"service"`
	Database DataBase     `yaml:"database"`
	Log      LogOption    `yaml:"log"`
	Async    Async        `yaml:"async"`
	Tenant   TenantConfig `yaml:"tenant"`

	UseLabel LabelSwitch `yaml:"useLabel"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *TaskServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the TaskServerSetting default value if user not configured.
func (s *TaskServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Database.trySetDefault()
	s.Log.trySetDefault()
	s.Async.trySetDefault()

	if s.OBSDatabase != nil {
		s.OBSDatabase.trySetDefault()
	}

	return
}

// Validate TaskServerSetting option.
func (s TaskServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Database.validate(); err != nil {
		return err
	}

	if s.OBSDatabase != nil {
		if err := s.OBSDatabase.validate(); err != nil {
			return err
		}
	}

	if s.AlarmCli != nil {
		if err := s.AlarmCli.validate(); err != nil {
			return err
		}
	}

	if s.SamPwdCli != nil {
		if err := s.SamPwdCli.validate(); err != nil {
			return err
		}
	}

	if err := s.Cmdb.validate(); err != nil {
		return fmt.Errorf("cmdb validate error: %w", err)
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *TaskServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// WoaServerSetting defines woa server used setting options.
type WoaServerSetting struct {
	Network           Network    `yaml:"network"`
	Service           Service    `yaml:"service"`
	Database          DataBase   `yaml:"database"`
	Log               LogOption  `yaml:"log"`
	Cmdb              ApiGateway `yaml:"cmdb"`
	FinOps            ApiGateway `yaml:"finops"`
	BkHcmURL          string     `yaml:"bkHcmUrl"`
	BkApigwHCMURL     string     `yaml:"bkApigwHCMUrl"`
	MongoDB           MongoDB    `yaml:"mongodb"`
	Watch             MongoDB    `yaml:"watch"`
	Redis             Redis      `yaml:"redis"`
	ClientConfig      `yaml:",inline"`
	ItsmFlows         []ItsmFlow        `yaml:"itsmFlows"`
	CancelItsmFlows   []ItsmFlow        `yaml:"cancelItsmFlows"`
	ResDissolve       ResourceDissolve  `yaml:"resourceDissolve"`
	Es                Es                `yaml:"elasticsearch"`
	Blacklist         string            `yaml:"blacklist"`
	UseMongo          bool              `yaml:"useMongo"`
	Recover           Recover           `yaml:"recover"`
	LocalTimezone     string            `yaml:"localTimezone"`
	RollingServer     RollingServer     `yaml:"rollingServer"`
	ResPlan           ResPlan           `yaml:"resPlan"`
	ResourceSync      ResourceSync      `yaml:"resourceSync"`
	Cmsi              CMSI              `yaml:"cmsi"`
	StuckCheck        StuckCheck        `yaml:"stuckCheck"`
	ApplyTicketConfig ApplyTicketConfig `yaml:"applyTicketConfig"`
	RecycleNotice     RecycleNotice     `yaml:"recycleNotice"`

	Tenant TenantConfig `yaml:"tenant"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *WoaServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the WoaServerSetting default value if user not configured.
func (s *WoaServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.StuckCheck.trySetDefault()

	return
}

// Validate TaskServerSetting option.
func (s WoaServerSetting) Validate() error {
	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if err := s.FinOps.validate(); err != nil {
		return err
	}

	if s.BkHcmURL == "" {
		return fmt.Errorf("bkHcmUrl should not be empty")
	}

	if s.BkApigwHCMURL == "" {
		return fmt.Errorf("bkApigwHCMUrl should not be empty")
	}

	// 开启Mongo之后才校验参数
	if s.UseMongo {
		if err := s.MongoDB.validate(); err != nil {
			return err
		}

		if err := s.Watch.validate(); err != nil {
			return err
		}
	}

	if err := s.Redis.validate(); err != nil {
		return err
	}

	if err := s.Database.validate(); err != nil {
		return err
	}

	if err := s.ClientConfig.validate(); err != nil {
		return err
	}

	if err := s.ResDissolve.validate(); err != nil {
		return err
	}

	if err := s.Es.validate(); err != nil {
		return err
	}

	if err := s.ResourceSync.Validate(); err != nil {
		return err
	}

	if err := s.ApplyTicketConfig.validate(); err != nil {
		return err
	}
	if err := s.RecycleNotice.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *WoaServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// AccountServerSetting defines task server used setting options.
type AccountServerSetting struct {
	// 自研云增加的配置写在这里
	FinOps         ApiGateway           `yaml:"finops"`
	Jarvis         Jarvis               `yaml:"jarvis"`
	ExchangeRate   ExchangeRate         `yaml:"exchangeRate"`
	IEGObsOption   IEGObsOption         `yaml:"obs"`
	Esb            Esb                  `yaml:"esb"`
	Network        Network              `yaml:"network"`
	Service        Service              `yaml:"service"`
	Controller     BillControllerOption `yaml:"controller"`
	Log            LogOption            `yaml:"log"`
	BillAllocation BillAllocationOption `yaml:"billAllocation"`
	TmpFileDir     string               `yaml:"tmpFileDir"`
	Tenant         TenantConfig         `yaml:"tenant"`
	Cmdb           ApiGateway           `yaml:"cmdb"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *AccountServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the TaskServerSetting default value if user not configured.
func (s *AccountServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Controller.trySetDefault()
	s.Log.trySetDefault()
	if s.TmpFileDir == "" {
		s.TmpFileDir = "/tmp"
	}

	//  内部版配置
	s.ExchangeRate.trySetDefault()
}

// Validate TaskServerSetting option.
func (s AccountServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Jarvis.validate(); err != nil {
		return err
	}

	if err := s.IEGObsOption.validate(); err != nil {
		return err
	}

	if err := s.BillAllocation.validate(); err != nil {
		return err
	}

	if err := s.Cmdb.validate(); err != nil {
		return err
	}

	if err := s.Esb.validate(); err != nil {
		return err
	}

	return nil
}

// TenantEnable get tenant is enabled.
func (s *AccountServerSetting) TenantEnable() bool {
	return s.Tenant.Enabled
}

// ChangeLogPath ...
type ChangeLogPath struct {
	Chinese string `yaml:"ch"`
	English string `yaml:"en"`
}

func (c *ChangeLogPath) trySetDefault() {
	if c.Chinese == "" {
		c.Chinese = "changelog/ch"
	}
	if c.English == "" {
		c.English = "changelog/en"
	}
}

// ApplyTicketConfig defines apply ticket config related settings.
type ApplyTicketConfig struct {
	PurchaseToResourcePool PurchaseToResourcePool `yaml:"purchaseToResourcePool"`
}

func (s *ApplyTicketConfig) validate() error {
	if err := s.PurchaseToResourcePool.validate(); err != nil {
		return err
	}

	return nil
}

// PurchaseToResourcePool defines purchase to resource pool related settings.
type PurchaseToResourcePool struct {
	User  string `yaml:"user"`
	BizID int64  `yaml:"bizID"`
}

func (s *PurchaseToResourcePool) validate() error {
	if s.User == "" {
		return fmt.Errorf("user should not be empty")
	}

	if s.BizID <= 0 {
		return fmt.Errorf("bizID should not be empty")
	}

	return nil
}
