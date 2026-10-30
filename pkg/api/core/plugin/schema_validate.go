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

package plugin

import (
	"crypto"
	_ "crypto/sha256" // 注册 SHA-256 到 crypto.Hash
	_ "crypto/sha512" // 注册 SHA-384 / SHA-512 到 crypto.Hash
	"encoding/hex"
	"strings"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
)

// supportedSchemaHashAlgos 支持的 schema_hash 算法白名单，映射到标准库 crypto.Hash。
var supportedSchemaHashAlgos = map[enumor.SchemaHashAlgo]crypto.Hash{
	enumor.SchemaHashAlgoSHA256: crypto.SHA256,
	enumor.SchemaHashAlgoSHA384: crypto.SHA384,
	enumor.SchemaHashAlgoSHA512: crypto.SHA512,
}

// ParseSchemaHash 解析 "<algo>:<hex>" 格式的 schema_hash，校验算法是否在支持列表内且摘要为对应长度的小写 hex。
func ParseSchemaHash(schemaHash string) (enumor.SchemaHashAlgo, string, crypto.Hash, error) {
	algoStr, hexDigest, ok := strings.Cut(schemaHash, ":")
	if !ok || len(algoStr) == 0 || len(hexDigest) == 0 {
		return "", "", 0, errf.Newf(errf.InvalidParameter,
			"invalid schema_hash format: %q, expected <algo>:<lowercase hex>", schemaHash)
	}

	algo := enumor.SchemaHashAlgo(algoStr)
	hashFunc, supported := supportedSchemaHashAlgos[algo]
	if !supported || !hashFunc.Available() {
		return "", "", 0, errf.Newf(errf.InvalidParameter,
			"unsupported schema_hash algorithm: %q, supported algorithms: [sha256, sha384, sha512]", algo)
	}

	expectedHexLen := hashFunc.Size() * 2
	if len(hexDigest) != expectedHexLen || !isLowerHex(hexDigest) {
		return "", "", 0, errf.Newf(errf.InvalidParameter,
			"invalid %s schema_hash digest: %q, expected %d lowercase hex chars",
			algo, hexDigest, expectedHexLen)
	}

	return algo, hexDigest, hashFunc, nil
}

func isLowerHex(s string) bool {
	if _, err := hex.DecodeString(s); err != nil {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'F' {
			return false
		}
	}
	return true
}

// ValidateSchemaHashFormat 按前缀算法校验 schemaHash 格式是否合法且算法受支持。
func ValidateSchemaHashFormat(schemaHash string) error {
	_, _, _, err := ParseSchemaHash(schemaHash)
	return err
}
