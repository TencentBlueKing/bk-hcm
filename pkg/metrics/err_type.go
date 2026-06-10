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
// The mapping rules:
//   - context.DeadlineExceeded / "deadline exceeded" / "timeout" → timeout
//   - context.Canceled / "context canceled" → cancel
//   - net.Error with Timeout() true → timeout
//   - net.Error otherwise / DNS errors / "connection refused" / EOF → network
//   - errf.ErrorF with InvalidParameter / DecodeRequestFailed → invalid_param
//   - errf.ErrorF with PermissionDenied / DoAuthorizeFailed / UserNoAppAccess → auth
//   - errf.ErrorF with CloudVendorError → cloud_error
//   - errf.ErrorF other codes → hcm_error
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
