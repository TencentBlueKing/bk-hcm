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

	"hcm/pkg/api/discovery"
	"hcm/pkg/cc"
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

	discoverServices := []cc.Name{cc.CloudServerName}
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

	return restful.NewContainer().Add(ws)
}

// Do proxy restful request to target server.
func (p *proxy) Do(req *restful.Request, resp *restful.Response) {
	r, w := req.Request, resp.ResponseWriter

	p.proxyRequest(req, w)

	rid := r.Header.Get(constant.RidKey)
	start := time.Now()

	url := r.URL.Scheme + "://" + r.URL.Host + r.RequestURI
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		logs.Errorf("new proxy request[%s] failed, err: %v, rid: %s", url, err, rid)
		fmt.Fprintf(w, err.Error())
		return
	}

	for k, v := range r.Header {
		if len(v) > 0 {
			proxyReq.Header.Set(k, v[0])
		}
	}

	response, err := p.cli.Do(proxyReq)
	if err != nil {
		logs.Errorf("do request[%s url: %s] failed, err: %v, rid: %s", r.Method, url, err, rid)
		fmt.Fprintf(w, err.Error())
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
		logs.Errorf("response request[url: %s] failed, err: %v, rid: %s", r.RequestURI, err, rid)
		return
	}

	logs.V(4).Infof("cost: %dms, action: %s, status code: %d, user: %s, app code: %s, url: %s, rid: %s",
		time.Since(start).Nanoseconds()/int64(time.Millisecond), r.Method, response.StatusCode,
		r.Header.Get(constant.UserKey), r.Header.Get(constant.AppCodeKey), url, rid)
	return
}

// proxyRequest get request service by url, discover service and proxy request to target server
func (p *proxy) proxyRequest(req *restful.Request, w http.ResponseWriter) {
	var service cc.Name

	// path format: /api/{api_version}/{service}/other
	paths := strings.Split(req.Request.URL.Path, "/")
	if len(paths) > 3 {
		servicePath := paths[3]
		switch servicePath {
		case "cloud":
			service = cc.CloudServerName
		}
	} else {
		logs.Errorf("received url path length not conform to the regulations, path: %s", req.Request.URL.Path)
		fmt.Fprintf(w, errf.New(http.StatusNotFound, "Not Found").Error())
		return
	}

	discovery, exists := p.discovery[service]
	if !exists {
		logs.Errorf("received request service %s is not supported, path: %s", service, req.Request.URL.Path)
		fmt.Fprintf(w, errf.New(http.StatusNotFound, "Service Not Supported").Error())
		return
	}

	servers, err := discovery.GetServers()
	if err != nil {
		logs.Errorf("received request service %s has no servers, path: %s", service, req.Request.URL.Path)
		fmt.Fprintf(w, errf.New(http.StatusNotFound, "Servers Not Found").Error())
		return
	}

	if strings.HasPrefix(servers[0], "https://") {
		req.Request.URL.Host = servers[0][8:]
		req.Request.URL.Scheme = "https"
	} else {
		req.Request.URL.Host = servers[0][7:]
		req.Request.URL.Scheme = "http"
	}
}
