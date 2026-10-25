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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseTime = time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)

// procSpec 手工构造的进程，idle 为空闲起始相对 baseTime 的偏移，boundAgo 为配置生效距 now 的时长。
type procSpec struct {
	id       uint64
	state    procState
	key      string // 空表示未配置
	boundAgo time.Duration
	inflight int
	pending  int
	idle     time.Duration
}

func pickPool(t *testing.T, maxProcs, maxLeases int, specs ...procSpec) *PluginPool {
	t.Helper()

	// 只测选进程规则，不经过 NewPluginPool：它会启动后台协程预热进程，与这里手工构造的进程表互相干扰。
	p := &PluginPool{
		cfg:   PoolConfig{MaxProcesses: maxProcs, MaxLeasesPerProcess: maxLeases, MaxConfigAge: 30 * time.Minute},
		procs: make(map[uint64]*pluginProcess),
	}

	now := baseTime.Add(time.Hour)
	for _, s := range specs {
		proc := &pluginProcess{id: s.id, state: s.state, inflight: s.inflight, idleSince: baseTime.Add(s.idle)}
		if s.pending > 0 {
			proc.pending = make([]*request, s.pending)
			for i := range proc.pending {
				proc.pending[i] = &request{state: requestPending, ch: make(chan result, 1)}
			}
		}
		if s.key != "" {
			proc.binding = &binding{key: s.key}
			if s.state == stateRunning {
				proc.binding.boundAt = now.Add(-s.boundAgo)
			}
		}
		p.procs[s.id] = proc
	}
	return p
}

func TestPickRules(t *testing.T) {
	now := baseTime.Add(time.Hour)
	fresh := time.Minute
	expired := time.Hour

	cases := []struct {
		name      string
		maxProcs  int
		maxLeases int
		procs     []procSpec
		wantAct   action
		wantProc  uint64
	}{
		{
			name: "empty pool creates", maxProcs: 2, maxLeases: 1,
			wantAct: actionCreate,
		},
		{
			name: "rule 1 matched with free slot", maxProcs: 2, maxLeases: 2,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "B", boundAgo: fresh},
				{id: 2, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionDeliver, wantProc: 2,
		},
		{
			name: "rule 1 prefers lower inflight", maxProcs: 3, maxLeases: 3,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "A", boundAgo: fresh, inflight: 2},
				{id: 2, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionDeliver, wantProc: 2,
		},
		{
			name: "rule 1 prefers newer boundAt on same inflight", maxProcs: 3, maxLeases: 3,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "A", boundAgo: 2 * time.Minute, inflight: 1},
				{id: 2, state: stateRunning, key: "A", boundAgo: time.Minute, inflight: 1},
			},
			wantAct: actionDeliver, wantProc: 2,
		},
		{
			name: "rule 1 tie breaks by id", maxProcs: 3, maxLeases: 3,
			procs: []procSpec{
				{id: 3, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
				{id: 2, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionDeliver, wantProc: 2,
		},
		{
			name: "rule 1 beats idle unbound", maxProcs: 3, maxLeases: 2,
			procs: []procSpec{
				{id: 1, state: stateRunning},
				{id: 2, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionDeliver, wantProc: 2,
		},
		{
			name: "full matched process is not delivered", maxProcs: 1, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "A", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionHold,
		},
		{
			name: "expired match falls to rule 2 on the same process", maxProcs: 1, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "A", boundAgo: expired},
			},
			wantAct: actionConfigure, wantProc: 1,
		},
		{
			name: "expired busy process is neither delivered nor reconfigured", maxProcs: 1, maxLeases: 2,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "A", boundAgo: expired, inflight: 1},
			},
			wantAct: actionHold,
		},
		{
			name: "rule 2 prefers unbound over other config", maxProcs: 2, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "B", boundAgo: fresh, idle: -time.Hour},
				{id: 2, state: stateRunning},
			},
			wantAct: actionConfigure, wantProc: 2,
		},
		{
			name: "rule 2 prefers longest idle", maxProcs: 2, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateRunning, idle: time.Minute},
				{id: 2, state: stateRunning, idle: 0},
			},
			wantAct: actionConfigure, wantProc: 2,
		},
		{
			name: "rule 3 reconfigures longest idle other config", maxProcs: 2, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "B", boundAgo: fresh, idle: time.Minute},
				{id: 2, state: stateRunning, key: "C", boundAgo: fresh, idle: 0},
			},
			wantAct: actionConfigure, wantProc: 2,
		},
		{
			name: "busy process with other config is not reconfigured", maxProcs: 1, maxLeases: 2,
			procs: []procSpec{
				{id: 1, state: stateRunning, key: "B", boundAgo: fresh, inflight: 1},
			},
			wantAct: actionHold,
		},
		{
			name: "rule 1b reserves configuring same key when slot free", maxProcs: 2, maxLeases: 2,
			procs: []procSpec{
				{id: 1, state: stateConfiguring, key: "A", pending: 1},
				{id: 2, state: stateRunning},
			},
			wantAct: actionReserve, wantProc: 1,
		},
		{
			name: "rule 1b prefers fewer pending on tie", maxProcs: 2, maxLeases: 3,
			procs: []procSpec{
				{id: 1, state: stateStarting, key: "A", pending: 2},
				{id: 2, state: stateConfiguring, key: "A", pending: 1},
			},
			wantAct: actionReserve, wantProc: 2,
		},
		{
			name: "rule 1b full pending falls through to rule 4 create", maxProcs: 2, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateStarting, key: "A", pending: 1},
			},
			wantAct: actionCreate,
		},
		{
			name: "closing process occupies capacity", maxProcs: 1, maxLeases: 1,
			procs: []procSpec{
				{id: 1, state: stateClosing},
			},
			wantAct: actionHold,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := pickPool(t, c.maxProcs, c.maxLeases, c.procs...)

			proc, act := p.pickLocked(now, "A")
			assert.Equal(t, c.wantAct, act)
			if c.wantProc == 0 {
				assert.Nil(t, proc)
				return
			}
			require.NotNil(t, proc)
			assert.Equal(t, c.wantProc, proc.id)
		})
	}
}

func TestBindingValid(t *testing.T) {
	now := baseTime
	age := time.Minute

	var nilBinding *binding
	assert.False(t, nilBinding.valid(now, age))
	assert.False(t, nilBinding.configuring())

	configuring := &binding{key: "A"}
	assert.True(t, configuring.configuring())
	assert.False(t, configuring.valid(now, age), "configuring binding is not effective")

	effective := &binding{key: "A", boundAt: now.Add(-30 * time.Second)}
	assert.False(t, effective.configuring())
	assert.True(t, effective.valid(now, age))

	boundary := &binding{key: "A", boundAt: now.Add(-age)}
	assert.False(t, boundary.valid(now, age), "expires exactly at maxConfigAge")
}
