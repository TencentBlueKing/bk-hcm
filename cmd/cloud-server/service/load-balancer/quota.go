package loadbalancer

import (
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListBizLoadBalancerQuotas 获取业务下的账号配额.
func (svc *lbSvc) ListBizLoadBalancerQuotas(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancerQuotas(cts, handler.BizOperateAuth)
}

// ListResLoadBalancerQuotas 获取资源下的账号配额.
func (svc *lbSvc) ListResLoadBalancerQuotas(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancerQuotas(cts, handler.ResOperateAuth)
}

// listLoadBalancerQuotas list load balancer quota.
func (svc *lbSvc) listLoadBalancerQuotas(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list quota load balancer request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	// validate biz and authorize
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Quota,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, accountID: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.listTCloudLoadBalancerQuota(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *lbSvc) listTCloudLoadBalancerQuota(kt *kit.Kit, body json.RawMessage) (any, error) {
	req := new(hcproto.TCloudListLoadBalancerQuotaReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Clb.ListQuota(kt, req)
	if err != nil {
		logs.Errorf("list tcloud load balancer quota failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result, nil
}
