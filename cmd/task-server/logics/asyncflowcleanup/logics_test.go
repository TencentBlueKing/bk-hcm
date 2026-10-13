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

package asyncflowcleanup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/stretchr/testify/assert"
)

// taskServerSetting 供各用例改写清理配置，InitRuntime 持有其指针，改字段即改 cc.TaskServer() 的返回值。
var taskServerSetting = new(cc.TaskServerSetting)

func TestMain(m *testing.M) {
	cc.InitRuntime(taskServerSetting)
	m.Run()
}

// setCleanupConfig 设置本次用例使用的清理配置。
func setCleanupConfig(enabled bool) {
	taskServerSetting.AsyncFlowAndTaskCleanup = cc.AsyncFlowAndTaskCleanup{
		Enabled:         cvt.ValToPtr(enabled),
		IntervalMin:     cvt.ValToPtr(constant.DefaultAsyncFlowCleanupIntervalMin),
		RetentionDays:   cvt.ValToPtr(constant.DefaultAsyncFlowCleanupRetentionDays),
		BatchIntervalMs: cvt.ValToPtr(constant.DefaultAsyncFlowCleanupBatchIntervalMs),
	}
}

// stubTenantLister 返回预设租户列表，用于在不依赖 data-service 的情况下驱动跨租户清理。
type stubTenantLister struct {
	tenantIDs []string
	err       error
	callCount int
}

// ListAllTenantIDs 返回预设租户列表。
func (s *stubTenantLister) ListAllTenantIDs(kt *kit.Kit) ([]string, error) {
	s.callCount++
	return s.tenantIDs, s.err
}

// newTestLogics 组装注入桩实现的 Logics，默认单租户 default（等价于未开启多租户的部署）。
func newTestLogics(store flowStore, tenantIDs ...string) *Logics {
	if len(tenantIDs) == 0 {
		tenantIDs = []string{constant.DefaultTenantID}
	}

	return newTestLogicsWithTenants(store, &stubTenantLister{tenantIDs: tenantIDs})
}

// newTestLogicsWithTenants 组装注入自定义租户桩的 Logics。
// 必须经由本函数构造而不是直接写结构体字面量，否则 startCursors 为 nil，记录起点时会 panic。
func newTestLogicsWithTenants(store flowStore, tenants tenantLister) *Logics {
	return &Logics{store: store, tenants: tenants, startCursors: make(map[string]string)}
}

// stubStore 记录调用次数并按预设次序返回结果，用于在不连库的情况下驱动清理主流程。
// pages 与 scanWindows 都用尽即视为查不到更多数据，因此不关心某一路的用例可以留空。
type stubStore struct {
	// pages 按调用次序返回给 ListExpiredFlowIDs 的每批 flow id
	pages [][]string
	// scanWindows 按调用次序返回给 ScanFlowsAfter 的每个扫描窗口
	scanWindows [][]flowBrief

	listCount int
	scanCount int
	delCount  int
	delFlowID []string
	// listTenantIDs 逐次记录查询超期 flow 时 kit 携带的租户，用于断言清理按租户逐个执行
	listTenantIDs []string
	// listCursors 逐次记录查询超期 flow 时传入的游标，用于断言水位推进
	listCursors []string
	// scanCursors 逐次记录定位扫描传入的游标，用于断言定位过程向后推进
	scanCursors []string
	// onScan 在每次定位扫描时回调，用于构造并发等待
	onScan func()
	// onDelete 在每次 Delete 调用后回调，用于构造批间中断
	onDelete func()
}

// ListExpiredFlowIDs 按调用次序返回预设批次。
func (s *stubStore) ListExpiredFlowIDs(kt *kit.Kit, cutoff, lastFlowID string, limit uint) ([]string, error) {
	s.listCount++
	s.listTenantIDs = append(s.listTenantIDs, kt.TenantID)
	s.listCursors = append(s.listCursors, lastFlowID)
	idx := s.listCount - 1

	if idx >= len(s.pages) {
		return nil, nil
	}

	return s.pages[idx], nil
}

// ScanFlowsAfter 按调用次序返回预设扫描窗口。
func (s *stubStore) ScanFlowsAfter(kt *kit.Kit, lastFlowID string, limit uint) ([]flowBrief, error) {
	s.scanCount++
	s.scanCursors = append(s.scanCursors, lastFlowID)
	idx := s.scanCount - 1

	if s.onScan != nil {
		s.onScan()
	}

	if idx >= len(s.scanWindows) {
		return nil, nil
	}

	return s.scanWindows[idx], nil
}

