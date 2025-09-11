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

// Package cc ...
package cc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/ssl"
	"hcm/pkg/version"

	etcd3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// Service defines Setting related runtime.
type Service struct {
	Etcd Etcd `yaml:"etcd"`
}

// trySetDefault set the Setting default value if user not configured.
func (s *Service) trySetDefault() {
	s.Etcd.trySetDefault()
}

// validate Setting related runtime.
func (s Service) validate() error {
	if err := s.Etcd.validate(); err != nil {
		return err
	}

	return nil
}

// Etcd defines etcd related runtime
type Etcd struct {
	// Endpoints is a list of URLs.
	Endpoints []string `yaml:"endpoints"`
	// DialTimeoutMS is the timeout seconds for failing
	// to establish a connection.
	DialTimeoutMS uint `yaml:"dialTimeoutMS"`
	// Username is a user's name for authentication.
	Username string `yaml:"username"`
	// Password is a password for authentication.
	Password string    `yaml:"password"`
	TLS      TLSConfig `yaml:"tls"`
}

// trySetDefault set the etcd default value if user not configured.
func (es *Etcd) trySetDefault() {
	if len(es.Endpoints) == 0 {
		es.Endpoints = []string{"127.0.0.1:2379"}
	}

	if es.DialTimeoutMS == 0 {
		es.DialTimeoutMS = 200
	}
}

// ToConfig convert to etcd config.
func (es Etcd) ToConfig() (etcd3.Config, error) {
	var tlsC *tls.Config
	if es.TLS.Enable() {
		var err error
		tlsC, err = ssl.ClientTLSConfVerify(es.TLS.InsecureSkipVerify, es.TLS.CAFile, es.TLS.CertFile,
			es.TLS.KeyFile, es.TLS.Password)
		if err != nil {
			return etcd3.Config{}, fmt.Errorf("init etcd tls config failed, err: %v", err)
		}
	}

	c := etcd3.Config{
		Endpoints:            es.Endpoints,
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(es.DialTimeoutMS) * time.Millisecond,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  tlsC,
		Username:             es.Username,
		Password:             es.Password,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	}

	// set grpc.WithBlock() to make sure quick fail when etcd endpoint is unavailable.
	// if not, etcd client may wait forever for incorrect(or unavailable) etcd endpoint
	// ref: https://github.com/etcd-io/etcd/issues/9877
	c.DialOptions = append(c.DialOptions, grpc.WithBlock())
	return c, nil
}

// validate etcd runtime
func (es Etcd) validate() error {
	if len(es.Endpoints) == 0 {
		return errors.New("etcd endpoints is not set")
	}

	if err := es.TLS.validate(); err != nil {
		return fmt.Errorf("etcd tls, %v", err)
	}

	return nil
}

// Limiter defines the request limit options
type Limiter struct {
	// QPS should >=1
	QPS uint `yaml:"qps"`
	// Burst should >= 1;
	Burst uint `yaml:"burst"`
}

// validate if the limiter is valid or not.
func (lm Limiter) validate() error {
	if lm.QPS <= 0 {
		return errors.New("invalid QPS value, should >= 1")
	}

	if lm.Burst <= 0 {
		return errors.New("invalid Burst value, should >= 1")
	}

	return nil
}

// trySetDefault try set the default value of limiter
func (lm *Limiter) trySetDefault() {
	if lm.QPS == 0 {
		lm.QPS = 1500
	}

	if lm.Burst == 0 {
		lm.Burst = 2000
	}
}

// Async defines async relating.
type Async struct {
	Scheduler  Parser     `yaml:"scheduler"`
	Executor   Executor   `yaml:"executor"`
	Dispatcher Dispatcher `yaml:"dispatcher"`
	WatchDog   WatchDog   `yaml:"watchDog"`
}

// Validate Async
func (a Async) Validate() error {
	// 这里不进行校验，统一由异步任务框架进行校验
	return nil
}

// Parser 公共组件，负责获取分配给当前节点的任务流，并解析成任务树后，派发当前要执行的任务给executor执行
type Parser struct {
	WatchIntervalSec                uint `yaml:"watchIntervalSec"`
	WorkerNumber                    uint `yaml:"workerNumber"`
	ScheduledFlowFetcherConcurrency uint `yaml:"scheduledFlowFetcherConcurrency"`
	CanceledFlowFetcherConcurrency  uint `yaml:"canceledFlowFetcherConcurrency"`
}

