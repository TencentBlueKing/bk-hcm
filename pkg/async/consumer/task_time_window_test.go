package consumer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTimeWindow(t *testing.T) {
	tests := []struct {
		name     string
		capacity uint
		duration uint
	}{
		{"正常创建", 10, 5},
		{"最小容量", 1, 1},
		{"大容量", 1000, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := NewTimeWindow(tt.capacity, tt.duration)
			if tw == nil {
				t.Fatal("NewTimeWindow返回nil")
			}
			if tw.capacity != tt.capacity {
				t.Errorf("容量设置错误，期望 %d，实际 %d", tt.capacity, tw.capacity)
			}
			if tw.duration != time.Duration(tt.duration)*time.Minute {
				t.Errorf("时间窗口设置错误，期望 %v，实际 %v",
					time.Duration(tt.duration)*time.Minute, tw.duration)
			}
			if tw.size != 0 {
				t.Errorf("初始大小应为0，实际 %d", tw.size)
			}
			if tw.head != 0 || tw.tail != 0 {
				t.Errorf("初始头尾指针应为0，实际 head=%d, tail=%d", tw.head, tw.tail)
			}
		})
	}
}

func TestTimeWindow_Push(t *testing.T) {
	t.Run("正常推入数据", func(t *testing.T) {
		tw := NewTimeWindow(3, 5)

		// 推入第一个数据
		tw.Push(1.5)
		if tw.size != 1 {
			t.Errorf("推入一个数据后，size应为1，实际 %d", tw.size)
		}
		if tw.tail != 1 {
			t.Errorf("推入一个数据后，tail应为1，实际 %d", tw.tail)
		}

		// 推入更多数据
		tw.Push(2.0)
		tw.Push(2.5)
		if tw.size != 3 {
			t.Errorf("推入三个数据后，size应为3，实际 %d", tw.size)
		}
	})

	t.Run("超容量覆盖测试", func(t *testing.T) {
		tw := NewTimeWindow(2, 5)

		// 填满队列
		tw.Push(1.0)
		tw.Push(2.0)

		// 超容量推入，应该覆盖最旧的数据
		tw.Push(3.0)

		if tw.size != 2 {
			t.Errorf("超容量后，size应保持为2，实际 %d", tw.size)
		}
		if tw.head != 1 {
			t.Errorf("超容量后，head应为1，实际 %d", tw.head)
		}
		if tw.tail != 1 {
			t.Errorf("超容量后，tail应为1，实际 %d", tw.tail)
		}
	})
}

func TestTimeWindow_GetAvg_EmptyQueue(t *testing.T) {
	tw := NewTimeWindow(5, 5)

	avgTime, neverExec := tw.GetAvg()

	if !neverExec {
		t.Error("空队列应该返回neverExec=true")
	}
	if avgTime != 0 {
		t.Errorf("空队列平均时间应为0，实际 %f", avgTime)
	}
}

func TestTimeWindow_GetAvg_WithinTimeWindow(t *testing.T) {
	tw := NewTimeWindow(5, 1) // 1分钟时间窗口

	// 推入一些数据
	tw.Push(1.0)
	tw.Push(2.0)
	tw.Push(3.0)

	avgTime, neverExec := tw.GetAvg()

	if neverExec {
		t.Error("有数据的队列不应该返回neverExec=true")
	}

	expectedAvg := (1.0 + 2.0 + 3.0) / 3.0
	if avgTime != expectedAvg {
		t.Errorf("平均时间计算错误，期望 %f，实际 %f", expectedAvg, avgTime)
	}
}

func TestTimeWindow_GetAvg_ExpiredData(t *testing.T) {
	tw := NewTimeWindow(5, 1) // 1分钟时间窗口

	// 手动设置过期数据
	tw.Lock()
	oldTime := time.Now().Add(-2 * time.Minute) // 2分钟前的数据
	tw.queue[0] = taskTypeExecTime{execTime: 1.0, entryTime: oldTime}
	tw.queue[1] = taskTypeExecTime{execTime: 2.0, entryTime: oldTime}
	tw.size = 2
	tw.tail = 2
	tw.Unlock()

	avgTime, neverExec := tw.GetAvg()

	if neverExec {
		t.Error("有数据的队列不应该返回neverExec=true")
	}

	// 应该返回过期数据的平均值作为参考
	expectedAvg := (1.0 + 2.0) / 2.0
	if avgTime != expectedAvg {
		t.Errorf("过期数据平均时间计算错误，期望 %f，实际 %f", expectedAvg, avgTime)
	}
}