// DeleteFlowsWithTasks 记录被删除的 flow id，返回固定的 task 条数。
func (s *stubStore) DeleteFlowsWithTasks(kt *kit.Kit, flowIDs []string) (int, error) {
	s.delCount++
	s.delFlowID = append(s.delFlowID, flowIDs...)

	if s.onDelete != nil {
		s.onDelete()
	}

	return len(flowIDs) * 2, nil
}

func newFlowIDs(prefix string, count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = fmt.Sprintf("%s-%04d", prefix, i)
	}

	return ids
}

// newHitRow 构造一条 name 命中的记录，即定位起点时要停下的目标。
func newHitRow(id string) flowBrief {
	return flowBrief{ID: id, Name: cleanupFlowNames[0]}
}

// newOtherNameRow 构造一条 name 不命中的记录，模拟永远不会被清理、只能被空扫过去的残留数据。
func newOtherNameRow(id string) flowBrief {
	return flowBrief{ID: id, Name: enumor.FlowBillDailySummary}
}

// newOtherNameWindow 构造一个满窗、全是 name 不命中记录的扫描窗口。
func newOtherNameWindow(prefix string) []flowBrief {
	rows := make([]flowBrief, scanWindow)
	for i, id := range newFlowIDs(prefix, int(scanWindow)) {
		rows[i] = newOtherNameRow(id)
	}

	return rows
}

// TestBuildExpiredFlowFilter 清理批次的过滤条件由 name、updated_at、id 三条规则组成。
func TestBuildExpiredFlowFilter(t *testing.T) {
	expr := buildExpiredFlowFilter("2024-01-01T00:00:00+08:00", "0000001")

	assert.Equal(t, filter.And, expr.Op)
	assert.Len(t, expr.Rules, 3)

	fields := make([]string, len(expr.Rules))
	for i, one := range expr.Rules {
		rule, ok := one.(*filter.AtomRule)
		assert.True(t, ok)
		fields[i] = rule.Field
	}
	assert.Equal(t, []string{"name", "updated_at", "id"}, fields)
}

// TestBuildScanFilterOnlyHasPrimaryKey 定位起点用的扫描条件里只能有主键游标一条规则。
//
// 这是定位这一步能成立的地基：name 与 updated_at 在 async_flow 上都没有索引，一旦有人把它们加回
// 这条 WHERE，LIMIT 约束的就从「扫描行数」变成「返回行数」，单次查询的扫描量会失去上界，
// 在两亿行的表上必然超时——那正是这一步要绕开的问题。
func TestBuildScanFilterOnlyHasPrimaryKey(t *testing.T) {
	expr := buildScanFilter("0000001")

	assert.Equal(t, filter.And, expr.Op)
	assert.Len(t, expr.Rules, 1)

	rule, ok := expr.Rules[0].(*filter.AtomRule)
	assert.True(t, ok)
	assert.Equal(t, "id", rule.Field)
	assert.Equal(t, filter.IDGreaterThan.Factory(), rule.Op)
	assert.Equal(t, "0000001", rule.Value)
}

// TestBuildScanFilterSQLWithEmptyCursor 锁定「冷启动第一窗」这条路径：
// 空串游标必须能通过表达式校验并渲染成 SQL。若有人把 RuleIDGreaterThan 换回 RuleGreaterThan，
// 其 ValidateValue 只接受数字或时间格式，会拒绝 base36 字符串主键，本用例会在 Validate 处失败。
func TestBuildScanFilterSQLWithEmptyCursor(t *testing.T) {
	expr := buildScanFilter("")

	exprOpt := filter.NewExprOption(filter.RuleFields(tableasync.AsyncFlowColumns.ColumnTypes()))
	assert.NoError(t, expr.Validate(exprOpt))

	where, values, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	assert.NoError(t, err)
	assert.Contains(t, where, "id > ")
	assert.NotContains(t, where, "name")
	assert.NotContains(t, where, "updated_at")

	// 占位符名带随机后缀，只能按字段名前缀定位
	idValue, ok := lookupPlaceholderValue(values, "id_")
	assert.True(t, ok, "where values %v should contain an id placeholder", values)
	assert.Equal(t, "", idValue)
}