// Executor 公共组件，负责执行异步任务
type Executor struct {
	WorkerNumber       uint `yaml:"workerNumber"`
	TaskExecTimeoutSec uint `yaml:"taskExecTimeoutSec"`
}

// Dispatcher 主节点组件，负责派发任务
type Dispatcher struct {
	WatchIntervalSec              uint `yaml:"watchIntervalSec"`
	PendingFlowFetcherConcurrency uint `yaml:"pendingFlowFetcherConcurrency"`
}

// WatchDog 主节点组件，负责异常任务修正（超时任务，任务处理节点已经挂掉的任务等）
type WatchDog struct {
	WatchIntervalSec uint `yaml:"watchIntervalSec"`
	TaskTimeoutSec   uint `yaml:"taskTimeoutSec"`
	WorkerNumber     uint `yaml:"workerNumber"`
}

// DataBase defines database related runtime
type DataBase struct {
	Resource ResourceDB `yaml:"resource"`
	// MaxSlowLogLatencyMS defines the max tolerance in millisecond to execute
	// the database command, if the cost time of execute have >= the MaxSlowLogLatencyMS
	// then this request will be logged.
	MaxSlowLogLatencyMS uint `yaml:"maxSlowLogLatencyMS"`
	// Limiter defines request's to ORM's limitation for each sharding, and
	// each sharding have the independent request limitation.
	Limiter *Limiter `yaml:"limiter"`
}

// trySetDefault set the sharding default value if user not configured.
func (s *DataBase) trySetDefault() {
	s.Resource.trySetDefault()

	if s.MaxSlowLogLatencyMS == 0 {
		s.MaxSlowLogLatencyMS = 100
	}

	if s.Limiter == nil {
		s.Limiter = new(Limiter)
	}

	s.Limiter.trySetDefault()
}

// validate sharding runtime
func (s DataBase) validate() error {
	if err := s.Resource.validate(); err != nil {
		return err
	}

	if s.MaxSlowLogLatencyMS <= 0 {
		return errors.New("invalid maxSlowLogLatencyMS")
	}

	if s.Limiter != nil {
		if err := s.Limiter.validate(); err != nil {
			return fmt.Errorf("sharding.limiter is invalid, %v", err)
		}
	}

	return nil
}

// ResourceDB defines database related runtime.
type ResourceDB struct {
	// Endpoints is a seed list of host:port addresses of database nodes.
	Endpoints []string `yaml:"endpoints"`
	Database  string   `yaml:"database"`
	User      string   `yaml:"user"`
	Password  string   `yaml:"password"`
	// DialTimeoutSec is timeout in seconds to wait for a
	// response from the db server
	// all the timeout default value reference:
	// https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html
	DialTimeoutSec    uint      `yaml:"dialTimeoutSec"`
	ReadTimeoutSec    uint      `yaml:"readTimeoutSec"`
	WriteTimeoutSec   uint      `yaml:"writeTimeoutSec"`
	MaxIdleTimeoutMin uint      `yaml:"maxIdleTimeoutMin"`
	MaxOpenConn       uint      `yaml:"maxOpenConn"`
	MaxIdleConn       uint      `yaml:"maxIdleConn"`
	TLS               TLSConfig `yaml:"tls"`
	TimeZone          string    `yaml:"timeZone"`
}

// trySetDefault set the database's default value if user not configured.
func (ds *ResourceDB) trySetDefault() {
	if len(ds.Endpoints) == 0 {
		ds.Endpoints = []string{"127.0.0.1:3306"}
	}

	if ds.DialTimeoutSec == 0 {
		ds.DialTimeoutSec = 15
	}

	if ds.ReadTimeoutSec == 0 {
		ds.ReadTimeoutSec = 10
	}

	if ds.WriteTimeoutSec == 0 {
		ds.WriteTimeoutSec = 10
	}

	if ds.MaxOpenConn == 0 {
		ds.MaxOpenConn = 500
	}

	if ds.MaxIdleConn == 0 {
		ds.MaxIdleConn = 5
	}
	if len(ds.TimeZone) == 0 {
		ds.TimeZone = "UTC"
	}
}

