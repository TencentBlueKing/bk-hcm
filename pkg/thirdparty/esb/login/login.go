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

package login

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
	"hcm/pkg/tools/uuid"
)

// Client esb login client
type Client interface {
	IsLogin(ctx context.Context, bkToken string) (*IsLoginResp, error)
}

// NewClient initialize a new login client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &login{
		client: client,
		config: config,
	}
}

type login struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// IsLogin oa login check
func (l *login) IsLogin(ctx context.Context, bkToken string) (*IsLoginResp, error) {
	resp := new(IsLoginResp)

	req := &IsLoginReq{
		BkToken: bkToken,
	}

	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())
	types.SetCommonHeader(&h, l.config)
	err := l.client.Post().
		SubResourcef("/bk_login/is_login/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("{\"code\":%d, \"message\":\"%s\"}", resp.Code, resp.Message)
	}

	return resp, nil
}
