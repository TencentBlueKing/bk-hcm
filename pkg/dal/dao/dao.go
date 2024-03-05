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

// Package dao ...
package dao

import (
	"fmt"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/dal/dao/application"
	daoasync "hcm/pkg/dal/dao/async"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/auth"
	"hcm/pkg/dal/dao/cloud"
	daoselection "hcm/pkg/dal/dao/cloud-selection"
	argstpl "hcm/pkg/dal/dao/cloud/argument-template"
	"hcm/pkg/dal/dao/cloud/bill"
	"hcm/pkg/dal/dao/cloud/cvm"
	"hcm/pkg/dal/dao/cloud/disk"
	diskcvmrel "hcm/pkg/dal/dao/cloud/disk-cvm-rel"
	"hcm/pkg/dal/dao/cloud/eip"
	eipcvmrel "hcm/pkg/dal/dao/cloud/eip-cvm-rel"
	cimage "hcm/pkg/dal/dao/cloud/image"
	networkinterface "hcm/pkg/dal/dao/cloud/network-interface"
	nicvmrel "hcm/pkg/dal/dao/cloud/network-interface-cvm-rel"
	"hcm/pkg/dal/dao/cloud/region"
	resourcegroup "hcm/pkg/dal/dao/cloud/resource-group"
	routetable "hcm/pkg/dal/dao/cloud/route-table"
	securitygroup "hcm/pkg/dal/dao/cloud/security-group"
	sgcvmrel "hcm/pkg/dal/dao/cloud/security-group-cvm-rel"
	daosubaccount "hcm/pkg/dal/dao/cloud/sub-account"
	daosync "hcm/pkg/dal/dao/cloud/sync"
	"hcm/pkg/dal/dao/cloud/zone"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	recyclerecord "hcm/pkg/dal/dao/recycle-record"
	daouser "hcm/pkg/dal/dao/user"
	"hcm/pkg/kit"
	"hcm/pkg/metrics"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
)

// Set defines all the DAO to be operated.
type Set interface {
	Audit() audit.Interface
	Auth() auth.Auth
	Account() cloud.Account
	SubAccount() daosubaccount.SubAccount
	SecurityGroup() securitygroup.SecurityGroup
	SGCvmRel() sgcvmrel.Interface
	TCloudSGRule() securitygroup.TCloudSGRule
	AwsSGRule() securitygroup.AwsSGRule
	HuaWeiSGRule() securitygroup.HuaWeiSGRule
	AzureSGRule() securitygroup.AzureSGRule
	GcpFirewallRule() cloud.GcpFirewallRule
	Cloud() cloud.Cloud
	AccountBizRel() cloud.AccountBizRel
	Vpc() cloud.Vpc
	Subnet() cloud.Subnet
	HuaWeiRegion() region.HuaWeiRegion
	AzureRG() resourcegroup.AzureRG
	AzureRegion() region.AzureRegion
	Zone() zone.Zone
	AccountSyncDetail() daosync.AccountSyncDetail
	TCloudRegion() region.TCloudRegion
	AwsRegion() region.AwsRegion
	GcpRegion() region.GcpRegion
	Cvm() cvm.Interface
	RouteTable() routetable.RouteTable
	Route() routetable.Route
	Application() application.Application
	ApprovalProcess() application.ApprovalProcess
	NetworkInterface() networkinterface.NetworkInterface
	RecycleRecord() recyclerecord.RecycleRecord
	Eip() eip.Eip
	Disk() disk.Disk
	NiCvmRel() nicvmrel.NiCvmRel
	Image() cimage.Image
	DiskCvmRel() diskcvmrel.DiskCvmRel
	EipCvmRel() eipcvmrel.EipCvmRel
	AccountBillConfig() bill.Interface
	AsyncFlow() daoasync.AsyncFlow
	AsyncFlowTask() daoasync.AsyncFlowTask
	UserCollection() daouser.Interface
	CloudSelectionScheme() daoselection.SchemeInterface
	CloudSelectionBizType() daoselection.BizTypeInterface
	CloudSelectionIdc() daoselection.IdcInterface
	ArgsTpl() argstpl.Interface

	Txn() *Txn
}