// TestPickStartCursor 命中记录之前的那条 id 作为起点，交给 id > cursor 的查询正好能取到命中记录。
// 返回命中记录本身的 id 会把它跳过去，这条用例挡住那个差一错误。
func TestPickStartCursor(t *testing.T) {
	rows := []flowBrief{newOtherNameRow("id-01"), newHitRow("hit-02"), newOtherNameRow("id-03")}

	cursor, found := pickStartCursor(rows, "")
	assert.True(t, found)
	assert.Equal(t, "id-01", cursor)
}

// TestPickStartCursorStopsAtEveryNameHit 只要 name 命中就停下，不管它是否已经超期。
//
// 这是起点游标能被长期保留的前提：游标之前必须全是 name 不命中、永远不会被清理的记录。
// 若为了少扫几行而跳过「name 命中但尚未超期」的记录，等它超期时就落在游标后面，
// 再也不会被扫到——这条用例挡住那个漏清回归。
func TestPickStartCursorStopsAtEveryNameHit(t *testing.T) {
	// 第二条是 name 命中的记录，不含超期信息，定位阶段一律视为停止点
	rows := []flowBrief{newOtherNameRow("id-01"), newHitRow("hit-02"), newHitRow("hit-03")}

	cursor, found := pickStartCursor(rows, "")
	assert.True(t, found)
	assert.Equal(t, "id-01", cursor)
}

// TestPickStartCursorFirstRowIsTarget 窗口第一条就是命中记录时，游标保持为传入值，
// 不能回退到更早的位置。
func TestPickStartCursorFirstRowIsTarget(t *testing.T) {
	cursor, found := pickStartCursor([]flowBrief{newHitRow("hit-01")}, "last-scanned")
	assert.True(t, found)
	assert.Equal(t, "last-scanned", cursor)
}

// TestPickStartCursorNoTargetInWindow 窗口内全是 name 不命中的记录时游标推进到窗口末尾，
// 下一窗从这里继续，已经空扫过的记录不再重扫。
func TestPickStartCursorNoTargetInWindow(t *testing.T) {
	rows := []flowBrief{newOtherNameRow("id-01"), newOtherNameRow("id-02"), newOtherNameRow("id-03")}

	cursor, found := pickStartCursor(rows, "")
	assert.False(t, found)
	assert.Equal(t, "id-03", cursor)
}

// TestPickStartCursorEmptyWindow 空窗口不推进游标。
func TestPickStartCursorEmptyWindow(t *testing.T) {
	cursor, found := pickStartCursor(nil, "last-scanned")
	assert.False(t, found)
	assert.Equal(t, "last-scanned", cursor)
}

// lookupPlaceholderValue 按字段名前缀取出渲染后的占位符值。
func lookupPlaceholderValue(values map[string]interface{}, prefix string) (interface{}, bool) {
	for key, value := range values {
		if strings.HasPrefix(key, prefix) {
			return value, true
		}
	}

	return nil, false
}

// TestBatchSizeWithinLimits scanWindow 受分页上限约束，flowBatchSize 受 IN 列表上限约束。
//
// scanWindow 直接作为 Page.Limit 下发，超过 core.DefaultMaxPageLimit 会在 DAO 的分页校验处被拒；
// flowBatchSize 决定删除语句里 id IN 列表的长度，超过 filter.DefaultMaxInLimit 会在表达式校验处被拒。
// 任一项越界，清理都彻底不可用。
func TestBatchSizeWithinLimits(t *testing.T) {
	assert.LessOrEqual(t, scanWindow, core.DefaultMaxPageLimit)
	assert.LessOrEqual(t, flowBatchSize, filter.DefaultMaxInLimit)
}

// TestSplitIDsWithinInLimit 1200 个 id 按 IN 上限切分为 500/500/200 三片。
func TestSplitIDsWithinInLimit(t *testing.T) {
	assert.Equal(t, uint(500), filter.DefaultMaxInLimit)

	chunks := slice.Split(newFlowIDs("task", 1200), int(filter.DefaultMaxInLimit))
	assert.Len(t, chunks, 3)
	assert.Len(t, chunks[0], 500)
	assert.Len(t, chunks[1], 500)
	assert.Len(t, chunks[2], 200)
}

