package tcloud

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/clb"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

type LoadBalancerClient struct {
	client rest.ClientInterface
}

func NewLoadBalancerClient(client rest.ClientInterface) *LoadBalancerClient {
	return &LoadBalancerClient{client: client}
}

// BatchCreateTCloudClb 批量创建腾讯云CLB
func (cli *LoadBalancerClient) BatchCreateTCloudClb(kt *kit.Kit, req *dataproto.TCloudCLBCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudCLBCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/clbs/batch/create")
}

// Get 获取clb 详情
func (cli *LoadBalancerClient) Get(kt *kit.Kit, id string) (*clb.Clb[clb.TCloudClbExtension], error) {

	return common.Request[common.Empty, clb.Clb[clb.TCloudClbExtension]](
		cli.client, rest.GET, kt, nil, "/clbs/%s", id)
}
