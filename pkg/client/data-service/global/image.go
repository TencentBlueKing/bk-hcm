package global

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
)

// ListImage 查询公共镜像列表
func (rc *restClient) ListImage(ctx context.Context, h http.Header, request *dataproto.ImageListReq) (
	*dataproto.ImageListResult, error) {

	resp := new(dataproto.ImageListResp)
	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/images/list").
		WithHeaders(h).
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
func (rc *restClient) DeleteImage(
	ctx context.Context,
	h http.Header,
	request *dataproto.ImageDeleteReq,
) (interface{}, error) {
	resp := new(core.DeleteResp)
	err := rc.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/images/batch").
		WithHeaders(h).
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
