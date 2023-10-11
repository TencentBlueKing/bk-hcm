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

// Package suite 测试配置
package suite

import (
	"fmt"
	"log"
	"os"
	"testing"

	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/dal/table"
	"hcm/pkg/logs"
	restclient "hcm/pkg/rest/client"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"

	// mysql driver for clear table
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/pflag"
)

var clientSet *client.ClientSet

var dbCfg dbConfig

type dbConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

var db *sqlx.DB

func init() {
	var etcdCfg cc.Etcd
	var concurrent int
	var sustainSeconds float64
	var totalRequest int64

	flag := pflag.CommandLine

	// flag.ParseErrorsWhitelist
	flag.StringSliceVar(&etcdCfg.Endpoints, "etcd-endpoints", []string{"127.0.0.1:2379"},
		"etcd endpoints, use the normal service etcd.")
	flag.StringVar(&etcdCfg.Username, "etcd-username", "", "etcd username")
	flag.StringVar(&etcdCfg.Password, "etcd-password", "", "etcd password")

	flag.IntVar(&concurrent, "concurrent", 1000, "concurrent request during the load test.")
	flag.Float64Var(&sustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&totalRequest, "total-request", 0,
		"the load test total request,it has higher priority than SustainSeconds")
	flag.StringVar(&dbCfg.IP, "mysql-ip", "127.0.0.1", "mysql ip address")
	flag.Int64Var(&dbCfg.Port, "mysql-port", 3306, "mysql port")
	flag.StringVar(&dbCfg.User, "mysql-user", "hcm", "mysql login user")
	flag.StringVar(&dbCfg.Password, "mysql-passwd", "admin", "mysql login password")
	flag.StringVar(&dbCfg.DB, "mysql-db", "hcm_suite_test", "mysql database")

	pflagArgs := os.Args[1:]
	osArgs := os.Args
	// 将参数分为两部分
	for i, arg := range os.Args {
		// 双横线参数为分界线，前面为go 测试参数，后面为pflag参数，如果不处理掉后面的双横线参数会导致test参数解析报错
		if arg[0] == '-' && arg[1] == '-' {
			osArgs, pflagArgs = os.Args[0:i], os.Args[i:]
			break
		}
	}
	os.Args = osArgs

	// if err it will panic, there is no meaning handling this error
	_ = flag.Parse(pflagArgs)
	testing.Init()

	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=UTC",
		dbCfg.User, dbCfg.Password, dbCfg.IP, dbCfg.Port, dbCfg.DB)
	db = sqlx.MustConnect("mysql", dsn)
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(5)

	tls := &ssl.TLSConfig{}
	restCli, err := restclient.NewClient(tls)
	if err != nil {
		log.Printf("suite test new rest client err: %v", err)
		os.Exit(0)
	}

	// new api server discovery client.
	discOpt := serviced.DiscoveryOption{Services: []cc.Name{cc.CloudServerName, cc.HCServiceName, cc.DataServiceName}}
	dis, err := serviced.NewDiscovery(cc.Service{Etcd: etcdCfg}, discOpt)
	if err != nil {
		logs.Errorf("new service discovery failed, err: %v", err)
		panic(err)
	}

	logs.Infof("create discovery success.")

	clientSet = client.NewClientSet(restCli, dis)
}

// ClearData 清空数据表
func ClearData() error {

	tables := []table.Name{
		table.AccountTable,
		table.AccountSyncDetailTable,
		table.AuditTable,
		table.VpcTable,
		table.SubnetTable,
		table.RouteTableTable,
		table.TCloudRegionTable,
		table.ZoneTable,
	}
	for _, tableName := range tables {
		if _, err := db.Exec("truncate table " + string(tableName)); err != nil {
			logs.Errorf("fail to truncate table %s, err: %v", tableName, err)
			return err
		}
	}

	return nil
}

// GetClientSet get suite-test client set .
func GetClientSet() *client.ClientSet {
	return clientSet
}