// NewDaoSet create the DAO set instance.
func NewDaoSet(opt cc.DataBase) (Set, error) {
	db, err := connect(opt.Resource)
	if err != nil {
		return nil, fmt.Errorf("init sharding failed, err: %v", err)
	}

	ormInst := orm.InitOrm(db, orm.MetricsRegisterer(metrics.Register()),
		orm.IngressLimiter(opt.Limiter.QPS, opt.Limiter.Burst), orm.SlowRequestMS(opt.MaxSlowLogLatencyMS))

	idGen := idgenerator.New(db, idgenerator.DefaultMaxRetryCount)

	s := &set{
		idGen: idGen,
		orm:   ormInst,
		db:    db,
		audit: audit.NewAudit(ormInst),
	}

	return s, nil
}

// connect to mysql
func connect(opt cc.ResourceDB) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", uri(opt))
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed, err: %v", err)
	}

	db.SetMaxOpenConns(int(opt.MaxOpenConn))
	db.SetMaxIdleConns(int(opt.MaxIdleConn))
	db.SetConnMaxLifetime(time.Duration(opt.MaxIdleTimeoutMin) * time.Minute)

	return db, nil
}

// uri generate the standard db connection string format uri.
func uri(opt cc.ResourceDB) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s",
		opt.User,
		opt.Password,
		strings.Join(opt.Endpoints, ","),
		opt.Database,
		opt.DialTimeoutSec,
		opt.ReadTimeoutSec,
		opt.WriteTimeoutSec,
		"utf8mb4",
	)
}

type set struct {
	idGen idgenerator.IDGenInterface
	orm   orm.Interface
	db    *sqlx.DB
	audit audit.Interface
}