// TestCleanupDisabled 开关关闭时不发起任何 DAO 调用。
func TestCleanupDisabled(t *testing.T) {
	setCleanupConfig(false)

	store := &stubStore{pages: [][]string{newFlowIDs("flow", 10)}}
	tenants := &stubTenantLister{tenantIDs: []string{constant.DefaultTenantID}}
	logics := newTestLogicsWithTenants(store, tenants)

	result, err := logics.Cleanup(kit.New())
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrCleanupDisabled)
	// 开关关闭时不应发起 data-service 调用
	assert.Equal(t, 0, tenants.callCount)
	assert.False(t, errors.Is(err, ErrCleanupRunning), "disabled must not be mistaken for running")
	assert.True(t, IsSkipped(err))
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
	assert.Equal(t, 0, store.scanCount)
	assert.Equal(t, 0, store.listCount)
	assert.Equal(t, 0, store.delCount)
}

// TestIsSkipped 只有两个哨兵错误算预期跳过，同为 Aborted 码的真实错误不能被误判。
func TestIsSkipped(t *testing.T) {
	assert.True(t, IsSkipped(ErrCleanupDisabled))
	assert.True(t, IsSkipped(ErrCleanupRunning))
	assert.False(t, IsSkipped(nil))
	assert.False(t, IsSkipped(errf.New(errf.Aborted, "some real failure from lower layer")))
	assert.False(t, IsSkipped(errors.New("plain error")))
}

// TestCleanupNoExpiredFlow 无可清理记录时正常返回、结果全零、不发起删除。
func TestCleanupNoExpiredFlow(t *testing.T) {
	setCleanupConfig(true)

	store := &stubStore{}
	logics := newTestLogics(store)

	result, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, 0, result.DeletedFlowCount)
	assert.Equal(t, 0, result.DeletedTaskCount)
	assert.Equal(t, 1, store.listCount)
	assert.Equal(t, 0, store.delCount)
}

// TestCleanupLocateStartCursor 清理前先定位起点，把那段永远不会被清理的记录空扫过去，
// 再把定位到的游标交给带条件的查询作为第一批的起点。
//
// 这条用例挡住本方案要解决的核心问题：直接从空串开始查，
// 带 name / updated_at 条件的那条查询要扫过整段前缀才能凑够一批命中，在两亿行的表上必然超时。
func TestCleanupLocateStartCursor(t *testing.T) {
	setCleanupConfig(true)

	// 第一窗满窗全是 name 不命中的残留记录，第二窗第三条才是命中记录
	unmatched := newOtherNameWindow("other")
	hitWindow := []flowBrief{newOtherNameRow("b-01"), newOtherNameRow("b-02"), newHitRow("hit-03")}
	store := &stubStore{scanWindows: [][]flowBrief{unmatched, hitWindow}}
	logics := newTestLogics(store)

	_, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)

	// 定位扫描逐窗向后推进，不重扫已经空扫过的记录
	assert.Equal(t, []string{"", unmatched[len(unmatched)-1].ID}, store.scanCursors)
	// 起点是命中记录的前一条，带条件的查询用 id > 它正好能取到命中记录
	assert.Equal(t, []string{"b-02"}, store.listCursors)
	assert.Equal(t, "b-02", logics.startCursors[constant.DefaultTenantID])
}

// TestCleanupLocateHasNoRateLimit 定位阶段不做限速等待。
// 冷启动要空扫过大量残留记录，每窗都睡一次会把定位拖到几小时。
func TestCleanupLocateHasNoRateLimit(t *testing.T) {
	setCleanupConfig(true)

	windows := make([][]flowBrief, 20)
	for i := range windows {
		windows[i] = newOtherNameWindow(fmt.Sprintf("other-%02d", i))
	}
	store := &stubStore{scanWindows: windows}
	logics := newTestLogics(store)

	start := time.Now()
	_, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, len(windows)+1, store.scanCount)
	assert.Equal(t, 0, store.delCount)
	assert.Less(t, time.Since(start), time.Duration(constant.DefaultAsyncFlowCleanupBatchIntervalMs)*
		time.Millisecond)
}

