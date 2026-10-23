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
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/pkg/kit"
)

// TestConcurrentStress 并发申请、持有、取消、归还，检查：
//   - 每个 lease 持有期间，其进程始终配置成调用方的配置（不会被别人重配）；
//   - 运行中周期性校验不变式；
//   - 结束后没有泄漏的 lease、排队请求或准备中的进程；
//   - 无失败时进程数从不超过上限，也不会被关闭重建。
func TestConcurrentStress(t *testing.T) {
	const (
		workers    = 32
		iterations = 60
		maxProcs   = 3
	)
	configs := []string{"A", "B", "C", "D", "E"}

	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: maxProcs, MaxLeasesPerProcess: 2, MaxQueueLength: 1000,
		MaxQueueWait: 200 * time.Millisecond})

	stop := make(chan struct{})
	var checker sync.WaitGroup
	checker.Add(1)
	go func() {
		defer checker.Done()
		for {
			select {
			case <-stop:
				return
			case <-time.After(time.Millisecond):
				p.poolLock.Lock()
				err := p.checkInvariantsLocked()
				p.poolLock.Unlock()
				if err != nil {
					t.Errorf("invariant violated: %v", err)
					return
				}
			}
		}
	}()

	var ok, timeout, canceled atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))

			for j := 0; j < iterations; j++ {
				secretID := configs[rng.Intn(len(configs))]
				kt := kit.New()
				var cancel context.CancelFunc = func() {}
				if rng.Intn(5) == 0 {
					kt.Ctx, cancel = context.WithTimeout(kt.Ctx, time.Duration(rng.Intn(3))*time.Millisecond)
				}

				l, err := p.Acquire(kt, providerConfig(secretID))
				cancel()
				switch {
				case errors.Is(err, ErrAcquireTimeout):
					timeout.Add(1)
					continue
				case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
					canceled.Add(1)
					continue
				case err != nil:
					t.Errorf("unexpected acquire error: %v", err)
					return
				}
				ok.Add(1)

				if err := checkLeaseConfig(l, secretID); err != nil {
					t.Error(err)
				}
				time.Sleep(time.Duration(rng.Intn(500)) * time.Microsecond)
				if err := checkLeaseConfig(l, secretID); err != nil {
					t.Error(err)
				}
				l.Release()
			}
		}(int64(i + 1))
	}
	wg.Wait()
	close(stop)
	checker.Wait()

	waitSettled(t, p)
	st := p.stats()
	assert.Equal(t, 0, st.inflight, "no leaked lease")
	assert.Equal(t, 0, st.waiters, "no leaked waiter")
	assert.LessOrEqual(t, st.total(), maxProcs)
	assert.LessOrEqual(t, s.starts.Load(), int64(maxProcs), "processes are reused, never recreated")
	assert.Positive(t, ok.Load())
	t.Logf("ok=%d timeout=%d canceled=%d starts=%d configures=%d", ok.Load(), timeout.Load(), canceled.Load(),
		s.starts.Load(), s.configures.Load())
}

// checkLeaseConfig lease 所在进程当前的配置必须是调用方的配置。
func checkLeaseConfig(l *Lease, secretID string) error {
	resp, err := l.PlanResourceChange(kit.New(), nil)
	if err != nil {
		return fmt.Errorf("plan on lease failed: %v", err)
	}
	if id := string(resp.PlannedPrivate); !strings.HasSuffix(id, "/"+secretID) {
		return fmt.Errorf("lease for %s runs on process %s", secretID, id)
	}
	return nil
}

// TestConcurrentCloseUnderLoad 负载中关闭池：所有调用都要返回，关闭后没有残留进程与请求。
func TestConcurrentCloseUnderLoad(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.startFn = func(*fakeManager) error { time.Sleep(time.Millisecond); return nil }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxQueueLength: 1000})

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				l, err := p.Acquire(kit.New(), providerConfig(fmt.Sprintf("C%d", (i+j)%4)))
				if err != nil {
					if !errors.Is(err, ErrPoolClosed) {
						t.Errorf("unexpected error: %v", err)
					}
					return
				}
				l.Release()
			}
		}(i)
	}

	time.Sleep(5 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	require.NoError(t, p.Close(ctx))
	wg.Wait()

	assert.Equal(t, 0, p.stats().total())
	assert.Equal(t, 0, p.stats().waiters)
	for _, m := range s.allManagers() {
		assert.True(t, m.Exited(), "process %d left running", m.id)
	}
	assertInvariants(t, p)
}
