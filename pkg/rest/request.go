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

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// VerbType http request verb type
type VerbType string

// http request method.
const (
	PUT    VerbType = http.MethodPut
	POST   VerbType = http.MethodPost
	GET    VerbType = http.MethodGet
	DELETE VerbType = http.MethodDelete
	PATCH  VerbType = http.MethodPatch
	HEAD   VerbType = http.MethodHead
)

// ContentType http request content type
type ContentType string

// ContentType http request content type
const (
	FormDataContent ContentType = "application/x-www-form-urlencoded"
	JsonContent     ContentType = "application/json"
)

// Request http request.
type Request struct {
	// http client.
	client *Client

	// request capability.
	capability *client.Capability

	verb    VerbType
	params  url.Values
	headers http.Header
	body    []byte
	ctx     context.Context

	// prefixed url
	baseURL string
	// sub path of the url, will be appended to baseURL
	subPath string
	// sub path format args
	subPathArgs []interface{}

	// metric additional labels
	metricDimension string

	// request timeout value
	timeout time.Duration

	// contentType http content type
	contentType ContentType

	err error
}

// WithMetricDimension add this request a addition dimension value, which helps us to separate
// request metrics with a dimension label.
func (r *Request) WithMetricDimension(value string) *Request {
	r.metricDimension = value
	return r
}

// WithParams add params to request.
func (r *Request) WithParams(params map[string]string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value)
	}
	return r
}

// WithParam add param to request.
func (r *Request) WithParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[paramName] = append(r.params[paramName], value)
	return r
}

// WithParamsFromURL add params to request from url.
func (r *Request) WithParamsFromURL(u *url.URL) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	params := u.Query()
	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value...)
	}
	return r
}

// WithHeaders add header to request.
func (r *Request) WithHeaders(header http.Header) *Request {
	if r.headers == nil {
		r.headers = header
		return r
	}

	for key, values := range header {
		for _, v := range values {
			r.headers.Add(key, v)
		}
	}
	return r
}

// WithContext add context to request.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// WithTimeout add timeout to request.
func (r *Request) WithTimeout(d time.Duration) *Request {
	r.timeout = d
	return r
}

// SubResourcef add subPath and subPath's args to request.
func (r *Request) SubResourcef(subPath string, args ...interface{}) *Request {
	r.subPathArgs = args
	return r.subResource(subPath)
}

// subResource add subPath to request.
func (r *Request) subResource(subPath string) *Request {
	subPath = strings.TrimLeft(subPath, "/")
	r.subPath = subPath
	return r
}

// WithContentType add content type to request.
func (r *Request) WithContentType(contentType ContentType) *Request {
	r.contentType = contentType
	return r
}

// Body add body to request.
func (r *Request) Body(body interface{}) *Request {
	if body == nil {
		r.body = []byte("")
		return r
	}

	valueOf := reflect.ValueOf(body)
	switch valueOf.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		if valueOf.IsNil() {
			r.body = []byte("")
			return r
		}
		break
	case reflect.String:
		r.body = []byte(body.(string))
		return r

	case reflect.Struct:
		break

	default:
		r.err = errors.New("body should be one of interface, map, pointer or slice value")
		r.body = []byte("")
		return r
	}

	data, err := json.Marshal(body)
	if err != nil {
		r.err = err
		r.body = []byte("")
		return r
	}

	r.body = data
	return r
}

// WrapURL get http complete url from request.
func (r *Request) WrapURL() *url.URL {
	finalURL := &url.URL{}
	if len(r.baseURL) != 0 {
		u, err := url.Parse(r.baseURL)
		if err != nil {
			r.err = err
			return new(url.URL)
		}
		*finalURL = *u
	}

	if len(r.subPathArgs) > 0 {
		finalURL.Path = finalURL.Path + fmt.Sprintf(r.subPath, r.subPathArgs...)
	} else {
		finalURL.Path = finalURL.Path + r.subPath
	}

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}

	finalURL.RawQuery = query.Encode()
	return finalURL
}

// checkToleranceLatency check request toleranceLatency.
func (r *Request) checkToleranceLatency(start *time.Time, url string, rid string) {
	if time.Since(*start) < r.capability.ToleranceLatencyTime {
		return
	}

	if r.isToleranceLatencyExclusionURL(url) {
		return
	}

	// request time larger than the maxToleranceLatencyTime time, then log the request
	logs.Infof("http request exceeded max latency time. cost: %d ms, code: %s, user: %s, %s, "+
		"url: %s, body: %s, rid: %s", time.Since(*start)/time.Millisecond, r.headers.Get(constant.AppCodeKey),
		r.headers.Get(constant.UserKey), r.verb, url, r.body, rid)
}

// isToleranceLatencyExclusionURL judge url if need to checkToleranceLatency.
func (r *Request) isToleranceLatencyExclusionURL(url string) bool {
	exclusionURL := make([]string, 0)

	for _, eurl := range exclusionURL {
		if strings.Contains(url, eurl) {
			return true
		}
	}

	return false
}

// Result http response result.
type Result struct {
	Rid        string
	Body       []byte
	Err        error
	StatusCode int
	Status     string
	Header     http.Header
}

