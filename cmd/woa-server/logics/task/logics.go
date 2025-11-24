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

// Package logics ...
package logics

import (
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/operation"
	"hcm/cmd/woa-server/logics/task/recycler"
	"hcm/cmd/woa-server/logics/task/scheduler"
	"hcm/cmd/woa-server/logics/task/statistics"
)

// Logics provides management interface for operations of model and instance and related resources like association
type Logics interface {
	Scheduler() scheduler.Interface
	Recycler() recycler.Interface
	Operation() operation.Interface
	Statistics() statistics.Interface
}

type logics struct {
	scheduler  scheduler.Interface
	recycler   recycler.Interface
	informer   informer.Interface
	operation  operation.Interface
	statistics statistics.Interface
}

// New create a logics manager
func New(schedulerIf scheduler.Interface, recyclerIf recycler.Interface,
	informerIf informer.Interface, operationIf operation.Interface,
	statisticsIf statistics.Interface) Logics {

	return &logics{
		scheduler:  schedulerIf,
		recycler:   recyclerIf,
		informer:   informerIf,
		operation:  operationIf,
		statistics: statisticsIf,
	}
}

// Scheduler scheduler interface
func (l *logics) Scheduler() scheduler.Interface {
	return l.scheduler
}

// Recycler recycler interface
func (l *logics) Recycler() recycler.Interface {
	return l.recycler
}

// Operation operation interface
func (l *logics) Operation() operation.Interface {
	return l.operation
}

// Statistics statistics interface
func (l *logics) Statistics() statistics.Interface {
	return l.statistics
}
