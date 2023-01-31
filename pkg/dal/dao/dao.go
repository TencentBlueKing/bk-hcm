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

package dao

import (
	"fmt"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/auth"
	"hcm/pkg/dal/dao/cloud"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/metrics"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
)

// Set defines all the DAO to be operated.
type Set interface {
	Auth() auth.Auth
	Account() cloud.Account
	SecurityGroup() cloud.SecurityGroup
	SecurityGroupBizRel() cloud.SecurityGroupBizRel
	TCloudSGRule() cloud.TCloudSGRule
	AwsSGRule() cloud.AwsSGRule
	HuaWeiSGRule() cloud.HuaWeiSGRule
	AzureSGRule() cloud.AzureSGRule
	GcpFirewallRule() cloud.GcpFirewallRule
	Cloud() cloud.Cloud
	AccountBizRel() cloud.AccountBizRel
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

	auditDao, err := audit.NewAuditDao(ormInst, db)
	if err != nil {
		return nil, fmt.Errorf("new audit dao failed, err: %v", err)
	}

	s := &set{
		idGen:    idGen,
		orm:      ormInst,
		db:       db,
		auditDao: auditDao,
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
	idGen    idgenerator.IDGenInterface
	orm      orm.Interface
	db       *sqlx.DB
	auditDao audit.AuditDao
}

// Account return account dao.
func (s *set) Account() cloud.Account {
	return &cloud.AccountDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
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

// AccountBizRel return AccountBizRel dao.
func (s *set) AccountBizRel() cloud.AccountBizRel {
	return &cloud.AccountBizRelDao{
		Orm: s.orm,
	}
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

// SecurityGroup return security group dao.
func (s *set) SecurityGroup() cloud.SecurityGroup {
	return &cloud.SecurityGroupDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// SecurityGroupBizRel return security group and biz rel dao.
func (s *set) SecurityGroupBizRel() cloud.SecurityGroupBizRel {
	return &cloud.SecurityGroupBizRelDao{
		Orm: s.orm,
	}
}

// TCloudSGRule return tcloud security group rule dao.
func (s *set) TCloudSGRule() cloud.TCloudSGRule {
	return &cloud.TCloudSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// GcpFirewallRule return gcp firewall rule dao.
func (s *set) GcpFirewallRule() cloud.GcpFirewallRule {
	return &cloud.GcpFirewallRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AwsSGRule return aws security group rule dao.
func (s *set) AwsSGRule() cloud.AwsSGRule {
	return &cloud.AwsSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// HuaWeiSGRule return huawei security group rule dao.
func (s *set) HuaWeiSGRule() cloud.HuaWeiSGRule {
	return &cloud.HuaWeiSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}

// AzureSGRule return azure security group rule dao.
func (s *set) AzureSGRule() cloud.AzureSGRule {
	return &cloud.AzureSGRuleDao{
		Orm:   s.orm,
		IDGen: s.idGen,
	}
}
