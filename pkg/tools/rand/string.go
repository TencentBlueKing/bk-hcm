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
	cryptoRand "crypto/rand"
	"math/big"
	"math/rand"
	"time"

	"hcm/pkg/logs"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// String randomly generate a string of specified length.
func String(n int) string {
	b := make([]rune, n)
	for i := range b {
		num, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			logs.Errorf("rand.Int failed: %v", err)
			num = big.NewInt(0)
		}
		b[i] = letterRunes[num.Int64()]
	}

	return string(b)
}

// RandomRange return a random value which is between the value of between[0] and between[1].
// so, do assure that between[0] < between[1].
func RandomRange(between [2]int) int {
	randX := rand.New(rand.NewSource(time.Now().UnixNano()))
	return randX.Intn(between[1]-between[0]) + between[0]
}

// Prefix random string with prefix, same as prefix+rand.String(n)
func Prefix(prefix string, n int) string {
	randX := rand.New(rand.NewSource(time.Now().UnixNano()))
	prefixLen := len(prefix)
	b := make([]rune, n+prefixLen)
	for i := range prefix {
		b[i] = rune(prefix[i])
	}
	for i := 0; i < n; i++ {
		b[i+prefixLen] = letterRunes[randX.Intn(len(letterRunes))]
	}

	return string(b)
}
