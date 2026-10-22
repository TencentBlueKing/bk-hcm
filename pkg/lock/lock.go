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

// Package lock 基于 etcd 的分布式锁：租约 + 事务，不续约，持有者崩溃后租约到期自动释放。
package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/pkg/cc"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// ErrLockFailed lock grabbing failed err
var ErrLockFailed = errors.New("lock grabbing failed")

// Manager 进程级默认锁管理器，服务启动时通过 InitManager 初始化。
var Manager *EtcdMutex

// InitManager 初始化进程级默认锁管理器，ttl 为默认租约时长（秒）。
func InitManager(cfg etcd3.Config, ttl int64) error {
	mux, err := New(cfg, ttl)
	if err != nil {
		return err
	}

	Manager = mux

	return nil
}

// AccountSyncKey 返回 cloud-server 账号同步使用的锁 key。
func AccountSyncKey(accountID string) string {
	return fmt.Sprintf("/hcm/lock/%s/sync/%s", cc.CloudServerName, accountID)
}

// IaCPluginKey 返回 IaC 插件注册、升级、启停共用的锁 key。
func IaCPluginKey(pluginName string) string {
	return fmt.Sprintf("/hcm/lock/iac_plugin/%s", pluginName)
}

// New create an etcd mutex, ttl is the default lease ttl in seconds.
func New(cfg etcd3.Config, ttl int64) (*EtcdMutex, error) {
	client, err := etcd3.New(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	return &EtcdMutex{
		ttl:    ttl,
		cli:    client,
		ctx:    ctx,
		cancel: cancelFunc,
		lease:  etcd3.NewLease(client),
	}, nil
}

// EtcdMutex 基于 etcd 的分布式锁。租约 ID 即持有者凭证：UnLock 只撤销自己的租约，不会影响其他持有者。
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

// TryLock try lock key with the default ttl, failed return error.
func (mux *EtcdMutex) TryLock(key string) (etcd3.LeaseID, error) {
	return mux.TryLockWithTTL(key, mux.ttl)
}

// TryLockWithTTL try lock key with the specified lease ttl(in seconds), failed return ErrLockFailed.
// 租约不续约，持有者崩溃后租约到期自动释放。
func (mux *EtcdMutex) TryLockWithTTL(key string, ttl int64) (etcd3.LeaseID, error) {
	grant, err := mux.lease.Grant(mux.ctx, ttl)
	if err != nil {
		return 0, err
	}

	txn := etcd3.NewKV(mux.cli).Txn(mux.ctx).If(etcd3.Compare(etcd3.CreateRevision(key), "=", 0)).
		Then(etcd3.OpPut(key, "", etcd3.WithLease(grant.ID))).Else()
	txnResp, err := txn.Commit()
	if err != nil {
		mux.revokeQuietly(grant.ID)
		return 0, err
	}

	if !txnResp.Succeeded {
		// 没拿到锁，租约没有用处，立即撤销而不是等它过期。
		mux.revokeQuietly(grant.ID)
		return 0, ErrLockFailed
	}

	return grant.ID, nil
}

// Lock 阻塞获取锁，直到拿到锁、ctx 取消或等待超过 maxWait（maxWait<=0 表示只受 ctx 控制）。
// 抢不到时监听该 key 的删除事件（持有者 UnLock 或租约到期都会触发），删除后再抢一次，不做轮询。
// 返回的 LeaseID 用于 UnLock。
func (mux *EtcdMutex) Lock(ctx context.Context, key string, ttl int64, maxWait time.Duration) (
	etcd3.LeaseID, error) {

	if maxWait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, maxWait)
		defer cancel()
	}

	for {
		leaseID, err := mux.TryLockWithTTL(key, ttl)
		if err == nil {
			return leaseID, nil
		}
		// 只有"锁被占用"才等待，其他错误（etcd 不可用等）交给调用方判断。
		if !errors.Is(err, ErrLockFailed) {
			return 0, err
		}

		if err = mux.waitKeyDeleted(ctx, key); err != nil {
			return 0, fmt.Errorf("wait lock failed: %w", err)
		}
	}
}

// waitKeyDeleted 阻塞直到 key 不存在或被删除。
func (mux *EtcdMutex) waitKeyDeleted(ctx context.Context, key string) error {
	resp, err := mux.cli.Get(ctx, key, etcd3.WithKeysOnly())
	if err != nil {
		return err
	}
	// 抢锁失败到这里的间隙里 key 已被释放，直接重试。
	if resp.Count == 0 {
		return nil
	}

	// 从 Get 的版本之后开始监听，不会漏掉中间发生的删除。
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	wch := mux.cli.Watch(watchCtx, key, etcd3.WithRev(resp.Header.Revision+1))
	for wresp := range wch {
		if err = wresp.Err(); err != nil {
			return err
		}
		for _, ev := range wresp.Events {
			if ev.Type == etcd3.EventTypeDelete {
				return nil
			}
		}
	}

	// watch 通道被关闭：ctx 已取消，或连接异常。
	if err = ctx.Err(); err != nil {
		return err
	}

	return errors.New("watch lock key closed unexpectedly")
}

// UnLock 撤销租约以释放锁。
func (mux *EtcdMutex) UnLock(leaseID etcd3.LeaseID) error {
	if _, err := mux.lease.Revoke(mux.ctx, leaseID); err != nil {
		return err
	}

	return nil
}

func (mux *EtcdMutex) revokeQuietly(leaseID etcd3.LeaseID) {
	// 加锁失败路径上的清理，失败只会让租约自然过期，不影响正确性。
	_, _ = mux.lease.Revoke(mux.ctx, leaseID)
}
