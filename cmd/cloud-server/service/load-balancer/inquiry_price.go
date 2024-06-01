package loadbalancer

import (
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InquiryPriceLoadBalancer inquiry price load balancer.
func (svc *lbSvc) InquiryPriceLoadBalancer(cts *rest.Contexts) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("inquiry price load balancer request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.LoadBalancer, Action: meta.Create, ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("inquiry price load balancer auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.inquiryPriceTCloudLoadBalancer(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *lbSvc) inquiryPriceTCloudLoadBalancer(kt *kit.Kit, body json.RawMessage) (any, error) {
	req := new(cslb.TCloudBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	hcReq := &hcproto.TCloudBatchCreateReq{
		BkBizID:                 constant.UnassignedBiz,
		AccountID:               req.AccountID,
		Region:                  req.Region,
		Name:                    req.Name,
		LoadBalancerType:        req.LoadBalancerType,
		AddressIPVersion:        req.AddressIPVersion,
		Zones:                   req.Zones,
		BackupZones:             req.BackupZones,
		CloudVpcID:              req.CloudVpcID,
		CloudSubnetID:           req.CloudSubnetID,
		Vip:                     req.Vip,
		CloudEipID:              req.CloudEipID,
		VipIsp:                  req.VipIsp,
		InternetChargeType:      req.InternetChargeType,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,
		BandwidthPackageID:      req.BandwidthPackageID,
		SlaType:                 req.SlaType,
		AutoRenew:               req.AutoRenew,
		RequireCount:            req.RequireCount,
		Memo:                    req.Memo,
	}

	result, err := svc.client.HCService().TCloud.Clb.InquiryPrice(kt, hcReq)
	if err != nil {
		logs.Errorf("inquiry price tcloud load balancer failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result, nil
}