// validate database runtime.
func (ds ResourceDB) validate() error {
	if len(ds.Endpoints) == 0 {
		return errors.New("database endpoints is not set")
	}

	if len(ds.Database) == 0 {
		return errors.New("database is not set")
	}

	if (ds.DialTimeoutSec > 0 && ds.DialTimeoutSec < 1) || ds.DialTimeoutSec > 60 {
		return errors.New("invalid database dialTimeoutMS, should be in [1:60]s")
	}

	if (ds.ReadTimeoutSec > 0 && ds.ReadTimeoutSec < 1) || ds.ReadTimeoutSec > 60 {
		return errors.New("invalid database readTimeoutMS, should be in [1:60]s")
	}

	if (ds.WriteTimeoutSec > 0 && ds.WriteTimeoutSec < 1) || ds.WriteTimeoutSec > 30 {
		return errors.New("invalid database writeTimeoutMS, should be in [1:30]s")
	}

	if err := ds.TLS.validate(); err != nil {
		return fmt.Errorf("database tls, %v", err)
	}

	return nil
}

// LogOption defines log's related configuration
type LogOption struct {
	LogDir           string `yaml:"logDir"`
	MaxPerFileSizeMB uint32 `yaml:"maxPerFileSizeMB"`
	MaxPerLineSizeKB uint32 `yaml:"maxPerLineSizeKB"`
	MaxFileNum       uint   `yaml:"maxFileNum"`
	LogAppend        bool   `yaml:"logAppend"`
	// log the log to std err only, it can not be used with AlsoToStdErr
	// at the same time.
	ToStdErr bool `yaml:"toStdErr"`
	// log the log to file and also to std err. it can not be used with ToStdErr
	// at the same time.
	AlsoToStdErr bool `yaml:"alsoToStdErr"`
	Verbosity    uint `yaml:"verbosity"`
}

// trySetDefault set the log's default value if user not configured.
func (log *LogOption) trySetDefault() {
	if len(log.LogDir) == 0 {
		log.LogDir = "./"
	}

	if log.MaxPerFileSizeMB == 0 {
		log.MaxPerFileSizeMB = 500
	}

	if log.MaxPerLineSizeKB == 0 {
		log.MaxPerLineSizeKB = 5
	}

	if log.MaxFileNum == 0 {
		log.MaxFileNum = 5
	}
}

// Logs convert it to logs.LogConfig.
func (log LogOption) Logs() logs.LogConfig {
	l := logs.LogConfig{
		LogDir:             log.LogDir,
		LogMaxSize:         log.MaxPerFileSizeMB,
		LogLineMaxSize:     log.MaxPerLineSizeKB,
		LogMaxNum:          log.MaxFileNum,
		RestartNoScrolling: log.LogAppend,
		ToStdErr:           log.ToStdErr,
		AlsoToStdErr:       log.AlsoToStdErr,
		Verbosity:          log.Verbosity,
	}

	return l
}