func TestTimeWindow_GetAvg_MixedData(t *testing.T) {
	tw := NewTimeWindow(5, 1) // 1分钟时间窗口

	// 手动设置混合数据：部分过期，部分有效
	tw.Lock()
	oldTime := time.Now().Add(-2 * time.Minute) // 过期数据
	newTime := time.Now()                       // 有效数据

	tw.queue[0] = taskTypeExecTime{execTime: 1.0, entryTime: oldTime} // 过期
	tw.queue[1] = taskTypeExecTime{execTime: 2.0, entryTime: oldTime} // 过期
	tw.queue[2] = taskTypeExecTime{execTime: 3.0, entryTime: newTime} // 有效
	tw.queue[3] = taskTypeExecTime{execTime: 4.0, entryTime: newTime} // 有效
	tw.size = 4
	tw.tail = 4
	tw.Unlock()

	avgTime, neverExec := tw.GetAvg()

	if neverExec {
		t.Error("有数据的队列不应该返回neverExec=true")
	}

	// 应该返回有效数据的平均值，并清理过期数据
	expectedAvg := (3.0 + 4.0) / 2.0
	if avgTime != expectedAvg {
		t.Errorf("混合数据平均时间计算错误，期望 %f，实际 %f", expectedAvg, avgTime)
	}

	// 检查过期数据是否被清理
	tw.Lock()
	if tw.size != 2 {
		t.Errorf("清理后size应为2，实际 %d", tw.size)
	}
	if tw.head != 2 {
		t.Errorf("清理后head应为2，实际 %d", tw.head)
	}
	tw.Unlock()
}

func TestTimeWindow_CircularBuffer(t *testing.T) {
	tw := NewTimeWindow(3, 5)

	// 测试环形缓冲区的正确性
	for i := 0; i < 10; i++ {
		tw.Push(float64(i))
	}

	// 应该只保留最后3个数据
	if tw.size != 3 {
		t.Errorf("环形缓冲区size应为3，实际 %d", tw.size)
	}

	avgTime, _ := tw.GetAvg()
	expectedAvg := (7.0 + 8.0 + 9.0) / 3.0 // 最后三个数据的平均值
	if avgTime != expectedAvg {
		t.Errorf("环形缓冲区平均时间计算错误，期望 %f，实际 %f", expectedAvg, avgTime)
	}
}

func TestTimeWindow_ConcurrentSafety(t *testing.T) {
	tw := NewTimeWindow(100, 5)

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// 并发推入数据
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				value := float64(id*numOperations + j)
				tw.Push(value)
			}
		}(i)
	}

	// 并发读取平均值并验证数据结构完整性
	var structuralErrors []string
	var errorsMutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// 直接调用GetAvg，不预先计算期望值
				// 因为GetAvg会修改队列状态（清理过期数据）
				avg, neverExec := tw.GetAvg()

				// 验证返回值的合理性
				if !neverExec && avg < 0 {
					errorsMutex.Lock()
					structuralErrors = append(structuralErrors,
						fmt.Sprintf("Reader %d: 负平均值 %f", readerID, avg))
					errorsMutex.Unlock()
				}

				// 验证队列结构完整性
				tw.Lock()
				currentSize := tw.size
				currentHead := tw.head
				currentTail := tw.tail
				currentCapacity := tw.capacity

				if currentSize > currentCapacity {
					errorsMutex.Lock()
					structuralErrors = append(structuralErrors,
						fmt.Sprintf("Reader %d: size超出容量 %d > %d", readerID, currentSize, currentCapacity))
					errorsMutex.Unlock()
				}

				if currentHead >= currentCapacity || currentTail >= currentCapacity {
					errorsMutex.Lock()
					structuralErrors = append(structuralErrors,
						fmt.Sprintf("Reader %d: 指针越界 head=%d, tail=%d, capacity=%d",
							readerID, currentHead, currentTail, currentCapacity))
					errorsMutex.Unlock()
				}
				tw.Unlock()

				time.Sleep(time.Microsecond) // 增加竞争
			}
		}(i)
	}

	wg.Wait()

	// 检查结构性错误
	if len(structuralErrors) > 0 {
		for _, err := range structuralErrors {
			t.Error(err)
		}
	}

	// 最终状态验证
	tw.Lock()
	finalSize := tw.size
	finalHead := tw.head
	finalTail := tw.tail
	tw.Unlock()

	if finalSize > tw.capacity {
		t.Errorf("最终size超出容量: size=%d, capacity=%d", finalSize, tw.capacity)
	}

	if finalHead >= tw.capacity || finalTail >= tw.capacity {
		t.Errorf("最终指针越界: head=%d, tail=%d, capacity=%d",
			finalHead, finalTail, tw.capacity)
	}

	t.Logf("并发测试完成: 最终队列大小%d, 结构性错误%d个",
		finalSize, len(structuralErrors))
}

