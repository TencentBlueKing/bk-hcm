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
	"regexp"
	"strings"
)

// maxNormalizedSegments caps the number of path segments kept in a
// normalized endpoint label, so deep URLs cannot inflate cardinality. The
// remaining suffix is collapsed to "/_".
const maxNormalizedSegments = 6

// numericSegment matches purely numeric segments (e.g. biz_id "123").
var numericSegment = regexp.MustCompile(`^[0-9]+$`)

// uuidSegment matches RFC4122-shaped uuids.
var uuidSegment = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// longHexSegment matches 24+ char hex / alnum segments, typical of mongo
// object ids and hcm internal IDs.
var longHexSegment = regexp.MustCompile(`^[0-9a-fA-F]{24,}$`)

// cloudIDSegment matches common cloud resource IDs of the form
// "<prefix>-<alnum suffix>" with a short alphabetic prefix and a long
// alnum suffix, e.g. "lb-xxxxxxxx", "ins-xxxxxxxx", "vpc-xxxxxxxx".
var cloudIDSegment = regexp.MustCompile(`^[a-zA-Z]{1,5}-[0-9a-zA-Z]{6,}$`)

// NormalizeEndpoint returns a low-cardinality version of `path` suitable for
// use as the `endpoint` label of http_request_* metrics. It is intended for
// proxy-style entry points (such as api-server) that don't have static route
// templates. Production HTTP services with explicit route templates SHOULD
// use the route template directly instead of calling this helper.
//
// Rules:
//   - leading/trailing "/" are normalized;
//   - up to maxNormalizedSegments segments are kept; the rest is collapsed
//     to "/_";
//   - segments that look like ids (numeric / UUID / long hex / cloud
//     resource id) are replaced with "{id}".
func NormalizeEndpoint(path string) string {
	if path == "" {
		return "/"
	}
	// strip query string (paranoia; URL.Path normally has none).
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "/"
	}
	parts := strings.Split(trimmed, "/")
	truncated := false
	if len(parts) > maxNormalizedSegments {
		parts = parts[:maxNormalizedSegments]
		truncated = true
	}
	for i, seg := range parts {
		if isIDLike(seg) {
			parts[i] = "{id}"
		}
	}
	out := "/" + strings.Join(parts, "/")
	if truncated {
		out += "/_"
	}
	return out
}

func isIDLike(seg string) bool {
	switch {
	case numericSegment.MatchString(seg):
		return true
	case uuidSegment.MatchString(seg):
		return true
	case longHexSegment.MatchString(seg):
		return true
	case cloudIDSegment.MatchString(seg):
		return true
	default:
		return false
	}
}