// Network defines all the network related options
type Network struct {
	// BindIP is ip where server working on
	BindIP string `yaml:"bindIP"`
	// Port is port where server listen to http port.
	Port uint      `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

// trySetFlagBindIP try set flag bind ip, bindIP only can set by one of the flag or configuration file.
func (n *Network) trySetFlagBindIP(ip net.IP) error {
	if len(ip) != 0 {
		if len(n.BindIP) != 0 {
			return errors.New("bind ip only can set by one of the flags or configuration file")
		}

		n.BindIP = ip.String()
		return nil
	}

	return nil
}

// trySetDefault set the network's default value if user not configured.
func (n *Network) trySetDefault() {
	if len(n.BindIP) == 0 {
		n.BindIP = "127.0.0.1"
	}
}

// validate network options
func (n Network) validate() error {
	if len(n.BindIP) == 0 {
		return errors.New("network bindIP is not set")
	}

	if ip := net.ParseIP(n.BindIP); ip == nil {
		return errors.New("invalid network bindIP")
	}

	if err := n.TLS.validate(); err != nil {
		return fmt.Errorf("network tls, %v", err)
	}

	return nil
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate.
	// For testing only.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Server requires TLS client certificate authentication
	CertFile string `yaml:"certFile"`
	// Server requires TLS client certificate authentication
	KeyFile string `yaml:"keyFile"`
	// Trusted root certificates for server
	CAFile string `yaml:"caFile"`
	// the password to decrypt the certificate
	Password string `yaml:"password"`
}

// Enable test tls if enable.
func (tls TLSConfig) Enable() bool {
	if len(tls.CertFile) == 0 &&
		len(tls.KeyFile) == 0 &&
		len(tls.CAFile) == 0 {
		return false
	}

	return true
}

// Validate validates TLS configuration
func (tls TLSConfig) Validate() error {
	if tls.Enable() {
		if len(tls.CertFile) > 0 && len(tls.KeyFile) == 0 {
			return fmt.Errorf("client key file is required when cert file is specified")
		}
		if len(tls.KeyFile) > 0 && len(tls.CertFile) == 0 {
			return fmt.Errorf("client cert file is required when key file is specified")
		}
	}
	return nil
}

// validate tls configs
func (tls TLSConfig) validate() error {
	if !tls.Enable() {
		return nil
	}

	// TODO: add tls config validate.

	return nil
}

// SysOption is the system's normal option, which is parsed from
// flag commandline.
type SysOption struct {
	ConfigFile string
	// BindIP Setting startup bind ip.
	BindIP net.IP
	// Versioned Setting if show current version info.
	Versioned bool

	// current env for service discovery
	Environment string
	// service label for service discovery
	Labels []string
	// if true always be follower
	DisableElection bool
}

// CheckV check if show current version info.
func (s SysOption) CheckV() {
	if s.Versioned {
		version.ShowVersion()
		os.Exit(0)
	}
}

// IAM defines all the iam related runtime.
type IAM struct {
	// Endpoints is a seed list of host:port addresses of iam nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode blueking belong to hcm's appcode.
	AppCode string `yaml:"appCode"`
	// AppSecret blueking belong to hcm app's secret.
	AppSecret string    `yaml:"appSecret"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate iam runtime.
func (s IAM) validate() error {
	if len(s.Endpoints) == 0 {
		return errors.New("iam endpoints is not set")
	}

	if len(s.AppCode) == 0 {
		return errors.New("iam appcode is not set")
	}

	if len(s.AppSecret) == 0 {
		return errors.New("iam app secret is not set")
	}

	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("iam tls validate failed, err: %v", err)
	}

	return nil
}

// Web 服务依赖所需特有配置， 包括登录、静态文件等配置的定义
type Web struct {
	StaticFileDirPath string `yaml:"staticFileDirPath"`

	BkLoginCookieName      string `yaml:"bkLoginCookieName"`
	BkLoginUrl             string `yaml:"bkLoginUrl"`
	BkComponentApiUrl      string `yaml:"bkComponentApiUrl"`
	BkItsmUrl              string `yaml:"bkItsmUrl"`
	BkDomain               string `yaml:"bkDomain"`
	BkCmdbCreateBizUrl     string `yaml:"bkCmdbCreateBizUrl"`
	BkCmdbCreateBizDocsUrl string `yaml:"bkCmdbCreateBizDocsUrl"`
	EnableCloudSelection   bool   `yaml:"enableCloudSelection"`
	EnableAccountBill      bool   `yaml:"enableAccountBill"`
}

func (s Web) validate() error {
	if len(s.BkLoginUrl) == 0 {
		return errors.New("bk_login_url is not set")
	}

	if len(s.BkComponentApiUrl) == 0 {
		return errors.New("bk_component_api_url is not set")
	}

	if len(s.BkItsmUrl) == 0 {
		return errors.New("bk_itsm_url is not set")
	}

	if len(s.BkDomain) == 0 {
		return errors.New("bk_domain is not set")
	}

	return nil
}

// Esb defines the esb related runtime.
type Esb struct {
	// Endpoints is a seed list of host:port addresses of esb nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of hcm to request esb.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of hcm to request esb.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of hcm to request esb.
	User string    `yaml:"user"`
	TLS  TLSConfig `yaml:"tls"`
}

