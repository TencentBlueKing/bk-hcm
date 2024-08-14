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

package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/client/discovery"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"

	"github.com/emicklei/go-restful/v3"
)

// proxy all server's rest proxy.
type proxy struct {
	discovery map[cc.Name]*discovery.APIDiscovery
	cli       *http.Client
}

// newProxy create new rest proxy.
func newProxy(dis serviced.Discover, cli *http.Client) (*proxy, error) {
	apiDiscovery := make(map[cc.Name]*discovery.APIDiscovery)

	discoverServices := []cc.Name{cc.CloudServerName, cc.AccountServerName}
	for _, service := range discoverServices {
		apiDiscovery[service] = discovery.NewAPIDiscovery(service, dis)
	}

	p := &proxy{
		discovery: apiDiscovery,
		cli:       cli,
	}

	return p, nil
}

func (p *proxy) apiSet() *restful.Container {
	ws := new(restful.WebService)

	ws.Path("/api/v1")
	ws.Filter(p.restFilter())
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("{.*}").To(p.Do))
	ws.Route(ws.POST("{.*}").To(p.Do))
	ws.Route(ws.PUT("{.*}").To(p.Do))
	ws.Route(ws.DELETE("{.*}").To(p.Do))
	ws.Route(ws.PATCH("{.*}").To(p.Do))

	return restful.NewContainer().Add(ws)
}

// Do proxy restful request to target server.
func (p *proxy) Do(req *restful.Request, resp *restful.Response) {
	r, w := req.Request, resp.ResponseWriter

	rid := r.Header.Get(constant.RidKey)
	start := time.Now()

	if err := p.prepareRequest(req); err != nil {
		_, _ = fmt.Fprintf(w, errf.NewFromErr(http.StatusNotFound, err).Error())
		logs.Errorf("prepare request to proxy failed, err: %v, rid: %s", err, rid)
		return
	}

	url := r.URL.Scheme + "://" + r.URL.Host + r.RequestURI
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		logs.Errorf("new proxy request[%s] failed, err: %v, rid: %s", url, err, rid)
		return
	}

	for k, v := range r.Header {
		if len(v) > 0 {
			proxyReq.Header.Set(k, v[0])
		}
	}

	response, err := p.cli.Do(proxyReq)
	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		logs.Errorf("do request[%s url: %s] failed, err: %v, rid: %s", r.Method, url, err, rid)
		return
	}
	defer response.Body.Close()

	for k, v := range response.Header {
		if len(v) > 0 {
			resp.Header().Set(k, v[0])
		}
	}

	resp.ResponseWriter.WriteHeader(response.StatusCode)

	if _, err := io.Copy(resp, response.Body); err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		logs.Errorf("response request[url: %s] failed, err: %v, rid: %s", r.RequestURI, err, rid)
		return
	}

	if logs.V(4) {
		logs.Infof("cost: %dms, action: %s, status code: %d, user: %s, app code: %s, url: %s, rid: %s",
			time.Since(start).Nanoseconds()/int64(time.Millisecond), r.Method, response.StatusCode,
			r.Header.Get(constant.UserKey), r.Header.Get(constant.AppCodeKey), url, rid)
	}

	return
}

// prepareRequest get request service by url, discover service and proxy request to target server
func (p *proxy) prepareRequest(req *restful.Request) error {
	var service cc.Name

	// path format: /api/{api_version}/{service}/other
	paths := strings.Split(req.Request.URL.Path, "/")
	if len(paths) <= 3 {
		return fmt.Errorf("received invalid url path: %s", req.Request.URL.Path)
	}

	servicePath := paths[3]
	switch servicePath {
	case "cloud":
		service = cc.CloudServerName
	case "account":
		service = cc.AccountServerName
	default:
		return fmt.Errorf("received unknown url path: %s", req.Request.URL.Path)
	}

	ds, exists := p.discovery[service]
	if !exists {
		return fmt.Errorf("received request service %s is not supported, path: %s", service, req.Request.URL.Path)
	}

	servers, err := ds.GetServers()
	if err != nil {
		return fmt.Errorf("received request to service %s has no servers, path: %s", service, req.Request.URL.Path)
	}

	if strings.HasPrefix(servers[0], "https://") {
		req.Request.URL.Host = servers[0][8:]
		req.Request.URL.Scheme = "https"
	} else {
		req.Request.URL.Host = servers[0][7:]
		req.Request.URL.Scheme = "http"
	}

	return nil
}
