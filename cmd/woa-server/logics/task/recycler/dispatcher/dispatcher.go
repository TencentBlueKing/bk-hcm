/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package dispatcher implements recycle order dispatcher
// which dispatches recycle order to different processors.
package dispatcher

import (
	"context"
	"errors"
	"strconv"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	rslogics "hcm/cmd/woa-server/logics/rolling-server"
	srlogics "hcm/cmd/woa-server/logics/short-rental"
	"hcm/cmd/woa-server/logics/task/recycler/detector"
	"hcm/cmd/woa-server/logics/task/recycler/notifier"
	"hcm/cmd/woa-server/logics/task/recycler/returner"
	"hcm/cmd/woa-server/logics/task/recycler/transit"
	"hcm/pkg"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/utils/wait"

	"k8s.io/client-go/util/workqueue"
)

// Dispatcher dispatch and deal recycle order
type Dispatcher struct {
	detector *detector.Detector
	returner *returner.Returner
	transit  *transit.Transit
	queue    workqueue.RateLimitingInterface
	ctx      context.Context
	rsLogic  rslogics.Logics
	srLogic  srlogics.Logics
	notifier *notifier.Notifier
}

// New create a dispatcher
func New(ctx context.Context, moduleDetector *detector.Detector, moduleReturner *returner.Returner,
	moduleTransit *transit.Transit, rsLogic rslogics.Logics, srLogic srlogics.Logics, notifier *notifier.Notifier) (
	*Dispatcher, error) {

	dispatcher := &Dispatcher{
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "recycle_dispatch"),
		ctx:      ctx,
		detector: moduleDetector,
		returner: moduleReturner,
		transit:  moduleTransit,
		rsLogic:  rsLogic,
		srLogic:  srLogic,
		notifier: notifier,
	}

	// TODO: get worker num from config
	go dispatcher.Run(20)

	return dispatcher, nil
}

// GetTransit get dispatcher member transit
// Deprecated: 不应该对外暴露
func (d *Dispatcher) GetTransit() *transit.Transit {
	return d.transit
}

// GetReturn get dispatcher member returner
// Deprecated: 不应该对外暴露
func (d *Dispatcher) GetReturn() *returner.Returner {
	return d.returner
}

// GetDetector get dispatcher member detector
// Deprecated: 不应该对外暴露
func (d *Dispatcher) GetDetector() *detector.Detector {
	return d.detector
}

// Run starts dispatcher
func (d *Dispatcher) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(d.runWorker, time.Second, d.ctx)
	}

	select {
	case <-d.ctx.Done():
		logs.Infof("dispatcher exits")
	}
}

// Pop gets head of recycle order queue
func (d *Dispatcher) Pop() (string, error) {
	obj, shutdown := d.queue.Get()
	if shutdown {
		return "", nil
	}

	defer d.queue.Done(obj)

	id, ok := obj.(string)
	if !ok {
		d.queue.Forget(obj)
		logs.Warnf("Expected string in queue but got %#v", obj)
		return "", errors.New("got non-string from queue")
	}

	d.queue.Forget(obj)

	return id, nil
}

// Add add recycle order to recycle order queue
// 兼容旧单据恢复、流转逻辑, 暂时保留
func (d *Dispatcher) Add(orderId string) {
	d.queue.Add(orderId)
}

// runWorker deals with recycle order
func (d *Dispatcher) runWorker() error {
	orderId, err := d.Pop()
	if err != nil {
		logs.Errorf("failed to deal recycle order, for get recycle order from informer err: %v", err)
		return err
	}
	logs.Infof("recycler dispatcher get one recycle order %s", orderId)
	if err := d.dispatchHandler(orderId); err != nil {
		logs.Errorf("failed to dispatch recycle order, err: %v, order id: %s", err, orderId)
		return err
	}

	logs.Infof("Successfully dispatch recycle order %s", orderId)

	return nil
}

// dispatchHandler recycle order dispatch handler
func (d *Dispatcher) dispatchHandler(orderId string) (err error) {

	var order *table.RecycleOrder
	// get recycle order by key
	order, err = d.getRecycleOrder(orderId)
	if err != nil {
		logs.Errorf("failed to get recycle order %s, err: %v", orderId, err)
		return err
	}

	// 记录任务执行指标
	startAt := time.Now()
	defer func() {
		d.recordMetrics(startAt, order, err)
	}()

	task := NewTask(order.Status, d.srLogic)
	taskCtx := &CommonContext{
		Order:      order,
		Dispatcher: d,
	}

	if err = task.State.Execute(taskCtx); err != nil {
		logs.Errorf("[%s] failed to execute task, err: %v, order id: %s, state: %s",
			constant.CvmRecycleFailed, err, order.SuborderID, task.State.Name())
		return err
	}

	logs.Infof("finished dispatch order %s, state: %s", orderId, task.State.Name())

	return nil
}

// getRecycleOrder gets recycle order by recycle order id
func (d *Dispatcher) getRecycleOrder(key string) (*table.RecycleOrder, error) {
	filter := &mapstr.MapStr{
		"suborder_id": key,
	}
	order, err := dao.Set().RecycleOrder().GetRecycleOrder(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get recycle order by order id: %d, err: %v", key, err)
		return nil, err
	}

	return order, nil
}

func (d *Dispatcher) getRecycleHosts(orderId string) ([]*table.RecycleHost, error) {
	filter := map[string]interface{}{
		"suborder_id": orderId,
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKMaxInstanceLimit,
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle hosts, err: %v", err)
		return nil, err
	}

	return insts, nil
}
func (d *Dispatcher) recordMetrics(startAt time.Time, order *table.RecycleOrder, err error) {
	labels := map[string]string{
		"status":    string(order.Status),
		"bk_biz_id": strconv.FormatInt(order.BizID, 10),
	}

	// 记录单次状态执行耗时
	currentStateCost := time.Since(startAt)
	dispatcherMetrics.OrderStateCostSec.With(labels).Observe(currentStateCost.Seconds())

	// 记录从提交到当前状态耗时
	if order.CommittedAt.After(order.CreateAt) {
		costSinceCommit := time.Since(order.CommittedAt)
		dispatcherMetrics.RecycleStateCostSinceCommitSec.With(labels).Observe(costSinceCommit.Seconds())
	}

	// 记录错误
	if err != nil {
		dispatcherMetrics.OrderStateErrCounter.With(labels).Inc()
	}
}

// Cancel 取消子单调度
func (d *Dispatcher) Cancel(kt *kit.Kit, suborderID string) error {
	return d.detector.Cancel(kt, suborderID)
}
