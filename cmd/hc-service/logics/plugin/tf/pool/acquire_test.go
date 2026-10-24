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

package pool

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
)

func TestAcquireCreatesConfiguresAndReuses(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2})

	l1 := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l1))
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 1, s.configures.Load())
	assertInvariants(t, p)

	l1.Release()
	assert.Equal(t, 0, p.testStats().inflight)

	l2 := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l2), "same config reuses the configured process")
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 1, s.configures.Load(), "no reconfigure for the same config")
	l2.Release()
	assertInvariants(t, p)
}

// TestStrictFIFOHeadBlocks 队首缺资源时，后面本可填空位的请求也要等（技术方案 4.3 的例子）。
func TestStrictFIFOHeadBlocks(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxLeasesPerProcess: 2})

	la := mustAcquire(t, p, "A")  // P1 = A，inflight 1
	lb1 := mustAcquire(t, p, "B") // P2 = B，inflight 1
	lb2 := mustAcquire(t, p, "B") // P2 = B，inflight 2（满）
	require.Equal(t, identityOf(t, lb1), identityOf(t, lb2))

	chC := acquireAsync(p, kit.New(), "C") // 队首：无匹配、无空闲、容量已满
	waitWaiters(t, p, 1)
	chA := acquireAsync(p, kit.New(), "A") // P1 有空位，但不能越过队首
	waitWaiters(t, p, 2)
	assertPending(t, chA)

	lb1.Release() // P2 仍有在途 lease，不能重配
	assertPending(t, chC)
	assertPending(t, chA)

	lb2.Release() // P2 空闲 → 队首 C 重配 P2，随后 A 落到 P1
	rc := waitResult(t, chC)
	ra := waitResult(t, chA)
	require.NoError(t, rc.err)
	require.NoError(t, ra.err)
	assert.Equal(t, identityOf(t, la), identityOf(t, ra.lease))
	assert.True(t, strings.HasSuffix(identityOf(t, rc.lease), "/C"))

	la.Release()
	rc.lease.Release()
	ra.lease.Release()
	waitSettled(t, p)
}

func TestCanceledHeadUnblocksNext(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2})

	la := mustAcquire(t, p, "A")

	ctx, cancel := context.WithCancel(context.Background())
	kt := kit.New()
	kt.Ctx = ctx
	chB := acquireAsync(p, kt, "B") // 队首，等不到资源
	waitWaiters(t, p, 1)
	chA := acquireAsync(p, kit.New(), "A")
	waitWaiters(t, p, 2)
	assertPending(t, chA)

	cancel()
	rb := waitResult(t, chB)
	assert.ErrorIs(t, rb.err, context.Canceled)

	ra := waitResult(t, chA)
	require.NoError(t, ra.err, "cancel of the head lets the next request through")
	assert.Equal(t, identityOf(t, la), identityOf(t, ra.lease))

	la.Release()
	ra.lease.Release()
	assert.Equal(t, 0, p.testStats().waiters)
	assertInvariants(t, p)
}

func TestQueueFullAndTimeout(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxQueueLength: 1, MaxQueueWait: 100 * time.Millisecond})

	la := mustAcquire(t, p, "A")

	chB := acquireAsync(p, kit.New(), "B")
	waitWaiters(t, p, 1)

	_, err := p.Acquire(kit.New(), providerConfig("C"))
	assert.ErrorIs(t, err, ErrTooManyWaiters)

	rb := waitResult(t, chB)
	assert.ErrorIs(t, rb.err, ErrAcquireTimeout)
	assert.Equal(t, 0, p.testStats().waiters, "timed out request leaves the queue")

	la.Release()
	assertInvariants(t, p)
}

func TestLeaseReleaseIsIdempotent(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2})

	l1 := mustAcquire(t, p, "A")
	l2 := mustAcquire(t, p, "A")
	assert.Equal(t, 2, p.testStats().inflight)

	l1.Release()
	l1.Release()
	assert.Equal(t, 1, p.testStats().inflight, "double release must not double decrement")

	_, err := l1.PlanResourceChange(kit.New(), &tfruntime.PlanResourceChangeRequest{})
	assert.ErrorIs(t, err, ErrLeaseReleased)
	_, err = l1.ApplyResourceChange(kit.New(), &tfruntime.ApplyResourceChangeRequest{})
	assert.ErrorIs(t, err, ErrLeaseReleased)

	l2.Release()
	assert.Equal(t, 0, p.testStats().inflight)
	assertInvariants(t, p)
}

// TestCancelAfterDeliveryReturnsLease 调用方取消时 lease 已经写进请求但还没被取走：取消方负责把名额还回去。
// 这个竞争在真实并发里很难稳定复现，这里直接停在两步之间。
func TestCancelAfterDeliveryReturnsLease(t *testing.T) {
	p := newTestPool(t, &fakeStarter{}, PoolConfig{MaxProcesses: 1})
	config := providerConfig("A")
	key, err := configKey(config)
	require.NoError(t, err)

	w, err := p.enqueue(kit.New(), config, key)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return len(w.ch) == 1 }, waitTimeout, time.Millisecond)
	assert.Equal(t, 1, p.testStats().inflight, "lease is delivered but nobody took it")

	p.cancel(w)

	assert.Equal(t, 0, p.testStats().inflight, "the undelivered lease is released by cancel")
	assertInvariants(t, p)
}
