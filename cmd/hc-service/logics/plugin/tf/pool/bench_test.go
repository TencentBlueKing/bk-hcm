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
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/pkg/kit"
)

func BenchmarkScheduleLocked(b *testing.B) {
	for _, tc := range []struct {
		name      string
		maxProcs  int
		maxLeases int
	}{
		{"procs_4_leases_1", 4, 1},
		{"procs_16_leases_4", 16, 4},
	} {
		b.Run(tc.name, func(b *testing.B) {
			p, err := NewPluginPool(Options{
				PluginName: "tcloud",
				ExecPath:   "/bin/fake",
				Starter:    &fakeStarter{},
				Config: PoolConfig{
					MaxProcesses:        tc.maxProcs,
					MaxLeasesPerProcess: tc.maxLeases,
					MinIdleProcesses:    1,
					MaxQueueLength:      10000,
					PoolCheckInterval:   time.Hour,
				},
			})
			if err != nil {
				b.Fatal(err)
			}
			defer func() { _ = p.Close(p.taskCtx) }()

			now := baseTime.Add(time.Hour)
			p.now = func() time.Time { return now }
			for i := 1; i <= tc.maxProcs; i++ {
				key := fmt.Sprintf("K%d", i)
				p.procs[uint64(i)] = &pluginProcess{
					id:       uint64(i),
					state:    stateRunning,
					binding:  &binding{key: key, boundAt: now.Add(-time.Minute)},
					inflight: tc.maxLeases - 1,
					exec:     &fakeExecutor{m: &fakeManager{id: i, closedCh: make(chan struct{})}},
				}
			}

			targetKey := fmt.Sprintf("K%d", tc.maxProcs)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := &request{key: targetKey, kt: kit.New(), enqueuedAt: now, state: requestQueued, ch: make(chan result, 1)}
				p.queue.PushBack(w)
				p.scheduleLocked()
				p.procs[uint64(tc.maxProcs)].inflight = tc.maxLeases - 1
			}
		})
	}
}

func BenchmarkQueuePeekWithWaiters(b *testing.B) {
	for _, qLen := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("waiters_%d", qLen), func(b *testing.B) {
			q := newQueue[*request]()
			for i := 0; i < qLen; i++ {
				q.PushBack(&request{key: "K"})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = q.Peek()
			}
		})
	}
}

// 对数正态：中位数是 median，sigma 是 ln(耗时) 的标准差。
// 约 p95 = median * exp(1.645*sigma)，p99 = median * exp(2.326*sigma)。
//
// 取值按本地拉起 terraform 插件、不把云 API 往返算进插件进程本身：
//   - 拉起（起进程 + 握手）：中位 400ms，σ=0.55 → p95≈1.0s，p99≈1.4s
//   - 配置（ConfigureProvider）：中位 80ms，σ=0.45 → p95≈170ms，p99≈230ms
//   - 关闭（通知退出并等进程结束）：中位 50ms，σ=0.40 → p95≈95ms，p99≈130ms
//   - lease 上的一次 Apply（打云 API）：中位 250ms，σ=0.65 → p95≈730ms，p99≈1.1s
//
// 关闭不在请求路径上，只发生在回收和池关闭时。
var (
	latencyStart     = latencyDist{median: 400 * time.Millisecond, sigma: 0.55}
	latencyConfigure = latencyDist{median: 80 * time.Millisecond, sigma: 0.45}
	latencyClose     = latencyDist{median: 50 * time.Millisecond, sigma: 0.40}
	latencyCall      = latencyDist{median: 250 * time.Millisecond, sigma: 0.65}
)

// latencyDist 是耗时的对数正态分布。
type latencyDist struct {
	median time.Duration
	sigma  float64
}

func (d latencyDist) sample(rng *rand.Rand) time.Duration {
	mu := math.Log(float64(d.median))
	v := time.Duration(math.Exp(mu + d.sigma*rng.NormFloat64()))
	if v < time.Millisecond {
		return time.Millisecond
	}
	return v
}

// latencySampler 给插件回调用的采样器。回调在不同协程里执行，rng 不能并发读。
type latencySampler struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func newLatencySampler(seed int64) *latencySampler {
	return &latencySampler{rng: rand.New(rand.NewSource(seed))}
}

func (s *latencySampler) duration(d latencyDist) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return d.sample(s.rng)
}

