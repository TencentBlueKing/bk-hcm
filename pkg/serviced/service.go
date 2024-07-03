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
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/logs"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// Service defines all the service and discovery
// related operations.
type Service interface {
	CurrentNodeKey() string
	// Register the service
	Register() error
	// Deregister the service
	Deregister() error
	State
}

// State defines the service's state related operations.
type State interface {
	// IsMaster test if this service instance is
	// master or not.
	IsMaster() bool
	// DisableMasterSlave disable/enable this service instance's master-slave check.
	// if disabled, treat this service as a slave instead of checking if it is master from service discovery.
	DisableMasterSlave(disable bool)
}

// NewService create a service instance.
func NewService(config cc.Service, opt ServiceOption) (Service, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	etcdOpt, err := config.Etcd.ToConfig()
	if err != nil {
		return nil, fmt.Errorf("get etcd config failed, err: %v", err)
	}

	cli, err := etcd3.New(etcdOpt)
	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		cli:    cli,
		svcOpt: opt,
		ctx:    ctx,
		cancel: cancel,
	}

	// keep synchronizing current node's master state.
	s.syncMasterState()
	return s, nil
}

type service struct {
	// key 当前节点的key
	key string

	cli    *etcd3.Client
	svcOpt ServiceOption

	// isRegisteredFlag service register flag.
	isRegisteredFlag  bool
	isRegisteredRWMux sync.RWMutex

	// leaseID is grant lease's id that used to put kv.
	leaseID      etcd3.LeaseID
	leaseIDRWMux sync.RWMutex

	// isMasterFlag service instance master state.
	isMasterFlag  bool
	isMasterRwMux sync.RWMutex

	// disableMasterSlaveFlag defines if the service instance's master-slave check is disabled and treated as slave.
	disableMasterSlaveFlag bool

	// watchChan is watch etcd service path's watch channel.
	watchChan etcd3.WatchChan

	ctx    context.Context
	cancel context.CancelFunc
}

// CurrentNodeKey return current node key.
func (s *service) CurrentNodeKey() string {
	return s.key
}

// Register the service
func (s *service) Register() error {
	if s.isRegistered() {
		return errors.New("current service is already registered, it can only register once")
	}

	// get service key and value.
	key := key(ServiceDiscoveryName(s.svcOpt.Name), s.svcOpt.Uid)
	s.key = key
	serverPath := url.URL{
		Scheme: s.svcOpt.Scheme,
		Host:   net.JoinHostPort(s.svcOpt.IP, strconv.Itoa(int(s.svcOpt.Port))),
	}
	value := serverPath.String()

	// grant lease, and put kv with lease.
	lease := etcd3.NewLease(s.cli)
	leaseResp, err := lease.Grant(s.ctx, defaultGrantLeaseTTL)
	if err != nil {
		logs.Errorf("grant lease failed, err: %v", err)
		return err
	}
	ctx, cancel := context.WithTimeout(s.ctx, defaultEtcdTimeout)
	defer cancel()
	_, err = s.cli.Put(ctx, key, value, etcd3.WithLease(leaseResp.ID))
	if err != nil {
		logs.Errorf("put kv with lease failed, key: %s, value: %s, err: %v", key, value, err)
		return err
	}
	s.updateLeaseID(leaseResp.ID)
	s.updateRegisterFlag(true)

	// start to keep alive lease.
	s.keepAlive(key, value)

	return nil
}

func (s *service) keepAlive(key string, value string) {
	go func() {
		lease := etcd3.NewLease(s.cli)
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				curLeaseID := s.getLeaseID()
				// if the current lease is 0, you need to lease the lease and use this put kv (bind lease).
				// if the lease is not 0, the put has been completed and the lease needs to be renewed.
				if curLeaseID == 0 {
					leaseResp, err := lease.Grant(s.ctx, defaultGrantLeaseTTL)
					if err != nil {
						logs.Errorf("grant lease failed, key: %s, err: %v", key, err)
						time.Sleep(defaultErrSleepTime)
						continue
					}

					ctx, cancel := context.WithTimeout(s.ctx, defaultEtcdTimeout)
					defer cancel()
					_, err = s.cli.Put(ctx, key, value, etcd3.WithLease(leaseResp.ID))
					if err != nil {
						logs.Errorf("put kv failed, key: %s, lease: %d, err: %v", key, leaseResp.ID, err)
						time.Sleep(defaultErrSleepTime)
						continue
					}

					s.updateRegisterFlag(true)
					s.updateLeaseID(leaseResp.ID)
				} else {
					// before keep alive, need to judge if service key exists.
					// if not exist, need to re-register.
					ctx, cancel := context.WithTimeout(s.ctx, defaultEtcdTimeout)
					defer cancel()
					resp, err := s.cli.Get(ctx, key)
					if err != nil {
						logs.Errorf("get key failed, lease: %d, err: %v", curLeaseID, err)
						s.keepAliveFailed()
						continue
					}
					if len(resp.Kvs) == 0 {
						logs.Warnf("current service key [%s, %s] is not exist, need to re-register", key, value)
						s.keepAliveFailed()
						continue
					}

					if _, err := lease.KeepAliveOnce(s.ctx, curLeaseID); err != nil {
						logs.Errorf("keep alive lease failed, lease: %d, err: %v", curLeaseID, err)
						s.keepAliveFailed()
						continue
					}
				}
				time.Sleep(defaultKeepAliveInterval)
			}
		}
	}()
}