// TestCleanupLocateResumesFromLastCursor 每轮都重新定位，但从上一轮记下的起点续扫，不回到表头。
//
// 每轮重新定位是必要的：上一轮删完之后，原本命中的位置只剩 name 不命中的记录，
// 起点不继续向后推，这段新沉淀的前缀就会重新把带条件的查询拖慢。
// 而续扫是这件事代价可接受的前提，回到表头的话首轮那段历史前缀每轮都要重走一遍。
func TestCleanupLocateResumesFromLastCursor(t *testing.T) {
	setCleanupConfig(true)

	firstWindow := newOtherNameWindow("other")
	store := &stubStore{scanWindows: [][]flowBrief{
		firstWindow,
		{newOtherNameRow("b-01"), newHitRow("hit-02")},
		// 第二轮：上一轮的命中位置已被删空，起点继续往后推到下一条命中记录之前
		{newOtherNameRow("c-01"), newHitRow("hit-02")},
	}}
	logics := newTestLogics(store)

	_, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, "b-01", logics.startCursors[constant.DefaultTenantID])

	_, err = logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, "c-01", logics.startCursors[constant.DefaultTenantID])
	// 第二轮的定位从上一轮的起点续扫，而不是回到空串
	assert.Equal(t, []string{"", firstWindow[len(firstWindow)-1].ID, "b-01"}, store.scanCursors)
}

// TestCleanupCursorNeverSkipsUnexpiredHit 起点游标绝不越过 name 命中的记录，
// 哪怕删除循环已经删到它后面去了。
//
// 删除循环的查询会跳过尚未超期的记录，把它的进度记成起点就等于永久跳过那些记录，
// 等它们超期时再也扫不到。这条用例钉住「起点只由定位决定」这个约束。
func TestCleanupCursorNeverSkipsUnexpiredHit(t *testing.T) {
	setCleanupConfig(true)

	// 定位停在 hit-02 之前，删除循环随后删掉了 id 更大的记录
	store := &stubStore{
		scanWindows: [][]flowBrief{{newOtherNameRow("a-01"), newHitRow("hit-02")}},
		pages:       [][]string{{"hit-99"}},
	}
	logics := newTestLogics(store)

	result, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, 1, result.DeletedFlowCount)
	// 起点仍是定位结果，没有被删除进度顶到 hit-99
	assert.Equal(t, "a-01", logics.startCursors[constant.DefaultTenantID])
}

// TestCleanupPaging 满批时以主键游标继续下一批，不足一批时结束。
func TestCleanupPaging(t *testing.T) {
	setCleanupConfig(true)

	firstPage := newFlowIDs("first", int(flowBatchSize))
	lastPage := newFlowIDs("last", 2)
	store := &stubStore{pages: [][]string{firstPage, lastPage}}
	logics := newTestLogics(store)

	result, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, 2, store.listCount)
	assert.Equal(t, 2, store.delCount)
	assert.Equal(t, len(firstPage)+len(lastPage), result.DeletedFlowCount)
	assert.Equal(t, 2*(len(firstPage)+len(lastPage)), result.DeletedTaskCount)
	assert.True(t, result.DurationMs > 0)
	assert.False(t, result.Interrupted)
	assert.Equal(t, append(append([]string{}, firstPage...), lastPage...), store.delFlowID)
	// 第二批从第一批最后一条继续
	assert.Equal(t, []string{"", firstPage[len(firstPage)-1]}, store.listCursors)
}

// TestCleanupRoundStartsFromLocatedCursor 每轮的第一批都从定位结果开始，
// 而不是从上一轮删到的位置开始。
//
// 删除进度跨过了尚未超期的记录，拿它当下一轮起点会把那些记录永久跳过；
// 从定位结果重新开始虽然要重扫一段，但已删记录本身已经不在表里，不会被重复处理。
func TestCleanupRoundStartsFromLocatedCursor(t *testing.T) {
	setCleanupConfig(true)

	firstPage := newFlowIDs("first", int(flowBatchSize))
	store := &stubStore{
		scanWindows: [][]flowBrief{{newOtherNameRow("a-01"), newHitRow("hit-02")}},
		pages:       [][]string{firstPage, newFlowIDs("last", 2)},
	}
	logics := newTestLogics(store)

	_, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)

	_, err = logics.Cleanup(kit.New())
	assert.NoError(t, err)
	// 第一轮轮内推进到 firstPage 末尾，第二轮回到定位结果 a-01 重新开始
	assert.Equal(t, []string{"a-01", firstPage[len(firstPage)-1], "a-01"}, store.listCursors)
}

