package notice

import (
	"errors"
	"net/http"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/rest"
	pkgnotice "hcm/pkg/thirdparty/api-gateway/notice"
)

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {
	svc := &service{
		client: c.NoticeCli,
	}

	h := rest.NewHandler()

	h.Add("GetCurrentAnnouncements", http.MethodGet,
		"/notice/get_current_announcements", svc.GetCurrentAnnouncements)

	h.Load(c.WebService)
}

type service struct {
	client pkgnotice.Client
}

// GetCurrentAnnouncements ...
func (s service) GetCurrentAnnouncements(cts *rest.Contexts) (interface{}, error) {
	if !cc.WebServer().Notice.Enable {
		return nil, errors.New("notification is not enabled")
	}
	params := make(map[string]string)
	for key, val := range cts.Request.Request.URL.Query() {
		params[key] = val[0]
	}
	params["platform"] = cc.WebServer().Notice.AppCode
	return s.client.GetCurAnn(cts.Kit, params)
}
