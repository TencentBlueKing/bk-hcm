package image

import (
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/hc-service/image"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"net/http"
)

func (svc *imageSvc) initTCloudImageService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("ListImage", http.MethodPost, "/vendors/tcloud/images/list", svc.ListImage)

	h.Load(cap.WebService)
}

// ListImage ...
func (svc *imageSvc) ListImage(cts *rest.Contexts) (interface{}, error) {

	req := new(image.TCloudImageListOption)
	err := cts.DecodeInto(req)
	if err != nil {
		return nil, err
	}
	err = req.Validate()
	if err != nil {
		return nil, err
	}
	cli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	result, err := cli.ListImage(cts.Kit, req.TCloudImageListOption)
	if err != nil {
		logs.Errorf("list images failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
