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

// Package lock ...
package lock

import (
	"context"
	"errors"
	"fmt"

	"hcm/pkg/cc"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// ErrLockFailed lock grabbing failed err
var ErrLockFailed = errors.New("lock grabbing failed")

// Manager lock manager.
var Manager *EtcdMutex

// InitManger init lock manager.
func InitManger(cfg etcd3.Config, ttl int64) error {

	client, err := etcd3.New(cfg)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	Manager = &EtcdMutex{
		ttl:    ttl,
		cli:    client,
		ctx:    ctx,
		cancel: cancelFunc,
		lease:  etcd3.NewLease(client),
	}

	return nil
}

// EtcdMutex ...
type EtcdMutex struct {
	ttl   int64
	cli   *etcd3.Client
	lease etcd3.Lease

	ctx    context.Context
	cancel context.CancelFunc
}

// Close ...
func (mux *EtcdMutex) Close() {
	mux.cancel()
}

// TryLock try lock key, failed return error.
func (mux *EtcdMutex) TryLock(key string) (etcd3.LeaseID, error) {
	grant, err := mux.lease.Grant(mux.ctx, mux.ttl)
	if err != nil {
		return 0, err
	}

	txn := etcd3.NewKV(mux.cli).Txn(mux.ctx).If(etcd3.Compare(etcd3.CreateRevision(key), "=", 0)).
		Then(etcd3.OpPut(key, "", etcd3.WithLease(grant.ID))).Else()
	txnResp, err := txn.Commit()
	if err != nil {
		return 0, err
	}

	if !txnResp.Succeeded {
		return 0, ErrLockFailed
	}

	return grant.ID, nil
}

// UnLock ...
func (mux *EtcdMutex) UnLock(leaseID etcd3.LeaseID) error {
	if _, err := mux.lease.Revoke(mux.ctx, leaseID); err != nil {
		return err
	}

	return nil
}

// Key return key.
func Key(accountID string) string {
	return fmt.Sprintf("/hcm/lock/%s/sync/%s", cc.CloudServerName, accountID)
}