func TestTimeWindow_EdgeCases(t *testing.T) {
	t.Run("容量为1的时间窗口", func(t *testing.T) {
		tw := NewTimeWindow(1, 5)

		tw.Push(1.0)
		tw.Push(2.0) // 应该覆盖第一个

		avgTime, neverExec := tw.GetAvg()
		if neverExec {
			t.Error("不应该返回neverExec=true")
		}
		if avgTime != 2.0 {
			t.Errorf("容量为1时平均时间应为2.0，实际 %f", avgTime)
		}
	})

	t.Run("零执行时间", func(t *testing.T) {
		tw := NewTimeWindow(5, 5)

		tw.Push(0.0)
		tw.Push(0.0)

		avgTime, neverExec := tw.GetAvg()
		if neverExec {
			t.Error("不应该返回neverExec=true")
		}
		if avgTime != 0.0 {
			t.Errorf("零执行时间的平均值应为0.0，实际 %f", avgTime)
		}
	})

	t.Run("负执行时间", func(t *testing.T) {
		tw := NewTimeWindow(5, 5)

		tw.Push(-1.0)
		tw.Push(1.0)

		avgTime, neverExec := tw.GetAvg()
		if neverExec {
			t.Error("不应该返回neverExec=true")
		}
		if avgTime != 0.0 {
			t.Errorf("负数和正数的平均值应为0.0，实际 %f", avgTime)
		}
	})
}

func TestTimeWindow_TimeWindowBoundary(t *testing.T) {
	tw := NewTimeWindow(5, 1) // 1分钟时间窗口

	// 推入一个数据
	tw.Push(1.0)

	// 立即获取平均值，应该在时间窗口内
	avgTime, neverExec := tw.GetAvg()
	if neverExec || avgTime != 1.0 {
		t.Error("刚推入的数据应该在时间窗口内")
	}

	// 等待超过时间窗口
	time.Sleep(61 * time.Second) // 等待超过1分钟

	// 再次获取平均值，数据应该过期
	avgTime, neverExec = tw.GetAvg()
	if neverExec {
		t.Error("过期数据仍应返回平均值作为参考")
	}
	if avgTime != 1.0 {
		t.Errorf("过期数据的平均值应为1.0，实际 %f", avgTime)
	}
}

// 基准测试
func BenchmarkTimeWindow_Push(b *testing.B) {
	tw := NewTimeWindow(1000, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.Push(float64(i))
	}
}

