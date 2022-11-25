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

package serviced

import (
	"errors"
	"fmt"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/tools/uuid"
)

const (
	// defaultKeepAliveInterval service key lease keep alive interval.
	defaultKeepAliveInterval = 5 * time.Second
	// defaultSyncMasterInterval sync master interval.
	defaultSyncMasterInterval = 10 * time.Second
	// defaultGrantLeaseTTL etcd lease ttl.
	defaultGrantLeaseTTL = 10
	// defaultErrSleepTime is exec failed need to wait time.
	defaultErrSleepTime = time.Second
)

// ServiceOption defines a service related options.
type ServiceOption struct {
	Name   cc.Name
	IP     string
	Port   uint
	Scheme string
	// Uid is a service's unique identity.
	Uid string
}

// Validate the service option
func (so ServiceOption) Validate() error {
	if len(so.Name) == 0 {
		return errors.New("service name is empty")
	}

	if len(so.IP) == 0 || so.IP == "0.0.0.0" {
		return errors.New("invalid service ip")
	}

	if so.Port == 0 {
		return errors.New("invalid service port")
	}

	if len(so.Uid) == 0 {
		return errors.New("invalid service uid")
	}

	return nil
}

// NewServiceOption generate a service option.
func NewServiceOption(name cc.Name, network cc.Network) ServiceOption {
	opt := ServiceOption{
		Name: name,
		IP:   network.BindIP,
		Port: network.Port,
		Uid:  uuid.UUID(),
	}

	if network.TLS.Enable() {
		opt.Scheme = "https"
	} else {
		opt.Scheme = "http"
	}

	return opt
}

// DiscoveryOption defines all the service discovery
// related options.
type DiscoveryOption struct {
	// Services defines all the services to discover
	Services []cc.Name
}

// Validate the service option
func (so DiscoveryOption) Validate() error {
	if len(so.Services) == 0 {
		return errors.New("services is empty")
	}

	return nil
}

// ServiceDiscoveryName return the service's register path in etcd.
func ServiceDiscoveryName(serviceName cc.Name) string {
	return fmt.Sprintf("/hcm/services/%s", serviceName)
}

// key return service's register key in etcd.
// e.g: /hcm/services/data-service/0fa709f2-8e35-11ec-83f6-acde48001122
func key(path, uid string) string {
	return fmt.Sprintf("%s/%s", path, uid)
}