// TestCleanupInterrupted 服务优雅退出打断批间等待时，本轮标记为中断，已删除的批次照常计入结果。
func TestCleanupInterrupted(t *testing.T) {
	setCleanupConfig(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 两批都是满批，用例才走得到批间等待
	firstPage := newFlowIDs("first", int(flowBatchSize))
	store := &stubStore{pages: [][]string{firstPage, newFlowIDs("second", int(flowBatchSize))}}
	// 第一批删完后取消上下文，让批间等待立即返回中断
	store.onDelete = cancel
	logics := newTestLogics(store)

	kt := kit.New()
	kt.Ctx = ctx

	result, err := logics.Cleanup(kt)
	assert.NoError(t, err)
	assert.True(t, result.Interrupted, "cleanup cut off by context done should be marked interrupted")
	assert.Equal(t, int(flowBatchSize), result.DeletedFlowCount)
	assert.Equal(t, 2*int(flowBatchSize), result.DeletedTaskCount)
	// 中断发生在第一批之后，不应再发起第二次查询
	assert.Equal(t, 1, store.listCount)
	assert.Equal(t, 1, store.delCount)
	// 剩余记录留待下一轮：已删的那批不在表里，下一轮从定位结果重新开始也不会重复处理
	assert.Equal(t, []string{""}, store.listCursors)
}

// TestCleanupSingleFlight 并发调用时第二个立即被拒绝，且不进入清理主体。
func TestCleanupSingleFlight(t *testing.T) {
	setCleanupConfig(true)

	entered := make(chan struct{})
	release := make(chan struct{})
	store := &stubStore{
		onScan: func() {
			close(entered)
			<-release
		},
	}
	logics := newTestLogics(store)

	done := make(chan error, 1)
	go func() {
		_, err := logics.Cleanup(kit.New())
		done <- err
	}()

	<-entered
	_, err := logics.Cleanup(kit.New())
	assert.ErrorIs(t, err, ErrCleanupRunning)
	assert.False(t, errors.Is(err, ErrCleanupDisabled), "running must not be mistaken for disabled")
	assert.True(t, IsSkipped(err))
	assert.Equal(t, 1, store.scanCount)
	assert.Equal(t, 0, store.listCount)

	close(release)
	select {
	case err = <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("first cleanup did not finish in time")
	}
}

// TestCleanupIterateTenants 清理逐个租户执行，每个租户的查询都携带该租户 id。
// 这条用例挡住「清理只覆盖调用方租户」的回归：调度器给的是 system 租户 kit，
// 而目标数据落在各业务租户下，不逐租户切换就会漏清大部分数据。
func TestCleanupIterateTenants(t *testing.T) {
	setCleanupConfig(true)

	store := &stubStore{pages: [][]string{
		newFlowIDs("flow-t1", 1),
		{},
		newFlowIDs("flow-t3", 2),
	}}
	tenants := &stubTenantLister{tenantIDs: []string{"tenant-1", "tenant-2", "tenant-3"}}
	logics := newTestLogicsWithTenants(store, tenants)

	kt := kit.New()
	kt.TenantID = constant.SystemTenantID

	result, err := logics.Cleanup(kt)
	assert.NoError(t, err)
	assert.Equal(t, 1, tenants.callCount)
	assert.Equal(t, []string{"tenant-1", "tenant-2", "tenant-3"}, store.listTenantIDs)
	assert.Equal(t, 3, result.DeletedFlowCount)
	// 起点游标按租户各记一份，互不干扰
	assert.Len(t, logics.startCursors, 3)
	// 调用方自己的 kit 不能被改动
	assert.Equal(t, constant.SystemTenantID, kt.TenantID)
}

// TestCleanupSingleTenantDeployment 未开启多租户时租户列表只有 default，
// 此时 DAO 不注入 tenant_id 条件，清理覆盖全表。
func TestCleanupSingleTenantDeployment(t *testing.T) {
	setCleanupConfig(true)

	store := &stubStore{pages: [][]string{newFlowIDs("flow", 1)}}
	logics := newTestLogics(store)

	result, err := logics.Cleanup(kit.New())
	assert.NoError(t, err)
	assert.Equal(t, []string{constant.DefaultTenantID}, store.listTenantIDs)
	assert.Equal(t, 1, result.DeletedFlowCount)
}

// TestCleanupListTenantFailed 租户列表查询失败时本轮终止，且不发起任何删除。
func TestCleanupListTenantFailed(t *testing.T) {
	setCleanupConfig(true)

	store := &stubStore{pages: [][]string{newFlowIDs("flow", 1)}}
	logics := newTestLogicsWithTenants(store, &stubTenantLister{err: errors.New("data-service unreachable")})

	result, err := logics.Cleanup(kit.New())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, IsSkipped(err), "a real failure must not be treated as an expected skip")
	assert.Equal(t, 0, store.scanCount)
	assert.Equal(t, 0, store.listCount)
	assert.Equal(t, 0, store.delCount)
}
