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

// Package async 异步任务框架
package async

import (
	// 注册测试用例
	_ "hcm/pkg/async/action/test"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/consumer"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/async/producer"
	"hcm/pkg/logs"
)

/*
Async 异步任务框架，提供异步任务下发和异步任务消费功能。
由两部分组成，Producer【生产者】负责异步任务下发操作，Consumer【消费者】负责异步任务消费。
*/
type Async interface {
	// GetProducer 获取生产者。生产者：负责异步任务下发，查询等职责。
	GetProducer() producer.Producer
	// GetConsumer 获取消费者。消费者：负责异步任务执行、重试、强制终止等职责。
	GetConsumer() consumer.Consumer
}

var _ Async = new(async)

// NewAsync new async.
func NewAsync(bd backend.Backend, ld leader.Leader, opt *Option) (Async, error) {
	opt.tryDefaultValue()
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	pdr, err := producer.NewProducer(bd, opt.Register)
	if err != nil {
		logs.Errorf("new producer failed, err: %v", err)
		return nil, err
	}

	csm, err := consumer.NewConsumer(bd, ld, opt.Register, opt.ConsumerOption)
	if err != nil {
		logs.Errorf("new consumer failed, err: %v", err)
		return nil, err
	}

	syn := &async{
		producer: pdr,
		consumer: csm,
	}

	return syn, nil
}

type async struct {
	// producer 生产者：负责异步任务下发，查询等职责。
	producer producer.Producer
	// consumer 消费者：负责异步任务执行、重试、强制终止等职责。
	consumer consumer.Consumer
}

// GetProducer 获取生产者。生产者：负责异步任务下发，查询等职责。
func (a *async) GetProducer() producer.Producer {
	return a.producer
}

// GetConsumer 获取消费者。消费者：负责异步任务执行、重试、强制终止等职责。
func (a *async) GetConsumer() consumer.Consumer {
	return a.consumer
}