// EipCvmRel return EipCvmRel dao.
func (s *set) EipCvmRel() eipcvmrel.EipCvmRel {
	return &eipcvmrel.EipCvmRelDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// DiskCvmRel return DiskCvmRel dao.
func (s *set) DiskCvmRel() diskcvmrel.DiskCvmRel {
	return &diskcvmrel.DiskCvmRelDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// Image return Image dao.
func (s *set) Image() cimage.Image {
	return &cimage.ImageDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// NiCvmRel return NiCvmRel dao.
func (s *set) NiCvmRel() nicvmrel.NiCvmRel {
	return &nicvmrel.NiCvmRelDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// Disk return Disk dao.
func (s *set) Disk() disk.Disk {
	return &disk.DiskDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// Eip return Eip dao.
func (s *set) Eip() eip.Eip {
	return &eip.EipDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// Zone return Zone dao.
func (s *set) Zone() zone.Zone {
	return &zone.ZoneDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AccountSyncDetail return AccountSyncDetail dao.
func (s *set) AccountSyncDetail() daosync.AccountSyncDetail {
	return &daosync.AccountSyncDetailDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AzureRegion return AzureRegion dao.
func (s *set) AzureRegion() region.AzureRegion {
	return &region.AzureRegionDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AzureRG return AzureRG dao.
func (s *set) AzureRG() resourcegroup.AzureRG {
	return &resourcegroup.AzureRGDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// HuaWeiRegion return HuaWeiRegion dao.
func (s *set) HuaWeiRegion() region.HuaWeiRegion {
	return &region.HuaWeiRegionDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// Account return account dao.
func (s *set) Account() cloud.Account {
	return &cloud.AccountDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// SubAccount return sub account dao.
func (s *set) SubAccount() daosubaccount.SubAccount {
	return &daosubaccount.SubAccountDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// AccountBizRel returns account biz relation dao.
func (s *set) AccountBizRel() cloud.AccountBizRel {
	return &cloud.AccountBizRelDao{
		Orm: s.orm,
	}
}

// Vpc returns vpc dao.
func (s *set) Vpc() cloud.Vpc {
	return cloud.NewVpcDao(s.orm, s.idGen, s.audit)
}

// Subnet returns subnet dao.
func (s *set) Subnet() cloud.Subnet {
	return cloud.NewSubnetDao(s.orm, s.idGen, s.audit)
}

// Auth return auth dao.
func (s *set) Auth() auth.Auth {
	return &auth.AuthDao{
		Orm: s.orm,
	}
}

// Cloud return cloud dao.
func (s *set) Cloud() cloud.Cloud {
	return &cloud.CloudDao{
		Orm: s.orm,
	}
}

// SecurityGroup return security group dao.
func (s *set) SecurityGroup() securitygroup.SecurityGroup {
	return &securitygroup.SecurityGroupDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// SGCvmRel return security group cvm rel dao.
func (s *set) SGCvmRel() sgcvmrel.Interface {
	return &sgcvmrel.Dao{
		Orm: s.orm,
	}
}

// TCloudSGRule return tcloud security group rule dao.
func (s *set) TCloudSGRule() securitygroup.TCloudSGRule {
	return &securitygroup.TCloudSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// GcpFirewallRule return gcp firewall rule dao.
func (s *set) GcpFirewallRule() cloud.GcpFirewallRule {
	return &cloud.GcpFirewallRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// AwsSGRule return aws security group rule dao.
func (s *set) AwsSGRule() securitygroup.AwsSGRule {
	return &securitygroup.AwsSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// HuaWeiSGRule return huawei security group rule dao.
func (s *set) HuaWeiSGRule() securitygroup.HuaWeiSGRule {
	return &securitygroup.HuaWeiSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// AzureSGRule return azure security group rule dao.
func (s *set) AzureSGRule() securitygroup.AzureSGRule {
	return &securitygroup.AzureSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// Cvm return cvm dao.
func (s *set) Cvm() cvm.Interface {
	return &cvm.Dao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// TCloudRegion returns tcloud region dao.
func (s *set) TCloudRegion() region.TCloudRegion {
	return region.NewTCloudRegionDao(s.orm, s.idGen)
}

// AwsRegion returns aws region dao.
func (s *set) AwsRegion() region.AwsRegion {
	return region.NewAwsRegionDao(s.orm, s.idGen)
}

// GcpRegion returns gcp region dao.
func (s *set) GcpRegion() region.GcpRegion {
	return region.NewGcpRegionDao(s.orm, s.idGen)
}

// RouteTable returns route table dao.
func (s *set) RouteTable() routetable.RouteTable {
	return routetable.NewRouteTableDao(s.orm, s.idGen, s.audit)
}

// Route returns route dao.
func (s *set) Route() routetable.Route {
	return routetable.NewRouteDao(s.orm, s.idGen, s.audit)
}

// Audit return audit dao.
func (s *set) Audit() audit.Interface {
	return s.audit
}

// Application return application dao.
func (s *set) Application() application.Application {
	return &application.ApplicationDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// ApprovalProcess return application dao.
func (s *set) ApprovalProcess() application.ApprovalProcess {
	return &application.ApprovalProcessDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// NetworkInterface return network interface dao.
func (s *set) NetworkInterface() networkinterface.NetworkInterface {
	return &networkinterface.NetworkInterfaceDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// RecycleRecord return recycle record dao.
func (s *set) RecycleRecord() recyclerecord.RecycleRecord {
	return recyclerecord.NewRecycleRecordDao(s.orm, s.idGen, s.audit)
}

// Txn define dao set Txn.
type Txn struct {
	orm orm.Interface
}

// AutoTxn auto Txn.
func (t *Txn) AutoTxn(kt *kit.Kit, run orm.TxnFunc) (interface{}, error) {
	return t.orm.AutoTxn(kt, run)
}

// Txn return Txn.
func (s *set) Txn() *Txn {
	return &Txn{
		orm: s.orm,
	}
}

// AccountBillConfig returns account bill config dao.
func (s *set) AccountBillConfig() bill.Interface {
	return &bill.AccountBillConfigDao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}

// UserCollection returns user collection dao.
func (s *set) UserCollection() daouser.Interface {
	return &daouser.Dao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AsyncFlow return AsyncFlow dao.
func (s *set) AsyncFlow() daoasync.AsyncFlow {
	return &daoasync.AsyncFlowDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AsyncFlowTask return AsyncFlowTask dao.
func (s *set) AsyncFlowTask() daoasync.AsyncFlowTask {
	return &daoasync.AsyncFlowTaskDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// CloudSelectionScheme returns cloud selection scheme dao.
func (s *set) CloudSelectionScheme() daoselection.SchemeInterface {
	return &daoselection.SchemeDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// CloudSelectionBizType return cloud selection biz type dao.
func (s *set) CloudSelectionBizType() daoselection.BizTypeInterface {
	return &daoselection.BizTypeDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// CloudSelectionIdc return cloud selection idc dao.
func (s *set) CloudSelectionIdc() daoselection.IdcInterface {
	return &daoselection.IdcDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// ArgsTpl return argument template dao.
func (s *set) ArgsTpl() argstpl.Interface {
	return &argstpl.Dao{
		Orm:   s.orm,
		IDGen: s.idGen,
		Audit: s.audit,
	}
}
