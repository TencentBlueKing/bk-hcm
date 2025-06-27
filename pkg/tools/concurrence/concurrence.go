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

// Package concurrence ...
package concurrence

import (
	gosync "sync"
)

// BaseExec 基础并发执行基础架子。
// Params:
//  1. concurrenceLimit: 并发执行最大写协程数量
//  2. params: 并发执行变量参数
//  3. execFunc: 并发执行函数
func BaseExec[T any](concurrenceLimit int, params []T, execFunc func(param T) error) error {

	pipeline := make(chan bool, concurrenceLimit)
	var firstErr error
	var wg gosync.WaitGroup
	for _, param := range params {
		pipeline <- true
		wg.Add(1)

		go func(param T) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			err := execFunc(param)
			if firstErr == nil && err != nil {
				firstErr = err
				return
			}
		}(param)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	return nil
}