// validate esb runtime.
func (s Esb) validate() error {
	if len(s.Endpoints) == 0 {
		return errors.New("esb endpoints is not set")
	}
	if len(s.AppCode) == 0 {
		return errors.New("esb app code is not set")
	}
	if len(s.AppSecret) == 0 {
		return errors.New("esb app secret is not set")
	}
	if len(s.User) == 0 {
		return errors.New("esb user is not set")
	}
	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("validate esb tls failed, err: %v", err)
	}
	return nil
}

// AesGcm Aes Gcm加密
type AesGcm struct {
	Key   string `yaml:"key"`
	Nonce string `yaml:"nonce"`
}

func (a AesGcm) validate() error {
	if len(a.Key) != 16 && len(a.Key) != 32 {
		return errors.New("invalid key, should be 16 or 32 bytes")
	}

	if len(a.Nonce) != 12 {
		return errors.New("invalid nonce, should be 12 bytes")
	}

	return nil
}

// Crypto 定义项目里需要用到的加密，包括选择的算法等
// TODO: 这里默认只支持AES Gcm算法，后续需要支持国密等的选择，可能还需要支持根据不同场景配置不同（比如不同场景，加密的密钥等都不一样）
type Crypto struct {
	AesGcm AesGcm `yaml:"aesGcm"`
}

func (c Crypto) validate() error {
	if err := c.AesGcm.validate(); err != nil {
		return err
	}

	return nil
}

// CloudResource 云资源配置
type CloudResource struct {
	Sync CloudResourceSync `yaml:"sync"`
}

func (c CloudResource) validate() error {
	if err := c.Sync.validate(); err != nil {
		return err
	}

	return nil
}

// CloudResourceSync 云资源同步配置
type CloudResourceSync struct {
	Enable                       bool   `yaml:"enable"`
	SyncIntervalMin              uint64 `yaml:"syncIntervalMin"`
	SyncFrequencyLimitingTimeMin uint64 `yaml:"syncFrequencyLimitingTimeMin"`
}

func (c CloudResourceSync) validate() error {
	if c.Enable {
		if c.SyncFrequencyLimitingTimeMin < 10 {
			return errors.New("syncFrequencyLimitingTimeMin must > 10")
		}
	}

	return nil
}

// Recycle configuration.
type Recycle struct {
	AutoDeleteTime uint `yaml:"autoDeleteTimeHour"`
}

func (a Recycle) validate() error {
	if a.AutoDeleteTime == 0 {
		return errors.New("autoDeleteTimeHour must > 0")
	}

	return nil
}

// BillConfig 账号账单配置
type BillConfig struct {
	Enable          bool   `yaml:"enable"`
	SyncIntervalMin uint64 `yaml:"syncIntervalMin"`
}

func (c BillConfig) validate() error {
	if c.Enable && c.SyncIntervalMin < 1 {
		return errors.New("BillConfig.SyncIntervalMin must >= 1")
	}

	return nil
}

// ApiGateway defines the api gateway config.
type ApiGateway struct {
	// Endpoints is a seed list of host:port addresses of api gateway.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of hcm to request api gateway.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of hcm to request api gateway.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of hcm to request api gateway.
	User string `yaml:"user"`
	// BkTicket is the BlueKing access ticket of hcm to request api gateway.
	BkTicket string `yaml:"bkTicket"`
	// BkToken is the BlueKing access token of hcm to request api gateway.
	BkToken string    `yaml:"bkToken"`
	TLS     TLSConfig `yaml:"tls"`
}

// validate hcm runtime.
func (gt ApiGateway) validate() error {
	if len(gt.Endpoints) == 0 {
		return errors.New("api gateway endpoints is not set")
	}
	if len(gt.AppCode) == 0 {
		return errors.New("app code is not set")
	}
	if len(gt.AppSecret) == 0 {
		return errors.New("app secret is not set")
	}

	if len(gt.BkToken) != 0 && len(gt.BkTicket) != 0 {
		return errors.New("bkToken or bkTicket only one is needed")
	}

	if err := gt.TLS.validate(); err != nil {
		return fmt.Errorf("validate tls failed, err: %v", err)
	}
	return nil
}

