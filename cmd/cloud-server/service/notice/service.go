package notice

import (
	"errors"
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
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

	//h.Add("/notice", h.Handle(svc))
	h.Add("GetCurrentAnnouncements", http.MethodGet,
		"/notice/get_current_announcements", svc.GetCurrentAnnouncements)

	h.Load(c.WebService)
}

type service struct {
	client pkgnotice.Client
}

func (s service) GetCurrentAnnouncements(cts *rest.Contexts) (interface{}, error) {
	if !cc.CloudServer().Notice.Enable {
		return nil, errors.New("notification is not enabled")
	}
	params := make(map[string]string)
	for key, val := range cts.Request.Request.URL.Query() {
		params[key] = val[0]
	}
	params["platform"] = cc.CloudServer().Notice.AppCode
	return s.client.GetCurAnn(cts.Kit, params)
}
