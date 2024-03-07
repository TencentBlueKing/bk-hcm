package tcloud

import (
	"hcm/pkg/api/core"
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