// TestRequestThroughput 在拉起、配置、关闭和 lease 业务调用都有随机耗时的情况下，
// 测量冷启动请求、稳态复用、以及配置种类多于进程数时的完成时间和吞吐。
func TestRequestThroughput(t *testing.T) {
	lat := newLatencySampler(1)
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.startFn = func(*fakeManager) error {
			time.Sleep(lat.duration(latencyStart))
			return nil
		}
		s.configureFn = func(*fakeManager, map[string]any) error {
			time.Sleep(lat.duration(latencyConfigure))
			return nil
		}
		s.closeFn = func(*fakeManager, context.Context) error {
			time.Sleep(lat.duration(latencyClose))
			return nil
		}
	})

	const (
		maxProcs  = 4
		maxLeases = 4
		workers   = 16
		loadFor   = 2 * time.Second
	)
	p := newTestPool(t, s, PoolConfig{
		MaxProcesses:        maxProcs,
		MaxLeasesPerProcess: maxLeases,
		MinIdleProcesses:    1,
		MaxQueueLength:      1000,
		MaxQueueWait:        5 * time.Second,
		PoolCheckInterval:   time.Hour,
		IdleReclaimTimeout:  time.Hour,
	})

	coldKeys := []string{"A", "B", "C", "D"}
	cold := runOnceEach(t, p, coldKeys)
	logLatency(t, "cold (start+configure+call)", cold.total, cold.acquire, cold.errs, cold.wall)

	steadyKeys := coldKeys
	steady := runLoad(t, p, steadyKeys, workers, loadFor)
	logLatency(t, "steady reuse", steady.total, steady.acquire, steady.errs, steady.wall)

	mixedKeys := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	mixed := runLoad(t, p, mixedKeys, workers, loadFor)
	logLatency(t, "mixed reconfigure", mixed.total, mixed.acquire, mixed.errs, mixed.wall)

	require.NotEmpty(t, cold.total)
	require.NotEmpty(t, steady.total)
	require.NotEmpty(t, mixed.total)
	assert.Zero(t, cold.errs)
	assert.Zero(t, steady.errs)
	assert.Zero(t, mixed.errs)
	waitSettled(t, p)
	assert.Equal(t, 0, p.testStats().inflight)
}

type loadResult struct {
	total   []time.Duration
	acquire []time.Duration
	errs    int
	wall    time.Duration
}

// runOnceEach 每个 key 一个请求，并行发出，用来看冷启动。
func runOnceEach(t *testing.T, p *PluginPool, keys []string) loadResult {
	t.Helper()

	var mu sync.Mutex
	var wg sync.WaitGroup
	out := loadResult{}
	start := time.Now()
	for i, key := range keys {
		wg.Add(1)
		go func(seed int64, key string) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			total, acquire, err := oneRequest(p, key, rng)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				out.errs++
				t.Errorf("cold acquire %s: %v", key, err)
				return
			}
			out.total = append(out.total, total)
			out.acquire = append(out.acquire, acquire)
		}(int64(100+i), key)
	}
	wg.Wait()
	out.wall = time.Since(start)
	return out
}

// runLoad 在 wall 时间内尽量多打请求。workers 大于名额数，多出来的请求在队列里等。
func runLoad(t *testing.T, p *PluginPool, keys []string, workers int, dur time.Duration) loadResult {
	t.Helper()

	var mu sync.Mutex
	var wg sync.WaitGroup
	out := loadResult{}
	deadline := time.Now().Add(dur)
	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			for time.Now().Before(deadline) {
				key := keys[rng.Intn(len(keys))]
				total, acquire, err := oneRequest(p, key, rng)
				mu.Lock()
				if err != nil {
					out.errs++
					t.Errorf("acquire %s: %v", key, err)
				} else {
					out.total = append(out.total, total)
					out.acquire = append(out.acquire, acquire)
				}
				mu.Unlock()
			}
		}(int64(1000 + i))
	}
	wg.Wait()
	out.wall = time.Since(start)
	return out
}

// oneRequest 完成一次请求：拿到 lease、按业务调用的分布占用它、归还。
// 返回整段耗时和其中 Acquire 的耗时。
func oneRequest(p *PluginPool, key string, rng *rand.Rand) (total, acquire time.Duration, err error) {
	start := time.Now()
	l, err := p.Acquire(kit.New(), providerConfig(key))
	acquire = time.Since(start)
	if err != nil {
		return time.Since(start), acquire, err
	}
	time.Sleep(latencyCall.sample(rng))
	if _, callErr := l.PlanResourceChange(kit.New(), nil); callErr != nil {
		l.Release()
		return time.Since(start), acquire, callErr
	}
	l.Release()
	return time.Since(start), acquire, nil
}

func logLatency(t *testing.T, name string, total, acquire []time.Duration, errs int, wall time.Duration) {
	t.Helper()
	n := len(total)
	throughput := 0.0
	if wall > 0 {
		throughput = float64(n) / wall.Seconds()
	}
	t.Logf("%s: n=%d errs=%d wall=%s throughput=%.1f req/s e2e p50=%s p95=%s p99=%s acquire p50=%s p95=%s",
		name, n, errs, wall.Truncate(time.Millisecond), throughput,
		percentile(total, 0.50), percentile(total, 0.95), percentile(total, 0.99),
		percentile(acquire, 0.50), percentile(acquire, 0.95))
}

func percentile(samples []time.Duration, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Round(p * float64(len(sorted)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx].Truncate(time.Millisecond)
}
