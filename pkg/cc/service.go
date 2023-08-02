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
)

// Setting defines all service Setting interface.
type Setting interface {
	trySetFlagBindIP(ip net.IP) error
	trySetDefault()
	Validate() error
}

// ApiServerSetting defines api server used setting options.
type ApiServerSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`
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

// CloudServerSetting defines cloud server used setting options.
type CloudServerSetting struct {
	Network       Network       `yaml:"network"`
	Service       Service       `yaml:"service"`
	Log           LogOption     `yaml:"log"`
	Crypto        Crypto        `yaml:"crypto"`
	Esb           Esb           `yaml:"esb"`
	BkHcmUrl      string        `yaml:"bkHcmUrl"`
	CloudResource CloudResource `yaml:"cloudResource"`
	Recycle       Recycle       `yaml:"recycle"`
	BillConfig    BillConfig    `yaml:"billConfig"`
	Itsm          ApiGateway    `yaml:"itsm"`
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
	if err := s.Esb.validate(); err != nil {
		return err
	}

	if s.BkHcmUrl == "" {
		return fmt.Errorf("bkHcmUrl should not be empty")
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

	return nil
}

// DataServiceSetting defines data service used setting options.
type DataServiceSetting struct {
	Network  Network   `yaml:"network"`
	Service  Service   `yaml:"service"`
	Log      LogOption `yaml:"log"`
	Database DataBase  `yaml:"database"`
	Crypto   Crypto    `yaml:"crypto"`
	Esb      Esb       `yaml:"esb"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *DataServiceSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetDefault set the DataServiceSetting default value if user not configured.
func (s *DataServiceSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Database.trySetDefault()

	return
}

// Validate DataServiceSetting option.
func (s DataServiceSetting) Validate() error {
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

	if err := s.Esb.validate(); err != nil {
		return err
	}

	return nil
}

// HCServiceSetting defines hc service used setting options.
type HCServiceSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`
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

	return nil
}

// AuthServerSetting defines auth server used setting options.
type AuthServerSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`
	Esb     Esb       `yaml:"esb"`

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

	if err := s.IAM.validate(); err != nil {
		return err
	}

	return nil
}

// WebServerSetting defines api server used setting options.
type WebServerSetting struct {
	Network Network    `yaml:"network"`
	Service Service    `yaml:"service"`
	Log     LogOption  `yaml:"log"`
	Web     Web        `yaml:"web"`
	Esb     Esb        `yaml:"esb"`
	Itsm    ApiGateway `yaml:"itsm"`
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

	if err := s.Itsm.validate(); err != nil {
		return err
	}

	return nil
}