// Into parse body to obj.
func (r *Result) Into(obj interface{}) error {
	if nil != r.Err {
		return r.Err
	}

	if 0 != len(r.Body) {
		err := json.Unmarshal(r.Body, obj)
		if nil != err {
			if r.StatusCode >= 300 {
				return fmt.Errorf("http request err: %s", string(r.Body))
			}

			logs.Errorf("invalid response body, unmarshal json failed, reply: %s, err: %s",
				r.Body, err.Error())
			return fmt.Errorf("http response err: %v, raw data: %s", err, r.Body)
		}
	} else if r.StatusCode >= 300 {
		return fmt.Errorf("http request failed: %s", r.Status)
	}
	return nil
}

const maxLatency = 200 * time.Millisecond

// tryThrottle try throttle
func (r *Request) tryThrottle(url string) {
	now := time.Now()

	if latency := time.Since(now); latency > maxLatency {
		logs.Infof("Throttling request took %d ms, verb: %s, request: %s", latency, r.verb, url)
	}
}

// Do http request do.
func (r *Request) Do() *Result {
	result := new(Result)

	rid := ridFromContext(r.ctx)
	if rid == "" {
		rid = r.headers.Get(constant.RidKey)
	}

	if r.err != nil {
		result.Err = r.err
		return result
	}

	client := r.capability.Client
	if client == nil {
		client = http.DefaultClient
	}

	hosts, err := r.capability.Discover.GetServers()
	if err != nil {
		result.Err = err
		return result
	}

	maxRetryCycle := 3
	for try := 0; try < maxRetryCycle; try++ {
		for index, host := range hosts {
			result, isComplete := r.doWithHost(client, host, try+index, rid)
			if isComplete {
				return result
			}
		}
	}

	result.Err = errors.New("unexpected error")
	return result
}

// doWithHost http request do with specific host.
func (r *Request) doWithHost(client client.HTTPClient, host string, retries int, rid string) (*Result, bool) {
	contentType := r.contentType

	switch r.contentType {
	case FormDataContent:
		r.body = []byte(r.params.Encode())
		r.params = url.Values{}
	default:
		contentType = JsonContent
	}

	url := host + r.WrapURL().String()
	req, err := r.getRequest(url, contentType)
	if err != nil {
		return &Result{Err: err, Rid: rid}, true
	}

	if retries > 0 {
		r.tryThrottle(url)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		// "Connection reset by peer" is a special err which in most scenario is a transient error.
		// Which means that we can retry it. And so does the GET operation.
		// While the other "write" operation can not simply retry it again, because they are not idempotent.
		logs.Errorf("http request %s %s with body %s, but %v, rid: %s", string(r.verb), url, r.body, err, rid)
		r.checkToleranceLatency(&start, url, rid)
		if !isConnectionReset(err) || r.verb != GET {
			return &Result{Err: err, Rid: rid}, true
		}

		// retry now
		time.Sleep(20 * time.Millisecond)
		return nil, false
	}

	// collect request metrics
	if r.client.requestDuration != nil {
		labels := prometheus.Labels{
			"handler":     r.subPath,
			"status_code": strconv.Itoa(resp.StatusCode),
			"dimension":   r.metricDimension,
		}

		r.client.requestDuration.With(labels).Observe(float64(time.Since(start) / time.Millisecond))
	}

	// record latency if needed
	r.checkToleranceLatency(&start, url, rid)

	var body []byte
	if resp.Body != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				// retry now
				time.Sleep(20 * time.Millisecond)
				return nil, false
			}
			logs.Errorf("http request %s %s with body %s, err: %v, rid: %s", string(r.verb), url, r.body,
				err, rid)
			return &Result{Err: err, Rid: rid}, true
		}
		body = data
	}

	if logs.V(4) {
		logs.Infof("http request cost: %dms, %s %s with body %s, response status: %s, response body: %s, rid: "+
			"%s", time.Since(start)/time.Millisecond, string(r.verb), url, r.body, resp.Status, body, rid)
	}

	return &Result{
		Rid:        rid,
		Body:       body,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
	}, true
}

func (r *Request) getRequest(url string, contentType ContentType) (*http.Request, error) {
	req, err := http.NewRequest(string(r.verb), url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}

	if r.ctx != nil {
		req.WithContext(r.ctx)
	}

	req.Header = cloneHeader(r.headers)
	if len(req.Header) == 0 {
		req.Header = make(http.Header)
	}

	req.Header.Del("Accept-Encoding")
	req.Header.Set("Content-Type", string(contentType))
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// isConnectionReset Returns if the given err is "connection reset by peer" error.
func isConnectionReset(err error) bool {
	if urlErr, ok := err.(*url.Error); ok {
		err = urlErr.Err
	}

	if opErr, ok := err.(*net.OpError); ok {
		err = opErr.Err
	}

	if osErr, ok := err.(*os.SyscallError); ok {
		err = osErr.Err
	}

	if errno, ok := err.(syscall.Errno); ok && errno == syscall.ECONNRESET {
		return true
	}

	return false
}

// ridFromContext get request id from context.
func ridFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid := ctx.Value(constant.RidKey)
	ridValue, ok := rid.(string)
	if ok == true {
		return ridValue
	}
	return ""
}

func cloneHeader(src http.Header) http.Header {
	tar := http.Header{}
	for key := range src {
		tar.Set(key, src.Get(key))
	}
	return tar
}
