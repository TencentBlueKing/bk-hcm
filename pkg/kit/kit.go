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

// Package kit ...
package kit

import (
	"context"
	"errors"
	"net/http"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/rand"
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

	// TenantID is tenant id.
	TenantID string

	// RequestSource 请求来源，字段为空是默认为 ApiCall 类型。
	// Note: hcm请求分为 ApiCall 和 BackgroundSync 请求。所以，RequestSource是内部使用字段，
	// 因为来自前端和第三方系统调用的请求均为 ApiCall，所以没必要将该字段暴漏出去，仅同步请求需要设
	// 置该字段为 BackgroundSync。
	RequestSource enumor.RequestSourceType
}

// SetTenant set tenant id
func (kt *Kit) SetTenant(tenantID string) {
	kt.TenantID = tenantID
}

// NewSubKit 在当前kit后缀加上6位随机字符串
func (kt *Kit) NewSubKit() *Kit {
	suffix := rand.String(6)
	return kt.NewSubKitWithSuffix(suffix)
}

// NewSubKitWithSuffix 生成子kit, 指定后缀
func (kt *Kit) NewSubKitWithSuffix(suffix string) *Kit {

	newSubKit := converter.ValToPtr(*kt)
	subRid := kt.Rid + "/" + suffix
	newSubKit.Rid = subRid
	newSubKit.Ctx = context.WithValue(kt.Ctx, constant.RidKey, subRid)
	return newSubKit
}

// NewSubKitWithTenant 生成子kit, 变更租户id
func (kt *Kit) NewSubKitWithTenant(tenantID string) *Kit {
	newSubKit := converter.ValToPtr(*kt)
	newSubKit.TenantID = tenantID
	return newSubKit
}

// NewSubKitWithRid 生成子kit, 指定rid, 继承 context
func (kt *Kit) NewSubKitWithRid(subRid string) *Kit {
	newSubKit := converter.ValToPtr(*kt)
	newSubKit.Rid = subRid
	newSubKit.Ctx = context.WithValue(kt.Ctx, constant.RidKey, subRid)
	return newSubKit
}

// NewSubKitWithCtx 生成子kit, 变更context
func (kt *Kit) NewSubKitWithCtx(newCtx context.Context) *Kit {
	newSubKit := converter.ValToPtr(*kt)
	newSubKit.Ctx = newCtx
	return newSubKit
}

// WithAsyncSource 生成子kit 设置对应的请求来源为 AsynchronousTasks
func (kt *Kit) WithAsyncSource() *Kit {
	newKit := converter.ValToPtr(*kt)
	newKit.RequestSource = enumor.AsynchronousTasks
	newKit.Ctx = context.WithValue(kt.Ctx, constant.RequestSourceKey, newKit.RequestSource)
	return newKit
}

// GetRequestSource RequestSource为空，返回 ApiCall 类型。
func (kt *Kit) GetRequestSource() enumor.RequestSourceType {
	if len(kt.RequestSource) == 0 {
		return enumor.ApiCall
	}

	return kt.RequestSource
}

// ContextWithRid ...
func (kt *Kit) ContextWithRid() context.Context {
	return context.WithValue(kt.Ctx, constant.RidKey, kt.Rid)
}

// CtxWithTimeoutMS create a new context with basic info and timout configuration.
func (kt *Kit) CtxWithTimeoutMS(timeoutMS int) context.CancelFunc {
	ctx := context.WithValue(context.TODO(), constant.RidKey, kt.Rid)
	var cancel context.CancelFunc
	kt.Ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	return cancel
}

// CtxBackgroundWithCancel create a new context with backgroup and cancel function
func (kt *Kit) CtxBackgroundWithCancel() context.CancelFunc {
	var cancel context.CancelFunc
	kt.Ctx, cancel = context.WithCancel(kt.Ctx)
	return cancel
}

// Validate context kit.
func (kt *Kit) Validate() error {
	if kt.Ctx == nil {
		return errors.New("context is required")
	}

	if len(kt.User) == 0 {
		return errors.New("user is required")
	}

	ridLen := len(kt.Rid)
	if ridLen == 0 {
		return errors.New("rid is required")
	}

	if ridLen < 16 || ridLen > 50 {
		return errors.New("rid length not right, length should in 16~50")
	}

	if len(kt.AppCode) == 0 {
		return errors.New("app code is required")
	}

	// 非Web来源的请求，多租户开关打开时再校验租户ID
	if kt.AppCode != constant.WebSourceAppCode && cc.TenantEnable() && len(kt.TenantID) == 0 {
		return errors.New("tenant id is required")
	}

	return nil
}

// Header generate header by kit
func (kt *Kit) Header() http.Header {
	return http.Header{
		constant.UserKey:          []string{kt.User},
		constant.RidKey:           []string{kt.Rid},
		constant.AppCodeKey:       []string{kt.AppCode},
		constant.TenantIDKey:      []string{kt.TenantID},
		constant.RequestSourceKey: []string{string(kt.RequestSource)},
	}
}

// FromHeader http request header to context kit and validate.
func FromHeader(ctx context.Context, header http.Header) (*Kit, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	kt := &Kit{
		Ctx:           ctx,
		User:          header.Get(constant.UserKey),
		Rid:           header.Get(constant.RidKey),
		AppCode:       header.Get(constant.AppCodeKey),
		TenantID:      header.Get(constant.TenantIDKey),
		RequestSource: enumor.RequestSourceType(header.Get(constant.RequestSourceKey)),
	}

	if kt.Ctx.Value(constant.RidKey) == nil {
		kt.Ctx = context.WithValue(kt.Ctx, constant.RidKey, kt.Rid)
	}

	if err := kt.Validate(); err != nil {
		return nil, err
	}

	return kt, nil
}

// SetBackendTenantID 设置后端操作的租户id
func (kt *Kit) SetBackendTenantID() {
	// 如果开启多租户，设置租户id为system，不开启则设置为default
	if cc.TenantEnable() {
		kt.TenantID = constant.SystemTenantID
		return
	}

	kt.TenantID = constant.DefaultTenantID
}