// GetAuthValue get auth value.
func (gt ApiGateway) GetAuthValue() string {

	if len(gt.BkTicket) != 0 {
		return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_ticket\":\"%s\"}",
			gt.AppCode, gt.AppSecret, gt.BkTicket)
	}

	if len(gt.BkToken) != 0 {
		return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"access_token\":\"%s\"}",
			gt.AppCode, gt.AppSecret, gt.BkToken)
	}

	return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}", gt.AppCode, gt.AppSecret)
}

// CloudSelection define cloud selection relation setting.
type CloudSelection struct {
	DefaultSampleOffset  int                       `yaml:"userDistributionSampleOffset"`
	AvgLatencySampleDays int                       `yaml:"avgLatencySampleDays"`
	CoverRate            float64                   `yaml:"coverRate"`
	CoverPingRanges      []ThreshHoldRanges        `yaml:"coverPingRanges"`
	IDCPriceRanges       []ThreshHoldRanges        `yaml:"idcPriceRanges"`
	AlgorithmPlugin      Plugin                    `yaml:"algorithmPlugin"`
	TableNames           CloudSelectionTableNames  `yaml:"tableNames"`
	DataSourceType       string                    `yaml:"dataSourceType"`
	BkBase               BkBase                    `yaml:"bkBase"`
	DefaultIdcPrice      map[enumor.Vendor]float64 `yaml:"defaultIdcPrice"`
}

// Plugin outside binary plugin
type Plugin struct {
	BinaryPath string   `yaml:"binaryPath"`
	Args       []string `yaml:"args"`
}

// BkBase define bkbase relation setting.
type BkBase struct {
	QueryLimit uint   `yaml:"queryLimit"`
	DataToken  string `yaml:"dataToken"`
	ApiGateway `yaml:"-,inline"`
}

// Validate ...
func (b BkBase) Validate() error {
	if err := b.ApiGateway.validate(); err != nil {
		return err
	}

	if len(b.DataToken) == 0 {
		return errors.New("data token is required")
	}

	return nil
}

// Validate define cloud selection relation setting.
func (c CloudSelection) Validate() error {
	switch c.DataSourceType {
	case "bk_base":
		if err := c.BkBase.validate(); err != nil {
			return err
		}

	default:
		return fmt.Errorf("data source: %s not support", c.DataSourceType)
	}

	return nil
}

// CloudSelectionTableNames ...
type CloudSelectionTableNames struct {
	LatencyPingProvinceIdc   string `yaml:"latencyPingProvinceIdc"`
	LatencyBizProvinceIdc    string `yaml:"latencyBizProvinceIdc"`
	UserCountryDistribution  string `yaml:"userCountryDistribution"`
	UserProvinceDistribution string `yaml:"userProvinceDistribution"`
	RecommendDataSource      string `yaml:"recommendDataSource"`
}