func BenchmarkTimeWindow_GetAvg(b *testing.B) {
	tw := NewTimeWindow(1000, 5)

	// 预填充数据
	for i := 0; i < 1000; i++ {
		tw.Push(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.GetAvg()
	}
}

// 测试并发操作的数据一致性
func TestTimeWindow_ConcurrentDataConsistency(t *testing.T) {
	tw := NewTimeWindow(50, 1) // 较小的容量便于测试

	var wg sync.WaitGroup
	numWriters := 5
	numReaders := 3
	numOperations := 100

	// 用于跟踪写入的数据
	type writeRecord struct {
		value     float64
		timestamp time.Time
	}
	var writeRecords []writeRecord
	var recordsMutex sync.Mutex

	// 并发写入
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				value := float64(writerID*1000 + j)
				timestamp := time.Now()
				tw.Push(value)

				recordsMutex.Lock()
				writeRecords = append(writeRecords, writeRecord{
					value:     value,
					timestamp: timestamp,
				})
				recordsMutex.Unlock()

				// 随机延迟，增加竞争
				time.Sleep(time.Microsecond * time.Duration(j%10))
			}
		}(i)
	}

	// 并发读取并验证
	var readErrors []string
	var errorsMutex sync.Mutex

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// 直接调用GetAvg，不预先计算期望值
				// 因为GetAvg会修改队列状态（清理过期数据）
				avg, neverExec := tw.GetAvg()

				// 验证返回值的合理性
				if !neverExec && avg < 0 {
					errorsMutex.Lock()
					readErrors = append(readErrors,
						fmt.Sprintf("Reader %d: 负平均值 %f", readerID, avg))
					errorsMutex.Unlock()
				}

				// 验证队列结构完整性
				tw.Lock()
				currentSize := tw.size
				currentHead := tw.head
				currentTail := tw.tail
				currentCapacity := tw.capacity

				if currentSize > currentCapacity {
					errorsMutex.Lock()
					readErrors = append(readErrors,
						fmt.Sprintf("Reader %d: size超出容量 %d > %d", readerID, currentSize, currentCapacity))
					errorsMutex.Unlock()
				}

				if currentHead >= currentCapacity || currentTail >= currentCapacity {
					errorsMutex.Lock()
					readErrors = append(readErrors,
						fmt.Sprintf("Reader %d: 指针越界 head=%d, tail=%d, capacity=%d",
							readerID, currentHead, currentTail, currentCapacity))
					errorsMutex.Unlock()
				}
				tw.Unlock()

				time.Sleep(time.Microsecond * time.Duration(j%5))
			}
		}(i)
	}

	wg.Wait()

	// 检查是否有读取错误
	if len(readErrors) > 0 {
		for _, err := range readErrors {
			t.Error(err)
		}
	}

	// 最终状态验证
	tw.Lock()
	finalSize := tw.size
	finalHead := tw.head
	finalTail := tw.tail
	tw.Unlock()

	// 验证最终状态的合理性
	if finalSize > tw.capacity {
		t.Errorf("最终size超出容量: size=%d, capacity=%d", finalSize, tw.capacity)
	}

	if finalHead >= tw.capacity || finalTail >= tw.capacity {
		t.Errorf("指针越界: head=%d, tail=%d, capacity=%d",
			finalHead, finalTail, tw.capacity)
	}

	t.Logf("数据一致性测试完成: 写入%d条记录, 最终队列大小%d, 读取错误%d个",
		len(writeRecords), finalSize, len(readErrors))
}

// 测试极端并发场景下的数据竞争
func TestTimeWindow_RaceConditionDetection(t *testing.T) {
	tw := NewTimeWindow(10, 1) // 很小的容量，增加竞争

	var wg sync.WaitGroup
	numGoroutines := 20
	numOperations := 50

	// 用于检测数据竞争的计数器
	var pushCount, getCount int64
	var inconsistencyCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// 随机选择操作类型
				if j%3 == 0 {
					// Push操作
					tw.Push(float64(id*numOperations + j))
					atomic.AddInt64(&pushCount, 1)
				} else {
					// GetAvg操作并验证结果
					// 直接调用GetAvg，不预先计算期望值
					avg, neverExec := tw.GetAvg()
					atomic.AddInt64(&getCount, 1)

					// 验证返回值的合理性
					if !neverExec && avg < 0 {
						atomic.AddInt64(&inconsistencyCount, 1)
					}
				}

				// 偶尔检查内部状态的一致性
				if j%25 == 0 {
					tw.Lock()
					size := tw.size
					head := tw.head
					tail := tw.tail
					capacity := tw.capacity
					tw.Unlock()

					// 检查状态一致性
					if size > capacity {
						atomic.AddInt64(&inconsistencyCount, 1)
						t.Errorf("Goroutine %d: size > capacity (%d > %d)", id, size, capacity)
					}
					if head >= capacity || tail >= capacity {
						atomic.AddInt64(&inconsistencyCount, 1)
						t.Errorf("Goroutine %d: 指针越界 head=%d, tail=%d, capacity=%d",
							id, head, tail, capacity)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	// 最终验证
	finalInconsistencies := atomic.LoadInt64(&inconsistencyCount)
	if finalInconsistencies > 0 {
		t.Errorf("检测到 %d 个数据不一致问题", finalInconsistencies)
	}

	t.Logf("竞争条件测试完成: Push操作%d次, GetAvg操作%d次, 不一致问题%d个",
		pushCount, getCount, finalInconsistencies)
}

func BenchmarkTimeWindow_ConcurrentOperations(b *testing.B) {
	tw := NewTimeWindow(1000, 5)

	// 预填充一些数据以获得更真实的性能数据
	for i := 0; i < 100; i++ {
		tw.Push(float64(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				tw.Push(float64(i))
			} else {
				avg, neverExec := tw.GetAvg()
				// 简单验证返回值的合理性
				if !neverExec && avg < 0 {
					b.Errorf("基准测试中发现负平均值: %f", avg)
				}
			}
			i++
		}
	})
}
