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

package suite

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	cloudserver "hcm/pkg/client/cloud-server"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/dal/table"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	restclient "hcm/pkg/rest/client"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/smartystreets/goconvey/convey"
)

type ClientSet struct {
	HCService    *hcservice.Client
	CloudService *cloudserver.Client
}

var clientSet ClientSet

var dbCfg dbConfig

type serverConfig struct {
	HcHost    string
	CLoudHost string
	DataHost  string
}

type dbConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

var db *sqlx.DB

func init() {
	var clientCfg serverConfig
	var concurrent int
	var sustainSeconds float64
	var totalRequest int64

	flag.StringVar(&clientCfg.DataHost, "data-host", "http://127.0.0.1:9600", "data http service address")
	flag.StringVar(&clientCfg.HcHost, "hc-host", "http://127.0.0.1:9601", "hc http server address")
	flag.StringVar(&clientCfg.CLoudHost, "cloud-host", "http://127.0.0.1:9602", "cloud http server address")
	flag.IntVar(&concurrent, "concurrent", 1000, "concurrent request during the load test.")
	flag.Float64Var(&sustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&totalRequest, "total-request", 0,
		"the load test total request,it has higher priority than SustainSeconds")
	flag.StringVar(&dbCfg.IP, "mysql-ip", "127.0.0.1", "mysql ip address")
	flag.Int64Var(&dbCfg.Port, "mysql-port", 3306, "mysql port")
	flag.StringVar(&dbCfg.User, "mysql-user", "hcm", "mysql login user")
	flag.StringVar(&dbCfg.Password, "mysql-passwd", "admin", "mysql login password")
	flag.StringVar(&dbCfg.DB, "mysql-db", "hcm_suite_test", "mysql database")
	testing.Init()
	flag.Parse()

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

	hcCap := &restclient.Capability{
		Client:     restCli,
		Discover:   serviced.NewSimpleDiscovery([]string{clientCfg.HcHost}),
		MetricOpts: restclient.MetricOption{Register: metrics.Register()},
	}
	hcCli := hcservice.NewClient(hcCap, "v1")

	cloudCap := &restclient.Capability{
		Client:     restCli,
		Discover:   serviced.NewSimpleDiscovery([]string{clientCfg.CLoudHost}),
		MetricOpts: restclient.MetricOption{Register: metrics.Register()},
	}
	cloudCli := cloudserver.NewClient(cloudCap, "v1")
	clientSet = ClientSet{HCService: hcCli, CloudService: cloudCli}
}

func ClearData() error {

	// 清空表然后重建，通过makefile实现
	tables := []table.Name{
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
func GetClientSet() *ClientSet {
	return &clientSet
}
