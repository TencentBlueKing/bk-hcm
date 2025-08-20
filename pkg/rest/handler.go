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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/emicklei/go-restful/v3"
	prm "github.com/prometheus/client_golang/prometheus"
)

var once sync.Once

// NewHandler create a new restfull handler
func NewHandler() *Handler {
	once.Do(func() {
		initMetric()
	})

	return &Handler{
		actions: make([]*action, 0),
	}
}

// action defines a http request action
type action struct {
	Verb    string
	Path    string
	Alias   string
	Handler func(contexts *Contexts) (reply interface{}, err error)
}

// Handler contains all the restfull http handler actions
type Handler struct {
	rootPath string
	actions  []*action
}

// Path defines the root path of the handler.
func (r *Handler) Path(path string) {
	r.rootPath = strings.TrimRight(path, "/")
}

// Add add a http handler
func (r *Handler) Add(alias, verb, path string, handler func(cts *Contexts) (interface{}, error)) {
	switch verb {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
	default:
		panic(fmt.Sprintf("add http handler failed, inavlid http verb: %s.", verb))
	}

	if len(path) == 0 {
		panic("add http handler, but got empty http path.")
	}

	if handler == nil {
		panic("add http handler, but got nil http handler")
	}

	r.actions = append(r.actions, &action{Verb: verb, Path: path, Alias: alias, Handler: handler})
}

// Load add actions to the restful webservice, and add to the rest container.
func (r *Handler) Load(ws *restful.WebService) {
	if len(r.actions) == 0 {
		panic("no actions has been added, can not load the handler")
	}

	aliasMap := make(map[string]string, len(r.actions))
	for _, action := range r.actions {
		path := action.Path
		if r.rootPath != "" {
			path = fmt.Sprintf("%s/%s", r.rootPath, strings.TrimLeft(action.Path, "/"))
		}
		if previousPath, exists := aliasMap[action.Alias]; exists {
			logs.Errorf("duplicate handler alias %s for path: %s, previous: %s", action.Alias, path, previousPath)
		} else {
			aliasMap[action.Alias] = path
		}

		switch action.Verb {
		case http.MethodPost:
			ws.Route(ws.POST(path).To(r.wrapperAction(action)))
		case http.MethodDelete:
			ws.Route(ws.DELETE(path).To(r.wrapperAction(action)))
		case http.MethodPut:
			ws.Route(ws.PUT(path).To(r.wrapperAction(action)))
		case http.MethodGet:
			ws.Route(ws.GET(path).To(r.wrapperAction(action)))
		case http.MethodPatch:
			ws.Route(ws.PATCH(path).To(r.wrapperAction(action)))
		default:
			panic(fmt.Sprintf("add handler to webservice, but got unsupport verb: %s .", action.Verb))
		}
	}

	return
}

func (r *Handler) wrapperAction(action *action) func(req *restful.Request, resp *restful.Response) {
	return func(req *restful.Request, resp *restful.Response) {
		cts := new(Contexts)
		cts.Request = req
		cts.resp = resp

		kt, err := kit.FromHeader(req.Request.Context(), req.Request.Header)
		if err != nil {
			rid := req.Request.Header.Get(constant.RidKey)
			logs.Errorf("invalid request for %s, err: %v, rid: %s", action.Alias, err, rid)
			cts.WithStatusCode(http.StatusBadRequest)
			cts.respError(err)
			restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID}).Inc()
			return
		}

		defer func() {
			if fatalErr := recover(); fatalErr != nil {
				cts.respError(fmt.Errorf("panic err: %v", fatalErr))
				logs.Errorf("[hcm server panic], err: %v, rid: %s, debug strace: %s", fatalErr, kt.Rid, debug.Stack())
				logs.CloseLogs()
			}
		}()

		cts.Kit = kt

		// print request log when log level is 4 or request is write request
		if (bool(logs.V(4)) || (!strings.Contains(req.Request.URL.Path, "/list/") &&
			!strings.Contains(req.Request.URL.Path, "/find/"))) && req.Request.Body != nil {

			byt, err := ioutil.ReadAll(req.Request.Body)
			if err != nil {
				logs.Errorf("restful request %s peek failed, err: %v, rid: %s", action.Alias, err, cts.Kit.Rid)

				cts.WithStatusCode(http.StatusBadRequest)
				cts.respError(errf.NewFromErr(errf.InvalidParameter, err))
				restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID}).Inc()
				return
			}

			req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

			compactJson := new(bytes.Buffer)
			compactBody := string(byt)
			if err := json.Compact(compactJson, byt); err == nil {
				compactBody = compactJson.String()
			}
			logs.Infof("%s received restful request, body: %s, rid: %s", action.Alias, compactBody, kt.Rid)
		}

		start := time.Now()
		reply, err := action.Handler(cts)
		if err != nil {
			if logs.V(2) {
				logs.Errorf("do restful request %s failed, err: %v, rid: %s", action.Alias, err, cts.Kit.Rid)
			}

			if reply != nil {
				cts.respErrorWithEntity(reply, err)
			} else {
				cts.respError(err)
			}

			restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID}).Inc()
			return
		}

		cts.respEntity(reply)

		restMetric.lagMS.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID}).
			Observe(float64(time.Since(start).Milliseconds()))
	}
}
