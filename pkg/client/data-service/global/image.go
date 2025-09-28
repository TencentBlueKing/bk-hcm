package global

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
)

// ListImage 查询公共镜像列表
func (rc *restClient) ListImage(kt *kit.Kit, request *core.ListReq) (
	*dataproto.ListResult, error) {

	resp := new(core.BaseResp[*dataproto.ListResult])

	err := rc.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// DeleteImage 删除公共镜像记录
func (rc *restClient) DeleteImage(kt *kit.Kit, request *dataproto.DeleteReq) error {

	resp := new(core.DeleteResp)
	err := rc.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images/batch").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}
