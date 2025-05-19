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

package rand

import (
	"bytes"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var globalRandX = rand.New(rand.NewSource(time.Now().UnixNano()))

var randMu sync.Mutex

var lastStringN []byte

// String randomly generate a string of specified length.
func String(n int) string {
	randMu.Lock()
	defer randMu.Unlock()
	b := make([]byte, n)
	for {
		for i := range b {
			b[i] = letterBytes[globalRandX.Intn(len(letterBytes))]
		}
		if bytes.Compare(b, lastStringN) != 0 {
			lastStringN = b
			return *(*string)(unsafe.Pointer(&b))
		}
	}
}

// RandomRange return a random value which is between the value of between[0] and between[1].
// so, do assure that between[0] < between[1].
func RandomRange(between [2]int) int {
	randMu.Lock()
	defer randMu.Unlock()
	return globalRandX.Intn(between[1]-between[0]) + between[0]
}

// Prefix random string with prefix, same as prefix+rand.String(n)
func Prefix(prefix string, n int) string {
	randMu.Lock()
	defer randMu.Unlock()
	prefixLen := len(prefix)
	b := make([]byte, n+prefixLen)
	for i := range prefix {
		b[i] = prefix[i]
	}
	for i := 0; i < n; i++ {
		b[i+prefixLen] = letterBytes[globalRandX.Intn(len(letterBytes))]
	}

	return string(b)
}
