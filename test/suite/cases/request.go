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

package cases

import (
	"context"
	"encoding/json"
	"net/http"

	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/uuid"
)

// GenQueryFilterByIds query app filter by id
func GenQueryFilterByIds(ids []uint32) (*pbstruct.Struct, error) {
	ft := filter.Expression{
		Op: filter.Or,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: ids,
			},
		},
	}
	marshal, err := json.Marshal(ft)
	if err != nil {
		return nil, err
	}

	pbStruct := new(pbstruct.Struct)
	if err := pbStruct.UnmarshalJSON(marshal); err != nil {
		return nil, err
	}
	return pbStruct, nil
}

// GenApiKit generate a new kit.Kit with UserKey and AppCodeKey header for testing.
func GenApiKit() *kit.Kit {
	kt := kit.New()
	kt.User = "suite"
	kt.AppCode = "test"
	return kt
}

// GenApiCtxHeader generate request context for api client
func GenApiCtxHeader() (context.Context, http.Header) {
	header := http.Header{}
	header.Set(constant.UserKey, "suite")
	header.Set(constant.RidKey, uuid.UUID())
	header.Set(constant.AppCodeKey, "test")
	return context.Background(), header
}
