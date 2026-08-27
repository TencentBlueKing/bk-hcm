/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package metrics

import (
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"strings"

	"hcm/pkg/criteria/errf"
)

// ErrType is a normalized error category used as a metric label value.
// It is a closed enum to keep label cardinality bounded.
type ErrType string

// Closed enumeration of err_type values used by *_fail_total metrics.
// New values MUST be added here intentionally; arbitrary error texts must
// never be passed as label values.
const (
	ErrTypeOK           ErrType = "ok"
	ErrTypeTimeout      ErrType = "timeout"
	ErrTypeNetwork      ErrType = "network"
	ErrTypeCloudError   ErrType = "cloud_error"
	ErrTypeHCMError     ErrType = "hcm_error"
	ErrTypeInvalidParam ErrType = "invalid_param"
	ErrTypeAuth         ErrType = "auth"
	ErrTypeCancel       ErrType = "cancel"
	ErrTypeUnknown      ErrType = "unknown"
)

// String returns the metric label value for the err_type.
func (e ErrType) String() string {
	return string(e)
}

// ClassifyError maps an error to a normalized ErrType. nil error returns
// ErrTypeOK; callers should treat ErrTypeOK as "no fail metric should be
// emitted". Unknown / unclassified errors fall back to ErrTypeUnknown.
//
// The mapping rules (evaluated in order):
//   - context.DeadlineExceeded / "deadline exceeded" / "timeout" → timeout
//   - context.Canceled / "context canceled" → cancel
//   - net.Error with Timeout() true → timeout
//   - net.Error otherwise / DNS errors / "connection refused" / EOF → network
//   - errf.ErrorF with InvalidParameter / DecodeRequestFailed → invalid_param
//   - errf.ErrorF with PermissionDenied / DoAuthorizeFailed / UserNoAppAccess → auth
//   - errf.ErrorF with CloudVendorError → cloud_error
//   - errf.ErrorF other codes → hcm_error
//   - cloud vendor SDK errors (tcloud / huawei / aws / azure / gcp) → cloud_error
//   - common message patterns ("4xx", "5xx", "context canceled" 等) → 对应类型
//   - everything else → unknown
func ClassifyError(err error) ErrType {
	if err == nil {
		return ErrTypeOK
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTypeTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrTypeCancel
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrTypeTimeout
		}
		return ErrTypeNetwork
	}

	var ef *errf.ErrorF
	if errors.As(err, &ef) {
		return classifyErrorF(ef)
	}

	if et := classifyCloudVendorError(err); et != ErrTypeOK {
		return et
	}

	if et := classifyByMessage(err.Error()); et != ErrTypeOK {
		return et
	}

	return ErrTypeUnknown
}

func classifyErrorF(ef *errf.ErrorF) ErrType {
	switch ef.Code {
	case errf.InvalidParameter, errf.DecodeRequestFailed:
		return ErrTypeInvalidParam
	case errf.PermissionDenied, errf.DoAuthorizeFailed, errf.UserNoAppAccess:
		return ErrTypeAuth
	case errf.CloudVendorError:
		return ErrTypeCloudError
	case errf.OK:
		return ErrTypeOK
	default:
		return ErrTypeHCMError
	}
}

// awsErrorLike matches the github.com/aws/aws-sdk-go/aws/awserr.Error
// interface contract. We declare the interface locally so that pkg/metrics
// stays free of cloud-vendor SDK imports; any AWS SDK error implements
// this shape implicitly.
type awsErrorLike interface {
	Code() string
	Message() string
	OrigErr() error
}

// tcloudErrorLike matches github.com/tencentcloud/tencentcloud-sdk-go's
// TencentCloudSDKError shape (GetCode / GetMessage / GetRequestId). All
// TCloud SDK errors satisfy this interface.
type tcloudErrorLike interface {
	GetCode() string
	GetMessage() string
	GetRequestId() string
}

// cloudVendorErrorTypes is a closed allowlist of cloud-vendor SDK error
// type names (as returned by reflect.TypeOf(err).String()). Listed types
// MUST be classified as ErrTypeCloudError. Keep this list small and
// vendor-rooted to avoid cardinality leak.
var cloudVendorErrorTypes = map[string]struct{}{
	// Huawei Cloud SDK v3 (github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr)
	"*sdkerr.ServiceResponseError": {},
	"*sdkerr.CredentialsTypeError": {},
	"*sdkerr.RequestError":         {},
	// Azure SDK (github.com/Azure/azure-sdk-for-go/sdk/azcore)
	"*azcore.ResponseError": {},
	// GCP API client (google.golang.org/api/googleapi)
	"*googleapi.Error": {},
}

// classifyCloudVendorError detects common cloud vendor SDK error types and
// classifies them as ErrTypeCloudError. Detection avoids importing any
// vendor SDK package by combining:
//
//   - interface-shape assertion for AWS (awserr.Error) and TCloud
//     (TencentCloudSDKError), since both expose stable method sets;
//   - reflect type-name allowlist for Huawei / Azure / GCP, whose error
//     types are concrete structs without a shared interface.
//
// Returns ErrTypeOK when err is not a recognized vendor SDK error so the
// caller can fall through to the next classification rule.
func classifyCloudVendorError(err error) ErrType {
	var awsErr awsErrorLike
	if errors.As(err, &awsErr) {
		return ErrTypeCloudError
	}

	var tcErr tcloudErrorLike
	if errors.As(err, &tcErr) {
		return ErrTypeCloudError
	}

	if _, ok := cloudVendorErrorTypes[reflect.TypeOf(err).String()]; ok {
		return ErrTypeCloudError
	}

	return ErrTypeOK
}

func classifyByMessage(msg string) ErrType {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "context canceled"), strings.Contains(lower, "context cancelled"):
		return ErrTypeCancel
	case strings.Contains(lower, "timeout"), strings.Contains(lower, "deadline exceeded"):
		return ErrTypeTimeout
	case strings.Contains(lower, "no such host"),
		strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "broken pipe"),
		strings.Contains(lower, "eof"):
		return ErrTypeNetwork
	// 兜底：当云 SDK error 被深层包装、type assertion 失效时，按错误文本
	// 中常见的厂商签名词识别为 cloud_error。关键字尽量保守，避免误伤业务
	// 错误。
	case strings.Contains(lower, "tencentcloudsdkerror"),
		strings.Contains(lower, "[apigateway-error]"),
		strings.Contains(lower, "huaweicloudsdkerror"),
		strings.Contains(lower, "responseerror"),
		strings.Contains(lower, "googleapi: error"),
		strings.Contains(lower, "awserr:"),
		strings.Contains(lower, "requestid:"):
		return ErrTypeCloudError
	default:
		return ErrTypeOK
	}
}

// ClassifyHTTPStatusCode maps an HTTP status code to a normalized ErrType.
// It is used by middleware-style打点 where only the response status code is
// available. 2xx returns ErrTypeOK (not a failure).
func ClassifyHTTPStatusCode(code int) ErrType {
	switch {
	case code >= 200 && code < 400:
		return ErrTypeOK
	case code == http.StatusUnauthorized, code == http.StatusForbidden:
		return ErrTypeAuth
	case code == http.StatusBadRequest, code == http.StatusUnprocessableEntity,
		code == http.StatusNotAcceptable:
		return ErrTypeInvalidParam
	case code == http.StatusRequestTimeout, code == http.StatusGatewayTimeout:
		return ErrTypeTimeout
	case code >= 500:
		return ErrTypeHCMError
	default:
		return ErrTypeUnknown
	}
}