// keepAliveFailed keep alive lease failed, need to exec action.
func (s *service) keepAliveFailed() {
	s.updateRegisterFlag(false)
	s.updateLeaseID(0)
	time.Sleep(defaultErrSleepTime)
}

// Deregister the service
func (s *service) Deregister() error {
	s.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdTimeout)
	defer cancel()
	if _, err := s.cli.Delete(ctx, key(ServiceDiscoveryName(s.svcOpt.Name),
		s.svcOpt.Uid)); err != nil {
		return err
	}

	s.updateRegisterFlag(false)
	return nil
}

// IsMaster test if this service instance is
// master or not.
func (s *service) IsMaster() bool {
	s.isMasterRwMux.RLock()
	defer s.isMasterRwMux.RUnlock()

	if s.disableMasterSlaveFlag {
		logs.Infof("master-slave is disabled, returns this service instance master state as slave")
		return false
	}
	return s.isMasterFlag
}

// DisableMasterSlave disable/enable this service instance's master-slave check.
// if disabled, treat this service as a slave instead of checking if it is master from service discovery.
func (s *service) DisableMasterSlave(disable bool) {
	s.isMasterRwMux.RLock()
	s.disableMasterSlaveFlag = disable
	s.isMasterRwMux.RUnlock()

	logs.Infof("master-slave disabled status: %v", disable)
	return
}

// HealthInfo is etcd health info, e.g. '{"health":"true"}'.
type HealthInfo struct {
	// Health is state flag, it's string not boolean.
	Health string `json:"health"`
}

// syncMasterState determine whether the current node is the primary node.
func (s *service) syncMasterState() {
	svrPath := ServiceDiscoveryName(s.svcOpt.Name)
	svrKey := key(svrPath, s.svcOpt.Uid)

	// watch service register path change event. if receive event, need to sync master state.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdTimeout)
		defer cancel()
		s.watchChan = s.cli.Watch(ctx, svrPath, etcd3.WithPrefix(), etcd3.WithPrevKV())

		for {
			resp, ok := <-s.watchChan
			// if the watchChan send is abnormally closed, you also need to finally
			// determine whether it is the master node.
			isMaster, err := s.isMaster(svrPath, svrKey)
			if err != nil {
				if logs.V(2) {
					logs.Errorf("sync service: %s master state failed, err: %v", s.svcOpt.Name, err)
				}
				time.Sleep(defaultErrSleepTime)
				continue
			}
			s.updateMasterFlag(isMaster)

			// if the abnormal pipe is closed, you need to retry watch
			if !ok || resp.Err() != nil {
				s.watchChan = s.cli.Watch(ctx, svrPath, etcd3.WithPrefix(),
					etcd3.WithPrevKV())
			}
		}
	}()

	// the bottom of the plan, sync master state regularly.
	go func() {
		for {
			isMaster, err := s.isMaster(svrPath, svrKey)
			if err != nil {
				if logs.V(2) {
					logs.Errorf("sync service: %s master state failed, err: %v", s.svcOpt.Name, err)
				}
				time.Sleep(defaultErrSleepTime)
				continue
			}
			s.updateMasterFlag(isMaster)

			time.Sleep(defaultSyncMasterInterval)
		}
	}()
}

// isMaster judge current service is master node.
func (s *service) isMaster(srvPath, srvKey string) (bool, error) {
	// get current instance version info.
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdTimeout)
	defer cancel()
	resp, err := s.cli.Get(ctx, srvKey, etcd3.WithPrefix(), etcd3.WithSerializable())
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, errors.New("current service not register, key: " + srvKey)
	}
	cr := resp.Kvs[0].CreateRevision

	// get first service instance version info.
	opts := etcd3.WithFirstCreate()
	opts = append(opts, etcd3.WithSerializable())
	resp, err = s.cli.Get(ctx, srvPath, opts...)
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, errors.New("current service not register, service path: " + srvPath)
	}
	firstCR := resp.Kvs[0].CreateRevision

	logs.V(7).Infof("current service(%s) master state: %v", srvKey, cr == firstCR)

	return cr == firstCR, nil
}

// updateMasterFlag update isMasterFlag by rw mux.
func (s *service) updateMasterFlag(isMaster bool) {
	s.isMasterRwMux.Lock()
	s.isMasterFlag = isMaster
	s.isMasterRwMux.Unlock()
	return
}

// updateRegisterFlag update isMasterFlag by rw mux.
func (s *service) updateRegisterFlag(isRegister bool) {
	s.isRegisteredRWMux.Lock()
	s.isRegisteredFlag = isRegister
	s.isRegisteredRWMux.Unlock()
	return
}

// updateLeaseID update leaseID by rw mux.
func (s *service) updateLeaseID(id etcd3.LeaseID) {
	s.leaseIDRWMux.Lock()
	s.leaseID = id
	s.leaseIDRWMux.Unlock()
	return
}

// isRegistered return is register flag by rw mux.
func (s *service) isRegistered() bool {
	s.isRegisteredRWMux.RLock()
	defer s.isRegisteredRWMux.RUnlock()
	return s.isRegisteredFlag
}

// getLeaseID return leaseID by rw mux.
func (s *service) getLeaseID() etcd3.LeaseID {
	s.leaseIDRWMux.RLock()
	defer s.leaseIDRWMux.RUnlock()
	return s.leaseID
}
