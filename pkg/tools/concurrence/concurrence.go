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

// BaseExecWithResult 带返回值的并发执行框架
// Params:
//  1. concurrenceLimit: 最大并发数
//  2. params: 输入参数列表
//  3. execFunc: 执行函数（返回结果和错误）
//
// Returns:
//  1. 结果切片（保序）
//  2. 首个遇到的错误
func BaseExecWithResult[T any, R any](concurrenceLimit int, params []T, execFunc func(param T) (R, error)) ([]R, error) {
	results := make([]R, len(params))
	var lock gosync.Mutex
	var firstErr error

	pipeline := make(chan bool, concurrenceLimit)
	var wg gosync.WaitGroup

	for i, param := range params {
		index := i // 捕获循环变量
		pipeline <- true
		wg.Add(1)

		go func(p T, idx int) {
			defer func() {
				wg.Done()
				<-pipeline
			}()
			if firstErr != nil {
				return // 如果已经有错误，直接返回
			}

			res, err := execFunc(p)
			lock.Lock()
			defer lock.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
			}

			results[idx] = res
		}(param, index)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}
