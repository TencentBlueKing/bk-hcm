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

// Package service ...
package service

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	authsvc "hcm/cmd/web-server/service/auth"
	"hcm/cmd/web-server/service/capability"
	"hcm/cmd/web-server/service/cloud/subnet"
	"hcm/cmd/web-server/service/cloud/vpc"
	"hcm/cmd/web-server/service/cmdb"
	"hcm/cmd/web-server/service/itsm"
	"hcm/cmd/web-server/service/notice"
	templateSvc "hcm/cmd/web-server/service/template"
	"hcm/cmd/web-server/service/user"
	"hcm/cmd/web-server/service/version"
	"hcm/pkg/cc"
	apiclient "hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	pkgbkuser "hcm/pkg/thirdparty/api-gateway/bkuser"
	pkgcmdb "hcm/pkg/thirdparty/api-gateway/cmdb"
	pkgitsm "hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/api-gateway/login"
	pkgnotice "hcm/pkg/thirdparty/api-gateway/notice"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/ssl"
	pkgversion "hcm/pkg/version"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the web server's work
type Service struct {
	// client 为调用其他微服务所需的Client集合
	client *apiclient.ClientSet
	// EsbClient 调用接入ESB的第三方系统API集合
	esbClient esb.Client
	// 透传代理请求到其他微服务
	proxy *proxy
	// authorizer 鉴权所需接口集合
	authorizer auth.Authorizer
	// itsmCli itsm client.
	itsmCli pkgitsm.Client
	// noticeCli notification center client
	noticeCli pkgnotice.Client
	cmdbCli   pkgcmdb.Client
	bkUserCli pkgbkuser.Client
	// loginCli login client.
	loginCli login.Client
}

// NewService create a service instance.
func NewService(dis serviced.Discover) (*Service, error) {
	network := cc.WebServer().Network

	var tlsConfig *ssl.TLSConfig
	if network.TLS.Enable() {
		tlsConfig = &ssl.TLSConfig{
			InsecureSkipVerify: network.TLS.InsecureSkipVerify,
			CertFile:           network.TLS.CertFile,
			KeyFile:            network.TLS.KeyFile,
			CAFile:             network.TLS.CAFile,
			Password:           network.TLS.Password,
		}
	}

	// 创建基本的Http请求Client
	httpClient, err := client.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}
	// 使用基本Http Client生成调用依赖其他微服务的Client集合
	apiClientSet := apiclient.NewClientSet(httpClient, dis)

	// 创建ESB Client
	esbConfig := cc.WebServer().Esb
	esbClient, err := esb.NewClient(&esbConfig, metrics.Register())
	if err != nil {
		return nil, err
	}

	itsmCfg := cc.WebServer().Itsm
	itsmCli, err := pkgitsm.NewClient(&itsmCfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	// create authorizer
	authorizer, err := auth.NewAuthorizer(dis, network.TLS)
	if err != nil {
		return nil, err
	}

	// 创建代理
	p, err := newProxy(dis, httpClient)
	if err != nil {
		return nil, err
	}

	bkUserCfg := cc.WebServer().BkUser
	bkUserCli, err := pkgbkuser.NewClient(&bkUserCfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	noticeCli, err := newNotificationClient(bkUserCli)
	if err != nil {
		logs.Errorf("failed to create notice client, err: %v", err)
		return nil, err
	}

	loginCfg := cc.WebServer().Login
	loginCli, err := login.NewClient(&loginCfg, bkUserCli, metrics.Register())
	if err != nil {
		return nil, err
	}

	cmdbCfg := cc.WebServer().Cmdb
	cmdbCli, err := pkgcmdb.NewClient(&cmdbCfg, bkUserCli, metrics.Register())
	if err != nil {
		return nil, err
	}

	return &Service{
		client:     apiClientSet,
		esbClient:  esbClient,
		proxy:      p,
		authorizer: authorizer,
		itsmCli:    itsmCli,
		noticeCli:  noticeCli,
		cmdbCli:    cmdbCli,
		bkUserCli:  bkUserCli,
		loginCli:   loginCli,
	}, nil
}

func newNotificationClient(bkUserCli pkgbkuser.Client) (pkgnotice.Client, error) {
	noticeCfg := cc.WebServer().Notice
	if !noticeCfg.Enable {
		return nil, nil
	}
	noticeCli, err := pkgnotice.NewClient(&noticeCfg.ApiGateway, bkUserCli, metrics.Register())
	if err != nil {
		logs.Errorf("failed to create notice client, err: %v", err)
		return nil, err
	}
	_, err = noticeCli.RegApp(kit.New())
	if err != nil {
		// 无api gateway权限可能会导致注册失败，阻塞服务启动
		logs.Errorf("register notice app failed, err: %v", err)
		return nil, err
	}
	return noticeCli, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {

	root := http.NewServeMux()
	// Basic API 使用net/http 路由处理
	// Healthz
	root.HandleFunc("/healthz", s.Healthz)
	// metric/debug/ctl
	handler.SetCommonHandler(root)

	// 其他使用 go-restful处理
	container := restful.NewContainer()
	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders: []string{
			"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers",
			"Cache-Control", "Content-Language", "Content-Type", "Content-Disposition",
		},
		AllowedHeaders: []string{
			"Accept", "Accept-encoding", "Authorization", "Content-Type", "Dnt",
			"Origin", "User-Agent", "X-Csrftoken", "X-Requested-With", "X-Bkapi-Request-Id", "X-Bk-Tenant-Id",
		},
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		CookiesAllowed: true,
		Container:      container,
	}
	container.Filter(cors.Filter)
	// Add container filter to respond to OPTIONS
	container.Filter(container.OPTIONSFilter)
	container.Add(s.staticFileSet())
	container.Add(s.apiSet())
	container.Add(s.proxyApiSet("/api/v1/cloud"))
	container.Add(s.proxyApiSet("/api/v1/account"))
	container.Add(s.indexSet())

	root.Handle("/", container)

	network := cc.WebServer().Network
	server := &http.Server{
		Addr:    net.JoinHostPort(network.BindIP, strconv.FormatUint(uint64(network.Port), 10)),
		Handler: root,
	}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := ssl.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	logs.Infof("listen restful server on %s with secure(%v) now.", server.Addr, network.TLS.Enable())

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()

			logs.Infof("start shutdown restful server gracefully...")

			ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logs.Errorf("shutdown restful server failed, err: %v", err)
				return
			}
			logs.Infof("shutdown restful server success...")
		}
	}()
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()
	return nil
}

