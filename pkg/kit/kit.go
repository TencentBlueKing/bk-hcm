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

package kit

import (
	"context"
	"errors"
	"net/http"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/tools/uuid"
)

// New initial a kit with rid and context.
func New() *Kit {
	rid := uuid.UUID()
	return &Kit{
		Rid: rid,
		Ctx: context.WithValue(context.TODO(), constant.RidKey, rid),
	}
}

// Kit defines the basic metadata info within a task.
type Kit struct {
	// Ctx is request context.
	Ctx context.Context

	// User's name.
	User string

	// Rid is request id.
	Rid string

	// AppCode is app code.
	AppCode string
}

// ContextWithRid ...
func (c *Kit) ContextWithRid() context.Context {
	return context.WithValue(c.Ctx, constant.RidKey, c.Rid)
}

// CtxWithTimeoutMS create a new context with basic info and timout configuration.
func (c *Kit) CtxWithTimeoutMS(timeoutMS int) context.CancelFunc {
	ctx := context.WithValue(context.TODO(), constant.RidKey, c.Rid)
	var cancel context.CancelFunc
	c.Ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	return cancel
}

// Validate context kit.
func (c *Kit) Validate() error {
	if c.Ctx == nil {
		return errors.New("context is required")
	}

	if len(c.User) == 0 {
		return errors.New("user is required")
	}

	ridLen := len(c.Rid)
	if ridLen == 0 {
		return errors.New("rid is required")
	}

	if ridLen < 16 || ridLen > 48 {
		return errors.New("rid length not right, length should in 16~48")
	}

	if len(c.AppCode) == 0 {
		return errors.New("app code is required")
	}

	return nil
}

// FromHeader http request header to context kit and validate.
func FromHeader(ctx context.Context, header http.Header) (*Kit, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	kt := &Kit{
		Ctx:  ctx,
		User: header.Get(constant.UserKey),
		Rid:  header.Get(constant.RidKey),
	}

	if err := kt.Validate(); err != nil {
		return nil, err
	}

	return kt, nil
}