// ThreshHoldRanges 评分范围
type ThreshHoldRanges struct {
	Score int   `yaml:"score" json:"score"`
	Range []int `yaml:"range" json:"range"`
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

var (
	defaultControllerSyncDuration         = 30 * time.Second
	defaultMainAccountSummarySyncDuration = 10 * time.Minute
	defaultRootAccountSummarySyncDuration = 10 * time.Minute
	defaultDailySummarySyncDuration       = 30 * time.Second
	defaultMonthTaskSyncDuration          = 30 * time.Second
)

// BillControllerOption bill controller option
type BillControllerOption struct {
	// 是否关闭整个账单同步，默认为不关闭
	Disable                        bool           `yaml:"disable"`
	ControllerSyncDuration         *time.Duration `yaml:"controllerSyncDuration,omitempty"`
	MainAccountSummarySyncDuration *time.Duration `yaml:"mainAccountSummarySyncDuration,omitempty"`
	RootAccountSummarySyncDuration *time.Duration `yaml:"rootAccountSummarySyncDuration,omitempty"`
	MonthTaskSyncDuration          *time.Duration `yaml:"monthTaskSyncDuration,omitempty"`
	DailySummarySyncDuration       *time.Duration `yaml:"dailySummarySyncDuration,omitempty"`
}

func (bco *BillControllerOption) trySetDefault() {
	if bco.ControllerSyncDuration == nil {
		bco.ControllerSyncDuration = &defaultControllerSyncDuration
	}
	if bco.MainAccountSummarySyncDuration == nil {
		bco.MainAccountSummarySyncDuration = &defaultMainAccountSummarySyncDuration
	}
	if bco.RootAccountSummarySyncDuration == nil {
		bco.RootAccountSummarySyncDuration = &defaultRootAccountSummarySyncDuration
	}
	if bco.MonthTaskSyncDuration == nil {
		bco.MonthTaskSyncDuration = &defaultMonthTaskSyncDuration
	}
	if bco.DailySummarySyncDuration == nil {
		bco.DailySummarySyncDuration = &defaultDailySummarySyncDuration
	}
}

// CMSI cmsi config
type CMSI struct {
	CC         []string `yaml:"cc"`
	Sender     string   `yaml:"sender"`
	ApiGateway `yaml:"-,inline"`
}

// Validate do validate
func (c *CMSI) validate() error {
	if err := c.ApiGateway.validate(); err != nil {
		return err
	}

	if c.CC == nil || len(c.CC) == 0 {
		c.CC = make([]string, 0)
	}

	if len(c.Sender) == 0 {
		return errors.New("sender cannot be empty")
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

// BillCommonExpense ...
type BillCommonExpense struct {
	ExcludeAccountCloudIDs []string `yaml:"excludeAccountCloudIDs" validate:"dive,required"`
}

// BillDeductItemsExpense ...
type BillDeductItemsExpense struct {
	DeductItemTypes map[string]map[string][]string `yaml:"deductItemTypes" validate:"dive,required"`
}

// CreditReturn ...
type CreditReturn struct {
	CreditID string `yaml:"creditId" validate:"required"`
	// which account this credit will return to
	AccountCloudID string `yaml:"accountCloudID" validate:"required"`
	CreditName     string `yaml:"creditName" `
}

// Validate ...
func (r CreditReturn) Validate() error {
	if r.CreditID == "" {
		return errors.New("credit id cannot be empty")
	}
	if r.AccountCloudID == "" {
		return errors.New("account cloud id cannot be empty")
	}
	return nil
}

// GcpCreditConfig ...
type GcpCreditConfig struct {
	// RootAccountCloudID which root account these savings plans belongs to
	RootAccountCloudID string         `yaml:"rootAccountCloudID" validate:"required"`
	ReturnConfigs      []CreditReturn `yaml:"returnConfigs" validate:"required,dive,required"`
}

// Validate ...
func (opt *GcpCreditConfig) Validate() error {
	if opt.RootAccountCloudID == "" {
		return errors.New("root account cloud id cannot be empty for gcp credits config")
	}
	if len(opt.ReturnConfigs) == 0 {
		return errors.New("return configs cannot be empty for gcp credits config")
	}
	for i := range opt.ReturnConfigs {
		if err := opt.ReturnConfigs[i].Validate(); err != nil {
			return errors.New(fmt.Sprintf("gcp credit return config index %d validation failed, %v", i, err))
		}
	}
	return nil
}

// BillAllocationOption ...
type BillAllocationOption struct {
	AwsSavingsPlans       []AwsSavingsPlansOption `yaml:"awsSavingsPlans"`
	AwsCommonExpense      BillCommonExpense       `yaml:"awsCommonExpense"`
	AwsDeductAccountItems BillDeductItemsExpense  `yaml:"awsDeductAccountItems"`
	GcpCredits            []GcpCreditConfig       `yaml:"gcpCredits"`
	GcpCommonExpense      BillCommonExpense       `yaml:"gcpCommonExpense"`
	HuaweiCommonExpense   BillCommonExpense       `yaml:"huaweiCommonExpense"`
}

func (opt *BillAllocationOption) validate() error {
	for i := range opt.AwsSavingsPlans {
		if err := opt.AwsSavingsPlans[i].validate(); err != nil {
			return errors.New(fmt.Sprintf("aws savings plans index %d validation failed, %v", i, err))
		}
	}
	return nil
}

// Notice ...
type Notice struct {
	Enable     bool `yaml:"enable"`
	ApiGateway `yaml:"-,inline"`
}

// Validate do validate
func (c *Notice) validate() error {
	if !c.Enable {
		return nil
	}
	if err := c.ApiGateway.validate(); err != nil {
		return err
	}

	return nil
}

// SyncConfig defines sync config.
type SyncConfig struct {
	DefaultConcurrent uint `yaml:"defaultConcurrent"`
	// 并发配置
	ConcurrentRules []SyncConcurrentRule `yaml:"concurrentRules"`
}

func (s *SyncConfig) trySetDefault() {
	if s.DefaultConcurrent == 0 {
		s.DefaultConcurrent = 1
	}
	for i := range s.ConcurrentRules {
		r := &s.ConcurrentRules[i]
		r.trySetDefault()
		if r.ListConcurrent == 0 {
			r.ListConcurrent = r.SyncConcurrent
		}
	}
}

// Validate ...
func (s SyncConfig) Validate() error {
	if s.DefaultConcurrent == 0 {
		return errors.New("defaultConcurrent is not set")
	}
	for _, c := range s.ConcurrentRules {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GetSyncConcurrent 获取同步并发，按顺序匹配，一旦匹配立即返回
func (s SyncConfig) GetSyncConcurrent(vendor enumor.Vendor, resource enumor.CloudResourceType, region string) (
	listing, syncing uint) {

	for _, c := range s.ConcurrentRules {
		if c.Match(vendor, resource, region) {
			return c.ListConcurrent, c.SyncConcurrent
		}
	}
	return s.DefaultConcurrent, s.DefaultConcurrent
}

// SyncConcurrentRule 同步并发配置
type SyncConcurrentRule struct {
	Rule     string                   `yaml:"rule"`
	vendor   enumor.Vendor            `yaml:"vendor"`
	resource enumor.CloudResourceType `yaml:"resource"`
	region   string                   `yaml:"region"`
	// for syncing resource
	SyncConcurrent uint `yaml:"syncConcurrent"`
	// for listing resource from cloud if not set will be set to SyncConcurrent
	ListConcurrent uint `yaml:"listConcurrent"`
}

// String ...
func (s SyncConcurrentRule) String() string {
	return fmt.Sprintf("rule: %s, syncConcurrent: %d, listConcurrent: %d", s.Rule, s.SyncConcurrent, s.ListConcurrent)
}

// ConcurrentWildcard 通配关键字
const ConcurrentWildcard = "*"

// Match ...
func (r *SyncConcurrentRule) Match(vendor enumor.Vendor, resource enumor.CloudResourceType, region string) bool {
	if r == nil {
		return false
	}
	if r.vendor != ConcurrentWildcard && r.vendor != vendor {
		return false
	}
	if r.resource != ConcurrentWildcard && r.resource != resource {
		return false
	}
	if r.region != ConcurrentWildcard && r.region != region {
		return false
	}
	return true
}

func (r *SyncConcurrentRule) trySetDefault() {
	if r.Rule == "" {
		return
	}
	parts := strings.Split(r.Rule, "/")
	if len(parts) > 0 {
		r.vendor = enumor.Vendor(parts[0])
	}

	if len(parts) > 1 {
		r.resource = enumor.CloudResourceType(parts[1])
	}
	if len(parts) > 2 {
		r.region = parts[2]
	}

}

// Validate ...
func (r *SyncConcurrentRule) Validate() error {
	if r == nil {
		return errors.New("sync concurrent rule is nil")
	}
	if len(r.Rule) == 0 {
		return errors.New("empty sync concurrent rule")
	}
	if r.vendor == "" {
		return errors.New("invalid sync concurrent rule: empty vendor")
	}
	if r.resource == "" {
		return errors.New("invalid sync concurrent rule: empty resource")
	}
	if r.region == "" {
		return errors.New("invalid sync concurrent rule: empty region")
	}

	if r.ListConcurrent == 0 {
		return errors.New("invalid list concurrent number")
	}
	if r.SyncConcurrent == 0 {
		return errors.New("invalid sync concurrent number")
	}
	return nil
}

// TenantConfig tenant config
type TenantConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ConcurrentConfig CLB import config
type ConcurrentConfig struct {
	CLBImportCount int `yaml:"clbImportCount"`
}

func (c *ConcurrentConfig) trySetDefault() {
	if c.CLBImportCount == 0 {
		c.CLBImportCount = 10
	}
}