// apiSet 处理Web特有的API
func (s *Service) apiSet() *restful.WebService {
	ws := new(restful.WebService)
	ws.Produces(restful.MIME_JSON)

	ws.Filter(NewCompleteRequestIDFilter())
	// Note: 所有API接口都需要经过用户认证
	ws.Path("/api/v1/web").Filter(
		NewUserAuthenticateFilter(s.loginCli, cc.WebServer().Web.BkLoginUrl, cc.WebServer().Web.BkLoginCookieName),
	)

	c := &capability.Capability{
		WebService: ws,
		ApiClient:  s.client,
		Authorizer: s.authorizer,
		ItsmCli:    s.itsmCli,
		NoticeCli:  s.noticeCli,
		LoginCli:   s.loginCli,
		CmdbCli:    s.cmdbCli,
	}

	user.InitUserService(c)
	cmdb.InitCmdbService(c)
	authsvc.InitAuthService(c)
	vpc.InitVpcService(c)
	subnet.InitService(c)
	itsm.InitService(c)
	version.InitVersionService(c)
	if cc.WebServer().Notice.Enable {
		notice.InitService(c)
	}
	templateSvc.InitTemplateService(c)

	return ws
}

// proxyApiSet 处理代理API
func (s *Service) proxyApiSet(apiPath string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Produces(restful.MIME_JSON)

	ws.Filter(NewCompleteRequestIDFilter())
	// Note: 所有API接口都需要经过用户认证
	ws.Path(apiPath).Filter(
		NewUserAuthenticateFilter(s.loginCli, cc.WebServer().Web.BkLoginUrl, cc.WebServer().Web.BkLoginCookieName),
	)
	ws.Route(ws.GET("{.*}").To(s.proxy.Do))
	ws.Route(ws.POST("{.*}").To(s.proxy.Do))
	ws.Route(ws.PUT("{.*}").To(s.proxy.Do))
	ws.Route(ws.PATCH("{.*}").To(s.proxy.Do))
	ws.Route(ws.DELETE("{.*}").To(s.proxy.Do))

	return ws
}

func (s *Service) staticFileSet() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/static")
	// 静态资源
	ws.Route(ws.GET("/{subpath:*}").To(s.staticFileHandleFunc))

	return ws
}

func (s *Service) staticFileHandleFunc(req *restful.Request, resp *restful.Response) {
	actual := path.Join(cc.WebServer().Web.StaticFileDirPath, req.PathParameter("subpath"))
	http.ServeFile(resp.ResponseWriter, req.Request, actual)
}

func (s *Service) indexSet() *restful.WebService {
	ws := new(restful.WebService)
	// 所有前缀未匹配到的URL都将返回index.html
	ws.Route(ws.GET("/").To(s.indexHandleFunc))
	ws.Route(ws.GET("/{subpath:*}").To(s.indexHandleFunc))

	return ws
}

// Index 首页
func (s *Service) indexHandleFunc(req *restful.Request, resp *restful.Response) {
	indexHtmlFile := filepath.Join(cc.WebServer().Web.StaticFileDirPath, "index.html")

	// Return a 404 if the template doesn't exist
	_, err := os.Stat(indexHtmlFile)
	if err != nil {
		if os.IsNotExist(err) {
			resp.WriteErrorString(http.StatusNotFound, "the template don't exists")
			return
		}
	}

	// 解析模板
	tmpl, err := template.ParseFiles(indexHtmlFile)
	if err != nil {
		// Log the detailed error
		logs.Errorf("the template of indexHandleFunc.html gave err: %v", err)
		// Return a generic "Internal Server Error" message
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 渲染模板
	content := map[string]interface{}{
		"BK_LOGIN_URL":                cc.WebServer().Web.BkLoginUrl,
		"BK_COMPONENT_API_URL":        cc.WebServer().Web.BkComponentApiUrl,
		"BK_ITSM_URL":                 cc.WebServer().Web.BkItsmUrl,
		"BK_DOMAIN":                   cc.WebServer().Web.BkDomain,
		"VERSION":                     pkgversion.VERSION,
		"BK_CMDB_CREATE_BIZ_URL":      cc.WebServer().Web.BkCmdbCreateBizUrl,
		"BK_CMDB_CREATE_BIZ_DOCS_URL": cc.WebServer().Web.BkCmdbCreateBizDocsUrl,
		"ENABLE_CLOUD_SELECTION":      cc.WebServer().Web.EnableCloudSelection,
		"ENABLE_ACCOUNT_BILL":         cc.WebServer().Web.EnableAccountBill,
		"ENABLE_NOTICE":               cc.WebServer().Notice.Enable,
		"USER_MANAGE_URL":             cc.WebServer().Web.BkUserManageUrl,
	}
	err = tmpl.Execute(resp.ResponseWriter, content)
	if err != nil {
		logs.Errorf("the template of indexHandleFunc.html render err: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

// Healthz service health check.
func (s *Service) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(r.Context(), cc.WebServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
